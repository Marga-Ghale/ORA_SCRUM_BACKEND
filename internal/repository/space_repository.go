package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Space struct {
	ID           string
	WorkspaceID  string   
	OwnerID      string
	Name         string
	Description  *string
	Icon         *string
	Color        *string
	Visibility   *string
	AllowedUsers []string
	AllowedTeams []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SpaceMember struct {
	ID       string
	SpaceID  string
	UserID   string
	Role     string
	JoinedAt time.Time
	User     *User
}

type SpaceRepository interface {
	Create(ctx context.Context, space *Space) error
	FindByID(ctx context.Context, id string) (*Space, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Space, error) // ✓ NEW
	FindByUserID(ctx context.Context, userID string) ([]*Space, error)
	Update(ctx context.Context, space *Space) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *SpaceMember) error
	FindMembers(ctx context.Context, spaceID string) ([]*SpaceMember, error)
	FindMember(ctx context.Context, spaceID, userID string) (*SpaceMember, error)
	FindMemberUserIDs(ctx context.Context, spaceID string) ([]string, error)
	UpdateMemberRole(ctx context.Context, spaceID, userID, role string) error
	RemoveMember(ctx context.Context, spaceID, userID string) error
	HasAccess(ctx context.Context, spaceID, userID string) (bool, error)
}

type pgSpaceRepository struct {
	pool *pgxpool.Pool
}

func NewSpaceRepository(pool *pgxpool.Pool) SpaceRepository {
	return &pgSpaceRepository{pool: pool}
}

func (r *pgSpaceRepository) Create(ctx context.Context, space *Space) error {
	query := `
		INSERT INTO spaces (workspace_id, name, description, icon, color, owner_id, visibility, allowed_users, allowed_teams)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		space.WorkspaceID, space.Name, space.Description, space.Icon, space.Color, space.OwnerID,
		space.Visibility, space.AllowedUsers, space.AllowedTeams,
	).Scan(&space.ID, &space.CreatedAt, &space.UpdatedAt)
}

func (r *pgSpaceRepository) FindByID(ctx context.Context, id string) (*Space, error) {
	query := `
		SELECT id, workspace_id, name, description, icon, color, owner_id, visibility, allowed_users, allowed_teams, created_at, updated_at
		FROM spaces WHERE id = $1
	`
	s := &Space{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.WorkspaceID, &s.Name, &s.Description, &s.Icon, &s.Color,
		&s.OwnerID, &s.Visibility, &s.AllowedUsers, &s.AllowedTeams,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ✓ NEW - Find all spaces in a workspace
func (r *pgSpaceRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Space, error) {
	query := `
		SELECT id, workspace_id, name, description, icon, color, owner_id, visibility, allowed_users, allowed_teams, created_at, updated_at
		FROM spaces
		WHERE workspace_id = $1
		ORDER BY name
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
			&s.ID, &s.WorkspaceID, &s.Name, &s.Description, &s.Icon, &s.Color,
			&s.OwnerID, &s.Visibility, &s.AllowedUsers, &s.AllowedTeams,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		spaces = append(spaces, s)
	}
	return spaces, nil
}

func (r *pgSpaceRepository) FindByUserID(ctx context.Context, userID string) ([]*Space, error) {
	query := `
		SELECT s.id, s.workspace_id, s.name, s.description, s.icon, s.color, s.owner_id, s.visibility, s.allowed_users, s.allowed_teams, s.created_at, s.updated_at
		FROM spaces s
		JOIN space_members sm ON s.id = sm.space_id
		WHERE sm.user_id = $1
		ORDER BY s.name
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spaces []*Space
	for rows.Next() {
		s := &Space{}
		if err := rows.Scan(
			&s.ID, &s.WorkspaceID, &s.Name, &s.Description, &s.Icon, &s.Color,
			&s.OwnerID, &s.Visibility, &s.AllowedUsers, &s.AllowedTeams,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		spaces = append(spaces, s)
	}
	return spaces, nil
}

func (r *pgSpaceRepository) Update(ctx context.Context, space *Space) error {
	query := `
		UPDATE spaces 
		SET name = $2, description = $3, icon = $4, color = $5, visibility = $6, allowed_users = $7, allowed_teams = $8, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		space.ID, space.Name, space.Description, space.Icon, space.Color,
		space.Visibility, space.AllowedUsers, space.AllowedTeams,
	)
	return err
}

func (r *pgSpaceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM spaces WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgSpaceRepository) AddMember(ctx context.Context, member *SpaceMember) error {
	query := `
		INSERT INTO space_members (space_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (space_id, user_id) DO UPDATE SET role = $3
		RETURNING id, joined_at
	`
	return r.pool.QueryRow(ctx, query, member.SpaceID, member.UserID, member.Role).
		Scan(&member.ID, &member.JoinedAt)
}

func (r *pgSpaceRepository) FindMembers(ctx context.Context, spaceID string) ([]*SpaceMember, error) {
	query := `
		SELECT sm.id, sm.space_id, sm.user_id, sm.role, sm.joined_at,
		       u.id, u.email, u.name, u.avatar, u.status
		FROM space_members sm
		JOIN users u ON sm.user_id = u.id
		WHERE sm.space_id = $1
		ORDER BY sm.joined_at
	`
	rows, err := r.pool.Query(ctx, query, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*SpaceMember
	for rows.Next() {
		m := &SpaceMember{User: &User{}}
		if err := rows.Scan(
			&m.ID, &m.SpaceID, &m.UserID, &m.Role, &m.JoinedAt,
			&m.User.ID, &m.User.Email, &m.User.Name, &m.User.Avatar, &m.User.Status,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *pgSpaceRepository) FindMember(ctx context.Context, spaceID, userID string) (*SpaceMember, error) {
	query := `
		SELECT id, space_id, user_id, role, joined_at
		FROM space_members WHERE space_id = $1 AND user_id = $2
	`
	m := &SpaceMember{}
	err := r.pool.QueryRow(ctx, query, spaceID, userID).Scan(
		&m.ID, &m.SpaceID, &m.UserID, &m.Role, &m.JoinedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *pgSpaceRepository) FindMemberUserIDs(ctx context.Context, spaceID string) ([]string, error) {
	query := `SELECT user_id FROM space_members WHERE space_id = $1`
	rows, err := r.pool.Query(ctx, query, spaceID)
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

func (r *pgSpaceRepository) UpdateMemberRole(ctx context.Context, spaceID, userID, role string) error {
	query := `UPDATE space_members SET role = $3 WHERE space_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, spaceID, userID, role)
	return err
}

func (r *pgSpaceRepository) RemoveMember(ctx context.Context, spaceID, userID string) error {
	query := `DELETE FROM space_members WHERE space_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, spaceID, userID)
	return err
}

func (r *pgSpaceRepository) HasAccess(ctx context.Context, spaceID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM space_members 
			WHERE space_id = $1 AND user_id = $2
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, spaceID, userID).Scan(&exists)
	return exists, err
}