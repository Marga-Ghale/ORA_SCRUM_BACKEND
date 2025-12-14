package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Sprint struct {
	ID        string
	Name      string
	Goal      *string
	ProjectID string
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SprintRepository interface {
	Create(ctx context.Context, sprint *Sprint) error
	FindByID(ctx context.Context, id string) (*Sprint, error)
	FindByProjectID(ctx context.Context, projectID string) ([]*Sprint, error)
	FindActive(ctx context.Context, projectID string) (*Sprint, error)
	FindEndingSoon(ctx context.Context, within time.Duration) ([]*Sprint, error)
	FindExpired(ctx context.Context) ([]*Sprint, error)
	Update(ctx context.Context, sprint *Sprint) error
	Delete(ctx context.Context, id string) error
}

type pgSprintRepository struct {
	pool *pgxpool.Pool
}

func NewSprintRepository(pool *pgxpool.Pool) SprintRepository {
	return &pgSprintRepository{pool: pool}
}

func (r *pgSprintRepository) Create(ctx context.Context, sprint *Sprint) error {
	if sprint.Status == "" {
		sprint.Status = "planning"
	}
	query := `
		INSERT INTO sprints (name, goal, project_id, status, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		sprint.Name, sprint.Goal, sprint.ProjectID, sprint.Status,
		sprint.StartDate, sprint.EndDate,
	).Scan(&sprint.ID, &sprint.CreatedAt, &sprint.UpdatedAt)
}

func (r *pgSprintRepository) FindByID(ctx context.Context, id string) (*Sprint, error) {
	query := `
		SELECT id, name, goal, project_id, status, start_date, end_date, created_at, updated_at
		FROM sprints WHERE id = $1
	`
	s := &Sprint{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Goal, &s.ProjectID, &s.Status,
		&s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *pgSprintRepository) FindByProjectID(ctx context.Context, projectID string) ([]*Sprint, error) {
	query := `
		SELECT id, name, goal, project_id, status, start_date, end_date, created_at, updated_at
		FROM sprints WHERE project_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		s := &Sprint{}
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Goal, &s.ProjectID, &s.Status,
			&s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sprints = append(sprints, s)
	}
	return sprints, nil
}

func (r *pgSprintRepository) FindActive(ctx context.Context, projectID string) (*Sprint, error) {
	query := `
		SELECT id, name, goal, project_id, status, start_date, end_date, created_at, updated_at
		FROM sprints WHERE project_id = $1 AND status = 'active'
		LIMIT 1
	`
	s := &Sprint{}
	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&s.ID, &s.Name, &s.Goal, &s.ProjectID, &s.Status,
		&s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *pgSprintRepository) FindEndingSoon(ctx context.Context, within time.Duration) ([]*Sprint, error) {
	query := `
		SELECT id, name, goal, project_id, status, start_date, end_date, created_at, updated_at
		FROM sprints 
		WHERE status = 'active' AND end_date IS NOT NULL 
		AND end_date > NOW() AND end_date < $1
	`
	deadline := time.Now().Add(within)
	rows, err := r.pool.Query(ctx, query, deadline)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		s := &Sprint{}
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Goal, &s.ProjectID, &s.Status,
			&s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sprints = append(sprints, s)
	}
	return sprints, nil
}

func (r *pgSprintRepository) FindExpired(ctx context.Context) ([]*Sprint, error) {
	query := `
		SELECT id, name, goal, project_id, status, start_date, end_date, created_at, updated_at
		FROM sprints 
		WHERE status = 'active' AND end_date IS NOT NULL AND end_date < NOW()
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		s := &Sprint{}
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Goal, &s.ProjectID, &s.Status,
			&s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sprints = append(sprints, s)
	}
	return sprints, nil
}

func (r *pgSprintRepository) Update(ctx context.Context, sprint *Sprint) error {
	query := `
		UPDATE sprints SET name = $2, goal = $3, status = $4, start_date = $5, end_date = $6, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		sprint.ID, sprint.Name, sprint.Goal, sprint.Status,
		sprint.StartDate, sprint.EndDate,
	)
	return err
}

func (r *pgSprintRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sprints WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
