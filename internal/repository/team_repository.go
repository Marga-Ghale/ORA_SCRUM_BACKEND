package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Team struct {
	ID          string
	Name        string
	Description *string
	Avatar      *string
	Color       *string
	WorkspaceID string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Members     []*TeamMember
}

type TeamMember struct {
	ID       string
	TeamID   string
	UserID   string
	Role     string
	JoinedAt time.Time
	User     *User
}

type TeamRepository interface {
	Create(ctx context.Context, team *Team) error
	FindByID(ctx context.Context, id string) (*Team, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Team, error)
	FindByUserID(ctx context.Context, userID string) ([]*Team, error)
	Update(ctx context.Context, team *Team) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *TeamMember) error
	FindMembers(ctx context.Context, teamID string) ([]*TeamMember, error)
	FindMember(ctx context.Context, teamID, userID string) (*TeamMember, error)
	FindMemberUserIDs(ctx context.Context, teamID string) ([]string, error)
	UpdateMemberRole(ctx context.Context, teamID, userID, role string) error
	RemoveMember(ctx context.Context, teamID, userID string) error
	IsMember(ctx context.Context, teamID, userID string) (bool, error)
}

type pgTeamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) TeamRepository {
	return &pgTeamRepository{pool: pool}
}

func (r *pgTeamRepository) Create(ctx context.Context, team *Team) error {
	query := `
		INSERT INTO teams (name, description, avatar, color, workspace_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		team.Name, team.Description, team.Avatar, team.Color, team.WorkspaceID, team.CreatedBy,
	).Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)
}

func (r *pgTeamRepository) FindByID(ctx context.Context, id string) (*Team, error) {
	query := `
		SELECT id, name, description, avatar, color, workspace_id, created_by, created_at, updated_at
		FROM teams WHERE id = $1
	`
	team := &Team{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.Description, &team.Avatar, &team.Color,
		&team.WorkspaceID, &team.CreatedBy, &team.CreatedAt, &team.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (r *pgTeamRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Team, error) {
	query := `
		SELECT id, name, description, avatar, color, workspace_id, created_by, created_at, updated_at
		FROM teams WHERE workspace_id = $1
		ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		if err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.Avatar, &team.Color,
			&team.WorkspaceID, &team.CreatedBy, &team.CreatedAt, &team.UpdatedAt,
		); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, nil
}

func (r *pgTeamRepository) FindByUserID(ctx context.Context, userID string) ([]*Team, error) {
	query := `
		SELECT t.id, t.name, t.description, t.avatar, t.color, t.workspace_id, t.created_by, t.created_at, t.updated_at
		FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.name
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		if err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.Avatar, &team.Color,
			&team.WorkspaceID, &team.CreatedBy, &team.CreatedAt, &team.UpdatedAt,
		); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, nil
}

func (r *pgTeamRepository) Update(ctx context.Context, team *Team) error {
	query := `
		UPDATE teams SET name = $2, description = $3, avatar = $4, color = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, team.ID, team.Name, team.Description, team.Avatar, team.Color)
	return err
}

func (r *pgTeamRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgTeamRepository) AddMember(ctx context.Context, member *TeamMember) error {
	query := `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, joined_at
	`
	return r.pool.QueryRow(ctx, query, member.TeamID, member.UserID, member.Role).
		Scan(&member.ID, &member.JoinedAt)
}

func (r *pgTeamRepository) FindMembers(ctx context.Context, teamID string) ([]*TeamMember, error) {
	query := `
		SELECT tm.id, tm.team_id, tm.user_id, tm.role, tm.joined_at,
		       u.id, u.email, u.name, u.avatar, u.status
		FROM team_members tm
		INNER JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1
		ORDER BY tm.joined_at
	`
	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*TeamMember
	for rows.Next() {
		member := &TeamMember{User: &User{}}
		if err := rows.Scan(
			&member.ID, &member.TeamID, &member.UserID, &member.Role, &member.JoinedAt,
			&member.User.ID, &member.User.Email, &member.User.Name, &member.User.Avatar, &member.User.Status,
		); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func (r *pgTeamRepository) FindMember(ctx context.Context, teamID, userID string) (*TeamMember, error) {
	query := `
		SELECT id, team_id, user_id, role, joined_at
		FROM team_members WHERE team_id = $1 AND user_id = $2
	`
	member := &TeamMember{}
	err := r.pool.QueryRow(ctx, query, teamID, userID).Scan(
		&member.ID, &member.TeamID, &member.UserID, &member.Role, &member.JoinedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (r *pgTeamRepository) FindMemberUserIDs(ctx context.Context, teamID string) ([]string, error) {
	query := `SELECT user_id FROM team_members WHERE team_id = $1`
	rows, err := r.pool.Query(ctx, query, teamID)
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

func (r *pgTeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID, role string) error {
	query := `UPDATE team_members SET role = $3 WHERE team_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, teamID, userID, role)
	return err
}

func (r *pgTeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, teamID, userID)
	return err
}

func (r *pgTeamRepository) IsMember(ctx context.Context, teamID, userID string) (bool, error) {
	member, err := r.FindMember(ctx, teamID, userID)
	if err != nil {
		return false, err
	}
	return member != nil, nil
}
