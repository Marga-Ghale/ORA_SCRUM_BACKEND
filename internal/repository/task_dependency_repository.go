package repository

import (
	"context"
	"database/sql"
	"time"
)

type TaskDependency struct {
	ID              string    `json:"id" db:"id"`
	TaskID          string    `json:"taskId" db:"task_id"`
	DependsOnTaskID string    `json:"dependsOnTaskId" db:"depends_on_task_id"`
	DependencyType  string    `json:"dependencyType" db:"dependency_type"` // blocks, blocked_by, relates_to
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

type TaskDependencyRepository interface {
	Create(ctx context.Context, dep *TaskDependency) error
	FindByTaskID(ctx context.Context, taskID string) ([]*TaskDependency, error)
	FindBlockedBy(ctx context.Context, taskID string) ([]*TaskDependency, error)
	Delete(ctx context.Context, taskID, dependsOnTaskID string) error
	DeleteByID(ctx context.Context, id string) error
}

type taskDependencyRepository struct {
	db *sql.DB
}

func NewTaskDependencyRepository(db *sql.DB) TaskDependencyRepository {
	return &taskDependencyRepository{db: db}
}

func (r *taskDependencyRepository) Create(ctx context.Context, dep *TaskDependency) error {
	query := `
		INSERT INTO task_dependencies (id, task_id, depends_on_task_id, dependency_type, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		RETURNING id, created_at`
	
	return r.db.QueryRowContext(ctx, query,
		dep.TaskID, dep.DependsOnTaskID, dep.DependencyType,
	).Scan(&dep.ID, &dep.CreatedAt)
}

func (r *taskDependencyRepository) FindByTaskID(ctx context.Context, taskID string) ([]*TaskDependency, error) {
	query := `SELECT * FROM task_dependencies WHERE task_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []*TaskDependency
	for rows.Next() {
		dep := &TaskDependency{}
		if err := rows.Scan(&dep.ID, &dep.TaskID, &dep.DependsOnTaskID, &dep.DependencyType, &dep.CreatedAt); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, rows.Err()
}

func (r *taskDependencyRepository) FindBlockedBy(ctx context.Context, taskID string) ([]*TaskDependency, error) {
	query := `SELECT * FROM task_dependencies WHERE depends_on_task_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []*TaskDependency
	for rows.Next() {
		dep := &TaskDependency{}
		if err := rows.Scan(&dep.ID, &dep.TaskID, &dep.DependsOnTaskID, &dep.DependencyType, &dep.CreatedAt); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, rows.Err()
}

func (r *taskDependencyRepository) Delete(ctx context.Context, taskID, dependsOnTaskID string) error {
	query := `DELETE FROM task_dependencies WHERE task_id = $1 AND depends_on_task_id = $2`
	_, err := r.db.ExecContext(ctx, query, taskID, dependsOnTaskID)
	return err
}

func (r *taskDependencyRepository) DeleteByID(ctx context.Context, id string) error {
	query := `DELETE FROM task_dependencies WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}