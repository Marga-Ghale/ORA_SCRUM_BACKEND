package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Comment struct {
	ID        string
	Content   string
	TaskID    string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	User      *User
}

type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) error
	FindByID(ctx context.Context, id string) (*Comment, error)
	FindByTaskID(ctx context.Context, taskID string) ([]*Comment, error)
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id string) error
}

type pgCommentRepository struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) CommentRepository {
	return &pgCommentRepository{pool: pool}
}

func (r *pgCommentRepository) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (content, task_id, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query, comment.Content, comment.TaskID, comment.UserID).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
}

func (r *pgCommentRepository) FindByID(ctx context.Context, id string) (*Comment, error) {
	query := `
		SELECT c.id, c.content, c.task_id, c.user_id, c.created_at, c.updated_at,
		       u.id, u.name, u.email, u.avatar
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`
	c := &Comment{User: &User{}}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Content, &c.TaskID, &c.UserID, &c.CreatedAt, &c.UpdatedAt,
		&c.User.ID, &c.User.Name, &c.User.Email, &c.User.Avatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *pgCommentRepository) FindByTaskID(ctx context.Context, taskID string) ([]*Comment, error) {
	query := `
		SELECT c.id, c.content, c.task_id, c.user_id, c.created_at, c.updated_at,
		       u.id, u.name, u.email, u.avatar
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.task_id = $1
		ORDER BY c.created_at
	`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{User: &User{}}
		if err := rows.Scan(
			&c.ID, &c.Content, &c.TaskID, &c.UserID, &c.CreatedAt, &c.UpdatedAt,
			&c.User.ID, &c.User.Name, &c.User.Email, &c.User.Avatar,
		); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *pgCommentRepository) Update(ctx context.Context, comment *Comment) error {
	query := `UPDATE comments SET content = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, comment.ID, comment.Content)
	return err
}

func (r *pgCommentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
