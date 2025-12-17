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


func toWorkspaceResponse(ws *repository.Workspace) models.WorkspaceResponse {
	return models.WorkspaceResponse{
		ID:           ws.ID,
		Name:         ws.Name,
		Description:  ws.Description,
		Icon:         ws.Icon,
		Color:        ws.Color,
		OwnerID:      ws.OwnerID,
		Visibility:   ws.Visibility,
		AllowedUsers: ws.AllowedUsers,
		AllowedTeams: ws.AllowedTeams,
		CreatedAt:    ws.CreatedAt,
		UpdatedAt:    ws.UpdatedAt,
	}
}




func toFolderResponse(f *repository.Folder) models.FolderResponse {
	return models.FolderResponse{
		ID:           f.ID,
		Name:         f.Name,
		Description:  f.Description,
		Icon:         f.Icon,
		Color:        f.Color,
		OwnerID:      f.OwnerID,
		Visibility:   f.Visibility,
		AllowedUsers: f.AllowedUsers,
		AllowedTeams: f.AllowedTeams,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}


func toSpaceResponse(s *repository.Space) models.SpaceResponse {
	return models.SpaceResponse{
		ID:           s.ID,
		WorkspaceID:  s.WorkspaceID, // ✓ Include parent
		Name:         s.Name,
		Description:  s.Description,
		Icon:         s.Icon,
		Color:        s.Color,
		OwnerID:      s.OwnerID,
		Visibility:   s.Visibility,
		AllowedUsers: s.AllowedUsers,
		AllowedTeams: s.AllowedTeams,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
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


func toTaskResponse(t *repository.Task) models.TaskResponse {
	return models.TaskResponse{
		ID:             t.ID,
		ProjectID:      t.ProjectID,
		SprintID:       t.SprintID,
		ParentTaskID:   t.ParentTaskID,
		Title:          t.Title,
		Description:    t.Description,
		Status:         t.Status,
		Priority:       t.Priority,
		AssigneeIDs:    t.AssigneeIDs,
		WatcherIDs:     t.WatcherIDs,
		LabelIDs:       t.LabelIDs,
		EstimatedHours: t.EstimatedHours,
		ActualHours:    t.ActualHours,
		StoryPoints:    t.StoryPoints,
		StartDate:      t.StartDate,
		DueDate:        t.DueDate,
		CompletedAt:    t.CompletedAt,
		Position:       t.Position,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
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
