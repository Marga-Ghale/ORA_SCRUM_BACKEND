package repository

import (
	"context"
	"database/sql"
	"time"
)

// ============================================
// ATTACHMENTS
// ============================================

type TaskAttachment struct {
	ID         string    `json:"id" db:"id"`
	TaskID     string    `json:"taskId" db:"task_id"`
	UserID     string    `json:"userId" db:"user_id"`
	Filename   string    `json:"filename" db:"filename"`
	FileURL    string    `json:"fileUrl" db:"file_url"`
	FileSize   int64     `json:"fileSize" db:"file_size"`
	MimeType   string    `json:"mimeType" db:"mime_type"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	User       *User     `json:"user,omitempty"` // populated via join
}

type TaskAttachmentRepository interface {
	Create(ctx context.Context, attachment *TaskAttachment) error
	FindByTaskID(ctx context.Context, taskID string) ([]*TaskAttachment, error)
	FindByID(ctx context.Context, id string) (*TaskAttachment, error)
	Delete(ctx context.Context, id string) error
}

// ============================================
// TASK ATTACHMENT REPOSITORY IMPLEMENTATION
// ============================================

type taskAttachmentRepository struct {
	db *sql.DB
}

func NewTaskAttachmentRepository(db *sql.DB) TaskAttachmentRepository {
	return &taskAttachmentRepository{db: db}
}

func (r *taskAttachmentRepository) Create(ctx context.Context, attachment *TaskAttachment) error {
	query := `
		INSERT INTO task_attachments (id, task_id, user_id, filename, file_url, file_size, mime_type, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`
	
	return r.db.QueryRowContext(ctx, query,
		attachment.TaskID, attachment.UserID, attachment.Filename,
		attachment.FileURL, attachment.FileSize, attachment.MimeType,
	).Scan(&attachment.ID, &attachment.CreatedAt)
}

func (r *taskAttachmentRepository) FindByTaskID(ctx context.Context, taskID string) ([]*TaskAttachment, error) {
	query := `
		SELECT a.id, a.task_id, a.user_id, a.filename, a.file_url, a.file_size, a.mime_type, a.created_at
		FROM task_attachments a
		WHERE a.task_id = $1
		ORDER BY a.created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*TaskAttachment
	for rows.Next() {
		a := &TaskAttachment{}
		err := rows.Scan(
			&a.ID, &a.TaskID, &a.UserID, &a.Filename,
			&a.FileURL, &a.FileSize, &a.MimeType, &a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	return attachments, rows.Err()
}

func (r *taskAttachmentRepository) FindByID(ctx context.Context, id string) (*TaskAttachment, error) {
	query := `
		SELECT id, task_id, user_id, filename, file_url, file_size, mime_type, created_at
		FROM task_attachments
		WHERE id = $1`
	
	a := &TaskAttachment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.TaskID, &a.UserID, &a.Filename,
		&a.FileURL, &a.FileSize, &a.MimeType, &a.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *taskAttachmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM task_attachments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}