package product

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Producthandler struct {
}

func NewproductHandler() *Producthandler {
	return &Producthandler{}
}

func (p *Producthandler) getProduct(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "get all product",
	})
}
