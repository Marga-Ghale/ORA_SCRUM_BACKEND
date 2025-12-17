package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID           string
	SpaceID      string   // ✓ Projects belong to spaces (required)
	FolderID     *string  // ✓ Optional folder (can be NULL)
	Name         string
	Key          string   // ✓ Project key (e.g., "PROJ")
	Description  *string
	Icon         *string
	Color        *string
	LeadID       *string  // ✓ Project lead
	Visibility   *string
	AllowedUsers []string
	AllowedTeams []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
	FindByFolderID(ctx context.Context, folderID string) ([]*Project, error)
	FindByUserID(ctx context.Context, userID string) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
	
	// Member operations
	AddMember(ctx context.Context, member *ProjectMember) error
	FindMembers(ctx context.Context, projectID string) ([]*ProjectMember, error)
	FindMember(ctx context.Context, projectID, userID string) (*ProjectMember, error)
	FindMemberUserIDs(ctx context.Context, projectID string) ([]string, error)
	UpdateMemberRole(ctx context.Context, projectID, userID, role string) error
	RemoveMember(ctx context.Context, projectID, userID string) error
	HasAccess(ctx context.Context, projectID, userID string) (bool, error)
}

type pgProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) ProjectRepository {
	return &pgProjectRepository{pool: pool}
}

func (r *pgProjectRepository) Create(ctx context.Context, project *Project) error {
	query := `
		INSERT INTO projects (space_id, folder_id, name, key, description, icon, color, lead_id, visibility, allowed_users, allowed_teams)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		project.SpaceID, project.FolderID, project.Name, project.Key, project.Description,
		project.Icon, project.Color, project.LeadID, project.Visibility,
		project.AllowedUsers, project.AllowedTeams,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
}

func (r *pgProjectRepository) FindByID(ctx context.Context, id string) (*Project, error) {
	query := `
		SELECT id, space_id, folder_id, name, key, description, icon, color, lead_id, visibility, allowed_users, allowed_teams, created_at, updated_at
		FROM projects WHERE id = $1
	`
	p := &Project{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.SpaceID, &p.FolderID, &p.Name, &p.Key, &p.Description,
		&p.Icon, &p.Color, &p.LeadID, &p.Visibility, &p.AllowedUsers, &p.AllowedTeams,
		&p.CreatedAt, &p.UpdatedAt,
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
		SELECT id, space_id, folder_id, name, key, description, icon, color, lead_id, visibility, allowed_users, allowed_teams, created_at, updated_at
		FROM projects
		WHERE space_id = $1
		ORDER BY name
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
			&p.ID, &p.SpaceID, &p.FolderID, &p.Name, &p.Key, &p.Description,
			&p.Icon, &p.Color, &p.LeadID, &p.Visibility, &p.AllowedUsers, &p.AllowedTeams,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *pgProjectRepository) FindByFolderID(ctx context.Context, folderID string) ([]*Project, error) {
	query := `
		SELECT id, space_id, folder_id, name, key, description, icon, color, lead_id, visibility, allowed_users, allowed_teams, created_at, updated_at
		FROM projects
		WHERE folder_id = $1
		ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(
			&p.ID, &p.SpaceID, &p.FolderID, &p.Name, &p.Key, &p.Description,
			&p.Icon, &p.Color, &p.LeadID, &p.Visibility, &p.AllowedUsers, &p.AllowedTeams,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *pgProjectRepository) FindByUserID(ctx context.Context, userID string) ([]*Project, error) {
	query := `
		SELECT p.id, p.space_id, p.folder_id, p.name, p.key, p.description, p.icon, p.color, p.lead_id, p.visibility, p.allowed_users, p.allowed_teams, p.created_at, p.updated_at
		FROM projects p
		JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1
		ORDER BY p.name
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(
			&p.ID, &p.SpaceID, &p.FolderID, &p.Name, &p.Key, &p.Description,
			&p.Icon, &p.Color, &p.LeadID, &p.Visibility, &p.AllowedUsers, &p.AllowedTeams,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *pgProjectRepository) Update(ctx context.Context, project *Project) error {
	query := `
		UPDATE projects 
		SET name = $2, key = $3, description = $4, icon = $5, color = $6, lead_id = $7, 
		    folder_id = $8, visibility = $9, allowed_users = $10, allowed_teams = $11, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		project.ID, project.Name, project.Key, project.Description, project.Icon, project.Color,
		project.LeadID, project.FolderID, project.Visibility, project.AllowedUsers, project.AllowedTeams,
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

func (r *pgProjectRepository) UpdateMemberRole(ctx context.Context, projectID, userID, role string) error {
	query := `UPDATE project_members SET role = $3 WHERE project_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, projectID, userID, role)
	return err
}

func (r *pgProjectRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, projectID, userID)
	return err
}

func (r *pgProjectRepository) HasAccess(ctx context.Context, projectID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM project_members 
			WHERE project_id = $1 AND user_id = $2
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, projectID, userID).Scan(&exists)
	return exists, err
}