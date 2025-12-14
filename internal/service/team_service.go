package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/email"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
)

// ============================================
// Team Service
// ============================================

// TeamService defines team operations
type TeamService interface {
	Create(ctx context.Context, workspaceID, creatorID, name string, description, avatar, color *string) (*repository.Team, error)
	GetByID(ctx context.Context, id string) (*repository.Team, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*repository.Team, error)
	ListByUser(ctx context.Context, userID string) ([]*repository.Team, error)
	Update(ctx context.Context, id, userID string, name, description, avatar, color *string) (*repository.Team, error)
	Delete(ctx context.Context, id, userID string) error

	// Member operations
	AddMember(ctx context.Context, teamID, userID, role, addedByID string) error
	AddMemberByEmail(ctx context.Context, teamID, email, role, addedByID string) error
	ListMembers(ctx context.Context, teamID string) ([]*repository.TeamMember, error)
	UpdateMemberRole(ctx context.Context, teamID, userID, role string) error
	RemoveMember(ctx context.Context, teamID, userID string) error
	IsMember(ctx context.Context, teamID, userID string) (bool, error)
}

type teamService struct {
	teamRepo      repository.TeamRepository
	userRepo      repository.UserRepository
	workspaceRepo repository.WorkspaceRepository
	notifSvc      *notification.Service
	emailSvc      *email.Service
	broadcaster   *socket.Broadcaster
}

// NewTeamService creates a new team service
func NewTeamService(
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	workspaceRepo repository.WorkspaceRepository,
	notifSvc *notification.Service,
	emailSvc *email.Service,
	broadcaster *socket.Broadcaster,
) TeamService {
	return &teamService{
		teamRepo:      teamRepo,
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
		notifSvc:      notifSvc,
		emailSvc:      emailSvc,
		broadcaster:   broadcaster,
	}
}

func (s *teamService) Create(ctx context.Context, workspaceID, creatorID, name string, description, avatar, color *string) (*repository.Team, error) {
	// Verify workspace membership
	member, _ := s.workspaceRepo.FindMember(ctx, workspaceID, creatorID)
	if member == nil {
		return nil, ErrForbidden
	}

	team := &repository.Team{
		Name:        name,
		Description: description,
		Avatar:      avatar,
		Color:       color,
		WorkspaceID: workspaceID,
		CreatedBy:   creatorID,
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	// Auto-add creator as owner
	teamMember := &repository.TeamMember{
		TeamID: team.ID,
		UserID: creatorID,
		Role:   "owner",
	}
	s.teamRepo.AddMember(ctx, teamMember)

	// Broadcast team creation
	if s.broadcaster != nil {
		s.broadcaster.BroadcastTeamCreated(workspaceID, map[string]interface{}{
			"id":          team.ID,
			"name":        team.Name,
			"workspaceId": team.WorkspaceID,
			"createdBy":   team.CreatedBy,
		})
	}

	return team, nil
}

func (s *teamService) GetByID(ctx context.Context, id string) (*repository.Team, error) {
	team, err := s.teamRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, ErrNotFound
	}

	// Load members
	members, _ := s.teamRepo.FindMembers(ctx, id)
	team.Members = members

	return team, nil
}

func (s *teamService) ListByWorkspace(ctx context.Context, workspaceID string) ([]*repository.Team, error) {
	teams, err := s.teamRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Load member counts
	for _, team := range teams {
		members, _ := s.teamRepo.FindMembers(ctx, team.ID)
		team.Members = members
	}

	return teams, nil
}

func (s *teamService) ListByUser(ctx context.Context, userID string) ([]*repository.Team, error) {
	return s.teamRepo.FindByUserID(ctx, userID)
}

func (s *teamService) Update(ctx context.Context, id, userID string, name, description, avatar, color *string) (*repository.Team, error) {
	team, err := s.teamRepo.FindByID(ctx, id)
	if err != nil || team == nil {
		return nil, ErrNotFound
	}

	// Check permission (owner or admin)
	member, _ := s.teamRepo.FindMember(ctx, id, userID)
	if member == nil || (member.Role != "owner" && member.Role != "admin") {
		return nil, ErrForbidden
	}

	if name != nil {
		team.Name = *name
	}
	if description != nil {
		team.Description = description
	}
	if avatar != nil {
		team.Avatar = avatar
	}
	if color != nil {
		team.Color = color
	}

	if err := s.teamRepo.Update(ctx, team); err != nil {
		return nil, err
	}

	// Broadcast update
	if s.broadcaster != nil {
		s.broadcaster.BroadcastTeamUpdated(team.WorkspaceID, map[string]interface{}{
			"id":          team.ID,
			"name":        team.Name,
			"description": team.Description,
		}, userID)
	}

	return team, nil
}

func (s *teamService) Delete(ctx context.Context, id, userID string) error {
	team, err := s.teamRepo.FindByID(ctx, id)
	if err != nil || team == nil {
		return ErrNotFound
	}

	// Check permission (only owner)
	member, _ := s.teamRepo.FindMember(ctx, id, userID)
	if member == nil || member.Role != "owner" {
		return ErrForbidden
	}

	workspaceID := team.WorkspaceID

	if err := s.teamRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Broadcast deletion
	if s.broadcaster != nil {
		s.broadcaster.BroadcastTeamDeleted(workspaceID, id)
	}

	return nil
}

func (s *teamService) AddMember(ctx context.Context, teamID, userID, role, addedByID string) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil || team == nil {
		return ErrNotFound
	}

	// Check if already a member
	existing, _ := s.teamRepo.FindMember(ctx, teamID, userID)
	if existing != nil {
		return ErrConflict
	}

	member := &repository.TeamMember{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}

	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		return err
	}

	// Get user info for notifications
	user, _ := s.userRepo.FindByID(ctx, userID)
	addedBy, _ := s.userRepo.FindByID(ctx, addedByID)
	workspace, _ := s.workspaceRepo.FindByID(ctx, team.WorkspaceID)

	// Send notification
	if s.notifSvc != nil && user != nil {
		s.notifSvc.SendBatchNotifications(ctx, []string{userID}, "",
			"TEAM_ADDED", "Added to Team",
			fmt.Sprintf("You've been added to team: %s", team.Name),
			map[string]interface{}{
				"teamId":      teamID,
				"teamName":    team.Name,
				"workspaceId": team.WorkspaceID,
				"action":      "view_team",
			})
	}

	// Send email
	if s.emailSvc != nil && user != nil && addedBy != nil && workspace != nil {
		s.emailSvc.SendTeamInvitation(user.Email, email.TeamInvitationData{
			UserName:      user.Name,
			TeamName:      team.Name,
			WorkspaceName: workspace.Name,
			AddedBy:       addedBy.Name,
			TeamURL:       fmt.Sprintf("/workspaces/%s/teams/%s", team.WorkspaceID, teamID),
		})
	}

	// Broadcast member added
	if s.broadcaster != nil && user != nil {
		s.broadcaster.BroadcastTeamMemberAdded(team.WorkspaceID, teamID, map[string]interface{}{
			"userId": userID,
			"name":   user.Name,
			"role":   role,
		})
	}

	return nil
}

func (s *teamService) AddMemberByEmail(ctx context.Context, teamID, userEmail, role, addedByID string) error {
	user, err := s.userRepo.FindByEmail(ctx, userEmail)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.AddMember(ctx, teamID, user.ID, role, addedByID)
}

func (s *teamService) ListMembers(ctx context.Context, teamID string) ([]*repository.TeamMember, error) {
	return s.teamRepo.FindMembers(ctx, teamID)
}

func (s *teamService) UpdateMemberRole(ctx context.Context, teamID, userID, role string) error {
	return s.teamRepo.UpdateMemberRole(ctx, teamID, userID, role)
}

func (s *teamService) RemoveMember(ctx context.Context, teamID, userID string) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil || team == nil {
		return ErrNotFound
	}

	if err := s.teamRepo.RemoveMember(ctx, teamID, userID); err != nil {
		return err
	}

	// Broadcast member removed
	if s.broadcaster != nil {
		s.broadcaster.BroadcastTeamMemberRemoved(team.WorkspaceID, teamID, userID)
	}

	return nil
}

func (s *teamService) IsMember(ctx context.Context, teamID, userID string) (bool, error) {
	return s.teamRepo.IsMember(ctx, teamID, userID)
}

// ============================================
// Invitation Service
// ============================================

// InvitationService defines invitation operations
type InvitationService interface {
	CreateWorkspaceInvitation(ctx context.Context, workspaceID, email, role, invitedByID string) (*repository.Invitation, error)
	CreateTeamInvitation(ctx context.Context, teamID, email, role, invitedByID string) (*repository.Invitation, error)
	CreateProjectInvitation(ctx context.Context, projectID, email, role, invitedByID string) (*repository.Invitation, error)
	AcceptInvitation(ctx context.Context, token, userID string) error
	GetPendingInvitations(ctx context.Context, email string) ([]*repository.Invitation, error)
	CancelInvitation(ctx context.Context, id string) error
}

type invitationService struct {
	invitationRepo repository.InvitationRepository
	workspaceRepo  repository.WorkspaceRepository
	teamRepo       repository.TeamRepository
	projectRepo    repository.ProjectRepository
	userRepo       repository.UserRepository
	emailSvc       *email.Service
}

// NewInvitationService creates a new invitation service
func NewInvitationService(
	invitationRepo repository.InvitationRepository,
	workspaceRepo repository.WorkspaceRepository,
	teamRepo repository.TeamRepository,
	projectRepo repository.ProjectRepository,
	userRepo repository.UserRepository,
	emailSvc *email.Service,
) InvitationService {
	return &invitationService{
		invitationRepo: invitationRepo,
		workspaceRepo:  workspaceRepo,
		teamRepo:       teamRepo,
		projectRepo:    projectRepo,
		userRepo:       userRepo,
		emailSvc:       emailSvc,
	}
}

func (s *invitationService) CreateWorkspaceInvitation(ctx context.Context, workspaceID, inviteEmail, role, invitedByID string) (*repository.Invitation, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return nil, ErrNotFound
	}

	// Check if user already exists and is a member
	existingUser, _ := s.userRepo.FindByEmail(ctx, inviteEmail)
	if existingUser != nil {
		member, _ := s.workspaceRepo.FindMember(ctx, workspaceID, existingUser.ID)
		if member != nil {
			return nil, ErrConflict
		}
	}

	invitation := &repository.Invitation{
		Email:     inviteEmail,
		Type:      "workspace",
		TargetID:  workspaceID,
		Role:      role,
		InvitedBy: invitedByID,
		Status:    "pending",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	// Send invitation email
	if s.emailSvc != nil {
		inviter, _ := s.userRepo.FindByID(ctx, invitedByID)
		inviterName := "Someone"
		if inviter != nil {
			inviterName = inviter.Name
		}

		s.emailSvc.SendWorkspaceInvitation(inviteEmail, email.WorkspaceInvitationData{
			InviterName:   inviterName,
			WorkspaceName: workspace.Name,
			Role:          role,
			InviteURL:     fmt.Sprintf("/invite/%s", invitation.Token),
		})
	}

	return invitation, nil
}

func (s *invitationService) CreateTeamInvitation(ctx context.Context, teamID, inviteEmail, role, invitedByID string) (*repository.Invitation, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil || team == nil {
		return nil, ErrNotFound
	}

	invitation := &repository.Invitation{
		Email:     inviteEmail,
		Type:      "team",
		TargetID:  teamID,
		Role:      role,
		InvitedBy: invitedByID,
		Status:    "pending",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *invitationService) CreateProjectInvitation(ctx context.Context, projectID, inviteEmail, role, invitedByID string) (*repository.Invitation, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil || project == nil {
		return nil, ErrNotFound
	}

	invitation := &repository.Invitation{
		Email:     inviteEmail,
		Type:      "project",
		TargetID:  projectID,
		Role:      role,
		InvitedBy: invitedByID,
		Status:    "pending",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *invitationService) AcceptInvitation(ctx context.Context, token, userID string) error {
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}
	if invitation == nil {
		return ErrNotFound
	}

	if invitation.Status != "pending" {
		return ErrConflict
	}

	if time.Now().After(invitation.ExpiresAt) {
		invitation.Status = "expired"
		s.invitationRepo.Update(ctx, invitation)
		return ErrInvalidToken
	}

	// Add user to the target
	switch invitation.Type {
	case "workspace":
		member := &repository.WorkspaceMember{
			WorkspaceID: invitation.TargetID,
			UserID:      userID,
			Role:        invitation.Role,
		}
		if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
			return err
		}

	case "team":
		member := &repository.TeamMember{
			TeamID: invitation.TargetID,
			UserID: userID,
			Role:   invitation.Role,
		}
		if err := s.teamRepo.AddMember(ctx, member); err != nil {
			return err
		}

	case "project":
		member := &repository.ProjectMember{
			ProjectID: invitation.TargetID,
			UserID:    userID,
			Role:      invitation.Role,
		}
		if err := s.projectRepo.AddMember(ctx, member); err != nil {
			return err
		}
	}

	// Update invitation status
	invitation.Status = "accepted"
	return s.invitationRepo.Update(ctx, invitation)
}

func (s *invitationService) GetPendingInvitations(ctx context.Context, userEmail string) ([]*repository.Invitation, error) {
	return s.invitationRepo.FindByEmail(ctx, userEmail)
}

func (s *invitationService) CancelInvitation(ctx context.Context, id string) error {
	return s.invitationRepo.Delete(ctx, id)
}

// ============================================
// Activity Service
// ============================================

// ActivityService defines activity log operations
type ActivityService interface {
	LogActivity(ctx context.Context, activityType, entityType, entityID, userID string, changes, metadata map[string]interface{}) error
	GetEntityActivities(ctx context.Context, entityType, entityID string, limit int) ([]*repository.Activity, error)
	GetUserActivities(ctx context.Context, userID string, limit int) ([]*repository.Activity, error)
}

type activityService struct {
	activityRepo repository.ActivityRepository
}

// NewActivityService creates a new activity service
func NewActivityService(activityRepo repository.ActivityRepository) ActivityService {
	return &activityService{activityRepo: activityRepo}
}

func (s *activityService) LogActivity(ctx context.Context, activityType, entityType, entityID, userID string, changes, metadata map[string]interface{}) error {
	activity := &repository.Activity{
		Type:       activityType,
		EntityType: entityType,
		EntityID:   entityID,
		UserID:     userID,
		Changes:    changes,
		Metadata:   metadata,
	}
	return s.activityRepo.Create(ctx, activity)
}

func (s *activityService) GetEntityActivities(ctx context.Context, entityType, entityID string, limit int) ([]*repository.Activity, error) {
	return s.activityRepo.FindByEntity(ctx, entityType, entityID, limit)
}

func (s *activityService) GetUserActivities(ctx context.Context, userID string, limit int) ([]*repository.Activity, error) {
	return s.activityRepo.FindByUser(ctx, userID, limit)
}

// ============================================
// Task Watcher Service
// ============================================

// TaskWatcherService defines task watcher operations
type TaskWatcherService interface {
	Watch(ctx context.Context, taskID, userID string) error
	Unwatch(ctx context.Context, taskID, userID string) error
	GetWatchers(ctx context.Context, taskID string) ([]string, error)
	IsWatching(ctx context.Context, taskID, userID string) (bool, error)
}

type taskWatcherService struct {
	watcherRepo repository.TaskWatcherRepository
}

// NewTaskWatcherService creates a new task watcher service
func NewTaskWatcherService(watcherRepo repository.TaskWatcherRepository) TaskWatcherService {
	return &taskWatcherService{watcherRepo: watcherRepo}
}

func (s *taskWatcherService) Watch(ctx context.Context, taskID, userID string) error {
	watcher := &repository.TaskWatcher{
		TaskID: taskID,
		UserID: userID,
	}
	return s.watcherRepo.Add(ctx, watcher)
}

func (s *taskWatcherService) Unwatch(ctx context.Context, taskID, userID string) error {
	return s.watcherRepo.Remove(ctx, taskID, userID)
}

func (s *taskWatcherService) GetWatchers(ctx context.Context, taskID string) ([]string, error) {
	return s.watcherRepo.GetWatcherUserIDs(ctx, taskID)
}

func (s *taskWatcherService) IsWatching(ctx context.Context, taskID, userID string) (bool, error) {
	return s.watcherRepo.IsWatching(ctx, taskID, userID)
}
