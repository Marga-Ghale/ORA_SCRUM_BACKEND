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

type WorkspaceMemberResponse struct {
	ID          string        `json:"id"`
	WorkspaceID string        `json:"workspaceId"`
	UserID      string        `json:"userId"`
	Role        string        `json:"role"`
	JoinedAt    time.Time     `json:"joinedAt"`
	User        *UserResponse `json:"user,omitempty"`
}

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=owner admin member viewer"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member viewer"`
}


type SpaceMemberResponse struct {
	ID       string        `json:"id"`
	SpaceID  string        `json:"spaceId"`
	UserID   string        `json:"userId"`
	Role     string        `json:"role"`
	JoinedAt time.Time     `json:"joinedAt"`
	User     *UserResponse `json:"user,omitempty"`
}

type InviteSpaceRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member viewer"`
}

type FolderMemberResponse struct {
	ID       string        `json:"id"`
	FolderID string        `json:"folderId"`
	UserID   string        `json:"userId"`
	Role     string        `json:"role"`
	JoinedAt time.Time     `json:"joinedAt"`
	User     *UserResponse `json:"user,omitempty"`
}

type InviteFolderRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member viewer"`
}

type ProjectMemberResponse struct {
	ID        string        `json:"id"`
	ProjectID string        `json:"projectId"`
	UserID    string        `json:"userId"`
	Role      string        `json:"role"`
	User      *UserResponse `json:"user,omitempty"`
	JoinedAt  time.Time     `json:"joinedAt"`
}

type AddProjectMemberRequest struct {
	UserID string `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=lead member viewer"`
}

// ============================================
// Sprint DTOs (Under Project)
// ============================================

type CreateSprintRequest struct {
	Name      string     `json:"name" binding:"required"`
	Goal      *string    `json:"goal"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

type UpdateSprintRequest struct {
	Name      *string    `json:"name"`
	Goal      *string    `json:"goal"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

type CompleteSprintRequest struct {
	MoveIncomplete string `json:"moveIncomplete"` // "backlog" or sprint ID
}

type SprintResponse struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"projectId"` // ✓ parent reference
	Name      string     `json:"name"`
	Goal      *string    `json:"goal"`
	Status    string     `json:"status"` // PLANNING, ACTIVE, COMPLETED
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// ============================================
// Task DTOs (Under Project)
// ============================================

// models/create_task_request.go

type CreateTaskRequest struct {
	ProjectID      string     `json:"projectId,omitempty"`
	SprintID       *string    `json:"sprintId"`
	ParentTaskID   *string    `json:"parentTaskId"`
	Title          string     `json:"title" binding:"required"`
	Description    *string   `json:"description"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	AssigneeIDs    []string   `json:"assigneeIds"`
	LabelIDs       []string   `json:"labelIds"`
	EstimatedHours *float64   `json:"estimatedHours"`
	StoryPoints    *int       `json:"storyPoints"`
	StartDate      *time.Time `json:"startDate"`
	DueDate        *time.Time `json:"dueDate"`
}


// models/update_task_request.go
type UpdateTaskRequest struct {
	Title          *string    `json:"title"`
	Description    *string    `json:"description"`
	Status         *string    `json:"status"`
	Priority       *string    `json:"priority"`
	SprintID       *string    `json:"sprintId"`
	AssigneeIDs    *[]string  `json:"assigneeIds"`
	LabelIDs       *[]string  `json:"labelIds"`
	EstimatedHours *float64   `json:"estimatedHours"`
	ActualHours    *float64   `json:"actualHours"`
	StoryPoints    *int       `json:"storyPoints"`
	StartDate      *time.Time `json:"startDate"`
	DueDate        *time.Time `json:"dueDate"`
}



// internal/models/task_model.go - TaskResponse
type TaskResponse struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"projectId"`
	SprintID       *string    `json:"sprintId,omitempty"`
	ParentTaskID   *string    `json:"parentTaskId,omitempty"`
	Title          string     `json:"title"`
	Description    *string    `json:"description,omitempty"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	Type           *string    `json:"type,omitempty"` // ← ADD THIS
	AssigneeIDs    []string   `json:"assigneeIds"`
	WatcherIDs     []string   `json:"watcherIds"`
	LabelIDs       []string   `json:"labelIds"`
	EstimatedHours *float64   `json:"estimatedHours,omitempty"`
	ActualHours    *float64   `json:"actualHours,omitempty"`
	StoryPoints    *int       `json:"storyPoints,omitempty"`
	StartDate      *time.Time `json:"startDate,omitempty"`
	DueDate        *time.Time `json:"dueDate,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	Blocked        bool       `json:"blocked"`
	Position       int        `json:"position"`
	CreatedBy      *string    `json:"createdBy,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type TaskFilters struct {
	ProjectID   string
	SprintID    *string
	AssigneeIDs []string
	Status      []string
	Priority    []string
	LabelIDs    []string
	Search      *string
	DueBefore   *time.Time
	DueAfter    *time.Time
	Overdue     *bool
	Blocked     *bool
	Limit       int
	Offset      int
}


type TaskFiltersRequest struct {
    ProjectID   string     `json:"projectId" binding:"required"`
    SprintID    *string    `json:"sprintId"`
    AssigneeIDs []string   `json:"assigneeIds"`
    Statuses    []string   `json:"statuses"`
    Priorities  []string   `json:"priorities"`
    LabelIDs    []string   `json:"labelIds"`
    TaskTypes   []string   `json:"taskTypes"`
    DueBefore   *time.Time `json:"dueBefore"`
    DueAfter    *time.Time `json:"dueAfter"`
    Blocked     *bool      `json:"blocked"`
    Overdue     *bool      `json:"overdue"`      
    SearchQuery *string    `json:"searchQuery"`
    Limit       int        `json:"limit"`
    Offset      int        `json:"offset"`
    SortBy      string     `json:"sortBy"`
    SortOrder   string     `json:"sortOrder"`
}




// ============================================
// Label DTOs
// ============================================

type CreateLabelRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=50"`
	Color string `json:"color" binding:"required,hexcolor"`
}

type UpdateLabelRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
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

// ============================================
// Checklist DTOs (NEW - Phase 1)
// ============================================

type UpdateChecklistItemRequest struct {
	Text      *string `json:"text"`
	IsChecked *bool   `json:"isChecked"`
	Position  *int    `json:"position"`
}


// ============================================
// Time Tracking DTOs
// ============================================

type StartTimerRequest struct {
	Description *string `json:"description"`
}

type StopTimerRequest struct {
	Description *string `json:"description"`
}

type AddManualTimeRequest struct {
	StartTime       time.Time `json:"startTime" binding:"required"`
	EndTime         time.Time `json:"endTime" binding:"required"`
	Description     *string   `json:"description"`
	DurationSeconds *int      `json:"durationSeconds"` // Optional, will be calculated if not provided
}

type TimeTrackingResponse struct {
	ID              string        `json:"id"`
	TaskID          string        `json:"taskId"`
	UserID          string        `json:"userId"`
	StartTime       time.Time     `json:"startTime"`
	EndTime         *time.Time    `json:"endTime"`
	DurationSeconds *int          `json:"durationSeconds"`
	Description     *string       `json:"description"`
	IsManual        bool          `json:"isManual"`
	User            *UserResponse `json:"user,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

type TimeTrackingSummaryResponse struct {
	TaskID          string `json:"taskId"`
	TotalSeconds    int    `json:"totalSeconds"`
	TotalHours      float64 `json:"totalHours"`
	EstimatedHours  *float64 `json:"estimatedHours"`
	RemainingHours  *float64 `json:"remainingHours"`
}

// ============================================
// Common Response Types
// ============================================

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"perPage"`
	TotalPages int         `json:"totalPages"`
}

// ============================================
// Utility Functions
// ============================================

func NewPaginatedResponse(data interface{}, total, page, perPage int) PaginatedResponse {
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}
	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}


// Add to models package


// Member requests
type AddMemberRequest struct {
	UserID string `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// Member responses
type UnifiedMemberResponse struct {
	ID            string        `json:"id"`
	EntityType    string        `json:"entityType"`    // "workspace", "space", "folder", "project"
	EntityID      string        `json:"entityId"`
	UserID        string        `json:"userId"`
	Role          string        `json:"role"`
	JoinedAt      time.Time     `json:"joinedAt"`
	IsInherited   bool          `json:"isInherited"`   // True if access from parent
	InheritedFrom string        `json:"inheritedFrom"` // "workspace", "space", "folder" or empty
	User          *UserResponse `json:"user,omitempty"`
}

type AccessCheckResponse struct {
	HasAccess     bool   `json:"hasAccess"`
	IsDirect      bool   `json:"isDirect"`
	InheritedFrom string `json:"inheritedFrom,omitempty"`
}

type AccessLevelResponse struct {
	Role          string `json:"role"`
	IsDirect      bool   `json:"isDirect"`
	InheritedFrom string `json:"inheritedFrom,omitempty"`
}

type UserAccessMapResponse struct {
	Workspaces []string `json:"workspaces"`
	Spaces     []string `json:"spaces"`
	Folders    []string `json:"folders"`
	Projects   []string `json:"projects"`
}