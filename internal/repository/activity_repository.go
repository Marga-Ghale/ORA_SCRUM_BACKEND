package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Activity struct {
	ID         string
	Type       string
	EntityType string
	EntityID   string
	UserID     string
	Changes    map[string]interface{}
	Metadata   map[string]interface{}
	CreatedAt  time.Time
}

type ActivityRepository interface {
	Create(ctx context.Context, activity *Activity) error
	FindByEntity(ctx context.Context, entityType, entityID string, limit int) ([]*Activity, error)
	FindByUser(ctx context.Context, userID string, limit int) ([]*Activity, error)
	DeleteOlderThan(ctx context.Context, olderThan time.Time) (int, error)
}

type pgActivityRepository struct {
	pool *pgxpool.Pool
}

func NewActivityRepository(pool *pgxpool.Pool) ActivityRepository {
	return &pgActivityRepository{pool: pool}
}

func (r *pgActivityRepository) Create(ctx context.Context, activity *Activity) error {
	query := `
		INSERT INTO activities (type, entity_type, entity_id, user_id, changes, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		activity.Type, activity.EntityType, activity.EntityID,
		activity.UserID, activity.Changes, activity.Metadata,
	).Scan(&activity.ID, &activity.CreatedAt)
}

func (r *pgActivityRepository) FindByEntity(ctx context.Context, entityType, entityID string, limit int) ([]*Activity, error) {
	query := `
		SELECT id, type, entity_type, entity_id, user_id, changes, metadata, created_at
		FROM activities WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`
	rows, err := r.pool.Query(ctx, query, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*Activity
	for rows.Next() {
		activity := &Activity{}
		if err := rows.Scan(
			&activity.ID, &activity.Type, &activity.EntityType, &activity.EntityID,
			&activity.UserID, &activity.Changes, &activity.Metadata, &activity.CreatedAt,
		); err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, nil
}

func (r *pgActivityRepository) FindByUser(ctx context.Context, userID string, limit int) ([]*Activity, error) {
	query := `
		SELECT id, type, entity_type, entity_id, user_id, changes, metadata, created_at
		FROM activities WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*Activity
	for rows.Next() {
		activity := &Activity{}
		if err := rows.Scan(
			&activity.ID, &activity.Type, &activity.EntityType, &activity.EntityID,
			&activity.UserID, &activity.Changes, &activity.Metadata, &activity.CreatedAt,
		); err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, nil
}

func (r *pgActivityRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int, error) {
	query := `DELETE FROM activities WHERE created_at < $1`
	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}
