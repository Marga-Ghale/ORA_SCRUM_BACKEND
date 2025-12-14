// internal/repository/space_repository.go
package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Space represents a space entity
type Space struct {
	ID           string
	Name         string
	Description  *string
	Icon         *string
	Color        *string
	WorkspaceID  string
	Visibility   *string  // "private", "workspace", "public"
	AllowedUsers []string // User IDs allowed for private spaces
	AllowedTeams []string // Team IDs allowed for private spaces
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SpaceRepository interface {
	Create(ctx context.Context, space *Space) error
	FindByID(ctx context.Context, id string) (*Space, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Space, error)
	Update(ctx context.Context, space *Space) error
	Delete(ctx context.Context, id string) error
}

type pgSpaceRepository struct {
	pool *pgxpool.Pool
}

func NewSpaceRepository(pool *pgxpool.Pool) SpaceRepository {
	return &pgSpaceRepository{pool: pool}
}

func (r *pgSpaceRepository) Create(ctx context.Context, space *Space) error {
	query := `
		INSERT INTO spaces (name, description, icon, color, workspace_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		space.Name, space.Description, space.Icon, space.Color, space.WorkspaceID,
	).Scan(&space.ID, &space.CreatedAt, &space.UpdatedAt)
}

func (r *pgSpaceRepository) FindByID(ctx context.Context, id string) (*Space, error) {
	query := `
		SELECT id, name, description, icon, color, workspace_id, created_at, updated_at
		FROM spaces WHERE id = $1
	`
	s := &Space{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Description, &s.Icon, &s.Color,
		&s.WorkspaceID, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *pgSpaceRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Space, error) {
	query := `
		SELECT id, name, description, icon, color, workspace_id, created_at, updated_at
		FROM spaces WHERE workspace_id = $1 ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spaces []*Space
	for rows.Next() {
		s := &Space{}
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Description, &s.Icon, &s.Color,
			&s.WorkspaceID, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		spaces = append(spaces, s)
	}
	return spaces, nil
}

func (r *pgSpaceRepository) Update(ctx context.Context, space *Space) error {
	query := `
		UPDATE spaces SET name = $2, description = $3, icon = $4, color = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, space.ID, space.Name, space.Description, space.Icon, space.Color)
	return err
}

func (r *pgSpaceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM spaces WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
