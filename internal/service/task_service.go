package service

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
)

type TaskService interface {
	// Task CRUD
	Create(ctx context.Context, req *models.CreateTaskRequest) (*repository.Task, error)
	GetByID(ctx context.Context, taskID, userID string) (*repository.Task, error)
	Update(ctx context.Context, taskID, userID string, req *models.UpdateTaskRequest) (*repository.Task, error)
	Delete(ctx context.Context, taskID, userID string) error
	
	// Listing
	ListByProject(ctx context.Context, projectID, userID string) ([]*repository.Task, error)
	ListBySprint(ctx context.Context, sprintID, userID string) ([]*repository.Task, error)
	ListSubtasks(ctx context.Context, parentTaskID, userID string) ([]*repository.Task, error)
	ListMyTasks(ctx context.Context, userID string) ([]*repository.Task, error)
	ListByStatus(ctx context.Context, projectID, status, userID string) ([]*repository.Task, error)
	
	// Task operations
	UpdateStatus(ctx context.Context, taskID, status, userID string) error
	UpdatePriority(ctx context.Context, taskID, priority, userID string) error
	AssignTask(ctx context.Context, taskID, assigneeID, actorID string) error
	UnassignTask(ctx context.Context, taskID, assigneeID, actorID string) error
	AddWatcher(ctx context.Context, taskID, watcherID, actorID string) error
	RemoveWatcher(ctx context.Context, taskID, watcherID, actorID string) error
	MarkComplete(ctx context.Context, taskID, userID string) error
	MoveToSprint(ctx context.Context, taskID, sprintID, userID string) error
	ConvertToSubtask(ctx context.Context, taskID, parentTaskID, userID string) error
	PromoteToTask(ctx context.Context, taskID, userID string) error

	// COMMENTS
	AddComment(ctx context.Context, taskID, userID, content string, mentionedUsers []string) (*repository.TaskComment, error)
	ListComments(ctx context.Context, taskID, userID string) ([]*repository.TaskComment, error)
	UpdateComment(ctx context.Context, commentID, userID, content string) error
	DeleteComment(ctx context.Context, commentID, userID string) error
	
	// ATTACHMENTS
	AddAttachment(ctx context.Context, taskID, userID, filename, fileURL string, fileSize int64, mimeType string) (*repository.TaskAttachment, error)
	ListAttachments(ctx context.Context, taskID, userID string) ([]*repository.TaskAttachment, error)
	DeleteAttachment(ctx context.Context, attachmentID, userID string) error
	
	// TIME TRACKING
	StartTimer(ctx context.Context, taskID, userID string) (*repository.TimeEntry, error)
	StopTimer(ctx context.Context, userID string) (*repository.TimeEntry, error)
	GetActiveTimer(ctx context.Context, userID string) (*repository.TimeEntry, error)
	LogTime(ctx context.Context, taskID, userID string, durationSeconds int, description *string) (*repository.TimeEntry, error)
	GetTimeEntries(ctx context.Context, taskID, userID string) ([]*repository.TimeEntry, error)
	GetTotalTime(ctx context.Context, taskID string) (int, error)
	
	// DEPENDENCIES
	AddDependency(ctx context.Context, taskID, dependsOnTaskID, depType, userID string) error
	RemoveDependency(ctx context.Context, taskID, dependsOnTaskID, userID string) error
	ListDependencies(ctx context.Context, taskID, userID string) ([]*repository.TaskDependency, error)
	ListBlockedBy(ctx context.Context, taskID, userID string) ([]*repository.TaskDependency, error)
	
	// CHECKLISTS
	CreateChecklist(ctx context.Context, taskID, userID, title string) (*repository.TaskChecklist, error)
	AddChecklistItem(ctx context.Context, checklistID, userID, content string, assigneeID *string) (*repository.ChecklistItem, error)
	ToggleChecklistItem(ctx context.Context, itemID, userID string) error
	DeleteChecklistItem(ctx context.Context, itemID, userID string) error
	ListChecklists(ctx context.Context, taskID, userID string) ([]*repository.TaskChecklist, error)
	
	// ACTIVITY
	GetActivity(ctx context.Context, taskID, userID string, limit int) ([]*repository.TaskActivity, error)
	
	// ADVANCED FILTERING
	FilterTasks(ctx context.Context, filters *repository.TaskFilters, userID string) ([]*repository.Task, int, error)
	FindOverdue(ctx context.Context, projectID, userID string) ([]*repository.Task, error)
	FindBlocked(ctx context.Context, projectID, userID string) ([]*repository.Task, error)
	
	// SCRUM SPECIFIC
	GetBacklog(ctx context.Context, projectID, userID string) ([]*repository.Task, error)
	GetSprintBoard(ctx context.Context, sprintID, userID string) (map[string][]*repository.Task, error)
	GetSprintVelocity(ctx context.Context, sprintID, userID string) (int, error)
	GetSprintBurndown(ctx context.Context, sprintID, userID string) (*SprintBurndown, error)
	UpdatePosition(ctx context.Context, taskID string, position int, userID string) error

	ReorderTasksInColumn(ctx context.Context, projectID, status, movedTaskID string, newPosition int, userID string) error
	
	// BULK OPERATIONS
	BulkUpdateStatus(ctx context.Context, taskIDs []string, status, userID string) error
	BulkAssign(ctx context.Context, taskIDs []string, assigneeID, actorID string) error
	BulkMoveToSprint(ctx context.Context, taskIDs []string, sprintID, userID string) error
}

type SprintBurndown struct {
	SprintID            string                   `json:"sprintId"`
	StartDate           time.Time                `json:"startDate"`
	EndDate             time.Time                `json:"endDate"`
	TotalStoryPoints    int                      `json:"totalStoryPoints"`
	CompletedPoints     int                      `json:"completedPoints"`
	RemainingPoints     int                      `json:"remainingPoints"`
	IdealBurndown       []BurndownPoint          `json:"idealBurndown"`
	ActualBurndown      []BurndownPoint          `json:"actualBurndown"`
	CompletionRate      float64                  `json:"completionRate"`
}

type BurndownPoint struct {
	Date   time.Time `json:"date"`
	Points int       `json:"points"`
}





type taskService struct {
	taskRepo        repository.TaskRepository
	commentRepo     repository.TaskCommentRepository
	attachmentRepo  repository.TaskAttachmentRepository
	timeEntryRepo   repository.TimeEntryRepository
	dependencyRepo  repository.TaskDependencyRepository
	checklistRepo   repository.TaskChecklistRepository
	activityRepo    repository.TaskActivityRepository
	projectRepo     repository.ProjectRepository
	sprintRepo      repository.SprintRepository
	userRepo        repository.UserRepository
	memberService   MemberService
	permService     PermissionService
	notificationSvc *notification.Service
	broadcaster     *socket.Broadcaster
}

// Constructor
func NewTaskService(
	taskRepo repository.TaskRepository,
	commentRepo repository.TaskCommentRepository,
	attachmentRepo repository.TaskAttachmentRepository,
	timeEntryRepo repository.TimeEntryRepository,
	dependencyRepo repository.TaskDependencyRepository,
	checklistRepo repository.TaskChecklistRepository,
	activityRepo repository.TaskActivityRepository,
	projectRepo repository.ProjectRepository,
	sprintRepo repository.SprintRepository,
	userRepo repository.UserRepository,
	memberService MemberService,
	permService PermissionService,
	notificationSvc *notification.Service,
	broadcaster *socket.Broadcaster,
) TaskService {
	return &taskService{
		taskRepo:        taskRepo,
		commentRepo:     commentRepo,
		attachmentRepo:  attachmentRepo,
		timeEntryRepo:   timeEntryRepo,
		dependencyRepo:  dependencyRepo,
		checklistRepo:   checklistRepo,
		activityRepo:    activityRepo,
		projectRepo:     projectRepo,
		sprintRepo:      sprintRepo,
		userRepo:        userRepo,
		memberService:   memberService,
		permService:     permService,
		notificationSvc: notificationSvc,
		broadcaster:     broadcaster,
	}
}

// ============================================
// CREATE - With Notifications
// ============================================

func (s *taskService) Create(ctx context.Context, req *models.CreateTaskRequest) (*repository.Task, error) {
	// Verify project exists
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil || project == nil {
		return nil, ErrNotFound
	}

	// Set defaults
	if req.Status == "" {
		req.Status = "todo"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	// Verify parent task belongs to same project (if provided)
	if req.ParentTaskID != nil {
		parentTask, err := s.taskRepo.FindByID(ctx, *req.ParentTaskID)
		if err != nil || parentTask == nil {
			return nil, ErrNotFound
		}
		if parentTask.ProjectID != req.ProjectID {
			return nil, ErrInvalidInput
		}
	}

	// Verify assignees have access to project
	for _, assigneeID := range req.AssigneeIDs {
		hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, req.ProjectID, assigneeID)
		if err != nil || !hasAccess {
			return nil, ErrUnauthorized
		}
	}

	// Verify creator has access to project
	if req.CreatedBy != nil {
		hasAccess, _, err := s.memberService.HasEffectiveAccess(
			ctx, EntityTypeProject, req.ProjectID, *req.CreatedBy,
		)
		if err != nil || !hasAccess {
			return nil, ErrUnauthorized
		}
	}

	task := &repository.Task{
		ProjectID:      req.ProjectID,
		SprintID:       req.SprintID,
		ParentTaskID:   req.ParentTaskID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		Priority:       req.Priority,
		Type:           req.Type,           // ✅ Include Type
		AssigneeIDs:    req.AssigneeIDs,
		LabelIDs:       req.LabelIDs,
		EstimatedHours: req.EstimatedHours,
		StoryPoints:    req.StoryPoints,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
		CreatedBy:      req.CreatedBy,
		WatcherIDs:     []string{}, // Initialize empty
	}

	// Auto-add creator as watcher
	if req.CreatedBy != nil {
		task.WatcherIDs = append(task.WatcherIDs, *req.CreatedBy)
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// ✅ CREATE SUBTASKS
	if len(req.Subtasks) > 0 {
		for _, subtaskReq := range req.Subtasks {
			subtask := &repository.Task{
				ProjectID:      task.ProjectID,
				SprintID:       task.SprintID,
				ParentTaskID:   &task.ID,  // ✅ Link to parent
				Title:          subtaskReq.Title,
				Description:    subtaskReq.Description,
				Status:         subtaskReq.Status,
				Priority:       subtaskReq.Priority,
				Type:           strPtr("subtask"), // ✅ Mark as subtask
				AssigneeIDs:    subtaskReq.AssigneeIDs,
				LabelIDs:       []string{},
				EstimatedHours: subtaskReq.EstimatedHours,
				StoryPoints:    subtaskReq.StoryPoints,
				CreatedBy:      req.CreatedBy,
				WatcherIDs:     []string{},
			}
			
			// Set defaults for subtask
			if subtask.Status == "" {
				subtask.Status = "todo"
			}
			if subtask.Priority == "" {
				subtask.Priority = "medium"
			}
			
			// Auto-add creator as watcher to subtask
			if req.CreatedBy != nil {
				subtask.WatcherIDs = append(subtask.WatcherIDs, *req.CreatedBy)
			}
			
			// Verify subtask assignees have access
			for _, assigneeID := range subtask.AssigneeIDs {
				hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, assigneeID)
				if err != nil || !hasAccess {
					continue // Skip invalid assignees
				}
			}
			
			// Create the subtask
			if err := s.taskRepo.Create(ctx, subtask); err != nil {
				log.Printf("Failed to create subtask: %v", err)
				continue // Continue creating other subtasks even if one fails
			}
		}
	}
	// ✅ END SUBTASK CREATION

	// ✅ NOTIFICATIONS START
	creatorID := ""
	if req.CreatedBy != nil {
		creatorID = *req.CreatedBy
	}

	// 1. Notify each assignee (excluding creator)
	for _, assigneeID := range req.AssigneeIDs {
		if assigneeID != creatorID {
			s.notificationSvc.SendTaskAssignedBy(
			ctx,
			assigneeID,
			creatorID,  // Pass the creator ID
			task.Title,
			task.ID,
			task.ProjectID,
		)
		}
	}

	// 2. Get all project members for notification
	members, err := s.memberService.ListEffectiveMembers(ctx, EntityTypeProject, req.ProjectID)
	if err == nil {
		// Create set of users to exclude (creator + assignees)
		excludeMap := make(map[string]bool)
		excludeMap[creatorID] = true
		for _, assigneeID := range req.AssigneeIDs {
			excludeMap[assigneeID] = true
		}

		// Collect member IDs to notify
		var memberIDs []string
		for _, member := range members {
			if !excludeMap[member.UserID] {
				memberIDs = append(memberIDs, member.UserID)
			}
		}

		// Notify about task creation
		if len(memberIDs) > 0 {
			taskKey := s.getTaskKey(task)
			s.notificationSvc.SendTaskCreated(
				ctx,
				memberIDs,
				creatorID,
				task.Title,
				taskKey,
				task.ID,
				task.ProjectID,
			)
		}
	}

	// // 3. Broadcast to project room via WebSocket
	// if s.broadcaster != nil {
	// 	s.broadcaster.BroadcastTaskCreated(
	// 		task.ProjectID,
	// 		s.taskToMap(task),
	// 					creatorID,

	// 	)
	// }
	// ✅ NOTIFICATIONS END

	return task, nil
}

// ============================================
// READ
// ============================================

func (s *taskService) GetByID(ctx context.Context, taskID, userID string) (*repository.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}

	// ✅ Check access via PermissionService
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}

	return task, nil
}

func (s *taskService) ListByProject(ctx context.Context, projectID, userID string) ([]*repository.Task, error) {
	// ✅ Check project access
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.taskRepo.FindByProjectID(ctx, projectID)
}

func (s *taskService) ListBySprint(ctx context.Context, sprintID, userID string) ([]*repository.Task, error) {
	// Get tasks in sprint
	tasks, err := s.taskRepo.FindBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// Filter by project access
	var accessibleTasks []*repository.Task
	for _, task := range tasks {
		hasAccess, _, _ := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, userID)
		if hasAccess {
			accessibleTasks = append(accessibleTasks, task)
		}
	}

	return accessibleTasks, nil
}

func (s *taskService) ListSubtasks(ctx context.Context, parentTaskID, userID string) ([]*repository.Task, error) {
	// Verify user can access parent task
	parentTask, err := s.taskRepo.FindByID(ctx, parentTaskID)
	if err != nil || parentTask == nil {
		return nil, ErrNotFound
	}

	if !s.permService.CanAccessTask(ctx, userID, parentTaskID) {
		return nil, ErrUnauthorized
	}

	return s.taskRepo.FindByParentTaskID(ctx, parentTaskID)
}

func (s *taskService) ListMyTasks(ctx context.Context, userID string) ([]*repository.Task, error) {
	return s.taskRepo.FindByAssigneeID(ctx, userID)
}

func (s *taskService) ListByStatus(ctx context.Context, projectID, status, userID string) ([]*repository.Task, error) {
	// Check project access
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.taskRepo.FindByStatus(ctx, projectID, status)
}


// ============================================
// UPDATE - With Notifications
// ============================================

func (s *taskService) Update(ctx context.Context, taskID, userID string, req *models.UpdateTaskRequest) (*repository.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, ErrNotFound
	}

	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}

	// Track changes for notification
	var changes []string
	oldStatus := task.Status
	oldAssignees := make([]string, len(task.AssigneeIDs))
	copy(oldAssignees, task.AssigneeIDs)

	// Update fields if provided
	if req.Title != nil && *req.Title != task.Title {
		task.Title = *req.Title
		changes = append(changes, "title")
	}
	if req.Description != nil {
		task.Description = req.Description
		changes = append(changes, "description")
	}
	if req.Status != nil && *req.Status != task.Status {
		task.Status = *req.Status
		changes = append(changes, "status")
		if *req.Status == "done" && task.CompletedAt == nil {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if req.Priority != nil && *req.Priority != task.Priority {
		task.Priority = *req.Priority
		changes = append(changes, "priority")
	}
		if req.Type != nil {
			task.Type = req.Type
			changes = append(changes, "type")
		}
	if req.SprintID != nil {
		task.SprintID = req.SprintID
		changes = append(changes, "sprint")
	}
	if req.AssigneeIDs != nil {
		// Verify all assignees have project access
		for _, assigneeID := range *req.AssigneeIDs {
			hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, assigneeID)
			if err != nil || !hasAccess {
				return nil, ErrUnauthorized
			}
		}
		task.AssigneeIDs = *req.AssigneeIDs
		changes = append(changes, "assignees")
	}
	if req.LabelIDs != nil {
		task.LabelIDs = *req.LabelIDs
		changes = append(changes, "labels")
	}
	if req.EstimatedHours != nil {
		task.EstimatedHours = req.EstimatedHours
		changes = append(changes, "estimated hours")
	}
	if req.ActualHours != nil {
		task.ActualHours = req.ActualHours
		changes = append(changes, "actual hours")
	}
	if req.StoryPoints != nil {
		task.StoryPoints = req.StoryPoints
		changes = append(changes, "story points")
	}
	if req.StartDate != nil {
		task.StartDate = req.StartDate
		changes = append(changes, "start date")
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
		changes = append(changes, "due date")
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// ✅ NOTIFICATIONS START
	// Track newly assigned users so we don't send them TASK_UPDATED (they get TASK_ASSIGNED instead)
	newAssigneeMap := make(map[string]bool)
	if req.AssigneeIDs != nil {
		newAssignees := s.findNewAssignees(oldAssignees, *req.AssigneeIDs)
		for _, id := range newAssignees {
			newAssigneeMap[id] = true
		}
	}

	if len(changes) > 0 {
		// 1. Notify assignees (excluding updater AND newly assigned users)
		notifiedUsers := make(map[string]bool)
		for _, assigneeID := range task.AssigneeIDs {
			// Skip if: it's the updater, OR they're newly assigned (they'll get TASK_ASSIGNED instead)
			if assigneeID != userID && !newAssigneeMap[assigneeID] {
				s.notificationSvc.SendTaskUpdatedBy(
					ctx,
					assigneeID,
					userID,
					task.Title,
					task.ID,
					task.ProjectID,
					changes,
				)
				notifiedUsers[assigneeID] = true
			}
		}

		// 2. Notify watchers (excluding updater and already notified)
		for _, watcherID := range task.WatcherIDs {
			if watcherID != userID && !notifiedUsers[watcherID] {
				s.notificationSvc.SendTaskUpdatedBy(
    ctx,
    watcherID,
    userID,  // ✅ Pass the updater ID
    task.Title,
    task.ID,
    task.ProjectID,
    changes,
)
			}
		}

		// 3. Broadcast update to project room
		if s.broadcaster != nil {
			s.broadcaster.BroadcastTaskUpdated(
				task.ProjectID,
				s.taskToMap(task),
				changes,
								userID,

			)
		}
	}

	

	// 4. Handle STATUS CHANGE specifically
	if req.Status != nil && *req.Status != oldStatus {
		// Notify assignees
		for _, assigneeID := range task.AssigneeIDs {
			if assigneeID != userID {
				s.notificationSvc.SendTaskStatusChangedBy(
    ctx,
    assigneeID,
    userID,  // ✅ Pass the changer ID
    task.Title,
    task.ID,
    task.ProjectID,
    oldStatus,
    *req.Status,
)
			}
		}

		// Broadcast status change
		if s.broadcaster != nil {
			s.broadcaster.BroadcastTaskStatusChanged(
				task.ProjectID,
				s.taskToMap(task),
				oldStatus,
				*req.Status,
								userID,

			)
		}
	}

	// 5. Handle NEW ASSIGNEES
	if req.AssigneeIDs != nil {
		newAssignees := s.findNewAssignees(oldAssignees, *req.AssigneeIDs)
		
		for _, newAssigneeID := range newAssignees {
			if newAssigneeID != userID {
				// Send notification (this also sends WebSocket via sendWebSocketNotification)
				s.notificationSvc.SendTaskAssignedBy(
					ctx,
					newAssigneeID,
					userID,
					task.Title,
					task.ID,
					task.ProjectID,
				)
				// NOTE: Removed BroadcastTaskAssigned - notification service already sends WebSocket

				// Auto-add new assignee as watcher
				if !contains(task.WatcherIDs, newAssigneeID) {
					task.WatcherIDs = append(task.WatcherIDs, newAssigneeID)
					s.taskRepo.Update(ctx, task)
				}
			}
		}
	}
	
	// ✅ NOTIFICATIONS END

	return task, nil
}



// ============================================
// DELETE - With Notifications
// ============================================

func (s *taskService) Delete(ctx context.Context, taskID, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	if !s.permService.CanDeleteTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}

	taskKey := s.getTaskKey(task)

	// ✅ NOTIFICATIONS START
	// 1. Collect users to notify (assignees + watchers, excluding deleter)
	notifyUsers := make(map[string]bool)
	for _, assigneeID := range task.AssigneeIDs {
		if assigneeID != userID {
			notifyUsers[assigneeID] = true
		}
	}
	for _, watcherID := range task.WatcherIDs {
		if watcherID != userID {
			notifyUsers[watcherID] = true
		}
	}

	// 2. Send deletion notifications
	for notifyUserID := range notifyUsers {
		s.notificationSvc.SendTaskDeleted(
			ctx,
			notifyUserID,
			task.Title,
			taskKey,
			task.ProjectID,
		)
	}

	// 3. Broadcast deletion
	if s.broadcaster != nil {
		s.broadcaster.BroadcastTaskDeleted(
			task.ProjectID,
			task.ID,
			taskKey,
			userID,
		)
	}
	// ✅ NOTIFICATIONS END

	return s.taskRepo.Delete(ctx, taskID)
}

// ============================================
// TASK OPERATIONS
// ============================================

// ============================================
// UPDATE STATUS - With Notifications
// ============================================

func (s *taskService) UpdateStatus(ctx context.Context, taskID, status, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}

	oldStatus := task.Status

	if err := s.taskRepo.UpdateStatus(ctx, taskID, status); err != nil {
		return err
	}

	// ✅ NOTIFICATIONS START
	// Notify assignees and watchers (excluding updater)
	notifiedUsers := make(map[string]bool)
	
	for _, assigneeID := range task.AssigneeIDs {
		if assigneeID != userID {
			s.notificationSvc.SendTaskStatusChangedBy(
    ctx,
    assigneeID,
    userID,  // ✅ Pass the changer ID
    task.Title,
    task.ID,
    task.ProjectID,
    oldStatus,
    status,
)
			notifiedUsers[assigneeID] = true
		}
	}

	for _, watcherID := range task.WatcherIDs {
		if watcherID != userID && !notifiedUsers[watcherID] {
			s.notificationSvc.SendTaskStatusChangedBy(
    ctx,
    watcherID,
    userID,  // ✅ Pass the changer ID
    task.Title,
    task.ID,
    task.ProjectID,
    oldStatus,
    status,
)
		}
	}

	// Broadcast status change
	if s.broadcaster != nil {
		task.Status = status // Update for broadcast
		s.broadcaster.BroadcastTaskStatusChanged(
			task.ProjectID,
			s.taskToMap(task),
			oldStatus,
			status,
			userID,
		)
	}
	// ✅ NOTIFICATIONS END

	return nil
}
func (s *taskService) UpdatePriority(ctx context.Context, taskID, priority, userID string) error {
	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}
	return s.taskRepo.UpdatePriority(ctx, taskID, priority)
}


// ============================================
// ASSIGN TASK - With Notifications
// ============================================

func (s *taskService) AssignTask(ctx context.Context, taskID, assigneeID, actorID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	if !s.permService.CanEditTask(ctx, actorID, taskID) {
		return ErrUnauthorized
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, assigneeID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	if err := s.taskRepo.AddAssignee(ctx, taskID, assigneeID); err != nil {
		return err
	}

	// ✅ NOTIFICATIONS START
	// Only notify if not self-assigning
	if assigneeID != actorID {
		s.notificationSvc.SendTaskAssignedBy(
    ctx,
    assigneeID,
    actorID,  // ✅ Pass the actor ID
    task.Title,
    task.ID,
    task.ProjectID,
)

		// Broadcast assignment
		if s.broadcaster != nil {
			s.broadcaster.BroadcastTaskAssigned(
				assigneeID,
				s.taskToMap(task),
				actorID,
			)
		}
	}

	// Auto-add assignee as watcher
	if !contains(task.WatcherIDs, assigneeID) {
		s.taskRepo.AddWatcher(ctx, taskID, assigneeID)
	}
	// ✅ NOTIFICATIONS END

	return nil
}


func (s *taskService) UnassignTask(ctx context.Context, taskID, assigneeID, actorID string) error {
	if !s.permService.CanEditTask(ctx, actorID, taskID) {
		return ErrUnauthorized
	}
	return s.taskRepo.RemoveAssignee(ctx, taskID, assigneeID)
}

func (s *taskService) AddWatcher(ctx context.Context, taskID, watcherID, actorID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	// ✅ Verify watcher has access to project
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, watcherID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	return s.taskRepo.AddWatcher(ctx, taskID, watcherID)
}

func (s *taskService) RemoveWatcher(ctx context.Context, taskID, watcherID, actorID string) error {
	return s.taskRepo.RemoveWatcher(ctx, taskID, watcherID)
}

func (s *taskService) MarkComplete(ctx context.Context, taskID, userID string) error {
	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}
	return s.taskRepo.MarkComplete(ctx, taskID)
}

func (s *taskService) MoveToSprint(ctx context.Context, taskID, sprintID, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}

	task.SprintID = &sprintID
	return s.taskRepo.Update(ctx, task)
}

// In task_service.go, add these methods:

func (s *taskService) ConvertToSubtask(ctx context.Context, taskID, parentTaskID, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}

	parentTask, err := s.taskRepo.FindByID(ctx, parentTaskID)
	if err != nil || parentTask == nil {
		return ErrNotFound
	}

	// Verify same project
	if task.ProjectID != parentTask.ProjectID {
		return ErrInvalidInput
	}

	// Prevent circular reference
	if parentTask.ParentTaskID != nil && *parentTask.ParentTaskID == taskID {
		return ErrInvalidInput
	}

	task.ParentTaskID = &parentTaskID
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// Broadcast update
	if s.broadcaster != nil {
		s.broadcaster.BroadcastTaskUpdated(
			task.ProjectID,
			s.taskToMap(task),
			[]string{"converted to subtask"},
						userID,

		)
	}

	return nil
}

// Add helper method to convert subtask to main task
func (s *taskService) PromoteToTask(ctx context.Context, taskID, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}

	if task.ParentTaskID == nil {
		return ErrInvalidInput // Already a main task
	}

	task.ParentTaskID = nil
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	if s.broadcaster != nil {
		s.broadcaster.BroadcastTaskUpdated(
			task.ProjectID,
			s.taskToMap(task),
			[]string{"promoted to main task"},
									userID,

		)
	}

	return nil
}

// ============================================
// ADD COMMENT - With Notifications
// ============================================

func (s *taskService) AddComment(
	ctx context.Context,
	taskID, userID, content string,
	mentionedUsers []string,
) (*repository.TaskComment, error) {

	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		log.Printf("[AddComment] unauthorized access userID=%s taskID=%s", userID, taskID)
		return nil, ErrUnauthorized
	}

	content = strings.TrimSpace(content)
	if content == "" {
		log.Printf("[AddComment] empty content userID=%s taskID=%s", userID, taskID)
		return nil, ErrBadRequest
	}

	// Get task info for notifications
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, ErrNotFound
	}

	comment := &repository.TaskComment{
		TaskID:         taskID,
		UserID:         userID,
		Content:        content,
		MentionedUsers: mentionedUsers,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		log.Printf("[AddComment] failed to create comment userID=%s taskID=%s err=%v",
			userID, taskID, err)
		return nil, err
	}

	// ✅ NOTIFICATIONS START
	// Get commenter info
	commenter, _ := s.userRepo.FindByID(ctx, userID)
	commenterName := "Someone"
	if commenter != nil {
		commenterName = commenter.Name
	}

	// Collect users to notify (excluding commenter)
	notifiedUsers := make(map[string]bool)

	// 1. Notify assignees
	for _, assigneeID := range task.AssigneeIDs {
		if assigneeID != userID {
			s.notificationSvc.SendTaskCommented(
				ctx,
				assigneeID,
				commenterName,
				task.Title,
				task.ID,
				task.ProjectID,
			)
			notifiedUsers[assigneeID] = true
		}
	}

	// 2. Notify watchers (excluding already notified)
	for _, watcherID := range task.WatcherIDs {
		if watcherID != userID && !notifiedUsers[watcherID] {
			s.notificationSvc.SendTaskCommented(
				ctx,
				watcherID,
				commenterName,
				task.Title,
				task.ID,
				task.ProjectID,
			)
			notifiedUsers[watcherID] = true
		}
	}

	// 3. Parse and send mention notifications
	if s.notificationSvc != nil {
		s.notificationSvc.ParseAndSendMentions(
			ctx,
			content,
			commenterName,
			task.Title,
			task.ID,
			task.ProjectID,
			userID,
		)
	}

	// 4. Broadcast comment
	if s.broadcaster != nil {
		s.broadcaster.BroadcastCommentAdded(
			task.ProjectID,
			task.ID,
			map[string]interface{}{
				"id":        comment.ID,
				"content":   comment.Content,
				"userId":    comment.UserID,
				"createdAt": comment.CreatedAt,
			},
			userID,
		)
	}
	// ✅ NOTIFICATIONS END

	// Activity logging
	if err := s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID: taskID,
		UserID: &userID,
		Action: "commented",
	}); err != nil {
		log.Printf("[AddComment] activity log failed commentID=%s taskID=%s err=%v",
			comment.ID, taskID, err)
	}

	return comment, nil
}



func (s *taskService) ListComments(
    ctx context.Context,
    taskID, userID string,
) ([]*repository.TaskComment, error) {

    if !s.permService.CanAccessTask(ctx, userID, taskID) {
        log.Printf(
            "[ListComments] unauthorized access userID=%s taskID=%s",
            userID, taskID,
        )
        return nil, ErrUnauthorized
    }

    comments, err := s.commentRepo.FindByTaskID(ctx, taskID)
    if err != nil {
        log.Printf(
            "[ListComments] failed userID=%s taskID=%s err=%v",
            userID, taskID, err,
        )
        return nil, err
    }

    return comments, nil
}



// ============================================
// UPDATE COMMENT - With Notifications
// ============================================

func (s *taskService) UpdateComment(
	ctx context.Context,
	commentID, userID, content string,
) error {

	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		log.Printf("[UpdateComment] find failed commentID=%s err=%v", commentID, err)
		return err
	}

	if comment == nil {
		log.Printf("[UpdateComment] not found commentID=%s", commentID)
		return ErrNotFound
	}

	if comment.UserID != userID {
		log.Printf("[UpdateComment] unauthorized userID=%s commentID=%s", userID, commentID)
		return ErrUnauthorized
	}

	content = strings.TrimSpace(content)
	if content == "" {
		log.Printf("[UpdateComment] empty content userID=%s commentID=%s", userID, commentID)
		return ErrBadRequest
	}

	comment.Content = content

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		log.Printf("[UpdateComment] update failed commentID=%s err=%v", commentID, err)
		return err
	}

	// ✅ BROADCAST COMMENT UPDATE
	if s.broadcaster != nil {
		task, _ := s.taskRepo.FindByID(ctx, comment.TaskID)
		if task != nil {
			s.broadcaster.BroadcastCommentUpdated(
				task.ProjectID,
				comment.TaskID,
				map[string]interface{}{
					"id":        comment.ID,
					"content":   comment.Content,
					"userId":    comment.UserID,
					"updatedAt": comment.UpdatedAt,
				},
				userID,
			)
		}
	}

	// Activity log
	if err := s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID: comment.TaskID,
		UserID: &userID,
		Action: "comment_updated",
	}); err != nil {
		log.Printf("[UpdateComment] activity log failed commentID=%s err=%v", commentID, err)
	}

	return nil
}

// ============================================
// DELETE COMMENT - With Notifications
// ============================================

func (s *taskService) DeleteComment(
	ctx context.Context,
	commentID, userID string,
) error {

	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		log.Printf("[DeleteComment] find failed commentID=%s err=%v", commentID, err)
		return err
	}

	if comment == nil {
		log.Printf("[DeleteComment] not found commentID=%s", commentID)
		return ErrNotFound
	}

	if comment.UserID != userID &&
		!s.permService.CanEditTask(ctx, userID, comment.TaskID) {
		log.Printf("[DeleteComment] unauthorized userID=%s commentID=%s taskID=%s",
			userID, commentID, comment.TaskID)
		return ErrUnauthorized
	}

	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		log.Printf("[DeleteComment] delete failed commentID=%s err=%v", commentID, err)
		return err
	}

	// ✅ BROADCAST COMMENT DELETION
	if s.broadcaster != nil {
		task, _ := s.taskRepo.FindByID(ctx, comment.TaskID)
		if task != nil {
			s.broadcaster.BroadcastCommentDeleted(
				task.ProjectID,
				comment.TaskID,
				commentID,
				userID,
			)
		}
	}

	// Activity log
	if err := s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID: comment.TaskID,
		UserID: &userID,
		Action: "comment_deleted",
	}); err != nil {
		log.Printf("[DeleteComment] activity log failed commentID=%s err=%v", commentID, err)
	}

	return nil
}


// ============================================
// ADD ATTACHMENT - With Notifications
// ============================================

func (s *taskService) AddAttachment(ctx context.Context, taskID, userID, filename, fileURL string, fileSize int64, mimeType string) (*repository.TaskAttachment, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}

	// Get task for notifications
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, ErrNotFound
	}

	attachment := &repository.TaskAttachment{
		TaskID:   taskID,
		UserID:   userID,
		Filename: filename,
		FileURL:  fileURL,
		FileSize: fileSize,
		MimeType: mimeType,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		return nil, err
	}

	// ✅ NOTIFICATIONS START
	// Get uploader info
	uploader, _ := s.userRepo.FindByID(ctx, userID)
	uploaderName := "Someone"
	if uploader != nil {
		uploaderName = uploader.Name
	}

	// Notify assignees and watchers (excluding uploader)
	notifiedUsers := make(map[string]bool)
	
	for _, assigneeID := range task.AssigneeIDs {
		if assigneeID != userID {
			s.notificationSvc.SendBatchNotifications(
				ctx,
				[]string{assigneeID},
				userID,
				"TASK_ATTACHMENT_ADDED",
				"Attachment Added",
				uploaderName+" added an attachment to task: "+task.Title,
				map[string]interface{}{
					"taskId":     taskID,
					"projectId":  task.ProjectID,
					"filename":   filename,
					"action":     "view_task",
				},
			)
			notifiedUsers[assigneeID] = true
		}
	}

	for _, watcherID := range task.WatcherIDs {
		if watcherID != userID && !notifiedUsers[watcherID] {
			s.notificationSvc.SendBatchNotifications(
				ctx,
				[]string{watcherID},
				userID,
				"TASK_ATTACHMENT_ADDED",
				"Attachment Added",
				uploaderName+" added an attachment to task: "+task.Title,
				map[string]interface{}{
					"taskId":     taskID,
					"projectId":  task.ProjectID,
					"filename":   filename,
					"action":     "view_task",
				},
			)
		}
	}
	// ✅ NOTIFICATIONS END

	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID:   taskID,
		UserID:   &userID,
		Action:   "added_attachment",
		NewValue: &filename,
	})

	return attachment, nil
}

func (s *taskService) ListAttachments(ctx context.Context, taskID, userID string) ([]*repository.TaskAttachment, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}
	return s.attachmentRepo.FindByTaskID(ctx, taskID)
}

func (s *taskService) DeleteAttachment(ctx context.Context, attachmentID, userID string) error {
	attachment, err := s.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil || attachment == nil {
		return ErrNotFound
	}

	// Only attachment uploader or task editors can delete
	if attachment.UserID != userID && !s.permService.CanEditTask(ctx, userID, attachment.TaskID) {
		return ErrUnauthorized
	}

	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID:   attachment.TaskID,
		UserID:   &userID,
		Action:   "deleted_attachment",
		OldValue: &attachment.Filename,
	})

	return s.attachmentRepo.Delete(ctx, attachmentID)
}

// ============================================
// TIME TRACKING IMPLEMENTATION
// ============================================

func (s *taskService) StartTimer(ctx context.Context, taskID, userID string) (*repository.TimeEntry, error) {
	// Check access
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}

	// Stop any existing timer
	active, _ := s.timeEntryRepo.FindActiveTimer(ctx, userID)
	if active != nil {
		s.timeEntryRepo.StopTimer(ctx, active.ID)
	}

	entry := &repository.TimeEntry{
		TaskID:    taskID,
		UserID:    userID,
		StartTime: time.Now(),
		IsManual:  false,
	}

	if err := s.timeEntryRepo.Create(ctx, entry); err != nil {
		return nil, err
	}

	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID: taskID,
		UserID: &userID,
		Action: "started_timer",
	})

	return entry, nil
}

func (s *taskService) StopTimer(ctx context.Context, userID string) (*repository.TimeEntry, error) {
	active, err := s.timeEntryRepo.FindActiveTimer(ctx, userID)
	if err != nil || active == nil {
		return nil, ErrNotFound
	}

	if err := s.timeEntryRepo.StopTimer(ctx, active.ID); err != nil {
		return nil, err
	}

	// Get updated entry
	entry, _ := s.timeEntryRepo.FindByTaskID(ctx, active.ID)
	
	// Update task actual hours
	totalSeconds, _ := s.timeEntryRepo.GetTotalTime(ctx, active.TaskID)
	task, _ := s.taskRepo.FindByID(ctx, active.TaskID)
	if task != nil {
		hours := float64(totalSeconds) / 3600.0
		task.ActualHours = &hours
		s.taskRepo.Update(ctx, task)
	}
	
	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID: active.TaskID,
		UserID: &userID,
		Action: "stopped_timer",
	})

	return entry[len(entry)-1], nil
}

func (s *taskService) GetActiveTimer(ctx context.Context, userID string) (*repository.TimeEntry, error) {
	entry, err := s.timeEntryRepo.FindActiveTimer(ctx, userID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, ErrNotFound
	}
	return entry, nil
}

func (s *taskService) LogTime(ctx context.Context, taskID, userID string, durationSeconds int, description *string) (*repository.TimeEntry, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}

	now := time.Now()
	entry := &repository.TimeEntry{
		TaskID:          taskID,
		UserID:          userID,
		StartTime:       now,
		EndTime:         &now,
		DurationSeconds: &durationSeconds,
		Description:     description,
		IsManual:        true,
	}

	if err := s.timeEntryRepo.Create(ctx, entry); err != nil {
		return nil, err
	}

	// Update task actual hours
	totalSeconds, _ := s.timeEntryRepo.GetTotalTime(ctx, taskID)
	task, _ := s.taskRepo.FindByID(ctx, taskID)
	if task != nil {
		hours := float64(totalSeconds) / 3600.0
		task.ActualHours = &hours
		s.taskRepo.Update(ctx, task)
	}

	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID: taskID,
		UserID: &userID,
		Action: "logged_time",
	})

	return entry, nil
}

func (s *taskService) GetTimeEntries(ctx context.Context, taskID, userID string) ([]*repository.TimeEntry, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}
	return s.timeEntryRepo.FindByTaskID(ctx, taskID)
}

func (s *taskService) GetTotalTime(ctx context.Context, taskID string) (int, error) {
	return s.timeEntryRepo.GetTotalTime(ctx, taskID)
}

// ============================================
// DEPENDENCIES IMPLEMENTATION
// ============================================

func (s *taskService) AddDependency(ctx context.Context, taskID, dependsOnTaskID, depType, userID string) error {
	// Verify both tasks exist and user has access
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	dependsOnTask, err := s.taskRepo.FindByID(ctx, dependsOnTaskID)
	if err != nil || dependsOnTask == nil {
		return ErrNotFound
	}

	// Verify same project
	if task.ProjectID != dependsOnTask.ProjectID {
		return ErrInvalidInput
	}

	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}

	dep := &repository.TaskDependency{
		TaskID:          taskID,
		DependsOnTaskID: dependsOnTaskID,
		DependencyType:  depType,
	}

	if err := s.dependencyRepo.Create(ctx, dep); err != nil {
		return err
	}

	// Mark task as blocked if dependency is not complete
	if depType == "blocks" && dependsOnTask.Status != "done" {
		task.Status = "blocked"
		s.taskRepo.Update(ctx, task)
	}

	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID:    taskID,
		UserID:    &userID,
		Action:    "added_dependency",
		FieldName: &depType,
		NewValue:  &dependsOnTaskID,
	})

	return nil
}

func (s *taskService) RemoveDependency(ctx context.Context, taskID, dependsOnTaskID, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}

	if err := s.dependencyRepo.Delete(ctx, taskID, dependsOnTaskID); err != nil {
		return err
	}

	// Check if task should still be blocked
	deps, _ := s.dependencyRepo.FindByTaskID(ctx, taskID)
	stillBlocked := false
	for _, dep := range deps {
		if dep.DependencyType == "blocks" {
			dependsTask, _ := s.taskRepo.FindByID(ctx, dep.DependsOnTaskID)
			if dependsTask != nil && dependsTask.Status != "done" {
				stillBlocked = true
				break
			}
		}
	}
	task.Blocked = stillBlocked
	s.taskRepo.Update(ctx, task)

	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID:   taskID,
		UserID:   &userID,
		Action:   "removed_dependency",
		OldValue: &dependsOnTaskID,
	})

	return nil
}

func (s *taskService) ListDependencies(ctx context.Context, taskID, userID string) ([]*repository.TaskDependency, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}
	return s.dependencyRepo.FindByTaskID(ctx, taskID)
}

func (s *taskService) ListBlockedBy(ctx context.Context, taskID, userID string) ([]*repository.TaskDependency, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}
	return s.dependencyRepo.FindBlockedBy(ctx, taskID)
}

// ============================================
// CHECKLISTS IMPLEMENTATION
// ============================================

func (s *taskService) CreateChecklist(ctx context.Context, taskID, userID, title string) (*repository.TaskChecklist, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}

	checklist := &repository.TaskChecklist{
		TaskID: taskID,
		Title:  title,
	}

	if err := s.checklistRepo.CreateChecklist(ctx, checklist); err != nil {
		return nil, err
	}

	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID:   taskID,
		UserID:   &userID,
		Action:   "created_checklist",
		NewValue: &title,
	})

	return checklist, nil
}

func (s *taskService) AddChecklistItem(ctx context.Context, checklistID, userID, content string, assigneeID *string) (*repository.ChecklistItem, error) {
	checklist, err := s.checklistRepo.FindChecklistByID(ctx, checklistID)
	if err != nil || checklist == nil {
		return nil, ErrNotFound
	}

	if !s.permService.CanAccessTask(ctx, userID, checklist.TaskID) {
		return nil, ErrUnauthorized
	}

	// Verify assignee has access if provided
	if assigneeID != nil {
		task, _ := s.taskRepo.FindByID(ctx, checklist.TaskID)
		if task != nil {
			hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, *assigneeID)
			if err != nil || !hasAccess {
				return nil, ErrUnauthorized
			}
		}
	}

	item := &repository.ChecklistItem{
		ChecklistID: checklistID,
		Content:     content,
		AssigneeID:  assigneeID,
	}

	if err := s.checklistRepo.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}


// ============================================
// CHECKLIST ITEM TOGGLE - With Notifications
// ============================================

func (s *taskService) ToggleChecklistItem(ctx context.Context, itemID, userID string) error {
	item, err := s.checklistRepo.FindItemByID(ctx, itemID)
	if err != nil || item == nil {
		return ErrNotFound
	}

	checklist, err := s.checklistRepo.FindChecklistByID(ctx, item.ChecklistID)
	if err != nil || checklist == nil {
		return ErrNotFound
	}

	if !s.permService.CanAccessTask(ctx, userID, checklist.TaskID) {
		return ErrUnauthorized
	}

	if err := s.checklistRepo.ToggleItem(ctx, itemID); err != nil {
		return err
	}

	// ✅ NOTIFICATIONS START
	// Get updated item
	updatedItem, _ := s.checklistRepo.FindItemByID(ctx, itemID)
	if updatedItem != nil && updatedItem.Completed {
		// Get task for notifications
		task, _ := s.taskRepo.FindByID(ctx, checklist.TaskID)
		if task != nil {
			// Get completer info
			completer, _ := s.userRepo.FindByID(ctx, userID)
			completerName := "Someone"
			if completer != nil {
				completerName = completer.Name
			}

			// Notify assignee if item has one (excluding completer)
			if item.AssigneeID != nil && *item.AssigneeID != userID {
				s.notificationSvc.SendBatchNotifications(
					ctx,
					[]string{*item.AssigneeID},
					userID,
					"CHECKLIST_ITEM_COMPLETED",
					"Checklist Item Completed",
					completerName+" completed a checklist item in task: "+task.Title,
					map[string]interface{}{
						"taskId":    checklist.TaskID,
						"projectId": task.ProjectID,
						"itemId":    itemID,
						"action":    "view_task",
					},
				)
			}
		}
	}
	// ✅ NOTIFICATIONS END

	return nil
}


func (s *taskService) DeleteChecklistItem(ctx context.Context, itemID, userID string) error {
	item, err := s.checklistRepo.FindItemByID(ctx, itemID)
	if err != nil || item == nil {
		return ErrNotFound
	}

	checklist, err := s.checklistRepo.FindChecklistByID(ctx, item.ChecklistID)
	if err != nil || checklist == nil {
		return ErrNotFound
	}

	if !s.permService.CanEditTask(ctx, userID, checklist.TaskID) {
		return ErrUnauthorized
	}

	return s.checklistRepo.DeleteItem(ctx, itemID)
}

func (s *taskService) ListChecklists(ctx context.Context, taskID, userID string) ([]*repository.TaskChecklist, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}
	return s.checklistRepo.FindByTaskID(ctx, taskID)
}

// ============================================
// ACTIVITY IMPLEMENTATION
// ============================================

func (s *taskService) GetActivity(ctx context.Context, taskID, userID string, limit int) ([]*repository.TaskActivity, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}

	if limit <= 0 {
		limit = 50 // Default limit
	}

	return s.activityRepo.FindByTaskID(ctx, taskID, limit)
}

// ============================================
// ADVANCED FILTERING
// ============================================

func (s *taskService) FilterTasks(ctx context.Context, filters *repository.TaskFilters, userID string) ([]*repository.Task, int, error) {
	// Check project access
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, filters.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, 0, ErrUnauthorized
	}

	return s.taskRepo.FindWithFilters(ctx, filters)
}

func (s *taskService) FindOverdue(ctx context.Context, projectID, userID string) ([]*repository.Task, error) {
	// Check project access
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.taskRepo.FindOverdue(ctx, projectID)
}

func (s *taskService) FindBlocked(ctx context.Context, projectID, userID string) ([]*repository.Task, error) {
	// Check project access
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.taskRepo.FindBlocked(ctx, projectID)
}

// ============================================
// SCRUM SPECIFIC IMPLEMENTATION
// ============================================

func (s *taskService) GetBacklog(ctx context.Context, projectID, userID string) ([]*repository.Task, error) {
	// Check project access
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.taskRepo.FindBacklog(ctx, projectID)
}

func (s *taskService) GetSprintBoard(ctx context.Context, sprintID, userID string) (map[string][]*repository.Task, error) {
	// Get all tasks in sprint
	tasks, err := s.taskRepo.FindBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// Group by status
	board := make(map[string][]*repository.Task)
	statuses := []string{"todo", "in_progress", "in_review", "done"}
	
	for _, status := range statuses {
		board[status] = []*repository.Task{}
	}

	for _, task := range tasks {
		// Check user has access to task's project
		hasAccess, _, _ := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, userID)
		if hasAccess {
			board[task.Status] = append(board[task.Status], task)
		}
	}

	return board, nil
}

func (s *taskService) GetSprintVelocity(ctx context.Context, sprintID, userID string) (int, error) {
	// Verify user has access to sprint
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return 0, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return 0, ErrUnauthorized
	}

	return s.taskRepo.GetSprintVelocity(ctx, sprintID)
}

func (s *taskService) GetSprintBurndown(ctx context.Context, sprintID, userID string) (*SprintBurndown, error) {
	// Get sprint
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Get total and completed story points
	totalPoints, _ := s.taskRepo.GetSprintVelocity(ctx, sprintID)
	completedPoints, _ := s.taskRepo.GetCompletedStoryPoints(ctx, sprintID)
	remainingPoints := totalPoints - completedPoints

	// Calculate ideal burndown
	sprintDays := int(sprint.EndDate.Sub(sprint.StartDate).Hours() / 24)
	if sprintDays == 0 {
		sprintDays = 1 // Prevent division by zero
	}
	pointsPerDay := float64(totalPoints) / float64(sprintDays)
	
	idealBurndown := []BurndownPoint{}
	for i := 0; i <= sprintDays; i++ {
		date := sprint.StartDate.AddDate(0, 0, i)
		points := totalPoints - int(float64(i)*pointsPerDay)
		if points < 0 {
			points = 0
		}
		idealBurndown = append(idealBurndown, BurndownPoint{
			Date:   date,
			Points: points,
		})
	}

	// Calculate actual burndown from activity history
	actualBurndown := []BurndownPoint{}
	tasks, _ := s.taskRepo.FindBySprintID(ctx, sprintID)
	
	// Create map of date -> completed points
	completedByDate := make(map[string]int)
	for _, task := range tasks {
		if task.CompletedAt != nil && task.StoryPoints != nil {
			dateStr := task.CompletedAt.Format("2006-01-02")
			completedByDate[dateStr] += *task.StoryPoints
		}
	}

	// Build actual burndown
	currentRemaining := totalPoints
	for i := 0; i <= sprintDays; i++ {
		date := sprint.StartDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		
		if completed, ok := completedByDate[dateStr]; ok {
			currentRemaining -= completed
		}
		
		if currentRemaining < 0 {
			currentRemaining = 0
		}
		
		actualBurndown = append(actualBurndown, BurndownPoint{
			Date:   date,
			Points: currentRemaining,
		})
	}

	completionRate := 0.0
	if totalPoints > 0 {
		completionRate = float64(completedPoints) / float64(totalPoints) * 100
	}

	return &SprintBurndown{
		SprintID:         sprintID,
		StartDate:        sprint.StartDate,
		EndDate:          sprint.EndDate,
		TotalStoryPoints: totalPoints,
		CompletedPoints:  completedPoints,
		RemainingPoints:  remainingPoints,
		IdealBurndown:    idealBurndown,
		ActualBurndown:   actualBurndown,
		CompletionRate:   completionRate,
	}, nil
}

// ============================================
// BULK OPERATIONS
// ============================================

func (s *taskService) BulkUpdateStatus(ctx context.Context, taskIDs []string, status, userID string) error {
	// Verify user can edit all tasks
	for _, taskID := range taskIDs {
		if !s.permService.CanEditTask(ctx, userID, taskID) {
			return ErrUnauthorized
		}
	}

	return s.taskRepo.BulkUpdateStatus(ctx, taskIDs, status)
}

func (s *taskService) BulkAssign(ctx context.Context, taskIDs []string, assigneeID, actorID string) error {
	// Verify actor can edit all tasks
	for _, taskID := range taskIDs {
		if !s.permService.CanEditTask(ctx, actorID, taskID) {
			return ErrUnauthorized
		}
	}

	// Verify assignee has access to all task projects
	for _, taskID := range taskIDs {
		task, err := s.taskRepo.FindByID(ctx, taskID)
		if err != nil || task == nil {
			return ErrNotFound
		}
		
		hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, assigneeID)
		if err != nil || !hasAccess {
			return ErrUnauthorized
		}
	}

	// Add assignee to all tasks
	for _, taskID := range taskIDs {
		if err := s.taskRepo.AddAssignee(ctx, taskID, assigneeID); err != nil {
			return err
		}
	}

	return nil
}

func (s *taskService) BulkMoveToSprint(ctx context.Context, taskIDs []string, sprintID, userID string) error {
	// Verify user can edit all tasks
	for _, taskID := range taskIDs {
		if !s.permService.CanEditTask(ctx, userID, taskID) {
			return ErrUnauthorized
		}
	}

	return s.taskRepo.BulkMoveToSprint(ctx, taskIDs, sprintID)
}


// ============================================
// DRAG AND DROP
// ============================================

func (s *taskService) UpdatePosition(ctx context.Context, taskID string, position int, userID string) error {
	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}
	return s.taskRepo.UpdatePosition(ctx, taskID, position)
}


// ✅ FIXED: service/task_service.go - ReorderTasksInColumn
func (s *taskService) ReorderTasksInColumn(
	ctx context.Context,
	projectID string,
	status string,
	movedTaskID string,
	newPosition int,
	userID string,
) error {
	log.Printf("🔄 ReorderTasksInColumn: project=%s, status=%s, movedTask=%s, newPos=%d",
		projectID, status, movedTaskID, newPosition)

	// Get ALL tasks in target column
	allTasks, err := s.taskRepo.FindByStatus(ctx, projectID, status)
	if err != nil {
		return err
	}

	log.Printf("📊 Found %d tasks in column %s", len(allTasks), status)

	// Separate parents from subtasks
	parents := make([]*repository.Task, 0)
	subtasksMap := make(map[string][]*repository.Task)
	
	for _, t := range allTasks {
		if t.ParentTaskID == nil {
			parents = append(parents, t)
		} else {
			if _, exists := subtasksMap[*t.ParentTaskID]; !exists {
				subtasksMap[*t.ParentTaskID] = make([]*repository.Task, 0)
			}
			subtasksMap[*t.ParentTaskID] = append(subtasksMap[*t.ParentTaskID], t)
		}
	}

	// Find moved task in parents list
	var movedTask *repository.Task
	movedIndex := -1
	for i, t := range parents {
		if t.ID == movedTaskID {
			movedTask = t
			movedIndex = i
			break
		}
	}

	if movedTask == nil {
		log.Printf("❌ Moved task not found in parents list")
		return ErrNotFound
	}

	log.Printf("🎯 Found moved task at index %d", movedIndex)

	// Build list without moved task
	otherParents := make([]*repository.Task, 0, len(parents)-1)
	for i, t := range parents {
		if i != movedIndex {
			otherParents = append(otherParents, t)
		}
	}

	// Clamp new position
	if newPosition < 0 {
		newPosition = 0
	}
	if newPosition > len(otherParents) {
		newPosition = len(otherParents)
	}

	log.Printf("📍 Inserting at position %d (out of %d tasks)", newPosition, len(otherParents))

	// Build final order with moved task inserted at new position
	finalOrder := make([]*repository.Task, 0, len(parents))
	finalOrder = append(finalOrder, otherParents[:newPosition]...)
	finalOrder = append(finalOrder, movedTask)
	finalOrder = append(finalOrder, otherParents[newPosition:]...)

	// ✅ Update positions in database - CRITICAL FIX
	for i, task := range finalOrder {
		log.Printf("✏️ Updating %s to position %d", task.Title, i)
		
		if err := s.taskRepo.UpdatePosition(ctx, task.ID, i); err != nil {
			log.Printf("❌ Failed to update position for task %s: %v", task.ID, err)
			return err
		}

		// ✅ Update subtasks (they stay right after parent in visual order)
		// But we don't need to update their DB position if they follow parent
		// Just log for debugging
		if subtasks := subtasksMap[task.ID]; len(subtasks) > 0 {
			log.Printf("   └─ Task has %d subtasks", len(subtasks))
		}
	}

	log.Printf("✅ Reordering complete")


	// ✅ CRITICAL FIX: Broadcast to ALL users (including the one who moved it!)
	// if s.broadcaster != nil {
	// 	log.Printf("📡 Broadcasting position update: project=%s, task=%s", projectID, movedTaskID)
		
	// 	// Fetch the updated task with all relations
	// 	updatedTask, err := s.taskRepo.FindByID(ctx, movedTaskID)
	// 	if err != nil {
	// 		log.Printf("⚠️ Failed to fetch updated task: %v", err)
	// 	} else {
	// 		// ✅ Pass empty string "" for excludeUserID to broadcast to EVERYONE
	// 		s.broadcaster.BroadcastTaskUpdated(
	// 			projectID,
	// 			s.taskToMap(updatedTask),
	// 			[]string{"position", "status"},
	// 			"", // ❌ CRITICAL: Empty string = broadcast to ALL users!
	// 		)
	// 		log.Printf("✅ Broadcasted task update to all users in project")
	// 	}
	// }


	return nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

func (s *taskService) taskToMap(task *repository.Task) map[string]interface{} {
	return map[string]interface{}{
		"id":          task.ID,
		"projectId":   task.ProjectID,
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"priority":    task.Priority,
		"assigneeIds": task.AssigneeIDs,
		"watcherIds":  task.WatcherIDs,
		"dueDate":     task.DueDate,
		"createdAt":   task.CreatedAt,
		"updatedAt":   task.UpdatedAt,
	}
}

func (s *taskService) getTaskKey(task *repository.Task) string {
	// If task has a Key field, use it
	// Otherwise generate a simple key
	if task.ID != "" {
		return task.ID
	}
	// Fallback: use task ID or generate project prefix + number
	return task.ID[:8] // Use first 8 chars of ID
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *taskService) findNewAssignees(oldAssignees, newAssignees []string) []string {
	oldMap := make(map[string]bool)
	for _, id := range oldAssignees {
		oldMap[id] = true
	}

	var result []string
	for _, id := range newAssignees {
		if !oldMap[id] {
			result = append(result, id)
		}
	}
	return result
}


