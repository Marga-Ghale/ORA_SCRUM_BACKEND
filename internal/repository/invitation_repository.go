package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	InvitationMethodDirect InvitationMethod = "direct"
)

// WorkspaceRole represents user role in workspace
type WorkspaceRole string

const (
	WorkspaceRoleOwner         WorkspaceRole = "owner"
	WorkspaceRoleAdmin         WorkspaceRole = "admin"
	WorkspaceRoleMember        WorkspaceRole = "member"
	WorkspaceRoleLimitedMember WorkspaceRole = "limited_member"
	WorkspaceRoleGuest         WorkspaceRole = "guest"
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
	LinkToken      *string          `json:"link_token,omitempty" db:"link_token"`
	Type           InvitationType   `json:"type" db:"type"`
	TargetID       string           `json:"target_id" db:"target_id"`
	TargetName     string           `json:"target_name,omitempty" db:"target_name"`
	Role           WorkspaceRole    `json:"role" db:"role"`
	Permission     PermissionLevel  `json:"permission" db:"permission"`
	InvitedByID    string           `json:"invited_by_id" db:"invited_by_id"`
	InvitedByName  string           `json:"invited_by_name,omitempty" db:"invited_by_name"`
	InviteeUserID  *string          `json:"invitee_user_id,omitempty" db:"invitee_user_id"`
	Status         InvitationStatus `json:"status" db:"status"`
	Method         InvitationMethod `json:"method" db:"method"`
	Message        *string          `json:"message,omitempty" db:"message"`
	ExpiresAt      *time.Time       `json:"expires_at,omitempty" db:"expires_at"`
	LinkExpiresAt  *time.Time       `json:"link_expires_at,omitempty" db:"link_expires_at"`
	AcceptedAt     *time.Time       `json:"accepted_at,omitempty" db:"accepted_at"`
	DeclinedAt     *time.Time       `json:"declined_at,omitempty" db:"declined_at"`
	ReminderSentAt *time.Time       `json:"reminder_sent_at,omitempty" db:"reminder_sent_at"`
	ReminderCount  int              `json:"reminder_count" db:"reminder_count"`
	MaxUses        *int             `json:"max_uses,omitempty" db:"max_uses"`
	UseCount       int              `json:"use_count" db:"use_count"`
	Metadata       *string          `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`
}

// InvitationPermissions for granular control
type InvitationPermissions struct {
	ID                string  `json:"id" db:"id"`
	InvitationID      string  `json:"invitation_id" db:"invitation_id"`
	CanEditTasks      bool    `json:"can_edit_tasks" db:"can_edit_tasks"`
	CanCreateTasks    bool    `json:"can_create_tasks" db:"can_create_tasks"`
	CanDeleteTasks    bool    `json:"can_delete_tasks" db:"can_delete_tasks"`
	CanComment        bool    `json:"can_comment" db:"can_comment"`
	CanCreateSubtasks bool    `json:"can_create_subtasks" db:"can_create_subtasks"`
	CanAssignTasks    bool    `json:"can_assign_tasks" db:"can_assign_tasks"`
	CanSeeTimeSpent   bool    `json:"can_see_time_spent" db:"can_see_time_spent"`
	CanTrackTime      bool    `json:"can_track_time" db:"can_track_time"`
	CanAddTags        bool    `json:"can_add_tags" db:"can_add_tags"`
	CanCreateViews    bool    `json:"can_create_views" db:"can_create_views"`
	CanInviteOthers   bool    `json:"can_invite_others" db:"can_invite_others"`
	CanManageSprints  bool    `json:"can_manage_sprints" db:"can_manage_sprints"`
	CanViewReports    bool    `json:"can_view_reports" db:"can_view_reports"`
	CanExport         bool    `json:"can_export" db:"can_export"`
	CustomPermissions *string `json:"custom_permissions,omitempty" db:"custom_permissions"`
}

// InvitationActivity logs all invitation-related events
type InvitationActivity struct {
	ID           string    `json:"id" db:"id"`
	InvitationID string    `json:"invitation_id" db:"invitation_id"`
	Action       string    `json:"action" db:"action"`
	ActorID      *string   `json:"actor_id,omitempty" db:"actor_id"`
	ActorType    string    `json:"actor_type" db:"actor_type"`
	IPAddress    *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string   `json:"user_agent,omitempty" db:"user_agent"`
	Details      *string   `json:"details,omitempty" db:"details"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// InvitationLinkSettings for shareable invitation links
type InvitationLinkSettings struct {
	ID                string          `json:"id" db:"id"`
	WorkspaceID       string          `json:"workspace_id" db:"workspace_id"`
	LinkToken         string          `json:"link_token" db:"link_token"`
	Type              InvitationType  `json:"type" db:"type"`
	TargetID          string          `json:"target_id" db:"target_id"`
	DefaultRole       WorkspaceRole   `json:"default_role" db:"default_role"`
	DefaultPermission PermissionLevel `json:"default_permission" db:"default_permission"`
	IsActive          bool            `json:"is_active" db:"is_active"`
	RequiresApproval  bool            `json:"requires_approval" db:"requires_approval"`
	AllowedDomains    *string         `json:"allowed_domains,omitempty" db:"allowed_domains"`
	BlockedDomains    *string         `json:"blocked_domains,omitempty" db:"blocked_domains"`
	MaxUses           *int            `json:"max_uses,omitempty" db:"max_uses"`
	UseCount          int             `json:"use_count" db:"use_count"`
	ExpiresAt         *time.Time      `json:"expires_at,omitempty" db:"expires_at"`
	CreatedByID       string          `json:"created_by_id" db:"created_by_id"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// AccessRequest for users requesting access to resources
type AccessRequest struct {
	ID           string         `json:"id" db:"id"`
	WorkspaceID  string         `json:"workspace_id" db:"workspace_id"`
	RequesterID  string         `json:"requester_id" db:"requester_id"`
	Email        string         `json:"email" db:"email"`
	Type         InvitationType `json:"type" db:"type"`
	TargetID     string         `json:"target_id" db:"target_id"`
	Message      *string        `json:"message,omitempty" db:"message"`
	Status       string         `json:"status" db:"status"`
	ProcessedBy  *string        `json:"processed_by,omitempty" db:"processed_by"`
	ProcessedAt  *time.Time     `json:"processed_at,omitempty" db:"processed_at"`
	DenialReason *string        `json:"denial_reason,omitempty" db:"denial_reason"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// BulkInvitationResult for batch inviting results
type BulkInvitationResult struct {
	ID           string         `json:"id" db:"id"`
	WorkspaceID  string         `json:"workspace_id" db:"workspace_id"`
	InvitedByID  string         `json:"invited_by_id" db:"invited_by_id"`
	Type         InvitationType `json:"type" db:"type"`
	TargetID     string         `json:"target_id" db:"target_id"`
	Role         WorkspaceRole  `json:"role" db:"role"`
	TotalCount   int            `json:"total_count" db:"total_count"`
	SuccessCount int            `json:"success_count" db:"success_count"`
	FailedCount  int            `json:"failed_count" db:"failed_count"`
	SkippedCount int            `json:"skipped_count" db:"skipped_count"`
	Status       string         `json:"status" db:"status"`
	FailedEmails *string        `json:"failed_emails,omitempty" db:"failed_emails"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
}

// InvitationFilter for querying invitations
type InvitationFilter struct {
	WorkspaceID    string
	Email          string
	Status         InvitationStatus
	Statuses       []InvitationStatus
	Type           InvitationType
	Types          []InvitationType
	TargetID       string
	InvitedByID    string
	InviteeUserID  string
	Method         InvitationMethod
	DateFrom       *time.Time
	DateTo         *time.Time
	IncludeExpired bool
	Limit          int
	Offset         int
	OrderBy        string
	OrderDir       string
}

// InvitationStats for analytics
type InvitationStats struct {
	TotalInvitations   int     `json:"total_invitations"`
	PendingCount       int     `json:"pending_count"`
	AcceptedCount      int     `json:"accepted_count"`
	DeclinedCount      int     `json:"declined_count"`
	ExpiredCount       int     `json:"expired_count"`
	CancelledCount     int     `json:"cancelled_count"`
	AcceptanceRate     float64 `json:"acceptance_rate"`
	AvgTimeToAcceptHrs float64 `json:"avg_time_to_accept_hrs"`
}

// Helper methods
func (i *Invitation) IsExpired() bool {
	if i.ExpiresAt == nil {
		return false
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

func (l *InvitationLinkSettings) CheckDomain(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := strings.ToLower(parts[1])

	if l.BlockedDomains != nil {
		var blocked []string
		if err := json.Unmarshal([]byte(*l.BlockedDomains), &blocked); err == nil {
			for _, b := range blocked {
				if strings.ToLower(b) == domain {
					return false
				}
			}
		}
	}

	if l.AllowedDomains != nil {
		var allowed []string
		if err := json.Unmarshal([]byte(*l.AllowedDomains), &allowed); err == nil {
			if len(allowed) > 0 {
				for _, a := range allowed {
					if strings.ToLower(a) == domain {
						return true
					}
				}
				return false
			}
		}
	}

	return true
}

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

// InvitationRepository interface
type InvitationRepository interface {
	Create(ctx context.Context, inv *Invitation) error
	CreateBatch(ctx context.Context, invitations []*Invitation) ([]string, []error)
	CreateWithPermissions(ctx context.Context, inv *Invitation, perms *InvitationPermissions) error

	FindByID(ctx context.Context, id string) (*Invitation, error)
	FindByToken(ctx context.Context, token string) (*Invitation, error)
	FindByLinkToken(ctx context.Context, linkToken string) (*Invitation, error)
	FindByEmail(ctx context.Context, email string) ([]*Invitation, error)
	FindByEmailInWorkspace(ctx context.Context, email, workspaceID string) ([]*Invitation, error)
	FindByInviteeUserID(ctx context.Context, userID string) ([]*Invitation, error)
	FindPendingByEmail(ctx context.Context, email string) ([]*Invitation, error)
	FindPendingByTarget(ctx context.Context, targetType InvitationType, targetID string) ([]*Invitation, error)
	FindByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*Invitation, int, error)
	FindByInviter(ctx context.Context, inviterID string, limit, offset int) ([]*Invitation, error)
	FindByFilter(ctx context.Context, filter *InvitationFilter) ([]*Invitation, int, error)
	FindPendingForReminder(ctx context.Context, minAge time.Duration, maxReminders int) ([]*Invitation, error)

	ExistsPendingForEmail(ctx context.Context, email string, targetType InvitationType, targetID string) (bool, error)
	ExistsPendingForUser(ctx context.Context, userID string, targetType InvitationType, targetID string) (bool, error)
	CountPendingByTarget(ctx context.Context, targetType InvitationType, targetID string) (int, error)
	CountByWorkspace(ctx context.Context, workspaceID string) (int, error)

	Update(ctx context.Context, inv *Invitation) error
	UpdateStatus(ctx context.Context, id string, status InvitationStatus) error
	UpdateRole(ctx context.Context, id string, role WorkspaceRole) error
	UpdatePermission(ctx context.Context, id string, permission PermissionLevel) error

	MarkAccepted(ctx context.Context, id string, userID string) error
	MarkDeclined(ctx context.Context, id string) error
	MarkExpired(ctx context.Context, id string) error
	MarkCancelled(ctx context.Context, id string) error
	MarkRevoked(ctx context.Context, id string) error

	UpdateReminderSent(ctx context.Context, id string) error
	ResetReminderCount(ctx context.Context, id string) error

	IncrementLinkUseCount(ctx context.Context, id string) error
	UpdateLinkExpiry(ctx context.Context, id string, expiresAt time.Time) error
	RegenerateToken(ctx context.Context, id string) (string, error)

	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) (int64, error)
	DeleteByTarget(ctx context.Context, targetType InvitationType, targetID string) (int64, error)
	DeleteByWorkspace(ctx context.Context, workspaceID string) (int64, error)

	GetStatsByWorkspace(ctx context.Context, workspaceID string) (*InvitationStats, error)
	GetStatsByTarget(ctx context.Context, targetType InvitationType, targetID string) (*InvitationStats, error)

	LogActivity(ctx context.Context, activity *InvitationActivity) error
	GetActivityByInvitation(ctx context.Context, invitationID string) ([]*InvitationActivity, error)

	GetPermissions(ctx context.Context, invitationID string) (*InvitationPermissions, error)
	CreatePermissions(ctx context.Context, perms *InvitationPermissions) error
	UpdatePermissions(ctx context.Context, perms *InvitationPermissions) error
	DeletePermissions(ctx context.Context, invitationID string) error

	CreateLinkSettings(ctx context.Context, settings *InvitationLinkSettings) error
	GetLinkSettings(ctx context.Context, id string) (*InvitationLinkSettings, error)
	GetLinkSettingsByToken(ctx context.Context, linkToken string) (*InvitationLinkSettings, error)
	GetLinkSettingsByTarget(ctx context.Context, targetType InvitationType, targetID string) ([]*InvitationLinkSettings, error)
	UpdateLinkSettings(ctx context.Context, settings *InvitationLinkSettings) error
	DeactivateLinkSettings(ctx context.Context, id string) error
	IncrementLinkSettingsUseCount(ctx context.Context, id string) error
	DeleteLinkSettings(ctx context.Context, id string) error

	CreateAccessRequest(ctx context.Context, req *AccessRequest) error
	GetAccessRequest(ctx context.Context, id string) (*AccessRequest, error)
	GetAccessRequestsByTarget(ctx context.Context, targetType InvitationType, targetID string, status string) ([]*AccessRequest, error)
	GetAccessRequestsByRequester(ctx context.Context, requesterID string) ([]*AccessRequest, error)
	UpdateAccessRequestStatus(ctx context.Context, id, status string, processedBy *string, denialReason *string) error
	DeleteAccessRequest(ctx context.Context, id string) error

	CreateBulkResult(ctx context.Context, result *BulkInvitationResult) error
	GetBulkResult(ctx context.Context, id string) (*BulkInvitationResult, error)
	UpdateBulkResult(ctx context.Context, result *BulkInvitationResult) error
}

type pgInvitationRepository struct {
	pool *pgxpool.Pool
}

func NewInvitationRepository(pool *pgxpool.Pool) InvitationRepository {
	return &pgInvitationRepository{pool: pool}
}

func (r *pgInvitationRepository) Create(ctx context.Context, inv *Invitation) error {
	if inv.ID == "" {
		inv.ID = uuid.New().String()
	}
	if inv.Token == "" {
		inv.Token = uuid.New().String()
	}
	if inv.Status == "" {
		inv.Status = InvitationStatusPending
	}
	if inv.Method == "" {
		inv.Method = InvitationMethodEmail
	}
	if inv.Permission == "" {
		inv.Permission = DefaultPermissionForRole(inv.Role)
	}

	query := `
		INSERT INTO invitations (
			id, workspace_id, email, token, link_token, type, target_id, target_name,
			role, permission, invited_by_id, invited_by_name, invitee_user_id,
			status, method, message, expires_at, link_expires_at,
			reminder_count, max_uses, use_count, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, NOW(), NOW()
		) RETURNING created_at, updated_at
	`

	return r.pool.QueryRow(ctx, query,
		inv.ID, inv.WorkspaceID, inv.Email, inv.Token, inv.LinkToken,
		inv.Type, inv.TargetID, inv.TargetName, inv.Role, inv.Permission,
		inv.InvitedByID, inv.InvitedByName, inv.InviteeUserID, inv.Status,
		inv.Method, inv.Message, inv.ExpiresAt, inv.LinkExpiresAt,
		inv.ReminderCount, inv.MaxUses, inv.UseCount, inv.Metadata,
	).Scan(&inv.CreatedAt, &inv.UpdatedAt)
}

func (r *pgInvitationRepository) CreateBatch(ctx context.Context, invitations []*Invitation) ([]string, []error) {
	ids := make([]string, len(invitations))
	errors := make([]error, len(invitations))

	for i, inv := range invitations {
		err := r.Create(ctx, inv)
		if err != nil {
			errors[i] = err
		} else {
			ids[i] = inv.ID
		}
	}

	return ids, errors
}

func (r *pgInvitationRepository) CreateWithPermissions(ctx context.Context, inv *Invitation, perms *InvitationPermissions) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.Create(ctx, inv); err != nil {
		return err
	}

	if perms != nil {
		perms.InvitationID = inv.ID
		if err := r.CreatePermissions(ctx, perms); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *pgInvitationRepository) FindByID(ctx context.Context, id string) (*Invitation, error) {
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations WHERE id = $1
	`
	return r.scanOne(ctx, query, id)
}

func (r *pgInvitationRepository) FindByToken(ctx context.Context, token string) (*Invitation, error) {
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations WHERE token = $1
	`
	return r.scanOne(ctx, query, token)
}

func (r *pgInvitationRepository) FindByLinkToken(ctx context.Context, linkToken string) (*Invitation, error) {
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations WHERE link_token = $1
	`
	return r.scanOne(ctx, query, linkToken)
}

func (r *pgInvitationRepository) FindByEmail(ctx context.Context, email string) ([]*Invitation, error) {
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations WHERE LOWER(email) = LOWER($1)
		ORDER BY created_at DESC
	`
	return r.scanMany(ctx, query, email)
}

func (r *pgInvitationRepository) FindByEmailInWorkspace(ctx context.Context, email, workspaceID string) ([]*Invitation, error) {
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations 
		WHERE LOWER(email) = LOWER($1) AND workspace_id = $2
		ORDER BY created_at DESC
	`
	return r.scanMany(ctx, query, email, workspaceID)
}

func (r *pgInvitationRepository) FindByInviteeUserID(ctx context.Context, userID string) ([]*Invitation, error) {
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations WHERE invitee_user_id = $1
		ORDER BY created_at DESC
	`
	return r.scanMany(ctx, query, userID)
}

func (r *pgInvitationRepository) FindPendingByEmail(ctx context.Context, email string) ([]*Invitation, error) {
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations 
		WHERE LOWER(email) = LOWER($1) AND status = 'pending'
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
	`
	return r.scanMany(ctx, query, email)
}

func (r *pgInvitationRepository) FindPendingByTarget(ctx context.Context, targetType InvitationType, targetID string) ([]*Invitation, error) {
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations 
		WHERE type = $1 AND target_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
	`
	return r.scanMany(ctx, query, targetType, targetID)
}

func (r *pgInvitationRepository) FindByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*Invitation, int, error) {
	countQuery := `SELECT COUNT(*) FROM invitations WHERE workspace_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, workspaceID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	invitations, err := r.scanMany(ctx, query, workspaceID, limit, offset)
	return invitations, total, err
}

func (r *pgInvitationRepository) FindByInviter(ctx context.Context, inviterID string, limit, offset int) ([]*Invitation, error) {
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations WHERE invited_by_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.scanMany(ctx, query, inviterID, limit, offset)
}

func (r *pgInvitationRepository) FindByFilter(ctx context.Context, filter *InvitationFilter) ([]*Invitation, int, error) {
	baseQuery := `FROM invitations WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if filter.WorkspaceID != "" {
		baseQuery += fmt.Sprintf(" AND workspace_id = $%d", argNum)
		args = append(args, filter.WorkspaceID)
		argNum++
	}
	if filter.Email != "" {
		baseQuery += fmt.Sprintf(" AND LOWER(email) = LOWER($%d)", argNum)
		args = append(args, filter.Email)
		argNum++
	}
	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filter.Status)
		argNum++
	}
	if len(filter.Statuses) > 0 {
		placeholders := make([]string, len(filter.Statuses))
		for i, s := range filter.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, s)
			argNum++
		}
		baseQuery += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}
	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND type = $%d", argNum)
		args = append(args, filter.Type)
		argNum++
	}
	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, t)
			argNum++
		}
		baseQuery += fmt.Sprintf(" AND type IN (%s)", strings.Join(placeholders, ","))
	}
	if filter.TargetID != "" {
		baseQuery += fmt.Sprintf(" AND target_id = $%d", argNum)
		args = append(args, filter.TargetID)
		argNum++
	}
	if filter.InvitedByID != "" {
		baseQuery += fmt.Sprintf(" AND invited_by_id = $%d", argNum)
		args = append(args, filter.InvitedByID)
		argNum++
	}
	if filter.InviteeUserID != "" {
		baseQuery += fmt.Sprintf(" AND invitee_user_id = $%d", argNum)
		args = append(args, filter.InviteeUserID)
		argNum++
	}
	if filter.Method != "" {
		baseQuery += fmt.Sprintf(" AND method = $%d", argNum)
		args = append(args, filter.Method)
		argNum++
	}
	if filter.DateFrom != nil {
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argNum)
		args = append(args, filter.DateFrom)
		argNum++
	}
	if filter.DateTo != nil {
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argNum)
		args = append(args, filter.DateTo)
		argNum++
	}
	if !filter.IncludeExpired {
		baseQuery += " AND (expires_at IS NULL OR expires_at > NOW())"
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := "created_at"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	orderDir := "DESC"
	if filter.OrderDir != "" {
		orderDir = filter.OrderDir
	}

	selectQuery := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at ` + baseQuery

	selectQuery += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	if filter.Limit > 0 {
		selectQuery += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}
	if filter.Offset > 0 {
		selectQuery += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	invitations, err := r.scanMany(ctx, selectQuery, args...)
	return invitations, total, err
}

func (r *pgInvitationRepository) FindPendingForReminder(ctx context.Context, minAge time.Duration, maxReminders int) ([]*Invitation, error) {
	cutoff := time.Now().Add(-minAge)
	query := `
		SELECT id, workspace_id, email, token, link_token, type, target_id, target_name,
			   role, permission, invited_by_id, invited_by_name, invitee_user_id,
			   status, method, message, expires_at, link_expires_at, accepted_at,
			   declined_at, reminder_sent_at, reminder_count, max_uses, use_count,
			   metadata, created_at, updated_at
		FROM invitations 
		WHERE status = 'pending'
		  AND (expires_at IS NULL OR expires_at > NOW())
		  AND reminder_count < $1
		  AND created_at < $2
		  AND (reminder_sent_at IS NULL OR reminder_sent_at < $2)
		ORDER BY created_at ASC
	`
	return r.scanMany(ctx, query, maxReminders, cutoff)
}

func (r *pgInvitationRepository) ExistsPendingForEmail(ctx context.Context, email string, targetType InvitationType, targetID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM invitations 
			WHERE LOWER(email) = LOWER($1) 
			AND type = $2 
			AND target_id = $3 
			AND status = 'pending'
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, email, targetType, targetID).Scan(&exists)
	return exists, err
}

func (r *pgInvitationRepository) ExistsPendingForUser(ctx context.Context, userID string, targetType InvitationType, targetID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM invitations 
			WHERE invitee_user_id = $1 
			AND type = $2 
			AND target_id = $3 
			AND status = 'pending'
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, targetType, targetID).Scan(&exists)
	return exists, err
}

func (r *pgInvitationRepository) CountPendingByTarget(ctx context.Context, targetType InvitationType, targetID string) (int, error) {
	query := `
		SELECT COUNT(*) FROM invitations 
		WHERE type = $1 AND target_id = $2 AND status = 'pending'
	`
	var count int
	err := r.pool.QueryRow(ctx, query, targetType, targetID).Scan(&count)
	return count, err
}

func (r *pgInvitationRepository) CountByWorkspace(ctx context.Context, workspaceID string) (int, error) {
	query := `SELECT COUNT(*) FROM invitations WHERE workspace_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, workspaceID).Scan(&count)
	return count, err
}

func (r *pgInvitationRepository) Update(ctx context.Context, inv *Invitation) error {
	query := `
		UPDATE invitations SET
			email = $2, role = $3, permission = $4, status = $5, message = $6,
			expires_at = $7, link_expires_at = $8, accepted_at = $9, declined_at = $10,
			reminder_sent_at = $11, reminder_count = $12, max_uses = $13, use_count = $14,
			metadata = $15, token = $16, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	return r.pool.QueryRow(ctx, query,
		inv.ID, inv.Email, inv.Role, inv.Permission, inv.Status, inv.Message,
		inv.ExpiresAt, inv.LinkExpiresAt, inv.AcceptedAt, inv.DeclinedAt,
		inv.ReminderSentAt, inv.ReminderCount, inv.MaxUses, inv.UseCount,
		inv.Metadata, inv.Token,
	).Scan(&inv.UpdatedAt)
}

func (r *pgInvitationRepository) UpdateStatus(ctx context.Context, id string, status InvitationStatus) error {
	query := `UPDATE invitations SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, status)
	return err
}

func (r *pgInvitationRepository) UpdateRole(ctx context.Context, id string, role WorkspaceRole) error {
	query := `UPDATE invitations SET role = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, role)
	return err
}

func (r *pgInvitationRepository) UpdatePermission(ctx context.Context, id string, permission PermissionLevel) error {
	query := `UPDATE invitations SET permission = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, permission)
	return err
}

func (r *pgInvitationRepository) MarkAccepted(ctx context.Context, id string, userID string) error {
	query := `
		UPDATE invitations 
		SET status = 'accepted', accepted_at = NOW(), invitee_user_id = $2, updated_at = NOW() 
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, userID)
	return err
}

func (r *pgInvitationRepository) MarkDeclined(ctx context.Context, id string) error {
	query := `UPDATE invitations SET status = 'declined', declined_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) MarkExpired(ctx context.Context, id string) error {
	query := `UPDATE invitations SET status = 'expired', updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) MarkCancelled(ctx context.Context, id string) error {
	query := `UPDATE invitations SET status = 'cancelled', updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) MarkRevoked(ctx context.Context, id string) error {
	query := `UPDATE invitations SET status = 'revoked', updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) UpdateReminderSent(ctx context.Context, id string) error {
	query := `
		UPDATE invitations 
		SET reminder_sent_at = NOW(), reminder_count = reminder_count + 1, updated_at = NOW() 
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) ResetReminderCount(ctx context.Context, id string) error {
	query := `
		UPDATE invitations 
		SET reminder_count = 0, reminder_sent_at = NULL, updated_at = NOW() 
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) IncrementLinkUseCount(ctx context.Context, id string) error {
	query := `UPDATE invitations SET use_count = use_count + 1, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) UpdateLinkExpiry(ctx context.Context, id string, expiresAt time.Time) error {
	query := `UPDATE invitations SET link_expires_at = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, expiresAt)
	return err
}

func (r *pgInvitationRepository) RegenerateToken(ctx context.Context, id string) (string, error) {
	newToken := uuid.New().String()
	query := `UPDATE invitations SET token = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, newToken)
	return newToken, err
}

func (r *pgInvitationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM invitations WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE invitations SET status = 'cancelled', updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM invitations WHERE expires_at < NOW() AND status = 'pending'`
	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *pgInvitationRepository) DeleteByTarget(ctx context.Context, targetType InvitationType, targetID string) (int64, error) {
	query := `DELETE FROM invitations WHERE type = $1 AND target_id = $2`
	result, err := r.pool.Exec(ctx, query, targetType, targetID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *pgInvitationRepository) DeleteByWorkspace(ctx context.Context, workspaceID string) (int64, error) {
	query := `DELETE FROM invitations WHERE workspace_id = $1`
	result, err := r.pool.Exec(ctx, query, workspaceID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *pgInvitationRepository) GetStatsByWorkspace(ctx context.Context, workspaceID string) (*InvitationStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'accepted') as accepted,
			COUNT(*) FILTER (WHERE status = 'declined') as declined,
			COUNT(*) FILTER (WHERE status = 'expired') as expired,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled,
			COALESCE(AVG(EXTRACT(EPOCH FROM (accepted_at - created_at)) / 3600) FILTER (WHERE status = 'accepted'), 0) as avg_time
		FROM invitations WHERE workspace_id = $1
	`
	stats := &InvitationStats{}
	err := r.pool.QueryRow(ctx, query, workspaceID).Scan(
		&stats.TotalInvitations, &stats.PendingCount, &stats.AcceptedCount,
		&stats.DeclinedCount, &stats.ExpiredCount, &stats.CancelledCount,
		&stats.AvgTimeToAcceptHrs,
	)
	if err != nil {
		return nil, err
	}

	if stats.AcceptedCount+stats.DeclinedCount > 0 {
		stats.AcceptanceRate = float64(stats.AcceptedCount) / float64(stats.AcceptedCount+stats.DeclinedCount) * 100
	}
	return stats, nil
}

func (r *pgInvitationRepository) GetStatsByTarget(ctx context.Context, targetType InvitationType, targetID string) (*InvitationStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'accepted') as accepted,
			COUNT(*) FILTER (WHERE status = 'declined') as declined,
			COUNT(*) FILTER (WHERE status = 'expired') as expired,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled,
			COALESCE(AVG(EXTRACT(EPOCH FROM (accepted_at - created_at)) / 3600) FILTER (WHERE status = 'accepted'), 0) as avg_time
		FROM invitations WHERE type = $1 AND target_id = $2
	`
	stats := &InvitationStats{}
	err := r.pool.QueryRow(ctx, query, targetType, targetID).Scan(
		&stats.TotalInvitations, &stats.PendingCount, &stats.AcceptedCount,
		&stats.DeclinedCount, &stats.ExpiredCount, &stats.CancelledCount,
		&stats.AvgTimeToAcceptHrs,
	)
	if err != nil {
		return nil, err
	}

	if stats.AcceptedCount+stats.DeclinedCount > 0 {
		stats.AcceptanceRate = float64(stats.AcceptedCount) / float64(stats.AcceptedCount+stats.DeclinedCount) * 100
	}
	return stats, nil
}

func (r *pgInvitationRepository) LogActivity(ctx context.Context, activity *InvitationActivity) error {
	if activity.ID == "" {
		activity.ID = uuid.New().String()
	}
	query := `
		INSERT INTO invitation_activities (id, invitation_id, action, actor_id, actor_type, ip_address, user_agent, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING created_at
	`
	return r.pool.QueryRow(ctx, query,
		activity.ID, activity.InvitationID, activity.Action, activity.ActorID,
		activity.ActorType, activity.IPAddress, activity.UserAgent, activity.Details,
	).Scan(&activity.CreatedAt)
}

func (r *pgInvitationRepository) GetActivityByInvitation(ctx context.Context, invitationID string) ([]*InvitationActivity, error) {
	query := `
		SELECT id, invitation_id, action, actor_id, actor_type, ip_address, user_agent, details, created_at
		FROM invitation_activities WHERE invitation_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, invitationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*InvitationActivity
	for rows.Next() {
		a := &InvitationActivity{}
		if err := rows.Scan(&a.ID, &a.InvitationID, &a.Action, &a.ActorID, &a.ActorType,
			&a.IPAddress, &a.UserAgent, &a.Details, &a.CreatedAt); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}

func (r *pgInvitationRepository) GetPermissions(ctx context.Context, invitationID string) (*InvitationPermissions, error) {
	query := `
		SELECT id, invitation_id, can_edit_tasks, can_create_tasks, can_delete_tasks,
			   can_comment, can_create_subtasks, can_assign_tasks, can_see_time_spent,
			   can_track_time, can_add_tags, can_create_views, can_invite_others,
			   can_manage_sprints, can_view_reports, can_export, custom_permissions
		FROM invitation_permissions WHERE invitation_id = $1
	`
	p := &InvitationPermissions{}
	err := r.pool.QueryRow(ctx, query, invitationID).Scan(
		&p.ID, &p.InvitationID, &p.CanEditTasks, &p.CanCreateTasks, &p.CanDeleteTasks,
		&p.CanComment, &p.CanCreateSubtasks, &p.CanAssignTasks, &p.CanSeeTimeSpent,
		&p.CanTrackTime, &p.CanAddTags, &p.CanCreateViews, &p.CanInviteOthers,
		&p.CanManageSprints, &p.CanViewReports, &p.CanExport, &p.CustomPermissions,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *pgInvitationRepository) CreatePermissions(ctx context.Context, perms *InvitationPermissions) error {
	if perms.ID == "" {
		perms.ID = uuid.New().String()
	}
	query := `
		INSERT INTO invitation_permissions (
			id, invitation_id, can_edit_tasks, can_create_tasks, can_delete_tasks,
			can_comment, can_create_subtasks, can_assign_tasks, can_see_time_spent,
			can_track_time, can_add_tags, can_create_views, can_invite_others,
			can_manage_sprints, can_view_reports, can_export, custom_permissions
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := r.pool.Exec(ctx, query,
		perms.ID, perms.InvitationID, perms.CanEditTasks, perms.CanCreateTasks, perms.CanDeleteTasks,
		perms.CanComment, perms.CanCreateSubtasks, perms.CanAssignTasks, perms.CanSeeTimeSpent,
		perms.CanTrackTime, perms.CanAddTags, perms.CanCreateViews, perms.CanInviteOthers,
		perms.CanManageSprints, perms.CanViewReports, perms.CanExport, perms.CustomPermissions,
	)
	return err
}

func (r *pgInvitationRepository) UpdatePermissions(ctx context.Context, perms *InvitationPermissions) error {
	query := `
		UPDATE invitation_permissions SET
			can_edit_tasks = $2, can_create_tasks = $3, can_delete_tasks = $4,
			can_comment = $5, can_create_subtasks = $6, can_assign_tasks = $7,
			can_see_time_spent = $8, can_track_time = $9, can_add_tags = $10,
			can_create_views = $11, can_invite_others = $12, can_manage_sprints = $13,
			can_view_reports = $14, can_export = $15, custom_permissions = $16
		WHERE invitation_id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		perms.InvitationID, perms.CanEditTasks, perms.CanCreateTasks, perms.CanDeleteTasks,
		perms.CanComment, perms.CanCreateSubtasks, perms.CanAssignTasks, perms.CanSeeTimeSpent,
		perms.CanTrackTime, perms.CanAddTags, perms.CanCreateViews, perms.CanInviteOthers,
		perms.CanManageSprints, perms.CanViewReports, perms.CanExport, perms.CustomPermissions,
	)
	return err
}

func (r *pgInvitationRepository) DeletePermissions(ctx context.Context, invitationID string) error {
	query := `DELETE FROM invitation_permissions WHERE invitation_id = $1`
	_, err := r.pool.Exec(ctx, query, invitationID)
	return err
}

func (r *pgInvitationRepository) CreateLinkSettings(ctx context.Context, settings *InvitationLinkSettings) error {
	if settings.ID == "" {
		settings.ID = uuid.New().String()
	}
	if settings.LinkToken == "" {
		settings.LinkToken = uuid.New().String()
	}
	query := `
		INSERT INTO invitation_link_settings (
			id, workspace_id, link_token, type, target_id, default_role, default_permission,
			is_active, requires_approval, allowed_domains, blocked_domains, max_uses,
			use_count, expires_at, created_by_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		settings.ID, settings.WorkspaceID, settings.LinkToken, settings.Type, settings.TargetID,
		settings.DefaultRole, settings.DefaultPermission, settings.IsActive, settings.RequiresApproval,
		settings.AllowedDomains, settings.BlockedDomains, settings.MaxUses, settings.UseCount,
		settings.ExpiresAt, settings.CreatedByID,
	).Scan(&settings.CreatedAt, &settings.UpdatedAt)
}

func (r *pgInvitationRepository) GetLinkSettings(ctx context.Context, id string) (*InvitationLinkSettings, error) {
	query := `
		SELECT id, workspace_id, link_token, type, target_id, default_role, default_permission,
			   is_active, requires_approval, allowed_domains, blocked_domains, max_uses,
			   use_count, expires_at, created_by_id, created_at, updated_at
		FROM invitation_link_settings WHERE id = $1
	`
	s := &InvitationLinkSettings{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.WorkspaceID, &s.LinkToken, &s.Type, &s.TargetID, &s.DefaultRole, &s.DefaultPermission,
		&s.IsActive, &s.RequiresApproval, &s.AllowedDomains, &s.BlockedDomains, &s.MaxUses,
		&s.UseCount, &s.ExpiresAt, &s.CreatedByID, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *pgInvitationRepository) GetLinkSettingsByToken(ctx context.Context, linkToken string) (*InvitationLinkSettings, error) {
	query := `
		SELECT id, workspace_id, link_token, type, target_id, default_role, default_permission,
			   is_active, requires_approval, allowed_domains, blocked_domains, max_uses,
			   use_count, expires_at, created_by_id, created_at, updated_at
		FROM invitation_link_settings WHERE link_token = $1
	`
	s := &InvitationLinkSettings{}
	err := r.pool.QueryRow(ctx, query, linkToken).Scan(
		&s.ID, &s.WorkspaceID, &s.LinkToken, &s.Type, &s.TargetID, &s.DefaultRole, &s.DefaultPermission,
		&s.IsActive, &s.RequiresApproval, &s.AllowedDomains, &s.BlockedDomains, &s.MaxUses,
		&s.UseCount, &s.ExpiresAt, &s.CreatedByID, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *pgInvitationRepository) GetLinkSettingsByTarget(ctx context.Context, targetType InvitationType, targetID string) ([]*InvitationLinkSettings, error) {
	query := `
		SELECT id, workspace_id, link_token, type, target_id, default_role, default_permission,
			   is_active, requires_approval, allowed_domains, blocked_domains, max_uses,
			   use_count, expires_at, created_by_id, created_at, updated_at
		FROM invitation_link_settings WHERE type = $1 AND target_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, targetType, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*InvitationLinkSettings
	for rows.Next() {
		s := &InvitationLinkSettings{}
		if err := rows.Scan(
			&s.ID, &s.WorkspaceID, &s.LinkToken, &s.Type, &s.TargetID, &s.DefaultRole, &s.DefaultPermission,
			&s.IsActive, &s.RequiresApproval, &s.AllowedDomains, &s.BlockedDomains, &s.MaxUses,
			&s.UseCount, &s.ExpiresAt, &s.CreatedByID, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (r *pgInvitationRepository) UpdateLinkSettings(ctx context.Context, settings *InvitationLinkSettings) error {
	query := `
		UPDATE invitation_link_settings SET
			default_role = $2, default_permission = $3, is_active = $4, requires_approval = $5,
			allowed_domains = $6, blocked_domains = $7, max_uses = $8, expires_at = $9, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	return r.pool.QueryRow(ctx, query,
		settings.ID, settings.DefaultRole, settings.DefaultPermission, settings.IsActive,
		settings.RequiresApproval, settings.AllowedDomains, settings.BlockedDomains,
		settings.MaxUses, settings.ExpiresAt,
	).Scan(&settings.UpdatedAt)
}

func (r *pgInvitationRepository) DeactivateLinkSettings(ctx context.Context, id string) error {
	query := `UPDATE invitation_link_settings SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) IncrementLinkSettingsUseCount(ctx context.Context, id string) error {
	query := `UPDATE invitation_link_settings SET use_count = use_count + 1, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) DeleteLinkSettings(ctx context.Context, id string) error {
	query := `DELETE FROM invitation_link_settings WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) CreateAccessRequest(ctx context.Context, req *AccessRequest) error {
	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	if req.Status == "" {
		req.Status = "pending"
	}
	query := `
		INSERT INTO access_requests (
			id, workspace_id, requester_id, email, type, target_id, message, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		req.ID, req.WorkspaceID, req.RequesterID, req.Email, req.Type,
		req.TargetID, req.Message, req.Status,
	).Scan(&req.CreatedAt, &req.UpdatedAt)
}

func (r *pgInvitationRepository) GetAccessRequest(ctx context.Context, id string) (*AccessRequest, error) {
	query := `
		SELECT id, workspace_id, requester_id, email, type, target_id, message, status,
			   processed_by, processed_at, denial_reason, created_at, updated_at
		FROM access_requests WHERE id = $1
	`
	req := &AccessRequest{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&req.ID, &req.WorkspaceID, &req.RequesterID, &req.Email, &req.Type, &req.TargetID,
		&req.Message, &req.Status, &req.ProcessedBy, &req.ProcessedAt, &req.DenialReason,
		&req.CreatedAt, &req.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return req, err
}

func (r *pgInvitationRepository) GetAccessRequestsByTarget(ctx context.Context, targetType InvitationType, targetID string, status string) ([]*AccessRequest, error) {
	query := `
		SELECT id, workspace_id, requester_id, email, type, target_id, message, status,
			   processed_by, processed_at, denial_reason, created_at, updated_at
		FROM access_requests WHERE type = $1 AND target_id = $2
	`
	args := []interface{}{targetType, targetID}
	if status != "" {
		query += " AND status = $3"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*AccessRequest
	for rows.Next() {
		req := &AccessRequest{}
		if err := rows.Scan(
			&req.ID, &req.WorkspaceID, &req.RequesterID, &req.Email, &req.Type, &req.TargetID,
			&req.Message, &req.Status, &req.ProcessedBy, &req.ProcessedAt, &req.DenialReason,
			&req.CreatedAt, &req.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *pgInvitationRepository) GetAccessRequestsByRequester(ctx context.Context, requesterID string) ([]*AccessRequest, error) {
	query := `
		SELECT id, workspace_id, requester_id, email, type, target_id, message, status,
			   processed_by, processed_at, denial_reason, created_at, updated_at
		FROM access_requests WHERE requester_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, requesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*AccessRequest
	for rows.Next() {
		req := &AccessRequest{}
		if err := rows.Scan(
			&req.ID, &req.WorkspaceID, &req.RequesterID, &req.Email, &req.Type, &req.TargetID,
			&req.Message, &req.Status, &req.ProcessedBy, &req.ProcessedAt, &req.DenialReason,
			&req.CreatedAt, &req.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *pgInvitationRepository) UpdateAccessRequestStatus(ctx context.Context, id, status string, processedBy *string, denialReason *string) error {
	query := `
		UPDATE access_requests SET
			status = $2, processed_by = $3, processed_at = NOW(), denial_reason = $4, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, status, processedBy, denialReason)
	return err
}

func (r *pgInvitationRepository) DeleteAccessRequest(ctx context.Context, id string) error {
	query := `DELETE FROM access_requests WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) CreateBulkResult(ctx context.Context, result *BulkInvitationResult) error {
	if result.ID == "" {
		result.ID = uuid.New().String()
	}
	query := `
		INSERT INTO bulk_invitation_results (
			id, workspace_id, invited_by_id, type, target_id, role, total_count,
			success_count, failed_count, skipped_count, status, failed_emails, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		RETURNING created_at
	`
	return r.pool.QueryRow(ctx, query,
		result.ID, result.WorkspaceID, result.InvitedByID, result.Type, result.TargetID,
		result.Role, result.TotalCount, result.SuccessCount, result.FailedCount,
		result.SkippedCount, result.Status, result.FailedEmails,
	).Scan(&result.CreatedAt)
}

func (r *pgInvitationRepository) GetBulkResult(ctx context.Context, id string) (*BulkInvitationResult, error) {
	query := `
		SELECT id, workspace_id, invited_by_id, type, target_id, role, total_count,
			   success_count, failed_count, skipped_count, status, failed_emails, created_at, completed_at
		FROM bulk_invitation_results WHERE id = $1
	`
	result := &BulkInvitationResult{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.WorkspaceID, &result.InvitedByID, &result.Type, &result.TargetID,
		&result.Role, &result.TotalCount, &result.SuccessCount, &result.FailedCount,
		&result.SkippedCount, &result.Status, &result.FailedEmails, &result.CreatedAt, &result.CompletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return result, err
}

func (r *pgInvitationRepository) UpdateBulkResult(ctx context.Context, result *BulkInvitationResult) error {
	query := `
		UPDATE bulk_invitation_results SET
			success_count = $2, failed_count = $3, skipped_count = $4, status = $5,
			failed_emails = $6, completed_at = $7
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		result.ID, result.SuccessCount, result.FailedCount, result.SkippedCount,
		result.Status, result.FailedEmails, result.CompletedAt,
	)
	return err
}

func (r *pgInvitationRepository) scanOne(ctx context.Context, query string, args ...interface{}) (*Invitation, error) {
	inv := &Invitation{}
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&inv.ID, &inv.WorkspaceID, &inv.Email, &inv.Token, &inv.LinkToken,
		&inv.Type, &inv.TargetID, &inv.TargetName, &inv.Role, &inv.Permission,
		&inv.InvitedByID, &inv.InvitedByName, &inv.InviteeUserID, &inv.Status,
		&inv.Method, &inv.Message, &inv.ExpiresAt, &inv.LinkExpiresAt,
		&inv.AcceptedAt, &inv.DeclinedAt, &inv.ReminderSentAt, &inv.ReminderCount,
		&inv.MaxUses, &inv.UseCount, &inv.Metadata, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *pgInvitationRepository) scanMany(ctx context.Context, query string, args ...interface{}) ([]*Invitation, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*Invitation
	for rows.Next() {
		inv := &Invitation{}
		if err := rows.Scan(
			&inv.ID, &inv.WorkspaceID, &inv.Email, &inv.Token, &inv.LinkToken,
			&inv.Type, &inv.TargetID, &inv.TargetName, &inv.Role, &inv.Permission,
			&inv.InvitedByID, &inv.InvitedByName, &inv.InviteeUserID, &inv.Status,
			&inv.Method, &inv.Message, &inv.ExpiresAt, &inv.LinkExpiresAt,
			&inv.AcceptedAt, &inv.DeclinedAt, &inv.ReminderSentAt, &inv.ReminderCount,
			&inv.MaxUses, &inv.UseCount, &inv.Metadata, &inv.CreatedAt, &inv.UpdatedAt,
		); err != nil {
			return nil, err
		}
		invitations = append(invitations, inv)
	}
	return invitations, nil
}