package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// handleServiceError maps service errors to HTTP responses
func handleServiceError(c *gin.Context, err error) {
	switch err {
	case service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
	case service.ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	case service.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	case service.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": "Resource already exists"})
	case service.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	case service.ErrInvalidToken:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

// TeamHandler handles team-related HTTP requests
type TeamHandler struct {
	teamSvc service.TeamService
}

// NewTeamHandler creates a new team handler
func NewTeamHandler(teamSvc service.TeamService) *TeamHandler {
	return &TeamHandler{teamSvc: teamSvc}
}

// CreateTeamRequest represents the request body for creating a team
type CreateTeamRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Avatar      *string `json:"avatar"`
	Color       *string `json:"color"`
}

// UpdateTeamRequest represents the request body for updating a team
type UpdateTeamRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Avatar      *string `json:"avatar"`
	Color       *string `json:"color"`
}

// AddTeamMemberRequest represents the request body for adding a team member
type AddTeamMemberRequest struct {
	UserID *string `json:"userId"`
	Email  *string `json:"email"`
	Role   string  `json:"role" binding:"required"`
}

// UpdateTeamMemberRoleRequest represents the request body for updating member role
type UpdateTeamMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// Create creates a new team
func (h *TeamHandler) Create(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamSvc.Create(c.Request.Context(), workspaceID, userID, req.Name, req.Description, req.Avatar, req.Color)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, team)
}

// Get retrieves a team by ID
func (h *TeamHandler) Get(c *gin.Context) {
	teamID := c.Param("id")

	team, err := h.teamSvc.GetByID(c.Request.Context(), teamID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, team)
}

// ListByWorkspace lists all teams in a workspace
func (h *TeamHandler) ListByWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")

	teams, err := h.teamSvc.ListByWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, teams)
}

// ListMyTeams lists all teams the current user is a member of
func (h *TeamHandler) ListMyTeams(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	teams, err := h.teamSvc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, teams)
}

// Update updates a team
func (h *TeamHandler) Update(c *gin.Context) {
	teamID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamSvc.Update(c.Request.Context(), teamID, userID, req.Name, req.Description, req.Avatar, req.Color)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, team)
}

// Delete deletes a team
func (h *TeamHandler) Delete(c *gin.Context) {
	teamID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.teamSvc.Delete(c.Request.Context(), teamID, userID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddMember adds a member to a team
func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID := c.Param("id")
	addedByID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	if req.UserID != nil {
		err = h.teamSvc.AddMember(c.Request.Context(), teamID, *req.UserID, req.Role, addedByID)
	} else if req.Email != nil {
		err = h.teamSvc.AddMemberByEmail(c.Request.Context(), teamID, *req.Email, req.Role, addedByID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId or email is required"})
		return
	}

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "member added"})
}

// ListMembers lists all members of a team
func (h *TeamHandler) ListMembers(c *gin.Context) {
	teamID := c.Param("id")

	members, err := h.teamSvc.ListMembers(c.Request.Context(), teamID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, members)
}

// UpdateMemberRole updates a team member's role
func (h *TeamHandler) UpdateMemberRole(c *gin.Context) {
	teamID := c.Param("id")
	userIDToUpdate := c.Param("userId")

	var req UpdateTeamMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.teamSvc.UpdateMemberRole(c.Request.Context(), teamID, userIDToUpdate, req.Role); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

// RemoveMember removes a member from a team
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID := c.Param("id")
	userIDToRemove := c.Param("userId")

	if err := h.teamSvc.RemoveMember(c.Request.Context(), teamID, userIDToRemove); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// Invitation Handler
// ============================================

// InvitationHandler handles invitation-related HTTP requests
type InvitationHandler struct {
	invitationSvc service.InvitationService
}

// NewInvitationHandler creates a new invitation handler
func NewInvitationHandler(invitationSvc service.InvitationService) *InvitationHandler {
	return &InvitationHandler{invitationSvc: invitationSvc}
}

// CreateInvitationRequest represents the request body for creating an invitation
type CreateInvitationRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

// CreateWorkspaceInvitation creates an invitation to join a workspace
func (h *InvitationHandler) CreateWorkspaceInvitation(c *gin.Context) {
	workspaceID := c.Param("id")
	inviterID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invitation, err := h.invitationSvc.CreateWorkspaceInvitation(c.Request.Context(), workspaceID, req.Email, req.Role, inviterID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// CreateProjectInvitation creates an invitation to join a project
func (h *InvitationHandler) CreateProjectInvitation(c *gin.Context) {
	projectID := c.Param("id")
	inviterID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invitation, err := h.invitationSvc.CreateProjectInvitation(c.Request.Context(), projectID, req.Email, req.Role, inviterID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// AcceptInvitation accepts an invitation
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.invitationSvc.AcceptInvitation(c.Request.Context(), token, userID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invitation accepted"})
}

// GetMyInvitations gets all pending invitations for the current user
func (h *InvitationHandler) GetMyInvitations(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	invitations, err := h.invitationSvc.GetPendingInvitations(c.Request.Context(), email)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, invitations)
}

// CancelInvitation cancels an invitation
func (h *InvitationHandler) CancelInvitation(c *gin.Context) {
	invitationID := c.Param("id")

	if err := h.invitationSvc.CancelInvitation(c.Request.Context(), invitationID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

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
