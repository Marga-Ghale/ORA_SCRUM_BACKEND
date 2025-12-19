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
