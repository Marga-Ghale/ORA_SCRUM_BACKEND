package repository

import (
	"context"
	"database/sql"
	"time"
)

// Sprint model
type Sprint struct {
	ID        string     `json:"id" db:"id"`
	ProjectID string     `json:"projectId" db:"project_id"`
	Name      string     `json:"name" db:"name"`
	Goal      *string    `json:"goal,omitempty" db:"goal"`
	Status    string     `json:"status" db:"status"` // "planning", "active", "completed"
	StartDate time.Time  `json:"startDate" db:"start_date"`
	EndDate   time.Time  `json:"endDate" db:"end_date"`
	CreatedBy string     `json:"createdBy" db:"created_by"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}

// SprintRepository interface
type SprintRepository interface {
	Create(ctx context.Context, sprint *Sprint) error
	FindByID(ctx context.Context, id string) (*Sprint, error)
	FindByProjectID(ctx context.Context, projectID string) ([]*Sprint, error)
	Update(ctx context.Context, sprint *Sprint) error
	UpdateStatus(ctx context.Context, id, status string) error
	Delete(ctx context.Context, id string) error
	FindActiveSprint(ctx context.Context, projectID string) (*Sprint, error)
	querySprints(ctx context.Context, query string, args ...interface{}) ([]*Sprint, error)
	FindSprintsEndingSoon(ctx context.Context, within time.Duration) ([]*Sprint, error)
	FindExpiredSprints(ctx context.Context) ([]*Sprint, error)

}

// sprintRepository implementation
type sprintRepository struct {
	db *sql.DB
}

// NewSprintRepository creates a new SprintRepository
func NewSprintRepository(db *sql.DB) SprintRepository {
	return &sprintRepository{db: db}
}

// Create inserts a new sprint
func (r *sprintRepository) Create(ctx context.Context, sprint *Sprint) error {
	query := `
		INSERT INTO sprints (
			id, project_id, name, goal, status, start_date, end_date, 
			created_by, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
		) RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		sprint.ProjectID,
		sprint.Name,
		sprint.Goal,
		sprint.Status,
		sprint.StartDate,
		sprint.EndDate,
		sprint.CreatedBy,
	).Scan(&sprint.ID, &sprint.CreatedAt, &sprint.UpdatedAt)
}

// FindByID retrieves a sprint by ID
func (r *sprintRepository) FindByID(ctx context.Context, id string) (*Sprint, error) {
	query := `SELECT * FROM sprints WHERE id = $1`

	sprint := &Sprint{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sprint.ID,
		&sprint.ProjectID,
		&sprint.Name,
		&sprint.Goal,
		&sprint.Status,
		&sprint.StartDate,
		&sprint.EndDate,
		&sprint.CreatedBy,
		&sprint.CreatedAt,
		&sprint.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return sprint, nil
}

// FindByProjectID retrieves all sprints for a project
func (r *sprintRepository) FindByProjectID(ctx context.Context, projectID string) ([]*Sprint, error) {
	query := `SELECT * FROM sprints WHERE project_id = $1 ORDER BY start_date DESC`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		sprint := &Sprint{}
		err := rows.Scan(
			&sprint.ID,
			&sprint.ProjectID,
			&sprint.Name,
			&sprint.Goal,
			&sprint.Status,
			&sprint.StartDate,
			&sprint.EndDate,
			&sprint.CreatedBy,
			&sprint.CreatedAt,
			&sprint.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sprints = append(sprints, sprint)
	}

	return sprints, rows.Err()
}

// FindActiveSprint retrieves the currently active sprint for a project
func (r *sprintRepository) FindActiveSprint(ctx context.Context, projectID string) (*Sprint, error) {
	query := `SELECT * FROM sprints WHERE project_id = $1 AND status = 'active' ORDER BY start_date DESC LIMIT 1`

	sprint := &Sprint{}
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&sprint.ID,
		&sprint.ProjectID,
		&sprint.Name,
		&sprint.Goal,
		&sprint.Status,
		&sprint.StartDate,
		&sprint.EndDate,
		&sprint.CreatedBy,
		&sprint.CreatedAt,
		&sprint.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return sprint, nil
}

// Update updates an existing sprint
func (r *sprintRepository) Update(ctx context.Context, sprint *Sprint) error {
	query := `
		UPDATE sprints SET
			name = $2,
			goal = $3,
			status = $4,
			start_date = $5,
			end_date = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		sprint.ID,
		sprint.Name,
		sprint.Goal,
		sprint.Status,
		sprint.StartDate,
		sprint.EndDate,
	).Scan(&sprint.UpdatedAt)
}

// UpdateStatus updates the status of a sprint
func (r *sprintRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `UPDATE sprints SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// Delete removes a sprint
func (r *sprintRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sprints WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// FindSprintsEndingSoon returns sprints ending within the next 'within' duration
func (r *sprintRepository) FindSprintsEndingSoon(ctx context.Context, within time.Duration) ([]*Sprint, error) {
	query := `
		SELECT * FROM sprints 
		WHERE end_date BETWEEN NOW() AND NOW() + $1::interval 
		  AND status != 'completed'
		ORDER BY end_date ASC`
	return r.querySprints(ctx, query, within.String())
}

// FindExpiredSprints returns sprints whose end_date has passed but not completed
func (r *sprintRepository) FindExpiredSprints(ctx context.Context) ([]*Sprint, error) {
	query := `SELECT * FROM sprints WHERE end_date < NOW() AND status != 'completed' ORDER BY end_date ASC`
	return r.querySprints(ctx, query)
}

// Helper to run sprint queries
func (r *sprintRepository) querySprints(ctx context.Context, query string, args ...interface{}) ([]*Sprint, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		s := &Sprint{}
		err := rows.Scan(
			&s.ID, &s.ProjectID, &s.Name, &s.Goal, &s.Status,
			&s.StartDate, &s.EndDate, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sprints = append(sprints, s)
	}
	return sprints, rows.Err()
}
