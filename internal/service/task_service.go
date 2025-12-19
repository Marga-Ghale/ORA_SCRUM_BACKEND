package service

import (
	"context"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

type TaskService interface {
	// Task CRUD
	Create(ctx context.Context, req *CreateTaskRequest) (*repository.Task, error)
	GetByID(ctx context.Context, taskID, userID string) (*repository.Task, error)
	Update(ctx context.Context, taskID, userID string, req *UpdateTaskRequest) (*repository.Task, error)
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
	taskRepo       repository.TaskRepository
	commentRepo    repository.TaskCommentRepository
	attachmentRepo repository.TaskAttachmentRepository
	timeEntryRepo  repository.TimeEntryRepository
	dependencyRepo repository.TaskDependencyRepository
	checklistRepo  repository.TaskChecklistRepository
	activityRepo   repository.TaskActivityRepository
	projectRepo    repository.ProjectRepository
	sprintRepo     repository.SprintRepository
	memberService  MemberService
	permService    PermissionService
}

type CreateTaskRequest struct {
	ProjectID      string
	SprintID       *string
	ParentTaskID   *string
	Title          string
	Description    *string
	Status         string
	Priority       string
	AssigneeIDs    []string
	LabelIDs       []string
	EstimatedHours *float64
	StoryPoints    *int
	StartDate      *time.Time
	DueDate        *time.Time
	CreatedBy      *string
}

type UpdateTaskRequest struct {
	Title          *string
	Description    *string
	Status         *string
	Priority       *string
	SprintID       *string
	AssigneeIDs    *[]string
	LabelIDs       *[]string
	EstimatedHours *float64
	ActualHours    *float64
	StoryPoints    *int
	StartDate      *time.Time
	DueDate        *time.Time
}

// ✅ Fixed constructor with all repositories
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
	memberService MemberService,
	permService PermissionService,
) TaskService {
	return &taskService{
		taskRepo:       taskRepo,
		commentRepo:    commentRepo,
		attachmentRepo: attachmentRepo,
		timeEntryRepo:  timeEntryRepo,
		dependencyRepo: dependencyRepo,
		checklistRepo:  checklistRepo,
		activityRepo:   activityRepo,
		projectRepo:    projectRepo,
		sprintRepo:     sprintRepo,
		memberService:  memberService,
		permService:    permService,
	}
}

// ============================================
// CREATE
// ============================================

func (s *taskService) Create(ctx context.Context, req *CreateTaskRequest) (*repository.Task, error) {
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

	// ✅ Verify assignees have access to project
	for _, assigneeID := range req.AssigneeIDs {
		hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, req.ProjectID, assigneeID)
		if err != nil || !hasAccess {
			return nil, ErrUnauthorized
		}
	}
	// ✅ Verify creator has access to project
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
		AssigneeIDs:    req.AssigneeIDs,
		LabelIDs:       req.LabelIDs,
		EstimatedHours: req.EstimatedHours,
		StoryPoints:    req.StoryPoints,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
		CreatedBy:      req.CreatedBy,  

	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

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
// UPDATE
// ============================================

func (s *taskService) Update(ctx context.Context, taskID, userID string, req *UpdateTaskRequest) (*repository.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, ErrNotFound
	}

	// ✅ Check edit permission via PermissionService
	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}

	// Update fields if provided
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
		if *req.Status == "done" && task.CompletedAt == nil {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.SprintID != nil {
		task.SprintID = req.SprintID
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
	}
	if req.LabelIDs != nil {
		task.LabelIDs = *req.LabelIDs
	}
	if req.EstimatedHours != nil {
		task.EstimatedHours = req.EstimatedHours
	}
	if req.ActualHours != nil {
		task.ActualHours = req.ActualHours
	}
	if req.StoryPoints != nil {
		task.StoryPoints = req.StoryPoints
	}
	if req.StartDate != nil {
		task.StartDate = req.StartDate
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) Delete(ctx context.Context, taskID, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	// ✅ Check delete permission via PermissionService
	if !s.permService.CanDeleteTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}

	return s.taskRepo.Delete(ctx, taskID)
}

// ============================================
// TASK OPERATIONS
// ============================================

func (s *taskService) UpdateStatus(ctx context.Context, taskID, status, userID string) error {
	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}
	return s.taskRepo.UpdateStatus(ctx, taskID, status)
}

func (s *taskService) UpdatePriority(ctx context.Context, taskID, priority, userID string) error {
	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}
	return s.taskRepo.UpdatePriority(ctx, taskID, priority)
}

func (s *taskService) AssignTask(ctx context.Context, taskID, assigneeID, actorID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	// Check actor can edit task
	if !s.permService.CanEditTask(ctx, actorID, taskID) {
		return ErrUnauthorized
	}

	// ✅ Verify assignee has access to project
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, assigneeID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	return s.taskRepo.AddAssignee(ctx, taskID, assigneeID)
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

func (s *taskService) ConvertToSubtask(ctx context.Context, taskID, parentTaskID, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return ErrNotFound
	}

	parentTask, err := s.taskRepo.FindByID(ctx, parentTaskID)
	if err != nil || parentTask == nil {
		return ErrNotFound
	}

	// Verify both tasks in same project
	if task.ProjectID != parentTask.ProjectID {
		return ErrInvalidInput
	}

	if !s.permService.CanEditTask(ctx, userID, taskID) {
		return ErrUnauthorized
	}

	task.ParentTaskID = &parentTaskID
	return s.taskRepo.Update(ctx, task)
}

// ============================================
// COMMENTS IMPLEMENTATION
// ============================================

func (s *taskService) AddComment(ctx context.Context, taskID, userID, content string, mentionedUsers []string) (*repository.TaskComment, error) {
	// Check access
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}

	comment := &repository.TaskComment{
		TaskID:         taskID,
		UserID:         userID,
		Content:        content,
		MentionedUsers: mentionedUsers,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID: taskID,
		UserID: &userID,
		Action: "commented",
	})

	// TODO: Send notifications to mentioned users
	
	return comment, nil
}

func (s *taskService) ListComments(ctx context.Context, taskID, userID string) ([]*repository.TaskComment, error) {
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
	}
	return s.commentRepo.FindByTaskID(ctx, taskID)
}

func (s *taskService) UpdateComment(ctx context.Context, commentID, userID, content string) error {
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil || comment == nil {
		return ErrNotFound
	}

	// Only comment author can update
	if comment.UserID != userID {
		return ErrUnauthorized
	}

	comment.Content = content
	return s.commentRepo.Update(ctx, comment)
}

func (s *taskService) DeleteComment(ctx context.Context, commentID, userID string) error {
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil || comment == nil {
		return ErrNotFound
	}

	// Only comment author or task editors can delete
	if comment.UserID != userID && !s.permService.CanEditTask(ctx, userID, comment.TaskID) {
		return ErrUnauthorized
	}

	return s.commentRepo.Delete(ctx, commentID)
}

// ============================================
// ATTACHMENTS IMPLEMENTATION
// ============================================

func (s *taskService) AddAttachment(ctx context.Context, taskID, userID, filename, fileURL string, fileSize int64, mimeType string) (*repository.TaskAttachment, error) {
	// Check access
	if !s.permService.CanAccessTask(ctx, userID, taskID) {
		return nil, ErrUnauthorized
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

	// Log activity
	s.activityRepo.Create(ctx, &repository.TaskActivity{
		TaskID:    taskID,
		UserID:    &userID,
		Action:    "added_attachment",
		NewValue:  &filename,
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

	return s.checklistRepo.ToggleItem(ctx, itemID)
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