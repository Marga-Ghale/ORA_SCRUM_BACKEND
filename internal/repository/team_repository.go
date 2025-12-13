package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================
// Team Models
// ============================================

// Team represents a team within a workspace
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

// TeamMember represents a team member
type TeamMember struct {
	ID       string
	TeamID   string
	UserID   string
	Role     string
	JoinedAt time.Time
	User     *User
}

// Invitation represents an invitation to join workspace/team/project
type Invitation struct {
	ID        string
	Email     string
	Token     string
	Type      string // workspace, team, project
	TargetID  string
	Role      string
	InvitedBy string
	Status    string // pending, accepted, expired
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Activity represents an activity log entry
type Activity struct {
	ID         string
	Type       string
	EntityType string
	EntityID   string
	UserID     string
	Changes    map[string]interface{}
	Metadata   map[string]interface{}
	CreatedAt  time.Time
}

// TaskWatcher represents a user watching a task
type TaskWatcher struct {
	ID        string
	TaskID    string
	UserID    string
	CreatedAt time.Time
}

// ============================================
// Team Repository Interface
// ============================================

// TeamRepository defines team data operations
type TeamRepository interface {
	Create(ctx context.Context, team *Team) error
	FindByID(ctx context.Context, id string) (*Team, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Team, error)
	FindByUserID(ctx context.Context, userID string) ([]*Team, error)
	Update(ctx context.Context, team *Team) error
	Delete(ctx context.Context, id string) error

	// Member operations
	AddMember(ctx context.Context, member *TeamMember) error
	FindMembers(ctx context.Context, teamID string) ([]*TeamMember, error)
	FindMember(ctx context.Context, teamID, userID string) (*TeamMember, error)
	FindMemberUserIDs(ctx context.Context, teamID string) ([]string, error)
	UpdateMemberRole(ctx context.Context, teamID, userID, role string) error
	RemoveMember(ctx context.Context, teamID, userID string) error
	IsMember(ctx context.Context, teamID, userID string) (bool, error)
}

// InvitationRepository defines invitation data operations
type InvitationRepository interface {
	Create(ctx context.Context, invitation *Invitation) error
	FindByID(ctx context.Context, id string) (*Invitation, error)
	FindByToken(ctx context.Context, token string) (*Invitation, error)
	FindByEmail(ctx context.Context, email string) ([]*Invitation, error)
	FindPendingByTarget(ctx context.Context, targetType, targetID string) ([]*Invitation, error)
	Update(ctx context.Context, invitation *Invitation) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) (int, error)
}

// ActivityRepository defines activity log operations
type ActivityRepository interface {
	Create(ctx context.Context, activity *Activity) error
	FindByEntity(ctx context.Context, entityType, entityID string, limit int) ([]*Activity, error)
	FindByUser(ctx context.Context, userID string, limit int) ([]*Activity, error)
	DeleteOlderThan(ctx context.Context, olderThan time.Time) (int, error)
}

// TaskWatcherRepository defines task watcher operations
type TaskWatcherRepository interface {
	Add(ctx context.Context, watcher *TaskWatcher) error
	Remove(ctx context.Context, taskID, userID string) error
	FindByTask(ctx context.Context, taskID string) ([]*TaskWatcher, error)
	FindByUser(ctx context.Context, userID string) ([]*TaskWatcher, error)
	IsWatching(ctx context.Context, taskID, userID string) (bool, error)
	GetWatcherUserIDs(ctx context.Context, taskID string) ([]string, error)
}

// ============================================
// PostgreSQL Team Repository Implementation
// ============================================

type pgTeamRepository struct {
	pool *pgxpool.Pool
}

// NewPgTeamRepository creates a new PostgreSQL team repository
func NewPgTeamRepository(pool *pgxpool.Pool) TeamRepository {
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

// ============================================
// PostgreSQL Invitation Repository Implementation
// ============================================

type pgInvitationRepository struct {
	pool *pgxpool.Pool
}

// NewPgInvitationRepository creates a new PostgreSQL invitation repository
func NewPgInvitationRepository(pool *pgxpool.Pool) InvitationRepository {
	return &pgInvitationRepository{pool: pool}
}

func (r *pgInvitationRepository) Create(ctx context.Context, invitation *Invitation) error {
	invitation.Token = uuid.New().String()
	query := `
		INSERT INTO invitations (email, token, type, target_id, role, invited_by, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		invitation.Email, invitation.Token, invitation.Type, invitation.TargetID,
		invitation.Role, invitation.InvitedBy, invitation.Status, invitation.ExpiresAt,
	).Scan(&invitation.ID, &invitation.CreatedAt)
}

func (r *pgInvitationRepository) FindByID(ctx context.Context, id string) (*Invitation, error) {
	query := `
		SELECT id, email, token, type, target_id, role, invited_by, status, expires_at, created_at
		FROM invitations WHERE id = $1
	`
	invitation := &Invitation{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&invitation.ID, &invitation.Email, &invitation.Token, &invitation.Type,
		&invitation.TargetID, &invitation.Role, &invitation.InvitedBy, &invitation.Status,
		&invitation.ExpiresAt, &invitation.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *pgInvitationRepository) FindByToken(ctx context.Context, token string) (*Invitation, error) {
	query := `
		SELECT id, email, token, type, target_id, role, invited_by, status, expires_at, created_at
		FROM invitations WHERE token = $1
	`
	invitation := &Invitation{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&invitation.ID, &invitation.Email, &invitation.Token, &invitation.Type,
		&invitation.TargetID, &invitation.Role, &invitation.InvitedBy, &invitation.Status,
		&invitation.ExpiresAt, &invitation.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *pgInvitationRepository) FindByEmail(ctx context.Context, email string) ([]*Invitation, error) {
	query := `
		SELECT id, email, token, type, target_id, role, invited_by, status, expires_at, created_at
		FROM invitations WHERE email = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*Invitation
	for rows.Next() {
		invitation := &Invitation{}
		if err := rows.Scan(
			&invitation.ID, &invitation.Email, &invitation.Token, &invitation.Type,
			&invitation.TargetID, &invitation.Role, &invitation.InvitedBy, &invitation.Status,
			&invitation.ExpiresAt, &invitation.CreatedAt,
		); err != nil {
			return nil, err
		}
		invitations = append(invitations, invitation)
	}
	return invitations, nil
}

func (r *pgInvitationRepository) FindPendingByTarget(ctx context.Context, targetType, targetID string) ([]*Invitation, error) {
	query := `
		SELECT id, email, token, type, target_id, role, invited_by, status, expires_at, created_at
		FROM invitations WHERE type = $1 AND target_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, targetType, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*Invitation
	for rows.Next() {
		invitation := &Invitation{}
		if err := rows.Scan(
			&invitation.ID, &invitation.Email, &invitation.Token, &invitation.Type,
			&invitation.TargetID, &invitation.Role, &invitation.InvitedBy, &invitation.Status,
			&invitation.ExpiresAt, &invitation.CreatedAt,
		); err != nil {
			return nil, err
		}
		invitations = append(invitations, invitation)
	}
	return invitations, nil
}

func (r *pgInvitationRepository) Update(ctx context.Context, invitation *Invitation) error {
	query := `UPDATE invitations SET status = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, invitation.ID, invitation.Status)
	return err
}

func (r *pgInvitationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM invitations WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) DeleteExpired(ctx context.Context) (int, error) {
	query := `DELETE FROM invitations WHERE expires_at < NOW() AND status = 'pending'`
	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

// ============================================
// PostgreSQL Activity Repository Implementation
// ============================================

type pgActivityRepository struct {
	pool *pgxpool.Pool
}

// NewPgActivityRepository creates a new PostgreSQL activity repository
func NewPgActivityRepository(pool *pgxpool.Pool) ActivityRepository {
	return &pgActivityRepository{pool: pool}
}

func (r *pgActivityRepository) Create(ctx context.Context, activity *Activity) error {
	query := `
		INSERT INTO activities (type, entity_type, entity_id, user_id, changes, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		activity.Type, activity.EntityType, activity.EntityID,
		activity.UserID, activity.Changes, activity.Metadata,
	).Scan(&activity.ID, &activity.CreatedAt)
}

func (r *pgActivityRepository) FindByEntity(ctx context.Context, entityType, entityID string, limit int) ([]*Activity, error) {
	query := `
		SELECT id, type, entity_type, entity_id, user_id, changes, metadata, created_at
		FROM activities WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`
	rows, err := r.pool.Query(ctx, query, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*Activity
	for rows.Next() {
		activity := &Activity{}
		if err := rows.Scan(
			&activity.ID, &activity.Type, &activity.EntityType, &activity.EntityID,
			&activity.UserID, &activity.Changes, &activity.Metadata, &activity.CreatedAt,
		); err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, nil
}

func (r *pgActivityRepository) FindByUser(ctx context.Context, userID string, limit int) ([]*Activity, error) {
	query := `
		SELECT id, type, entity_type, entity_id, user_id, changes, metadata, created_at
		FROM activities WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*Activity
	for rows.Next() {
		activity := &Activity{}
		if err := rows.Scan(
			&activity.ID, &activity.Type, &activity.EntityType, &activity.EntityID,
			&activity.UserID, &activity.Changes, &activity.Metadata, &activity.CreatedAt,
		); err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, nil
}

func (r *pgActivityRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int, error) {
	query := `DELETE FROM activities WHERE created_at < $1`
	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

// ============================================
// PostgreSQL Task Watcher Repository Implementation
// ============================================

type pgTaskWatcherRepository struct {
	pool *pgxpool.Pool
}

// NewPgTaskWatcherRepository creates a new PostgreSQL task watcher repository
func NewPgTaskWatcherRepository(pool *pgxpool.Pool) TaskWatcherRepository {
	return &pgTaskWatcherRepository{pool: pool}
}

func (r *pgTaskWatcherRepository) Add(ctx context.Context, watcher *TaskWatcher) error {
	query := `
		INSERT INTO task_watchers (task_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (task_id, user_id) DO NOTHING
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, watcher.TaskID, watcher.UserID).
		Scan(&watcher.ID, &watcher.CreatedAt)
	if err == pgx.ErrNoRows {
		// Already exists, that's fine
		return nil
	}
	return err
}

func (r *pgTaskWatcherRepository) Remove(ctx context.Context, taskID, userID string) error {
	query := `DELETE FROM task_watchers WHERE task_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, taskID, userID)
	return err
}

func (r *pgTaskWatcherRepository) FindByTask(ctx context.Context, taskID string) ([]*TaskWatcher, error) {
	query := `
		SELECT id, task_id, user_id, created_at
		FROM task_watchers WHERE task_id = $1
	`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchers []*TaskWatcher
	for rows.Next() {
		watcher := &TaskWatcher{}
		if err := rows.Scan(&watcher.ID, &watcher.TaskID, &watcher.UserID, &watcher.CreatedAt); err != nil {
			return nil, err
		}
		watchers = append(watchers, watcher)
	}
	return watchers, nil
}

func (r *pgTaskWatcherRepository) FindByUser(ctx context.Context, userID string) ([]*TaskWatcher, error) {
	query := `
		SELECT id, task_id, user_id, created_at
		FROM task_watchers WHERE user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchers []*TaskWatcher
	for rows.Next() {
		watcher := &TaskWatcher{}
		if err := rows.Scan(&watcher.ID, &watcher.TaskID, &watcher.UserID, &watcher.CreatedAt); err != nil {
			return nil, err
		}
		watchers = append(watchers, watcher)
	}
	return watchers, nil
}

func (r *pgTaskWatcherRepository) IsWatching(ctx context.Context, taskID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM task_watchers WHERE task_id = $1 AND user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, taskID, userID).Scan(&exists)
	return exists, err
}

func (r *pgTaskWatcherRepository) GetWatcherUserIDs(ctx context.Context, taskID string) ([]string, error) {
	query := `SELECT user_id FROM task_watchers WHERE task_id = $1`
	rows, err := r.pool.Query(ctx, query, taskID)
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
