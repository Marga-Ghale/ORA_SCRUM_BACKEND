package handlers

import (
	"net/http"
	"strconv"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SprintAnalyticsHandler struct {
	analyticsService service.SprintAnalyticsService
}

func NewSprintAnalyticsHandler(analyticsService service.SprintAnalyticsService) *SprintAnalyticsHandler {
	return &SprintAnalyticsHandler{analyticsService: analyticsService}
}

// ============================================
// SPRINT REPORTS
// ============================================

// GET /api/sprints/:sprintId/report
func (h *SprintAnalyticsHandler) GetSprintReport(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("sprintId")
	report, err := h.analyticsService.GetSprintReport(c.Request.Context(), sprintID, userID)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

// POST /api/sprints/:sprintId/report/generate
func (h *SprintAnalyticsHandler) GenerateSprintReport(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("sprintId")
	report, err := h.analyticsService.GenerateSprintReport(c.Request.Context(), sprintID, userID)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

// ============================================
// VELOCITY
// ============================================

// GET /api/projects/:id/velocity
func (h *SprintAnalyticsHandler) GetVelocityHistory(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	history, err := h.analyticsService.GetVelocityHistory(c.Request.Context(), projectID, userID, limit)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projectId": projectID,
		"history":   history,
	})
}

// GET /api/projects/:id/velocity/trend
func (h *SprintAnalyticsHandler) GetVelocityTrend(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	sprintCountStr := c.DefaultQuery("sprints", "6")
	sprintCount, _ := strconv.Atoi(sprintCountStr)

	trend, err := h.analyticsService.GetVelocityTrend(c.Request.Context(), projectID, userID, sprintCount)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, trend)
}

// ============================================
// CYCLE TIME
// ============================================

// GET /api/sprints/:sprintId/cycle-time
func (h *SprintAnalyticsHandler) GetSprintCycleTime(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("sprintId")
	stats, err := h.analyticsService.GetCycleTimeStats(c.Request.Context(), sprintID, userID)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sprintId": sprintID,
		"tasks":    stats,
	})
}

// GET /api/projects/:id/cycle-time
func (h *SprintAnalyticsHandler) GetProjectCycleTime(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)

	avg, err := h.analyticsService.GetProjectCycleTimeAvg(c.Request.Context(), projectID, userID, days)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, avg)
}

// GET /api/tasks/:id/status-history
func (h *SprintAnalyticsHandler) GetTaskStatusHistory(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	history, err := h.analyticsService.GetTaskStatusHistory(c.Request.Context(), taskID, userID)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"taskId":  taskID,
		"history": history,
	})
}

// ============================================
// GANTT CHART
// ============================================

// GET /api/projects/:id/gantt
func (h *SprintAnalyticsHandler) GetGanttData(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	sprintID := c.Query("sprintId")

	var sprintIDPtr *string
	if sprintID != "" {
		sprintIDPtr = &sprintID
	}

	data, err := h.analyticsService.GetGanttData(c.Request.Context(), projectID, userID, sprintIDPtr)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, data)
}

// ============================================
// DASHBOARDS
// ============================================

// GET /api/sprints/:sprintId/analytics
func (h *SprintAnalyticsHandler) GetSprintAnalyticsDashboard(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("sprintId")
	dashboard, err := h.analyticsService.GetSprintAnalyticsDashboard(c.Request.Context(), sprintID, userID)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GET /api/projects/:id/analytics
func (h *SprintAnalyticsHandler) GetProjectAnalyticsDashboard(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	dashboard, err := h.analyticsService.GetProjectAnalyticsDashboard(c.Request.Context(), projectID, userID)
	if err != nil {
		handleAnalyticsError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// ============================================
// ERROR HANDLER
// ============================================

func handleAnalyticsError(c *gin.Context, err error) {
	switch err {
	case service.ErrUnauthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
	case service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
	case service.ErrInvalidInput:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}