package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Attachment struct {
	ID         string
	Name       string
	URL        string
	Size       int64
	MimeType   *string
	EntityType string // "task", "comment", "project"
	EntityID   string
	UploadedBy string
	CreatedAt  time.Time
	Uploader   *User // populated on join
}

type AttachmentRepository interface {
	Create(ctx context.Context, attachment *Attachment) error
	FindByID(ctx context.Context, id string) (*Attachment, error)
	FindByEntity(ctx context.Context, entityType, entityID string) ([]*Attachment, error)
	Delete(ctx context.Context, id string) error
	DeleteByEntity(ctx context.Context, entityType, entityID string) error
}

type pgAttachmentRepository struct {
	pool *pgxpool.Pool
}

func NewAttachmentRepository(pool *pgxpool.Pool) AttachmentRepository {
	return &pgAttachmentRepository{pool: pool}
}

func (r *pgAttachmentRepository) Create(ctx context.Context, attachment *Attachment) error {
	query := `
		INSERT INTO attachments (name, url, size, mime_type, entity_type, entity_id, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		attachment.Name, attachment.URL, attachment.Size, attachment.MimeType,
		attachment.EntityType, attachment.EntityID, attachment.UploadedBy,
	).Scan(&attachment.ID, &attachment.CreatedAt)
}

func (r *pgAttachmentRepository) FindByID(ctx context.Context, id string) (*Attachment, error) {
	query := `
		SELECT a.id, a.name, a.url, a.size, a.mime_type, a.entity_type, a.entity_id, 
		       a.uploaded_by, a.created_at,
		       u.id, u.email, u.name, u.avatar, u.status
		FROM attachments a
		LEFT JOIN users u ON a.uploaded_by = u.id
		WHERE a.id = $1
	`
	att := &Attachment{Uploader: &User{}}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&att.ID, &att.Name, &att.URL, &att.Size, &att.MimeType,
		&att.EntityType, &att.EntityID, &att.UploadedBy, &att.CreatedAt,
		&att.Uploader.ID, &att.Uploader.Email, &att.Uploader.Name,
		&att.Uploader.Avatar, &att.Uploader.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return att, nil
}

func (r *pgAttachmentRepository) FindByEntity(ctx context.Context, entityType, entityID string) ([]*Attachment, error) {
	query := `
		SELECT a.id, a.name, a.url, a.size, a.mime_type, a.entity_type, a.entity_id, 
		       a.uploaded_by, a.created_at,
		       u.id, u.email, u.name, u.avatar, u.status
		FROM attachments a
		LEFT JOIN users u ON a.uploaded_by = u.id
		WHERE a.entity_type = $1 AND a.entity_id = $2
		ORDER BY a.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*Attachment
	for rows.Next() {
		att := &Attachment{Uploader: &User{}}
		if err := rows.Scan(
			&att.ID, &att.Name, &att.URL, &att.Size, &att.MimeType,
			&att.EntityType, &att.EntityID, &att.UploadedBy, &att.CreatedAt,
			&att.Uploader.ID, &att.Uploader.Email, &att.Uploader.Name,
			&att.Uploader.Avatar, &att.Uploader.Status,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, att)
	}
	return attachments, nil
}

func (r *pgAttachmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM attachments WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgAttachmentRepository) DeleteByEntity(ctx context.Context, entityType, entityID string) error {
	query := `DELETE FROM attachments WHERE entity_type = $1 AND entity_id = $2`
	_, err := r.pool.Exec(ctx, query, entityType, entityID)
	return err
}