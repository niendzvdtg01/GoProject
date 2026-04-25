package user

import (
	"backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (u *UserHandler) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := utils.ValidationPositive("Id", idStr)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "ID must be valid id",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "New user",
		"user_id": id,
	})
}

func (u *UserHandler) GetUserByUuid(ctx *gin.Context) {
	uuidStr := ctx.Param("uuid")

	uid, err := utils.ValidationUuid("Uuid", uuidStr)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "successfully",
		"uuidStr": uid,
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
