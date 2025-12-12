package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	Auth         *AuthHandler
	User         *UserHandler
	Workspace    *WorkspaceHandler
	Space        *SpaceHandler
	Project      *ProjectHandler
	Sprint       *SprintHandler
	Task         *TaskHandler
	Comment      *CommentHandler
	Label        *LabelHandler
	Notification *NotificationHandler
}

// NewHandlers creates all handlers
func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		Auth:         &AuthHandler{authService: services.Auth},
		User:         &UserHandler{userService: services.User},
		Workspace:    &WorkspaceHandler{workspaceService: services.Workspace},
		Space:        &SpaceHandler{spaceService: services.Space},
		Project:      &ProjectHandler{projectService: services.Project},
		Sprint:       &SprintHandler{sprintService: services.Sprint},
		Task:         &TaskHandler{taskService: services.Task},
		Comment:      &CommentHandler{commentService: services.Comment},
		Label:        &LabelHandler{labelService: services.Label},
		Notification: &NotificationHandler{notificationService: services.Notification},
	}
}

// ============================================
// Auth Handler
// ============================================

type AuthHandler struct {
	authService service.AuthService
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, accessToken, refreshToken, err := h.authService.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if err == service.ErrUserExists {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.authService.Logout(c.Request.Context(), req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ============================================
// User Handler
// ============================================

type UserHandler struct {
	userService service.UserService
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Update(c.Request.Context(), userID, req.Name, req.Avatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

// ============================================
// Workspace Handler
// ============================================

type WorkspaceHandler struct {
	workspaceService service.WorkspaceService
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	workspaces, err := h.workspaceService.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workspaces"})
		return
	}

	response := make([]models.WorkspaceResponse, len(workspaces))
	for i, ws := range workspaces {
		response[i] = toWorkspaceResponse(ws)
	}

	c.JSON(http.StatusOK, response)
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.workspaceService.Create(c.Request.Context(), userID, req.Name, req.Description, req.Icon, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workspace"})
		return
	}

	c.JSON(http.StatusCreated, toWorkspaceResponse(workspace))
}

func (h *WorkspaceHandler) Get(c *gin.Context) {
	id := c.Param("id")

	workspace, err := h.workspaceService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
		return
	}

	c.JSON(http.StatusOK, toWorkspaceResponse(workspace))
}

func (h *WorkspaceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.workspaceService.Update(c.Request.Context(), id, req.Name, req.Description, req.Icon, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workspace"})
		return
	}

	c.JSON(http.StatusOK, toWorkspaceResponse(workspace))
}

func (h *WorkspaceHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.workspaceService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workspace"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	id := c.Param("id")

	members, err := h.workspaceService.ListMembers(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	response := make([]models.WorkspaceMemberResponse, len(members))
	for i, m := range members {
		response[i] = toWorkspaceMemberResponse(m)
	}

	c.JSON(http.StatusOK, response)
}

func (h *WorkspaceHandler) AddMember(c *gin.Context) {
	id := c.Param("id")

	var req models.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.workspaceService.AddMember(c.Request.Context(), id, req.Email, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

func (h *WorkspaceHandler) UpdateMemberRole(c *gin.Context) {
	id := c.Param("id")
	userID := c.Param("userId")

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.workspaceService.UpdateMemberRole(c.Request.Context(), id, userID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	id := c.Param("id")
	userID := c.Param("userId")

	if err := h.workspaceService.RemoveMember(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// Space Handler
// ============================================

type SpaceHandler struct {
	spaceService service.SpaceService
}

func (h *SpaceHandler) ListByWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")

	spaces, err := h.spaceService.ListByWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch spaces"})
		return
	}

	response := make([]models.SpaceResponse, len(spaces))
	for i, s := range spaces {
		response[i] = toSpaceResponse(s)
	}

	c.JSON(http.StatusOK, response)
}

func (h *SpaceHandler) Create(c *gin.Context) {
	workspaceID := c.Param("id")

	var req models.CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	space, err := h.spaceService.Create(c.Request.Context(), workspaceID, req.Name, req.Description, req.Icon, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create space"})
		return
	}

	c.JSON(http.StatusCreated, toSpaceResponse(space))
}

func (h *SpaceHandler) Get(c *gin.Context) {
	id := c.Param("id")

	space, err := h.spaceService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Space not found"})
		return
	}

	c.JSON(http.StatusOK, toSpaceResponse(space))
}

func (h *SpaceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	space, err := h.spaceService.Update(c.Request.Context(), id, req.Name, req.Description, req.Icon, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update space"})
		return
	}

	c.JSON(http.StatusOK, toSpaceResponse(space))
}

func (h *SpaceHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.spaceService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete space"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// Project Handler
// ============================================

type ProjectHandler struct {
	projectService service.ProjectService
}

func (h *ProjectHandler) ListBySpace(c *gin.Context) {
	spaceID := c.Param("id")

	projects, err := h.projectService.ListBySpace(c.Request.Context(), spaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	response := make([]models.ProjectResponse, len(projects))
	for i, p := range projects {
		response[i] = toProjectResponse(p)
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	spaceID := c.Param("id")

	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.Create(c.Request.Context(), spaceID, req.Name, req.Key, req.Description, req.Icon, req.Color, req.LeadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, toProjectResponse(project))
}

func (h *ProjectHandler) Get(c *gin.Context) {
	id := c.Param("id")

	project, err := h.projectService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, toProjectResponse(project))
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.Update(c.Request.Context(), id, req.Name, req.Key, req.Description, req.Icon, req.Color, req.LeadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, toProjectResponse(project))
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.projectService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ProjectHandler) ListMembers(c *gin.Context) {
	id := c.Param("id")

	members, err := h.projectService.ListMembers(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	response := make([]models.ProjectMemberResponse, len(members))
	for i, m := range members {
		response[i] = toProjectMemberResponse(m)
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProjectHandler) AddMember(c *gin.Context) {
	id := c.Param("id")

	var req models.AddProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectService.AddMember(c.Request.Context(), id, req.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	id := c.Param("id")
	userID := c.Param("userId")

	if err := h.projectService.RemoveMember(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

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

	sprint, err := h.sprintService.Start(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start sprint"})
		return
	}

	c.JSON(http.StatusOK, toSprintResponse(sprint))
}

func (h *SprintHandler) Complete(c *gin.Context) {
	id := c.Param("id")

	var req models.CompleteSprintRequest
	c.ShouldBindJSON(&req)

	sprint, err := h.sprintService.Complete(c.Request.Context(), id, req.MoveIncomplete)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete sprint"})
		return
	}

	c.JSON(http.StatusOK, toSprintResponse(sprint))
}

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

	task, err := h.taskService.Update(c.Request.Context(), id, updates)
	if err != nil {
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

	if err := h.taskService.Delete(c.Request.Context(), id); err != nil {
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

// ============================================
// Comment Handler
// ============================================

type CommentHandler struct {
	commentService service.CommentService
}

func (h *CommentHandler) ListByTask(c *gin.Context) {
	taskID := c.Param("id")

	comments, err := h.commentService.ListByTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	response := make([]models.CommentResponse, len(comments))
	for i, cm := range comments {
		response[i] = toCommentResponse(cm)
	}

	c.JSON(http.StatusOK, response)
}

func (h *CommentHandler) Create(c *gin.Context) {
	taskID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.commentService.Create(c.Request.Context(), taskID, userID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, toCommentResponse(comment))
}

func (h *CommentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.commentService.Update(c.Request.Context(), id, userID, req.Content)
	if err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this comment"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	c.JSON(http.StatusOK, toCommentResponse(comment))
}

func (h *CommentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.commentService.Delete(c.Request.Context(), id, userID); err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this comment"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// Label Handler
// ============================================

type LabelHandler struct {
	labelService service.LabelService
}

func (h *LabelHandler) ListByProject(c *gin.Context) {
	projectID := c.Param("id")

	labels, err := h.labelService.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch labels"})
		return
	}

	response := make([]models.LabelResponse, len(labels))
	for i, l := range labels {
		response[i] = toLabelResponse(l)
	}

	c.JSON(http.StatusOK, response)
}

func (h *LabelHandler) Create(c *gin.Context) {
	projectID := c.Param("id")

	var req models.CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	label, err := h.labelService.Create(c.Request.Context(), projectID, req.Name, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create label"})
		return
	}

	c.JSON(http.StatusCreated, toLabelResponse(label))
}

func (h *LabelHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	label, err := h.labelService.Update(c.Request.Context(), id, req.Name, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update label"})
		return
	}

	c.JSON(http.StatusOK, toLabelResponse(label))
}

func (h *LabelHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.labelService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete label"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// Notification Handler
// ============================================

type NotificationHandler struct {
	notificationService service.NotificationService
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	unreadOnly := c.Query("unread") == "true"

	notifications, err := h.notificationService.List(c.Request.Context(), userID, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	response := make([]models.NotificationResponse, len(notifications))
	for i, n := range notifications {
		response[i] = toNotificationResponse(n)
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) Count(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	total, unread, err := h.notificationService.Count(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count notifications"})
		return
	}

	c.JSON(http.StatusOK, models.NotificationCountResponse{
		Total:  total,
		Unread: unread,
	})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id := c.Param("id")

	if err := h.notificationService.MarkAsRead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.notificationService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *NotificationHandler) DeleteAll(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.notificationService.DeleteAll(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notifications"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================
// Response Mappers
// ============================================

func toUserResponse(u *repository.User) models.UserResponse {
	return models.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Avatar:    u.Avatar,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

func toWorkspaceResponse(w *repository.Workspace) models.WorkspaceResponse {
	return models.WorkspaceResponse{
		ID:          w.ID,
		Name:        w.Name,
		Description: w.Description,
		Icon:        w.Icon,
		Color:       w.Color,
		OwnerID:     w.OwnerID,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

func toWorkspaceMemberResponse(m *repository.WorkspaceMember) models.WorkspaceMemberResponse {
	resp := models.WorkspaceMemberResponse{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Role:        m.Role,
		JoinedAt:    m.JoinedAt,
	}
	if m.User != nil {
		resp.User = toUserResponse(m.User)
	}
	return resp
}

func toSpaceResponse(s *repository.Space) models.SpaceResponse {
	return models.SpaceResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Icon:        s.Icon,
		Color:       s.Color,
		WorkspaceID: s.WorkspaceID,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func toProjectResponse(p *repository.Project) models.ProjectResponse {
	return models.ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Key:         p.Key,
		Description: p.Description,
		Icon:        p.Icon,
		Color:       p.Color,
		SpaceID:     p.SpaceID,
		LeadID:      p.LeadID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toProjectMemberResponse(m *repository.ProjectMember) models.ProjectMemberResponse {
	resp := models.ProjectMemberResponse{
		ID:        m.ID,
		ProjectID: m.ProjectID,
		UserID:    m.UserID,
		Role:      m.Role,
		JoinedAt:  m.JoinedAt,
	}
	if m.User != nil {
		resp.User = toUserResponse(m.User)
	}
	return resp
}

func toSprintResponse(s *repository.Sprint) models.SprintResponse {
	return models.SprintResponse{
		ID:        s.ID,
		Name:      s.Name,
		Goal:      s.Goal,
		ProjectID: s.ProjectID,
		Status:    s.Status,
		StartDate: s.StartDate,
		EndDate:   s.EndDate,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func toTaskResponse(t *repository.Task) models.TaskResponse {
	resp := models.TaskResponse{
		ID:          t.ID,
		Key:         t.Key,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		Type:        t.Type,
		ProjectID:   t.ProjectID,
		SprintID:    t.SprintID,
		AssigneeID:  t.AssigneeID,
		ReporterID:  t.ReporterID,
		ParentID:    t.ParentID,
		StoryPoints: t.StoryPoints,
		DueDate:     t.DueDate,
		Labels:      t.Labels,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.Assignee != nil {
		userResp := toUserResponse(t.Assignee)
		resp.Assignee = &userResp
	}
	if t.Reporter != nil {
		userResp := toUserResponse(t.Reporter)
		resp.Reporter = &userResp
	}
	if resp.Labels == nil {
		resp.Labels = []string{}
	}
	return resp
}

func toCommentResponse(c *repository.Comment) models.CommentResponse {
	resp := models.CommentResponse{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	if c.User != nil {
		resp.User = toUserResponse(c.User)
	}
	return resp
}

func toLabelResponse(l *repository.Label) models.LabelResponse {
	return models.LabelResponse{
		ID:        l.ID,
		Name:      l.Name,
		Color:     l.Color,
		ProjectID: l.ProjectID,
		CreatedAt: l.CreatedAt,
	}
}

func toNotificationResponse(n *repository.Notification) models.NotificationResponse {
	resp := models.NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Read:      n.Read,
		CreatedAt: n.CreatedAt,
	}
	if n.Data != nil {
		resp.Data = &n.Data
	}
	return resp
}
