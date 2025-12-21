package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// TeamHandler handles team-related HTTP requests
type TeamHandler struct {
	teamSvc service.TeamService
}

// NewTeamHandler creates a new team handler
func NewTeamHandler(teamSvc service.TeamService) *TeamHandler {
	return &TeamHandler{teamSvc: teamSvc}
}

// CreateTeamRequest represents the request body for creating a team
type CreateTeamRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Avatar      *string `json:"avatar"`
	Color       *string `json:"color"`
}

// UpdateTeamRequest represents the request body for updating a team
type UpdateTeamRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Avatar      *string `json:"avatar"`
	Color       *string `json:"color"`
}

// AddTeamMemberRequest represents the request body for adding a team member
type AddTeamMemberRequest struct {
	UserID *string `json:"userId"`
	Email  *string `json:"email"`
	Role   string  `json:"role" binding:"required"`
}

// UpdateTeamMemberRoleRequest represents the request body for updating member role
type UpdateTeamMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// Create creates a new team
func (h *TeamHandler) Create(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamSvc.Create(c.Request.Context(), workspaceID, userID, req.Name, req.Description, req.Avatar, req.Color)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, team)
}

// Get retrieves a team by ID
func (h *TeamHandler) Get(c *gin.Context) {
	teamID := c.Param("id")

	team, err := h.teamSvc.GetByID(c.Request.Context(), teamID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, team)
}

// ListByWorkspace lists all teams in a workspace
func (h *TeamHandler) ListByWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")

	teams, err := h.teamSvc.ListByWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, teams)
}

// ListMyTeams lists all teams the current user is a member of
func (h *TeamHandler) ListMyTeams(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	teams, err := h.teamSvc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, teams)
}

// Update updates a team
func (h *TeamHandler) Update(c *gin.Context) {
	teamID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamSvc.Update(c.Request.Context(), teamID, userID, req.Name, req.Description, req.Avatar, req.Color)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, team)
}

// Delete deletes a team
func (h *TeamHandler) Delete(c *gin.Context) {
	teamID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.teamSvc.Delete(c.Request.Context(), teamID, userID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddMember adds a member to a team
func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID := c.Param("id")
	addedByID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	if req.UserID != nil {
		err = h.teamSvc.AddMember(c.Request.Context(), teamID, *req.UserID, req.Role, addedByID)
	} else if req.Email != nil {
		err = h.teamSvc.AddMemberByEmail(c.Request.Context(), teamID, *req.Email, req.Role, addedByID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId or email is required"})
		return
	}

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "member added"})
}

// ListMembers lists all members of a team
func (h *TeamHandler) ListMembers(c *gin.Context) {
	teamID := c.Param("id")

	members, err := h.teamSvc.ListMembers(c.Request.Context(), teamID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, members)
}

// UpdateMemberRole updates a team member's role
func (h *TeamHandler) UpdateMemberRole(c *gin.Context) {
	teamID := c.Param("id")
	userIDToUpdate := c.Param("userId")

	var req UpdateTeamMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.teamSvc.UpdateMemberRole(c.Request.Context(), teamID, userIDToUpdate, req.Role); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

// RemoveMember removes a member from a team
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID := c.Param("id")
	userIDToRemove := c.Param("userId")

	if err := h.teamSvc.RemoveMember(c.Request.Context(), teamID, userIDToRemove); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
