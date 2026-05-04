package routing

import (
	"backend/internal/config"
	user "backend/internal/handler"
	"backend/internal/middleware"
	database "backend/internal/respository"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(auth *middleware.AuthMiddleware, authService *service.AuthService, userService *service.UserService, users *database.UserRepository, teamService *service.TeamManagementService) *gin.Engine {
	server := gin.Default()

	server.Use(config.NewCorsConfig().CustomCORS())

	serverRouting := server.Group("/api")
	{
		userhandler := user.NewUserHandler(userService, users)
		authhandler := user.NewAuthHandler(authService)
		teamhandler := user.NewTeamHandler(teamService)

		authApi := serverRouting.Group("/auth")
		{
			authApi.POST("/login", authhandler.Login)
			authApi.POST("/logout", auth.AuthRequired(), authhandler.Logout)
		}

		userApi := serverRouting.Group("/users")
		{
			userApi.POST("/register", userhandler.Register)
			userApi.GET("", auth.AuthRequired(), userhandler.ListUsers)
		}

		teamApi := serverRouting.Group("/teams")
		{
			teamApi.POST("", auth.AuthRequired(), teamhandler.CreateTeam)
			teamApi.POST("/:teamName/members", auth.AuthRequired(), teamhandler.AddMember)
			teamApi.DELETE("/:teamName/members/:memberName", auth.AuthRequired(), teamhandler.RemoveMember)
			teamApi.DELETE("/:teamName", auth.AuthRequired(), teamhandler.DeleteTeam)
		}

	}

	return server
}
