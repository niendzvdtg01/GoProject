package main

import (
	routing "backend/api"
	"backend/internal/config"
	"backend/internal/middleware"
	database "backend/internal/respository"
	"backend/internal/service"
	"backend/package/utils"
	"context"
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
	if err := userRepository.EnsureUsersTable(context.Background()); err != nil {
		panic(err)
	}
	//
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET is required")
	}
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)
	authService := service.NewAuthService(userRepository, authMiddleware)
	//
	if err := utils.RegisterValidators(); err != nil {
		panic(err)
	}

	server := routing.SetupRouter(authMiddleware, authService, userRepository)

	server.Run(":8080")
}
