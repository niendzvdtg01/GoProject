package handler

import (
	"backend/internal/repository"
	"backend/internal/service"
	"backend/package/dtorequest"
	"backend/package/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
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
		switch {
		case errors.Is(err, repository.ErrEmailAlreadyExists):
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (u *UserHandler) ListUsers(ctx *gin.Context) {
	users, err := u.userService.ListUsers(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": users})
}
