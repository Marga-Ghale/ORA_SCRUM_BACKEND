package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
)

type WorkspaceService interface {
	Create(ctx context.Context, userID, name string, description, icon, color, visibility *string, allowedUsers, allowedTeams []string) (*repository.Workspace, error)
	GetByID(ctx context.Context, id string) (*repository.Workspace, error)
	List(ctx context.Context, userID string) ([]*repository.Workspace, error)
	Update(ctx context.Context, id string, name, description, icon, color, visibility *string, allowedUsers, allowedTeams *[]string) (*repository.Workspace, error)
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, workspaceID, email, role, inviterID string) error
	AddMemberByID(ctx context.Context, workspaceID, userID, role, inviterID string) error
	ListMembers(ctx context.Context, workspaceID string) ([]*repository.WorkspaceMember, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
	HasAccess(ctx context.Context, workspaceID, userID string) (bool, error)
}

type workspaceService struct {
	workspaceRepo repository.WorkspaceRepository
	userRepo      repository.UserRepository
	notifSvc      *notification.Service
	broadcaster   *socket.Broadcaster // ✅ NEW: Added broadcaster
}

func NewWorkspaceService(workspaceRepo repository.WorkspaceRepository, userRepo repository.UserRepository, notifSvc *notification.Service,
		broadcaster *socket.Broadcaster, // ✅ ADD
	) WorkspaceService {
	return &workspaceService{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		notifSvc:      notifSvc,
		broadcaster:  broadcaster,
	}
}

// ✅ NEW: SetBroadcaster sets the broadcaster for real-time updates
func (s *workspaceService) SetBroadcaster(b *socket.Broadcaster) {
	s.broadcaster = b
}

func (s *workspaceService) Create(ctx context.Context, userID, name string, description, icon, color, visibility *string, allowedUsers, allowedTeams []string) (*repository.Workspace, error) {
	defaultVisibility := "private"
	if visibility == nil {
		visibility = &defaultVisibility
	}

	workspace := &repository.Workspace{
		Name:         name,
		Description:  description,
		Icon:         icon,
		Color:        color,
		OwnerID:      userID,
		Visibility:   visibility,
		AllowedUsers: allowedUsers,
		AllowedTeams: allowedTeams,
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	member := &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        "owner",
	}
	if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	// ✅ NEW: Broadcast workspace creation to the creator
	// Note: New workspaces only have the creator as member, so we send directly to them
	if s.broadcaster != nil {
		s.broadcaster.BroadcastWorkspaceCreated(userID, map[string]interface{}{
			"id":          workspace.ID,
			"name":        workspace.Name,
			"description": workspace.Description,
			"icon":        workspace.Icon,
			"color":       workspace.Color,
			"ownerId":     workspace.OwnerID,
			"visibility":  workspace.Visibility,
			"createdAt":   workspace.CreatedAt,
			"updatedAt":   workspace.UpdatedAt,
		})
	}

	return workspace, nil
}

func (s *workspaceService) GetByID(ctx context.Context, id string) (*repository.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrNotFound
	}
	return workspace, nil
}

func (s *workspaceService) List(ctx context.Context, userID string) ([]*repository.Workspace, error) {
	return s.workspaceRepo.FindByUserID(ctx, userID)
}

func (s *workspaceService) Update(ctx context.Context, id string, name, description, icon, color, visibility *string, allowedUsers, allowedTeams *[]string) (*repository.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, id)
	if err != nil || workspace == nil {
		return nil, ErrNotFound
	}

	if name != nil {
		workspace.Name = *name
	}
	workspace.Description = description
	workspace.Icon = icon
	workspace.Color = color
	workspace.Visibility = visibility

	if allowedUsers != nil {
		workspace.AllowedUsers = *allowedUsers
	}
	if allowedTeams != nil {
		workspace.AllowedTeams = *allowedTeams
	}

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	// ✅ NEW: Broadcast workspace update to all workspace members
	if s.broadcaster != nil {
		s.broadcaster.BroadcastWorkspaceUpdated(workspace.ID, map[string]interface{}{
			"id":          workspace.ID,
			"name":        workspace.Name,
			"description": workspace.Description,
			"icon":        workspace.Icon,
			"color":       workspace.Color,
			"ownerId":     workspace.OwnerID,
			"visibility":  workspace.Visibility,
			"updatedAt":   workspace.UpdatedAt,
		}, "")
	}

	return workspace, nil
}

func (s *workspaceService) Delete(ctx context.Context, id string) error {
	// ✅ NEW: Broadcast deletion BEFORE actually deleting
	// so clients are still in the room
	if s.broadcaster != nil {
		s.broadcaster.BroadcastWorkspaceDeleted(id, "")
	}

	return s.workspaceRepo.Delete(ctx, id)
}

func (s *workspaceService) AddMember(ctx context.Context, workspaceID, email, role, inviterID string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	existing, _ := s.workspaceRepo.FindMember(ctx, workspaceID, user.ID)
	if existing != nil {
		return ErrConflict
	}

	member := &repository.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      user.ID,
		Role:        role,
	}

	if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
		return err
	}

	workspace, _ := s.workspaceRepo.FindByID(ctx, workspaceID)
	if workspace != nil && s.notifSvc != nil {
		inviterName := ""
		if inviter, _ := s.userRepo.FindByID(ctx, inviterID); inviter != nil {
			inviterName = inviter.Name
		}
		s.notifSvc.SendWorkspaceInvitation(ctx, user.ID, workspace.Name, workspaceID, inviterName)
	}

	// ✅ NEW: Broadcast member added to workspace
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(workspaceID, socket.MessageMemberAdded, map[string]interface{}{
			"workspaceId": workspaceID,
			"userId":      user.ID,
			"role":        role,
			"userName":    user.Name,
			"userEmail":   user.Email,
		}, "")
	}

	return nil
}

func (s *workspaceService) AddMemberByID(ctx context.Context, workspaceID, userID, role, inviterID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	existing, _ := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if existing != nil {
		return ErrConflict
	}

	member := &repository.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
	}

	if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
		return err
	}

	workspace, _ := s.workspaceRepo.FindByID(ctx, workspaceID)
	if workspace != nil && s.notifSvc != nil {
		inviterName := ""
		if inviter, _ := s.userRepo.FindByID(ctx, inviterID); inviter != nil {
			inviterName = inviter.Name
		}
		s.notifSvc.SendWorkspaceInvitation(ctx, userID, workspace.Name, workspaceID, inviterName)
	}

	// ✅ NEW: Broadcast member added to workspace
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(workspaceID, socket.MessageMemberAdded, map[string]interface{}{
			"workspaceId": workspaceID,
			"userId":      user.ID,
			"role":        role,
			"userName":    user.Name,
			"userEmail":   user.Email,
		}, "")
	}

	return nil
}

func (s *workspaceService) ListMembers(ctx context.Context, workspaceID string) ([]*repository.WorkspaceMember, error) {
	return s.workspaceRepo.FindMembers(ctx, workspaceID)
}

func (s *workspaceService) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	if err := s.workspaceRepo.UpdateMemberRole(ctx, workspaceID, userID, role); err != nil {
		return err
	}

	// ✅ NEW: Broadcast role update
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(workspaceID, socket.MessageMemberAdded, map[string]interface{}{
			"workspaceId": workspaceID,
			"userId":      userID,
			"role":        role,
			"action":      "role_updated",
		}, "")
	}

	return nil
}

func (s *workspaceService) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	if err := s.workspaceRepo.RemoveMember(ctx, workspaceID, userID); err != nil {
		return err
	}

	// ✅ NEW: Broadcast member removed
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(workspaceID, socket.MessageMemberRemoved, map[string]interface{}{
			"workspaceId": workspaceID,
			"userId":      userID,
		}, "")
	}

	return nil
}

func (s *workspaceService) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	member, err := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}
	return member != nil, nil
}

func (s *workspaceService) HasAccess(ctx context.Context, workspaceID, userID string) (bool, error) {
	return s.workspaceRepo.HasAccess(ctx, workspaceID, userID)
}