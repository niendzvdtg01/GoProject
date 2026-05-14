package service

import (
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/package/dtorequest"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
)

func newAuthSetup(t *testing.T) (*AuthService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	auth := middleware.NewAuthMiddleware("test-secret")
	svc := NewAuthService(repository.NewUserRepository(db), auth)
	return svc, mock, func() { db.Close() }
}

func hashPassword(t *testing.T, pw string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return string(h)
}

func TestLogin_Success(t *testing.T) {
	svc, mock, cleanup := newAuthSetup(t)
	defer cleanup()

	hash := hashPassword(t, "password123")
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash", "role", "created_at"}).
		AddRow("uid-1", "alice", "alice@example.com", hash, "member", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(rows)

	result, err := svc.Login(dtorequest.LoginRequest{Email: "alice@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
	if result.User.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", result.User.Email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestLogin_EmailNormalized(t *testing.T) {
	svc, mock, cleanup := newAuthSetup(t)
	defer cleanup()

	hash := hashPassword(t, "password123")
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash", "role", "created_at"}).
		AddRow("uid-1", "alice", "alice@example.com", hash, "member", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(rows)

	// Email with uppercase and spaces should be normalized
	_, err := svc.Login(dtorequest.LoginRequest{Email: "  ALICE@EXAMPLE.COM  ", Password: "password123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, mock, cleanup := newAuthSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("nobody@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Login(dtorequest.LoginRequest{Email: "nobody@example.com", Password: "pass"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, mock, cleanup := newAuthSetup(t)
	defer cleanup()

	hash := hashPassword(t, "correctpassword")
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash", "role", "created_at"}).
		AddRow("uid-1", "alice", "alice@example.com", hash, "member", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(rows)

	_, err := svc.Login(dtorequest.LoginRequest{Email: "alice@example.com", Password: "wrongpassword"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_DBError(t *testing.T) {
	svc, mock, cleanup := newAuthSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnError(errors.New("connection refused"))

	_, err := svc.Login(dtorequest.LoginRequest{Email: "alice@example.com", Password: "pass"})
	if err == nil {
		t.Fatal("expected error for DB failure")
	}
}

func TestLogout_Success(t *testing.T) {
	auth := middleware.NewAuthMiddleware("test-secret")
	token, _ := auth.GenerateToken("uid-1", "alice", "member")

	svc := NewAuthService(repository.NewUserRepository(nil), auth)
	err := svc.Logout(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogout_InvalidToken(t *testing.T) {
	auth := middleware.NewAuthMiddleware("test-secret")
	svc := NewAuthService(repository.NewUserRepository(nil), auth)
	err := svc.Logout("not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
