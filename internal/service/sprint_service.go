package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Sprint Service
// ============================================

type SprintService interface {
	Create(ctx context.Context, projectID, name string, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error)
	GetByID(ctx context.Context, id string) (*repository.Sprint, error)
	ListByProject(ctx context.Context, projectID string) ([]*repository.Sprint, error)
	GetActive(ctx context.Context, projectID string) (*repository.Sprint, error)
	Update(ctx context.Context, id string, name, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error)
	Delete(ctx context.Context, id string) error
	Start(ctx context.Context, id, userID string) (*repository.Sprint, error)
	Complete(ctx context.Context, id, moveIncomplete, userID string) (*repository.Sprint, error)
}

type sprintService struct {
	sprintRepo  repository.SprintRepository
	taskRepo    repository.TaskRepository
	projectRepo repository.ProjectRepository
	notifSvc    *notification.Service
}

func NewSprintService(sprintRepo repository.SprintRepository, taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository, notifSvc *notification.Service) SprintService {
	return &sprintService{
		sprintRepo:  sprintRepo,
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		notifSvc:    notifSvc,
	}
}

func (s *sprintService) Create(ctx context.Context, projectID, name string, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error) {
	sprint := &repository.Sprint{
		ProjectID: projectID,
		Name:      name,
		Goal:      goal,
		Status:    "planning", // lowercase
		StartDate: startDate,
		EndDate:   endDate,
	}

	if err := s.sprintRepo.Create(ctx, sprint); err != nil {
		return nil, err
	}
	return sprint, nil
}

func (s *sprintService) GetByID(ctx context.Context, id string) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sprint == nil {
		return nil, ErrNotFound
	}
	return sprint, nil
}

func (s *sprintService) ListByProject(ctx context.Context, projectID string) ([]*repository.Sprint, error) {
	return s.sprintRepo.FindByProjectID(ctx, projectID)
}

func (s *sprintService) GetActive(ctx context.Context, projectID string) (*repository.Sprint, error) {
	return s.sprintRepo.FindActive(ctx, projectID)
}

func (s *sprintService) Update(ctx context.Context, id string, name, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	if name != nil {
		sprint.Name = *name
	}
	if goal != nil {
		sprint.Goal = goal
	}
	if startDate != nil {
		sprint.StartDate = startDate
	}
	if endDate != nil {
		sprint.EndDate = endDate
	}

	if err := s.sprintRepo.Update(ctx, sprint); err != nil {
		return nil, err
	}
	return sprint, nil
}

func (s *sprintService) Delete(ctx context.Context, id string) error {
	return s.sprintRepo.Delete(ctx, id)
}

func (s *sprintService) Start(ctx context.Context, id, userID string) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	// Check if there's already an active sprint
	active, _ := s.sprintRepo.FindActive(ctx, sprint.ProjectID)
	if active != nil && active.ID != id {
		return nil, fmt.Errorf("another sprint is already active")
	}

	now := time.Now()
	sprint.Status = "active" // lowercase
	sprint.StartDate = &now

	if err := s.sprintRepo.Update(ctx, sprint); err != nil {
		return nil, err
	}

	// Send notifications
	if s.notifSvc != nil && s.projectRepo != nil {
		memberIDs, _ := s.projectRepo.FindMemberUserIDs(ctx, sprint.ProjectID)
		if len(memberIDs) > 0 {
			s.notifSvc.SendSprintStartedToMembers(ctx, memberIDs, sprint.Name, sprint.ID, sprint.ProjectID)
		}
	}

	return sprint, nil
}

func (s *sprintService) Complete(ctx context.Context, id, moveIncomplete, userID string) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	totalTasks, completedTasks, _ := s.taskRepo.CountBySprintID(ctx, id)

	now := time.Now()
	sprint.Status = "completed" // lowercase
	sprint.EndDate = &now

	// Move incomplete tasks
	if moveIncomplete != "" {
		tasks, _ := s.taskRepo.FindBySprintID(ctx, id)
		for _, task := range tasks {
			if task.Status != "done" && task.Status != "cancelled" { // lowercase
				if moveIncomplete == "backlog" {
					task.SprintID = nil
				} else {
					task.SprintID = &moveIncomplete
				}
				s.taskRepo.Update(ctx, task)
			}
		}
	}

	if err := s.sprintRepo.Update(ctx, sprint); err != nil {
		return nil, err
	}

	// Send notifications
	if s.notifSvc != nil && s.projectRepo != nil {
		memberIDs, _ := s.projectRepo.FindMemberUserIDs(ctx, sprint.ProjectID)
		if len(memberIDs) > 0 {
			s.notifSvc.SendSprintCompletedToMembers(ctx, memberIDs, sprint.Name, sprint.ID, sprint.ProjectID, completedTasks, totalTasks)
		}
	}

	return sprint, nil
}
