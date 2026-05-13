package handler

import (
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ImportHandler struct {
	userService *service.UserService
}

func NewImportHandler(userService *service.UserService) *ImportHandler {
	return &ImportHandler{userService: userService}
}

func (h *ImportHandler) ImportUsers(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	summary := h.userService.ImportUser(file, c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"succeeded": summary.Succeeded,
		"failed":    summary.Failed,
		"errors":    summary.Errors,
	})
}
