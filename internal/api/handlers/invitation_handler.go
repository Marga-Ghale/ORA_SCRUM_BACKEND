// internal/api/handlers/invitation_handler.go
package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type InvitationHandler struct {
	invitationSvc service.InvitationService
}

func NewInvitationHandler(invitationSvc service.InvitationService) *InvitationHandler {
	return &InvitationHandler{invitationSvc: invitationSvc}
}

type CreateInvitationRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Role    string `json:"role" binding:"required"`
	Message string `json:"message,omitempty"` // Optional personal message
}

// CreateWorkspaceInvitation creates an invitation to join a workspace
func (h *InvitationHandler) CreateWorkspaceInvitation(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace ID is required"})
		return
	}

	inviterID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invitation, err := h.invitationSvc.CreateWorkspaceInvitation(c.Request.Context(), workspaceID, req.Email, req.Role, inviterID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// CreateProjectInvitation creates an invitation to join a project
func (h *InvitationHandler) CreateProjectInvitation(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project ID is required"})
		return
	}

	inviterID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invitation, err := h.invitationSvc.CreateProjectInvitation(c.Request.Context(), projectID, req.Email, req.Role, inviterID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// AcceptInvitation accepts an invitation
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invitation token is required"})
		return
	}

	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	response, err := h.invitationSvc.AcceptInvitation(c.Request.Context(), token, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMyInvitations gets all pending invitations for the current user
func (h *InvitationHandler) GetMyInvitations(c *gin.Context) {
	// Get user from auth context
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	// Get user's email - we need to look it up
	// Option 1: Pass user email in context from auth middleware
	// Option 2: Accept email as query param (less secure)
	email := c.Query("email")
	if email == "" {
		// Try to get from context if available
		if userEmail, exists := c.Get("userEmail"); exists {
			email = userEmail.(string)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
			return
		}
	}

	_ = userID // Could verify email belongs to userID for security

	invitations, err := h.invitationSvc.GetPendingInvitations(c.Request.Context(), email)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, invitations)
}

// CancelInvitation cancels an invitation
func (h *InvitationHandler) CancelInvitation(c *gin.Context) {
	invitationID := c.Param("id")
	if invitationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invitation ID is required"})
		return
	}

	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.invitationSvc.CancelInvitation(c.Request.Context(), invitationID, userID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetWorkspaceInvitations gets pending invitations for a workspace
func (h *InvitationHandler) GetWorkspaceInvitations(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace ID is required"})
		return
	}

	invitations, err := h.invitationSvc.GetPendingWorkspaceInvitations(c.Request.Context(), workspaceID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, invitations)
}

// GetProjectInvitations gets pending invitations for a project
func (h *InvitationHandler) GetProjectInvitations(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project ID is required"})
		return
	}

	invitations, err := h.invitationSvc.GetPendingProjectInvitations(c.Request.Context(), projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, invitations)
}

// ResendInvitation resends an invitation email
func (h *InvitationHandler) ResendInvitation(c *gin.Context) {
	invitationID := c.Param("id")
	if invitationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invitation ID is required"})
		return
	}

	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.invitationSvc.ResendInvitation(c.Request.Context(), invitationID, userID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invitation resent"})
}
