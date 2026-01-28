package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type GoalHandler struct {
	goalService service.GoalService
}

func NewGoalHandler(goalService service.GoalService) *GoalHandler {
	return &GoalHandler{goalService: goalService}
}

// ============================================
// GOAL CRUD
// ============================================

func (h *GoalHandler) Create(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req service.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.CreatedBy = userID

	goal, err := h.goalService.Create(c.Request.Context(), &req)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, goal)
}

func (h *GoalHandler) Get(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	goalID := c.Param("id")
	goal, err := h.goalService.GetByID(c.Request.Context(), goalID, userID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (h *GoalHandler) ListByWorkspace(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	workspaceID := c.Param("id")
	goalType := c.Query("type")
	status := c.Query("status")

	var goalTypePtr, statusPtr *string
	if goalType != "" {
		goalTypePtr = &goalType
	}
	if status != "" {
		statusPtr = &status
	}

	goals, err := h.goalService.ListByWorkspace(c.Request.Context(), workspaceID, userID, goalTypePtr, statusPtr)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, goals)
}

func (h *GoalHandler) ListByProject(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	goals, err := h.goalService.ListByProject(c.Request.Context(), projectID, userID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, goals)
}

func (h *GoalHandler) ListBySprint(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("sprintId")
	goals, err := h.goalService.ListBySprint(c.Request.Context(), sprintID, userID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, goals)
}

func (h *GoalHandler) Update(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	goalID := c.Param("id")
	var req service.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal, err := h.goalService.Update(c.Request.Context(), goalID, userID, &req)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (h *GoalHandler) UpdateProgress(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	goalID := c.Param("id")
	var req struct {
		CurrentValue float64 `json:"currentValue" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.goalService.UpdateProgress(c.Request.Context(), goalID, userID, req.CurrentValue)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress updated"})
}

func (h *GoalHandler) UpdateStatus(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	goalID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.goalService.UpdateStatus(c.Request.Context(), goalID, userID, req.Status)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

func (h *GoalHandler) Delete(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	goalID := c.Param("id")
	err := h.goalService.Delete(c.Request.Context(), goalID, userID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// KEY RESULTS
// ============================================

func (h *GoalHandler) AddKeyResult(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	goalID := c.Param("id")
	var req service.CreateKeyResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kr, err := h.goalService.AddKeyResult(c.Request.Context(), goalID, userID, &req)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, kr)
}

func (h *GoalHandler) UpdateKeyResult(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	keyResultID := c.Param("krId")
	var req service.UpdateKeyResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.goalService.UpdateKeyResult(c.Request.Context(), keyResultID, userID, &req)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key result updated"})
}

func (h *GoalHandler) UpdateKeyResultProgress(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	keyResultID := c.Param("krId")
	var req struct {
		CurrentValue float64 `json:"currentValue" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.goalService.UpdateKeyResultProgress(c.Request.Context(), keyResultID, userID, req.CurrentValue)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress updated"})
}

func (h *GoalHandler) DeleteKeyResult(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	keyResultID := c.Param("krId")
	err := h.goalService.DeleteKeyResult(c.Request.Context(), keyResultID, userID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// TASK LINKING
// ============================================

func (h *GoalHandler) LinkTask(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	goalID := c.Param("id")
	var req struct {
		TaskID string `json:"taskId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.goalService.LinkTask(c.Request.Context(), goalID, req.TaskID, userID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task linked"})
}

func (h *GoalHandler) UnlinkTask(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	goalID := c.Param("id")
	taskID := c.Param("taskId")

	err := h.goalService.UnlinkTask(c.Request.Context(), goalID, taskID, userID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task unlinked"})
}

func (h *GoalHandler) GetGoalsByTask(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("taskId")
	goals, err := h.goalService.GetGoalsByTask(c.Request.Context(), taskID, userID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, goals)
}

// ============================================
// ANALYTICS
// ============================================

func (h *GoalHandler) GetGoalProgress(c *gin.Context) {
	goalID := c.Param("id")

	progress, err := h.goalService.GetGoalProgress(c.Request.Context(), goalID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"goalId":   goalID,
		"progress": progress,
	})
}

func (h *GoalHandler) GetSprintGoalsSummary(c *gin.Context) {
	sprintID := c.Param("sprintId")

	summary, err := h.goalService.GetSprintGoalsSummary(c.Request.Context(), sprintID)
	if err != nil {
		handleGoalError(c, err)
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ============================================
// ERROR HANDLER
// ============================================

func handleGoalError(c *gin.Context, err error) {
	switch err {
	case service.ErrUnauthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
	case service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Goal not found"})
	case service.ErrInvalidInput:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}