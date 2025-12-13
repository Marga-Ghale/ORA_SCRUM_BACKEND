package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Sprint Service
// ============================================

type SprintService interface {
	Create(ctx context.Context, projectID, name string, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error)
	GetByID(ctx context.Context, id string) (*repository.Sprint, error)
	ListByProject(ctx context.Context, projectID string) ([]*repository.Sprint, error)
	GetActive(ctx context.Context, projectID string) (*repository.Sprint, error)
	Update(ctx context.Context, id string, name, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error)
	Delete(ctx context.Context, id string) error
	Start(ctx context.Context, id, userID string) (*repository.Sprint, error)
	Complete(ctx context.Context, id, moveIncomplete, userID string) (*repository.Sprint, error)
}

type sprintService struct {
	sprintRepo  repository.SprintRepository
	taskRepo    repository.TaskRepository
	projectRepo repository.ProjectRepository
	notifSvc    *notification.Service
}

func NewSprintService(sprintRepo repository.SprintRepository, taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository, notifSvc *notification.Service) SprintService {
	return &sprintService{
		sprintRepo:  sprintRepo,
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		notifSvc:    notifSvc,
	}
}

func (s *sprintService) Create(ctx context.Context, projectID, name string, goal *string, startDate, endDate *time.Time) (*repository.Sprint, error) {
	sprint := &repository.Sprint{
		ProjectID: projectID,
		Name:      name,
		Goal:      goal,
		Status:    "planning", // lowercase
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

func (s *sprintService) GetActive(ctx context.Context, projectID string) (*repository.Sprint, error) {
	return s.sprintRepo.FindActive(ctx, projectID)
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

func (s *sprintService) Start(ctx context.Context, id, userID string) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	// Check if there's already an active sprint
	active, _ := s.sprintRepo.FindActive(ctx, sprint.ProjectID)
	if active != nil && active.ID != id {
		return nil, fmt.Errorf("another sprint is already active")
	}

	now := time.Now()
	sprint.Status = "active" // lowercase
	sprint.StartDate = &now

	if err := s.sprintRepo.Update(ctx, sprint); err != nil {
		return nil, err
	}

	// Send notifications
	if s.notifSvc != nil && s.projectRepo != nil {
		memberIDs, _ := s.projectRepo.FindMemberUserIDs(ctx, sprint.ProjectID)
		if len(memberIDs) > 0 {
			s.notifSvc.SendSprintStartedToMembers(ctx, memberIDs, sprint.Name, sprint.ID, sprint.ProjectID)
		}
	}

	return sprint, nil
}

func (s *sprintService) Complete(ctx context.Context, id, moveIncomplete, userID string) (*repository.Sprint, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, id)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	totalTasks, completedTasks, _ := s.taskRepo.CountBySprintID(ctx, id)

	now := time.Now()
	sprint.Status = "completed" // lowercase
	sprint.EndDate = &now

	// Move incomplete tasks
	if moveIncomplete != "" {
		tasks, _ := s.taskRepo.FindBySprintID(ctx, id)
		for _, task := range tasks {
			if task.Status != "done" && task.Status != "cancelled" { // lowercase
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

	// Send notifications
	if s.notifSvc != nil && s.projectRepo != nil {
		memberIDs, _ := s.projectRepo.FindMemberUserIDs(ctx, sprint.ProjectID)
		if len(memberIDs) > 0 {
			s.notifSvc.SendSprintCompletedToMembers(ctx, memberIDs, sprint.Name, sprint.ID, sprint.ProjectID, completedTasks, totalTasks)
		}
	}

	return sprint, nil
}

// ============================================
// Task Service
// ============================================

type TaskService interface {
	Create(ctx context.Context, projectID, reporterID, title string, description *string, status, priority, taskType *string, assigneeID, sprintID, parentID *string, storyPoints *int, dueDate *time.Time, labels []string) (*repository.Task, error)
	GetByID(ctx context.Context, id string) (*repository.Task, error)
	GetByKey(ctx context.Context, key string) (*repository.Task, error)
	ListByProject(ctx context.Context, projectID string, filters *repository.TaskFilters) ([]*repository.Task, error)
	ListBySprint(ctx context.Context, sprintID string) ([]*repository.Task, error)
	ListBacklog(ctx context.Context, projectID string) ([]*repository.Task, error)
	Update(ctx context.Context, id, updaterID string, updates map[string]interface{}) (*repository.Task, error)
	Delete(ctx context.Context, id, deleterID string) error
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
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil || project == nil {
		return nil, ErrNotFound
	}

	taskNum, _ := s.projectRepo.GetNextTaskNumber(ctx, projectID)
	taskKey := fmt.Sprintf("%s-%d", project.Key, taskNum)

	// Use lowercase defaults
	statusVal := "backlog"
	if status != nil && *status != "" {
		statusVal = *status
	}

	priorityVal := "medium"
	if priority != nil && *priority != "" {
		priorityVal = *priority
	}

	typeVal := "task"
	if taskType != nil && *taskType != "" {
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

	// ============================================
	// SEND NOTIFICATIONS
	// ============================================
	if s.notifSvc != nil {
		// 1. Notify assignee if task is assigned to someone else
		if assigneeID != nil && *assigneeID != reporterID {
			s.notifSvc.SendTaskAssigned(ctx, *assigneeID, task.Title, task.ID, projectID)
		}

		// 2. Notify all project members about new task (except the creator)
		memberIDs, _ := s.projectRepo.FindMemberUserIDs(ctx, projectID)
		if len(memberIDs) > 0 {
			s.notifSvc.SendTaskCreated(ctx, memberIDs, reporterID, task.Title, task.Key, task.ID, projectID)
		}
	}

	// Populate user info
	s.populateTaskUsers(ctx, task)

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
	s.populateTaskUsers(ctx, task)
	return task, nil
}

func (s *taskService) GetByKey(ctx context.Context, key string) (*repository.Task, error) {
	task, err := s.taskRepo.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}
	s.populateTaskUsers(ctx, task)
	return task, nil
}

func (s *taskService) ListByProject(ctx context.Context, projectID string, filters *repository.TaskFilters) ([]*repository.Task, error) {
	tasks, err := s.taskRepo.FindByProjectID(ctx, projectID, filters)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		s.populateTaskUsers(ctx, task)
	}
	return tasks, nil
}

func (s *taskService) ListBySprint(ctx context.Context, sprintID string) ([]*repository.Task, error) {
	tasks, err := s.taskRepo.FindBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		s.populateTaskUsers(ctx, task)
	}
	return tasks, nil
}

func (s *taskService) ListBacklog(ctx context.Context, projectID string) ([]*repository.Task, error) {
	tasks, err := s.taskRepo.FindBacklog(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		s.populateTaskUsers(ctx, task)
	}
	return tasks, nil
}

func (s *taskService) Update(ctx context.Context, id, updaterID string, updates map[string]interface{}) (*repository.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil || task == nil {
		return nil, ErrNotFound
	}

	oldAssigneeID := task.AssigneeID
	oldStatus := task.Status
	var changes []string

	// Apply updates and track changes
	if v, ok := updates["title"].(string); ok && v != task.Title {
		task.Title = v
		changes = append(changes, "title")
	}
	if v, ok := updates["description"].(*string); ok {
		task.Description = v
		changes = append(changes, "description")
	}
	if v, ok := updates["status"].(string); ok && v != task.Status {
		task.Status = v
		changes = append(changes, "status")
	}
	if v, ok := updates["priority"].(string); ok && v != task.Priority {
		task.Priority = v
		changes = append(changes, "priority")
	}
	if v, ok := updates["type"].(string); ok && v != task.Type {
		task.Type = v
		changes = append(changes, "type")
	}
	if v, ok := updates["assigneeId"].(*string); ok {
		task.AssigneeID = v
		changes = append(changes, "assignee")
	}
	if v, ok := updates["sprintId"].(*string); ok {
		task.SprintID = v
		changes = append(changes, "sprint")
	}
	if v, ok := updates["storyPoints"].(*int); ok {
		task.StoryPoints = v
		changes = append(changes, "story points")
	}
	if v, ok := updates["dueDate"].(*time.Time); ok {
		task.DueDate = v
		changes = append(changes, "due date")
	}
	if v, ok := updates["orderIndex"].(int); ok {
		task.OrderIndex = v
	}
	if v, ok := updates["labels"].([]string); ok {
		task.Labels = v
		changes = append(changes, "labels")
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// Send notifications
	if s.notifSvc != nil {
		// Notify new assignee if changed
		if task.AssigneeID != nil {
			if oldAssigneeID == nil || *oldAssigneeID != *task.AssigneeID {
				if *task.AssigneeID != updaterID {
					s.notifSvc.SendTaskAssigned(ctx, *task.AssigneeID, task.Title, task.ID, task.ProjectID)
				}
			}
		}

		// Notify about status change
		if task.Status != oldStatus {
			usersToNotify := s.getTaskStakeholders(ctx, task, updaterID)
			s.notifSvc.SendTaskStatusChangedToUsers(ctx, usersToNotify, updaterID, task.Title, task.ID, task.ProjectID, oldStatus, task.Status)
		}

		// Notify about other significant updates (if there are changes besides status)
		if len(changes) > 0 && (len(changes) > 1 || changes[0] != "status") {
			usersToNotify := s.getTaskStakeholders(ctx, task, updaterID)
			if len(usersToNotify) > 0 {
				s.notifSvc.SendTaskUpdatedToUsers(ctx, usersToNotify, updaterID, task.Title, task.ID, task.ProjectID, changes)
			}
		}
	}

	s.populateTaskUsers(ctx, task)
	return task, nil
}

func (s *taskService) Delete(ctx context.Context, id, deleterID string) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil || task == nil {
		return ErrNotFound
	}

	// Notify stakeholders before deletion
	if s.notifSvc != nil {
		usersToNotify := s.getTaskStakeholders(ctx, task, deleterID)
		s.notifSvc.SendTaskDeletedToUsers(ctx, usersToNotify, deleterID, task.Title, task.Key, task.ProjectID)
	}

	return s.taskRepo.Delete(ctx, id)
}

func (s *taskService) BulkUpdate(ctx context.Context, updates []repository.BulkTaskUpdate) error {
	return s.taskRepo.BulkUpdate(ctx, updates)
}

// Helper methods
func (s *taskService) populateTaskUsers(ctx context.Context, task *repository.Task) {
	if task.AssigneeID != nil {
		task.Assignee, _ = s.userRepo.FindByID(ctx, *task.AssigneeID)
	}
	task.Reporter, _ = s.userRepo.FindByID(ctx, task.ReporterID)
}

func (s *taskService) getTaskStakeholders(ctx context.Context, task *repository.Task, excludeUserID string) []string {
	stakeholders := make(map[string]bool)

	// Add assignee
	if task.AssigneeID != nil && *task.AssigneeID != excludeUserID {
		stakeholders[*task.AssigneeID] = true
	}

	// Add reporter
	if task.ReporterID != excludeUserID {
		stakeholders[task.ReporterID] = true
	}

	result := make([]string, 0, len(stakeholders))
	for userID := range stakeholders {
		result = append(result, userID)
	}
	return result
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

	// Get commenter info for notifications
	commenter, _ := s.userRepo.FindByID(ctx, userID)
	commenterName := "Someone"
	if commenter != nil {
		commenterName = commenter.Name
	}

	// Send notifications
	task, _ := s.taskRepo.FindByID(ctx, taskID)
	if task != nil && s.notifSvc != nil {
		// Notify assignee
		if task.AssigneeID != nil && *task.AssigneeID != userID {
			s.notifSvc.SendTaskCommented(ctx, *task.AssigneeID, commenterName, task.Title, task.ID, task.ProjectID)
		}

		// Notify reporter (if different from assignee and commenter)
		if task.ReporterID != userID {
			if task.AssigneeID == nil || *task.AssigneeID != task.ReporterID {
				s.notifSvc.SendTaskCommented(ctx, task.ReporterID, commenterName, task.Title, task.ID, task.ProjectID)
			}
		}

		// Parse and send mention notifications
		s.notifSvc.ParseAndSendMentions(ctx, content, commenterName, task.Title, task.ID, task.ProjectID, userID)
	}

	comment.User = commenter
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
	// Check for duplicate name in project
	existing, _ := s.labelRepo.FindByName(ctx, projectID, name)
	if existing != nil {
		return nil, ErrConflict
	}

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
		// Check for duplicate name
		existing, _ := s.labelRepo.FindByName(ctx, label.ProjectID, *name)
		if existing != nil && existing.ID != id {
			return nil, ErrConflict
		}
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
