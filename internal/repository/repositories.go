package repository

import (
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	// Core repositories (pgxpool)
	UserRepo         UserRepository
	WorkspaceRepo    WorkspaceRepository
	FolderRepo       FolderRepository
	SpaceRepo        SpaceRepository
	ProjectRepo      ProjectRepository
	TeamRepo         TeamRepository
	InvitationRepo   InvitationRepository
	ActivityRepo     ActivityRepository
	ChatRepo         ChatRepository
	LabelRepo        LabelRepository
	NotificationRepo NotificationRepository

	GoalRepo            GoalRepository
	SprintAnalyticsRepo SprintAnalyticsRepository

	// Task-related repositories (sql.DB)
	SprintRepo         SprintRepository
	TaskRepo           TaskRepository
	TaskDependencyRepo TaskDependencyRepository
	TaskAttachmentRepo TaskAttachmentRepository
	TaskChecklistRepo  TaskChecklistRepository
	TaskCommentRepo    TaskCommentRepository
	TaskActivityRepo   TaskActivityRepository
	TimeEntryRepo      TimeEntryRepository
	SprintCommitmentRepo SprintCommitmentRepository
}

func NewRepositories(pool *pgxpool.Pool, db *sql.DB) *Repositories {
	return &Repositories{
		// pgxpool repos
		UserRepo:         NewUserRepository(pool),
		WorkspaceRepo:    NewWorkspaceRepository(pool),
		FolderRepo:       NewFolderRepository(pool),
		SpaceRepo:        NewSpaceRepository(pool),
		ProjectRepo:      NewProjectRepository(pool),
		TeamRepo:         NewTeamRepository(pool),
		InvitationRepo:   NewInvitationRepository(pool),
		ActivityRepo:     NewActivityRepository(pool),
		ChatRepo:         NewChatRepository(pool),
		LabelRepo:        NewLabelRepository(pool),
		NotificationRepo: NewNotificationRepository(pool),

		// sql.DB repos (all task-related)
		SprintRepo:         NewSprintRepository(db),
		SprintAnalyticsRepo: NewSprintAnalyticsRepository(db),
		GoalRepo:         NewGoalRepository(db),
		TaskRepo:           NewTaskRepository(db),
		TaskDependencyRepo: NewTaskDependencyRepository(db),
		TaskAttachmentRepo: NewTaskAttachmentRepository(db),
		TaskChecklistRepo:  NewTaskChecklistRepository(db),
		TaskCommentRepo:    NewTaskCommentRepository(db),
		TaskActivityRepo:   NewTaskActivityRepository(db),
		TimeEntryRepo:      NewTimeEntryRepository(db),
		SprintCommitmentRepo: NewSprintCommitmentRepository(db),
	}
}