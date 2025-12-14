package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Workspace Service
// ============================================

type WorkspaceService interface {
	Create(ctx context.Context, userID, name string, description, icon, color *string) (*repository.Workspace, error)
	GetByID(ctx context.Context, id string) (*repository.Workspace, error)
	List(ctx context.Context, userID string) ([]*repository.Workspace, error)
	Update(ctx context.Context, id string, name, description, icon, color *string) (*repository.Workspace, error)
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, workspaceID, email, role, inviterID string) error
	AddMemberByID(ctx context.Context, workspaceID, userID, role, inviterID string) error
	ListMembers(ctx context.Context, workspaceID string) ([]*repository.WorkspaceMember, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
}

type workspaceService struct {
	workspaceRepo repository.WorkspaceRepository
	userRepo      repository.UserRepository
	notifSvc      *notification.Service
}

func NewWorkspaceService(workspaceRepo repository.WorkspaceRepository, userRepo repository.UserRepository, notifSvc *notification.Service) WorkspaceService {
	return &workspaceService{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		notifSvc:      notifSvc,
	}
}

func (s *workspaceService) Create(ctx context.Context, userID, name string, description, icon, color *string) (*repository.Workspace, error) {
	workspace := &repository.Workspace{
		Name:        name,
		Description: description,
		Icon:        icon,
		Color:       color,
		OwnerID:     userID,
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	member := &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        "OWNER",
	}
	if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, err
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

func (s *workspaceService) Update(ctx context.Context, id string, name, description, icon, color *string) (*repository.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, id)
	if err != nil || workspace == nil {
		return nil, ErrNotFound
	}

	// Name is required - only update if provided
	if name != nil {
		workspace.Name = *name
	}

	// Nullable fields - always update to allow clearing (setting to NULL)
	workspace.Description = description
	workspace.Icon = icon
	workspace.Color = color

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, err
	}
	return workspace, nil
}

func (s *workspaceService) Delete(ctx context.Context, id string) error {
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

	return nil
}

func (s *workspaceService) ListMembers(ctx context.Context, workspaceID string) ([]*repository.WorkspaceMember, error) {
	members, err := s.workspaceRepo.FindMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		user, _ := s.userRepo.FindByID(ctx, m.UserID)
		m.User = user
	}

	return members, nil
}

func (s *workspaceService) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	return s.workspaceRepo.UpdateMemberRole(ctx, workspaceID, userID, role)
}

func (s *workspaceService) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	return s.workspaceRepo.RemoveMember(ctx, workspaceID, userID)
}

func (s *workspaceService) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	member, err := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}
	return member != nil, nil
}
