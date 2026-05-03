package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/respository"
	"backend/package/dtorequest"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users *respository.UserRepository
	auth  *middleware.AuthMiddleware
}

func NewUserService(users *respository.UserRepository, auth *middleware.AuthMiddleware) *UserService {
	return &UserService{
		users: users,
		auth:  auth,
	}
}

func (s *UserService) Register(ctx context.Context, input dtorequest.RegisterRequest) (AuthResult, error) {

	// Store only the bcrypt hash; never persist the raw password.
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		UserID:       uuid.NewString(),
		Username:     strings.TrimSpace(input.Username),
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: string(passwordHash),
	}

	if err := s.users.CreateUser(ctx, user); err != nil {
		return AuthResult{}, err
	}

	// Return a signed JWT immediately so the user can start an authenticated session.
	token, err := s.auth.GenerateToken(user.UserID, user.Username)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	return AuthResult{
		Token: token,
		User:  user.Public(),
	}, nil
}
