package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (u *UserHandler) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "ID must be a number",
		})
	}
	if id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Id must be a negative number!",
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "New user",
	})
}
func (u *UserHandler) PostUser(ctx *gin.Context) {
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Create user",
	})
}
func (u *UserHandler) PutUser(ctx *gin.Context) {
	ctx.JSON(http.StatusAccepted, gin.H{
		"message": "Update user",
	})
}
func (u *UserHandler) DeleteUser(ctx *gin.Context) {
	ctx.JSON(http.StatusNoContent, gin.H{
		"message": "Delete user",
	})
}
