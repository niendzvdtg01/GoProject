package product

import (
	"backend/utils"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:[-.][a-z0-9]+)*$`)
var searchRegex = regexp.MustCompile(`^[a-zA-Z0-9\s]*$`)

type Producthandler struct {
}

func NewproductHandler() *Producthandler {
	return &Producthandler{}
}

func (p *Producthandler) GetProductBySlug(ctx *gin.Context) {
	slugStr := ctx.Param("slug")

	if err := utils.ValidationCharacter("error:Slug  must contain only lowercase letter, numbers, hyphens, *", slugStr, slugRegex); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "get all product",
		"slug":    slugStr,
	})
}

func (p *Producthandler) GetProductBySearch(ctx *gin.Context) {
	search := ctx.Query("search")

	if err := utils.ValidationRequire("Search", search); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.ValidationStringLength("Search", search, 3, 50); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.ValidationCharacter("search only have numbers, charater, spaces", search, searchRegex); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limitStr := ctx.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "limit must be a positive number",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "List all products",
		"search":  search,
		"limit":   limit,
	})

}
