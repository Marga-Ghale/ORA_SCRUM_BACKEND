package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TimeTracking struct {
	ID              string
	TaskID          string
	UserID          string
	StartTime       time.Time
	EndTime         *time.Time
	DurationSeconds *int
	Description     *string
	IsManual        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	User            *User // populated on join
	Task            *Task // populated on join
}

type TimeTrackingRepository interface {
	Create(ctx context.Context, timeEntry *TimeTracking) error
	FindByID(ctx context.Context, id string) (*TimeTracking, error)
	FindByTask(ctx context.Context, taskID string) ([]*TimeTracking, error)
	FindByUser(ctx context.Context, userID string, startDate, endDate *time.Time) ([]*TimeTracking, error)
	FindActiveByUser(ctx context.Context, userID string) (*TimeTracking, error)
	Update(ctx context.Context, timeEntry *TimeTracking) error
	StopTimer(ctx context.Context, id string, endTime time.Time) error
	Delete(ctx context.Context, id string) error
	GetTotalTimeByTask(ctx context.Context, taskID string) (int, error)
	GetTotalTimeByUser(ctx context.Context, userID string, startDate, endDate *time.Time) (int, error)
	GetTotalTimeBySprint(ctx context.Context, sprintID string) (int, error)
}

type pgTimeTrackingRepository struct {
	pool *pgxpool.Pool
}

func NewTimeTrackingRepository(pool *pgxpool.Pool) TimeTrackingRepository {
	return &pgTimeTrackingRepository{pool: pool}
}

func (r *pgTimeTrackingRepository) Create(ctx context.Context, timeEntry *TimeTracking) error {
	query := `
		INSERT INTO time_tracking (task_id, user_id, start_time, end_time, duration_seconds, description, is_manual)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		timeEntry.TaskID, timeEntry.UserID, timeEntry.StartTime, timeEntry.EndTime,
		timeEntry.DurationSeconds, timeEntry.Description, timeEntry.IsManual,
	).Scan(&timeEntry.ID, &timeEntry.CreatedAt, &timeEntry.UpdatedAt)
}

func (r *pgTimeTrackingRepository) FindByID(ctx context.Context, id string) (*TimeTracking, error) {
	query := `
		SELECT tt.id, tt.task_id, tt.user_id, tt.start_time, tt.end_time, 
		       tt.duration_seconds, tt.description, tt.is_manual, tt.created_at, tt.updated_at,
		       u.id, u.email, u.name, u.avatar, u.status
		FROM time_tracking tt
		LEFT JOIN users u ON tt.user_id = u.id
		WHERE tt.id = $1
	`
	te := &TimeTracking{User: &User{}}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&te.ID, &te.TaskID, &te.UserID, &te.StartTime, &te.EndTime,
		&te.DurationSeconds, &te.Description, &te.IsManual, &te.CreatedAt, &te.UpdatedAt,
		&te.User.ID, &te.User.Email, &te.User.Name, &te.User.Avatar, &te.User.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return te, nil
}

func (r *pgTimeTrackingRepository) FindByTask(ctx context.Context, taskID string) ([]*TimeTracking, error) {
	query := `
		SELECT tt.id, tt.task_id, tt.user_id, tt.start_time, tt.end_time, 
		       tt.duration_seconds, tt.description, tt.is_manual, tt.created_at, tt.updated_at,
		       u.id, u.email, u.name, u.avatar, u.status
		FROM time_tracking tt
		LEFT JOIN users u ON tt.user_id = u.id
		WHERE tt.task_id = $1
		ORDER BY tt.start_time DESC
	`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*TimeTracking
	for rows.Next() {
		te := &TimeTracking{User: &User{}}
		if err := rows.Scan(
			&te.ID, &te.TaskID, &te.UserID, &te.StartTime, &te.EndTime,
			&te.DurationSeconds, &te.Description, &te.IsManual, &te.CreatedAt, &te.UpdatedAt,
			&te.User.ID, &te.User.Email, &te.User.Name, &te.User.Avatar, &te.User.Status,
		); err != nil {
			return nil, err
		}
		entries = append(entries, te)
	}
	return entries, nil
}

func (r *pgTimeTrackingRepository) FindByUser(ctx context.Context, userID string, startDate, endDate *time.Time) ([]*TimeTracking, error) {
	query := `
		SELECT tt.id, tt.task_id, tt.user_id, tt.start_time, tt.end_time, 
		       tt.duration_seconds, tt.description, tt.is_manual, tt.created_at, tt.updated_at
		FROM time_tracking tt
		WHERE tt.user_id = $1
	`
	args := []interface{}{userID}
	argCount := 1

	if startDate != nil {
		argCount++
		query += ` AND tt.start_time >= $` + string(rune('0'+argCount))
		args = append(args, *startDate)
	}
	if endDate != nil {
		argCount++
		query += ` AND tt.start_time <= $` + string(rune('0'+argCount))
		args = append(args, *endDate)
	}

	query += ` ORDER BY tt.start_time DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*TimeTracking
	for rows.Next() {
		te := &TimeTracking{}
		if err := rows.Scan(
			&te.ID, &te.TaskID, &te.UserID, &te.StartTime, &te.EndTime,
			&te.DurationSeconds, &te.Description, &te.IsManual, &te.CreatedAt, &te.UpdatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, te)
	}
	return entries, nil
}

func (r *pgTimeTrackingRepository) FindActiveByUser(ctx context.Context, userID string) (*TimeTracking, error) {
	query := `
		SELECT tt.id, tt.task_id, tt.user_id, tt.start_time, tt.end_time, 
		       tt.duration_seconds, tt.description, tt.is_manual, tt.created_at, tt.updated_at
		FROM time_tracking tt
		WHERE tt.user_id = $1 AND tt.end_time IS NULL
		ORDER BY tt.start_time DESC
		LIMIT 1
	`
	te := &TimeTracking{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&te.ID, &te.TaskID, &te.UserID, &te.StartTime, &te.EndTime,
		&te.DurationSeconds, &te.Description, &te.IsManual, &te.CreatedAt, &te.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return te, nil
}

func (r *pgTimeTrackingRepository) Update(ctx context.Context, timeEntry *TimeTracking) error {
	query := `
		UPDATE time_tracking 
		SET start_time = $2, end_time = $3, duration_seconds = $4, description = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		timeEntry.ID, timeEntry.StartTime, timeEntry.EndTime, timeEntry.DurationSeconds, timeEntry.Description,
	)
	return err
}

func (r *pgTimeTrackingRepository) StopTimer(ctx context.Context, id string, endTime time.Time) error {
	query := `
		UPDATE time_tracking 
		SET end_time = $2, 
		    duration_seconds = EXTRACT(EPOCH FROM ($2 - start_time))::INT,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, endTime)
	return err
}

func (r *pgTimeTrackingRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM time_tracking WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgTimeTrackingRepository) GetTotalTimeByTask(ctx context.Context, taskID string) (int, error) {
	query := `
		SELECT COALESCE(SUM(duration_seconds), 0) 
		FROM time_tracking 
		WHERE task_id = $1 AND duration_seconds IS NOT NULL
	`
	var total int
	err := r.pool.QueryRow(ctx, query, taskID).Scan(&total)
	return total, err
}

func (r *pgTimeTrackingRepository) GetTotalTimeByUser(ctx context.Context, userID string, startDate, endDate *time.Time) (int, error) {
	query := `
		SELECT COALESCE(SUM(duration_seconds), 0) 
		FROM time_tracking 
		WHERE user_id = $1 AND duration_seconds IS NOT NULL
	`
	args := []interface{}{userID}
	argCount := 1

	if startDate != nil {
		argCount++
		query += ` AND start_time >= $` + string(rune('0'+argCount))
		args = append(args, *startDate)
	}
	if endDate != nil {
		argCount++
		query += ` AND start_time <= $` + string(rune('0'+argCount))
		args = append(args, *endDate)
	}

	var total int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&total)
	return total, err
}

func (r *pgTimeTrackingRepository) GetTotalTimeBySprint(ctx context.Context, sprintID string) (int, error) {
	query := `
		SELECT COALESCE(SUM(tt.duration_seconds), 0) 
		FROM time_tracking tt
		JOIN tasks t ON tt.task_id = t.id
		WHERE t.sprint_id = $1 AND tt.duration_seconds IS NOT NULL
	`
	var total int
	err := r.pool.QueryRow(ctx, query, sprintID).Scan(&total)
	return total, err
}