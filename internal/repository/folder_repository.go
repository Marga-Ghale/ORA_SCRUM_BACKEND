package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Folder struct {
	ID           string
	SpaceID      string   // ✓ ADDED - parent space
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

type FolderMember struct {
	ID       string
	FolderID string
	UserID   string
	Role     string
	JoinedAt time.Time
	User     *User
}

type FolderRepository interface {
	Create(ctx context.Context, folder *Folder) error
	FindByID(ctx context.Context, id string) (*Folder, error)
	FindBySpaceID(ctx context.Context, spaceID string) ([]*Folder, error) // ✓ NEW
	FindByUserID(ctx context.Context, userID string) ([]*Folder, error)
	Update(ctx context.Context, folder *Folder) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *FolderMember) error
	FindMembers(ctx context.Context, folderID string) ([]*FolderMember, error)
	FindMember(ctx context.Context, folderID, userID string) (*FolderMember, error)
	FindMemberUserIDs(ctx context.Context, folderID string) ([]string, error)
	UpdateMemberRole(ctx context.Context, folderID, userID, role string) error
	RemoveMember(ctx context.Context, folderID, userID string) error
	HasAccess(ctx context.Context, folderID, userID string) (bool, error)
}

type pgFolderRepository struct {
	pool *pgxpool.Pool
}

func NewFolderRepository(pool *pgxpool.Pool) FolderRepository {
	return &pgFolderRepository{pool: pool}
}

func (r *pgFolderRepository) Create(ctx context.Context, folder *Folder) error {
	query := `
		INSERT INTO folders (space_id, name, description, icon, color, owner_id, visibility, allowed_users, allowed_teams)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		folder.SpaceID, folder.Name, folder.Description, folder.Icon, folder.Color, folder.OwnerID,
		folder.Visibility, folder.AllowedUsers, folder.AllowedTeams,
	).Scan(&folder.ID, &folder.CreatedAt, &folder.UpdatedAt)
}

func (r *pgFolderRepository) FindByID(ctx context.Context, id string) (*Folder, error) {
	query := `
		SELECT id, space_id, name, description, icon, color, owner_id, visibility, allowed_users, allowed_teams, created_at, updated_at
		FROM folders WHERE id = $1
	`
	f := &Folder{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&f.ID, &f.SpaceID, &f.Name, &f.Description, &f.Icon, &f.Color,
		&f.OwnerID, &f.Visibility, &f.AllowedUsers, &f.AllowedTeams,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ✓ NEW - Find all folders in a space
func (r *pgFolderRepository) FindBySpaceID(ctx context.Context, spaceID string) ([]*Folder, error) {
	query := `
		SELECT id, space_id, name, description, icon, color, owner_id, visibility, allowed_users, allowed_teams, created_at, updated_at
		FROM folders
		WHERE space_id = $1
		ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*Folder
	for rows.Next() {
		f := &Folder{}
		if err := rows.Scan(
			&f.ID, &f.SpaceID, &f.Name, &f.Description, &f.Icon, &f.Color,
			&f.OwnerID, &f.Visibility, &f.AllowedUsers, &f.AllowedTeams,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}
	return folders, nil
}

func (r *pgFolderRepository) FindByUserID(ctx context.Context, userID string) ([]*Folder, error) {
	query := `
		SELECT f.id, f.space_id, f.name, f.description, f.icon, f.color, f.owner_id, f.visibility, f.allowed_users, f.allowed_teams, f.created_at, f.updated_at
		FROM folders f
		JOIN folder_members fm ON f.id = fm.folder_id
		WHERE fm.user_id = $1
		ORDER BY f.name
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*Folder
	for rows.Next() {
		f := &Folder{}
		if err := rows.Scan(
			&f.ID, &f.SpaceID, &f.Name, &f.Description, &f.Icon, &f.Color,
			&f.OwnerID, &f.Visibility, &f.AllowedUsers, &f.AllowedTeams,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}
	return folders, nil
}

func (r *pgFolderRepository) Update(ctx context.Context, folder *Folder) error {
	query := `
		UPDATE folders 
		SET name = $2, description = $3, icon = $4, color = $5, visibility = $6, allowed_users = $7, allowed_teams = $8, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		folder.ID, folder.Name, folder.Description, folder.Icon, folder.Color,
		folder.Visibility, folder.AllowedUsers, folder.AllowedTeams,
	)
	return err
}

func (r *pgFolderRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM folders WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgFolderRepository) AddMember(ctx context.Context, member *FolderMember) error {
	query := `
		INSERT INTO folder_members (folder_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (folder_id, user_id) DO UPDATE SET role = $3
		RETURNING id, joined_at
	`
	return r.pool.QueryRow(ctx, query, member.FolderID, member.UserID, member.Role).
		Scan(&member.ID, &member.JoinedAt)
}

func (r *pgFolderRepository) FindMembers(ctx context.Context, folderID string) ([]*FolderMember, error) {
	query := `
		SELECT fm.id, fm.folder_id, fm.user_id, fm.role, fm.joined_at,
		       u.id, u.email, u.name, u.avatar, u.status
		FROM folder_members fm
		JOIN users u ON fm.user_id = u.id
		WHERE fm.folder_id = $1
		ORDER BY fm.joined_at
	`
	rows, err := r.pool.Query(ctx, query, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*FolderMember
	for rows.Next() {
		m := &FolderMember{User: &User{}}
		if err := rows.Scan(
			&m.ID, &m.FolderID, &m.UserID, &m.Role, &m.JoinedAt,
			&m.User.ID, &m.User.Email, &m.User.Name, &m.User.Avatar, &m.User.Status,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *pgFolderRepository) FindMember(ctx context.Context, folderID, userID string) (*FolderMember, error) {
	query := `
		SELECT id, folder_id, user_id, role, joined_at
		FROM folder_members WHERE folder_id = $1 AND user_id = $2
	`
	m := &FolderMember{}
	err := r.pool.QueryRow(ctx, query, folderID, userID).Scan(
		&m.ID, &m.FolderID, &m.UserID, &m.Role, &m.JoinedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *pgFolderRepository) FindMemberUserIDs(ctx context.Context, folderID string) ([]string, error) {
	query := `SELECT user_id FROM folder_members WHERE folder_id = $1`
	rows, err := r.pool.Query(ctx, query, folderID)
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

func (r *pgFolderRepository) UpdateMemberRole(ctx context.Context, folderID, userID, role string) error {
	query := `UPDATE folder_members SET role = $3 WHERE folder_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, folderID, userID, role)
	return err
}

func (r *pgFolderRepository) RemoveMember(ctx context.Context, folderID, userID string) error {
	query := `DELETE FROM folder_members WHERE folder_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, folderID, userID)
	return err
}

func (r *pgFolderRepository) HasAccess(ctx context.Context, folderID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM folder_members 
			WHERE folder_id = $1 AND user_id = $2
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, folderID, userID).Scan(&exists)
	return exists, err
}