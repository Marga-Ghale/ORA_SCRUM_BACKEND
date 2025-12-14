package repository

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID          string
	Name        string
	Key         string
	Description *string
	Icon        *string
	Color       *string
	SpaceID     string
	LeadID      *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectMember struct {
	ID        string
	ProjectID string
	UserID    string
	Role      string
	JoinedAt  time.Time
	User      *User
}

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	FindByID(ctx context.Context, id string) (*Project, error)
	FindBySpaceID(ctx context.Context, spaceID string) ([]*Project, error)
	FindByKey(ctx context.Context, key string) (*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *ProjectMember) error
	FindMembers(ctx context.Context, projectID string) ([]*ProjectMember, error)
	FindMemberUserIDs(ctx context.Context, projectID string) ([]string, error)
	FindMember(ctx context.Context, projectID, userID string) (*ProjectMember, error)
	RemoveMember(ctx context.Context, projectID, userID string) error
	GetNextTaskNumber(ctx context.Context, projectID string) (int, error)
}

type pgProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) ProjectRepository {
	return &pgProjectRepository{pool: pool}
}

func (r *pgProjectRepository) Create(ctx context.Context, project *Project) error {
	query := `
		INSERT INTO projects (name, key, description, icon, color, space_id, lead_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		project.Name, strings.ToUpper(project.Key), project.Description, project.Icon,
		project.Color, project.SpaceID, project.LeadID,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
}

func (r *pgProjectRepository) FindByID(ctx context.Context, id string) (*Project, error) {
	query := `
		SELECT id, name, key, description, icon, color, space_id, lead_id, created_at, updated_at
		FROM projects WHERE id = $1
	`
	p := &Project{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Key, &p.Description, &p.Icon,
		&p.Color, &p.SpaceID, &p.LeadID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *pgProjectRepository) FindBySpaceID(ctx context.Context, spaceID string) ([]*Project, error) {
	query := `
		SELECT id, name, key, description, icon, color, space_id, lead_id, created_at, updated_at
		FROM projects WHERE space_id = $1 ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Key, &p.Description, &p.Icon,
			&p.Color, &p.SpaceID, &p.LeadID, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *pgProjectRepository) FindByKey(ctx context.Context, key string) (*Project, error) {
	query := `
		SELECT id, name, key, description, icon, color, space_id, lead_id, created_at, updated_at
		FROM projects WHERE UPPER(key) = UPPER($1)
	`
	p := &Project{}
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&p.ID, &p.Name, &p.Key, &p.Description, &p.Icon,
		&p.Color, &p.SpaceID, &p.LeadID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *pgProjectRepository) Update(ctx context.Context, project *Project) error {
	query := `
		UPDATE projects SET name = $2, key = $3, description = $4, icon = $5, color = $6, lead_id = $7, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		project.ID, project.Name, project.Key, project.Description,
		project.Icon, project.Color, project.LeadID,
	)
	return err
}

func (r *pgProjectRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgProjectRepository) AddMember(ctx context.Context, member *ProjectMember) error {
	query := `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, user_id) DO UPDATE SET role = $3
		RETURNING id, joined_at
	`
	return r.pool.QueryRow(ctx, query, member.ProjectID, member.UserID, member.Role).
		Scan(&member.ID, &member.JoinedAt)
}

func (r *pgProjectRepository) FindMembers(ctx context.Context, projectID string) ([]*ProjectMember, error) {
	query := `
		SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.joined_at,
		       u.id, u.email, u.name, u.avatar, u.status
		FROM project_members pm
		JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at
	`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*ProjectMember
	for rows.Next() {
		m := &ProjectMember{User: &User{}}
		if err := rows.Scan(
			&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.JoinedAt,
			&m.User.ID, &m.User.Email, &m.User.Name, &m.User.Avatar, &m.User.Status,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *pgProjectRepository) FindMemberUserIDs(ctx context.Context, projectID string) ([]string, error) {
	query := `SELECT user_id FROM project_members WHERE project_id = $1`
	rows, err := r.pool.Query(ctx, query, projectID)
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

func (r *pgProjectRepository) FindMember(ctx context.Context, projectID, userID string) (*ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, joined_at
		FROM project_members WHERE project_id = $1 AND user_id = $2
	`
	m := &ProjectMember{}
	err := r.pool.QueryRow(ctx, query, projectID, userID).Scan(
		&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.JoinedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *pgProjectRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, projectID, userID)
	return err
}

func (r *pgProjectRepository) GetNextTaskNumber(ctx context.Context, projectID string) (int, error) {
	query := `
		UPDATE projects SET task_counter = task_counter + 1
		WHERE id = $1
		RETURNING task_counter
	`
	var counter int
	err := r.pool.QueryRow(ctx, query, projectID).Scan(&counter)
	return counter, err
}
