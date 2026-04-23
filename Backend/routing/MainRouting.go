package routing

import (
	"backend/controller/api/v1/product"
	"backend/controller/api/v1/user"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	server := gin.Default()

	serverRouting := server.Group("/api")
	{
		userApi := serverRouting.Group("/user/")
		{
			userhandler := user.NewUserHandler()
			userApi.GET("/test/:id", userhandler.GetUser)
			userApi.GET("/admin/:uuid", userhandler.GetUserByUuid)
		}

		productApi := serverRouting.Group("/product")
		{
			productHandler := product.NewproductHandler()
			productApi.GET("/test/:slug", productHandler.GetProductBySlug)
		}
	}

	return server
}
