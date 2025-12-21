// ============================================
// FILE: internal/models/project.go
// ============================================
package models

import "time"

// Request models
type CreateProjectRequest struct {
    Name        string  `json:"name" binding:"required"`
    Key         string  `json:"key" binding:"required"`
    FolderID    *string `json:"folderId"`      // ✅ Change to camelCase
    Description *string `json:"description"`
    Icon        *string `json:"icon"`
    Color       *string `json:"color"`
    LeadID      *string `json:"leadId"`        // ✅ Also change this
}

type UpdateProjectRequest struct {
    Name        *string  `json:"name"`
    Key         *string  `json:"key"`
    FolderID    **string `json:"folderId"`     // ✅ Change to camelCase
    Description *string  `json:"description"`
    Icon        *string  `json:"icon"`
    Color       *string  `json:"color"`
    LeadID      *string  `json:"leadId"`       // ✅ Also change this
}
type ProjectResponse struct {
	ID           string     `json:"id"`
	SpaceID      string     `json:"spaceId"`
	FolderID     *string    `json:"folderId,omitempty"`
	Name         string     `json:"name"`
	Key          string     `json:"key"`
	Description  *string    `json:"description,omitempty"`
	Icon         *string    `json:"icon,omitempty"`
	Color        *string    `json:"color,omitempty"`
	LeadID       *string    `json:"leadId,omitempty"`
	Visibility   *string    `json:"visibility,omitempty"`
	AllowedUsers []string   `json:"allowedUsers"`  
	AllowedTeams []string   `json:"allowedTeams"`  
	CreatedBy    *string    `json:"createdBy,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}