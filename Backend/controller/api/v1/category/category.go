package category

import (
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

	if !validCategory[category] {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Category must be one of: php, python, golang",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message":  "category found",
		"category": category,
	})

}
