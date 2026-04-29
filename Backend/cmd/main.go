package main

import (
	routing "backend/api"
	"backend/package/utils"
)

func main() {
	if err := utils.RegisterValidators(); err != nil {
		panic(err)
	}

	server := routing.SetupRouter()

	server.Run(":8080")
}
