package handler

import (
	"backend/internal/service"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 10 * 1024 * 1024 // 10 MB

type ImportHandler struct {
	userService *service.UserService
}

func NewImportHandler(userService *service.UserService) *ImportHandler {
	return &ImportHandler{userService: userService}
}

func (h *ImportHandler) ImportUsers(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		if err.Error() == "http: request body too large" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large, maximum 10MB"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if fileHeader.Size > maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large, maximum 10MB"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer src.Close()

	tmpFile, err := os.CreateTemp("", "import-*.csv")
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file"})
		return
	}
	tmpPath := tmpFile.Name()

	if _, err = io.Copy(tmpFile, src); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	tmpFile.Close()

	taskID, err := h.userService.CreateImportTask(fileHeader.Filename, userIDStr, c.Request.Context())
	if err != nil {
		os.Remove(tmpPath)
		if errors.Is(err, service.ErrTooManyActiveImports) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	go h.userService.ProcessImportAsync(taskID, tmpPath, userIDStr)

	c.Header("Location", fmt.Sprintf("/api/import-tasks/%d", taskID))
	c.JSON(http.StatusAccepted, gin.H{
		"task_id":    taskID,
		"status_url": fmt.Sprintf("/api/import-tasks/%d", taskID),
		"message":    "Import is being processed",
	})
}

func (h *ImportHandler) GetImportTask(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	task, err := h.userService.GetImportTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}
