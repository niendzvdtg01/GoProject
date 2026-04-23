package main

import (
	"backend/routing"
)

func main() {

	server := routing.SetupRouter()

	server.Run(":8080")
}
