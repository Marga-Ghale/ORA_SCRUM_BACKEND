package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// Sprint Handler
// ============================================

type SprintHandler struct {
	sprintService service.SprintService
}

func (h *SprintHandler) ListByProject(c *gin.Context) {
	projectID := c.Param("id")

	sprints, err := h.sprintService.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sprints"})
		return
	}

	response := make([]models.SprintResponse, len(sprints))
	for i, s := range sprints {
		response[i] = toSprintResponse(s)
	}

	c.JSON(http.StatusOK, response)
}

func (h *SprintHandler) Create(c *gin.Context) {
	projectID := c.Param("id")

	var req models.CreateSprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sprint, err := h.sprintService.Create(c.Request.Context(), projectID, req.Name, req.Goal, req.StartDate, req.EndDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sprint"})
		return
	}

	c.JSON(http.StatusCreated, toSprintResponse(sprint))
}

func (h *SprintHandler) Get(c *gin.Context) {
	id := c.Param("id")

	sprint, err := h.sprintService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}

	c.JSON(http.StatusOK, toSprintResponse(sprint))
}

func (h *SprintHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateSprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sprint, err := h.sprintService.Update(c.Request.Context(), id, req.Name, req.Goal, req.StartDate, req.EndDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sprint"})
		return
	}

	c.JSON(http.StatusOK, toSprintResponse(sprint))
}

func (h *SprintHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.sprintService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sprint"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *SprintHandler) Start(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	// Pass userID for notifications
	sprint, err := h.sprintService.Start(c.Request.Context(), id, userID)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toSprintResponse(sprint))
}

func (h *SprintHandler) Complete(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.CompleteSprintRequest
	c.ShouldBindJSON(&req)

	// Pass userID for notifications
	sprint, err := h.sprintService.Complete(c.Request.Context(), id, req.MoveIncomplete, userID)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete sprint"})
		return
	}

	c.JSON(http.StatusOK, toSprintResponse(sprint))
}
