package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// Workspace Handler
// ============================================

type WorkspaceHandler struct {
	workspaceService service.WorkspaceService
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	workspaces, err := h.workspaceService.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workspaces"})
		return
	}

	response := make([]models.WorkspaceResponse, len(workspaces))
	for i, ws := range workspaces {
		response[i] = toWorkspaceResponse(ws)
	}

	c.JSON(http.StatusOK, response)
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.workspaceService.Create(c.Request.Context(), userID, req.Name, req.Description, req.Icon, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workspace"})
		return
	}

	c.JSON(http.StatusCreated, toWorkspaceResponse(workspace))
}

func (h *WorkspaceHandler) Get(c *gin.Context) {
	id := c.Param("id")

	workspace, err := h.workspaceService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
		return
	}

	c.JSON(http.StatusOK, toWorkspaceResponse(workspace))
}

func (h *WorkspaceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.workspaceService.Update(c.Request.Context(), id, req.Name, req.Description, req.Icon, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workspace"})
		return
	}

	c.JSON(http.StatusOK, toWorkspaceResponse(workspace))
}

func (h *WorkspaceHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.workspaceService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workspace"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	id := c.Param("id")

	members, err := h.workspaceService.ListMembers(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	response := make([]models.WorkspaceMemberResponse, len(members))
	for i, m := range members {
		response[i] = toWorkspaceMemberResponse(m)
	}

	c.JSON(http.StatusOK, response)
}

func (h *WorkspaceHandler) AddMember(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pass inviterID for notification
	if err := h.workspaceService.AddMember(c.Request.Context(), id, req.Email, req.Role, userID); err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "User is already a member"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

func (h *WorkspaceHandler) UpdateMemberRole(c *gin.Context) {
	id := c.Param("id")
	memberUserID := c.Param("userId")

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.workspaceService.UpdateMemberRole(c.Request.Context(), id, memberUserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	id := c.Param("id")
	memberUserID := c.Param("userId")

	if err := h.workspaceService.RemoveMember(c.Request.Context(), id, memberUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddMemberByID adds an existing user to workspace by their user ID
func (h *WorkspaceHandler) AddMemberByID(c *gin.Context) {
	workspaceID := c.Param("id")
	inviterID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req struct {
		UserID string `json:"userId" binding:"required"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	if err := h.workspaceService.AddMemberByID(c.Request.Context(), workspaceID, req.UserID, req.Role, inviterID); err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "User is already a member"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}
