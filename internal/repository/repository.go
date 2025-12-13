// internal/repository/repository.go
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================
// Models / Entities
// ============================================

type User struct {
	ID           string
	Email        string
	Password     string
	Name         string
	Avatar       *string
	Status       string
	LastActiveAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Workspace struct {
	ID          string
	Name        string
	Description *string
	Icon        *string
	Color       *string
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceMember struct {
	ID          string
	WorkspaceID string
	UserID      string
	Role        string
	JoinedAt    time.Time
	User        *User
}

type Space struct {
	ID          string
	Name        string
	Description *string
	Icon        *string
	Color       *string
	WorkspaceID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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

type Task struct {
	ID          string
	Key         string
	Title       string
	Description *string
	Status      string
	Priority    string
	Type        string
	ProjectID   string
	SprintID    *string
	AssigneeID  *string
	ReporterID  string
	ParentID    *string
	StoryPoints *int
	DueDate     *time.Time
	OrderIndex  int
	Labels      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Assignee    *User
	Reporter    *User
}

type TaskFilters struct {
	Status     []string
	Priority   []string
	Type       []string
	AssigneeID []string
	SprintID   *string
	Labels     []string
	Search     string
	Limit      int
	Offset     int
}

type BulkTaskUpdate struct {
	ID         string
	Status     *string
	SprintID   *string
	OrderIndex *int
}

type Comment struct {
	ID        string
	Content   string
	TaskID    string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	User      *User
}

type Label struct {
	ID        string
	Name      string
	Color     string
	ProjectID string
	CreatedAt time.Time
}

type Notification struct {
	ID        string
	UserID    string
	Type      string
	Title     string
	Message   string
	Read      bool
	Data      map[string]interface{}
	CreatedAt time.Time
}

// ============================================
// Repository Interfaces
// ============================================

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByName(ctx context.Context, name string) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	UpdateLastActive(ctx context.Context, userID string) error
	UpdateStatusForInactive(ctx context.Context, inactiveDuration time.Duration) error
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteUserRefreshTokens(ctx context.Context, userID string) error
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
}

type SpaceRepository interface {
	Create(ctx context.Context, space *Space) error
	FindByID(ctx context.Context, id string) (*Space, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Space, error)
	Update(ctx context.Context, space *Space) error
	Delete(ctx context.Context, id string) error
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

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindByKey(ctx context.Context, key string) (*Task, error)
	FindByProjectID(ctx context.Context, projectID string, filters *TaskFilters) ([]*Task, error)
	FindBySprintID(ctx context.Context, sprintID string) ([]*Task, error)
	FindBacklog(ctx context.Context, projectID string) ([]*Task, error)
	FindOverdue(ctx context.Context) ([]*Task, error)
	FindDueSoon(ctx context.Context, within time.Duration) ([]*Task, error)
	FindByAssignee(ctx context.Context, assigneeID string) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
	BulkUpdate(ctx context.Context, updates []BulkTaskUpdate) error
	CountBySprintID(ctx context.Context, sprintID string) (total int, completed int, err error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) error
	FindByID(ctx context.Context, id string) (*Comment, error)
	FindByTaskID(ctx context.Context, taskID string) ([]*Comment, error)
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id string) error
}

type LabelRepository interface {
	Create(ctx context.Context, label *Label) error
	FindByID(ctx context.Context, id string) (*Label, error)
	FindByProjectID(ctx context.Context, projectID string) ([]*Label, error)
	FindByName(ctx context.Context, projectID, name string) (*Label, error)
	Update(ctx context.Context, label *Label) error
	Delete(ctx context.Context, id string) error
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	FindByID(ctx context.Context, id string) (*Notification, error)
	FindByUserID(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error)
	CountByUserID(ctx context.Context, userID string) (total int, unread int, err error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
	DeleteAll(ctx context.Context, userID string) error
	DeleteOlderThan(ctx context.Context, olderThan time.Time, readOnly bool) (int, error)
}

// ============================================
// Repositories Container
// ============================================

type Repositories struct {
	UserRepo         UserRepository
	WorkspaceRepo    WorkspaceRepository
	SpaceRepo        SpaceRepository
	ProjectRepo      ProjectRepository
	SprintRepo       SprintRepository
	TaskRepo         TaskRepository
	CommentRepo      CommentRepository
	LabelRepo        LabelRepository
	NotificationRepo NotificationRepository
}

// NewRepositories creates in-memory repositories (for testing/fallback)
func NewRepositories() *Repositories {
	return &Repositories{
		UserRepo:         newInMemoryUserRepository(),
		WorkspaceRepo:    newInMemoryWorkspaceRepository(),
		SpaceRepo:        newInMemorySpaceRepository(),
		ProjectRepo:      newInMemoryProjectRepository(),
		SprintRepo:       newInMemorySprintRepository(),
		TaskRepo:         newInMemoryTaskRepository(),
		CommentRepo:      newInMemoryCommentRepository(),
		LabelRepo:        newInMemoryLabelRepository(),
		NotificationRepo: newInMemoryNotificationRepository(),
	}
}

// NewPgRepositories creates PostgreSQL-backed repositories
func NewPgRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepo:         &pgUserRepository{pool: pool},
		WorkspaceRepo:    &pgWorkspaceRepository{pool: pool},
		SpaceRepo:        &pgSpaceRepository{pool: pool},
		ProjectRepo:      &pgProjectRepository{pool: pool},
		SprintRepo:       &pgSprintRepository{pool: pool},
		TaskRepo:         &pgTaskRepository{pool: pool},
		CommentRepo:      &pgCommentRepository{pool: pool},
		LabelRepo:        &pgLabelRepository{pool: pool},
		NotificationRepo: &pgNotificationRepository{pool: pool},
	}
}

// ============================================
// PostgreSQL User Repository
// ============================================

type pgUserRepository struct {
	pool *pgxpool.Pool
}

func (r *pgUserRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (email, password, name, avatar, status, last_active_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	user.LastActiveAt = &now
	if user.Status == "" {
		user.Status = "online"
	}
	return r.pool.QueryRow(ctx, query,
		user.Email, user.Password, user.Name, user.Avatar, user.Status, now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *pgUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, password, name, avatar, status, last_active_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	user := &User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar,
		&user.Status, &user.LastActiveAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *pgUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password, name, avatar, status, last_active_at, created_at, updated_at
		FROM users WHERE LOWER(email) = LOWER($1)
	`
	user := &User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar,
		&user.Status, &user.LastActiveAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *pgUserRepository) FindByName(ctx context.Context, name string) (*User, error) {
	query := `
		SELECT id, email, password, name, avatar, status, last_active_at, created_at, updated_at
		FROM users WHERE LOWER(name) LIKE LOWER($1)
		LIMIT 1
	`
	user := &User{}
	err := r.pool.QueryRow(ctx, query, "%"+name+"%").Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar,
		&user.Status, &user.LastActiveAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *pgUserRepository) FindAll(ctx context.Context) ([]*User, error) {
	query := `
		SELECT id, email, password, name, avatar, status, last_active_at, created_at, updated_at
		FROM users ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(
			&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar,
			&user.Status, &user.LastActiveAt, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *pgUserRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users SET email = $2, name = $3, avatar = $4, status = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.Name, user.Avatar, user.Status)
	return err
}

func (r *pgUserRepository) UpdateLastActive(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_active_at = NOW(), status = 'online' WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *pgUserRepository) UpdateStatusForInactive(ctx context.Context, inactiveDuration time.Duration) error {
	query := `
		UPDATE users SET status = 'away'
		WHERE status = 'online' AND last_active_at < $1
	`
	threshold := time.Now().Add(-inactiveDuration)
	_, err := r.pool.Exec(ctx, query, threshold)
	return err
}

func (r *pgUserRepository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query, token.Token, token.UserID, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt)
}

func (r *pgUserRepository) FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	query := `
		SELECT id, token, user_id, expires_at, created_at
		FROM refresh_tokens WHERE token = $1
	`
	rt := &RefreshToken{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&rt.ID, &rt.Token, &rt.UserID, &rt.ExpiresAt, &rt.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *pgUserRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.pool.Exec(ctx, query, token)
	return err
}

func (r *pgUserRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// ============================================
// PostgreSQL Workspace Repository
// ============================================

type pgWorkspaceRepository struct {
	pool *pgxpool.Pool
}

func (r *pgWorkspaceRepository) Create(ctx context.Context, workspace *Workspace) error {
	query := `
		INSERT INTO workspaces (name, description, icon, color, owner_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		workspace.Name, workspace.Description, workspace.Icon, workspace.Color, workspace.OwnerID,
	).Scan(&workspace.ID, &workspace.CreatedAt, &workspace.UpdatedAt)
}

func (r *pgWorkspaceRepository) FindByID(ctx context.Context, id string) (*Workspace, error) {
	query := `
		SELECT id, name, description, icon, color, owner_id, created_at, updated_at
		FROM workspaces WHERE id = $1
	`
	ws := &Workspace{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ws.ID, &ws.Name, &ws.Description, &ws.Icon, &ws.Color,
		&ws.OwnerID, &ws.CreatedAt, &ws.UpdatedAt,
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
		SELECT w.id, w.name, w.description, w.icon, w.color, w.owner_id, w.created_at, w.updated_at
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
			&ws.OwnerID, &ws.CreatedAt, &ws.UpdatedAt,
		); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

func (r *pgWorkspaceRepository) Update(ctx context.Context, workspace *Workspace) error {
	query := `
		UPDATE workspaces SET name = $2, description = $3, icon = $4, color = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		workspace.ID, workspace.Name, workspace.Description, workspace.Icon, workspace.Color,
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

// ============================================
// PostgreSQL Space Repository
// ============================================

type pgSpaceRepository struct {
	pool *pgxpool.Pool
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

// ============================================
// PostgreSQL Project Repository
// ============================================

type pgProjectRepository struct {
	pool *pgxpool.Pool
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

// ============================================
// PostgreSQL Sprint Repository
// ============================================

type pgSprintRepository struct {
	pool *pgxpool.Pool
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

// ============================================
// PostgreSQL Task Repository
// ============================================

type pgTaskRepository struct {
	pool *pgxpool.Pool
}

func (r *pgTaskRepository) Create(ctx context.Context, task *Task) error {
	if task.Status == "" {
		task.Status = "backlog"
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}
	if task.Type == "" {
		task.Type = "task"
	}
	if task.Labels == nil {
		task.Labels = []string{}
	}

	query := `
		INSERT INTO tasks (key, title, description, status, priority, type, project_id, sprint_id, 
		                   assignee_id, reporter_id, parent_id, story_points, due_date, order_index, labels)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		task.Key, task.Title, task.Description, task.Status, task.Priority, task.Type,
		task.ProjectID, task.SprintID, task.AssigneeID, task.ReporterID, task.ParentID,
		task.StoryPoints, task.DueDate, task.OrderIndex, task.Labels,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *pgTaskRepository) FindByID(ctx context.Context, id string) (*Task, error) {
	query := `
		SELECT t.id, t.key, t.title, t.description, t.status, t.priority, t.type,
		       t.project_id, t.sprint_id, t.assignee_id, t.reporter_id, t.parent_id,
		       t.story_points, t.due_date, t.order_index, t.labels, t.created_at, t.updated_at,
		       a.id, a.name, a.email, a.avatar,
		       rep.id, rep.name, rep.email, rep.avatar
		FROM tasks t
		LEFT JOIN users a ON t.assignee_id = a.id
		LEFT JOIN users rep ON t.reporter_id = rep.id
		WHERE t.id = $1
	`
	task := &Task{}
	var assigneeID, assigneeName, assigneeEmail, assigneeAvatar sql.NullString
	var reporterID, reporterName, reporterEmail, reporterAvatar sql.NullString

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.Key, &task.Title, &task.Description, &task.Status, &task.Priority, &task.Type,
		&task.ProjectID, &task.SprintID, &task.AssigneeID, &task.ReporterID, &task.ParentID,
		&task.StoryPoints, &task.DueDate, &task.OrderIndex, &task.Labels, &task.CreatedAt, &task.UpdatedAt,
		&assigneeID, &assigneeName, &assigneeEmail, &assigneeAvatar,
		&reporterID, &reporterName, &reporterEmail, &reporterAvatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if assigneeID.Valid {
		task.Assignee = &User{ID: assigneeID.String, Name: assigneeName.String, Email: assigneeEmail.String}
		if assigneeAvatar.Valid {
			task.Assignee.Avatar = &assigneeAvatar.String
		}
	}
	if reporterID.Valid {
		task.Reporter = &User{ID: reporterID.String, Name: reporterName.String, Email: reporterEmail.String}
		if reporterAvatar.Valid {
			task.Reporter.Avatar = &reporterAvatar.String
		}
	}

	return task, nil
}

func (r *pgTaskRepository) FindByKey(ctx context.Context, key string) (*Task, error) {
	query := `SELECT id FROM tasks WHERE UPPER(key) = UPPER($1)`
	var id string
	err := r.pool.QueryRow(ctx, query, key).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *pgTaskRepository) FindByProjectID(ctx context.Context, projectID string, filters *TaskFilters) ([]*Task, error) {
	query := `
		SELECT t.id, t.key, t.title, t.description, t.status, t.priority, t.type,
		       t.project_id, t.sprint_id, t.assignee_id, t.reporter_id, t.parent_id,
		       t.story_points, t.due_date, t.order_index, t.labels, t.created_at, t.updated_at,
		       a.id, a.name, a.email, a.avatar,
		       rep.id, rep.name, rep.email, rep.avatar
		FROM tasks t
		LEFT JOIN users a ON t.assignee_id = a.id
		LEFT JOIN users rep ON t.reporter_id = rep.id
		WHERE t.project_id = $1
	`
	args := []interface{}{projectID}
	argNum := 1

	if filters != nil {
		if len(filters.Status) > 0 {
			argNum++
			query += fmt.Sprintf(" AND t.status = ANY($%d)", argNum)
			args = append(args, filters.Status)
		}
		if len(filters.Priority) > 0 {
			argNum++
			query += fmt.Sprintf(" AND t.priority = ANY($%d)", argNum)
			args = append(args, filters.Priority)
		}
		if len(filters.Type) > 0 {
			argNum++
			query += fmt.Sprintf(" AND t.type = ANY($%d)", argNum)
			args = append(args, filters.Type)
		}
		if filters.SprintID != nil {
			argNum++
			query += fmt.Sprintf(" AND t.sprint_id = $%d", argNum)
			args = append(args, *filters.SprintID)
		}
		if filters.Search != "" {
			argNum++
			query += fmt.Sprintf(" AND (LOWER(t.title) LIKE LOWER($%d) OR LOWER(t.key) LIKE LOWER($%d))", argNum, argNum)
			args = append(args, "%"+filters.Search+"%")
		}
	}

	query += " ORDER BY t.order_index, t.created_at DESC"

	if filters != nil && filters.Limit > 0 {
		argNum++
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filters.Limit)
	}
	if filters != nil && filters.Offset > 0 {
		argNum++
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var assigneeID, assigneeName, assigneeEmail, assigneeAvatar sql.NullString
		var reporterID, reporterName, reporterEmail, reporterAvatar sql.NullString

		if err := rows.Scan(
			&task.ID, &task.Key, &task.Title, &task.Description, &task.Status, &task.Priority, &task.Type,
			&task.ProjectID, &task.SprintID, &task.AssigneeID, &task.ReporterID, &task.ParentID,
			&task.StoryPoints, &task.DueDate, &task.OrderIndex, &task.Labels, &task.CreatedAt, &task.UpdatedAt,
			&assigneeID, &assigneeName, &assigneeEmail, &assigneeAvatar,
			&reporterID, &reporterName, &reporterEmail, &reporterAvatar,
		); err != nil {
			return nil, err
		}

		if assigneeID.Valid {
			task.Assignee = &User{ID: assigneeID.String, Name: assigneeName.String, Email: assigneeEmail.String}
			if assigneeAvatar.Valid {
				task.Assignee.Avatar = &assigneeAvatar.String
			}
		}
		if reporterID.Valid {
			task.Reporter = &User{ID: reporterID.String, Name: reporterName.String, Email: reporterEmail.String}
			if reporterAvatar.Valid {
				task.Reporter.Avatar = &reporterAvatar.String
			}
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindBySprintID(ctx context.Context, sprintID string) ([]*Task, error) {
	query := `
		SELECT t.id, t.key, t.title, t.description, t.status, t.priority, t.type,
		       t.project_id, t.sprint_id, t.assignee_id, t.reporter_id, t.parent_id,
		       t.story_points, t.due_date, t.order_index, t.labels, t.created_at, t.updated_at,
		       a.id, a.name, a.email, a.avatar,
		       rep.id, rep.name, rep.email, rep.avatar
		FROM tasks t
		LEFT JOIN users a ON t.assignee_id = a.id
		LEFT JOIN users rep ON t.reporter_id = rep.id
		WHERE t.sprint_id = $1
		ORDER BY t.order_index
	`
	rows, err := r.pool.Query(ctx, query, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var assigneeID, assigneeName, assigneeEmail, assigneeAvatar sql.NullString
		var reporterID, reporterName, reporterEmail, reporterAvatar sql.NullString

		if err := rows.Scan(
			&task.ID, &task.Key, &task.Title, &task.Description, &task.Status, &task.Priority, &task.Type,
			&task.ProjectID, &task.SprintID, &task.AssigneeID, &task.ReporterID, &task.ParentID,
			&task.StoryPoints, &task.DueDate, &task.OrderIndex, &task.Labels, &task.CreatedAt, &task.UpdatedAt,
			&assigneeID, &assigneeName, &assigneeEmail, &assigneeAvatar,
			&reporterID, &reporterName, &reporterEmail, &reporterAvatar,
		); err != nil {
			return nil, err
		}

		if assigneeID.Valid {
			task.Assignee = &User{ID: assigneeID.String, Name: assigneeName.String, Email: assigneeEmail.String}
			if assigneeAvatar.Valid {
				task.Assignee.Avatar = &assigneeAvatar.String
			}
		}
		if reporterID.Valid {
			task.Reporter = &User{ID: reporterID.String, Name: reporterName.String, Email: reporterEmail.String}
			if reporterAvatar.Valid {
				task.Reporter.Avatar = &reporterAvatar.String
			}
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindBacklog(ctx context.Context, projectID string) ([]*Task, error) {
	query := `SELECT id FROM tasks WHERE project_id = $1 AND sprint_id IS NULL ORDER BY order_index`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		task, _ := r.FindByID(ctx, id)
		if task != nil {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindOverdue(ctx context.Context) ([]*Task, error) {
	query := `SELECT id FROM tasks WHERE due_date < NOW() AND status NOT IN ('done', 'cancelled')`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		task, _ := r.FindByID(ctx, id)
		if task != nil {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindDueSoon(ctx context.Context, within time.Duration) ([]*Task, error) {
	query := `SELECT id FROM tasks WHERE due_date > NOW() AND due_date < $1 AND status NOT IN ('done', 'cancelled')`
	deadline := time.Now().Add(within)
	rows, err := r.pool.Query(ctx, query, deadline)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		task, _ := r.FindByID(ctx, id)
		if task != nil {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindByAssignee(ctx context.Context, assigneeID string) ([]*Task, error) {
	query := `SELECT id FROM tasks WHERE assignee_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, assigneeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		task, _ := r.FindByID(ctx, id)
		if task != nil {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *pgTaskRepository) Update(ctx context.Context, task *Task) error {
	query := `
		UPDATE tasks SET 
			title = $2, description = $3, status = $4, priority = $5, type = $6,
			sprint_id = $7, assignee_id = $8, story_points = $9, due_date = $10,
			order_index = $11, labels = $12, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.Priority, task.Type,
		task.SprintID, task.AssigneeID, task.StoryPoints, task.DueDate,
		task.OrderIndex, task.Labels,
	)
	return err
}

func (r *pgTaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgTaskRepository) BulkUpdate(ctx context.Context, updates []BulkTaskUpdate) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, u := range updates {
		if u.Status != nil {
			_, err = tx.Exec(ctx, `UPDATE tasks SET status = $2, updated_at = NOW() WHERE id = $1`, u.ID, *u.Status)
			if err != nil {
				return err
			}
		}
		if u.SprintID != nil {
			_, err = tx.Exec(ctx, `UPDATE tasks SET sprint_id = $2, updated_at = NOW() WHERE id = $1`, u.ID, *u.SprintID)
			if err != nil {
				return err
			}
		}
		if u.OrderIndex != nil {
			_, err = tx.Exec(ctx, `UPDATE tasks SET order_index = $2, updated_at = NOW() WHERE id = $1`, u.ID, *u.OrderIndex)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *pgTaskRepository) CountBySprintID(ctx context.Context, sprintID string) (total int, completed int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'done') as completed
		FROM tasks WHERE sprint_id = $1
	`
	err = r.pool.QueryRow(ctx, query, sprintID).Scan(&total, &completed)
	return
}

// ============================================
// PostgreSQL Comment Repository
// ============================================

type pgCommentRepository struct {
	pool *pgxpool.Pool
}

func (r *pgCommentRepository) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (content, task_id, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query, comment.Content, comment.TaskID, comment.UserID).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
}

func (r *pgCommentRepository) FindByID(ctx context.Context, id string) (*Comment, error) {
	query := `
		SELECT c.id, c.content, c.task_id, c.user_id, c.created_at, c.updated_at,
		       u.id, u.name, u.email, u.avatar
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`
	c := &Comment{User: &User{}}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Content, &c.TaskID, &c.UserID, &c.CreatedAt, &c.UpdatedAt,
		&c.User.ID, &c.User.Name, &c.User.Email, &c.User.Avatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *pgCommentRepository) FindByTaskID(ctx context.Context, taskID string) ([]*Comment, error) {
	query := `
		SELECT c.id, c.content, c.task_id, c.user_id, c.created_at, c.updated_at,
		       u.id, u.name, u.email, u.avatar
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.task_id = $1
		ORDER BY c.created_at
	`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{User: &User{}}
		if err := rows.Scan(
			&c.ID, &c.Content, &c.TaskID, &c.UserID, &c.CreatedAt, &c.UpdatedAt,
			&c.User.ID, &c.User.Name, &c.User.Email, &c.User.Avatar,
		); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *pgCommentRepository) Update(ctx context.Context, comment *Comment) error {
	query := `UPDATE comments SET content = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, comment.ID, comment.Content)
	return err
}

func (r *pgCommentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// ============================================
// PostgreSQL Label Repository
// ============================================

type pgLabelRepository struct {
	pool *pgxpool.Pool
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

// ============================================
// PostgreSQL Notification Repository
// ============================================

type pgNotificationRepository struct {
	pool *pgxpool.Pool
}

func (r *pgNotificationRepository) Create(ctx context.Context, notification *Notification) error {
	dataJSON, _ := json.Marshal(notification.Data)
	if notification.Data == nil {
		dataJSON = []byte("{}")
	}
	query := `
		INSERT INTO notifications (user_id, type, title, message, read, data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		notification.UserID, notification.Type, notification.Title,
		notification.Message, notification.Read, dataJSON,
	).Scan(&notification.ID, &notification.CreatedAt)
}

func (r *pgNotificationRepository) FindByID(ctx context.Context, id string) (*Notification, error) {
	query := `SELECT id, user_id, type, title, message, read, data, created_at FROM notifications WHERE id = $1`
	n := &Notification{}
	var dataJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Read, &dataJSON, &n.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(dataJSON, &n.Data)
	return n, nil
}

func (r *pgNotificationRepository) FindByUserID(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, read, data, created_at 
		FROM notifications WHERE user_id = $1
	`
	if unreadOnly {
		query += ` AND read = FALSE`
	}
	query += ` ORDER BY created_at DESC LIMIT 100`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		n := &Notification{}
		var dataJSON []byte
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Read, &dataJSON, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(dataJSON, &n.Data)
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *pgNotificationRepository) CountByUserID(ctx context.Context, userID string) (total int, unread int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE read = FALSE) as unread
		FROM notifications WHERE user_id = $1
	`
	err = r.pool.QueryRow(ctx, query, userID).Scan(&total, &unread)
	return
}

func (r *pgNotificationRepository) MarkAsRead(ctx context.Context, id string) error {
	query := `UPDATE notifications SET read = TRUE WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgNotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notifications SET read = TRUE WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *pgNotificationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgNotificationRepository) DeleteAll(ctx context.Context, userID string) error {
	query := `DELETE FROM notifications WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *pgNotificationRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time, readOnly bool) (int, error) {
	query := `DELETE FROM notifications WHERE created_at < $1`
	if readOnly {
		query += ` AND read = TRUE`
	}
	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

// ============================================
// In-Memory Repository Implementations (Fallback)
// ============================================

// In-memory User Repository
type inMemoryUserRepository struct {
	users         map[string]*User
	refreshTokens map[string]*RefreshToken
}

func newInMemoryUserRepository() *inMemoryUserRepository {
	return &inMemoryUserRepository{
		users:         make(map[string]*User),
		refreshTokens: make(map[string]*RefreshToken),
	}
}

func (r *inMemoryUserRepository) Create(ctx context.Context, user *User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	now := time.Now()
	user.LastActiveAt = &now
	if user.Status == "" {
		user.Status = "online"
	}
	r.users[user.ID] = user
	return nil
}

func (r *inMemoryUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

func (r *inMemoryUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	for _, user := range r.users {
		if strings.EqualFold(user.Email, email) {
			return user, nil
		}
	}
	return nil, nil
}

func (r *inMemoryUserRepository) FindByName(ctx context.Context, name string) (*User, error) {
	nameLower := strings.ToLower(name)
	for _, user := range r.users {
		if strings.Contains(strings.ToLower(user.Name), nameLower) {
			return user, nil
		}
	}
	return nil, nil
}

func (r *inMemoryUserRepository) FindAll(ctx context.Context) ([]*User, error) {
	users := make([]*User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}

func (r *inMemoryUserRepository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

func (r *inMemoryUserRepository) UpdateLastActive(ctx context.Context, userID string) error {
	if user, ok := r.users[userID]; ok {
		now := time.Now()
		user.LastActiveAt = &now
		user.Status = "online"
	}
	return nil
}

func (r *inMemoryUserRepository) UpdateStatusForInactive(ctx context.Context, inactiveDuration time.Duration) error {
	threshold := time.Now().Add(-inactiveDuration)
	for _, user := range r.users {
		if user.LastActiveAt != nil && user.LastActiveAt.Before(threshold) && user.Status == "online" {
			user.Status = "away"
		}
	}
	return nil
}

func (r *inMemoryUserRepository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	token.ID = uuid.New().String()
	token.CreatedAt = time.Now()
	r.refreshTokens[token.Token] = token
	return nil
}

func (r *inMemoryUserRepository) FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	if rt, ok := r.refreshTokens[token]; ok {
		return rt, nil
	}
	return nil, nil
}

func (r *inMemoryUserRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	delete(r.refreshTokens, token)
	return nil
}

func (r *inMemoryUserRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	for token, rt := range r.refreshTokens {
		if rt.UserID == userID {
			delete(r.refreshTokens, token)
		}
	}
	return nil
}

// Remaining in-memory implementations (Workspace, Space, Project, Sprint, Task, Comment, Label, Notification)
// These are simplified stubs - use the PostgreSQL versions for production

type inMemoryWorkspaceRepository struct {
	workspaces map[string]*Workspace
	members    map[string][]*WorkspaceMember
}

func newInMemoryWorkspaceRepository() *inMemoryWorkspaceRepository {
	return &inMemoryWorkspaceRepository{
		workspaces: make(map[string]*Workspace),
		members:    make(map[string][]*WorkspaceMember),
	}
}

func (r *inMemoryWorkspaceRepository) Create(ctx context.Context, workspace *Workspace) error {
	workspace.ID = uuid.New().String()
	workspace.CreatedAt = time.Now()
	workspace.UpdatedAt = time.Now()
	r.workspaces[workspace.ID] = workspace
	return nil
}

func (r *inMemoryWorkspaceRepository) FindByID(ctx context.Context, id string) (*Workspace, error) {
	if ws, ok := r.workspaces[id]; ok {
		return ws, nil
	}
	return nil, nil
}

func (r *inMemoryWorkspaceRepository) FindByUserID(ctx context.Context, userID string) ([]*Workspace, error) {
	var result []*Workspace
	for wsID, members := range r.members {
		for _, m := range members {
			if m.UserID == userID {
				if ws, ok := r.workspaces[wsID]; ok {
					result = append(result, ws)
				}
				break
			}
		}
	}
	return result, nil
}

func (r *inMemoryWorkspaceRepository) Update(ctx context.Context, workspace *Workspace) error {
	workspace.UpdatedAt = time.Now()
	r.workspaces[workspace.ID] = workspace
	return nil
}

func (r *inMemoryWorkspaceRepository) Delete(ctx context.Context, id string) error {
	delete(r.workspaces, id)
	delete(r.members, id)
	return nil
}

func (r *inMemoryWorkspaceRepository) AddMember(ctx context.Context, member *WorkspaceMember) error {
	member.ID = uuid.New().String()
	member.JoinedAt = time.Now()
	r.members[member.WorkspaceID] = append(r.members[member.WorkspaceID], member)
	return nil
}

func (r *inMemoryWorkspaceRepository) FindMembers(ctx context.Context, workspaceID string) ([]*WorkspaceMember, error) {
	return r.members[workspaceID], nil
}

func (r *inMemoryWorkspaceRepository) FindMember(ctx context.Context, workspaceID, userID string) (*WorkspaceMember, error) {
	for _, m := range r.members[workspaceID] {
		if m.UserID == userID {
			return m, nil
		}
	}
	return nil, nil
}

func (r *inMemoryWorkspaceRepository) FindMemberUserIDs(ctx context.Context, workspaceID string) ([]string, error) {
	var userIDs []string
	for _, m := range r.members[workspaceID] {
		userIDs = append(userIDs, m.UserID)
	}
	return userIDs, nil
}

func (r *inMemoryWorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	for _, m := range r.members[workspaceID] {
		if m.UserID == userID {
			m.Role = role
			return nil
		}
	}
	return nil
}

func (r *inMemoryWorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	members := r.members[workspaceID]
	for i, m := range members {
		if m.UserID == userID {
			r.members[workspaceID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return nil
}

// Space in-memory
type inMemorySpaceRepository struct {
	spaces map[string]*Space
}

func newInMemorySpaceRepository() *inMemorySpaceRepository {
	return &inMemorySpaceRepository{spaces: make(map[string]*Space)}
}

func (r *inMemorySpaceRepository) Create(ctx context.Context, space *Space) error {
	space.ID = uuid.New().String()
	space.CreatedAt = time.Now()
	space.UpdatedAt = time.Now()
	r.spaces[space.ID] = space
	return nil
}

func (r *inMemorySpaceRepository) FindByID(ctx context.Context, id string) (*Space, error) {
	if s, ok := r.spaces[id]; ok {
		return s, nil
	}
	return nil, nil
}

func (r *inMemorySpaceRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Space, error) {
	var result []*Space
	for _, s := range r.spaces {
		if s.WorkspaceID == workspaceID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *inMemorySpaceRepository) Update(ctx context.Context, space *Space) error {
	space.UpdatedAt = time.Now()
	r.spaces[space.ID] = space
	return nil
}

func (r *inMemorySpaceRepository) Delete(ctx context.Context, id string) error {
	delete(r.spaces, id)
	return nil
}

// Project in-memory
type inMemoryProjectRepository struct {
	projects    map[string]*Project
	members     map[string][]*ProjectMember
	taskCounter map[string]int
}

func newInMemoryProjectRepository() *inMemoryProjectRepository {
	return &inMemoryProjectRepository{
		projects:    make(map[string]*Project),
		members:     make(map[string][]*ProjectMember),
		taskCounter: make(map[string]int),
	}
}

func (r *inMemoryProjectRepository) Create(ctx context.Context, project *Project) error {
	project.ID = uuid.New().String()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	r.projects[project.ID] = project
	r.taskCounter[project.ID] = 0
	return nil
}

func (r *inMemoryProjectRepository) FindByID(ctx context.Context, id string) (*Project, error) {
	if p, ok := r.projects[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (r *inMemoryProjectRepository) FindBySpaceID(ctx context.Context, spaceID string) ([]*Project, error) {
	var result []*Project
	for _, p := range r.projects {
		if p.SpaceID == spaceID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *inMemoryProjectRepository) FindByKey(ctx context.Context, key string) (*Project, error) {
	for _, p := range r.projects {
		if strings.EqualFold(p.Key, key) {
			return p, nil
		}
	}
	return nil, nil
}

func (r *inMemoryProjectRepository) Update(ctx context.Context, project *Project) error {
	project.UpdatedAt = time.Now()
	r.projects[project.ID] = project
	return nil
}

func (r *inMemoryProjectRepository) Delete(ctx context.Context, id string) error {
	delete(r.projects, id)
	delete(r.members, id)
	delete(r.taskCounter, id)
	return nil
}

func (r *inMemoryProjectRepository) AddMember(ctx context.Context, member *ProjectMember) error {
	member.ID = uuid.New().String()
	member.JoinedAt = time.Now()
	r.members[member.ProjectID] = append(r.members[member.ProjectID], member)
	return nil
}

func (r *inMemoryProjectRepository) FindMembers(ctx context.Context, projectID string) ([]*ProjectMember, error) {
	return r.members[projectID], nil
}

func (r *inMemoryProjectRepository) FindMemberUserIDs(ctx context.Context, projectID string) ([]string, error) {
	var userIDs []string
	for _, m := range r.members[projectID] {
		userIDs = append(userIDs, m.UserID)
	}
	return userIDs, nil
}

func (r *inMemoryProjectRepository) FindMember(ctx context.Context, projectID, userID string) (*ProjectMember, error) {
	for _, m := range r.members[projectID] {
		if m.UserID == userID {
			return m, nil
		}
	}
	return nil, nil
}

func (r *inMemoryProjectRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	members := r.members[projectID]
	for i, m := range members {
		if m.UserID == userID {
			r.members[projectID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *inMemoryProjectRepository) GetNextTaskNumber(ctx context.Context, projectID string) (int, error) {
	r.taskCounter[projectID]++
	return r.taskCounter[projectID], nil
}

// Sprint in-memory
type inMemorySprintRepository struct {
	sprints map[string]*Sprint
}

func newInMemorySprintRepository() *inMemorySprintRepository {
	return &inMemorySprintRepository{sprints: make(map[string]*Sprint)}
}

func (r *inMemorySprintRepository) Create(ctx context.Context, sprint *Sprint) error {
	sprint.ID = uuid.New().String()
	sprint.CreatedAt = time.Now()
	sprint.UpdatedAt = time.Now()
	if sprint.Status == "" {
		sprint.Status = "planning"
	}
	r.sprints[sprint.ID] = sprint
	return nil
}

func (r *inMemorySprintRepository) FindByID(ctx context.Context, id string) (*Sprint, error) {
	if s, ok := r.sprints[id]; ok {
		return s, nil
	}
	return nil, nil
}

func (r *inMemorySprintRepository) FindByProjectID(ctx context.Context, projectID string) ([]*Sprint, error) {
	var result []*Sprint
	for _, s := range r.sprints {
		if s.ProjectID == projectID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *inMemorySprintRepository) FindActive(ctx context.Context, projectID string) (*Sprint, error) {
	for _, s := range r.sprints {
		if s.ProjectID == projectID && s.Status == "active" {
			return s, nil
		}
	}
	return nil, nil
}

func (r *inMemorySprintRepository) FindEndingSoon(ctx context.Context, within time.Duration) ([]*Sprint, error) {
	var result []*Sprint
	now := time.Now()
	deadline := now.Add(within)
	for _, s := range r.sprints {
		if s.Status == "active" && s.EndDate != nil && s.EndDate.After(now) && s.EndDate.Before(deadline) {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *inMemorySprintRepository) FindExpired(ctx context.Context) ([]*Sprint, error) {
	var result []*Sprint
	now := time.Now()
	for _, s := range r.sprints {
		if s.Status == "active" && s.EndDate != nil && s.EndDate.Before(now) {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *inMemorySprintRepository) Update(ctx context.Context, sprint *Sprint) error {
	sprint.UpdatedAt = time.Now()
	r.sprints[sprint.ID] = sprint
	return nil
}

func (r *inMemorySprintRepository) Delete(ctx context.Context, id string) error {
	delete(r.sprints, id)
	return nil
}

// Task in-memory
type inMemoryTaskRepository struct {
	tasks map[string]*Task
}

func newInMemoryTaskRepository() *inMemoryTaskRepository {
	return &inMemoryTaskRepository{tasks: make(map[string]*Task)}
}

func (r *inMemoryTaskRepository) Create(ctx context.Context, task *Task) error {
	task.ID = uuid.New().String()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	if task.Status == "" {
		task.Status = "backlog"
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}
	if task.Type == "" {
		task.Type = "task"
	}
	if task.Labels == nil {
		task.Labels = []string{}
	}
	r.tasks[task.ID] = task
	return nil
}

func (r *inMemoryTaskRepository) FindByID(ctx context.Context, id string) (*Task, error) {
	if t, ok := r.tasks[id]; ok {
		return t, nil
	}
	return nil, nil
}

func (r *inMemoryTaskRepository) FindByKey(ctx context.Context, key string) (*Task, error) {
	for _, t := range r.tasks {
		if strings.EqualFold(t.Key, key) {
			return t, nil
		}
	}
	return nil, nil
}

func (r *inMemoryTaskRepository) FindByProjectID(ctx context.Context, projectID string, filters *TaskFilters) ([]*Task, error) {
	var result []*Task
	for _, t := range r.tasks {
		if t.ProjectID != projectID {
			continue
		}
		if filters != nil {
			if len(filters.Status) > 0 && !contains(filters.Status, t.Status) {
				continue
			}
			if len(filters.Priority) > 0 && !contains(filters.Priority, t.Priority) {
				continue
			}
			if len(filters.Type) > 0 && !contains(filters.Type, t.Type) {
				continue
			}
			if filters.Search != "" {
				searchLower := strings.ToLower(filters.Search)
				if !strings.Contains(strings.ToLower(t.Title), searchLower) && !strings.Contains(strings.ToLower(t.Key), searchLower) {
					continue
				}
			}
		}
		result = append(result, t)
	}
	return result, nil
}

func (r *inMemoryTaskRepository) FindBySprintID(ctx context.Context, sprintID string) ([]*Task, error) {
	var result []*Task
	for _, t := range r.tasks {
		if t.SprintID != nil && *t.SprintID == sprintID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *inMemoryTaskRepository) FindBacklog(ctx context.Context, projectID string) ([]*Task, error) {
	var result []*Task
	for _, t := range r.tasks {
		if t.ProjectID == projectID && t.SprintID == nil {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *inMemoryTaskRepository) FindOverdue(ctx context.Context) ([]*Task, error) {
	var result []*Task
	now := time.Now()
	for _, t := range r.tasks {
		if t.DueDate != nil && t.DueDate.Before(now) && t.Status != "done" && t.Status != "cancelled" {
			result = append(result, t)
		}
	}
	return result, nil
}
func (r *inMemoryTaskRepository) FindDueSoon(ctx context.Context, within time.Duration) ([]*Task, error) {
	var result []*Task
	now := time.Now()
	deadline := now.Add(within)
	for _, t := range r.tasks {
		if t.DueDate != nil && t.DueDate.After(now) && t.DueDate.Before(deadline) && t.Status != "done" && t.Status != "cancelled" {
			result = append(result, t)
		}
	}
	return result, nil
}
func (r *inMemoryTaskRepository) FindByAssignee(ctx context.Context, assigneeID string) ([]*Task, error) {
	var result []*Task
	for _, t := range r.tasks {
		if t.AssigneeID != nil && *t.AssigneeID == assigneeID {
			result = append(result, t)
		}
	}
	return result, nil
}
func (r *inMemoryTaskRepository) Update(ctx context.Context, task *Task) error {
	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	return nil
}
func (r *inMemoryTaskRepository) Delete(ctx context.Context, id string) error {
	delete(r.tasks, id)
	return nil
}
func (r *inMemoryTaskRepository) BulkUpdate(ctx context.Context, updates []BulkTaskUpdate) error {
	for _, u := range updates {
		if t, ok := r.tasks[u.ID]; ok {
			if u.Status != nil {
				t.Status = *u.Status
			}
			if u.SprintID != nil {
				t.SprintID = u.SprintID
			}
			if u.OrderIndex != nil {
				t.OrderIndex = *u.OrderIndex
			}
			t.UpdatedAt = time.Now()
		}
	}
	return nil
}
func (r *inMemoryTaskRepository) CountBySprintID(ctx context.Context, sprintID string) (total int, completed int, err error) {
	for _, t := range r.tasks {
		if t.SprintID != nil && *t.SprintID == sprintID {
			total++
			if t.Status == "done" {
				completed++
			}
		}
	}
	return total, completed, nil
}

// Comment in-memory
type inMemoryCommentRepository struct {
	comments map[string]*Comment
}

func newInMemoryCommentRepository() *inMemoryCommentRepository {
	return &inMemoryCommentRepository{comments: make(map[string]*Comment)}
}
func (r *inMemoryCommentRepository) Create(ctx context.Context, comment *Comment) error {
	comment.ID = uuid.New().String()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	r.comments[comment.ID] = comment
	return nil
}
func (r *inMemoryCommentRepository) FindByID(ctx context.Context, id string) (*Comment, error) {
	if c, ok := r.comments[id]; ok {
		return c, nil
	}
	return nil, nil
}
func (r *inMemoryCommentRepository) FindByTaskID(ctx context.Context, taskID string) ([]*Comment, error) {
	var result []*Comment
	for _, c := range r.comments {
		if c.TaskID == taskID {
			result = append(result, c)
		}
	}
	return result, nil
}
func (r *inMemoryCommentRepository) Update(ctx context.Context, comment *Comment) error {
	comment.UpdatedAt = time.Now()
	r.comments[comment.ID] = comment
	return nil
}
func (r *inMemoryCommentRepository) Delete(ctx context.Context, id string) error {
	delete(r.comments, id)
	return nil
}

// Label in-memory
type inMemoryLabelRepository struct {
	labels map[string]*Label
}

func newInMemoryLabelRepository() *inMemoryLabelRepository {
	return &inMemoryLabelRepository{labels: make(map[string]*Label)}
}
func (r *inMemoryLabelRepository) Create(ctx context.Context, label *Label) error {
	label.ID = uuid.New().String()
	label.CreatedAt = time.Now()
	r.labels[label.ID] = label
	return nil
}
func (r *inMemoryLabelRepository) FindByID(ctx context.Context, id string) (*Label, error) {
	if l, ok := r.labels[id]; ok {
		return l, nil
	}
	return nil, nil
}
func (r *inMemoryLabelRepository) FindByProjectID(ctx context.Context, projectID string) ([]*Label, error) {
	var result []*Label
	for _, l := range r.labels {
		if l.ProjectID == projectID {
			result = append(result, l)
		}
	}
	return result, nil
}
func (r *inMemoryLabelRepository) FindByName(ctx context.Context, projectID, name string) (*Label, error) {
	for _, l := range r.labels {
		if l.ProjectID == projectID && strings.EqualFold(l.Name, name) {
			return l, nil
		}
	}
	return nil, nil
}
func (r *inMemoryLabelRepository) Update(ctx context.Context, label *Label) error {
	r.labels[label.ID] = label
	return nil
}
func (r *inMemoryLabelRepository) Delete(ctx context.Context, id string) error {
	delete(r.labels, id)
	return nil
}

// Notification in-memory
type inMemoryNotificationRepository struct {
	notifications map[string]*Notification
}

func newInMemoryNotificationRepository() *inMemoryNotificationRepository {
	return &inMemoryNotificationRepository{notifications: make(map[string]*Notification)}
}
func (r *inMemoryNotificationRepository) Create(ctx context.Context, notification *Notification) error {
	notification.ID = uuid.New().String()
	notification.CreatedAt = time.Now()
	r.notifications[notification.ID] = notification
	return nil
}
func (r *inMemoryNotificationRepository) FindByID(ctx context.Context, id string) (*Notification, error) {
	if n, ok := r.notifications[id]; ok {
		return n, nil
	}
	return nil, nil
}
func (r *inMemoryNotificationRepository) FindByUserID(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error) {
	var result []*Notification
	for _, n := range r.notifications {
		if n.UserID == userID {
			if unreadOnly && n.Read {
				continue
			}
			result = append(result, n)
		}
	}
	return result, nil
}
func (r *inMemoryNotificationRepository) CountByUserID(ctx context.Context, userID string) (total int, unread int, err error) {
	for _, n := range r.notifications {
		if n.UserID == userID {
			total++
			if !n.Read {
				unread++
			}
		}
	}
	return total, unread, nil
}
func (r *inMemoryNotificationRepository) MarkAsRead(ctx context.Context, id string) error {
	if n, ok := r.notifications[id]; ok {
		n.Read = true
	}
	return nil
}
func (r *inMemoryNotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	for _, n := range r.notifications {
		if n.UserID == userID {
			n.Read = true
		}
	}
	return nil
}
func (r *inMemoryNotificationRepository) Delete(ctx context.Context, id string) error {
	delete(r.notifications, id)
	return nil
}
func (r *inMemoryNotificationRepository) DeleteAll(ctx context.Context, userID string) error {
	for id, n := range r.notifications {
		if n.UserID == userID {
			delete(r.notifications, id)
		}
	}
	return nil
}
func (r *inMemoryNotificationRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time, readOnly bool) (int, error) {
	deleted := 0
	for id, n := range r.notifications {
		if n.CreatedAt.Before(olderThan) {
			if readOnly && !n.Read {
				continue
			}
			delete(r.notifications, id)
			deleted++
		}
	}
	return deleted, nil
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
