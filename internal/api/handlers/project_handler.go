package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// Project Handler
// ============================================

type ProjectHandler struct {
	projectService service.ProjectService
}

func (h *ProjectHandler) ListBySpace(c *gin.Context) {
	spaceID := c.Param("id")

	projects, err := h.projectService.ListBySpace(c.Request.Context(), spaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	response := make([]models.ProjectResponse, len(projects))
	for i, p := range projects {
		response[i] = toProjectResponse(p)
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	spaceID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pass creatorID for auto-adding as member
	project, err := h.projectService.Create(c.Request.Context(), spaceID, userID, req.Name, req.Key, req.Description, req.Icon, req.Color, req.LeadID)
	if err != nil {
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Project key already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, toProjectResponse(project))
}

func (h *ProjectHandler) Get(c *gin.Context) {
	id := c.Param("id")

	project, err := h.projectService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, toProjectResponse(project))
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.Update(c.Request.Context(), id, req.Name, req.Key, req.Description, req.Icon, req.Color, req.LeadID)
	if err != nil {
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Project key already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, toProjectResponse(project))
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.projectService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ProjectHandler) ListMembers(c *gin.Context) {
	id := c.Param("id")

	members, err := h.projectService.ListMembers(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	response := make([]models.ProjectMemberResponse, len(members))
	for i, m := range members {
		response[i] = toProjectMemberResponse(m)
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProjectHandler) AddMember(c *gin.Context) {
	id := c.Param("id")
	inviterID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.AddProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pass inviterID for notification
	if err := h.projectService.AddMember(c.Request.Context(), id, req.UserID, req.Role, inviterID); err != nil {
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "User is already a member"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	id := c.Param("id")
	memberUserID := c.Param("userId")

	if err := h.projectService.RemoveMember(c.Request.Context(), id, memberUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddMemberByID adds an existing user to project by their user ID
func (h *ProjectHandler) AddMemberByID(c *gin.Context) {
	projectID := c.Param("id")
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

	if err := h.projectService.AddMember(c.Request.Context(), projectID, req.UserID, req.Role, inviterID); err != nil {
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "User is already a member"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}
