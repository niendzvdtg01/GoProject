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

func (c *CategoryHandler) GetCategoryHandlerv1(ctx *gin.Context) {
	category := ctx.Param("category")
	if err := utils.ValidationInList("category", category, validCategory); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message":  "category found",
		"category": category,
	})

}
