package main

import (
	routing "backend/api"
	"backend/internal/config"
	database "backend/internal/respository"
	"backend/package/utils"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}
	//
	dbConfig := config.NewDBConfig()

	if err := database.ConnectDB(dbConfig.GetDSN()); err != nil {
		panic(err)
	}

	defer database.CloseDB()
	//
	if err := utils.RegisterValidators(); err != nil {
		panic(err)
	}

	server := routing.SetupRouter()

	server.Run(":8080")
}
