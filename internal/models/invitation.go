package models

import (
	"time"
)

// InvitationType represents what the invitation is for
type InvitationType string

const (
	InvitationTypeWorkspace InvitationType = "workspace"
	InvitationTypeSpace     InvitationType = "space"
	InvitationTypeFolder    InvitationType = "folder"
	InvitationTypeProject   InvitationType = "project"
	InvitationTypeTeam      InvitationType = "team"
	InvitationTypeTask      InvitationType = "task"
)

// InvitationStatus represents current state of invitation
type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "pending"
	InvitationStatusAccepted  InvitationStatus = "accepted"
	InvitationStatusDeclined  InvitationStatus = "declined"
	InvitationStatusExpired   InvitationStatus = "expired"
	InvitationStatusCancelled InvitationStatus = "cancelled"
	InvitationStatusRevoked   InvitationStatus = "revoked"
)

// InvitationMethod represents how the invitation was sent
type InvitationMethod string

const (
	InvitationMethodEmail  InvitationMethod = "email"
	InvitationMethodLink   InvitationMethod = "link"
	InvitationMethodDirect InvitationMethod = "direct" // For existing users
)

// WorkspaceRole represents user role in workspace
type WorkspaceRole string

const (
	WorkspaceRoleOwner        WorkspaceRole = "owner"
	WorkspaceRoleAdmin        WorkspaceRole = "admin"
	WorkspaceRoleMember       WorkspaceRole = "member"
	WorkspaceRoleLimitedMember WorkspaceRole = "limited_member"
	WorkspaceRoleGuest        WorkspaceRole = "guest"
)

// PermissionLevel for items/locations
type PermissionLevel string

const (
	PermissionFullEdit PermissionLevel = "full_edit"
	PermissionEdit     PermissionLevel = "edit"
	PermissionComment  PermissionLevel = "comment"
	PermissionViewOnly PermissionLevel = "view_only"
)

// Invitation represents an invitation to join workspace/team/project etc.
type Invitation struct {
	ID             string           `json:"id" db:"id"`
	WorkspaceID    string           `json:"workspace_id" db:"workspace_id"`
	Email          string           `json:"email" db:"email"`
	Token          string           `json:"token" db:"token"`
	LinkToken      *string          `json:"link_token,omitempty" db:"link_token"` // For shareable links
	Type           InvitationType   `json:"type" db:"type"`
	TargetID       string           `json:"target_id" db:"target_id"`
	TargetName     string           `json:"target_name,omitempty" db:"target_name"` // Denormalized for display
	Role           WorkspaceRole    `json:"role" db:"role"`
	Permission     PermissionLevel  `json:"permission" db:"permission"`
	InvitedByID    string           `json:"invited_by_id" db:"invited_by_id"`
	InvitedByName  string           `json:"invited_by_name,omitempty" db:"invited_by_name"`
	InviteeUserID  *string          `json:"invitee_user_id,omitempty" db:"invitee_user_id"` // If invitee already has account
	Status         InvitationStatus `json:"status" db:"status"`
	Method         InvitationMethod `json:"method" db:"method"`
	Message        *string          `json:"message,omitempty" db:"message"`
	ExpiresAt      *time.Time       `json:"expires_at,omitempty" db:"expires_at"`
	LinkExpiresAt  *time.Time       `json:"link_expires_at,omitempty" db:"link_expires_at"`
	AcceptedAt     *time.Time       `json:"accepted_at,omitempty" db:"accepted_at"`
	DeclinedAt     *time.Time       `json:"declined_at,omitempty" db:"declined_at"`
	ReminderSentAt *time.Time       `json:"reminder_sent_at,omitempty" db:"reminder_sent_at"`
	ReminderCount  int              `json:"reminder_count" db:"reminder_count"`
	MaxUses        *int             `json:"max_uses,omitempty" db:"max_uses"`       // For link invitations
	UseCount       int              `json:"use_count" db:"use_count"`               // How many times link used
	Metadata       *string          `json:"metadata,omitempty" db:"metadata"`       // JSON for additional data
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`
}

// InvitationPermissions for granular control (like ClickUp's guest permissions)
type InvitationPermissions struct {
	ID                 string `json:"id" db:"id"`
	InvitationID       string `json:"invitation_id" db:"invitation_id"`
	CanEditTasks       bool   `json:"can_edit_tasks" db:"can_edit_tasks"`
	CanCreateTasks     bool   `json:"can_create_tasks" db:"can_create_tasks"`
	CanDeleteTasks     bool   `json:"can_delete_tasks" db:"can_delete_tasks"`
	CanComment         bool   `json:"can_comment" db:"can_comment"`
	CanCreateSubtasks  bool   `json:"can_create_subtasks" db:"can_create_subtasks"`
	CanAssignTasks     bool   `json:"can_assign_tasks" db:"can_assign_tasks"`
	CanSeeTimeSpent    bool   `json:"can_see_time_spent" db:"can_see_time_spent"`
	CanTrackTime       bool   `json:"can_track_time" db:"can_track_time"`
	CanAddTags         bool   `json:"can_add_tags" db:"can_add_tags"`
	CanCreateViews     bool   `json:"can_create_views" db:"can_create_views"`
	CanInviteOthers    bool   `json:"can_invite_others" db:"can_invite_others"`
	CanManageSprints   bool   `json:"can_manage_sprints" db:"can_manage_sprints"`
	CanViewReports     bool   `json:"can_view_reports" db:"can_view_reports"`
	CanExport          bool   `json:"can_export" db:"can_export"`
	CustomPermissions  *string `json:"custom_permissions,omitempty" db:"custom_permissions"` // JSON
}

// InvitationActivity logs all invitation-related events
type InvitationActivity struct {
	ID           string    `json:"id" db:"id"`
	InvitationID string    `json:"invitation_id" db:"invitation_id"`
	Action       string    `json:"action" db:"action"` // created, sent, resent, accepted, declined, expired, cancelled, revoked
	ActorID      *string   `json:"actor_id,omitempty" db:"actor_id"`
	ActorType    string    `json:"actor_type" db:"actor_type"` // user, system
	IPAddress    *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string   `json:"user_agent,omitempty" db:"user_agent"`
	Details      *string   `json:"details,omitempty" db:"details"` // JSON
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// InvitationLinkSettings for shareable invitation links
type InvitationLinkSettings struct {
	ID               string          `json:"id" db:"id"`
	WorkspaceID      string          `json:"workspace_id" db:"workspace_id"`
	LinkToken        string          `json:"link_token" db:"link_token"`
	Type             InvitationType  `json:"type" db:"type"`
	TargetID         string          `json:"target_id" db:"target_id"`
	DefaultRole      WorkspaceRole   `json:"default_role" db:"default_role"`
	DefaultPermission PermissionLevel `json:"default_permission" db:"default_permission"`
	IsActive         bool            `json:"is_active" db:"is_active"`
	RequiresApproval bool            `json:"requires_approval" db:"requires_approval"`
	AllowedDomains   *string         `json:"allowed_domains,omitempty" db:"allowed_domains"` // JSON array
	BlockedDomains   *string         `json:"blocked_domains,omitempty" db:"blocked_domains"` // JSON array
	MaxUses          *int            `json:"max_uses,omitempty" db:"max_uses"`
	UseCount         int             `json:"use_count" db:"use_count"`
	ExpiresAt        *time.Time      `json:"expires_at,omitempty" db:"expires_at"`
	CreatedByID      string          `json:"created_by_id" db:"created_by_id"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// AccessRequest for users requesting access to resources
type AccessRequest struct {
	ID           string           `json:"id" db:"id"`
	WorkspaceID  string           `json:"workspace_id" db:"workspace_id"`
	RequesterID  string           `json:"requester_id" db:"requester_id"`
	Email        string           `json:"email" db:"email"`
	Type         InvitationType   `json:"type" db:"type"`
	TargetID     string           `json:"target_id" db:"target_id"`
	Message      *string          `json:"message,omitempty" db:"message"`
	Status       string           `json:"status" db:"status"` // pending, approved, denied
	ProcessedBy  *string          `json:"processed_by,omitempty" db:"processed_by"`
	ProcessedAt  *time.Time       `json:"processed_at,omitempty" db:"processed_at"`
	DenialReason *string          `json:"denial_reason,omitempty" db:"denial_reason"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
}

// BulkInvitation for batch inviting users
type BulkInvitation struct {
	ID             string           `json:"id" db:"id"`
	WorkspaceID    string           `json:"workspace_id" db:"workspace_id"`
	InvitedByID    string           `json:"invited_by_id" db:"invited_by_id"`
	Type           InvitationType   `json:"type" db:"type"`
	TargetID       string           `json:"target_id" db:"target_id"`
	Role           WorkspaceRole    `json:"role" db:"role"`
	TotalCount     int              `json:"total_count" db:"total_count"`
	SuccessCount   int              `json:"success_count" db:"success_count"`
	FailedCount    int              `json:"failed_count" db:"failed_count"`
	Status         string           `json:"status" db:"status"` // processing, completed, failed
	FailedEmails   *string          `json:"failed_emails,omitempty" db:"failed_emails"` // JSON
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
}

// Helper methods

func (i *Invitation) IsExpired() bool {
	if i.ExpiresAt == nil {
		return false // Never expires
	}
	return time.Now().After(*i.ExpiresAt)
}

func (i *Invitation) IsLinkExpired() bool {
	if i.LinkExpiresAt == nil {
		return false
	}
	return time.Now().After(*i.LinkExpiresAt)
}

func (i *Invitation) CanAccept() bool {
	return i.Status == InvitationStatusPending && !i.IsExpired()
}

func (i *Invitation) CanResend() bool {
	return i.Status == InvitationStatusPending
}

func (i *Invitation) CanCancel() bool {
	return i.Status == InvitationStatusPending
}

func (l *InvitationLinkSettings) IsValid() bool {
	if !l.IsActive {
		return false
	}
	if l.ExpiresAt != nil && time.Now().After(*l.ExpiresAt) {
		return false
	}
	if l.MaxUses != nil && l.UseCount >= *l.MaxUses {
		return false
	}
	return true
}

// ValidRoles returns valid roles for invitation type
func ValidRolesForType(t InvitationType) []WorkspaceRole {
	switch t {
	case InvitationTypeWorkspace:
		return []WorkspaceRole{WorkspaceRoleAdmin, WorkspaceRoleMember, WorkspaceRoleLimitedMember, WorkspaceRoleGuest}
	case InvitationTypeSpace, InvitationTypeFolder, InvitationTypeProject:
		return []WorkspaceRole{WorkspaceRoleMember, WorkspaceRoleLimitedMember, WorkspaceRoleGuest}
	case InvitationTypeTeam:
		return []WorkspaceRole{WorkspaceRoleMember}
	case InvitationTypeTask:
		return []WorkspaceRole{WorkspaceRoleGuest, WorkspaceRoleLimitedMember}
	default:
		return []WorkspaceRole{WorkspaceRoleMember}
	}
}

// DefaultPermissionForRole returns default permission for a role
func DefaultPermissionForRole(role WorkspaceRole) PermissionLevel {
	switch role {
	case WorkspaceRoleOwner, WorkspaceRoleAdmin:
		return PermissionFullEdit
	case WorkspaceRoleMember:
		return PermissionEdit
	case WorkspaceRoleLimitedMember:
		return PermissionComment
	case WorkspaceRoleGuest:
		return PermissionViewOnly
	default:
		return PermissionViewOnly
	}
}