package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/package/dtorequest"
	"backend/package/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var input dtorequest.LoginRequest
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

func (h *AuthHandler) Logout(ctx *gin.Context) {
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
