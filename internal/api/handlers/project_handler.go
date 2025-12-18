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

func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// ListBySpace - List projects in a space
// GET /spaces/:id/projects
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

// Create - Create a new project
// POST /spaces/:id/projects
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

	project, err := h.projectService.Create(
		c.Request.Context(),
		spaceID,
		req.FolderID,
		userID,
		req.Name,
		req.Key,
		req.Description,
		req.Icon,
		req.Color,
		req.LeadID,
	)
	if err != nil {
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Project key already exists"})
			return
		}
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Space or folder not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, toProjectResponse(project))
}

// Get - Get a project by ID
// GET /projects/:id
func (h *ProjectHandler) Get(c *gin.Context) {
	id := c.Param("id")

	project, err := h.projectService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, toProjectResponse(project))
}

// Update - Update a project
// PUT /projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.Update(
		c.Request.Context(),
		id,
		req.Name,
		req.Key,
		req.Description,
		req.Icon,
		req.Color,
		req.LeadID,
		*req.FolderID,
	)
	if err != nil {
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Project key already exists"})
			return
		}
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, toProjectResponse(project))
}

// Delete - Delete a project
// DELETE /projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.projectService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.Status(http.StatusNoContent)
}
