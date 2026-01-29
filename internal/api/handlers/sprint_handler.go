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

	if err := h.sprintService.StartSprint(c.Request.Context(), sprintID, userID); err != nil {
		log.Printf("❌ [Sprint Start] Failed - SprintID: %s, Error: %v", sprintID, err)
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sprint started"})
}


func (h *SprintHandler) Complete(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("id")
	log.Printf("📝 [Sprint Complete] Completing sprint - SprintID: %s, UserID: %s", sprintID, userID)

	// 1. Complete the sprint (update status)
	if err := h.sprintService.CompleteSprint(c.Request.Context(), sprintID, userID); err != nil {
		log.Printf("❌ [Sprint Complete] Failed - SprintID: %s, Error: %v", sprintID, err)
		handleServiceError(c, err)
		return
	}

	// 2. ✅ RECORD VELOCITY - THIS WAS MISSING!
	if err := h.analyticsService.RecordSprintVelocity(c.Request.Context(), sprintID); err != nil {
		log.Printf("⚠️ [Sprint Complete] Failed to record velocity - SprintID: %s, Error: %v", sprintID, err)
		// Don't fail the request, just log the error
	} else {
		log.Printf("✅ [Sprint Complete] Velocity recorded - SprintID: %s", sprintID)
	}

	log.Printf("✅ [Sprint Complete] Success - SprintID: %s", sprintID)
	c.JSON(http.StatusOK, gin.H{"message": "Sprint completed"})
}