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
	ErrInvalidEntityType  = errors.New("invalid entity type")
	ErrInvalidInput       = errors.New("invalid input")
	ErrHasSubtasks        = errors.New("task has subtasks and cannot be deleted")
	ErrBadRequest         = errors.New("comment content is required")
	 ErrLastOwner = errors.New("cannot remove or demote the last owner")
)

// ============================================
// Services Container
// ============================================

type Services struct {
	Auth         AuthService
	User         UserService
	Folder       FolderService
	Workspace    WorkspaceService
	Space        SpaceService
	Project      ProjectService
	Task         TaskService
	Label        LabelService
	Notification NotificationService
	Team         TeamService
	Invitation   InvitationService
	Activity     ActivityService
	Chat         ChatService
	Permission   PermissionService
	Member       MemberService
	Broadcaster  *socket.Broadcaster
	notifiServicev1  *notification.Service
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
	// ✅ Create MemberService first (needed by other services)
	memberService := NewMemberService(
		deps.Repos.WorkspaceRepo,
		deps.Repos.SpaceRepo,
		deps.Repos.FolderRepo,
		deps.Repos.ProjectRepo,
		deps.Repos.UserRepo,
		deps.NotifSvc,
		deps.Broadcaster,
	)

	// ✅ Create PermissionService (needed by TaskService)
	permissionService := NewPermissionService(
		deps.Repos.WorkspaceRepo,
		deps.Repos.SpaceRepo,
		deps.Repos.ProjectRepo,
		deps.Repos.TaskRepo,
		deps.Repos.TeamRepo,
		deps.Repos.FolderRepo, // ✅ Added missing folderRepo
		memberService, // ✅ ADD THIS
	)

	return &Services{
		Auth:      NewAuthService(deps.Config, deps.Repos.UserRepo),
		User:      NewUserService(deps.Repos.UserRepo),
		Workspace: NewWorkspaceService(deps.Repos.WorkspaceRepo, deps.Repos.UserRepo, deps.NotifSvc,deps.Broadcaster,
),
		Space: NewSpaceService(
			deps.Repos.SpaceRepo,
			deps.Repos.WorkspaceRepo,
			memberService,
								deps.Broadcaster,

		),
		Folder: NewFolderService(
			deps.Repos.FolderRepo,
			deps.Repos.SpaceRepo,
			memberService,
								deps.Broadcaster,

		),
		Project: NewProjectService(
			deps.Repos.ProjectRepo,
			deps.Repos.SpaceRepo,
			deps.Repos.FolderRepo,
			memberService,
								deps.Broadcaster,

		),
		// ✅ CORRECTED TaskService with ALL required repos and services
		Task: NewTaskService(
					deps.Repos.TaskRepo,
					deps.Repos.TaskCommentRepo,
					deps.Repos.TaskAttachmentRepo,
					deps.Repos.TimeEntryRepo,
					deps.Repos.TaskDependencyRepo,
					deps.Repos.TaskChecklistRepo,
					deps.Repos.TaskActivityRepo,
					deps.Repos.ProjectRepo,
					deps.Repos.SprintRepo,
					deps.Repos.UserRepo,
					memberService,
					permissionService,
					deps.NotifSvc,   
					deps.Broadcaster,
				),
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
		Chat:        NewChatService(deps.Repos.ChatRepo, deps.Repos.UserRepo, deps.NotifSvc, deps.Broadcaster),
		Permission:  permissionService,
		Member:      memberService,
		Broadcaster: deps.Broadcaster,
	}
}