package news

import "github.com/gin-gonic/gin"

type NewsHandler struct {
}

func NewNewsHandler() *NewsHandler {
	return &NewsHandler{}
}

func (*NewsHandler) GetNewsByIdV1(ctx *gin.Context) {

}
