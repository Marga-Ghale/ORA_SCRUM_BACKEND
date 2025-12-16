package service

import (
	"context"
	"fmt"

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
	// addedBy, _ := s.userRepo.FindByID(ctx, addedByID)
	// workspace, _ := s.workspaceRepo.FindByID(ctx, team.WorkspaceID)

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

	// // Send email
	// if s.emailSvc != nil && user != nil && addedBy != nil && workspace != nil {
	// 	s.emailSvc.SendTeamInvitation(user.Email, email.TeamInvitationData{
	// 		UserName:      user.Name,
	// 		TeamName:      team.Name,
	// 		WorkspaceName: workspace.Name,
	// 		AddedBy:       addedBy.Name,
	// 		TeamURL:       fmt.Sprintf("/workspaces/%s/teams/%s", team.WorkspaceID, teamID),
	// 	})
	// }

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
