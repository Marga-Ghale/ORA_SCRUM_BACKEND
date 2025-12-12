package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/config"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrNotFound           = errors.New("resource not found")
	ErrUnauthorized       = errors.New("unauthorized")
)

// ============================================
// Services Container
// ============================================

type Services struct {
	Auth        AuthService
	User        UserService
	Workspace   WorkspaceService
	Space       SpaceService
	Project     ProjectService
	Sprint      SprintService
	Task        TaskService
	Comment     CommentService
	Label       LabelService
	Notification NotificationService
}

func NewServices(cfg *config.Config, repos *repository.Repositories, notifSvc *notification.Service) *Services {
	return &Services{
		Auth:        NewAuthService(cfg, repos.UserRepo),
		User:        NewUserService(repos.UserRepo),
		Workspace:   NewWorkspaceService(repos.WorkspaceRepo, repos.UserRepo, notifSvc),
		Space:       NewSpaceService(repos.SpaceRepo),
		Project:     NewProjectService(repos.ProjectRepo, repos.UserRepo, notifSvc),
		Sprint:      NewSprintService(repos.SprintRepo, repos.TaskRepo, notifSvc),
		Task:        NewTaskService(repos.TaskRepo, repos.ProjectRepo, repos.UserRepo, notifSvc),
		Comment:     NewCommentService(repos.CommentRepo, repos.TaskRepo, repos.UserRepo, notifSvc),
		Label:       NewLabelService(repos.LabelRepo),
		Notification: NewNotificationService(repos.NotificationRepo),
	}
}

// ============================================
// Auth Service
// ============================================

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*repository.User, string, string, error)
	Login(ctx context.Context, email, password string) (*repository.User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(token string) (*jwt.Token, error)
	GetUserIDFromToken(token *jwt.Token) (string, error)
}

type authService struct {
	cfg      *config.Config
	userRepo repository.UserRepository
}

func NewAuthService(cfg *config.Config, userRepo repository.UserRepository) AuthService {
	return &authService{cfg: cfg, userRepo: userRepo}
}

func (s *authService) Register(ctx context.Context, name, email, password string) (*repository.User, string, string, error) {
	// Check if user exists
	existingUser, _ := s.userRepo.FindByEmail(ctx, email)
	if existingUser != nil {
		return nil, "", "", ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	// Create user
	user := &repository.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Status:   "ONLINE",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", err
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*repository.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	// Update status
	user.Status = "ONLINE"
	s.userRepo.Update(ctx, user)

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// Verify refresh token exists
	rt, err := s.userRepo.FindRefreshToken(ctx, refreshToken)
	if err != nil || rt == nil {
		return "", "", ErrInvalidToken
	}

	// Check expiry
	if time.Now().After(rt.ExpiresAt) {
		s.userRepo.DeleteRefreshToken(ctx, refreshToken)
		return "", "", ErrInvalidToken
	}

	// Delete old token
	s.userRepo.DeleteRefreshToken(ctx, refreshToken)

	// Generate new tokens
	accessToken, newRefreshToken, err := s.generateTokens(ctx, rt.UserID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.userRepo.DeleteRefreshToken(ctx, refreshToken)
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *authService) GetUserIDFromToken(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}

func (s *authService) generateTokens(ctx context.Context, userID string) (string, string, error) {
	// Access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * time.Duration(s.cfg.JWTExpiry)).Unix(),
		"iat": time.Now().Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshTokenString := uuid.New().String()
	refreshTokenExpiry := time.Now().Add(time.Hour * 24 * time.Duration(s.cfg.RefreshExpiry))

	rt := &repository.RefreshToken{
		Token:     refreshTokenString,
		UserID:    userID,
		ExpiresAt: refreshTokenExpiry,
	}

	if err := s.userRepo.SaveRefreshToken(ctx, rt); err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// ============================================
// User Service
// ============================================

type UserService interface {
	GetByID(ctx context.Context, id string) (*repository.User, error)
	Update(ctx context.Context, id string, name, avatar *string) (*repository.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetByID(ctx context.Context, id string) (*repository.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, id string, name, avatar *string) (*repository.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if name != nil {
		user.Name = *name
	}
	if avatar != nil {
		user.Avatar = avatar
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ============================================
// Workspace Service
// ============================================

type WorkspaceService interface {
	Create(ctx context.Context, userID, name string, description, icon, color *string) (*repository.Workspace, error)
	GetByID(ctx context.Context, id string) (*repository.Workspace, error)
	List(ctx context.Context, userID string) ([]*repository.Workspace, error)
	Update(ctx context.Context, id string, name, description, icon, color *string) (*repository.Workspace, error)
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, workspaceID, email, role string) error
	ListMembers(ctx context.Context, workspaceID string) ([]*repository.WorkspaceMember, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
}

type workspaceService struct {
	workspaceRepo repository.WorkspaceRepository
	userRepo      repository.UserRepository
	notifSvc      *notification.Service
}

func NewWorkspaceService(workspaceRepo repository.WorkspaceRepository, userRepo repository.UserRepository, notifSvc *notification.Service) WorkspaceService {
	return &workspaceService{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		notifSvc:      notifSvc,
	}
}

func (s *workspaceService) Create(ctx context.Context, userID, name string, description, icon, color *string) (*repository.Workspace, error) {
	workspace := &repository.Workspace{
		Name:        name,
		Description: description,
		Icon:        icon,
		Color:       color,
		OwnerID:     userID,
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	// Add owner as member
	member := &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        "OWNER",
	}
	s.workspaceRepo.AddMember(ctx, member)

	return workspace, nil
}

func (s *workspaceService) GetByID(ctx context.Context, id string) (*repository.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrNotFound
	}
	return workspace, nil
}

func (s *workspaceService) List(ctx context.Context, userID string) ([]*repository.Workspace, error) {
	return s.workspaceRepo.FindByUserID(ctx, userID)
}

func (s *workspaceService) Update(ctx context.Context, id string, name, description, icon, color *string) (*repository.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, id)
	if err != nil || workspace == nil {
		return nil, ErrNotFound
	}

	if name != nil {
		workspace.Name = *name
	}
	if description != nil {
		workspace.Description = description
	}
	if icon != nil {
		workspace.Icon = icon
	}
	if color != nil {
		workspace.Color = color
	}

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (s *workspaceService) Delete(ctx context.Context, id string) error {
	return s.workspaceRepo.Delete(ctx, id)
}

func (s *workspaceService) AddMember(ctx context.Context, workspaceID, email, role string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	member := &repository.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      user.ID,
		Role:        role,
	}

	if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
		return err
	}

	// Send notification
	workspace, _ := s.workspaceRepo.FindByID(ctx, workspaceID)
	if workspace != nil {
		s.notifSvc.SendWorkspaceInvitation(ctx, user.ID, workspace.Name, workspaceID)
	}

	return nil
}

func (s *workspaceService) ListMembers(ctx context.Context, workspaceID string) ([]*repository.WorkspaceMember, error) {
	members, err := s.workspaceRepo.FindMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Populate user info
	for _, m := range members {
		user, _ := s.userRepo.FindByID(ctx, m.UserID)
		m.User = user
	}

	return members, nil
}

func (s *workspaceService) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	return s.workspaceRepo.UpdateMemberRole(ctx, workspaceID, userID, role)
}

func (s *workspaceService) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	return s.workspaceRepo.RemoveMember(ctx, workspaceID, userID)
}

// ============================================
// Space Service
// ============================================

type SpaceService interface {
	Create(ctx context.Context, workspaceID, name string, description, icon, color *string) (*repository.Space, error)
	GetByID(ctx context.Context, id string) (*repository.Space, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*repository.Space, error)
	Update(ctx context.Context, id string, name, description, icon, color *string) (*repository.Space, error)
	Delete(ctx context.Context, id string) error
}

type spaceService struct {
	spaceRepo repository.SpaceRepository
}

func NewSpaceService(spaceRepo repository.SpaceRepository) SpaceService {
	return &spaceService{spaceRepo: spaceRepo}
}

func (s *spaceService) Create(ctx context.Context, workspaceID, name string, description, icon, color *string) (*repository.Space, error) {
	space := &repository.Space{
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		Icon:        icon,
		Color:       color,
	}

	if err := s.spaceRepo.Create(ctx, space); err != nil {
		return nil, err
	}

	return space, nil
}

func (s *spaceService) GetByID(ctx context.Context, id string) (*repository.Space, error) {
	space, err := s.spaceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if space == nil {
		return nil, ErrNotFound
	}
	return space, nil
}

func (s *spaceService) ListByWorkspace(ctx context.Context, workspaceID string) ([]*repository.Space, error) {
	return s.spaceRepo.FindByWorkspaceID(ctx, workspaceID)
}

func (s *spaceService) Update(ctx context.Context, id string, name, description, icon, color *string) (*repository.Space, error) {
	space, err := s.spaceRepo.FindByID(ctx, id)
	if err != nil || space == nil {
		return nil, ErrNotFound
	}

	if name != nil {
		space.Name = *name
	}
	if description != nil {
		space.Description = description
	}
	if icon != nil {
		space.Icon = icon
	}
	if color != nil {
		space.Color = color
	}

	if err := s.spaceRepo.Update(ctx, space); err != nil {
		return nil, err
	}

	return space, nil
}

func (s *spaceService) Delete(ctx context.Context, id string) error {
	return s.spaceRepo.Delete(ctx, id)
}

// ============================================
// Project Service
// ============================================

type ProjectService interface {
	Create(ctx context.Context, spaceID, name, key string, description, icon, color, leadID *string) (*repository.Project, error)
	GetByID(ctx context.Context, id string) (*repository.Project, error)
	ListBySpace(ctx context.Context, spaceID string) ([]*repository.Project, error)
	Update(ctx context.Context, id string, name, key, description, icon, color, leadID *string) (*repository.Project, error)
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, projectID, userID, role string) error
	ListMembers(ctx context.Context, projectID string) ([]*repository.ProjectMember, error)
	RemoveMember(ctx context.Context, projectID, userID string) error
}

type projectService struct {
	projectRepo repository.ProjectRepository
	userRepo    repository.UserRepository
	notifSvc    *notification.Service
}

func NewProjectService(projectRepo repository.ProjectRepository, userRepo repository.UserRepository, notifSvc *notification.Service) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
		notifSvc:    notifSvc,
	}
}

func (s *projectService) Create(ctx context.Context, spaceID, name, key string, description, icon, color, leadID *string) (*repository.Project, error) {
	project := &repository.Project{
		SpaceID:     spaceID,
		Name:        name,
		Key:         key,
		Description: description,
		Icon:        icon,
		Color:       color,
		LeadID:      leadID,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) GetByID(ctx context.Context, id string) (*repository.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrNotFound
	}
	return project, nil
}

func (s *projectService) ListBySpace(ctx context.Context, spaceID string) ([]*repository.Project, error) {
	return s.projectRepo.FindBySpaceID(ctx, spaceID)
}

func (s *projectService) Update(ctx context.Context, id string, name, key, description, icon, color, leadID *string) (*repository.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil || project == nil {
		return nil, ErrNotFound
	}

	if name != nil {
		project.Name = *name
	}
	if key != nil {
		project.Key = *key
	}
	if description != nil {
		project.Description = description
	}
	if icon != nil {
		project.Icon = icon
	}
	if color != nil {
		project.Color = color
	}
	if leadID != nil {
		project.LeadID = leadID
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) Delete(ctx context.Context, id string) error {
	return s.projectRepo.Delete(ctx, id)
}

func (s *projectService) AddMember(ctx context.Context, projectID, userID, role string) error {
	member := &repository.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}

	if err := s.projectRepo.AddMember(ctx, member); err != nil {
		return err
	}

	// Send notification
	project, _ := s.projectRepo.FindByID(ctx, projectID)
	if project != nil {
		s.notifSvc.SendProjectInvitation(ctx, userID, project.Name, projectID)
	}

	return nil
}

func (s *projectService) ListMembers(ctx context.Context, projectID string) ([]*repository.ProjectMember, error) {
	members, err := s.projectRepo.FindMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		user, _ := s.userRepo.FindByID(ctx, m.UserID)
		m.User = user
	}

	return members, nil
}

func (s *projectService) RemoveMember(ctx context.Context, projectID, userID string) error {
	return s.projectRepo.RemoveMember(ctx, projectID, userID)
}

// ============================================
// Sprint Service
// ============================================

type SprintService interface {
	Create(ctx context.Context, projectID, name string, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error)
	GetByID(ctx context.Context, id string) (*repository.Sprint, error)
	ListByProject(ctx context.Context, projectID string) ([]*repository.Sprint, error)
	Update(ctx context.Context, id string, name, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error)
	Delete(ctx context.Context, id string) error
	Start(ctx context.Context, id string) (*repository.Sprint, error)
	Complete(ctx context.Context, id, moveIncomplete string) (*repository.Sprint, error)
}

type sprintService struct {
	sprintRepo repository.SprintRepository
	taskRepo   repository.TaskRepository
	notifSvc   *notification.Service
}

func NewSprintService(sprintRepo repository.SprintRepository, taskRepo repository.TaskRepository, notifSvc *notification.Service) SprintService {
	return &sprintService{
		sprintRepo: sprintRepo,
		taskRepo:   taskRepo,
		notifSvc:   notifSvc,
	}
}

func (s *sprintService) Create(ctx context.Context, projectID, name string, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error) {
	sprint := &repository.Sprint{
		ProjectID: projectID,
		Name:      name,
		Goal:      goal,
		Status:    "PLANNING",
		StartDate: startDate,
		EndDate:   endDate,
	}

	if err := s.sprintRepo.Create(ctx, sprint); err != nil {
		return nil, err
	}

	return sprint, nil
}

func (s *sprintService) GetByID(ctx context.Context, id string) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sprint == nil {
		return nil, ErrNotFound
	}
	return sprint, nil
}

func (s *sprintService) ListByProject(ctx context.Context, projectID string) ([]*repository.Sprint, error) {
	return s.sprintRepo.FindByProjectID(ctx, projectID)
}

func (s *sprintService) Update(ctx context.Context, id string, name, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	if name != nil {
		sprint.Name = *name
	}
	if goal != nil {
		sprint.Goal = goal
	}
	if startDate != nil {
		sprint.StartDate = startDate
	}
	if endDate != nil {
		sprint.EndDate = endDate
	}

	if err := s.sprintRepo.Update(ctx, sprint); err != nil {
		return nil, err
	}

	return sprint, nil
}

func (s *sprintService) Delete(ctx context.Context, id string) error {
	return s.sprintRepo.Delete(ctx, id)
}

func (s *sprintService) Start(ctx context.Context, id string) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	now := time.Now()
	sprint.Status = "ACTIVE"
	sprint.StartDate = &now

	if err := s.sprintRepo.Update(ctx, sprint); err != nil {
		return nil, err
	}

	// Send notifications to project members (would need project member list)
	s.notifSvc.SendSprintStarted(ctx, "", sprint.Name, sprint.ID) // UserID would come from project members

	return sprint, nil
}

func (s *sprintService) Complete(ctx context.Context, id, moveIncomplete string) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	now := time.Now()
	sprint.Status = "COMPLETED"
	sprint.EndDate = &now

	// Move incomplete tasks
	if moveIncomplete != "" {
		tasks, _ := s.taskRepo.FindBySprintID(ctx, id)
		for _, task := range tasks {
			if task.Status != "DONE" {
				if moveIncomplete == "backlog" {
					task.SprintID = nil
				} else {
					task.SprintID = &moveIncomplete
				}
				s.taskRepo.Update(ctx, task)
			}
		}
	}

	if err := s.sprintRepo.Update(ctx, sprint); err != nil {
		return nil, err
	}

	s.notifSvc.SendSprintCompleted(ctx, "", sprint.Name, sprint.ID)

	return sprint, nil
}

// ============================================
// Task Service
// ============================================

type TaskService interface {
	Create(ctx context.Context, projectID, reporterID, title string, description *string, status, priority, taskType *string, assigneeID, sprintID, parentID *string, storyPoints *int, dueDate *time.Time, labels []string) (*repository.Task, error)
	GetByID(ctx context.Context, id string) (*repository.Task, error)
	ListByProject(ctx context.Context, projectID string, filters *repository.TaskFilters) ([]*repository.Task, error)
	ListBySprint(ctx context.Context, sprintID string) ([]*repository.Task, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) (*repository.Task, error)
	Delete(ctx context.Context, id string) error
	BulkUpdate(ctx context.Context, updates []repository.BulkTaskUpdate) error
}

type taskService struct {
	taskRepo    repository.TaskRepository
	projectRepo repository.ProjectRepository
	userRepo    repository.UserRepository
	notifSvc    *notification.Service
}

func NewTaskService(taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository, userRepo repository.UserRepository, notifSvc *notification.Service) TaskService {
	return &taskService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
		notifSvc:    notifSvc,
	}
}

func (s *taskService) Create(ctx context.Context, projectID, reporterID, title string, description *string, status, priority, taskType *string, assigneeID, sprintID, parentID *string, storyPoints *int, dueDate *time.Time, labels []string) (*repository.Task, error) {
	// Get project for key generation
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil || project == nil {
		return nil, ErrNotFound
	}

	// Generate task key
	taskNum, _ := s.projectRepo.GetNextTaskNumber(ctx, projectID)
	taskKey := fmt.Sprintf("%s-%d", project.Key, taskNum)

	// Default values
	statusVal := "BACKLOG"
	if status != nil {
		statusVal = *status
	}
	priorityVal := "MEDIUM"
	if priority != nil {
		priorityVal = *priority
	}
	typeVal := "TASK"
	if taskType != nil {
		typeVal = *taskType
	}

	task := &repository.Task{
		Key:         taskKey,
		Title:       title,
		Description: description,
		Status:      statusVal,
		Priority:    priorityVal,
		Type:        typeVal,
		ProjectID:   projectID,
		SprintID:    sprintID,
		AssigneeID:  assigneeID,
		ReporterID:  reporterID,
		ParentID:    parentID,
		StoryPoints: storyPoints,
		DueDate:     dueDate,
		Labels:      labels,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// Send notification if assigned
	if assigneeID != nil && *assigneeID != reporterID {
		s.notifSvc.SendTaskAssigned(ctx, *assigneeID, task.Title, task.ID, projectID)
	}

	// Populate user info
	if task.AssigneeID != nil {
		task.Assignee, _ = s.userRepo.FindByID(ctx, *task.AssigneeID)
	}
	task.Reporter, _ = s.userRepo.FindByID(ctx, task.ReporterID)

	return task, nil
}

func (s *taskService) GetByID(ctx context.Context, id string) (*repository.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}

	// Populate user info
	if task.AssigneeID != nil {
		task.Assignee, _ = s.userRepo.FindByID(ctx, *task.AssigneeID)
	}
	task.Reporter, _ = s.userRepo.FindByID(ctx, task.ReporterID)

	return task, nil
}

func (s *taskService) ListByProject(ctx context.Context, projectID string, filters *repository.TaskFilters) ([]*repository.Task, error) {
	tasks, err := s.taskRepo.FindByProjectID(ctx, projectID, filters)
	if err != nil {
		return nil, err
	}

	// Populate user info
	for _, task := range tasks {
		if task.AssigneeID != nil {
			task.Assignee, _ = s.userRepo.FindByID(ctx, *task.AssigneeID)
		}
		task.Reporter, _ = s.userRepo.FindByID(ctx, task.ReporterID)
	}

	return tasks, nil
}

func (s *taskService) ListBySprint(ctx context.Context, sprintID string) ([]*repository.Task, error) {
	tasks, err := s.taskRepo.FindBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.AssigneeID != nil {
			task.Assignee, _ = s.userRepo.FindByID(ctx, *task.AssigneeID)
		}
		task.Reporter, _ = s.userRepo.FindByID(ctx, task.ReporterID)
	}

	return tasks, nil
}

func (s *taskService) Update(ctx context.Context, id string, updates map[string]interface{}) (*repository.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil || task == nil {
		return nil, ErrNotFound
	}

	oldAssigneeID := task.AssigneeID

	// Apply updates
	if v, ok := updates["title"].(string); ok {
		task.Title = v
	}
	if v, ok := updates["description"].(*string); ok {
		task.Description = v
	}
	if v, ok := updates["status"].(string); ok {
		task.Status = v
	}
	if v, ok := updates["priority"].(string); ok {
		task.Priority = v
	}
	if v, ok := updates["type"].(string); ok {
		task.Type = v
	}
	if v, ok := updates["assigneeId"].(*string); ok {
		task.AssigneeID = v
	}
	if v, ok := updates["sprintId"].(*string); ok {
		task.SprintID = v
	}
	if v, ok := updates["storyPoints"].(*int); ok {
		task.StoryPoints = v
	}
	if v, ok := updates["dueDate"].(*time.Time); ok {
		task.DueDate = v
	}
	if v, ok := updates["orderIndex"].(int); ok {
		task.OrderIndex = v
	}
	if v, ok := updates["labels"].([]string); ok {
		task.Labels = v
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// Send notification if assignee changed
	if task.AssigneeID != nil && (oldAssigneeID == nil || *oldAssigneeID != *task.AssigneeID) {
		s.notifSvc.SendTaskAssigned(ctx, *task.AssigneeID, task.Title, task.ID, task.ProjectID)
	}

	// Populate user info
	if task.AssigneeID != nil {
		task.Assignee, _ = s.userRepo.FindByID(ctx, *task.AssigneeID)
	}
	task.Reporter, _ = s.userRepo.FindByID(ctx, task.ReporterID)

	return task, nil
}

func (s *taskService) Delete(ctx context.Context, id string) error {
	return s.taskRepo.Delete(ctx, id)
}

func (s *taskService) BulkUpdate(ctx context.Context, updates []repository.BulkTaskUpdate) error {
	return s.taskRepo.BulkUpdate(ctx, updates)
}

// ============================================
// Comment Service
// ============================================

type CommentService interface {
	Create(ctx context.Context, taskID, userID, content string) (*repository.Comment, error)
	GetByID(ctx context.Context, id string) (*repository.Comment, error)
	ListByTask(ctx context.Context, taskID string) ([]*repository.Comment, error)
	Update(ctx context.Context, id, userID, content string) (*repository.Comment, error)
	Delete(ctx context.Context, id, userID string) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	taskRepo    repository.TaskRepository
	userRepo    repository.UserRepository
	notifSvc    *notification.Service
}

func NewCommentService(commentRepo repository.CommentRepository, taskRepo repository.TaskRepository, userRepo repository.UserRepository, notifSvc *notification.Service) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		taskRepo:    taskRepo,
		userRepo:    userRepo,
		notifSvc:    notifSvc,
	}
}

func (s *commentService) Create(ctx context.Context, taskID, userID, content string) (*repository.Comment, error) {
	comment := &repository.Comment{
		TaskID:  taskID,
		UserID:  userID,
		Content: content,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// Send notification to task assignee/reporter
	task, _ := s.taskRepo.FindByID(ctx, taskID)
	if task != nil {
		if task.AssigneeID != nil && *task.AssigneeID != userID {
			s.notifSvc.SendTaskCommented(ctx, *task.AssigneeID, task.Title, task.ID)
		}
		if task.ReporterID != userID && (task.AssigneeID == nil || *task.AssigneeID != task.ReporterID) {
			s.notifSvc.SendTaskCommented(ctx, task.ReporterID, task.Title, task.ID)
		}
	}

	// Populate user
	comment.User, _ = s.userRepo.FindByID(ctx, userID)

	return comment, nil
}

func (s *commentService) GetByID(ctx context.Context, id string) (*repository.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, ErrNotFound
	}
	comment.User, _ = s.userRepo.FindByID(ctx, comment.UserID)
	return comment, nil
}

func (s *commentService) ListByTask(ctx context.Context, taskID string) ([]*repository.Comment, error) {
	comments, err := s.commentRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	for _, c := range comments {
		c.User, _ = s.userRepo.FindByID(ctx, c.UserID)
	}

	return comments, nil
}

func (s *commentService) Update(ctx context.Context, id, userID, content string) (*repository.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil || comment == nil {
		return nil, ErrNotFound
	}

	if comment.UserID != userID {
		return nil, ErrUnauthorized
	}

	comment.Content = content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}

	comment.User, _ = s.userRepo.FindByID(ctx, comment.UserID)
	return comment, nil
}

func (s *commentService) Delete(ctx context.Context, id, userID string) error {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil || comment == nil {
		return ErrNotFound
	}

	if comment.UserID != userID {
		return ErrUnauthorized
	}

	return s.commentRepo.Delete(ctx, id)
}

// ============================================
// Label Service
// ============================================

type LabelService interface {
	Create(ctx context.Context, projectID, name, color string) (*repository.Label, error)
	GetByID(ctx context.Context, id string) (*repository.Label, error)
	ListByProject(ctx context.Context, projectID string) ([]*repository.Label, error)
	Update(ctx context.Context, id string, name, color *string) (*repository.Label, error)
	Delete(ctx context.Context, id string) error
}

type labelService struct {
	labelRepo repository.LabelRepository
}

func NewLabelService(labelRepo repository.LabelRepository) LabelService {
	return &labelService{labelRepo: labelRepo}
}

func (s *labelService) Create(ctx context.Context, projectID, name, color string) (*repository.Label, error) {
	label := &repository.Label{
		ProjectID: projectID,
		Name:      name,
		Color:     color,
	}

	if err := s.labelRepo.Create(ctx, label); err != nil {
		return nil, err
	}

	return label, nil
}

func (s *labelService) GetByID(ctx context.Context, id string) (*repository.Label, error) {
	label, err := s.labelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if label == nil {
		return nil, ErrNotFound
	}
	return label, nil
}

func (s *labelService) ListByProject(ctx context.Context, projectID string) ([]*repository.Label, error) {
	return s.labelRepo.FindByProjectID(ctx, projectID)
}

func (s *labelService) Update(ctx context.Context, id string, name, color *string) (*repository.Label, error) {
	label, err := s.labelRepo.FindByID(ctx, id)
	if err != nil || label == nil {
		return nil, ErrNotFound
	}

	if name != nil {
		label.Name = *name
	}
	if color != nil {
		label.Color = *color
	}

	if err := s.labelRepo.Update(ctx, label); err != nil {
		return nil, err
	}

	return label, nil
}

func (s *labelService) Delete(ctx context.Context, id string) error {
	return s.labelRepo.Delete(ctx, id)
}

// ============================================
// Notification Service (for handlers)
// ============================================

type NotificationService interface {
	List(ctx context.Context, userID string, unreadOnly bool) ([]*repository.Notification, error)
	Count(ctx context.Context, userID string) (total int, unread int, err error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
	DeleteAll(ctx context.Context, userID string) error
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
}

func NewNotificationService(notificationRepo repository.NotificationRepository) NotificationService {
	return &notificationService{notificationRepo: notificationRepo}
}

func (s *notificationService) List(ctx context.Context, userID string, unreadOnly bool) ([]*repository.Notification, error) {
	return s.notificationRepo.FindByUserID(ctx, userID, unreadOnly)
}

func (s *notificationService) Count(ctx context.Context, userID string) (total int, unread int, err error) {
	return s.notificationRepo.CountByUserID(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id string) error {
	return s.notificationRepo.MarkAsRead(ctx, id)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) Delete(ctx context.Context, id string) error {
	return s.notificationRepo.Delete(ctx, id)
}

func (s *notificationService) DeleteAll(ctx context.Context, userID string) error {
	return s.notificationRepo.DeleteAll(ctx, userID)
}
