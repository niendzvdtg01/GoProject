package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/respository"
	"backend/package/dtorequest"
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrInvalidRole = errors.New("role must be manager or member")
var ErrTokenMissingExpiry = errors.New("token missing expiry")

type AuthService struct {
	users *respository.UserRepository
	auth  *middleware.AuthMiddleware
}
type AuthResult struct {
	Token string           `json:"token"`
	User  model.PublicUser `json:"user"`
}

func NewAuthService(users *respository.UserRepository, auth *middleware.AuthMiddleware) *AuthService {
	return &AuthService{
		users: users,
		auth:  auth,
	}
}

func (s *AuthService) Login(ctx context.Context, input dtorequest.LoginRequest) (AuthResult, error) {
	user, err := s.users.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if errors.Is(err, respository.ErrUserNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	// bcrypt comparison is intentionally used instead of comparing password strings.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	token, err := s.auth.GenerateToken(user.UserID, user.Username)
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

	if claims.ExpiresAt == nil {
		return ErrTokenMissingExpiry
	}

	// JWT is stateless, so logout revokes the token id until its natural expiration.
	s.auth.RevokeToken(claims.ID, claims.ExpiresAt.Time)
	return nil
}
