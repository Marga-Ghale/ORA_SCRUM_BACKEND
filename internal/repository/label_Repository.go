package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Label struct {
	ID        string
	Name      string
	Color     string
	ProjectID string
	CreatedAt time.Time
}

type LabelRepository interface {
	Create(ctx context.Context, label *Label) error
	FindByID(ctx context.Context, id string) (*Label, error)
	FindByProjectID(ctx context.Context, projectID string) ([]*Label, error)
	FindByName(ctx context.Context, projectID, name string) (*Label, error)
	Update(ctx context.Context, label *Label) error
	Delete(ctx context.Context, id string) error
}

type pgLabelRepository struct {
	pool *pgxpool.Pool
}

func NewLabelRepository(pool *pgxpool.Pool) LabelRepository {
	return &pgLabelRepository{pool: pool}
}

func (r *pgLabelRepository) Create(ctx context.Context, label *Label) error {
	query := `
		INSERT INTO labels (name, color, project_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query, label.Name, label.Color, label.ProjectID).
		Scan(&label.ID, &label.CreatedAt)
}

func (r *pgLabelRepository) FindByID(ctx context.Context, id string) (*Label, error) {
	query := `SELECT id, name, color, project_id, created_at FROM labels WHERE id = $1`
	l := &Label{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&l.ID, &l.Name, &l.Color, &l.ProjectID, &l.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *pgLabelRepository) FindByProjectID(ctx context.Context, projectID string) ([]*Label, error) {
	query := `SELECT id, name, color, project_id, created_at FROM labels WHERE project_id = $1 ORDER BY name`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*Label
	for rows.Next() {
		l := &Label{}
		if err := rows.Scan(&l.ID, &l.Name, &l.Color, &l.ProjectID, &l.CreatedAt); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, nil
}

func (r *pgLabelRepository) FindByName(ctx context.Context, projectID, name string) (*Label, error) {
	query := `SELECT id, name, color, project_id, created_at FROM labels WHERE project_id = $1 AND LOWER(name) = LOWER($2)`
	l := &Label{}
	err := r.pool.QueryRow(ctx, query, projectID, name).Scan(&l.ID, &l.Name, &l.Color, &l.ProjectID, &l.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *pgLabelRepository) Update(ctx context.Context, label *Label) error {
	query := `UPDATE labels SET name = $2, color = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, label.ID, label.Name, label.Color)
	return err
}

func (r *pgLabelRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM labels WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
