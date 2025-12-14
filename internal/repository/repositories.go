package repository

import "github.com/jackc/pgx/v5/pgxpool"

// Repositories contains all repository instances
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
	TeamRepo         TeamRepository
	InvitationRepo   InvitationRepository
	ActivityRepo     ActivityRepository
	TaskWatcherRepo  TaskWatcherRepository
}

// NewRepositories creates all PostgreSQL-backed repositories
func NewRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepo:         NewUserRepository(pool),
		WorkspaceRepo:    NewWorkspaceRepository(pool),
		SpaceRepo:        NewSpaceRepository(pool),
		ProjectRepo:      NewProjectRepository(pool),
		SprintRepo:       NewSprintRepository(pool),
		TaskRepo:         NewTaskRepository(pool),
		CommentRepo:      NewCommentRepository(pool),
		LabelRepo:        NewLabelRepository(pool),
		NotificationRepo: NewNotificationRepository(pool),
		TeamRepo:         NewTeamRepository(pool),
		InvitationRepo:   NewInvitationRepository(pool),
		ActivityRepo:     NewActivityRepository(pool),
		TaskWatcherRepo:  NewTaskWatcherRepository(pool),
	}
}
