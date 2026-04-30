package user

import (
	"backend/internal/middleware"
	database "backend/internal/respository"
	"backend/internal/service"
	"backend/package/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authService *service.AuthService
	users       *database.UserRepository
}

func NewUserHandler(authService *service.AuthService, users *database.UserRepository) *UserHandler {
	return &UserHandler{
		authService: authService,
		users:       users,
	}
}

func (h *UserHandler) Register(ctx *gin.Context) {
	var input service.RegisterInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.HandleValidatorErrors(err))
		return
	}

	result, err := h.authService.Register(ctx.Request.Context(), input)
	if err != nil {
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

func (h *UserHandler) Login(ctx *gin.Context) {
	var input service.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.HandleValidatorErrors(err))
		return
	}

	result, err := h.authService.Login(ctx.Request.Context(), input)
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

func (h *UserHandler) Logout(ctx *gin.Context) {
	token, ok := middleware.ExtractBearerToken(ctx.GetHeader("Authorization"))
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "valid bearer token required"})
		return
	}

	if err := h.authService.Logout(token); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *UserHandler) ListUsers(ctx *gin.Context) {
	users, err := h.users.ListUsers(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": users})
}
