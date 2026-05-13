package main

import (
	routing "backend/api"
	"backend/internal/config"
	"backend/internal/middleware"
	database "backend/internal/repository"
	"backend/internal/service"
	"backend/package/utils"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		if err := godotenv.Load("Backend/.env"); err != nil {
			panic("Error loading .env file")
		}
	}
	//
	dbConfig := config.NewDBConfig()

	if err := database.ConnectDB(dbConfig.GetDSN()); err != nil {
		panic(err)
	}

	defer database.CloseDB()
	//
	userRepository := database.NewUserRepository(database.DB)
	teamRepository := database.NewTeamRepository(database.DB)
	teamMemberRepository := database.NewTeamMemberRepository(database.DB)
	folderRepository := database.NewFolderRepository(database.DB)
	noteRepository := database.NewNoteRepository(database.DB)
	permissionRepository := database.NewPermissionRepository(database.DB)
	//
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET is required")
	}
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)
	authService := service.NewAuthService(userRepository, authMiddleware)
	userService := service.NewUserService(userRepository, authMiddleware)
	teamService := service.NewTeamManagementService(teamRepository, teamMemberRepository, userRepository)
	folderService := service.NewFolderService(folderRepository, userRepository, permissionRepository, teamMemberRepository)
	noteService := service.NewNoteService(noteRepository, folderRepository, userRepository, permissionRepository, teamMemberRepository)
	sharingService := service.NewSharing(noteRepository, folderRepository, userRepository, permissionRepository)
	//
	if err := utils.RegisterValidators(); err != nil {
		panic(err)
	}

	server := routing.SetupRouter(authMiddleware, authService, userService, userRepository, teamService, folderService, noteService, sharingService)

	server.Run(":8080")
}
