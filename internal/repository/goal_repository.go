package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// ============================================
// GOAL MODELS
// ============================================

type Goal struct {
	ID           string     `json:"id" db:"id"`
	WorkspaceID  string     `json:"workspaceId" db:"workspace_id"`
	ProjectID    *string    `json:"projectId,omitempty" db:"project_id"`
	SprintID     *string    `json:"sprintId,omitempty" db:"sprint_id"`
	Title        string     `json:"title" db:"title"`
	Description  *string    `json:"description,omitempty" db:"description"`
	GoalType     string     `json:"goalType" db:"goal_type"` // sprint, project, quarterly, annual
	Status       string     `json:"status" db:"status"`       // active, completed, cancelled, at_risk
	TargetValue  *float64   `json:"targetValue,omitempty" db:"target_value"`
	CurrentValue float64    `json:"currentValue" db:"current_value"`
	Unit         *string    `json:"unit,omitempty" db:"unit"` // story_points, tasks, percentage, custom
	StartDate    *time.Time `json:"startDate,omitempty" db:"start_date"`
	TargetDate   *time.Time `json:"targetDate,omitempty" db:"target_date"`
	CompletedAt  *time.Time `json:"completedAt,omitempty" db:"completed_at"`
	OwnerID      *string    `json:"ownerId,omitempty" db:"owner_id"`
	CreatedBy    *string    `json:"createdBy,omitempty" db:"created_by"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`

	// Computed/Joined
	Progress    float64       `json:"progress"`              // 0-100
	KeyResults  []*KeyResult  `json:"keyResults,omitempty"`
	LinkedTasks []string      `json:"linkedTasks,omitempty"` // task IDs
}

type KeyResult struct {
	ID           string    `json:"id" db:"id"`
	GoalID       string    `json:"goalId" db:"goal_id"`
	Title        string    `json:"title" db:"title"`
	Description  *string   `json:"description,omitempty" db:"description"`
	TargetValue  float64   `json:"targetValue" db:"target_value"`
	CurrentValue float64   `json:"currentValue" db:"current_value"`
	Unit         *string   `json:"unit,omitempty" db:"unit"`
	Status       string    `json:"status" db:"status"` // pending, in_progress, completed, missed
	Weight       float64   `json:"weight" db:"weight"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`

	// Computed
	Progress float64 `json:"progress"` // 0-100
}

type GoalTask struct {
	ID        string    `json:"id" db:"id"`
	GoalID    string    `json:"goalId" db:"goal_id"`
	TaskID    string    `json:"taskId" db:"task_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// ============================================
// GOAL REPOSITORY INTERFACE
// ============================================

type GoalRepository interface {
	// Goal CRUD
	Create(ctx context.Context, goal *Goal) error
	FindByID(ctx context.Context, id string) (*Goal, error)
	FindByWorkspace(ctx context.Context, workspaceID string, goalType *string, status *string) ([]*Goal, error)
	FindByProject(ctx context.Context, projectID string) ([]*Goal, error)
	FindBySprint(ctx context.Context, sprintID string) ([]*Goal, error)
	Update(ctx context.Context, goal *Goal) error
	UpdateProgress(ctx context.Context, id string, currentValue float64) error
	UpdateStatus(ctx context.Context, id string, status string) error
	Delete(ctx context.Context, id string) error

	// Key Results
	CreateKeyResult(ctx context.Context, kr *KeyResult) error
	FindKeyResultsByGoal(ctx context.Context, goalID string) ([]*KeyResult, error)
	UpdateKeyResult(ctx context.Context, kr *KeyResult) error
	UpdateKeyResultProgress(ctx context.Context, id string, currentValue float64) error
	DeleteKeyResult(ctx context.Context, id string) error

	// Goal-Task linking
	LinkTask(ctx context.Context, goalID, taskID string) error
	UnlinkTask(ctx context.Context, goalID, taskID string) error
	GetLinkedTaskIDs(ctx context.Context, goalID string) ([]string, error)
	GetGoalsByTask(ctx context.Context, taskID string) ([]*Goal, error)

	// Analytics
	GetGoalProgress(ctx context.Context, goalID string) (float64, error)
	GetSprintGoalsSummary(ctx context.Context, sprintID string) (completed int, total int, err error)
}

// ============================================
// IMPLEMENTATION
// ============================================

type goalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) GoalRepository {
	return &goalRepository{db: db}
}

// ============================================
// GOAL CRUD
// ============================================

func (r *goalRepository) Create(ctx context.Context, goal *Goal) error {
	query := `
		INSERT INTO goals (
			workspace_id, project_id, sprint_id, title, description,
			goal_type, status, target_value, current_value, unit,
			start_date, target_date, owner_id, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		goal.WorkspaceID, goal.ProjectID, goal.SprintID, goal.Title, goal.Description,
		goal.GoalType, goal.Status, goal.TargetValue, goal.CurrentValue, goal.Unit,
		goal.StartDate, goal.TargetDate, goal.OwnerID, goal.CreatedBy,
	).Scan(&goal.ID, &goal.CreatedAt, &goal.UpdatedAt)
}

func (r *goalRepository) FindByID(ctx context.Context, id string) (*Goal, error) {
	query := `
		SELECT id, workspace_id, project_id, sprint_id, title, description,
			   goal_type, status, target_value, current_value, unit,
			   start_date, target_date, completed_at, owner_id, created_by,
			   created_at, updated_at
		FROM goals WHERE id = $1`

	goal := &Goal{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&goal.ID, &goal.WorkspaceID, &goal.ProjectID, &goal.SprintID,
		&goal.Title, &goal.Description, &goal.GoalType, &goal.Status,
		&goal.TargetValue, &goal.CurrentValue, &goal.Unit,
		&goal.StartDate, &goal.TargetDate, &goal.CompletedAt,
		&goal.OwnerID, &goal.CreatedBy, &goal.CreatedAt, &goal.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Calculate progress
	goal.Progress = r.calculateProgress(goal.CurrentValue, goal.TargetValue)

	// Load key results
	goal.KeyResults, _ = r.FindKeyResultsByGoal(ctx, id)

	// Load linked task IDs
	goal.LinkedTasks, _ = r.GetLinkedTaskIDs(ctx, id)

	return goal, nil
}

func (r *goalRepository) FindByWorkspace(ctx context.Context, workspaceID string, goalType *string, status *string) ([]*Goal, error) {
	query := `
		SELECT id, workspace_id, project_id, sprint_id, title, description,
			   goal_type, status, target_value, current_value, unit,
			   start_date, target_date, completed_at, owner_id, created_by,
			   created_at, updated_at
		FROM goals 
		WHERE workspace_id = $1`
	
	args := []interface{}{workspaceID}
	argIdx := 2

	if goalType != nil {
		query += ` AND goal_type = $` + string(rune('0'+argIdx))
		args = append(args, *goalType)
		argIdx++
	}

	if status != nil {
		query += ` AND status = $` + string(rune('0'+argIdx))
		args = append(args, *status)
	}

	query += ` ORDER BY created_at DESC`

	return r.queryGoals(ctx, query, args...)
}

func (r *goalRepository) FindByProject(ctx context.Context, projectID string) ([]*Goal, error) {
	query := `
		SELECT id, workspace_id, project_id, sprint_id, title, description,
			   goal_type, status, target_value, current_value, unit,
			   start_date, target_date, completed_at, owner_id, created_by,
			   created_at, updated_at
		FROM goals 
		WHERE project_id = $1
		ORDER BY created_at DESC`

	return r.queryGoals(ctx, query, projectID)
}

func (r *goalRepository) FindBySprint(ctx context.Context, sprintID string) ([]*Goal, error) {
	query := `
		SELECT id, workspace_id, project_id, sprint_id, title, description,
			   goal_type, status, target_value, current_value, unit,
			   start_date, target_date, completed_at, owner_id, created_by,
			   created_at, updated_at
		FROM goals 
		WHERE sprint_id = $1
		ORDER BY created_at DESC`

	return r.queryGoals(ctx, query, sprintID)
}

func (r *goalRepository) Update(ctx context.Context, goal *Goal) error {
	query := `
		UPDATE goals SET
			title = $2, description = $3, goal_type = $4, status = $5,
			target_value = $6, current_value = $7, unit = $8,
			start_date = $9, target_date = $10, completed_at = $11,
			owner_id = $12, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		goal.ID, goal.Title, goal.Description, goal.GoalType, goal.Status,
		goal.TargetValue, goal.CurrentValue, goal.Unit,
		goal.StartDate, goal.TargetDate, goal.CompletedAt, goal.OwnerID,
	).Scan(&goal.UpdatedAt)
}

func (r *goalRepository) UpdateProgress(ctx context.Context, id string, currentValue float64) error {
	query := `UPDATE goals SET current_value = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, currentValue)
	return err
}

func (r *goalRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE goals SET 
			status = $2, 
			completed_at = CASE WHEN $2 = 'completed' THEN NOW() ELSE completed_at END,
			updated_at = NOW() 
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

func (r *goalRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM goals WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ============================================
// KEY RESULTS
// ============================================

func (r *goalRepository) CreateKeyResult(ctx context.Context, kr *KeyResult) error {
	query := `
		INSERT INTO goal_key_results (goal_id, title, description, target_value, current_value, unit, status, weight)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		kr.GoalID, kr.Title, kr.Description, kr.TargetValue,
		kr.CurrentValue, kr.Unit, kr.Status, kr.Weight,
	).Scan(&kr.ID, &kr.CreatedAt, &kr.UpdatedAt)
}

func (r *goalRepository) FindKeyResultsByGoal(ctx context.Context, goalID string) ([]*KeyResult, error) {
	query := `
		SELECT id, goal_id, title, description, target_value, current_value,
			   unit, status, weight, created_at, updated_at
		FROM goal_key_results
		WHERE goal_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*KeyResult
	for rows.Next() {
		kr := &KeyResult{}
		err := rows.Scan(
			&kr.ID, &kr.GoalID, &kr.Title, &kr.Description,
			&kr.TargetValue, &kr.CurrentValue, &kr.Unit, &kr.Status,
			&kr.Weight, &kr.CreatedAt, &kr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		kr.Progress = r.calculateProgress(kr.CurrentValue, &kr.TargetValue)
		results = append(results, kr)
	}

	return results, rows.Err()
}

func (r *goalRepository) UpdateKeyResult(ctx context.Context, kr *KeyResult) error {
	query := `
		UPDATE goal_key_results SET
			title = $2, description = $3, target_value = $4, current_value = $5,
			unit = $6, status = $7, weight = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		kr.ID, kr.Title, kr.Description, kr.TargetValue,
		kr.CurrentValue, kr.Unit, kr.Status, kr.Weight,
	).Scan(&kr.UpdatedAt)
}

func (r *goalRepository) UpdateKeyResultProgress(ctx context.Context, id string, currentValue float64) error {
	query := `
		UPDATE goal_key_results SET 
			current_value = $2,
			status = CASE 
				WHEN $2 >= target_value THEN 'completed'
				WHEN $2 > 0 THEN 'in_progress'
				ELSE status 
			END,
			updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, currentValue)
	return err
}

func (r *goalRepository) DeleteKeyResult(ctx context.Context, id string) error {
	query := `DELETE FROM goal_key_results WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ============================================
// GOAL-TASK LINKING
// ============================================

func (r *goalRepository) LinkTask(ctx context.Context, goalID, taskID string) error {
	query := `
		INSERT INTO goal_tasks (goal_id, task_id)
		VALUES ($1, $2)
		ON CONFLICT (goal_id, task_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, goalID, taskID)
	return err
}

func (r *goalRepository) UnlinkTask(ctx context.Context, goalID, taskID string) error {
	query := `DELETE FROM goal_tasks WHERE goal_id = $1 AND task_id = $2`
	_, err := r.db.ExecContext(ctx, query, goalID, taskID)
	return err
}

func (r *goalRepository) GetLinkedTaskIDs(ctx context.Context, goalID string) ([]string, error) {
	query := `SELECT task_id FROM goal_tasks WHERE goal_id = $1`
	rows, err := r.db.QueryContext(ctx, query, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taskIDs []string
	for rows.Next() {
		var taskID string
		if err := rows.Scan(&taskID); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, taskID)
	}
	return taskIDs, rows.Err()
}

func (r *goalRepository) GetGoalsByTask(ctx context.Context, taskID string) ([]*Goal, error) {
	query := `
		SELECT g.id, g.workspace_id, g.project_id, g.sprint_id, g.title, g.description,
			   g.goal_type, g.status, g.target_value, g.current_value, g.unit,
			   g.start_date, g.target_date, g.completed_at, g.owner_id, g.created_by,
			   g.created_at, g.updated_at
		FROM goals g
		JOIN goal_tasks gt ON g.id = gt.goal_id
		WHERE gt.task_id = $1`

	return r.queryGoals(ctx, query, taskID)
}

// ============================================
// ANALYTICS
// ============================================

func (r *goalRepository) GetGoalProgress(ctx context.Context, goalID string) (float64, error) {
	// If goal has key results, calculate weighted average
	query := `
		SELECT 
			CASE 
				WHEN COUNT(*) = 0 THEN 
					(SELECT COALESCE(current_value / NULLIF(target_value, 0) * 100, 0) FROM goals WHERE id = $1)
				ELSE
					SUM((current_value / NULLIF(target_value, 0)) * weight) / NULLIF(SUM(weight), 0) * 100
			END as progress
		FROM goal_key_results
		WHERE goal_id = $1`

	var progress float64
	err := r.db.QueryRowContext(ctx, query, goalID).Scan(&progress)
	return progress, err
}

func (r *goalRepository) GetSprintGoalsSummary(ctx context.Context, sprintID string) (completed int, total int, err error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) as total
		FROM goals
		WHERE sprint_id = $1`

	err = r.db.QueryRowContext(ctx, query, sprintID).Scan(&completed, &total)
	return
}

// ============================================
// HELPERS
// ============================================

func (r *goalRepository) queryGoals(ctx context.Context, query string, args ...interface{}) ([]*Goal, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []*Goal
	for rows.Next() {
		goal := &Goal{}
		err := rows.Scan(
			&goal.ID, &goal.WorkspaceID, &goal.ProjectID, &goal.SprintID,
			&goal.Title, &goal.Description, &goal.GoalType, &goal.Status,
			&goal.TargetValue, &goal.CurrentValue, &goal.Unit,
			&goal.StartDate, &goal.TargetDate, &goal.CompletedAt,
			&goal.OwnerID, &goal.CreatedBy, &goal.CreatedAt, &goal.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		goal.Progress = r.calculateProgress(goal.CurrentValue, goal.TargetValue)
		goals = append(goals, goal)
	}

	return goals, rows.Err()
}

func (r *goalRepository) calculateProgress(current float64, target *float64) float64 {
	if target == nil || *target == 0 {
		return 0
	}
	progress := (current / *target) * 100
	if progress > 100 {
		progress = 100
	}
	return progress
}

// Ensure pq is used
var _ = pq.Array