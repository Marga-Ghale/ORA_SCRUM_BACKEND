package handlers

import (
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	Auth         *AuthHandler
	User         *UserHandler
	Workspace    *WorkspaceHandler
	Folder  	 *FolderHandler
	Space        *SpaceHandler
	Project      *ProjectHandler
	Task         *TaskHandler
	Label        *LabelHandler
	Notification *NotificationHandler
	Member	   	 *MemberHandler
}

// NewHandlers creates all handlers
func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		Auth:         &AuthHandler{authService: services.Auth},
		User:         &UserHandler{userService: services.User},
		Workspace:    &WorkspaceHandler{workspaceService: services.Workspace},
		Folder:       &FolderHandler{folderService: services.Folder},
		Space:        &SpaceHandler{spaceService: services.Space},
		Project:      &ProjectHandler{projectService: services.Project},
		Task:         &TaskHandler{taskService: services.Task},
		Label:        &LabelHandler{labelService: services.Label},
		Notification: &NotificationHandler{notificationService: services.Notification},
		Member:       &MemberHandler{memberService: services.Member},
	}
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


// ============================================
// COMPREHENSIVE TASK RESPONSE MAPPER
// ============================================

func toTaskResponse(t *repository.Task) models.TaskResponse {
	if t == nil {
		return models.TaskResponse{}
	}

	return models.TaskResponse{
		ID:             t.ID,
		Title:          t.Title,
		Description:    t.Description,
		Status:         t.Status,
		Priority:       t.Priority,
		Type:           t.Type,
		ProjectID:      t.ProjectID,
		SprintID:       t.SprintID,
		ParentTaskID:   t.ParentTaskID,
		AssigneeIDs:    safeStringSlice(t.AssigneeIDs),
		WatcherIDs:     safeStringSlice(t.WatcherIDs),
		LabelIDs:       safeStringSlice(t.LabelIDs),
		StoryPoints:    t.StoryPoints,
		EstimatedHours: t.EstimatedHours,
		ActualHours:    t.ActualHours,
		StartDate:      t.StartDate,
		DueDate:        t.DueDate,
		CompletedAt:    t.CompletedAt,
		Blocked:        t.Blocked,
		Position:       t.Position,
		CreatedBy:      t.CreatedBy,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		SubtaskCount:   0,    // Will be populated separately if needed
		Subtasks:       nil,  // Will be populated separately if needed
	}
}


// Enhanced converter with subtasks
// Enhanced converter with subtasks
func toTaskResponseWithSubtasks(t *repository.Task, subtasks []*repository.Task) models.TaskResponse {
	response := toTaskResponse(t)
	response.SubtaskCount = len(subtasks)
	
	if len(subtasks) > 0 {
		response.Subtasks = make([]models.TaskResponse, len(subtasks)) // ✅ Changed from []*models.TaskResponse to []models.TaskResponse
		for i, st := range subtasks {
			response.Subtasks[i] = toTaskResponse(st) // ✅ This now works
		}
	}
	
	return response
}

// Helper to ensure nil slices become empty slices
func safeStringSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// Helper to ensure nil int slices become empty slices
func safeIntSlice(s []int) []int {
	if s == nil {
		return []int{}
	}
	return s
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



// ============================================
// Helper Functions
// ============================================

// Helper function to convert repository.Folder to models.FolderResponse
func toFolderResponse(f *repository.Folder) models.FolderResponse {
	resp := models.FolderResponse{
		ID:        f.ID,
		SpaceID:   f.SpaceID,
		Name:      f.Name,
		OwnerID:   f.OwnerID,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}

	// Handle optional fields
	if f.Description != nil {
		resp.Description = *f.Description
	}
	if f.Icon != nil {
		resp.Icon = *f.Icon
	}
	if f.Color != nil {
		resp.Color = *f.Color
	}
	if f.Visibility != nil {
		resp.Visibility = *f.Visibility
	}

	// Handle arrays
	if f.AllowedUsers != nil {
		resp.AllowedUsers = f.AllowedUsers
	} else {
		resp.AllowedUsers = []string{}
	}

	if f.AllowedTeams != nil {
		resp.AllowedTeams = f.AllowedTeams
	} else {
		resp.AllowedTeams = []string{}
	}

	return resp
}



// ============================================
// Helper Functions
// ============================================

// Helper function to convert repository.Space to models.SpaceResponse
func toSpaceResponse(s *repository.Space) models.SpaceResponse {
	resp := models.SpaceResponse{
		ID:          s.ID,
		WorkspaceID: s.WorkspaceID,
		Name:        s.Name,
		OwnerID:     s.OwnerID,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}

	// Handle optional fields
	if s.Description != nil {
		resp.Description = *s.Description
	}
	if s.Icon != nil {
		resp.Icon = *s.Icon
	}
	if s.Color != nil {
		resp.Color = *s.Color
	}
	if s.Visibility != nil {
		resp.Visibility = *s.Visibility
	}

	// Handle arrays
	if s.AllowedUsers != nil {
		resp.AllowedUsers = s.AllowedUsers
	} else {
		resp.AllowedUsers = []string{}
	}

	if s.AllowedTeams != nil {
		resp.AllowedTeams = s.AllowedTeams
	} else {
		resp.AllowedTeams = []string{}
	}

	return resp
}

// Helper function to convert repository.Workspace to models.WorkspaceResponse
func toWorkspaceResponse(ws *repository.Workspace) models.WorkspaceResponse {
	resp := models.WorkspaceResponse{
		ID:        ws.ID,
		Name:      ws.Name,
		OwnerID:   ws.OwnerID,
		CreatedAt: ws.CreatedAt,
		UpdatedAt: ws.UpdatedAt,
	}

	// Handle optional fields
	if ws.Description != nil {
		resp.Description = *ws.Description
	}
	if ws.Icon != nil {
		resp.Icon = *ws.Icon
	}
	if ws.Color != nil {
		resp.Color = *ws.Color
	}
	if ws.Visibility != nil {
		resp.Visibility = *ws.Visibility
	}
	
	// Handle arrays
	if ws.AllowedUsers != nil {
		resp.AllowedUsers = ws.AllowedUsers
	} else {
		resp.AllowedUsers = []string{}
	}
	
	if ws.AllowedTeams != nil {
		resp.AllowedTeams = ws.AllowedTeams
	} else {
		resp.AllowedTeams = []string{}
	}

	return resp
}