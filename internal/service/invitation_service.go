package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/email"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

type InvitationService interface {
	// Core invitation operations
	CreateInvitation(ctx context.Context, inv *repository.Invitation) error
	CreateWithPermissions(ctx context.Context, inv *repository.Invitation, perms *repository.InvitationPermissions) error
	CreateBatch(ctx context.Context, invitations []*repository.Invitation) ([]string, []error)

	// Retrieval
	GetByID(ctx context.Context, id string) (*repository.Invitation, error)
	GetByToken(ctx context.Context, token string) (*repository.Invitation, error)
	GetByLinkToken(ctx context.Context, linkToken string) (*repository.Invitation, error)
	GetMyInvitations(ctx context.Context, email string) ([]*repository.Invitation, error)

	// Acceptance and lifecycle
	AcceptByID(ctx context.Context, id string, userID string) error
	AcceptByToken(ctx context.Context, token string, userID string) error
	DeclineByID(ctx context.Context, id string) error
	CancelInvitation(ctx context.Context, id string, actorID string) error
	ResendInvitation(ctx context.Context, id string, actorID *string) (*repository.Invitation, error)

	// Workspace and Project specific invitations
	CreateWorkspaceInvitation(ctx context.Context, workspaceID, email, role, inviterID string) (*repository.Invitation, error)
	CreateProjectInvitation(ctx context.Context, workspaceID, projectID, email, role, inviterID string) (*repository.Invitation, error)

	// List operations
	ListByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*repository.Invitation, int, error)
	ListByProject(ctx context.Context, projectID string, limit, offset int) ([]*repository.Invitation, int, error)

	// Link invitations
	CreateLinkSettings(ctx context.Context, settings *repository.InvitationLinkSettings) error
	UseLink(ctx context.Context, linkToken string, emailAddr string) (*repository.Invitation, *repository.InvitationLinkSettings, error)
	GetLinkSettingsByToken(ctx context.Context, token string) (*repository.InvitationLinkSettings, error)

	// Stats and analytics
	GetStatsByWorkspace(ctx context.Context, workspaceID string) (*repository.InvitationStats, error)

	// Token management
	RegenerateToken(ctx context.Context, id string) (string, error)

	// Activity and permissions
	LogActivity(ctx context.Context, a *repository.InvitationActivity) error
	GetActivity(ctx context.Context, invitationID string) ([]*repository.InvitationActivity, error)
	GetPermissions(ctx context.Context, invitationID string) (*repository.InvitationPermissions, error)
	CreatePermissions(ctx context.Context, perms *repository.InvitationPermissions) error
	UpdatePermissions(ctx context.Context, perms *repository.InvitationPermissions) error
	DeletePermissions(ctx context.Context, invitationID string) error

	// Access requests
	CreateAccessRequest(ctx context.Context, req *repository.AccessRequest) error
}

type invitationService struct {
	invRepo      repository.InvitationRepository
	workspaceRepo repository.WorkspaceRepository
	teamRepo     repository.TeamRepository
	projectRepo  repository.ProjectRepository
	userRepo     repository.UserRepository
	spaceRepo    repository.SpaceRepository
	emailSvc     *email.Service
	defaultTTL   time.Duration
}

func NewInvitationService(
	invRepo repository.InvitationRepository,
	workspaceRepo repository.WorkspaceRepository,
	teamRepo repository.TeamRepository,
	projectRepo repository.ProjectRepository,
	userRepo repository.UserRepository,
	spaceRepo repository.SpaceRepository,
	emailSvc *email.Service,
) InvitationService {
	return &invitationService{
		invRepo:      invRepo,
		workspaceRepo: workspaceRepo,
		teamRepo:     teamRepo,
		projectRepo:  projectRepo,
		userRepo:     userRepo,
		spaceRepo:    spaceRepo,
		emailSvc:     emailSvc,
		defaultTTL:   30 * 24 * time.Hour,
	}
}

func normalizeEmail(e string) string {
	return strings.ToLower(strings.TrimSpace(e))
}

func allowedRoleForType(t repository.InvitationType, r repository.WorkspaceRole) bool {
	for _, v := range repository.ValidRolesForType(t) {
		if v == r {
			return true
		}
	}
	return false
}

func strPtr(s string) *string { return &s }

func (s *invitationService) CreateInvitation(ctx context.Context, inv *repository.Invitation) error {
	if inv == nil {
		return errors.New("invitation is nil")
	}
	if strings.TrimSpace(inv.WorkspaceID) == "" {
		return errors.New("workspace_id required")
	}
	if strings.TrimSpace(inv.Email) == "" {
		return errors.New("email required")
	}
	inv.Email = normalizeEmail(inv.Email)

	if inv.Type == "" {
		inv.Type = repository.InvitationTypeWorkspace
	}
	if inv.Role == "" {
		inv.Role = repository.WorkspaceRoleMember
	}
	if !allowedRoleForType(inv.Type, inv.Role) {
		return errors.New("invalid role for invitation type")
	}
	if inv.Permission == "" {
		inv.Permission = repository.DefaultPermissionForRole(inv.Role)
	}
	if inv.Status == "" {
		inv.Status = repository.InvitationStatusPending
	}
	if inv.Method == "" {
		inv.Method = repository.InvitationMethodEmail
	}
	if inv.ExpiresAt == nil {
		t := time.Now().Add(s.defaultTTL)
		inv.ExpiresAt = &t
	}

	if inv.TargetID != "" && inv.Email != "" {
		exists, err := s.invRepo.ExistsPendingForEmail(ctx, inv.Email, inv.Type, inv.TargetID)
		if err == nil && exists {
			return errors.New("pending invitation already exists for this email and target")
		}
	}

	if err := s.invRepo.Create(ctx, inv); err != nil {
		return err
	}

	_ = s.invRepo.LogActivity(ctx, &repository.InvitationActivity{
		InvitationID: inv.ID,
		Action:       "created",
		ActorID:      &inv.InvitedByID,
		ActorType:    "user",
	})

	if s.emailSvc != nil && inv.Method == repository.InvitationMethodEmail {
		go func(inv *repository.Invitation) {
			workspaceName := inv.WorkspaceID
			if ws, err := s.workspaceRepo.FindByID(context.Background(), inv.WorkspaceID); err == nil && ws != nil {
				workspaceName = ws.Name
			}
			_ = s.emailSvc.SendInvitation(workspaceName, inv.Email, inv.InvitedByName, inv.Token)
		}(inv)
	}

	return nil
}

func (s *invitationService) CreateWithPermissions(ctx context.Context, inv *repository.Invitation, perms *repository.InvitationPermissions) error {
	if inv == nil {
		return errors.New("invitation is nil")
	}
	if inv.Email != "" {
		inv.Email = normalizeEmail(inv.Email)
	}
	if inv.Permission == "" {
		inv.Permission = repository.DefaultPermissionForRole(inv.Role)
	}
	return s.invRepo.CreateWithPermissions(ctx, inv, perms)
}

func (s *invitationService) CreateBatch(ctx context.Context, invitations []*repository.Invitation) ([]string, []error) {
	for _, inv := range invitations {
		if inv != nil && inv.Email != "" {
			inv.Email = normalizeEmail(inv.Email)
		}
	}
	return s.invRepo.CreateBatch(ctx, invitations)
}

func (s *invitationService) GetByID(ctx context.Context, id string) (*repository.Invitation, error) {
	if id == "" {
		return nil, errors.New("id required")
	}
	return s.invRepo.FindByID(ctx, id)
}

func (s *invitationService) GetByToken(ctx context.Context, token string) (*repository.Invitation, error) {
	if token == "" {
		return nil, errors.New("token required")
	}
	return s.invRepo.FindByToken(ctx, token)
}

func (s *invitationService) GetByLinkToken(ctx context.Context, linkToken string) (*repository.Invitation, error) {
	if linkToken == "" {
		return nil, errors.New("link token required")
	}
	return s.invRepo.FindByLinkToken(ctx, linkToken)
}

func (s *invitationService) GetMyInvitations(ctx context.Context, email string) ([]*repository.Invitation, error) {
	if email == "" {
		return nil, errors.New("email required")
	}
	return s.invRepo.FindPendingByEmail(ctx, normalizeEmail(email))
}

func (s *invitationService) AcceptByID(ctx context.Context, id string, userID string) error {
	if id == "" {
		return errors.New("id required")
	}
	if userID == "" {
		return errors.New("user_id required")
	}

	inv, err := s.invRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.New("invitation not found")
	}
	if !inv.CanAccept() {
		return errors.New("invitation cannot be accepted")
	}
	if inv.MaxUses != nil && inv.UseCount >= *inv.MaxUses {
		return errors.New("invitation max uses reached")
	}

	if err := s.invRepo.MarkAccepted(ctx, id, userID); err != nil {
		return err
	}

	_ = s.invRepo.IncrementLinkUseCount(ctx, id)
	_ = s.invRepo.LogActivity(ctx, &repository.InvitationActivity{
		InvitationID: id,
		Action:       "accepted",
		ActorID:      &userID,
		ActorType:    "user",
	})

	return s.addUserToTarget(ctx, inv, userID)
}

func (s *invitationService) AcceptByToken(ctx context.Context, token string, userID string) error {
	if token == "" {
		return errors.New("token required")
	}
	if userID == "" {
		return errors.New("user_id required")
	}

	inv, err := s.invRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.New("invitation not found")
	}
	if !inv.CanAccept() {
		return errors.New("invitation cannot be accepted")
	}
	if inv.MaxUses != nil && inv.UseCount >= *inv.MaxUses {
		return errors.New("invitation max uses reached")
	}

	if err := s.invRepo.MarkAccepted(ctx, inv.ID, userID); err != nil {
		return err
	}

	_ = s.invRepo.IncrementLinkUseCount(ctx, inv.ID)
	_ = s.invRepo.LogActivity(ctx, &repository.InvitationActivity{
		InvitationID: inv.ID,
		Action:       "accepted",
		ActorID:      &userID,
		ActorType:    "user",
	})

	return s.addUserToTarget(ctx, inv, userID)
}

func (s *invitationService) addUserToTarget(ctx context.Context, inv *repository.Invitation, userID string) error {
	switch inv.Type {
	case repository.InvitationTypeWorkspace:
		member := &repository.WorkspaceMember{
			WorkspaceID: inv.TargetID,
			UserID:      userID,
			Role:        string(inv.Role),
		}
		return s.workspaceRepo.AddMember(ctx, member)

	case repository.InvitationTypeProject:
		member := &repository.ProjectMember{
			ProjectID: inv.TargetID,
			UserID:    userID,
			Role:      string(inv.Role),
		}
		return s.projectRepo.AddMember(ctx, member)

	case repository.InvitationTypeTeam:
		member := &repository.TeamMember{
			TeamID: inv.TargetID,
			UserID: userID,
			Role:   string(inv.Role),
		}
		return s.teamRepo.AddMember(ctx, member)

	default:
		return errors.New("unsupported invitation type")
	}
}

func (s *invitationService) DeclineByID(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id required")
	}
	inv, err := s.invRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.New("invitation not found")
	}
	if err := s.invRepo.MarkDeclined(ctx, id); err != nil {
		return err
	}
	_ = s.invRepo.LogActivity(ctx, &repository.InvitationActivity{
		InvitationID: id,
		Action:       "declined",
		ActorType:    "user",
	})
	return nil
}

func (s *invitationService) CancelInvitation(ctx context.Context, id string, actorID string) error {
	if id == "" {
		return errors.New("id required")
	}
	inv, err := s.invRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.New("invitation not found")
	}
	if !inv.CanCancel() {
		return errors.New("invitation cannot be cancelled")
	}
	if err := s.invRepo.MarkCancelled(ctx, id); err != nil {
		return err
	}
	_ = s.invRepo.LogActivity(ctx, &repository.InvitationActivity{
		InvitationID: id,
		Action:       "cancelled",
		ActorID:      &actorID,
		ActorType:    "user",
	})
	return nil
}

func (s *invitationService) ResendInvitation(ctx context.Context, id string, actorID *string) (*repository.Invitation, error) {
	if id == "" {
		return nil, errors.New("id required")
	}
	inv, err := s.invRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, errors.New("invitation not found")
	}
	if !inv.CanResend() {
		return nil, errors.New("invitation cannot be resent")
	}
	if err := s.invRepo.UpdateReminderSent(ctx, id); err != nil {
		return nil, err
	}
	_ = s.invRepo.LogActivity(ctx, &repository.InvitationActivity{
		InvitationID: id,
		Action:       "resent",
		ActorID:      actorID,
		ActorType:    "user",
	})

	if s.emailSvc != nil && inv.Method == repository.InvitationMethodEmail {
		go func(inv *repository.Invitation) {
			workspaceName := inv.WorkspaceID
			if ws, err := s.workspaceRepo.FindByID(context.Background(), inv.WorkspaceID); err == nil && ws != nil {
				workspaceName = ws.Name
			}
			_ = s.emailSvc.SendInvitation(workspaceName, inv.Email, inv.InvitedByName, inv.Token)
		}(inv)
	}

	return s.invRepo.FindByID(ctx, id)
}

func (s *invitationService) CreateWorkspaceInvitation(ctx context.Context, workspaceID, email, role, inviterID string) (*repository.Invitation, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return nil, errors.New("workspace not found")
	}

	inviter, _ := s.userRepo.FindByID(ctx, inviterID)
	inviterName := "Someone"
	if inviter != nil {
		inviterName = inviter.Name
	}

	inv := &repository.Invitation{
		WorkspaceID:   workspaceID,
		Email:         normalizeEmail(email),
		Type:          repository.InvitationTypeWorkspace,
		TargetID:      workspaceID,
		TargetName:    workspace.Name,
		Role:          repository.WorkspaceRole(role),
		InvitedByID:   inviterID,
		InvitedByName: inviterName,
	}

	if err := s.CreateInvitation(ctx, inv); err != nil {
		return nil, err
	}

	return inv, nil
}

func (s *invitationService) CreateProjectInvitation(ctx context.Context, workspaceID, projectID, email, role, inviterID string) (*repository.Invitation, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil || project == nil {
		return nil, errors.New("project not found")
	}

	inviter, _ := s.userRepo.FindByID(ctx, inviterID)
	inviterName := "Someone"
	if inviter != nil {
		inviterName = inviter.Name
	}

	inv := &repository.Invitation{
		WorkspaceID:   workspaceID,
		Email:         normalizeEmail(email),
		Type:          repository.InvitationTypeProject,
		TargetID:      projectID,
		TargetName:    project.Name,
		Role:          repository.WorkspaceRole(role),
		InvitedByID:   inviterID,
		InvitedByName: inviterName,
	}

	if err := s.CreateInvitation(ctx, inv); err != nil {
		return nil, err
	}

	return inv, nil
}

func (s *invitationService) ListByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*repository.Invitation, int, error) {
	return s.invRepo.FindByWorkspace(ctx, workspaceID, limit, offset)
}

func (s *invitationService) ListByProject(ctx context.Context, projectID string, limit, offset int) ([]*repository.Invitation, int, error) {
	invs, err := s.invRepo.FindPendingByTarget(ctx, repository.InvitationTypeProject, projectID)
	if err != nil {
		return nil, 0, err
	}
	total := len(invs)
	if offset >= total {
		return []*repository.Invitation{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return invs[offset:end], total, nil
}

func (s *invitationService) CreateLinkSettings(ctx context.Context, settings *repository.InvitationLinkSettings) error {
	if settings == nil {
		return errors.New("settings required")
	}
	if strings.TrimSpace(settings.WorkspaceID) == "" {
		return errors.New("workspace_id required")
	}
	if settings.DefaultRole == "" {
		settings.DefaultRole = repository.WorkspaceRoleMember
	}
	if settings.DefaultPermission == "" {
		settings.DefaultPermission = repository.DefaultPermissionForRole(settings.DefaultRole)
	}
	if settings.AllowedDomains != nil {
		var arr []string
		if err := json.Unmarshal([]byte(*settings.AllowedDomains), &arr); err != nil {
			return errors.New("allowed_domains must be a JSON array of strings")
		}
	}
	if settings.BlockedDomains != nil {
		var arr []string
		if err := json.Unmarshal([]byte(*settings.BlockedDomains), &arr); err != nil {
			return errors.New("blocked_domains must be a JSON array of strings")
		}
	}
	if settings.MaxUses != nil && *settings.MaxUses <= 0 {
		return errors.New("max_uses must be > 0")
	}
	if settings.ExpiresAt != nil && settings.ExpiresAt.Before(time.Now()) {
		return errors.New("expires_at must be in the future")
	}
	return s.invRepo.CreateLinkSettings(ctx, settings)
}

func (s *invitationService) UseLink(ctx context.Context, linkToken string, emailAddr string) (*repository.Invitation, *repository.InvitationLinkSettings, error) {
	if linkToken == "" {
		return nil, nil, errors.New("link token required")
	}
	if emailAddr == "" {
		return nil, nil, errors.New("email required")
	}
	emailAddr = normalizeEmail(emailAddr)

	ls, err := s.invRepo.GetLinkSettingsByToken(ctx, linkToken)
	if err != nil {
		return nil, nil, err
	}
	if ls == nil {
		return nil, nil, errors.New("link not found")
	}
	if !ls.IsValid() {
		return nil, nil, errors.New("link is not valid (inactive/expired/max-uses)")
	}
	if !ls.CheckDomain(emailAddr) {
		return nil, nil, errors.New("email domain not allowed by link settings")
	}

	if err := s.invRepo.IncrementLinkSettingsUseCount(ctx, ls.ID); err != nil {
		return nil, nil, err
	}

	inv := &repository.Invitation{
		WorkspaceID:   ls.WorkspaceID,
		Email:         emailAddr,
		Type:          ls.Type,
		TargetID:      ls.TargetID,
		Role:          ls.DefaultRole,
		Permission:    ls.DefaultPermission,
		Method:        repository.InvitationMethodLink,
		LinkToken:     &ls.LinkToken,
		Status:        repository.InvitationStatusPending,
		ExpiresAt:     ls.ExpiresAt,
		MaxUses:       ls.MaxUses,
		InvitedByID:   ls.CreatedByID,
		InvitedByName: "",
	}
	if err := s.invRepo.Create(ctx, inv); err != nil {
		return nil, nil, err
	}

	_ = s.invRepo.LogActivity(ctx, &repository.InvitationActivity{
		InvitationID: inv.ID,
		Action:       "created_from_link",
		ActorID:      &ls.CreatedByID,
		ActorType:    "system",
		Details:      strPtr("created via invitation link"),
	})

	if s.emailSvc != nil {
		go func(inv *repository.Invitation) {
			workspaceName := inv.WorkspaceID
			if ws, err := s.workspaceRepo.FindByID(context.Background(), inv.WorkspaceID); err == nil && ws != nil {
				workspaceName = ws.Name
			}
			_ = s.emailSvc.SendInvitation(workspaceName, inv.Email, inv.InvitedByName, inv.Token)
		}(inv)
	}

	return inv, ls, nil
}

func (s *invitationService) GetLinkSettingsByToken(ctx context.Context, token string) (*repository.InvitationLinkSettings, error) {
	if token == "" {
		return nil, errors.New("token required")
	}
	return s.invRepo.GetLinkSettingsByToken(ctx, token)
}

func (s *invitationService) GetStatsByWorkspace(ctx context.Context, workspaceID string) (*repository.InvitationStats, error) {
	if workspaceID == "" {
		return nil, errors.New("workspace_id required")
	}
	return s.invRepo.GetStatsByWorkspace(ctx, workspaceID)
}

func (s *invitationService) RegenerateToken(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", errors.New("id required")
	}
	return s.invRepo.RegenerateToken(ctx, id)
}

func (s *invitationService) LogActivity(ctx context.Context, a *repository.InvitationActivity) error {
	return s.invRepo.LogActivity(ctx, a)
}

func (s *invitationService) GetActivity(ctx context.Context, invitationID string) ([]*repository.InvitationActivity, error) {
	return s.invRepo.GetActivityByInvitation(ctx, invitationID)
}

func (s *invitationService) GetPermissions(ctx context.Context, invitationID string) (*repository.InvitationPermissions, error) {
	return s.invRepo.GetPermissions(ctx, invitationID)
}

func (s *invitationService) CreatePermissions(ctx context.Context, perms *repository.InvitationPermissions) error {
	return s.invRepo.CreatePermissions(ctx, perms)
}

func (s *invitationService) UpdatePermissions(ctx context.Context, perms *repository.InvitationPermissions) error {
	return s.invRepo.UpdatePermissions(ctx, perms)
}

func (s *invitationService) DeletePermissions(ctx context.Context, invitationID string) error {
	return s.invRepo.DeletePermissions(ctx, invitationID)
}

func (s *invitationService) CreateAccessRequest(ctx context.Context, req *repository.AccessRequest) error {
	return s.invRepo.CreateAccessRequest(ctx, req)
}