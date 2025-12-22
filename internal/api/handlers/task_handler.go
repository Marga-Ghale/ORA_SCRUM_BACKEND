package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func logAPIError(c *gin.Context, action string, err error, fields map[string]interface{}) {
	log.Printf(
		"[API_ERROR] action=%s method=%s path=%s userID=%v fields=%v err=%v",
		action,
		c.Request.Method,
		c.FullPath(),
		c.GetString("userID"),
		fields,
		err,
	)
}


// ============================================
// TASK CRUD
// ============================================

func (h *TaskHandler) Create(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	// ✅ Get projectID from URL
	projectID := c.Param("id")

	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createReq := &service.CreateTaskRequest{
		ProjectID:      projectID,        // ✅ From URL
		SprintID:       req.SprintID,
		ParentTaskID:   req.ParentTaskID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		Priority:       req.Priority,
		AssigneeIDs:    req.AssigneeIDs,
		LabelIDs:       req.LabelIDs,
		EstimatedHours: req.EstimatedHours,
		StoryPoints:    req.StoryPoints,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
		CreatedBy:      &userID,          // ✅ From authenticated user
	}

	task, err := h.taskService.Create(c.Request.Context(), createReq)
if err != nil {
	logAPIError(c, "Task.Create", err, map[string]interface{}{
		"projectID": projectID,
		"title":     req.Title,
	})
	handleServiceError(c, err)
	return
}

	c.JSON(http.StatusCreated, toTaskResponse(task))


	
}

func (h *TaskHandler) Get(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	task, err := h.taskService.GetByID(c.Request.Context(), taskID, userID)
if err != nil {
	logAPIError(c, "Task.Get", err, map[string]interface{}{
		"taskID": taskID,
	})
	handleServiceError(c, err)
	return
}


	c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Update(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &service.UpdateTaskRequest{
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		Priority:       req.Priority,
		SprintID:       req.SprintID,
		AssigneeIDs:    req.AssigneeIDs,
		LabelIDs:       req.LabelIDs,
		EstimatedHours: req.EstimatedHours,
		ActualHours:    req.ActualHours,
		StoryPoints:    req.StoryPoints,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
	}

	task, err := h.taskService.Update(c.Request.Context(), taskID, userID, updateReq)
if err != nil {
	logAPIError(c, "Task.Update", err, map[string]interface{}{
		"taskID": taskID,
	})
	handleServiceError(c, err)
	return
}


	c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Delete(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	err := h.taskService.Delete(c.Request.Context(), taskID, userID)
if err != nil {
	logAPIError(c, "Task.Delete", err, map[string]interface{}{
		"taskID": taskID,
	})
	handleServiceError(c, err)
	return
}


	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// TASK LISTING
// ============================================

func (h *TaskHandler) ListByProject(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	fmt.Printf("DEBUG: projectID=%s, userID=%s\n", projectID, userID) // ADD THIS
	
	tasks, err := h.taskService.ListByProject(c.Request.Context(), projectID, userID)
if err != nil {
	logAPIError(c, "Task.ListByProject", err, map[string]interface{}{
		"projectID": projectID,
	})
	handleServiceError(c, err)
	return
}


	c.JSON(http.StatusOK, toTaskResponseList(tasks))
}


func (h *TaskHandler) ListBySprint(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("sprintId")
	tasks, err := h.taskService.ListBySprint(c.Request.Context(), sprintID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTaskResponseList(tasks))
}

func (h *TaskHandler) ListSubtasks(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	subtasks, err := h.taskService.ListSubtasks(c.Request.Context(), taskID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTaskResponseList(subtasks))
}

func (h *TaskHandler) ListMyTasks(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	tasks, err := h.taskService.ListMyTasks(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, toTaskResponseList(tasks))
}

func (h *TaskHandler) ListByStatus(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	status := c.Query("status")

	tasks, err := h.taskService.ListByStatus(c.Request.Context(), projectID, status, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTaskResponseList(tasks))
}

// ============================================
// TASK OPERATIONS
// ============================================

func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.UpdateStatus(c.Request.Context(), taskID, req.Status, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

func (h *TaskHandler) UpdatePriority(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req struct {
		Priority string `json:"priority" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.UpdatePriority(c.Request.Context(), taskID, req.Priority, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Priority updated successfully"})
}

// func (h *TaskHandler) AssignTask(c *gin.Context) {
// 	userID, ok := middleware.RequireUserID(c)
// 	if !ok {
// 		return
// 	}

// 	taskID := c.Param("id")
// 	var req struct {
// 		AssigneeID string `json:"assigneeId" binding:"required"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	err := h.taskService.AssignTask(c.Request.Context(), taskID, req.AssigneeID, userID)
// 	if err != nil {
// 		handleServiceError(c, err)
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Task assigned successfully"})
// }



func (h *TaskHandler) AssignTask(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req struct {
		AssigneeID string `json:"assigneeId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.AssignTask(c.Request.Context(), taskID, req.AssigneeID, userID)
if err != nil {
	logAPIError(c, "Task.Assign", err, map[string]interface{}{
		"taskID":     taskID,
		"assigneeID": req.AssigneeID,
	})
	handleServiceError(c, err)
	return
}


	c.JSON(http.StatusOK, gin.H{"message": "Task assigned successfully"})
}

func (h *TaskHandler) UnassignTask(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	assigneeID := c.Param("assigneeId")

	err := h.taskService.UnassignTask(c.Request.Context(), taskID, assigneeID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignee removed successfully"})
}

func (h *TaskHandler) AddWatcher(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req struct {
		WatcherID string `json:"watcherId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.AddWatcher(c.Request.Context(), taskID, req.WatcherID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Watcher added successfully"})
}

func (h *TaskHandler) RemoveWatcher(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	watcherID := c.Param("watcherId")

	err := h.taskService.RemoveWatcher(c.Request.Context(), taskID, watcherID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Watcher removed successfully"})
}

func (h *TaskHandler) MarkComplete(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	err := h.taskService.MarkComplete(c.Request.Context(), taskID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task marked as complete"})
}

func (h *TaskHandler) MoveToSprint(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req struct {
		SprintID string `json:"sprintId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.MoveToSprint(c.Request.Context(), taskID, req.SprintID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task moved to sprint successfully"})
}

func (h *TaskHandler) ConvertToSubtask(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req struct {
		ParentTaskID string `json:"parentTaskId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.ConvertToSubtask(c.Request.Context(), taskID, req.ParentTaskID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task converted to subtask successfully"})
}

// ============================================
// COMMENTS
// ============================================

func (h *TaskHandler) AddComment(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.taskService.AddComment(c.Request.Context(), taskID, userID, req.Content, req.MentionedUsers)
	if err != nil {
	logAPIError(c, "Task.AddComment", err, map[string]interface{}{
		"taskID": taskID,
	})
	handleServiceError(c, err)
	return
}


	c.JSON(http.StatusCreated, toCommentResponse(comment))
}

func (h *TaskHandler) ListComments(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	comments, err := h.taskService.ListComments(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, toCommentResponseList(comments))
}

func (h *TaskHandler) UpdateComment(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	commentID := c.Param("commentId")
	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.UpdateComment(c.Request.Context(), commentID, userID, req.Content)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment updated successfully"})
}

func (h *TaskHandler) DeleteComment(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	commentID := c.Param("commentId")
	err := h.taskService.DeleteComment(c.Request.Context(), commentID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// ATTACHMENTS
// ============================================

func (h *TaskHandler) AddAttachment(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req models.CreateAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attachment, err := h.taskService.AddAttachment(c.Request.Context(), taskID, userID, req.Filename, req.FileURL, req.FileSize, req.MimeType)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toAttachmentResponse(attachment))
}

func (h *TaskHandler) ListAttachments(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	attachments, err := h.taskService.ListAttachments(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attachments"})
		return
	}

	c.JSON(http.StatusOK, toAttachmentResponseList(attachments))
}

func (h *TaskHandler) DeleteAttachment(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	attachmentID := c.Param("attachmentId")
	err := h.taskService.DeleteAttachment(c.Request.Context(), attachmentID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// TIME TRACKING
// ============================================

func (h *TaskHandler) StartTimer(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	entry, err := h.taskService.StartTimer(c.Request.Context(), taskID, userID)
if err != nil {
	logAPIError(c, "Task.StartTimer", err, map[string]interface{}{
		"taskID": taskID,
	})
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start timer"})
	return
}


	c.JSON(http.StatusOK, toTimeEntryResponse(entry))
}

func (h *TaskHandler) StopTimer(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	entry, err := h.taskService.StopTimer(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active timer"})
		return
	}

	c.JSON(http.StatusOK, toTimeEntryResponse(entry))
}

func (h *TaskHandler) GetActiveTimer(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	entry, err := h.taskService.GetActiveTimer(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active timer"})
		return
	}

	c.JSON(http.StatusOK, toTimeEntryResponse(entry))
}

func (h *TaskHandler) LogTime(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req models.LogTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.taskService.LogTime(c.Request.Context(), taskID, userID, req.DurationSeconds, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log time"})
		return
	}

	c.JSON(http.StatusCreated, toTimeEntryResponse(entry))
}

func (h *TaskHandler) GetTimeEntries(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	entries, err := h.taskService.GetTimeEntries(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch time entries"})
		return
	}

	c.JSON(http.StatusOK, toTimeEntryResponseList(entries))
}

func (h *TaskHandler) GetTotalTime(c *gin.Context) {
	taskID := c.Param("id")
	totalSeconds, err := h.taskService.GetTotalTime(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate total time"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"taskId":       taskID,
		"totalSeconds": totalSeconds,
		"totalHours":   float64(totalSeconds) / 3600.0,
	})
}

// ============================================
// DEPENDENCIES
// ============================================

func (h *TaskHandler) AddDependency(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req models.CreateDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.AddDependency(c.Request.Context(), taskID, req.DependsOnTaskID, req.DependencyType, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Dependency added successfully"})
}

func (h *TaskHandler) RemoveDependency(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	dependsOnTaskID := c.Param("dependsOnTaskId")

	err := h.taskService.RemoveDependency(c.Request.Context(), taskID, dependsOnTaskID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dependency removed successfully"})
}

func (h *TaskHandler) ListDependencies(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	deps, err := h.taskService.ListDependencies(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dependencies"})
		return
	}

	c.JSON(http.StatusOK, toDependencyResponseList(deps))
}

func (h *TaskHandler) ListBlockedBy(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	deps, err := h.taskService.ListBlockedBy(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blocked by tasks"})
		return
	}

	c.JSON(http.StatusOK, toDependencyResponseList(deps))
}

// ============================================
// CHECKLISTS
// ============================================

func (h *TaskHandler) CreateChecklist(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	var req models.CreateChecklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checklist, err := h.taskService.CreateChecklist(c.Request.Context(), taskID, userID, req.Title)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toChecklistResponse(checklist))
}

func (h *TaskHandler) AddChecklistItem(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	checklistID := c.Param("checklistId")
	var req models.CreateChecklistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.taskService.AddChecklistItem(c.Request.Context(), checklistID, userID, req.Content, req.AssigneeID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toChecklistItemResponse(item))
}

func (h *TaskHandler) ToggleChecklistItem(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	itemID := c.Param("itemId")
	err := h.taskService.ToggleChecklistItem(c.Request.Context(), itemID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Checklist item toggled successfully"})
}

func (h *TaskHandler) DeleteChecklistItem(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	itemID := c.Param("itemId")
	err := h.taskService.DeleteChecklistItem(c.Request.Context(), itemID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *TaskHandler) ListChecklists(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	checklists, err := h.taskService.ListChecklists(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch checklists"})
		return
	}

	c.JSON(http.StatusOK, toChecklistResponseList(checklists))
}

// ============================================
// ACTIVITY
// ============================================

func (h *TaskHandler) GetActivity(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	activities, err := h.taskService.GetActivity(c.Request.Context(), taskID, userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activity"})
		return
	}

	c.JSON(http.StatusOK, toActivityResponseList(activities))
}

// ============================================
// ADVANCED FILTERING
// ============================================
func (h *TaskHandler) FilterTasks(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.TaskFiltersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters := &repository.TaskFilters{
		ProjectID:   req.ProjectID,
		SprintID:    req.SprintID,
		AssigneeIDs: req.AssigneeIDs,
		Status:      req.Statuses,    // []string matches
		Priority:    req.Priorities,  // []string matches
		LabelIDs:    req.LabelIDs,
		Search:      req.SearchQuery, // *string matches
		DueBefore:   req.DueBefore,
		DueAfter:    req.DueAfter,
		Overdue:     req.Overdue,
		Blocked:     req.Blocked,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	tasks, total, err := h.taskService.FilterTasks(c.Request.Context(), filters, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to filter tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":  toTaskResponseList(tasks),
		"total":  total,
		"limit":  filters.Limit,
		"offset": filters.Offset,
	})
}

func (h *TaskHandler) FindOverdue(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	tasks, err := h.taskService.FindOverdue(c.Request.Context(), projectID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTaskResponseList(tasks))
}

func (h *TaskHandler) FindBlocked(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	tasks, err := h.taskService.FindBlocked(c.Request.Context(), projectID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTaskResponseList(tasks))
}

// ============================================
// SCRUM SPECIFIC
// ============================================

func (h *TaskHandler) GetBacklog(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projectID := c.Param("id")
	tasks, err := h.taskService.GetBacklog(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch backlog"})
		return
	}

	c.JSON(http.StatusOK, toTaskResponseList(tasks))
}

func (h *TaskHandler) GetSprintBoard(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("sprintId")
	board, err := h.taskService.GetSprintBoard(c.Request.Context(), sprintID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch board"})
		return
	}

	// Convert board map to response
	response := gin.H{
		"todo":        toTaskResponseList(board["todo"]),
		"in_progress": toTaskResponseList(board["in_progress"]),
		"in_review":   toTaskResponseList(board["in_review"]),
		"done":        toTaskResponseList(board["done"]),
	}

	c.JSON(http.StatusOK, response)
}

func (h *TaskHandler) GetSprintVelocity(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("sprintId")
	velocity, err := h.taskService.GetSprintVelocity(c.Request.Context(), sprintID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate velocity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sprintId": sprintID,
		"velocity": velocity,
	})
}

func (h *TaskHandler) GetSprintBurndown(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	sprintID := c.Param("sprintId")
	burndown, err := h.taskService.GetSprintBurndown(c.Request.Context(), sprintID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch burndown"})
		return
	}

	c.JSON(http.StatusOK, toBurndownResponse(burndown))
}

// ============================================
// BULK OPERATIONS
// ============================================

func (h *TaskHandler) BulkUpdateStatus(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.BulkUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.BulkUpdateStatus(c.Request.Context(), req.TaskIDs, req.Status, userID)
if err != nil {
	logAPIError(c, "Task.BulkUpdateStatus", err, map[string]interface{}{
		"taskCount": len(req.TaskIDs),
		"status":    req.Status,
	})
	handleServiceError(c, err)
	return
}


	c.JSON(http.StatusOK, gin.H{"message": "Tasks updated successfully"})
}

func (h *TaskHandler) BulkAssign(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.BulkAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.BulkAssign(c.Request.Context(), req.TaskIDs, req.AssigneeID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tasks assigned successfully"})
}

func (h *TaskHandler) BulkMoveToSprint(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.BulkMoveToSprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.BulkMoveToSprint(c.Request.Context(), req.TaskIDs, req.SprintID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tasks moved to sprint successfully"})
}

// ============================================
// HELPER FUNCTIONS
// ============================================

func handleServiceError(c *gin.Context, err error) {
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



func toTaskResponseList(tasks []*repository.Task) []models.TaskResponse {
	response := make([]models.TaskResponse, len(tasks))
	for i, t := range tasks {
		response[i] = toTaskResponse(t)
	}
	return response
}

func toCommentResponse(c *repository.TaskComment) models.CommentResponse {
	return models.CommentResponse{
		ID:             c.ID,
		TaskID:         c.TaskID,
		UserID:         c.UserID,
		Content:        c.Content,
		MentionedUsers: c.MentionedUsers,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func toCommentResponseList(comments []*repository.TaskComment) []models.CommentResponse {
	response := make([]models.CommentResponse, len(comments))
	for i, c := range comments {
		response[i] = toCommentResponse(c)
	}
	return response
}

func toAttachmentResponse(a *repository.TaskAttachment) models.AttachmentResponse {
	return models.AttachmentResponse{
		ID:        a.ID,
		TaskID:    a.TaskID,
		UserID:    a.UserID,
		Filename:  a.Filename,
		FileURL:   a.FileURL,
		FileSize:  a.FileSize,
		MimeType:  a.MimeType,
		CreatedAt: a.CreatedAt,
	}
}

func toAttachmentResponseList(attachments []*repository.TaskAttachment) []models.AttachmentResponse {
	response := make([]models.AttachmentResponse, len(attachments))
	for i, a := range attachments {
		response[i] = toAttachmentResponse(a)
	}
	return response
}

func toTimeEntryResponse(e *repository.TimeEntry) models.TimeEntryResponse {
	return models.TimeEntryResponse{
		ID:              e.ID,
		TaskID:          e.TaskID,
		UserID:          e.UserID,
		StartTime:       e.StartTime,
		EndTime:         e.EndTime,
		DurationSeconds: e.DurationSeconds,
		Description:     e.Description,
		IsManual:        e.IsManual,
		CreatedAt:       e.CreatedAt,
	}
}

func toTimeEntryResponseList(entries []*repository.TimeEntry) []models.TimeEntryResponse {
	response := make([]models.TimeEntryResponse, len(entries))
	for i, e := range entries {
		response[i] = toTimeEntryResponse(e)
	}
	return response
}

func toDependencyResponse(d *repository.TaskDependency) models.DependencyResponse {
	return models.DependencyResponse{
		ID:              d.ID,
		TaskID:          d.TaskID,
		DependsOnTaskID: d.DependsOnTaskID,
		DependencyType:  d.DependencyType,
		CreatedAt:       d.CreatedAt,
	}
}

func toDependencyResponseList(deps []*repository.TaskDependency) []models.DependencyResponse {
	response := make([]models.DependencyResponse, len(deps))
	for i, d := range deps {
		response[i] = toDependencyResponse(d)
	}
	return response
}

func toChecklistItemResponse(item *repository.ChecklistItem) models.ChecklistItemResponse {
	resp := models.ChecklistItemResponse{
		ID:          item.ID,
		ChecklistID: item.ChecklistID,
		Content:     item.Content,
		Position:    item.Position,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		AssigneeID:  item.AssigneeID, // just pass the *string
	}

	return resp
}



func toChecklistResponse(cl *repository.TaskChecklist) models.ChecklistResponse {
	items := make([]models.ChecklistItemResponse, 0)
	if cl.Items != nil {
		items = make([]models.ChecklistItemResponse, len(cl.Items))
		for i, item := range cl.Items {
			items[i] = toChecklistItemResponse(item)
		}
	}
	return models.ChecklistResponse{
		ID:        cl.ID,
		TaskID:    cl.TaskID,
		Title:     cl.Title,
		Items:     items,
		CreatedAt: cl.CreatedAt,
		UpdatedAt: cl.UpdatedAt,  // ✅ Added missing field
	}
}


func toChecklistResponseList(checklists []*repository.TaskChecklist) []models.ChecklistResponse {
	response := make([]models.ChecklistResponse, len(checklists))
	for i, cl := range checklists {
		response[i] = toChecklistResponse(cl)
	}
	return response
}


func toActivityResponse(a *repository.TaskActivity) models.ActivityResponse {
	return models.ActivityResponse{
		ID:        a.ID,
		TaskID:    a.TaskID,
		UserID:    a.UserID,
		Action:    a.Action,
		FieldName: a.FieldName,
		OldValue:  a.OldValue,
		NewValue:  a.NewValue,
		CreatedAt: a.CreatedAt,
	}
}

func toActivityResponseList(activities []*repository.TaskActivity) []models.ActivityResponse {
	response := make([]models.ActivityResponse, len(activities))
	for i, a := range activities {
		response[i] = toActivityResponse(a)
	}
	return response
}

func toBurndownResponse(b *service.SprintBurndown) models.SprintBurndownResponse {
	idealPoints := make([]models.BurndownPoint, len(b.IdealBurndown))
	for i, p := range b.IdealBurndown {
		idealPoints[i] = models.BurndownPoint{
			Date:   p.Date,
			Points: p.Points,
		}
	}

	actualPoints := make([]models.BurndownPoint, len(b.ActualBurndown))
	for i, p := range b.ActualBurndown {
		actualPoints[i] = models.BurndownPoint{
			Date:   p.Date,
			Points: p.Points,
		}
	}

	return models.SprintBurndownResponse{
		SprintID:         b.SprintID,
		StartDate:        b.StartDate,
		EndDate:          b.EndDate,
		TotalStoryPoints: b.TotalStoryPoints,
		CompletedPoints:  b.CompletedPoints,
		RemainingPoints:  b.RemainingPoints,
		IdealBurndown:    idealPoints,
		ActualBurndown:   actualPoints,
		CompletionRate:   b.CompletionRate,
	}
}