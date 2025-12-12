package models

import "time"

// ============================================
// Auth DTOs
// ============================================

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
}

// ============================================
// User DTOs
// ============================================

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Avatar    *string   `json:"avatar,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type UpdateUserRequest struct {
	Name   *string `json:"name,omitempty"`
	Avatar *string `json:"avatar,omitempty"`
}

// ============================================
// Workspace DTOs
// ============================================

type CreateWorkspaceRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type UpdateWorkspaceRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type WorkspaceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	Color       *string   `json:"color,omitempty"`
	OwnerID     string    `json:"ownerId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type WorkspaceMemberResponse struct {
	ID          string       `json:"id"`
	WorkspaceID string       `json:"workspaceId"`
	UserID      string       `json:"userId"`
	Role        string       `json:"role"`
	User        UserResponse `json:"user"`
	JoinedAt    time.Time    `json:"joinedAt"`
}

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=ADMIN MEMBER VIEWER"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=ADMIN MEMBER VIEWER"`
}

// ============================================
// Space DTOs
// ============================================

type CreateSpaceRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type UpdateSpaceRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type SpaceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	Color       *string   `json:"color,omitempty"`
	WorkspaceID string    `json:"workspaceId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ============================================
// Project DTOs
// ============================================

type CreateProjectRequest struct {
	Name        string  `json:"name" binding:"required"`
	Key         string  `json:"key" binding:"required,min=2,max=10"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
	LeadID      *string `json:"leadId,omitempty"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Key         *string `json:"key,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
	LeadID      *string `json:"leadId,omitempty"`
}

type ProjectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	Description *string   `json:"description,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	Color       *string   `json:"color,omitempty"`
	SpaceID     string    `json:"spaceId"`
	LeadID      *string   `json:"leadId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProjectMemberResponse struct {
	ID        string       `json:"id"`
	ProjectID string       `json:"projectId"`
	UserID    string       `json:"userId"`
	Role      string       `json:"role"`
	User      UserResponse `json:"user"`
	JoinedAt  time.Time    `json:"joinedAt"`
}

type AddProjectMemberRequest struct {
	UserID string `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=LEAD MEMBER VIEWER"`
}

// ============================================
// Sprint DTOs
// ============================================

type CreateSprintRequest struct {
	Name      string     `json:"name" binding:"required"`
	Goal      *string    `json:"goal,omitempty"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}

type UpdateSprintRequest struct {
	Name      *string    `json:"name,omitempty"`
	Goal      *string    `json:"goal,omitempty"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}

type CompleteSprintRequest struct {
	MoveIncomplete string `json:"moveIncomplete,omitempty"` // "backlog" or sprint ID
}

type SprintResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Goal      *string    `json:"goal,omitempty"`
	ProjectID string     `json:"projectId"`
	Status    string     `json:"status"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// ============================================
// Task DTOs
// ============================================

type CreateTaskRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description *string  `json:"description,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Priority    *string  `json:"priority,omitempty"`
	Type        *string  `json:"type,omitempty"`
	AssigneeID  *string  `json:"assigneeId,omitempty"`
	SprintID    *string  `json:"sprintId,omitempty"`
	ParentID    *string  `json:"parentId,omitempty"`
	StoryPoints *int     `json:"storyPoints,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	Type        *string    `json:"type,omitempty"`
	AssigneeID  *string    `json:"assigneeId,omitempty"`
	SprintID    *string    `json:"sprintId,omitempty"`
	ParentID    *string    `json:"parentId,omitempty"`
	StoryPoints *int       `json:"storyPoints,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	Labels      []string   `json:"labels,omitempty"`
	OrderIndex  *int       `json:"orderIndex,omitempty"`
}

type BulkUpdateTaskRequest struct {
	Tasks []BulkTaskUpdate `json:"tasks" binding:"required"`
}

type BulkTaskUpdate struct {
	ID         string  `json:"id" binding:"required"`
	Status     *string `json:"status,omitempty"`
	SprintID   *string `json:"sprintId,omitempty"`
	OrderIndex *int    `json:"order,omitempty"`
}

type TaskResponse struct {
	ID          string              `json:"id"`
	Key         string              `json:"key"`
	Title       string              `json:"title"`
	Description *string             `json:"description,omitempty"`
	Status      string              `json:"status"`
	Priority    string              `json:"priority"`
	Type        string              `json:"type"`
	ProjectID   string              `json:"projectId"`
	SprintID    *string             `json:"sprintId,omitempty"`
	AssigneeID  *string             `json:"assigneeId,omitempty"`
	ReporterID  string              `json:"reporterId"`
	ParentID    *string             `json:"parentId,omitempty"`
	StoryPoints *int                `json:"storyPoints,omitempty"`
	DueDate     *time.Time          `json:"dueDate,omitempty"`
	Labels      []string            `json:"labels"`
	Assignee    *UserResponse       `json:"assignee,omitempty"`
	Reporter    *UserResponse       `json:"reporter,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

type TaskFilters struct {
	Status     []string `form:"status"`
	Priority   []string `form:"priority"`
	Type       []string `form:"type"`
	AssigneeID []string `form:"assigneeId"`
	SprintID   string   `form:"sprintId"`
	Labels     []string `form:"labels"`
	Search     string   `form:"search"`
	Limit      int      `form:"limit"`
	Offset     int      `form:"offset"`
}

// ============================================
// Comment DTOs
// ============================================

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type CommentResponse struct {
	ID        string       `json:"id"`
	TaskID    string       `json:"taskId"`
	UserID    string       `json:"userId"`
	Content   string       `json:"content"`
	User      UserResponse `json:"user"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// ============================================
// Label DTOs
// ============================================

type CreateLabelRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required"`
}

type UpdateLabelRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

type LabelResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	ProjectID string    `json:"projectId"`
	CreatedAt time.Time `json:"createdAt"`
}

// ============================================
// Notification DTOs
// ============================================

type NotificationResponse struct {
	ID        string                  `json:"id"`
	UserID    string                  `json:"userId"`
	Type      string                  `json:"type"`
	Title     string                  `json:"title"`
	Message   string                  `json:"message"`
	Read      bool                    `json:"read"`
	Data      *map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time               `json:"createdAt"`
}

type NotificationCountResponse struct {
	Total  int `json:"total"`
	Unread int `json:"unread"`
}
