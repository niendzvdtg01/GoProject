package user

import (
	"backend/package/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
}

type GetUserParamStruct struct {
	ID int `uri:"id" binding:"gt=0"`
}

type GetUserByUuidStruct struct {
	Uuid string `uri:"uuid" binding:"uuid"`
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (u *UserHandler) GetUser(ctx *gin.Context) {
	var params GetUserParamStruct
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": utils.HandleValidatorErrors(err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "New user",
		"user_id": params.ID,
	})
}

func (u *UserHandler) GetUserByUuid(ctx *gin.Context) {
	var uuid GetUserByUuidStruct
	if err := ctx.ShouldBindUri(&uuid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": utils.HandleValidatorErrors(err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "successfully",
		"uuidStr": uuid.Uuid,
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
