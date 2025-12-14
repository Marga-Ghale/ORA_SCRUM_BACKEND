package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

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
