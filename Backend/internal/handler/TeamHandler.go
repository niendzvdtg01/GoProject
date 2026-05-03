package handler

import (
	"backend/internal/service"
	"backend/package/dtorequest"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService *service.TeamManagementService
}

func NewTeamHandler(teamService *service.TeamManagementService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req struct {
		TeamName string                     `json:"teamName" binding:"required"`
		Members  []dtorequest.MemberRequest `json:"members"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (assuming auth middleware sets it)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	teamID, err := h.teamService.CreateTeam(req.TeamName, userID.(string), req.Members)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"teamID": teamID})
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	teamName := c.Param("teamName")
	var req struct {
		MemberName string `json:"memberName" binding:"required"`
		Role       string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	team, err := h.teamService.AddMemberByName(teamName, userID.(string), req.MemberName, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamName := c.Param("teamName")
	memberName := c.Param("memberName")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.teamService.RemoveMemberByName(teamName, userID.(string), memberName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed"})
}

func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	teamName := c.Param("teamName")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.teamService.DeleteTeam(teamName, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted"})
}
