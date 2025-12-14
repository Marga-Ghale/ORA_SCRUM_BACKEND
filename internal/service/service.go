package service

import (
	"errors"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/config"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/email"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
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
	Team         TeamService
	Invitation   InvitationService
	Activity     ActivityService
	TaskWatcher  TaskWatcherService
	Chat         ChatService
	Permission   PermissionService
	Broadcaster  *socket.Broadcaster
}

// ServiceDeps contains all dependencies needed to create services
type ServiceDeps struct {
	Config      *config.Config
	Repos       *repository.Repositories
	NotifSvc    *notification.Service
	EmailSvc    *email.Service
	Broadcaster *socket.Broadcaster
}

func NewServices(deps *ServiceDeps) *Services {
	return &Services{
		Auth:         NewAuthService(deps.Config, deps.Repos.UserRepo),
		User:         NewUserService(deps.Repos.UserRepo),
		Workspace:    NewWorkspaceService(deps.Repos.WorkspaceRepo, deps.Repos.UserRepo, deps.NotifSvc),
		Space:        NewSpaceService(deps.Repos.SpaceRepo),
		Project:      NewProjectService(deps.Repos.ProjectRepo, deps.Repos.UserRepo, deps.NotifSvc),
		Sprint:       NewSprintService(deps.Repos.SprintRepo, deps.Repos.TaskRepo, deps.Repos.ProjectRepo, deps.NotifSvc),
		Task:         NewTaskService(deps.Repos.TaskRepo, deps.Repos.ProjectRepo, deps.Repos.UserRepo, deps.NotifSvc),
		Comment:      NewCommentService(deps.Repos.CommentRepo, deps.Repos.TaskRepo, deps.Repos.UserRepo, deps.NotifSvc),
		Label:        NewLabelService(deps.Repos.LabelRepo),
		Notification: NewNotificationService(deps.Repos.NotificationRepo),
		Team:         NewTeamService(deps.Repos.TeamRepo, deps.Repos.UserRepo, deps.Repos.WorkspaceRepo, deps.NotifSvc, deps.EmailSvc, deps.Broadcaster),
		Invitation: NewInvitationService(
			deps.Repos.InvitationRepo,
			deps.Repos.WorkspaceRepo,
			deps.Repos.TeamRepo,
			deps.Repos.ProjectRepo,
			deps.Repos.UserRepo,
			deps.Repos.SpaceRepo,
			deps.EmailSvc,
		),
		Activity:    NewActivityService(deps.Repos.ActivityRepo),
		TaskWatcher: NewTaskWatcherService(deps.Repos.TaskWatcherRepo),
		Chat:        NewChatService(deps.Repos.ChatRepo, deps.Repos.UserRepo, deps.NotifSvc, deps.Broadcaster),
		Permission:  NewPermissionService(deps.Repos.WorkspaceRepo, deps.Repos.SpaceRepo, deps.Repos.ProjectRepo, deps.Repos.TaskRepo, deps.Repos.TeamRepo),
		Broadcaster: deps.Broadcaster,
	}
}
