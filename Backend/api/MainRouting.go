package routing

import (
	"backend/internal/config"
	user "backend/internal/handler"
	"backend/internal/middleware"
	database "backend/internal/repository"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(auth *middleware.AuthMiddleware, authService *service.AuthService, userService *service.UserService, users *database.UserRepository, teamService *service.TeamManagementService, folderService *service.FolderService, noteService *service.NoteService, sharingService *service.SharingService) *gin.Engine {
	server := gin.Default()

	server.Use(config.NewCorsConfig().CustomCORS())

	serverRouting := server.Group("/api")
	{
		userhandler := user.NewUserHandler(userService)
		authhandler := user.NewAuthHandler(authService)
		teamhandler := user.NewTeamHandler(teamService)
		folderHandler := user.NewFolderHandler(folderService)
		noteHandler := user.NewNoteHandler(noteService)
		sharingHandler := user.NewSharingHandler(sharingService)
		importHandler := user.NewImportHandler(userService)

		authApi := serverRouting.Group("/auth")
		{
			authApi.POST("/login", authhandler.Login)
			authApi.POST("/logout", auth.AuthRequired(), authhandler.Logout)
		}

		userApi := serverRouting.Group("/users")
		{
			userApi.POST("/register", userhandler.Register)
			userApi.GET("", auth.AuthRequired(), userhandler.ListUsers)
			userApi.POST("/import", auth.AuthRequired(), importHandler.ImportUsers)
		}

		teamApi := serverRouting.Group("/teams")
		{
			teamApi.GET("", auth.AuthRequired(), teamhandler.ListTeams)
			teamApi.POST("", auth.AuthRequired(), auth.RequireManager(), teamhandler.CreateTeam)
			teamApi.POST("/:teamName/members", auth.AuthRequired(), auth.RequireManager(), teamhandler.AddMember)
			teamApi.DELETE("/:teamName/members/:memberName", auth.AuthRequired(), auth.RequireManager(), teamhandler.RemoveMember)
			teamApi.DELETE("/:teamName", auth.AuthRequired(), auth.RequireManager(), teamhandler.DeleteTeam)
		}

		folderApi := serverRouting.Group("/folders")
		{
			folderApi.POST("", auth.AuthRequired(), folderHandler.CreateFolder)
			folderApi.GET("", auth.AuthRequired(), folderHandler.ListFolders)
			folderApi.GET("/:id", auth.AuthRequired(), folderHandler.GetFolder)
			folderApi.PUT("/:id", auth.AuthRequired(), folderHandler.UpdateFolder)
			folderApi.DELETE("/:id", auth.AuthRequired(), folderHandler.DeleteFolder)
			folderApi.POST("/:id/notes", auth.AuthRequired(), noteHandler.CreateNote)
			folderApi.GET("/:id/notes", auth.AuthRequired(), noteHandler.ListNotes)
		}

		noteApi := serverRouting.Group("/notes")
		{
			noteApi.GET("/:id", auth.AuthRequired(), noteHandler.GetNote)
			noteApi.PUT("/:id", auth.AuthRequired(), noteHandler.UpdateNote)
			noteApi.DELETE("/:id", auth.AuthRequired(), noteHandler.DeleteNote)
		}

		shareApi := serverRouting.Group("/share")
		{
			shareApi.POST("", auth.AuthRequired(), sharingHandler.ShareAsset)
			shareApi.DELETE("", auth.AuthRequired(), sharingHandler.RevokeAccess)
		}
	}

	return server
}
 