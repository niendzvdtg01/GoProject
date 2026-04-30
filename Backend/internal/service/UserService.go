package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	database "backend/internal/respository"
	"backend/package/dtorequest"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users *database.UserRepository
	auth  *middleware.AuthMiddleware
}

func NewUserService(users *database.UserRepository, auth *middleware.AuthMiddleware) *UserService {
	return &UserService{
		users: users,
		auth:  auth,
	}
}

func (s *UserService) Register(ctx context.Context, input dtorequest.RegisterRequest) (AuthResult, error) {
	// Roles are accepted only at creation time and must match the domain model.
	role := strings.ToLower(strings.TrimSpace(input.Role))
	if !model.IsValidRole(role) {
		return AuthResult{}, ErrInvalidRole
	}

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
		Role:         role,
	}

	if err := s.users.CreateUser(ctx, user); err != nil {
		return AuthResult{}, err
	}

	// Return a signed JWT immediately so the user can start an authenticated session.
	token, err := s.auth.GenerateToken(user.UserID, user.Username, user.Role)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	return AuthResult{
		Token: token,
		User:  user.Public(),
	}, nil
}
