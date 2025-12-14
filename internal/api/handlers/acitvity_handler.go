package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// Activity Handler
// ============================================

// ActivityHandler handles activity-related HTTP requests
type ActivityHandler struct {
	activitySvc service.ActivityService
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(activitySvc service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activitySvc: activitySvc}
}

// GetTaskActivities gets activities for a task
func (h *ActivityHandler) GetTaskActivities(c *gin.Context) {
	taskID := c.Param("id")

	activities, err := h.activitySvc.GetEntityActivities(c.Request.Context(), "task", taskID, 50)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, activities)
}

// GetProjectActivities gets activities for a project
func (h *ActivityHandler) GetProjectActivities(c *gin.Context) {
	projectID := c.Param("id")

	activities, err := h.activitySvc.GetEntityActivities(c.Request.Context(), "project", projectID, 50)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, activities)
}

// GetMyActivities gets the current user's activities
func (h *ActivityHandler) GetMyActivities(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	activities, err := h.activitySvc.GetUserActivities(c.Request.Context(), userID, 50)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, activities)
}
