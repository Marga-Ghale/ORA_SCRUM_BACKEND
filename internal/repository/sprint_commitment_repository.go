package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// ============================================
// MODELS
// ============================================

type SprintCommitment struct {
	ID              string    `json:"id" db:"id"`
	SprintID        string    `json:"sprintId" db:"sprint_id"`
	CommittedTasks  int       `json:"committedTasks" db:"committed_tasks"`
	CommittedPoints int       `json:"committedPoints" db:"committed_points"`
	TaskIDs         []string  `json:"taskIds" db:"task_ids"`
	SnapshotAt      time.Time `json:"snapshotAt" db:"snapshot_at"`
}

type SprintScopeChange struct {
	ID          string    `json:"id" db:"id"`
	SprintID    string    `json:"sprintId" db:"sprint_id"`
	TaskID      string    `json:"taskId" db:"task_id"`
	ChangeType  string    `json:"changeType" db:"change_type"` // added, removed
	StoryPoints int       `json:"storyPoints" db:"story_points"`
	ChangedBy   *string   `json:"changedBy,omitempty" db:"changed_by"`
	ChangedAt   time.Time `json:"changedAt" db:"changed_at"`
}

// ============================================
// INTERFACE
// ============================================

type SprintCommitmentRepository interface {
	// Commitment snapshot
	SaveCommitment(ctx context.Context, commitment *SprintCommitment) error
	GetCommitment(ctx context.Context, sprintID string) (*SprintCommitment, error)
	
	// Scope changes
	RecordScopeChange(ctx context.Context, change *SprintScopeChange) error
	GetScopeChanges(ctx context.Context, sprintID string) ([]*SprintScopeChange, error)
	GetAddedTasksCount(ctx context.Context, sprintID string) (tasks int, points int, err error)
	GetRemovedTasksCount(ctx context.Context, sprintID string) (tasks int, points int, err error)
	
	// Status history
	RecordStatusChange(ctx context.Context, taskID, fromStatus, toStatus string, changedBy *string) error
	GetTaskStatusHistory(ctx context.Context, taskID string) ([]*TaskStatusHistory, error)
}

// ============================================
// IMPLEMENTATION
// ============================================

type sprintCommitmentRepository struct {
	db *sql.DB
}

func NewSprintCommitmentRepository(db *sql.DB) SprintCommitmentRepository {
	return &sprintCommitmentRepository{db: db}
}

func (r *sprintCommitmentRepository) SaveCommitment(ctx context.Context, commitment *SprintCommitment) error {
	query := `
		INSERT INTO sprint_commitments (sprint_id, committed_tasks, committed_points, task_ids, snapshot_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (sprint_id) DO UPDATE SET
			committed_tasks = EXCLUDED.committed_tasks,
			committed_points = EXCLUDED.committed_points,
			task_ids = EXCLUDED.task_ids,
			snapshot_at = NOW()
		RETURNING id, snapshot_at`

	return r.db.QueryRowContext(ctx, query,
		commitment.SprintID,
		commitment.CommittedTasks,
		commitment.CommittedPoints,
		pq.Array(commitment.TaskIDs),
	).Scan(&commitment.ID, &commitment.SnapshotAt)
}

func (r *sprintCommitmentRepository) GetCommitment(ctx context.Context, sprintID string) (*SprintCommitment, error) {
	query := `
		SELECT id, sprint_id, committed_tasks, committed_points, task_ids, snapshot_at
		FROM sprint_commitments
		WHERE sprint_id = $1`

	commitment := &SprintCommitment{}
	err := r.db.QueryRowContext(ctx, query, sprintID).Scan(
		&commitment.ID,
		&commitment.SprintID,
		&commitment.CommittedTasks,
		&commitment.CommittedPoints,
		pq.Array(&commitment.TaskIDs),
		&commitment.SnapshotAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return commitment, err
}

func (r *sprintCommitmentRepository) RecordScopeChange(ctx context.Context, change *SprintScopeChange) error {
	query := `
		INSERT INTO sprint_scope_changes (sprint_id, task_id, change_type, story_points, changed_by, changed_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, changed_at`

	return r.db.QueryRowContext(ctx, query,
		change.SprintID,
		change.TaskID,
		change.ChangeType,
		change.StoryPoints,
		change.ChangedBy,
	).Scan(&change.ID, &change.ChangedAt)
}

func (r *sprintCommitmentRepository) GetScopeChanges(ctx context.Context, sprintID string) ([]*SprintScopeChange, error) {
	query := `
		SELECT id, sprint_id, task_id, change_type, story_points, changed_by, changed_at
		FROM sprint_scope_changes
		WHERE sprint_id = $1
		ORDER BY changed_at ASC`

	rows, err := r.db.QueryContext(ctx, query, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []*SprintScopeChange
	for rows.Next() {
		c := &SprintScopeChange{}
		err := rows.Scan(&c.ID, &c.SprintID, &c.TaskID, &c.ChangeType, &c.StoryPoints, &c.ChangedBy, &c.ChangedAt)
		if err != nil {
			return nil, err
		}
		changes = append(changes, c)
	}
	return changes, rows.Err()
}

func (r *sprintCommitmentRepository) GetAddedTasksCount(ctx context.Context, sprintID string) (tasks int, points int, err error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(story_points), 0)
		FROM sprint_scope_changes
		WHERE sprint_id = $1 AND change_type = 'added'`

	err = r.db.QueryRowContext(ctx, query, sprintID).Scan(&tasks, &points)
	return
}

func (r *sprintCommitmentRepository) GetRemovedTasksCount(ctx context.Context, sprintID string) (tasks int, points int, err error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(story_points), 0)
		FROM sprint_scope_changes
		WHERE sprint_id = $1 AND change_type = 'removed'`

	err = r.db.QueryRowContext(ctx, query, sprintID).Scan(&tasks, &points)
	return
}

func (r *sprintCommitmentRepository) RecordStatusChange(ctx context.Context, taskID, fromStatus, toStatus string, changedBy *string) error {
	query := `
		INSERT INTO task_status_history (task_id, from_status, to_status, changed_by, changed_at)
		VALUES ($1, $2, $3, $4, NOW())`

	_, err := r.db.ExecContext(ctx, query, taskID, fromStatus, toStatus, changedBy)
	return err
}

func (r *sprintCommitmentRepository) GetTaskStatusHistory(ctx context.Context, taskID string) ([]*TaskStatusHistory, error) {
	query := `
		SELECT id, task_id, from_status, to_status, changed_by, changed_at
		FROM task_status_history
		WHERE task_id = $1
		ORDER BY changed_at ASC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*TaskStatusHistory
	for rows.Next() {
		h := &TaskStatusHistory{}
		err := rows.Scan(&h.ID, &h.TaskID, &h.FromStatus, &h.ToStatus, &h.ChangedBy, &h.ChangedAt)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}