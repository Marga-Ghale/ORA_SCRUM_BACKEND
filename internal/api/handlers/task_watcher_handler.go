package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// Task Watcher Handler
// ============================================

// TaskWatcherHandler handles task watcher-related HTTP requests
type TaskWatcherHandler struct {
	watcherSvc service.TaskWatcherService
}

// NewTaskWatcherHandler creates a new task watcher handler
func NewTaskWatcherHandler(watcherSvc service.TaskWatcherService) *TaskWatcherHandler {
	return &TaskWatcherHandler{watcherSvc: watcherSvc}
}

// WatchTask starts watching a task
func (h *TaskWatcherHandler) WatchTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.watcherSvc.Watch(c.Request.Context(), taskID, userID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "watching task"})
}

// UnwatchTask stops watching a task
func (h *TaskWatcherHandler) UnwatchTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.watcherSvc.Unwatch(c.Request.Context(), taskID, userID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetWatchers gets all watchers of a task
func (h *TaskWatcherHandler) GetWatchers(c *gin.Context) {
	taskID := c.Param("id")

	watchers, err := h.watcherSvc.GetWatchers(c.Request.Context(), taskID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"watchers": watchers})
}

// IsWatching checks if the current user is watching a task
func (h *TaskWatcherHandler) IsWatching(c *gin.Context) {
	taskID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	watching, err := h.watcherSvc.IsWatching(c.Request.Context(), taskID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"watching": watching})
}
