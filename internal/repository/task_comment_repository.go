package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// TaskComment model
type TaskComment struct {
	ID             string    `json:"id" db:"id"`
	TaskID         string    `json:"taskId" db:"task_id"`
	UserID         string    `json:"userId" db:"user_id"`
	Content        string    `json:"content" db:"content"`
	MentionedUsers []string  `json:"mentionedUsers" db:"mentioned_users"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

// TaskCommentRepository interface
type TaskCommentRepository interface {
	Create(ctx context.Context, comment *TaskComment) error
	FindByID(ctx context.Context, id string) (*TaskComment, error)
	FindByTaskID(ctx context.Context, taskID string) ([]*TaskComment, error)
	Update(ctx context.Context, comment *TaskComment) error
	Delete(ctx context.Context, id string) error
}

// taskCommentRepository implementation
type taskCommentRepository struct {
	db *sql.DB
}

// NewTaskCommentRepository creates a new TaskCommentRepository
func NewTaskCommentRepository(db *sql.DB) TaskCommentRepository {
	return &taskCommentRepository{db: db}
}

// Create inserts a new comment
func (r *taskCommentRepository) Create(ctx context.Context, comment *TaskComment) error {
	query := `
		INSERT INTO task_comments (
			id, task_id, user_id, content, mentioned_users, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW()
		) RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		comment.TaskID,
		comment.UserID,
		comment.Content,
		pq.Array(comment.MentionedUsers),
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
}

// FindByID retrieves a comment by ID
func (r *taskCommentRepository) FindByID(ctx context.Context, id string) (*TaskComment, error) {
	query := `SELECT * FROM task_comments WHERE id = $1`

	comment := &TaskComment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.TaskID,
		&comment.UserID,
		&comment.Content,
		pq.Array(&comment.MentionedUsers),
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// FindByTaskID retrieves all comments for a task
func (r *taskCommentRepository) FindByTaskID(ctx context.Context, taskID string) ([]*TaskComment, error) {
	query := `SELECT * FROM task_comments WHERE task_id = $1 ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*TaskComment
	for rows.Next() {
		comment := &TaskComment{}
		err := rows.Scan(
			&comment.ID,
			&comment.TaskID,
			&comment.UserID,
			&comment.Content,
			pq.Array(&comment.MentionedUsers),
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

// Update updates an existing comment
func (r *taskCommentRepository) Update(ctx context.Context, comment *TaskComment) error {
	query := `
		UPDATE task_comments SET
			content = $2,
			mentioned_users = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		comment.ID,
		comment.Content,
		pq.Array(comment.MentionedUsers),
	).Scan(&comment.UpdatedAt)
}

// Delete removes a comment
func (r *taskCommentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM task_comments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}