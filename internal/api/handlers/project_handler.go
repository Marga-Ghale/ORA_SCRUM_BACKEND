package handlers

import (
	"log"
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
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
func (h *ProjectHandler) ListBySpace(c *gin.Context) {
	spaceID := c.Param("id")

	projects, err := h.projectService.ListBySpace(c.Request.Context(), spaceID)
	if err != nil {
		log.Printf("[ProjectHandler][ListBySpace] spaceID=%s error=%v", spaceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	response := make([]models.ProjectResponse, len(projects))
	for i, p := range projects {
		response[i] = toProjectResponse(p)
	}

	c.JSON(http.StatusOK, response)
}

// ListByFolder - List projects in a folder
func (h *ProjectHandler) ListByFolder(c *gin.Context) {
	folderID := c.Param("id")

	projects, err := h.projectService.ListByFolder(c.Request.Context(), folderID)
	if err != nil {
		log.Printf("[ProjectHandler][ListByFolder] folderID=%s error=%v", folderID, err)
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
func (h *ProjectHandler) Create(c *gin.Context) {
	spaceID := c.Param("id")

	userID, ok := middleware.RequireUserID(c)
	if !ok {
		log.Printf("[ProjectHandler][Create] missing userID")
		return
	}

	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ProjectHandler][Create] invalid payload error=%v", err)
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
		log.Printf(
			"[ProjectHandler][Create] spaceID=%s userID=%s error=%v",
			spaceID,
			userID,
			err,
		)

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
func (h *ProjectHandler) Get(c *gin.Context) {
	id := c.Param("id")

	project, err := h.projectService.GetByID(c.Request.Context(), id)
	if err != nil {
		log.Printf("[ProjectHandler][Get] projectID=%s error=%v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, toProjectResponse(project))
}

// Update - Update a project
func (h *ProjectHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ProjectHandler][Update] projectID=%s invalid payload error=%v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ✅ Handle double pointer correctly
	var folderIDUpdate *string
	if req.FolderID != nil {
		folderIDUpdate = *req.FolderID
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
		folderIDUpdate,  // ✅ Use converted value
	)
	if err != nil {
		log.Printf("[ProjectHandler][Update] projectID=%s error=%v", id, err)

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
func (h *ProjectHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.projectService.Delete(c.Request.Context(), id); err != nil {
		log.Printf("[ProjectHandler][Delete] projectID=%s error=%v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.Status(http.StatusNoContent)
}


// ============================================
// Helper Functions
// ============================================
func toProjectResponse(p *repository.Project) models.ProjectResponse {
	return models.ProjectResponse{
		ID:          p.ID,
		SpaceID:     p.SpaceID,
		FolderID:    p.FolderID,
		Name:        p.Name,
		Key:         p.Key,
		Description: p.Description,
		Icon:        p.Icon,
		Color:       p.Color,
		LeadID:      p.LeadID,
		Visibility:  p.Visibility,
		AllowedUsers: func() []string {
			if p.AllowedUsers != nil {
				return p.AllowedUsers
			}
			return []string{}
		}(),
		AllowedTeams: func() []string {
			if p.AllowedTeams != nil {
				return p.AllowedTeams
			}
			return []string{}
		}(),
		CreatedBy:   p.CreatedBy,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}