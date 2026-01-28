package service

import (
	"context"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
)

// ============================================
// GOAL SERVICE INTERFACE
// ============================================

type GoalService interface {
	// Goal CRUD
	Create(ctx context.Context, req *CreateGoalRequest) (*repository.Goal, error)
	GetByID(ctx context.Context, goalID, userID string) (*repository.Goal, error)
	ListByWorkspace(ctx context.Context, workspaceID, userID string, goalType, status *string) ([]*repository.Goal, error)
	ListByProject(ctx context.Context, projectID, userID string) ([]*repository.Goal, error)
	ListBySprint(ctx context.Context, sprintID, userID string) ([]*repository.Goal, error)
	Update(ctx context.Context, goalID, userID string, req *UpdateGoalRequest) (*repository.Goal, error)
	UpdateProgress(ctx context.Context, goalID, userID string, currentValue float64) error
	UpdateStatus(ctx context.Context, goalID, userID, status string) error
	Delete(ctx context.Context, goalID, userID string) error

	// Key Results
	AddKeyResult(ctx context.Context, goalID, userID string, req *CreateKeyResultRequest) (*repository.KeyResult, error)
	UpdateKeyResult(ctx context.Context, keyResultID, userID string, req *UpdateKeyResultRequest) error
	UpdateKeyResultProgress(ctx context.Context, keyResultID, userID string, currentValue float64) error
	DeleteKeyResult(ctx context.Context, keyResultID, userID string) error

	// Task Linking
	LinkTask(ctx context.Context, goalID, taskID, userID string) error
	UnlinkTask(ctx context.Context, goalID, taskID, userID string) error
	GetGoalsByTask(ctx context.Context, taskID, userID string) ([]*repository.Goal, error)

	// Analytics
	GetGoalProgress(ctx context.Context, goalID string) (float64, error)
	GetSprintGoalsSummary(ctx context.Context, sprintID string) (*SprintGoalsSummary, error)

	// Auto-update (called when tasks complete)
	RecalculateGoalProgress(ctx context.Context, goalID string) error
}

// ============================================
// REQUEST/RESPONSE MODELS
// ============================================

type CreateGoalRequest struct {
	WorkspaceID string     `json:"workspaceId" binding:"required"`
	ProjectID   *string    `json:"projectId,omitempty"`
	SprintID    *string    `json:"sprintId,omitempty"`
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description,omitempty"`
	GoalType    string     `json:"goalType"` // sprint, project, quarterly, annual
	TargetValue *float64   `json:"targetValue,omitempty"`
	Unit        *string    `json:"unit,omitempty"` // story_points, tasks, percentage, custom
	StartDate   *time.Time `json:"startDate,omitempty"`
	TargetDate  *time.Time `json:"targetDate,omitempty"`
	OwnerID     *string    `json:"ownerId,omitempty"`
	CreatedBy   string     `json:"-"`
}

type UpdateGoalRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	GoalType    *string    `json:"goalType,omitempty"`
	Status      *string    `json:"status,omitempty"`
	TargetValue *float64   `json:"targetValue,omitempty"`
	Unit        *string    `json:"unit,omitempty"`
	StartDate   *time.Time `json:"startDate,omitempty"`
	TargetDate  *time.Time `json:"targetDate,omitempty"`
	OwnerID     *string    `json:"ownerId,omitempty"`
}

type CreateKeyResultRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description *string  `json:"description,omitempty"`
	TargetValue float64  `json:"targetValue" binding:"required"`
	Unit        *string  `json:"unit,omitempty"`
	Weight      *float64 `json:"weight,omitempty"`
}

type UpdateKeyResultRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	TargetValue *float64 `json:"targetValue,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	Weight      *float64 `json:"weight,omitempty"`
	Status      *string  `json:"status,omitempty"`
}

type SprintGoalsSummary struct {
	SprintID        string            `json:"sprintId"`
	GoalsCompleted  int               `json:"goalsCompleted"`
	GoalsTotal      int               `json:"goalsTotal"`
	CompletionRate  float64           `json:"completionRate"`
	Goals           []*repository.Goal `json:"goals"`
}

// ============================================
// IMPLEMENTATION
// ============================================

type goalService struct {
	goalRepo        repository.GoalRepository
	taskRepo        repository.TaskRepository
	sprintRepo      repository.SprintRepository
	memberService   MemberService
	notificationSvc *notification.Service
	broadcaster     *socket.Broadcaster
}

func NewGoalService(
	goalRepo repository.GoalRepository,
	taskRepo repository.TaskRepository,
	sprintRepo repository.SprintRepository,
	memberService MemberService,
	notificationSvc *notification.Service,
	broadcaster *socket.Broadcaster,
) GoalService {
	return &goalService{
		goalRepo:        goalRepo,
		taskRepo:        taskRepo,
		sprintRepo:      sprintRepo,
		memberService:   memberService,
		notificationSvc: notificationSvc,
		broadcaster:     broadcaster,
	}
}

// ============================================
// GOAL CRUD
// ============================================

func (s *goalService) Create(ctx context.Context, req *CreateGoalRequest) (*repository.Goal, error) {
	// Verify workspace access
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, req.WorkspaceID, req.CreatedBy)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Set defaults
	if req.GoalType == "" {
		req.GoalType = "sprint"
	}

	goal := &repository.Goal{
		WorkspaceID:  req.WorkspaceID,
		ProjectID:    req.ProjectID,
		SprintID:     req.SprintID,
		Title:        req.Title,
		Description:  req.Description,
		GoalType:     req.GoalType,
		Status:       "active",
		TargetValue:  req.TargetValue,
		CurrentValue: 0,
		Unit:         req.Unit,
		StartDate:    req.StartDate,
		TargetDate:   req.TargetDate,
		OwnerID:      req.OwnerID,
		CreatedBy:    &req.CreatedBy,
	}

	if err := s.goalRepo.Create(ctx, goal); err != nil {
		return nil, err
	}

	// Broadcast goal creation
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(req.WorkspaceID, "goal:created", map[string]interface{}{
			"goal": goal,
		}, "")
	}

	return goal, nil
}

func (s *goalService) GetByID(ctx context.Context, goalID, userID string) (*repository.Goal, error) {
	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	if goal == nil {
		return nil, ErrNotFound
	}

	// Verify access
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, goal.WorkspaceID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return goal, nil
}

func (s *goalService) ListByWorkspace(ctx context.Context, workspaceID, userID string, goalType, status *string) ([]*repository.Goal, error) {
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, workspaceID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.goalRepo.FindByWorkspace(ctx, workspaceID, goalType, status)
}

func (s *goalService) ListByProject(ctx context.Context, projectID, userID string) ([]*repository.Goal, error) {
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.goalRepo.FindByProject(ctx, projectID)
}

func (s *goalService) ListBySprint(ctx context.Context, sprintID, userID string) ([]*repository.Goal, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.goalRepo.FindBySprint(ctx, sprintID)
}

func (s *goalService) Update(ctx context.Context, goalID, userID string, req *UpdateGoalRequest) (*repository.Goal, error) {
	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil || goal == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, goal.WorkspaceID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Update fields
	if req.Title != nil {
		goal.Title = *req.Title
	}
	if req.Description != nil {
		goal.Description = req.Description
	}
	if req.GoalType != nil {
		goal.GoalType = *req.GoalType
	}
	if req.Status != nil {
		goal.Status = *req.Status
		if *req.Status == "completed" {
			now := time.Now()
			goal.CompletedAt = &now
		}
	}
	if req.TargetValue != nil {
		goal.TargetValue = req.TargetValue
	}
	if req.Unit != nil {
		goal.Unit = req.Unit
	}
	if req.StartDate != nil {
		goal.StartDate = req.StartDate
	}
	if req.TargetDate != nil {
		goal.TargetDate = req.TargetDate
	}
	if req.OwnerID != nil {
		goal.OwnerID = req.OwnerID
	}

	if err := s.goalRepo.Update(ctx, goal); err != nil {
		return nil, err
	}

	// Broadcast update
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(goal.WorkspaceID, "goal:updated", map[string]interface{}{
			"goal": goal,
		},"")
	}

	return goal, nil
}

func (s *goalService) UpdateProgress(ctx context.Context, goalID, userID string, currentValue float64) error {
	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil || goal == nil {
		return ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, goal.WorkspaceID, userID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	if err := s.goalRepo.UpdateProgress(ctx, goalID, currentValue); err != nil {
		return err
	}

	// Check if goal should auto-complete
	if goal.TargetValue != nil && currentValue >= *goal.TargetValue {
		s.goalRepo.UpdateStatus(ctx, goalID, "completed")
	}

	return nil
}

func (s *goalService) UpdateStatus(ctx context.Context, goalID, userID, status string) error {
	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil || goal == nil {
		return ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, goal.WorkspaceID, userID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	return s.goalRepo.UpdateStatus(ctx, goalID, status)
}

func (s *goalService) Delete(ctx context.Context, goalID, userID string) error {
	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil || goal == nil {
		return ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, goal.WorkspaceID, userID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	return s.goalRepo.Delete(ctx, goalID)
}

// ============================================
// KEY RESULTS
// ============================================

func (s *goalService) AddKeyResult(ctx context.Context, goalID, userID string, req *CreateKeyResultRequest) (*repository.KeyResult, error) {
	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil || goal == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, goal.WorkspaceID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	weight := 1.0
	if req.Weight != nil {
		weight = *req.Weight
	}

	kr := &repository.KeyResult{
		GoalID:       goalID,
		Title:        req.Title,
		Description:  req.Description,
		TargetValue:  req.TargetValue,
		CurrentValue: 0,
		Unit:         req.Unit,
		Status:       "pending",
		Weight:       weight,
	}

	if err := s.goalRepo.CreateKeyResult(ctx, kr); err != nil {
		return nil, err
	}

	return kr, nil
}

func (s *goalService) UpdateKeyResult(ctx context.Context, keyResultID, userID string, req *UpdateKeyResultRequest) error {
	// Get key result's goal to check access
	keyResults, err := s.goalRepo.FindKeyResultsByGoal(ctx, keyResultID)
	if err != nil || len(keyResults) == 0 {
		return ErrNotFound
	}

	// Find the specific key result
	var kr *repository.KeyResult
	for _, k := range keyResults {
		if k.ID == keyResultID {
			kr = k
			break
		}
	}
	if kr == nil {
		return ErrNotFound
	}

	// Update fields
	if req.Title != nil {
		kr.Title = *req.Title
	}
	if req.Description != nil {
		kr.Description = req.Description
	}
	if req.TargetValue != nil {
		kr.TargetValue = *req.TargetValue
	}
	if req.Unit != nil {
		kr.Unit = req.Unit
	}
	if req.Weight != nil {
		kr.Weight = *req.Weight
	}
	if req.Status != nil {
		kr.Status = *req.Status
	}

	return s.goalRepo.UpdateKeyResult(ctx, kr)
}

func (s *goalService) UpdateKeyResultProgress(ctx context.Context, keyResultID, userID string, currentValue float64) error {
	return s.goalRepo.UpdateKeyResultProgress(ctx, keyResultID, currentValue)
}

func (s *goalService) DeleteKeyResult(ctx context.Context, keyResultID, userID string) error {
	return s.goalRepo.DeleteKeyResult(ctx, keyResultID)
}

// ============================================
// TASK LINKING
// ============================================

func (s *goalService) LinkTask(ctx context.Context, goalID, taskID, userID string) error {
	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil || goal == nil {
		return ErrNotFound
	}

	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, goal.WorkspaceID, userID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	return s.goalRepo.LinkTask(ctx, goalID, taskID)
}

func (s *goalService) UnlinkTask(ctx context.Context, goalID, taskID, userID string) error {
	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil || goal == nil {
		return ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, goal.WorkspaceID, userID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	return s.goalRepo.UnlinkTask(ctx, goalID, taskID)
}

func (s *goalService) GetGoalsByTask(ctx context.Context, taskID, userID string) ([]*repository.Goal, error) {
	return s.goalRepo.GetGoalsByTask(ctx, taskID)
}

// ============================================
// ANALYTICS
// ============================================

func (s *goalService) GetGoalProgress(ctx context.Context, goalID string) (float64, error) {
	return s.goalRepo.GetGoalProgress(ctx, goalID)
}

func (s *goalService) GetSprintGoalsSummary(ctx context.Context, sprintID string) (*SprintGoalsSummary, error) {
	completed, total, err := s.goalRepo.GetSprintGoalsSummary(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	goals, err := s.goalRepo.FindBySprint(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	completionRate := 0.0
	if total > 0 {
		completionRate = float64(completed) / float64(total) * 100
	}

	return &SprintGoalsSummary{
		SprintID:       sprintID,
		GoalsCompleted: completed,
		GoalsTotal:     total,
		CompletionRate: completionRate,
		Goals:          goals,
	}, nil
}

// ============================================
// AUTO-UPDATE (Called when tasks complete)
// ============================================

func (s *goalService) RecalculateGoalProgress(ctx context.Context, goalID string) error {
	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil || goal == nil {
		return nil
	}

	// Get linked tasks
	taskIDs, err := s.goalRepo.GetLinkedTaskIDs(ctx, goalID)
	if err != nil {
		return err
	}

	if len(taskIDs) == 0 {
		return nil
	}

	// Count completed tasks or sum story points
	var completed float64
	var total float64

	for _, taskID := range taskIDs {
		task, err := s.taskRepo.FindByID(ctx, taskID)
		if err != nil || task == nil {
			continue
		}

		if goal.Unit != nil && *goal.Unit == "story_points" {
			if task.StoryPoints != nil {
				total += float64(*task.StoryPoints)
				if task.Status == "done" {
					completed += float64(*task.StoryPoints)
				}
			}
		} else {
			total++
			if task.Status == "done" {
				completed++
			}
		}
	}

	// Update goal progress
	return s.goalRepo.UpdateProgress(ctx, goalID, completed)
}