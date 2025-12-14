package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Project Service
// ============================================

type ProjectService interface {
	Create(ctx context.Context, spaceID, creatorID, name, key string, description, icon, color, leadID *string) (*repository.Project, error)
	GetByID(ctx context.Context, id string) (*repository.Project, error)
	ListBySpace(ctx context.Context, spaceID string) ([]*repository.Project, error)
	Update(ctx context.Context, id string, name, key, description, icon, color, leadID *string) (*repository.Project, error)
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, projectID, userID, role, inviterID string) error
	ListMembers(ctx context.Context, projectID string) ([]*repository.ProjectMember, error)
	GetMemberUserIDs(ctx context.Context, projectID string) ([]string, error)
	RemoveMember(ctx context.Context, projectID, userID string) error
	IsMember(ctx context.Context, projectID, userID string) (bool, error)
}

type projectService struct {
	projectRepo repository.ProjectRepository
	userRepo    repository.UserRepository
	notifSvc    *notification.Service
}

func NewProjectService(projectRepo repository.ProjectRepository, userRepo repository.UserRepository, notifSvc *notification.Service) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
		notifSvc:    notifSvc,
	}
}

func (s *projectService) Create(ctx context.Context, spaceID, creatorID, name, key string, description, icon, color, leadID *string) (*repository.Project, error) {
	existing, _ := s.projectRepo.FindByKey(ctx, key)
	if existing != nil {
		return nil, ErrConflict
	}

	project := &repository.Project{
		SpaceID:     spaceID,
		Name:        name,
		Key:         key,
		Description: description,
		Icon:        icon,
		Color:       color,
		LeadID:      leadID,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	member := &repository.ProjectMember{
		ProjectID: project.ID,
		UserID:    creatorID,
		Role:      "LEAD",
	}
	s.projectRepo.AddMember(ctx, member)

	return project, nil
}

func (s *projectService) GetByID(ctx context.Context, id string) (*repository.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrNotFound
	}
	return project, nil
}

func (s *projectService) ListBySpace(ctx context.Context, spaceID string) ([]*repository.Project, error) {
	return s.projectRepo.FindBySpaceID(ctx, spaceID)
}

func (s *projectService) Update(ctx context.Context, id string, name, key, description, icon, color, leadID *string) (*repository.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil || project == nil {
		return nil, ErrNotFound
	}

	// Required fields - only update if provided
	if name != nil {
		project.Name = *name
	}
	if key != nil {
		existing, _ := s.projectRepo.FindByKey(ctx, *key)
		if existing != nil && existing.ID != id {
			return nil, ErrConflict
		}
		project.Key = *key
	}

	// Nullable fields - always update to allow clearing (setting to NULL)
	project.Description = description
	project.Icon = icon
	project.Color = color
	project.LeadID = leadID

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}
func (s *projectService) Delete(ctx context.Context, id string) error {
	return s.projectRepo.Delete(ctx, id)
}

func (s *projectService) AddMember(ctx context.Context, projectID, userID, role, inviterID string) error {
	existing, _ := s.projectRepo.FindMember(ctx, projectID, userID)
	if existing != nil {
		return ErrConflict
	}

	member := &repository.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}

	if err := s.projectRepo.AddMember(ctx, member); err != nil {
		return err
	}

	project, _ := s.projectRepo.FindByID(ctx, projectID)
	if project != nil && s.notifSvc != nil {
		inviterName := ""
		if inviter, _ := s.userRepo.FindByID(ctx, inviterID); inviter != nil {
			inviterName = inviter.Name
		}
		s.notifSvc.SendProjectInvitation(ctx, userID, project.Name, projectID, inviterName)
	}

	return nil
}

func (s *projectService) ListMembers(ctx context.Context, projectID string) ([]*repository.ProjectMember, error) {
	members, err := s.projectRepo.FindMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		user, _ := s.userRepo.FindByID(ctx, m.UserID)
		m.User = user
	}

	return members, nil
}

func (s *projectService) GetMemberUserIDs(ctx context.Context, projectID string) ([]string, error) {
	return s.projectRepo.FindMemberUserIDs(ctx, projectID)
}

func (s *projectService) RemoveMember(ctx context.Context, projectID, userID string) error {
	return s.projectRepo.RemoveMember(ctx, projectID, userID)
}

func (s *projectService) IsMember(ctx context.Context, projectID, userID string) (bool, error) {
	member, err := s.projectRepo.FindMember(ctx, projectID, userID)
	if err != nil {
		return false, err
	}
	return member != nil, nil
}
