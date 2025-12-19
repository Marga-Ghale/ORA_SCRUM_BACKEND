package models

import "time"

// ============================================
// Member Management Models
// ============================================

type AddMemberRequest struct {
    UserID string `json:"userId" binding:"required"`
    Role   string `json:"role" binding:"required"`
}

type InviteMemberRequest struct {
    Email string `json:"email" binding:"required,email"`
    Role  string `json:"role" binding:"required"`
}

type UpdateMemberRoleRequest struct {
    Role string `json:"role" binding:"required"`
}

type UnifiedMemberResponse struct {
    ID            string        `json:"id"`
    EntityType    string        `json:"entityType"`
    EntityID      string        `json:"entityId"`
    UserID        string        `json:"userId"`
    Role          string        `json:"role"`
    JoinedAt      time.Time     `json:"joinedAt"`
    IsInherited   bool          `json:"isInherited"`
    InheritedFrom string        `json:"inheritedFrom,omitempty"`
    User          *UserResponse `json:"user,omitempty"`
}

