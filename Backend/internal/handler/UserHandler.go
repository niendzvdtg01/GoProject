package handler

import (
	"backend/internal/middleware"
	database "backend/internal/respository"
	"backend/internal/service"
	"backend/package/dtorequest"
	"backend/package/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authService *service.AuthService
	userService *service.UserService
	users       *database.UserRepository
}

func NewUserHandler(authService *service.AuthService, userService *service.UserService, users *database.UserRepository) *UserHandler {
	return &UserHandler{
		authService: authService,
		userService: userService,
		users:       users,
	}
}

func (u *UserHandler) Register(ctx *gin.Context) {
	var input dtorequest.RegisterRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.HandleValidatorErrors(err))
		return
	}

	result, err := u.userService.Register(ctx.Request.Context(), input)
	if err != nil {
		// Map domain errors to stable HTTP responses for the API client.
		switch {
		case errors.Is(err, database.ErrEmailAlreadyExists):
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		case errors.Is(err, service.ErrInvalidRole):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (u *UserHandler) Login(ctx *gin.Context) {
	var input dtorequest.LoginRequest
	result, err := u.authService.Login(ctx.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	token, ok := middleware.ExtractBearerToken(ctx.GetHeader("Authorization"))
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "valid bearer token required"})
		return
	}

	if err := u.authService.Logout(token); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (u *UserHandler) ListUsers(ctx *gin.Context) {
	// The route is manager-only; the handler only handles fetching and response shape.
	users, err := u.users.ListUsers(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": users})
}
