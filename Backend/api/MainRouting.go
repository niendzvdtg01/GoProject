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
		userhandler := user.NewUserHandler(userService, users)
		authhandler := user.NewAuthHandler(authService)

		authApi := serverRouting.Group("/auth")
		{
			authApi.POST("/login", authhandler.Login)
			authApi.POST("/logout", auth.AuthRequired(), authhandler.Logout)
		}

		userApi := serverRouting.Group("/users")
		{
			userApi.POST("/register", userhandler.Register)
		}

	}

	return server
}
