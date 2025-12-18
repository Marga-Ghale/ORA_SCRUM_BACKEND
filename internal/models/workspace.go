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