package main

import (
	routing "backend/api"
	"backend/internal/config"
	database "backend/internal/respository"
	"backend/package/utils"
)

func main() {
	dbConfig := config.NewDBConfig()

	if err := database.ConnectDB(dbConfig.GetDSN()); err != nil {
		panic(err)
	}

	if err := utils.RegisterValidators(); err != nil {
		panic(err)
	}

	server := routing.SetupRouter()

	server.Run(":8080")
}
