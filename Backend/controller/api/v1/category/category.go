package category

import (
	"backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
}

//map object

var validCategory = map[string]bool{
	"php":    true,
	"python": true,
	"java":   false,
	"golang": true,
}

func NewCategoryhandler() *CategoryHandler {
	return &CategoryHandler{}
}

type GetCategoryByParam struct {
	Category string `uri:"category" binding:"oneof=php go python"`
}

func (c *CategoryHandler) GetCategoryHandlerv1(ctx *gin.Context) {
	var param GetCategoryByParam
	if err := ctx.ShouldBindUri(&param); err != nil {
		ctx.JSON(http.StatusBadGateway, utils.HandleValidatorErrors(err))
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{
		"message": "Get category by category (v1)",
	})

}
