package handler

import (
	"backend/internal/service"
	"backend/package/dtorequest"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SharingHandler struct {
	sharingService *service.SharingService
}

func NewSharingHandler(sharingService *service.SharingService) *SharingHandler {
	return &SharingHandler{sharingService: sharingService}
}

func (h *SharingHandler) ShareAsset(c *gin.Context) {
	var req dtorequest.ShareAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grantedBy, _ := c.Get("user_id")

	if err := h.sharingService.ShareAsset(req, grantedBy.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access granted"})
}

func (h *SharingHandler) RevokeAccess(c *gin.Context) {
	var req dtorequest.RevokeAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	revokedBy, _ := c.Get("user_id")

	if err := h.sharingService.RevokeAccess(req, revokedBy.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access revoked"})
}
