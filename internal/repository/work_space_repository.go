package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Workspace struct {
	ID           string
	OwnerID      string
	Name         string
	Description  *string
	Icon         *string
	Color        *string
	Visibility   *string  // "private", "workspace", "public"
	AllowedUsers []string // User IDs allowed for private workspace
	AllowedTeams []string // Team IDs allowed for private workspace
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type WorkspaceMember struct {
	ID          string
	WorkspaceID string
	UserID      string
	Role        string
	JoinedAt    time.Time
	User        *User
}

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	FindByID(ctx context.Context, id string) (*Workspace, error)
	FindByUserID(ctx context.Context, userID string) ([]*Workspace, error)
	Update(ctx context.Context, workspace *Workspace) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *WorkspaceMember) error
	FindMembers(ctx context.Context, workspaceID string) ([]*WorkspaceMember, error)
	FindMember(ctx context.Context, workspaceID, userID string) (*WorkspaceMember, error)
	FindMemberUserIDs(ctx context.Context, workspaceID string) ([]string, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
	HasAccess(ctx context.Context, workspaceID, userID string) (bool, error)
}

type pgWorkspaceRepository struct {
	pool *pgxpool.Pool
}

func NewWorkspaceRepository(pool *pgxpool.Pool) WorkspaceRepository {
	return &pgWorkspaceRepository{pool: pool}
}

func (r *pgWorkspaceRepository) Create(ctx context.Context, workspace *Workspace) error {
	query := `
		INSERT INTO workspaces (name, description, icon, color, owner_id, visibility, allowed_users, allowed_teams)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		workspace.Name, workspace.Description, workspace.Icon, workspace.Color, workspace.OwnerID,
		workspace.Visibility, workspace.AllowedUsers, workspace.AllowedTeams,
	).Scan(&workspace.ID, &workspace.CreatedAt, &workspace.UpdatedAt)
}

func (r *pgWorkspaceRepository) FindByID(ctx context.Context, id string) (*Workspace, error) {
	query := `
		SELECT id, name, description, icon, color, owner_id, visibility, allowed_users, allowed_teams, created_at, updated_at
		FROM workspaces WHERE id = $1
	`
	ws := &Workspace{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ws.ID, &ws.Name, &ws.Description, &ws.Icon, &ws.Color,
		&ws.OwnerID, &ws.Visibility, &ws.AllowedUsers, &ws.AllowedTeams,
		&ws.CreatedAt, &ws.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (r *pgWorkspaceRepository) FindByUserID(ctx context.Context, userID string) ([]*Workspace, error) {
	query := `
		SELECT w.id, w.name, w.description, w.icon, w.color, w.owner_id, w.visibility, w.allowed_users, w.allowed_teams, w.created_at, w.updated_at
		FROM workspaces w
		JOIN workspace_members wm ON w.id = wm.workspace_id
		WHERE wm.user_id = $1
		ORDER BY w.name
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*Workspace
	for rows.Next() {
		ws := &Workspace{}
		if err := rows.Scan(
			&ws.ID, &ws.Name, &ws.Description, &ws.Icon, &ws.Color,
			&ws.OwnerID, &ws.Visibility, &ws.AllowedUsers, &ws.AllowedTeams,
			&ws.CreatedAt, &ws.UpdatedAt,
		); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

func (r *pgWorkspaceRepository) Update(ctx context.Context, workspace *Workspace) error {
	query := `
		UPDATE workspaces 
		SET name = $2, description = $3, icon = $4, color = $5, visibility = $6, allowed_users = $7, allowed_teams = $8, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		workspace.ID, workspace.Name, workspace.Description, workspace.Icon, workspace.Color,
		workspace.Visibility, workspace.AllowedUsers, workspace.AllowedTeams,
	)
	return err
}

func (r *pgWorkspaceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM workspaces WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgWorkspaceRepository) AddMember(ctx context.Context, member *WorkspaceMember) error {
	query := `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = $3
		RETURNING id, joined_at
	`
	return r.pool.QueryRow(ctx, query, member.WorkspaceID, member.UserID, member.Role).
		Scan(&member.ID, &member.JoinedAt)
}

func (r *pgWorkspaceRepository) FindMembers(ctx context.Context, workspaceID string) ([]*WorkspaceMember, error) {
	query := `
		SELECT wm.id, wm.workspace_id, wm.user_id, wm.role, wm.joined_at,
		       u.id, u.email, u.name, u.avatar, u.status
		FROM workspace_members wm
		JOIN users u ON wm.user_id = u.id
		WHERE wm.workspace_id = $1
		ORDER BY wm.joined_at
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*WorkspaceMember
	for rows.Next() {
		m := &WorkspaceMember{User: &User{}}
		if err := rows.Scan(
			&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &m.JoinedAt,
			&m.User.ID, &m.User.Email, &m.User.Name, &m.User.Avatar, &m.User.Status,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *pgWorkspaceRepository) FindMember(ctx context.Context, workspaceID, userID string) (*WorkspaceMember, error) {
	query := `
		SELECT id, workspace_id, user_id, role, joined_at
		FROM workspace_members WHERE workspace_id = $1 AND user_id = $2
	`
	m := &WorkspaceMember{}
	err := r.pool.QueryRow(ctx, query, workspaceID, userID).Scan(
		&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &m.JoinedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *pgWorkspaceRepository) FindMemberUserIDs(ctx context.Context, workspaceID string) ([]string, error) {
	query := `SELECT user_id FROM workspace_members WHERE workspace_id = $1`
	rows, err := r.pool.Query(ctx, query, workspaceID)
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

func (r *pgWorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	query := `UPDATE workspace_members SET role = $3 WHERE workspace_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, workspaceID, userID, role)
	return err
}

func (r *pgWorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	query := `DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, workspaceID, userID)
	return err
}

func (r *pgWorkspaceRepository) HasAccess(ctx context.Context, workspaceID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM workspace_members 
			WHERE workspace_id = $1 AND user_id = $2
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, workspaceID, userID).Scan(&exists)
	return exists, err
}