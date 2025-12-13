package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        string
	UserID    string
	Type      string
	Title     string
	Message   string
	Read      bool
	Data      map[string]interface{}
	CreatedAt time.Time
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	FindByID(ctx context.Context, id string) (*Notification, error)
	FindByUserID(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error)
	CountByUserID(ctx context.Context, userID string) (total int, unread int, err error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
	DeleteAll(ctx context.Context, userID string) error
	DeleteOlderThan(ctx context.Context, olderThan time.Time, readOnly bool) (int, error)
}

type pgNotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) NotificationRepository {
	return &pgNotificationRepository{pool: pool}
}

func (r *pgNotificationRepository) Create(ctx context.Context, notification *Notification) error {
	dataJSON, _ := json.Marshal(notification.Data)
	if notification.Data == nil {
		dataJSON = []byte("{}")
	}
	query := `
		INSERT INTO notifications (user_id, type, title, message, read, data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		notification.UserID, notification.Type, notification.Title,
		notification.Message, notification.Read, dataJSON,
	).Scan(&notification.ID, &notification.CreatedAt)
}

func (r *pgNotificationRepository) FindByID(ctx context.Context, id string) (*Notification, error) {
	query := `SELECT id, user_id, type, title, message, read, data, created_at FROM notifications WHERE id = $1`
	n := &Notification{}
	var dataJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Read, &dataJSON, &n.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(dataJSON, &n.Data)
	return n, nil
}

func (r *pgNotificationRepository) FindByUserID(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, read, data, created_at 
		FROM notifications WHERE user_id = $1
	`
	if unreadOnly {
		query += ` AND read = FALSE`
	}
	query += ` ORDER BY created_at DESC LIMIT 100`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		n := &Notification{}
		var dataJSON []byte
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Read, &dataJSON, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(dataJSON, &n.Data)
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *pgNotificationRepository) CountByUserID(ctx context.Context, userID string) (total int, unread int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE read = FALSE) as unread
		FROM notifications WHERE user_id = $1
	`
	err = r.pool.QueryRow(ctx, query, userID).Scan(&total, &unread)
	return
}

func (r *pgNotificationRepository) MarkAsRead(ctx context.Context, id string) error {
	query := `UPDATE notifications SET read = TRUE WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgNotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notifications SET read = TRUE WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *pgNotificationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgNotificationRepository) DeleteAll(ctx context.Context, userID string) error {
	query := `DELETE FROM notifications WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *pgNotificationRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time, readOnly bool) (int, error) {
	query := `DELETE FROM notifications WHERE created_at < $1`
	if readOnly {
		query += ` AND read = TRUE`
	}
	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}
