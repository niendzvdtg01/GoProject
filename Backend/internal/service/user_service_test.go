package service

import (
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/package/dtorequest"
	"context"
	"errors"
	"mime/multipart"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

// testMultipartFile wraps strings.Reader to satisfy multipart.File
type testMultipartFile struct{ *strings.Reader }

func (f *testMultipartFile) Close() error { return nil }

var _ multipart.File = (*testMultipartFile)(nil)

func newCSVFile(content string) *testMultipartFile {
	return &testMultipartFile{strings.NewReader(content)}
}

func newUserSetup(t *testing.T) (*UserService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	auth := middleware.NewAuthMiddleware("test-secret")
	svc := NewUserService(repository.NewUserRepository(db), auth)
	return svc, mock, func() { db.Close() }
}

func TestRegister_Success(t *testing.T) {
	svc, mock, cleanup := newUserSetup(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WithArgs(sqlmock.AnyArg(), "alice", "alice@example.com", sqlmock.AnyArg(), "member").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Register(context.Background(), dtorequest.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
		Role:     "member",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
	if result.User.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", result.User.Email)
	}
}

func TestRegister_EmailNormalized(t *testing.T) {
	svc, mock, cleanup := newUserSetup(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WithArgs(sqlmock.AnyArg(), "alice", "alice@example.com", sqlmock.AnyArg(), "member").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Register(context.Background(), dtorequest.RegisterRequest{
		Username: "  alice  ",
		Email:    "  ALICE@EXAMPLE.COM  ",
		Password: "password123",
		Role:     "member",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User.Username != "alice" {
		t.Errorf("expected trimmed username alice, got %s", result.User.Username)
	}
	if result.User.Email != "alice@example.com" {
		t.Errorf("expected lowercased email, got %s", result.User.Email)
	}
}

func TestRegister_InvalidRole(t *testing.T) {
	svc, _, cleanup := newUserSetup(t)
	defer cleanup()

	_, err := svc.Register(context.Background(), dtorequest.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
		Role:     "admin",
	})
	if !errors.Is(err, ErrInvalidRole) {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	svc, mock, cleanup := newUserSetup(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry"})

	_, err := svc.Register(context.Background(), dtorequest.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
		Role:     "member",
	})
	if !errors.Is(err, repository.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestRegister_ManagerRole(t *testing.T) {
	svc, mock, cleanup := newUserSetup(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WithArgs(sqlmock.AnyArg(), "bob", "bob@example.com", sqlmock.AnyArg(), "manager").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Register(context.Background(), dtorequest.RegisterRequest{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "password123",
		Role:     "manager",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User.Role != "manager" {
		t.Errorf("expected role manager, got %s", result.User.Role)
	}
}

func TestListUsers_Success(t *testing.T) {
	svc, mock, cleanup := newUserSetup(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "role", "created_at"}).
		AddRow("uid-1", "alice", "alice@example.com", "member", now).
		AddRow("uid-2", "bob", "bob@example.com", "manager", now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, role, created_at FROM users ORDER BY created_at DESC")).
		WillReturnRows(rows)

	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
	if users[0].Email != "alice@example.com" {
		t.Errorf("expected alice, got %s", users[0].Email)
	}
}

func TestListUsers_Empty(t *testing.T) {
	svc, mock, cleanup := newUserSetup(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "role", "created_at"})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, role, created_at FROM users ORDER BY created_at DESC")).
		WillReturnRows(rows)

	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestImportUser_Success(t *testing.T) {
	svc, mock, cleanup := newUserSetup(t)
	defer cleanup()

	// Two valid rows (+ header)
	csv := "username,email,password,role\nalice,alice@example.com,password1,member\nbob,bob@example.com,password2,manager\n"

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WillReturnResult(sqlmock.NewResult(2, 1))

	summary := svc.ImportUser(newCSVFile(csv), context.Background())
	if summary.Succeeded != 2 {
		t.Errorf("expected 2 succeeded, got %d", summary.Succeeded)
	}
	if summary.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", summary.Failed)
	}
}

func TestImportUser_InvalidRowFormat(t *testing.T) {
	svc, _, cleanup := newUserSetup(t)
	defer cleanup()

	// Row with only 2 columns (needs at least 3)
	csv := "username,email\nalice,alice@example.com\n"

	summary := svc.ImportUser(newCSVFile(csv), context.Background())
	if summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", summary.Failed)
	}
}

func TestImportUser_DefaultRoleMember(t *testing.T) {
	svc, mock, cleanup := newUserSetup(t)
	defer cleanup()

	// Row with only 3 columns — role defaults to member
	csv := "username,email,password\nalice,alice@example.com,password1\n"

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WithArgs(sqlmock.AnyArg(), "alice", "alice@example.com", sqlmock.AnyArg(), "member").
		WillReturnResult(sqlmock.NewResult(1, 1))

	summary := svc.ImportUser(newCSVFile(csv), context.Background())
	if summary.Succeeded != 1 {
		t.Errorf("expected 1 succeeded, got %d", summary.Succeeded)
	}
}

func TestImportUser_DBError(t *testing.T) {
	svc, mock, cleanup := newUserSetup(t)
	defer cleanup()

	csv := "username,email,password,role\nalice,alice@example.com,password1,member\n"

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WillReturnError(errors.New("db error"))

	summary := svc.ImportUser(newCSVFile(csv), context.Background())
	if summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", summary.Failed)
	}
	if len(summary.Errors) != 1 {
		t.Errorf("expected 1 error entry, got %d", len(summary.Errors))
	}
}
