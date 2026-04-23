package product

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:[-.][a-z0-9]+)*$`)

type Producthandler struct {
}

func NewproductHandler() *Producthandler {
	return &Producthandler{}
}

func (p *Producthandler) GetProductBySlug(ctx *gin.Context) {
	slugStr := ctx.Param("slug")
	if !slugRegex.MatchString(slugStr) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Slug  must contain only lowercase letter, numbers, hyphens, *",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "get all product",
		"slug":    slugStr,
	})
}
