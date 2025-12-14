package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Space Service
// ============================================

type SpaceService interface {
	Create(ctx context.Context, workspaceID, name string, description, icon, color *string) (*repository.Space, error)
	GetByID(ctx context.Context, id string) (*repository.Space, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*repository.Space, error)
	Update(ctx context.Context, id string, name, description, icon, color *string) (*repository.Space, error)
	Delete(ctx context.Context, id string) error
}

type spaceService struct {
	spaceRepo repository.SpaceRepository
}

func NewSpaceService(spaceRepo repository.SpaceRepository) SpaceService {
	return &spaceService{spaceRepo: spaceRepo}
}

func (s *spaceService) Create(ctx context.Context, workspaceID, name string, description, icon, color *string) (*repository.Space, error) {
	space := &repository.Space{
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		Icon:        icon,
		Color:       color,
	}

	if err := s.spaceRepo.Create(ctx, space); err != nil {
		return nil, err
	}
	return space, nil
}

func (s *spaceService) GetByID(ctx context.Context, id string) (*repository.Space, error) {
	space, err := s.spaceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if space == nil {
		return nil, ErrNotFound
	}
	return space, nil
}

func (s *spaceService) ListByWorkspace(ctx context.Context, workspaceID string) ([]*repository.Space, error) {
	return s.spaceRepo.FindByWorkspaceID(ctx, workspaceID)
}

func (s *spaceService) Update(ctx context.Context, id string, name, description, icon, color *string) (*repository.Space, error) {
	space, err := s.spaceRepo.FindByID(ctx, id)
	if err != nil || space == nil {
		return nil, ErrNotFound
	}

	if name != nil {
		space.Name = *name
	}
	if description != nil {
		space.Description = description
	}
	if icon != nil {
		space.Icon = icon
	}
	if color != nil {
		space.Color = color
	}

	if err := s.spaceRepo.Update(ctx, space); err != nil {
		return nil, err
	}
	return space, nil
}

func (s *spaceService) Delete(ctx context.Context, id string) error {
	return s.spaceRepo.Delete(ctx, id)
}
