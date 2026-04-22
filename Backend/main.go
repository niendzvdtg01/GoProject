package main

import (
	"backend/controller/api/v1/user"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	userhandler := user.NewUserHandler()
	server.GET("/test", userhandler.GetUser)

	server.Run(":8080")
}
