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
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("resource already exists")
)

// ============================================
// Services Container
// ============================================

type Services struct {
	Auth         AuthService
	User         UserService
	Workspace    WorkspaceService
	Space        SpaceService
	Project      ProjectService
	Sprint       SprintService
	Task         TaskService
	Comment      CommentService
	Label        LabelService
	Notification NotificationService
}

func NewServices(cfg *config.Config, repos *repository.Repositories, notifSvc *notification.Service) *Services {
	return &Services{
		Auth:         NewAuthService(cfg, repos.UserRepo),
		User:         NewUserService(repos.UserRepo),
		Workspace:    NewWorkspaceService(repos.WorkspaceRepo, repos.UserRepo, notifSvc),
		Space:        NewSpaceService(repos.SpaceRepo),
		Project:      NewProjectService(repos.ProjectRepo, repos.UserRepo, notifSvc),
		Sprint:       NewSprintService(repos.SprintRepo, repos.TaskRepo, repos.ProjectRepo, notifSvc),
		Task:         NewTaskService(repos.TaskRepo, repos.ProjectRepo, repos.UserRepo, notifSvc),
		Comment:      NewCommentService(repos.CommentRepo, repos.TaskRepo, repos.UserRepo, notifSvc),
		Label:        NewLabelService(repos.LabelRepo),
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
	existingUser, _ := s.userRepo.FindByEmail(ctx, email)
	if existingUser != nil {
		return nil, "", "", ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &repository.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Status:   "online", // lowercase
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
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

	user.Status = "online" // lowercase
	s.userRepo.Update(ctx, user)
	s.userRepo.UpdateLastActive(ctx, user.ID)

	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user, accessToken, refreshToken, nil
}
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	rt, err := s.userRepo.FindRefreshToken(ctx, refreshToken)
	if err != nil || rt == nil {
		return "", "", ErrInvalidToken
	}

	if time.Now().After(rt.ExpiresAt) {
		s.userRepo.DeleteRefreshToken(ctx, refreshToken)
		return "", "", ErrInvalidToken
	}

	s.userRepo.DeleteRefreshToken(ctx, refreshToken)
	s.userRepo.UpdateLastActive(ctx, rt.UserID)

	accessToken, newRefreshToken, err := s.generateTokens(ctx, rt.UserID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate tokens: %w", err)
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
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * time.Duration(s.cfg.JWTExpiry)).Unix(),
		"iat": time.Now().Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

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
	GetByEmail(ctx context.Context, email string) (*repository.User, error)
	Update(ctx context.Context, id string, name, avatar *string) (*repository.User, error)
	UpdateLastActive(ctx context.Context, id string) error
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

func (s *userService) GetByEmail(ctx context.Context, email string) (*repository.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
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

func (s *userService) UpdateLastActive(ctx context.Context, id string) error {
	return s.userRepo.UpdateLastActive(ctx, id)
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
	AddMember(ctx context.Context, workspaceID, email, role, inviterID string) error
	ListMembers(ctx context.Context, workspaceID string) ([]*repository.WorkspaceMember, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
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

	member := &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        "OWNER",
	}
	if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

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

func (s *workspaceService) AddMember(ctx context.Context, workspaceID, email, role, inviterID string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	existing, _ := s.workspaceRepo.FindMember(ctx, workspaceID, user.ID)
	if existing != nil {
		return ErrConflict
	}

	member := &repository.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      user.ID,
		Role:        role,
	}

	if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
		return err
	}

	workspace, _ := s.workspaceRepo.FindByID(ctx, workspaceID)
	if workspace != nil && s.notifSvc != nil {
		inviterName := ""
		if inviter, _ := s.userRepo.FindByID(ctx, inviterID); inviter != nil {
			inviterName = inviter.Name
		}
		s.notifSvc.SendWorkspaceInvitation(ctx, user.ID, workspace.Name, workspaceID, inviterName)
	}

	return nil
}

func (s *workspaceService) ListMembers(ctx context.Context, workspaceID string) ([]*repository.WorkspaceMember, error) {
	members, err := s.workspaceRepo.FindMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

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

func (s *workspaceService) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	member, err := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}
	return member != nil, nil
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
	Create(ctx context.Context, spaceID, creatorID, name, key string, description, icon, color, leadID *string) (*repository.Project, error)
	GetByID(ctx context.Context, id string) (*repository.Project, error)
	ListBySpace(ctx context.Context, spaceID string) ([]*repository.Project, error)
	Update(ctx context.Context, id string, name, key, description, icon, color, leadID *string) (*repository.Project, error)
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, projectID, userID, role, inviterID string) error
	ListMembers(ctx context.Context, projectID string) ([]*repository.ProjectMember, error)
	GetMemberUserIDs(ctx context.Context, projectID string) ([]string, error)
	RemoveMember(ctx context.Context, projectID, userID string) error
	IsMember(ctx context.Context, projectID, userID string) (bool, error)
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

func (s *projectService) Create(ctx context.Context, spaceID, creatorID, name, key string, description, icon, color, leadID *string) (*repository.Project, error) {
	existing, _ := s.projectRepo.FindByKey(ctx, key)
	if existing != nil {
		return nil, ErrConflict
	}

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

	member := &repository.ProjectMember{
		ProjectID: project.ID,
		UserID:    creatorID,
		Role:      "LEAD",
	}
	s.projectRepo.AddMember(ctx, member)

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
		existing, _ := s.projectRepo.FindByKey(ctx, *key)
		if existing != nil && existing.ID != id {
			return nil, ErrConflict
		}
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

func (s *projectService) AddMember(ctx context.Context, projectID, userID, role, inviterID string) error {
	existing, _ := s.projectRepo.FindMember(ctx, projectID, userID)
	if existing != nil {
		return ErrConflict
	}

	member := &repository.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}

	if err := s.projectRepo.AddMember(ctx, member); err != nil {
		return err
	}

	project, _ := s.projectRepo.FindByID(ctx, projectID)
	if project != nil && s.notifSvc != nil {
		inviterName := ""
		if inviter, _ := s.userRepo.FindByID(ctx, inviterID); inviter != nil {
			inviterName = inviter.Name
		}
		s.notifSvc.SendProjectInvitation(ctx, userID, project.Name, projectID, inviterName)
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

func (s *projectService) GetMemberUserIDs(ctx context.Context, projectID string) ([]string, error) {
	return s.projectRepo.FindMemberUserIDs(ctx, projectID)
}

func (s *projectService) RemoveMember(ctx context.Context, projectID, userID string) error {
	return s.projectRepo.RemoveMember(ctx, projectID, userID)
}

func (s *projectService) IsMember(ctx context.Context, projectID, userID string) (bool, error) {
	member, err := s.projectRepo.FindMember(ctx, projectID, userID)
	if err != nil {
		return false, err
	}
	return member != nil, nil
}
