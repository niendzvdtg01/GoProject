package handler

import (
	database "backend/internal/respository"
	"backend/internal/service"
	"backend/package/dtorequest"
	"backend/package/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
	users       *database.UserRepository
}

func NewUserHandler(userService *service.UserService, users *database.UserRepository) *UserHandler {
	return &UserHandler{
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

func (u *UserHandler) ListUsers(ctx *gin.Context) {
	// The route is manager-only; the handler only handles fetching and response shape.
	users, err := u.users.ListUsers(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": users})
}
