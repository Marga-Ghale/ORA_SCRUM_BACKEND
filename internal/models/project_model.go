package models

import "time"

// Request models
type CreateProjectRequest struct {
	Name        string  `json:"name" binding:"required"`
	Key         string  `json:"key" binding:"required"`
	FolderID    *string `json:"folder_id"` // Optional - can be null
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	Color       *string `json:"color"`
	LeadID      *string `json:"lead_id"` // Optional - project lead
}

type UpdateProjectRequest struct {
	Name        *string  `json:"name"`
	Key         *string  `json:"key"`
	FolderID    **string `json:"folder_id"` // Double pointer to allow setting to null
	Description *string  `json:"description"`
	Icon        *string  `json:"icon"`
	Color       *string  `json:"color"`
	LeadID      *string  `json:"lead_id"`
}

// Response models
type ProjectResponse struct {
    ID           string   `json:"id"`
    SpaceID      string   `json:"space_id"`
    FolderID     *string  `json:"folder_id,omitempty"`  // <-- must include this
    Name         string   `json:"name"`
    Key          string   `json:"key"`
    Description  *string  `json:"description,omitempty"`
    Icon         *string  `json:"icon,omitempty"`
    Color        *string  `json:"color,omitempty"`
    LeadID       *string  `json:"lead_id,omitempty"`
    Visibility   *string  `json:"visibility,omitempty"`
    AllowedUsers []string `json:"allowed_users,omitempty"`
    AllowedTeams []string `json:"allowed_teams,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
