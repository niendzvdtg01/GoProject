package news

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type NewsHandler struct {
}

type CC interface {
}

func NewNewsHandler() *NewsHandler {
	return &NewsHandler{}
}

func (n *NewsHandler) GetNewsByIdV1(ctx *gin.Context) {
	slug := ctx.Param("slug")

	if slug == "" {
		ctx.JSON(http.StatusAccepted, gin.H{
			"message": "Get news V1",
			"slug":    "No news",
		})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{
			"message": "Get news V1",
			"slug":    slug,
		})
	}
}
