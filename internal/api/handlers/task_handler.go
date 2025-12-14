package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// Task Handler
// ============================================

type TaskHandler struct {
	taskService service.TaskService
}

func (h *TaskHandler) ListByProject(c *gin.Context) {
	projectID := c.Param("id")

	var filters repository.TaskFilters
	c.ShouldBindQuery(&filters)

	tasks, err := h.taskService.ListByProject(c.Request.Context(), projectID, &filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	response := make([]models.TaskResponse, len(tasks))
	for i, t := range tasks {
		response[i] = toTaskResponse(t)
	}

	c.JSON(http.StatusOK, response)
}

func (h *TaskHandler) ListBySprint(c *gin.Context) {
	sprintID := c.Param("id")

	tasks, err := h.taskService.ListBySprint(c.Request.Context(), sprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	response := make([]models.TaskResponse, len(tasks))
	for i, t := range tasks {
		response[i] = toTaskResponse(t)
	}

	c.JSON(http.StatusOK, response)
}

func (h *TaskHandler) Create(c *gin.Context) {
	projectID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskService.Create(
		c.Request.Context(),
		projectID,
		userID,
		req.Title,
		req.Description,
		req.Status,
		req.Priority,
		req.Type,
		req.AssigneeID,
		req.SprintID,
		req.ParentID,
		req.StoryPoints,
		req.DueDate,
		req.Labels,
	)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, toTaskResponse(task))
}

func (h *TaskHandler) Get(c *gin.Context) {
	id := c.Param("id")

	task, err := h.taskService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.AssigneeID != nil {
		updates["assigneeId"] = req.AssigneeID
	}
	if req.SprintID != nil {
		updates["sprintId"] = req.SprintID
	}
	if req.StoryPoints != nil {
		updates["storyPoints"] = req.StoryPoints
	}
	if req.DueDate != nil {
		updates["dueDate"] = req.DueDate
	}
	if req.OrderIndex != nil {
		updates["orderIndex"] = *req.OrderIndex
	}
	if req.Labels != nil {
		updates["labels"] = req.Labels
	}

	// Pass userID for notifications
	task, err := h.taskService.Update(c.Request.Context(), id, userID, updates)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) PartialUpdate(c *gin.Context) {
	h.Update(c) // Same logic for PATCH
}

func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	// Pass userID for notifications
	if err := h.taskService.Delete(c.Request.Context(), id, userID); err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *TaskHandler) BulkUpdate(c *gin.Context) {
	var req models.BulkUpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make([]repository.BulkTaskUpdate, len(req.Tasks))
	for i, t := range req.Tasks {
		updates[i] = repository.BulkTaskUpdate{
			ID:         t.ID,
			Status:     t.Status,
			SprintID:   t.SprintID,
			OrderIndex: t.OrderIndex,
		}
	}

	if err := h.taskService.BulkUpdate(c.Request.Context(), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk update tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tasks updated successfully"})
}
