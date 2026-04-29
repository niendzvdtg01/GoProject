package routing

import (
	"backend/internal/controller/api/v1/news"
	"backend/internal/controller/v1/category"
	"backend/internal/controller/v1/product"
	user "backend/internal/handler"

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
			productApi.GET("/search", productHandler.GetProductBySearch)
		}

		categoryApi := serverRouting.Group("/category")
		{
			categoryHandler := category.NewCategoryhandler()
			categoryApi.GET("/test/:category", categoryHandler.GetCategoryHandlerv1)
		}

		newsApi := serverRouting.Group("/news")
		{
			newsHandler := news.NewNewsHandler()
			newsApi.GET("/:slug", newsHandler.GetNewsByIdV1)
			newsApi.GET("/", newsHandler.GetNewsByIdV1)
		}
	}

	return server
}
