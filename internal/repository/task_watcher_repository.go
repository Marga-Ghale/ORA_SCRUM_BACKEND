package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskWatcher struct {
	ID        string
	TaskID    string
	UserID    string
	CreatedAt time.Time
}

type TaskWatcherRepository interface {
	Add(ctx context.Context, watcher *TaskWatcher) error
	Remove(ctx context.Context, taskID, userID string) error
	FindByTask(ctx context.Context, taskID string) ([]*TaskWatcher, error)
	FindByUser(ctx context.Context, userID string) ([]*TaskWatcher, error)
	IsWatching(ctx context.Context, taskID, userID string) (bool, error)
	GetWatcherUserIDs(ctx context.Context, taskID string) ([]string, error)
}

type pgTaskWatcherRepository struct {
	pool *pgxpool.Pool
}

func NewTaskWatcherRepository(pool *pgxpool.Pool) TaskWatcherRepository {
	return &pgTaskWatcherRepository{pool: pool}
}

func (r *pgTaskWatcherRepository) Add(ctx context.Context, watcher *TaskWatcher) error {
	query := `
		INSERT INTO task_watchers (task_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (task_id, user_id) DO NOTHING
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, watcher.TaskID, watcher.UserID).
		Scan(&watcher.ID, &watcher.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil
	}
	return err
}

func (r *pgTaskWatcherRepository) Remove(ctx context.Context, taskID, userID string) error {
	query := `DELETE FROM task_watchers WHERE task_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, taskID, userID)
	return err
}

func (r *pgTaskWatcherRepository) FindByTask(ctx context.Context, taskID string) ([]*TaskWatcher, error) {
	query := `
		SELECT id, task_id, user_id, created_at
		FROM task_watchers WHERE task_id = $1
	`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchers []*TaskWatcher
	for rows.Next() {
		watcher := &TaskWatcher{}
		if err := rows.Scan(&watcher.ID, &watcher.TaskID, &watcher.UserID, &watcher.CreatedAt); err != nil {
			return nil, err
		}
		watchers = append(watchers, watcher)
	}
	return watchers, nil
}

func (r *pgTaskWatcherRepository) FindByUser(ctx context.Context, userID string) ([]*TaskWatcher, error) {
	query := `
		SELECT id, task_id, user_id, created_at
		FROM task_watchers WHERE user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchers []*TaskWatcher
	for rows.Next() {
		watcher := &TaskWatcher{}
		if err := rows.Scan(&watcher.ID, &watcher.TaskID, &watcher.UserID, &watcher.CreatedAt); err != nil {
			return nil, err
		}
		watchers = append(watchers, watcher)
	}
	return watchers, nil
}

func (r *pgTaskWatcherRepository) IsWatching(ctx context.Context, taskID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM task_watchers WHERE task_id = $1 AND user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, taskID, userID).Scan(&exists)
	return exists, err
}

func (r *pgTaskWatcherRepository) GetWatcherUserIDs(ctx context.Context, taskID string) ([]string, error) {
	query := `SELECT user_id FROM task_watchers WHERE task_id = $1`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}
