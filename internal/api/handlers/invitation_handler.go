package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type InvitationHandler struct {
	invSvc service.InvitationService
}

func NewInvitationHandler(invSvc service.InvitationService) *InvitationHandler {
	return &InvitationHandler{invSvc: invSvc}
}

// CreateWorkspaceInvitation godoc
// @Summary Create workspace invitation
// @Tags invitations
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body CreateInvitationRequest true "Invitation details"
// @Success 201 {object} map[string]interface{}
// @Router /workspaces/{id}/invitations [post]
func (h *InvitationHandler) CreateWorkspaceInvitation(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := c.GetString("user_id")

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv, err := h.invSvc.CreateWorkspaceInvitation(c.Request.Context(), workspaceID, req.Email, req.Role, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Invitation sent successfully",
		"invitation": inv,
	})
}

// CreateProjectInvitation godoc
// @Summary Create project invitation
// @Tags invitations
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body CreateInvitationRequest true "Invitation details"
// @Success 201 {object} map[string]interface{}
// @Router /projects/{id}/invitations [post]
func (h *InvitationHandler) CreateProjectInvitation(c *gin.Context) {
	projectID := c.Param("id")
	userID := c.GetString("user_id")

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv, err := h.invSvc.CreateProjectInvitation(c.Request.Context(), req.WorkspaceID, projectID, req.Email, req.Role, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Invitation sent successfully",
		"invitation": inv,
	})
}

// GetWorkspaceInvitations godoc
// @Summary Get workspace invitations
// @Tags invitations
// @Produce json
// @Param id path string true "Workspace ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Router /workspaces/{id}/invitations [get]
func (h *InvitationHandler) GetWorkspaceInvitations(c *gin.Context) {
	workspaceID := c.Param("id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	invitations, total, err := h.invSvc.ListByWorkspace(c.Request.Context(), workspaceID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetProjectInvitations godoc
// @Summary Get project invitations
// @Tags invitations
// @Produce json
// @Param id path string true "Project ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Router /projects/{id}/invitations [get]
func (h *InvitationHandler) GetProjectInvitations(c *gin.Context) {
	projectID := c.Param("id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	invitations, total, err := h.invSvc.ListByProject(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetMyInvitations godoc
// @Summary Get my pending invitations
// @Tags invitations
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /invitations/pending [get]
func (h *InvitationHandler) GetMyInvitations(c *gin.Context) {
	userEmail := c.GetString("user_email")

	invitations, err := h.invSvc.GetMyInvitations(c.Request.Context(), userEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
	})
}

// AcceptInvitation godoc
// @Summary Accept invitation
// @Tags invitations
// @Produce json
// @Param token path string true "Invitation Token"
// @Success 200 {object} map[string]interface{}
// @Router /invitations/accept/{token} [post]
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	userID := c.GetString("user_id")

	err := h.invSvc.AcceptByToken(c.Request.Context(), token, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation accepted successfully",
	})
}

// ResendInvitation godoc
// @Summary Resend invitation
// @Tags invitations
// @Produce json
// @Param id path string true "Invitation ID"
// @Success 200 {object} map[string]interface{}
// @Router /invitations/resend/{id} [post]
func (h *InvitationHandler) ResendInvitation(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	inv, err := h.invSvc.ResendInvitation(c.Request.Context(), id, &userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Invitation resent successfully",
		"invitation": inv,
	})
}

// CancelInvitation godoc
// @Summary Cancel invitation
// @Tags invitations
// @Produce json
// @Param id path string true "Invitation ID"
// @Success 200 {object} map[string]interface{}
// @Router /invitations/{id} [delete]
func (h *InvitationHandler) CancelInvitation(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	err := h.invSvc.CancelInvitation(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation cancelled successfully",
	})
}

// AcceptInvitationByLink godoc
// @Summary Accept invitation by link
// @Tags invitations
// @Accept json
// @Produce json
// @Param request body AcceptLinkRequest true "Link details"
// @Success 200 {object} map[string]interface{}
// @Router /invitations/accept-link [post]
func (h *InvitationHandler) AcceptInvitationByLink(c *gin.Context) {
	var req AcceptLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")

	inv, settings, err := h.invSvc.UseLink(c.Request.Context(), req.LinkToken, userEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.invSvc.AcceptByID(c.Request.Context(), inv.ID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Invitation accepted successfully",
		"invitation": inv,
		"settings":   settings,
	})
}

// CreateLinkInvitation godoc
// @Summary Create link invitation
// @Tags invitations
// @Accept json
// @Produce json
// @Param request body CreateLinkInvitationRequest true "Link invitation details"
// @Success 201 {object} map[string]interface{}
// @Router /invitations/link [post]
func (h *InvitationHandler) CreateLinkInvitation(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateLinkInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings := &repository.InvitationLinkSettings{
		WorkspaceID:       req.WorkspaceID,
		Type:              repository.InvitationType(req.Type),
		TargetID:          req.TargetID,
		DefaultRole:       repository.WorkspaceRole(req.DefaultRole),
		DefaultPermission: repository.PermissionLevel(req.DefaultPermission),
		IsActive:          true,
		RequiresApproval:  req.RequiresApproval,
		AllowedDomains:    req.AllowedDomains,
		BlockedDomains:    req.BlockedDomains,
		MaxUses:           req.MaxUses,
		ExpiresAt:         req.ExpiresAt,
		CreatedByID:       userID,
	}

	err := h.invSvc.CreateLinkSettings(c.Request.Context(), settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Link invitation created successfully",
		"settings": settings,
	})
}

// GetLinkInvitation godoc
// @Summary Get link invitation by token
// @Tags invitations
// @Produce json
// @Param token path string true "Link Token"
// @Success 200 {object} map[string]interface{}
// @Router /invitations/link/{token} [get]
func (h *InvitationHandler) GetLinkInvitation(c *gin.Context) {
	token := c.Param("token")

	settings, err := h.invSvc.GetLinkSettingsByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	if !settings.IsValid() {
		c.JSON(http.StatusGone, gin.H{"error": "Link is expired or inactive"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})
}

// GetInvitationStats godoc
// @Summary Get invitation statistics
// @Tags invitations
// @Produce json
// @Param workspaceId query string true "Workspace ID"
// @Success 200 {object} map[string]interface{}
// @Router /invitations/stats [get]
func (h *InvitationHandler) GetInvitationStats(c *gin.Context) {
	workspaceID := c.Query("workspaceId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspaceId is required"})
		return
	}

	stats, err := h.invSvc.GetStatsByWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

type CreateInvitationRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Role        string `json:"role" binding:"required"`
	WorkspaceID string `json:"workspace_id,omitempty"`
}

type AcceptLinkRequest struct {
	LinkToken string `json:"link_token" binding:"required"`
}

type CreateLinkInvitationRequest struct {
	WorkspaceID       string     `json:"workspace_id" binding:"required"`
	Type              string     `json:"type" binding:"required"`
	TargetID          string     `json:"target_id" binding:"required"`
	DefaultRole       string     `json:"default_role" binding:"required"`
	DefaultPermission string     `json:"default_permission"`
	RequiresApproval  bool       `json:"requires_approval"`
	AllowedDomains    *string    `json:"allowed_domains,omitempty"`
	BlockedDomains    *string    `json:"blocked_domains,omitempty"`
	MaxUses           *int       `json:"max_uses,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}