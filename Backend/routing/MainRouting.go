package routing

import (
	"backend/controller/api/v1/user"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	server := gin.Default()

	userRouting := server.Group("/api")
	{
		userApi := userRouting.Group("/user/")
		{
			userhandler := user.NewUserHandler()
			userApi.GET("/test/:id", userhandler.GetUser)
			userApi.GET("/admin/:uuid", userhandler.GetUserByUuid)
		}
	}

	return server
}
