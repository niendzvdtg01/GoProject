package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/package/dtorequest"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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
	Errors    []ImportResult `json:"errors"`
}

func NewUserService(users *repository.UserRepository, auth *middleware.AuthMiddleware, importTasks *repository.ImportTaskRepository) *UserService {
	return &UserService{
		users:       users,
		auth:        auth,
		importTasks: importTasks,
	}
}

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

func (u *UserService) ImportUser(file multipart.File, fileName, userID string, ctx context.Context) ImportSummary {
	taskID, err := u.importTasks.CreateTask(ctx, userID, fileName)
	if err != nil {
		taskID = 0
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		if taskID > 0 {
			u.importTasks.MarkFailed(ctx, taskID, "failed to read CSV file")
		}
		return ImportSummary{TaskID: taskID, Errors: []ImportResult{}}
	}

	totalRows := len(records) - 1
	if totalRows < 0 {
		totalRows = 0
	}

	if taskID > 0 {
		u.importTasks.MarkProcessing(ctx, taskID, totalRows)
	}

	workerCount := 5

	jobs := make(chan dtorequest.RegisterRequest, 100)
	results := make(chan ImportResult, len(records))

	var wg sync.WaitGroup

	for range workerCount {
		wg.Add(1)
		go u.worker(jobs, results, &wg, ctx)
	}

	for i, row := range records {
		if i == 0 {
			continue
		}

		if len(row) < 3 {
			results <- ImportResult{
				Success: false,
				Error:   "invalid row format",
			}
			continue
		}

		role := "member"
		if len(row) >= 4 {
			role = row[3]
		}

		jobs <- dtorequest.RegisterRequest{
			Username: row[0],
			Email:    row[1],
			Password: row[2],
			Role:     role,
		}
	}

	close(jobs)
	wg.Wait()
	close(results)

	summary := ImportSummary{TaskID: taskID, Errors: []ImportResult{}}
	for result := range results {
		if result.Success {
			summary.Succeeded++
		} else {
			summary.Failed++
			summary.Errors = append(summary.Errors, result)
		}
	}

	if taskID > 0 {
		errorLogJSON, _ := json.Marshal(summary.Errors)
		u.importTasks.MarkCompleted(ctx, taskID, summary.Succeeded, summary.Failed, string(errorLogJSON))
	}

	return summary
}

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
