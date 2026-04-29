package routing

import (
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

	}

	return server
}
