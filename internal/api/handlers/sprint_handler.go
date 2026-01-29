package handlers

import (
	"log"
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SprintHandler struct {
	sprintService service.SprintService
	analyticsService service.SprintAnalyticsService  

}

func NewSprintHandler(
	sprintService service.SprintService,
	analyticsService service.SprintAnalyticsService,  // ✅ ADD THIS
) *SprintHandler {
	return &SprintHandler{
		sprintService:    sprintService,
		analyticsService: analyticsService,  // ✅ ADD THIS
	}
}

func (h *SprintHandler) Create(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var sprint repository.Sprint
	if err := c.ShouldBindJSON(&sprint); err != nil {
		log.Printf("❌ [Sprint Create] JSON binding failed - Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sprint.ProjectID = c.Param("id")
	
	if sprint.Status == "" {
		sprint.Status = "planning"
	}

	log.Printf("📝 [Sprint Create] Creating sprint - Name: %s, ProjectID: %s, UserID: %s", 
		sprint.Name, sprint.ProjectID, userID)

	if err := h.sprintService.Create(c.Request.Context(), &sprint, userID); err != nil {
		log.Printf("❌ [Sprint Create] Failed - Error: %v", err)
		handleServiceError(c, err)
		return
	}

	log.Printf("✅ [Sprint Create] Success - SprintID: %s", sprint.ID)
	c.JSON(http.StatusCreated, sprint)
}

func (h *SprintHandler) Get(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("id")
	log.Printf("📝 [Sprint Get] Fetching sprint - SprintID: %s, UserID: %s", sprintID, userID)

	sprint, err := h.sprintService.Get(c.Request.Context(), sprintID, userID)
	if err != nil {
		log.Printf("❌ [Sprint Get] Failed - SprintID: %s, Error: %v", sprintID, err)
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, sprint)
}

func (h *SprintHandler) ListByProject(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	log.Printf("📝 [Sprint List] Fetching sprints - ProjectID: %s, UserID: %s", projectID, userID)

	sprints, err := h.sprintService.ListByProject(c.Request.Context(), projectID, userID)
	if err != nil {
		log.Printf("❌ [Sprint List] Failed - ProjectID: %s, Error: %v", projectID, err)
		handleServiceError(c, err)
		return
	}

	log.Printf("✅ [Sprint List] Found %d sprints", len(sprints))
	c.JSON(http.StatusOK, sprints)
}

func (h *SprintHandler) GetActive(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	log.Printf("📝 [Sprint GetActive] Fetching active sprint - ProjectID: %s, UserID: %s", projectID, userID)

	sprint, err := h.sprintService.GetActiveSprint(c.Request.Context(), projectID, userID)
	if err != nil {
		log.Printf("❌ [Sprint GetActive] Failed - ProjectID: %s, Error: %v", projectID, err)
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, sprint)
}

func (h *SprintHandler) Update(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var sprint repository.Sprint
	if err := c.ShouldBindJSON(&sprint); err != nil {
		log.Printf("❌ [Sprint Update] JSON binding failed - Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sprint.ID = c.Param("id")
	log.Printf("📝 [Sprint Update] Updating sprint - SprintID: %s, UserID: %s", sprint.ID, userID)

	if err := h.sprintService.Update(c.Request.Context(), &sprint, userID); err != nil {
		log.Printf("❌ [Sprint Update] Failed - SprintID: %s, Error: %v", sprint.ID, err)
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, sprint)
}

func (h *SprintHandler) Delete(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("id")
	log.Printf("📝 [Sprint Delete] Deleting sprint - SprintID: %s, UserID: %s", sprintID, userID)

	if err := h.sprintService.Delete(c.Request.Context(), sprintID, userID); err != nil {
		log.Printf("❌ [Sprint Delete] Failed - SprintID: %s, Error: %v", sprintID, err)
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *SprintHandler) Start(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("id")
	log.Printf("📝 [Sprint Start] Starting sprint - SprintID: %s, UserID: %s", sprintID, userID)

	response, err := h.sprintService.StartSprint(c.Request.Context(), sprintID, userID)
	if err != nil {
		log.Printf("❌ [Sprint Start] Failed - SprintID: %s, Error: %v", sprintID, err)
		handleServiceError(c, err)
		return
	}

	log.Printf("✅ [Sprint Start] Success - SprintID: %s, Committed: %d tasks, %d points", 
		sprintID, response.CommittedTasks, response.CommittedPoints)

	c.JSON(http.StatusOK, response)
}


func (h *SprintHandler) Complete(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("id")
	
	// Parse options from request body (optional)
	var options service.SprintCompleteOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		// Default to backlog if no options provided
		options.MoveIncompleteTo = "backlog"
	}

	log.Printf("📝 [Sprint Complete] Completing sprint - SprintID: %s, MoveIncompleteTo: %s", sprintID, options.MoveIncompleteTo)

	response, err := h.sprintService.CompleteSprintWithOptions(c.Request.Context(), sprintID, userID, &options)
	if err != nil {
		log.Printf("❌ [Sprint Complete] Failed - SprintID: %s, Error: %v", sprintID, err)
		handleServiceError(c, err)
		return
	}

	// Record velocity
	if err := h.analyticsService.RecordSprintVelocity(c.Request.Context(), sprintID); err != nil {
		log.Printf("⚠️ [Sprint Complete] Failed to record velocity - SprintID: %s, Error: %v", sprintID, err)
	}

	log.Printf("✅ [Sprint Complete] Success - Completed: %d/%d tasks, Moved %d incomplete tasks to %s",
		response.CompletedTasks, response.CompletedTasks+response.IncompleteTasks,
		response.IncompleteTasks, response.TasksMovedTo)

	c.JSON(http.StatusOK, response)
}

// Add new endpoint for sprint summary
func (h *SprintHandler) GetSummary(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("id")
	summary, err := h.sprintService.GetSprintSummary(c.Request.Context(), sprintID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, summary)
}
// POST /api/sprints/:id/complete-with-options
func (h *SprintHandler) CompleteWithOptions(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("id")

	var req struct {
		MoveIncompleteTo string `json:"moveIncompleteTo"` // "backlog", "next_sprint", or sprint ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.MoveIncompleteTo = "backlog" // Default
	}

	log.Printf("📝 [Sprint Complete] SprintID: %s, MoveIncompleteTo: %s", sprintID, req.MoveIncompleteTo)

	response, err := h.sprintService.CompleteSprintWithOptions(c.Request.Context(), sprintID, userID, &service.SprintCompleteOptions{
		MoveIncompleteTo: req.MoveIncompleteTo,
	})
	if err != nil {
		log.Printf("❌ [Sprint Complete] Failed - SprintID: %s, Error: %v", sprintID, err)
		handleServiceError(c, err)
		return
	}

	// Record velocity
	if err := h.analyticsService.RecordSprintVelocity(c.Request.Context(), sprintID); err != nil {
		log.Printf("⚠️ Failed to record velocity: %v", err)
	}

	log.Printf("✅ [Sprint Complete] Success - Completed: %d tasks (%d pts), Incomplete: %d tasks (%d pts), Moved to: %s",
		response.CompletedTasks,
		response.CompletedPoints,
		response.IncompleteTasks,
		response.IncompletePoints,
		response.TasksMovedTo,
	)

	c.JSON(http.StatusOK, response)
}