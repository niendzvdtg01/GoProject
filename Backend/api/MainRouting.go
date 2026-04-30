package routing

import (
	user "backend/internal/handler"
	"backend/internal/middleware"
	database "backend/internal/respository"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(auth *middleware.AuthMiddleware, authService *service.AuthService, userService *service.UserService, users *database.UserRepository) *gin.Engine {
	server := gin.Default()

	serverRouting := server.Group("/api")
	{
		userhandler := user.NewUserHandler(authService, userService, users)

		authApi := serverRouting.Group("/auth")
		{
			authApi.POST("/login", userhandler.Login)
			authApi.POST("/logout", auth.AuthRequired(), userhandler.Logout)
		}

		userApi := serverRouting.Group("/users")
		{
			userApi.GET("", auth.AuthRequired(), auth.RoleRequired("manager"), userhandler.ListUsers)
			userApi.POST("/register", userhandler.Register)
		}

	}

	return server
}
