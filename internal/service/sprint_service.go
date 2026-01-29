// internal/service/sprint_service.go
package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

type SprintService interface {
	Create(ctx context.Context, sprint *repository.Sprint, userID string) error
	Get(ctx context.Context, sprintID, userID string) (*repository.Sprint, error)
	ListByProject(ctx context.Context, projectID, userID string) ([]*repository.Sprint, error)
	GetActiveSprint(ctx context.Context, projectID, userID string) (*repository.Sprint, error)
	Update(ctx context.Context, sprint *repository.Sprint, userID string) error
	Delete(ctx context.Context, sprintID, userID string) error
	StartSprint(ctx context.Context, sprintID, userID string) error
	CompleteSprint(ctx context.Context, sprintID, userID string) error
}

type sprintService struct {
	sprintRepo  repository.SprintRepository
	projectRepo repository.ProjectRepository
	memberSvc   MemberService
}

func NewSprintService(
	sprintRepo repository.SprintRepository,
	projectRepo repository.ProjectRepository,
	memberSvc MemberService,
) SprintService {
	return &sprintService{
		sprintRepo:  sprintRepo,
		projectRepo: projectRepo,
		memberSvc:   memberSvc,
	}
}

func (s *sprintService) Create(ctx context.Context, sprint *repository.Sprint, userID string) error {
	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	sprint.CreatedBy = userID
	return s.sprintRepo.Create(ctx, sprint)
}

func (s *sprintService) Get(ctx context.Context, sprintID, userID string) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return sprint, nil
}

func (s *sprintService) ListByProject(ctx context.Context, projectID, userID string) ([]*repository.Sprint, error) {
	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.sprintRepo.FindByProjectID(ctx, projectID)
}

func (s *sprintService) GetActiveSprint(ctx context.Context, projectID, userID string) (*repository.Sprint, error) {
	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.sprintRepo.FindActiveSprint(ctx, projectID)
}

func (s *sprintService) Update(
	ctx context.Context,
	sprint *repository.Sprint,
	userID string,
) error {

	existing, err := s.sprintRepo.FindByID(ctx, sprint.ID)
	if err != nil || existing == nil {
		return ErrNotFound
	}

	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(
		ctx,
		EntityTypeProject,
		existing.ProjectID,
		userID,
	)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	// ✅ Name (string)
	if sprint.Name != "" {
		existing.Name = sprint.Name
	}

	// ✅ Goal (*string)
	if sprint.Goal != nil {
		existing.Goal = sprint.Goal
	}

	// ✅ Dates
	if !sprint.StartDate.IsZero() {
		existing.StartDate = sprint.StartDate
	}

	if !sprint.EndDate.IsZero() {
		existing.EndDate = sprint.EndDate
	}

	// ✅ Status
	if sprint.Status != "" {
		existing.Status = sprint.Status
	}

	return s.sprintRepo.Update(ctx, existing)
}


func (s *sprintService) Delete(ctx context.Context, sprintID, userID string) error {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return ErrNotFound
	}

	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	return s.sprintRepo.Delete(ctx, sprintID)
}

func (s *sprintService) StartSprint(ctx context.Context, sprintID, userID string) error {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return ErrNotFound
	}

	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	return s.sprintRepo.UpdateStatus(ctx, sprintID, "active")
}

func (s *sprintService) CompleteSprint(ctx context.Context, sprintID, userID string) error {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return ErrNotFound
	}

	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	return s.sprintRepo.UpdateStatus(ctx, sprintID, "completed")
}