package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/package/dtorequest"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users *repository.UserRepository
	auth  *middleware.AuthMiddleware
}

func NewUserService(users *repository.UserRepository, auth *middleware.AuthMiddleware) *UserService {
	return &UserService{
		users: users,
		auth:  auth,
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
