package repository

import (
	"context"
	"database/sql"
	"time"
)

// TaskActivity model - tracks all changes and activities on a task
type TaskActivity struct {
	ID        string    `json:"id" db:"id"`
	TaskID    string    `json:"taskId" db:"task_id"`
	UserID    *string   `json:"userId,omitempty" db:"user_id"`  // Nullable for system actions
	Action    string    `json:"action" db:"action"`              // created, updated, status_changed, assigned, etc.
	FieldName *string   `json:"fieldName,omitempty" db:"field_name"`
	OldValue  *string   `json:"oldValue,omitempty" db:"old_value"`
	NewValue  *string   `json:"newValue,omitempty" db:"new_value"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// TaskActivityRepository interface
type TaskActivityRepository interface {
	Create(ctx context.Context, activity *TaskActivity) error
	FindByID(ctx context.Context, id string) (*TaskActivity, error)
	FindByTaskID(ctx context.Context, taskID string, limit int) ([]*TaskActivity, error)
	FindByUserID(ctx context.Context, userID string, limit int) ([]*TaskActivity, error)
	FindByProjectID(ctx context.Context, projectID string, limit int) ([]*TaskActivity, error)
	Delete(ctx context.Context, id string) error
}

// taskActivityRepository implementation
type taskActivityRepository struct {
	db *sql.DB
}

// NewTaskActivityRepository creates a new TaskActivityRepository
func NewTaskActivityRepository(db *sql.DB) TaskActivityRepository {
	return &taskActivityRepository{db: db}
}

// Create inserts a new activity record
func (r *taskActivityRepository) Create(ctx context.Context, activity *TaskActivity) error {
	query := `
		INSERT INTO task_activities (
			id, task_id, user_id, action, field_name, old_value, new_value, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW()
		) RETURNING id, created_at`

	return r.db.QueryRowContext(
		ctx, query,
		activity.TaskID,
		activity.UserID,
		activity.Action,
		activity.FieldName,
		activity.OldValue,
		activity.NewValue,
	).Scan(&activity.ID, &activity.CreatedAt)
}

// FindByID retrieves an activity by ID
func (r *taskActivityRepository) FindByID(ctx context.Context, id string) (*TaskActivity, error) {
	query := `SELECT * FROM task_activities WHERE id = $1`

	activity := &TaskActivity{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&activity.ID,
		&activity.TaskID,
		&activity.UserID,
		&activity.Action,
		&activity.FieldName,
		&activity.OldValue,
		&activity.NewValue,
		&activity.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return activity, nil
}

// FindByTaskID retrieves all activities for a task
func (r *taskActivityRepository) FindByTaskID(ctx context.Context, taskID string, limit int) ([]*TaskActivity, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}

	query := `SELECT * FROM task_activities WHERE task_id = $1 ORDER BY created_at DESC LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, taskID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*TaskActivity
	for rows.Next() {
		activity := &TaskActivity{}
		err := rows.Scan(
			&activity.ID,
			&activity.TaskID,
			&activity.UserID,
			&activity.Action,
			&activity.FieldName,
			&activity.OldValue,
			&activity.NewValue,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, rows.Err()
}

// FindByUserID retrieves all activities by a user
func (r *taskActivityRepository) FindByUserID(ctx context.Context, userID string, limit int) ([]*TaskActivity, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `SELECT * FROM task_activities WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*TaskActivity
	for rows.Next() {
		activity := &TaskActivity{}
		err := rows.Scan(
			&activity.ID,
			&activity.TaskID,
			&activity.UserID,
			&activity.Action,
			&activity.FieldName,
			&activity.OldValue,
			&activity.NewValue,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, rows.Err()
}

// FindByProjectID retrieves all activities for tasks in a project
func (r *taskActivityRepository) FindByProjectID(ctx context.Context, projectID string, limit int) ([]*TaskActivity, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT ta.* 
		FROM task_activities ta
		JOIN tasks t ON ta.task_id = t.id
		WHERE t.project_id = $1
		ORDER BY ta.created_at DESC 
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*TaskActivity
	for rows.Next() {
		activity := &TaskActivity{}
		err := rows.Scan(
			&activity.ID,
			&activity.TaskID,
			&activity.UserID,
			&activity.Action,
			&activity.FieldName,
			&activity.OldValue,
			&activity.NewValue,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, rows.Err()
}

// Delete removes an activity record
func (r *taskActivityRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM task_activities WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}