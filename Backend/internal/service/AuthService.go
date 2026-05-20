package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/package/dtorequest"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrInvalidRole = errors.New("role must be manager or member")
var ErrTokenMissingExpiry = errors.New("token missing expiry")

type AuthService struct {
	users *repository.UserRepository
	auth  *middleware.AuthMiddleware
}

type AuthResult struct {
	Token string           `json:"token"`
	User  model.PublicUser `json:"user"`
}

func NewAuthService(users *repository.UserRepository, auth *middleware.AuthMiddleware) *AuthService {
	return &AuthService{
		users: users,
		auth:  auth,
	}
}

// Login validates credentials and returns a signed JWT; both wrong-email and wrong-password map to ErrInvalidCredentials to prevent user enumeration.
func (s *AuthService) Login(input dtorequest.LoginRequest) (AuthResult, error) {
	user, err := s.users.GetUserByEmail(strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
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

// Logout adds the token's JTI to the in-memory revocation list; passing the expiry lets the background sweeper reclaim memory automatically.
func (s *AuthService) Logout(token string) error {
	claims, err := s.auth.ValidateToken(token)
	if err != nil {
		return err
	}

	if claims.ExpiresAt == nil {
		return ErrTokenMissingExpiry
	}

	s.auth.RevokeToken(claims.ID, claims.ExpiresAt.Time)
	return nil
}
