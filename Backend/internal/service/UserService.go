package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/package/dtorequest"
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrTooManyActiveImports = errors.New("you already have an active import task in progress, wait for it to complete")

const maxConcurrentImportsPerUser = 1

type UserService struct {
	users       *repository.UserRepository
	auth        *middleware.AuthMiddleware
	importTasks *repository.ImportTaskRepository
}

type ImportResult struct {
	Success bool   `json:"success"`
	Email   string `json:"email"`
	Error   string `json:"error,omitempty"`
}

type ImportSummary struct {
	TaskID    int64          `json:"task_id,omitempty"`
	Succeeded int            `json:"succeeded"`
	Failed    int            `json:"failed"`
	Processed int            `json:"processed"`
	Errors    []ImportResult `json:"errors"`
}

func NewUserService(users *repository.UserRepository, auth *middleware.AuthMiddleware, importTasks *repository.ImportTaskRepository) *UserService {
	return &UserService{
		users:       users,
		auth:        auth,
		importTasks: importTasks,
	}
}

// Register creates a user account and returns a ready-to-use JWT; role is validated before hashing to avoid wasted bcrypt work.
func (s *UserService) Register(ctx context.Context, input dtorequest.RegisterRequest) (AuthResult, error) {
	if input.Role != "manager" && input.Role != "member" {
		return AuthResult{}, ErrInvalidRole
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		UserID:       uuid.NewString(),
		Username:     strings.TrimSpace(input.Username),
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: string(passwordHash),
		Role:         input.Role,
	}

	if err := s.users.CreateUser(ctx, user); err != nil {
		return AuthResult{}, err
	}

	token, err := s.auth.GenerateToken(user.UserID, user.Username, user.Role)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	return AuthResult{
		Token: token,
		User:  user.Public(),
	}, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]model.PublicUser, error) {
	return s.users.ListUsers(ctx)
}

// ImportUser runs a CSV import synchronously; use CreateImportTask + ProcessImportAsync from HTTP handlers to avoid blocking.
func (u *UserService) ImportUser(file io.Reader, fileName, userID string, ctx context.Context) ImportSummary {
	taskID, err := u.importTasks.CreateTask(ctx, userID, fileName)
	if err != nil {
		return ImportSummary{Errors: []ImportResult{{Success: false, Error: err.Error()}}}
	}
	summary, err := u.importProcess(ctx, taskID, file, 0)
	if err != nil {
		u.importTasks.MarkFailed(context.Background(), taskID, err.Error())
	}
	return summary
}

// CreateImportTask persists an import_tasks row in "pending" state and returns the task ID before the background goroutine starts.
// Returns ErrTooManyActiveImports if the user already has a pending or processing task.
func (u *UserService) CreateImportTask(fileName, userID string, ctx context.Context) (int64, error) {
	count, err := u.importTasks.CountActiveByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("check active tasks: %w", err)
	}
	if count >= maxConcurrentImportsPerUser {
		return 0, ErrTooManyActiveImports
	}
	return u.importTasks.CreateTask(ctx, userID, fileName)
}

// ProcessImportAsync is the goroutine entry point for async CSV imports; cleans up the temp file, caps runtime at 1 hour, and converts panics to task failures.
func (u *UserService) ProcessImportAsync(taskID int64, filePath, userID string) {
	defer os.Remove(filePath)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in import task %d: %v\n%s", taskID, r, debug.Stack())
			u.importTasks.MarkFailed(context.Background(), taskID, fmt.Sprintf("internal error: %v", r))
		}
	}()

	totalRows, _ := countCSVRows(filePath)
	f, err := os.Open(filePath)
	if err != nil {
		u.importTasks.MarkFailed(context.Background(), taskID, fmt.Sprintf("failed to open file: %v", err))
		return
	}
	defer f.Close()

	if _, err := u.importProcess(ctx, taskID, f, totalRows); err != nil {
		log.Printf("import task %d failed: %v", taskID, err)
		u.importTasks.MarkFailed(context.Background(), taskID, err.Error())
	}
}

// worker drains jobs and calls Register for each row, reusing its business-rule validation for the bulk path.
func (u *UserService) worker(jobs <-chan dtorequest.RegisterRequest, results chan<- ImportResult, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	for user := range jobs {
		_, err := u.Register(ctx, user)
		if err != nil {
			results <- ImportResult{
				Success: false,
				Email:   user.Email,
				Error:   err.Error(),
			}
		} else {
			results <- ImportResult{
				Success: true,
				Email:   user.Email,
			}
		}
	}
}

// importProcess runs a fan-out/fan-in worker pool over the CSV and returns the final ImportSummary with dual-trigger progress updates (every 500 rows or 2 s).
func (u *UserService) importProcess(ctx context.Context, taskID int64, file io.Reader, totalRows int) (ImportSummary, error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // tolerate varying column counts

	// Skip header before marking processing so it isn't counted as a failed row.
	if _, err := reader.Read(); err != nil {
		return ImportSummary{}, fmt.Errorf("cannot read header: %w", err)
	}

	u.importTasks.MarkProcessing(ctx, taskID, totalRows)

	var (
		workerCount   = runtime.NumCPU()
		jobBufferSize = 100
	)

	jobs := make(chan dtorequest.RegisterRequest, jobBufferSize)
	results := make(chan ImportResult, jobBufferSize)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go u.worker(jobs, results, &wg, ctx)
	}

	go func() { // closer: signals aggregator when all workers finish
		wg.Wait()
		close(results)
	}()

	go func() { // producer: parses CSV rows and dispatches to the worker pool
		defer close(jobs)
		rowNum := 1
		for {
			row, err := reader.Read()
			if err == io.EOF {
				return
			}
			rowNum++

			if err != nil {
				select {
				case results <- ImportResult{Success: false, Error: fmt.Sprintf("row %d: parse error", rowNum)}:
				case <-ctx.Done():
					return
				}
				continue
			}

			if len(row) < 3 {
				select {
				case results <- ImportResult{Success: false, Error: fmt.Sprintf("row %d: invalid format", rowNum)}:
				case <-ctx.Done():
					return
				}
				continue
			}

			role := "member" // default when role column is absent
			if len(row) >= 4 {
				role = row[3]
			}

			req := dtorequest.RegisterRequest{
				Username: row[0],
				Email:    row[1],
				Password: row[2],
				Role:     role,
			}
			select {
			case jobs <- req:
			case <-ctx.Done():
				return
			}
		}
	}()

	summary := ImportSummary{TaskID: taskID, Errors: []ImportResult{}}
	const progressBatch = 500
	progressTicker := time.NewTicker(2 * time.Second)
	defer progressTicker.Stop()
	lastReported := 0

	for {
		select {
		case result, ok := <-results:
			if !ok {
				goto done // all workers finished
			}
			summary.Processed++
			if result.Success {
				summary.Succeeded++
			} else {
				summary.Failed++
				if len(summary.Errors) < 1000 {
					summary.Errors = append(summary.Errors, result)
				}
			}
			if summary.Processed-lastReported >= progressBatch {
				u.importTasks.UpdateProgress(context.Background(), taskID, summary.Processed)
				lastReported = summary.Processed
			}
		case <-progressTicker.C:
			if summary.Processed > lastReported {
				u.importTasks.UpdateProgress(context.Background(), taskID, summary.Processed)
				lastReported = summary.Processed
			}
		case <-ctx.Done():
			return summary, ctx.Err()
		}
	}

done:
	errorLogJSON, _ := json.Marshal(summary.Errors)
	err := u.importTasks.MarkCompleted(
		context.Background(), taskID, summary.Succeeded, summary.Failed, string(errorLogJSON))
	return summary, err
}

// countCSVRows counts data rows (excluding header) using a 1 MB buffer to handle long lines without reading the whole file into memory.
func countCSVRows(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		count++
	}
	return count - 1, scanner.Err() // subtract header row
}

func (u *UserService) GetImportTask(ctx context.Context, taskID int64) (*model.ImportTask, error) {
	return u.importTasks.GetTask(ctx, taskID)
}
