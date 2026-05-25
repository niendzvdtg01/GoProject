package handler

import (
	"backend/internal/repository"
	"backend/internal/service"
	"backend/package/dtorequest"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FolderHandler struct {
	folderService *service.FolderService
}

func NewFolderHandler(folderService *service.FolderService) *FolderHandler {
	return &FolderHandler{folderService: folderService}
}

func (h *FolderHandler) CreateFolder(c *gin.Context) {
	var req dtorequest.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ownerID, _ := c.Get("user_id")

	folder, err := h.folderService.CreateFolder(c.Request.Context(), ownerID.(string), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

func (h *FolderHandler) ListFolders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	folders, err := h.folderService.ListUserFolders(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list folders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"folders": folders})
}

func (h *FolderHandler) GetFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}

	requesterID, _ := c.Get("user_id")

	folder, err := h.folderService.GetFolder(c.Request.Context(), id, requesterID.(string))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrFolderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get folder"})
		}
		return
	}

	c.JSON(http.StatusOK, folder)
}

func (h *FolderHandler) UpdateFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}

	var req dtorequest.UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requesterID, _ := c.Get("user_id")

	folder, err := h.folderService.UpdateFolder(c.Request.Context(), id, requesterID.(string), req.Name)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrFolderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update folder"})
		}
		return
	}

	c.JSON(http.StatusOK, folder)
}

func (h *FolderHandler) DeleteFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}

	requesterID, _ := c.Get("user_id")

	if err := h.folderService.DeleteFolder(c.Request.Context(), id, requesterID.(string)); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrFolderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete folder"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "folder deleted"})
}
