package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	database "backend/internal/respository"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrInvalidRole = errors.New("role must be manager or member")

type AuthService struct {
	users *database.UserRepository
	auth  *middleware.AuthMiddleware
}

type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Role     string `json:"role" binding:"required,oneof=manager member"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required"`
}

type AuthResult struct {
	Token string           `json:"token"`
	User  model.PublicUser `json:"user"`
}

func NewAuthService(users *database.UserRepository, auth *middleware.AuthMiddleware) *AuthService {
	return &AuthService{
		users: users,
		auth:  auth,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	role := strings.ToLower(strings.TrimSpace(input.Role))
	if !model.IsValidRole(role) {
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
		Role:         role,
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

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	user, err := s.users.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
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

func (s *AuthService) Logout(token string) error {
	claims, err := s.auth.ValidateToken(token)
	if err != nil {
		return err
	}
	s.auth.RevokeToken(claims.ID, claims.ExpiresAt.Time)
	return nil
}
