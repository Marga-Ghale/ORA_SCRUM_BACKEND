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
		OrderIndex:  t.OrderIndex,
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
