package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Label Service
// ============================================

type LabelService interface {
	Create(ctx context.Context, projectID, name, color string) (*repository.Label, error)
	GetByID(ctx context.Context, id string) (*repository.Label, error)
	ListByProject(ctx context.Context, projectID string) ([]*repository.Label, error)
	Update(ctx context.Context, id string, name, color *string) (*repository.Label, error)
	Delete(ctx context.Context, id string) error
}

type labelService struct {
	labelRepo repository.LabelRepository
}

func NewLabelService(labelRepo repository.LabelRepository) LabelService {
	return &labelService{labelRepo: labelRepo}
}

func (s *labelService) Create(ctx context.Context, projectID, name, color string) (*repository.Label, error) {
	// Check for duplicate name in project
	existing, _ := s.labelRepo.FindByName(ctx, projectID, name)
	if existing != nil {
		return nil, ErrConflict
	}

	label := &repository.Label{
		ProjectID: projectID,
		Name:      name,
		Color:     color,
	}

	if err := s.labelRepo.Create(ctx, label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *labelService) GetByID(ctx context.Context, id string) (*repository.Label, error) {
	label, err := s.labelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if label == nil {
		return nil, ErrNotFound
	}
	return label, nil
}

func (s *labelService) ListByProject(ctx context.Context, projectID string) ([]*repository.Label, error) {
	return s.labelRepo.FindByProjectID(ctx, projectID)
}

func (s *labelService) Update(ctx context.Context, id string, name, color *string) (*repository.Label, error) {
	label, err := s.labelRepo.FindByID(ctx, id)
	if err != nil || label == nil {
		return nil, ErrNotFound
	}

	if name != nil {
		// Check for duplicate name
		existing, _ := s.labelRepo.FindByName(ctx, label.ProjectID, *name)
		if existing != nil && existing.ID != id {
			return nil, ErrConflict
		}
		label.Name = *name
	}
	if color != nil {
		label.Color = *color
	}

	if err := s.labelRepo.Update(ctx, label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *labelService) Delete(ctx context.Context, id string) error {
	return s.labelRepo.Delete(ctx, id)
}
