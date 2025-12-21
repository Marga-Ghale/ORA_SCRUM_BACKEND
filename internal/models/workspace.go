package models

import "time"

// Request models
type CreateWorkspaceRequest struct {
	Name         string   `json:"name" binding:"required"`
	Description  *string  `json:"description"`
	Icon         *string  `json:"icon"`
	Color        *string  `json:"color"`
	Visibility   *string  `json:"visibility"`
	AllowedUsers []string `json:"allowed_users"`
	AllowedTeams []string `json:"allowed_teams"`
}

type UpdateWorkspaceRequest struct {
	Name         *string   `json:"name"`
	Description  *string   `json:"description"`
	Icon         *string   `json:"icon"`
	Color        *string   `json:"color"`
	Visibility   *string   `json:"visibility"`
	AllowedUsers *[]string `json:"allowed_users"`
	AllowedTeams *[]string `json:"allowed_teams"`
}

// Response models
type WorkspaceResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Icon         string    `json:"icon,omitempty"`
	Color        string    `json:"color,omitempty"`
	OwnerID      string    `json:"owner_id"`
	Visibility   string    `json:"visibility,omitempty"`
	AllowedUsers []string  `json:"allowed_users"`
	AllowedTeams []string  `json:"allowed_teams"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}


// ============================================
// Member Management Models
// ============================================

type AddWorkspaceMemberRequest struct {
	UserID string `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required"` // owner, admin, member, viewer
}

type AddWorkspaceMemberByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}


type WorkspaceMemberResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	UserID      string    `json:"userId"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
	User        *UserInfo `json:"user,omitempty"`
}

type UserInfo struct {
	ID     string  `json:"id"`
	Email  string  `json:"email"`
	Name   string  `json:"name"`
	Avatar *string `json:"avatar,omitempty"`
	Status *string `json:"status,omitempty"`
}