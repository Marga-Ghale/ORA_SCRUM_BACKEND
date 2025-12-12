package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ============================================
// Mock Implementations (Replace with Prisma)
// ============================================

// Repositories struct holds all repository implementations
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

func NewRepositories() *Repositories {
	return &Repositories{
		UserRepo:         NewUserRepository(),
		WorkspaceRepo:    NewWorkspaceRepository(),
		SpaceRepo:        NewSpaceRepository(),
		ProjectRepo:      NewProjectRepository(),
		SprintRepo:       NewSprintRepository(),
		TaskRepo:         NewTaskRepository(),
		CommentRepo:      NewCommentRepository(),
		LabelRepo:        NewLabelRepository(),
		NotificationRepo: NewNotificationRepository(),
	}
}

// ============================================
// User Repository
// ============================================

type User struct {
	ID        string
	Email     string
	Password  string
	Name      string
	Avatar    *string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RefreshToken struct {
	ID        string
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteUserRefreshTokens(ctx context.Context, userID string) error
}

type userRepository struct {
	users         map[string]*User
	refreshTokens map[string]*RefreshToken
}

func NewUserRepository() UserRepository {
	return &userRepository{
		users:         make(map[string]*User),
		refreshTokens: make(map[string]*RefreshToken),
	}
}

func (r *userRepository) Create(ctx context.Context, user *User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*User, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (r *userRepository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

func (r *userRepository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	token.ID = uuid.New().String()
	token.CreatedAt = time.Now()
	r.refreshTokens[token.Token] = token
	return nil
}

func (r *userRepository) FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	if rt, ok := r.refreshTokens[token]; ok {
		return rt, nil
	}
	return nil, nil
}

func (r *userRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	delete(r.refreshTokens, token)
	return nil
}

func (r *userRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	for token, rt := range r.refreshTokens {
		if rt.UserID == userID {
			delete(r.refreshTokens, token)
		}
	}
	return nil
}

// ============================================
// Workspace Repository
// ============================================

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

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	FindByID(ctx context.Context, id string) (*Workspace, error)
	FindByUserID(ctx context.Context, userID string) ([]*Workspace, error)
	Update(ctx context.Context, workspace *Workspace) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *WorkspaceMember) error
	FindMembers(ctx context.Context, workspaceID string) ([]*WorkspaceMember, error)
	FindMember(ctx context.Context, workspaceID, userID string) (*WorkspaceMember, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
}

type workspaceRepository struct {
	workspaces map[string]*Workspace
	members    map[string][]*WorkspaceMember
}

func NewWorkspaceRepository() WorkspaceRepository {
	return &workspaceRepository{
		workspaces: make(map[string]*Workspace),
		members:    make(map[string][]*WorkspaceMember),
	}
}

func (r *workspaceRepository) Create(ctx context.Context, workspace *Workspace) error {
	workspace.ID = uuid.New().String()
	workspace.CreatedAt = time.Now()
	workspace.UpdatedAt = time.Now()
	r.workspaces[workspace.ID] = workspace
	return nil
}

func (r *workspaceRepository) FindByID(ctx context.Context, id string) (*Workspace, error) {
	if ws, ok := r.workspaces[id]; ok {
		return ws, nil
	}
	return nil, nil
}

func (r *workspaceRepository) FindByUserID(ctx context.Context, userID string) ([]*Workspace, error) {
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

func (r *workspaceRepository) Update(ctx context.Context, workspace *Workspace) error {
	workspace.UpdatedAt = time.Now()
	r.workspaces[workspace.ID] = workspace
	return nil
}

func (r *workspaceRepository) Delete(ctx context.Context, id string) error {
	delete(r.workspaces, id)
	delete(r.members, id)
	return nil
}

func (r *workspaceRepository) AddMember(ctx context.Context, member *WorkspaceMember) error {
	member.ID = uuid.New().String()
	member.JoinedAt = time.Now()
	r.members[member.WorkspaceID] = append(r.members[member.WorkspaceID], member)
	return nil
}

func (r *workspaceRepository) FindMembers(ctx context.Context, workspaceID string) ([]*WorkspaceMember, error) {
	return r.members[workspaceID], nil
}

func (r *workspaceRepository) FindMember(ctx context.Context, workspaceID, userID string) (*WorkspaceMember, error) {
	for _, m := range r.members[workspaceID] {
		if m.UserID == userID {
			return m, nil
		}
	}
	return nil, nil
}

func (r *workspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	for _, m := range r.members[workspaceID] {
		if m.UserID == userID {
			m.Role = role
			return nil
		}
	}
	return nil
}

func (r *workspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	members := r.members[workspaceID]
	for i, m := range members {
		if m.UserID == userID {
			r.members[workspaceID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return nil
}

// ============================================
// Space Repository
// ============================================

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

type SpaceRepository interface {
	Create(ctx context.Context, space *Space) error
	FindByID(ctx context.Context, id string) (*Space, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Space, error)
	Update(ctx context.Context, space *Space) error
	Delete(ctx context.Context, id string) error
}

type spaceRepository struct {
	spaces map[string]*Space
}

func NewSpaceRepository() SpaceRepository {
	return &spaceRepository{
		spaces: make(map[string]*Space),
	}
}

func (r *spaceRepository) Create(ctx context.Context, space *Space) error {
	space.ID = uuid.New().String()
	space.CreatedAt = time.Now()
	space.UpdatedAt = time.Now()
	r.spaces[space.ID] = space
	return nil
}

func (r *spaceRepository) FindByID(ctx context.Context, id string) (*Space, error) {
	if s, ok := r.spaces[id]; ok {
		return s, nil
	}
	return nil, nil
}

func (r *spaceRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*Space, error) {
	var result []*Space
	for _, s := range r.spaces {
		if s.WorkspaceID == workspaceID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *spaceRepository) Update(ctx context.Context, space *Space) error {
	space.UpdatedAt = time.Now()
	r.spaces[space.ID] = space
	return nil
}

func (r *spaceRepository) Delete(ctx context.Context, id string) error {
	delete(r.spaces, id)
	return nil
}

// ============================================
// Project Repository
// ============================================

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
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *ProjectMember) error
	FindMembers(ctx context.Context, projectID string) ([]*ProjectMember, error)
	RemoveMember(ctx context.Context, projectID, userID string) error
	GetNextTaskNumber(ctx context.Context, projectID string) (int, error)
}

type projectRepository struct {
	projects    map[string]*Project
	members     map[string][]*ProjectMember
	taskCounter map[string]int
}

func NewProjectRepository() ProjectRepository {
	return &projectRepository{
		projects:    make(map[string]*Project),
		members:     make(map[string][]*ProjectMember),
		taskCounter: make(map[string]int),
	}
}

func (r *projectRepository) Create(ctx context.Context, project *Project) error {
	project.ID = uuid.New().String()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	r.projects[project.ID] = project
	r.taskCounter[project.ID] = 0
	return nil
}

func (r *projectRepository) FindByID(ctx context.Context, id string) (*Project, error) {
	if p, ok := r.projects[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (r *projectRepository) FindBySpaceID(ctx context.Context, spaceID string) ([]*Project, error) {
	var result []*Project
	for _, p := range r.projects {
		if p.SpaceID == spaceID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *projectRepository) Update(ctx context.Context, project *Project) error {
	project.UpdatedAt = time.Now()
	r.projects[project.ID] = project
	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id string) error {
	delete(r.projects, id)
	delete(r.members, id)
	return nil
}

func (r *projectRepository) AddMember(ctx context.Context, member *ProjectMember) error {
	member.ID = uuid.New().String()
	member.JoinedAt = time.Now()
	r.members[member.ProjectID] = append(r.members[member.ProjectID], member)
	return nil
}

func (r *projectRepository) FindMembers(ctx context.Context, projectID string) ([]*ProjectMember, error) {
	return r.members[projectID], nil
}

func (r *projectRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	members := r.members[projectID]
	for i, m := range members {
		if m.UserID == userID {
			r.members[projectID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *projectRepository) GetNextTaskNumber(ctx context.Context, projectID string) (int, error) {
	r.taskCounter[projectID]++
	return r.taskCounter[projectID], nil
}

// ============================================
// Sprint Repository
// ============================================

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

type SprintRepository interface {
	Create(ctx context.Context, sprint *Sprint) error
	FindByID(ctx context.Context, id string) (*Sprint, error)
	FindByProjectID(ctx context.Context, projectID string) ([]*Sprint, error)
	FindActive(ctx context.Context, projectID string) (*Sprint, error)
	Update(ctx context.Context, sprint *Sprint) error
	Delete(ctx context.Context, id string) error
}

type sprintRepository struct {
	sprints map[string]*Sprint
}

func NewSprintRepository() SprintRepository {
	return &sprintRepository{
		sprints: make(map[string]*Sprint),
	}
}

func (r *sprintRepository) Create(ctx context.Context, sprint *Sprint) error {
	sprint.ID = uuid.New().String()
	sprint.CreatedAt = time.Now()
	sprint.UpdatedAt = time.Now()
	r.sprints[sprint.ID] = sprint
	return nil
}

func (r *sprintRepository) FindByID(ctx context.Context, id string) (*Sprint, error) {
	if s, ok := r.sprints[id]; ok {
		return s, nil
	}
	return nil, nil
}

func (r *sprintRepository) FindByProjectID(ctx context.Context, projectID string) ([]*Sprint, error) {
	var result []*Sprint
	for _, s := range r.sprints {
		if s.ProjectID == projectID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *sprintRepository) FindActive(ctx context.Context, projectID string) (*Sprint, error) {
	for _, s := range r.sprints {
		if s.ProjectID == projectID && s.Status == "ACTIVE" {
			return s, nil
		}
	}
	return nil, nil
}

func (r *sprintRepository) Update(ctx context.Context, sprint *Sprint) error {
	sprint.UpdatedAt = time.Now()
	r.sprints[sprint.ID] = sprint
	return nil
}

func (r *sprintRepository) Delete(ctx context.Context, id string) error {
	delete(r.sprints, id)
	return nil
}

// ============================================
// Task Repository
// ============================================

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

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindByProjectID(ctx context.Context, projectID string, filters *TaskFilters) ([]*Task, error)
	FindBySprintID(ctx context.Context, sprintID string) ([]*Task, error)
	FindBacklog(ctx context.Context, projectID string) ([]*Task, error)
	FindOverdue(ctx context.Context) ([]*Task, error)
	FindDueSoon(ctx context.Context, within time.Duration) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
	BulkUpdate(ctx context.Context, updates []BulkTaskUpdate) error
}

type BulkTaskUpdate struct {
	ID         string
	Status     *string
	SprintID   *string
	OrderIndex *int
}

type taskRepository struct {
	tasks map[string]*Task
}

func NewTaskRepository() TaskRepository {
	return &taskRepository{
		tasks: make(map[string]*Task),
	}
}

func (r *taskRepository) Create(ctx context.Context, task *Task) error {
	task.ID = uuid.New().String()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	return nil
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*Task, error) {
	if t, ok := r.tasks[id]; ok {
		return t, nil
	}
	return nil, nil
}

func (r *taskRepository) FindByProjectID(ctx context.Context, projectID string, filters *TaskFilters) ([]*Task, error) {
	var result []*Task
	for _, t := range r.tasks {
		if t.ProjectID == projectID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *taskRepository) FindBySprintID(ctx context.Context, sprintID string) ([]*Task, error) {
	var result []*Task
	for _, t := range r.tasks {
		if t.SprintID != nil && *t.SprintID == sprintID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *taskRepository) FindBacklog(ctx context.Context, projectID string) ([]*Task, error) {
	var result []*Task
	for _, t := range r.tasks {
		if t.ProjectID == projectID && t.SprintID == nil {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *taskRepository) FindOverdue(ctx context.Context) ([]*Task, error) {
	var result []*Task
	now := time.Now()
	for _, t := range r.tasks {
		if t.DueDate != nil && t.DueDate.Before(now) && t.Status != "DONE" {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *taskRepository) FindDueSoon(ctx context.Context, within time.Duration) ([]*Task, error) {
	var result []*Task
	now := time.Now()
	deadline := now.Add(within)
	for _, t := range r.tasks {
		if t.DueDate != nil && t.DueDate.After(now) && t.DueDate.Before(deadline) && t.Status != "DONE" {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *taskRepository) Update(ctx context.Context, task *Task) error {
	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	return nil
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	delete(r.tasks, id)
	return nil
}

func (r *taskRepository) BulkUpdate(ctx context.Context, updates []BulkTaskUpdate) error {
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

// ============================================
// Comment Repository
// ============================================

type Comment struct {
	ID        string
	Content   string
	TaskID    string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	User      *User
}

type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) error
	FindByID(ctx context.Context, id string) (*Comment, error)
	FindByTaskID(ctx context.Context, taskID string) ([]*Comment, error)
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id string) error
}

type commentRepository struct {
	comments map[string]*Comment
}

func NewCommentRepository() CommentRepository {
	return &commentRepository{
		comments: make(map[string]*Comment),
	}
}

func (r *commentRepository) Create(ctx context.Context, comment *Comment) error {
	comment.ID = uuid.New().String()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	r.comments[comment.ID] = comment
	return nil
}

func (r *commentRepository) FindByID(ctx context.Context, id string) (*Comment, error) {
	if c, ok := r.comments[id]; ok {
		return c, nil
	}
	return nil, nil
}

func (r *commentRepository) FindByTaskID(ctx context.Context, taskID string) ([]*Comment, error) {
	var result []*Comment
	for _, c := range r.comments {
		if c.TaskID == taskID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *Comment) error {
	comment.UpdatedAt = time.Now()
	r.comments[comment.ID] = comment
	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id string) error {
	delete(r.comments, id)
	return nil
}

// ============================================
// Label Repository
// ============================================

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
	Update(ctx context.Context, label *Label) error
	Delete(ctx context.Context, id string) error
}

type labelRepository struct {
	labels map[string]*Label
}

func NewLabelRepository() LabelRepository {
	return &labelRepository{
		labels: make(map[string]*Label),
	}
}

func (r *labelRepository) Create(ctx context.Context, label *Label) error {
	label.ID = uuid.New().String()
	label.CreatedAt = time.Now()
	r.labels[label.ID] = label
	return nil
}

func (r *labelRepository) FindByID(ctx context.Context, id string) (*Label, error) {
	if l, ok := r.labels[id]; ok {
		return l, nil
	}
	return nil, nil
}

func (r *labelRepository) FindByProjectID(ctx context.Context, projectID string) ([]*Label, error) {
	var result []*Label
	for _, l := range r.labels {
		if l.ProjectID == projectID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (r *labelRepository) Update(ctx context.Context, label *Label) error {
	r.labels[label.ID] = label
	return nil
}

func (r *labelRepository) Delete(ctx context.Context, id string) error {
	delete(r.labels, id)
	return nil
}

// ============================================
// Notification Repository
// ============================================

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

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	FindByID(ctx context.Context, id string) (*Notification, error)
	FindByUserID(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error)
	CountByUserID(ctx context.Context, userID string) (total int, unread int, err error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
	DeleteAll(ctx context.Context, userID string) error
}

type notificationRepository struct {
	notifications map[string]*Notification
}

func NewNotificationRepository() NotificationRepository {
	return &notificationRepository{
		notifications: make(map[string]*Notification),
	}
}

func (r *notificationRepository) Create(ctx context.Context, notification *Notification) error {
	notification.ID = uuid.New().String()
	notification.CreatedAt = time.Now()
	r.notifications[notification.ID] = notification
	return nil
}

func (r *notificationRepository) FindByID(ctx context.Context, id string) (*Notification, error) {
	if n, ok := r.notifications[id]; ok {
		return n, nil
	}
	return nil, nil
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error) {
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

func (r *notificationRepository) CountByUserID(ctx context.Context, userID string) (total int, unread int, err error) {
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

func (r *notificationRepository) MarkAsRead(ctx context.Context, id string) error {
	if n, ok := r.notifications[id]; ok {
		n.Read = true
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	for _, n := range r.notifications {
		if n.UserID == userID {
			n.Read = true
		}
	}
	return nil
}

func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	delete(r.notifications, id)
	return nil
}

func (r *notificationRepository) DeleteAll(ctx context.Context, userID string) error {
	for id, n := range r.notifications {
		if n.UserID == userID {
			delete(r.notifications, id)
		}
	}
	return nil
}
