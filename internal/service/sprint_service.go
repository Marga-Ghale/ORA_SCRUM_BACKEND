// internal/service/sprint_service.go
package service

import (
	"context"
	"log"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

type SprintService interface {
	Create(ctx context.Context, sprint *repository.Sprint, userID string) error
	Get(ctx context.Context, sprintID, userID string) (*repository.Sprint, error)
	ListByProject(ctx context.Context, projectID, userID string) ([]*repository.Sprint, error)
	GetActiveSprint(ctx context.Context, projectID, userID string) (*repository.Sprint, error)
	Update(ctx context.Context, sprint *repository.Sprint, userID string) error
	Delete(ctx context.Context, sprintID, userID string) error
	StartSprint(ctx context.Context, sprintID, userID string) (*SprintStartResponse, error)
	CompleteSprint(ctx context.Context, sprintID, userID string) error
	CompleteSprintWithOptions(ctx context.Context, sprintID, userID string, options *SprintCompleteOptions) (*SprintCompleteResponse, error)
	GetSprintSummary(ctx context.Context, sprintID, userID string) (*SprintSummary, error)
}

// New types for sprint operations
type SprintStartResponse struct {
	Sprint          *repository.Sprint `json:"sprint"`
	CommittedTasks  int                `json:"committedTasks"`
	CommittedPoints int                `json:"committedPoints"`
	Warning         string             `json:"warning,omitempty"`
}

type SprintCompleteOptions struct {
	MoveIncompleteTo string `json:"moveIncompleteTo"` // "backlog", "next_sprint", or sprint ID
}

type SprintCompleteResponse struct {
	Sprint           *repository.Sprint `json:"sprint"`
	CompletedTasks   int                `json:"completedTasks"`
	CompletedPoints  int                `json:"completedPoints"`
	IncompleteTasks  int                `json:"incompleteTasks"`
	IncompletePoints int                `json:"incompletePoints"`
	TasksMovedTo     string             `json:"tasksMovedTo,omitempty"`
	MovedTaskIDs     []string           `json:"movedTaskIds,omitempty"`
}
type SprintSummary struct {
	SprintID         string `json:"sprintId"`
	Status           string `json:"status"`
	CommittedTasks   int    `json:"committedTasks"`
	CommittedPoints  int    `json:"committedPoints"`
	CompletedTasks   int    `json:"completedTasks"`
	CompletedPoints  int    `json:"completedPoints"`
	IncompleteTasks  int    `json:"incompleteTasks"`
	IncompletePoints int    `json:"incompletePoints"`
	AddedTasks       int    `json:"addedTasks"`
	AddedPoints      int    `json:"addedPoints"`
	RemovedTasks     int    `json:"removedTasks"`
	RemovedPoints    int    `json:"removedPoints"`
	DaysRemaining    int    `json:"daysRemaining"`
	DaysElapsed      int    `json:"daysElapsed"`
}

type sprintService struct {
	sprintRepo     repository.SprintRepository
	projectRepo    repository.ProjectRepository
	taskRepo       repository.TaskRepository
	commitmentRepo repository.SprintCommitmentRepository
	goalRepo       repository.GoalRepository  
	memberSvc      MemberService

}

func NewSprintService(
	sprintRepo repository.SprintRepository,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	commitmentRepo repository.SprintCommitmentRepository,
	goalRepo repository.GoalRepository,  
	memberSvc MemberService,
) SprintService {
	return &sprintService{
		sprintRepo:     sprintRepo,
		projectRepo:    projectRepo,
		taskRepo:       taskRepo,
		commitmentRepo: commitmentRepo,
		goalRepo:       goalRepo, 
		memberSvc:      memberSvc,
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


func (s *sprintService) StartSprint(ctx context.Context, sprintID, userID string) (*SprintStartResponse, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Check if another sprint is already active
	activeSprint, err := s.sprintRepo.FindActiveSprint(ctx, sprint.ProjectID)
	if err != nil {
		return nil, err
	}
	if activeSprint != nil && activeSprint.ID != sprintID {
		return nil, ErrSprintAlreadyActive
	}

	// Get tasks in this sprint for commitment snapshot
	tasks, err := s.taskRepo.FindBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// Calculate commitment
	var committedTasks int
	var committedPoints int
	var taskIDs []string

	for _, task := range tasks {
		if task.ParentTaskID == nil { // Only count parent tasks
			committedTasks++
			if task.StoryPoints != nil {
				committedPoints += *task.StoryPoints
			}
			taskIDs = append(taskIDs, task.ID)
		}
	}

	// Save commitment snapshot
	commitment := &repository.SprintCommitment{
		SprintID:        sprintID,
		CommittedTasks:  committedTasks,
		CommittedPoints: committedPoints,
		TaskIDs:         taskIDs,
	}
	if err := s.commitmentRepo.SaveCommitment(ctx, commitment); err != nil {
		log.Printf("⚠️ Failed to save commitment snapshot: %v", err)
	}

	// Check if over-committing (compare with average velocity)
	var warning string
	// TODO: Get average velocity and compare
	// For now, just start the sprint

	// Start the sprint
	if err := s.sprintRepo.UpdateStatus(ctx, sprintID, "active"); err != nil {
		return nil, err
	}

	// Refresh sprint data
	sprint, _ = s.sprintRepo.FindByID(ctx, sprintID)

	return &SprintStartResponse{
		Sprint:          sprint,
		CommittedTasks:  committedTasks,
		CommittedPoints: committedPoints,
		Warning:         warning,
	}, nil
}



func (s *sprintService) CompleteSprint(ctx context.Context, sprintID, userID string) error {
	// Default behavior - move incomplete to backlog
	_, err := s.CompleteSprintWithOptions(ctx, sprintID, userID, &SprintCompleteOptions{
		MoveIncompleteTo: "backlog",
	})
	return err
}

func (s *sprintService) CompleteSprintWithOptions(ctx context.Context, sprintID, userID string, options *SprintCompleteOptions) (*SprintCompleteResponse, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Get all tasks in sprint
	tasks, err := s.taskRepo.FindBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	var completedTasks, completedPoints int
	var incompleteTasks, incompletePoints int
	var incompleteTaskIDs []string

	for _, task := range tasks {
		if task.ParentTaskID != nil {
			continue // Skip subtasks
		}

		points := 0
		if task.StoryPoints != nil {
			points = *task.StoryPoints
		}

		if task.Status == "done" {
			completedTasks++
			completedPoints += points
		} else {
			incompleteTasks++
			incompletePoints += points
			incompleteTaskIDs = append(incompleteTaskIDs, task.ID)
		}
	}

	// Handle incomplete tasks based on option
	var movedTo string
	if len(incompleteTaskIDs) > 0 && options != nil {
		switch options.MoveIncompleteTo {
		case "backlog":
			// Move to backlog (set sprint_id to NULL)
			for _, taskID := range incompleteTaskIDs {
				task, _ := s.taskRepo.FindByID(ctx, taskID)
				if task != nil {
					task.SprintID = nil
					s.taskRepo.Update(ctx, task)
				}
			}
			movedTo = "backlog"

		case "next_sprint":
			// Find or create next sprint
			sprints, _ := s.sprintRepo.FindByProjectID(ctx, sprint.ProjectID)
			var nextSprint *repository.Sprint
			for _, sp := range sprints {
				if sp.Status == "planning" && sp.StartDate.After(sprint.EndDate) {
					nextSprint = sp
					break
				}
			}
			if nextSprint != nil {
				s.taskRepo.BulkMoveToSprint(ctx, incompleteTaskIDs, nextSprint.ID)
				movedTo = nextSprint.Name
			} else {
				// No next sprint, move to backlog
				for _, taskID := range incompleteTaskIDs {
					task, _ := s.taskRepo.FindByID(ctx, taskID)
					if task != nil {
						task.SprintID = nil
						s.taskRepo.Update(ctx, task)
					}
				}
				movedTo = "backlog (no next sprint found)"
			}

		default:
			// Specific sprint ID provided
			if options.MoveIncompleteTo != "" {
				targetSprint, _ := s.sprintRepo.FindByID(ctx, options.MoveIncompleteTo)
				if targetSprint != nil {
					s.taskRepo.BulkMoveToSprint(ctx, incompleteTaskIDs, options.MoveIncompleteTo)
					movedTo = targetSprint.Name
				}
			}
		}
	}

	// Complete the sprint
	if err := s.sprintRepo.UpdateStatus(ctx, sprintID, "completed"); err != nil {
		return nil, err
	}

	s.updateSprintGoalsStatus(ctx, sprintID)


	// Refresh sprint data
	sprint, _ = s.sprintRepo.FindByID(ctx, sprintID)

	return &SprintCompleteResponse{
		Sprint:           sprint,
		CompletedTasks:   completedTasks,
		CompletedPoints:  completedPoints,
		IncompleteTasks:  incompleteTasks,
		IncompletePoints: incompletePoints,
		TasksMovedTo:     movedTo,
		MovedTaskIDs:     incompleteTaskIDs,
	}, nil
}

func (s *sprintService) GetSprintSummary(ctx context.Context, sprintID, userID string) (*SprintSummary, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberSvc.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Get commitment snapshot
	commitment, _ := s.commitmentRepo.GetCommitment(ctx, sprintID)
	committedTasks := 0
	committedPoints := 0
	if commitment != nil {
		committedTasks = commitment.CommittedTasks
		committedPoints = commitment.CommittedPoints
	}

	// Get current task stats
	tasks, _ := s.taskRepo.FindBySprintID(ctx, sprintID)
	var completedTasks, completedPoints int
	var incompleteTasks, incompletePoints int

	for _, task := range tasks {
		if task.ParentTaskID != nil {
			continue
		}
		points := 0
		if task.StoryPoints != nil {
			points = *task.StoryPoints
		}
		if task.Status == "done" {
			completedTasks++
			completedPoints += points
		} else {
			incompleteTasks++
			incompletePoints += points
		}
	}

	// Get scope changes
	addedTasks, addedPoints, _ := s.commitmentRepo.GetAddedTasksCount(ctx, sprintID)
	removedTasks, removedPoints, _ := s.commitmentRepo.GetRemovedTasksCount(ctx, sprintID)

	// Calculate days
	now := time.Now()
	daysElapsed := int(now.Sub(sprint.StartDate).Hours() / 24)
	daysRemaining := int(sprint.EndDate.Sub(now).Hours() / 24)
	if daysElapsed < 0 {
		daysElapsed = 0
	}
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	return &SprintSummary{
		SprintID:         sprintID,
		Status:           sprint.Status,
		CommittedTasks:   committedTasks,
		CommittedPoints:  committedPoints,
		CompletedTasks:   completedTasks,
		CompletedPoints:  completedPoints,
		IncompleteTasks:  incompleteTasks,
		IncompletePoints: incompletePoints,
		AddedTasks:       addedTasks,
		AddedPoints:      addedPoints,
		RemovedTasks:     removedTasks,
		RemovedPoints:    removedPoints,
		DaysRemaining:    daysRemaining,
		DaysElapsed:      daysElapsed,
	}, nil
}



func (s *sprintService) updateSprintGoalsStatus(ctx context.Context, sprintID string) {
	if s.goalRepo == nil {
		return
	}

	goals, err := s.goalRepo.FindBySprint(ctx, sprintID)
	if err != nil {
		log.Printf("⚠️ Failed to get sprint goals: %v", err)
		return
	}

	for _, goal := range goals {
		var newStatus string

		// Determine goal status based on progress
		if goal.Progress >= 100 {
			newStatus = "completed"
		} else if goal.Progress >= 70 {
			newStatus = "partially_met"
		} else {
			newStatus = "not_met"
		}

		// Update goal status
		if err := s.goalRepo.UpdateStatus(ctx, goal.ID, newStatus); err != nil {
			log.Printf("⚠️ Failed to update goal %s status: %v", goal.ID, err)
		} else {
			log.Printf("✅ Goal '%s' marked as %s (%.0f%% complete)", goal.Title, newStatus, goal.Progress)
		}
	}
}