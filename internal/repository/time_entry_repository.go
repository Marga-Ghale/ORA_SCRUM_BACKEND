package repository

import (
	"context"
	"database/sql"
	"time"
)

// TimeEntry model
type TimeEntry struct {
	ID              string     `json:"id" db:"id"`
	TaskID          string     `json:"taskId" db:"task_id"`
	UserID          string     `json:"userId" db:"user_id"`
	StartTime       time.Time  `json:"startTime" db:"start_time"`
	EndTime         *time.Time `json:"endTime,omitempty" db:"end_time"`
	DurationSeconds *int       `json:"durationSeconds,omitempty" db:"duration_seconds"`
	Description     *string    `json:"description,omitempty" db:"description"`
	IsManual        bool       `json:"isManual" db:"is_manual"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
}

// TimeEntryRepository interface
type TimeEntryRepository interface {
	Create(ctx context.Context, entry *TimeEntry) error
	FindByID(ctx context.Context, id string) (*TimeEntry, error)
	FindByTaskID(ctx context.Context, taskID string) ([]*TimeEntry, error)
	FindByUserID(ctx context.Context, userID string) ([]*TimeEntry, error)
	FindActiveTimer(ctx context.Context, userID string) (*TimeEntry, error)
	StopTimer(ctx context.Context, id string) error
	GetTotalTime(ctx context.Context, taskID string) (int, error)
	Delete(ctx context.Context, id string) error
}

// timeEntryRepository implementation
type timeEntryRepository struct {
	db *sql.DB
}

// NewTimeEntryRepository creates a new TimeEntryRepository
func NewTimeEntryRepository(db *sql.DB) TimeEntryRepository {
	return &timeEntryRepository{db: db}
}

// Create inserts a new time entry
func (r *timeEntryRepository) Create(ctx context.Context, entry *TimeEntry) error {
	query := `
		INSERT INTO time_entries (
			id, task_id, user_id, start_time, end_time, duration_seconds, 
			description, is_manual, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW()
		) RETURNING id, created_at`

	return r.db.QueryRowContext(
		ctx, query,
		entry.TaskID,
		entry.UserID,
		entry.StartTime,
		entry.EndTime,
		entry.DurationSeconds,
		entry.Description,
		entry.IsManual,
	).Scan(&entry.ID, &entry.CreatedAt)
}

// FindByID retrieves a time entry by ID
func (r *timeEntryRepository) FindByID(ctx context.Context, id string) (*TimeEntry, error) {
	query := `SELECT * FROM time_entries WHERE id = $1`

	entry := &TimeEntry{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.TaskID,
		&entry.UserID,
		&entry.StartTime,
		&entry.EndTime,
		&entry.DurationSeconds,
		&entry.Description,
		&entry.IsManual,
		&entry.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// FindByTaskID retrieves all time entries for a task
func (r *timeEntryRepository) FindByTaskID(ctx context.Context, taskID string) ([]*TimeEntry, error) {
	query := `SELECT * FROM time_entries WHERE task_id = $1 ORDER BY start_time DESC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*TimeEntry
	for rows.Next() {
		entry := &TimeEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.TaskID,
			&entry.UserID,
			&entry.StartTime,
			&entry.EndTime,
			&entry.DurationSeconds,
			&entry.Description,
			&entry.IsManual,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// FindByUserID retrieves all time entries for a user
func (r *timeEntryRepository) FindByUserID(ctx context.Context, userID string) ([]*TimeEntry, error) {
	query := `SELECT * FROM time_entries WHERE user_id = $1 ORDER BY start_time DESC LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*TimeEntry
	for rows.Next() {
		entry := &TimeEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.TaskID,
			&entry.UserID,
			&entry.StartTime,
			&entry.EndTime,
			&entry.DurationSeconds,
			&entry.Description,
			&entry.IsManual,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// FindActiveTimer retrieves the active timer for a user
func (r *timeEntryRepository) FindActiveTimer(ctx context.Context, userID string) (*TimeEntry, error) {
	query := `SELECT * FROM time_entries WHERE user_id = $1 AND end_time IS NULL AND is_manual = false ORDER BY start_time DESC LIMIT 1`

	entry := &TimeEntry{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&entry.ID,
		&entry.TaskID,
		&entry.UserID,
		&entry.StartTime,
		&entry.EndTime,
		&entry.DurationSeconds,
		&entry.Description,
		&entry.IsManual,
		&entry.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// StopTimer stops an active timer
func (r *timeEntryRepository) StopTimer(ctx context.Context, id string) error {
	query := `
		UPDATE time_entries SET
			end_time = NOW(),
			duration_seconds = EXTRACT(EPOCH FROM (NOW() - start_time))::INTEGER
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetTotalTime calculates total time spent on a task in seconds
func (r *timeEntryRepository) GetTotalTime(ctx context.Context, taskID string) (int, error) {
	query := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN end_time IS NOT NULL THEN duration_seconds
				ELSE EXTRACT(EPOCH FROM (NOW() - start_time))::INTEGER
			END
		), 0) 
		FROM time_entries 
		WHERE task_id = $1`

	var totalSeconds int
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(&totalSeconds)
	return totalSeconds, err
}

// Delete removes a time entry
func (r *timeEntryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM time_entries WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}