package notification

import (
	"context"
	"fmt"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// Service handles sending notifications
type Service struct {
	notificationRepo repository.NotificationRepository
}

// NewService creates a new notification service
func NewService(notificationRepo repository.NotificationRepository) *Service {
	return &Service{
		notificationRepo: notificationRepo,
	}
}

// SendTaskAssigned sends a notification when a task is assigned
func (s *Service) SendTaskAssigned(ctx context.Context, userID, taskTitle, taskID, projectID string) error {
	notification := &repository.Notification{
		UserID:  userID,
		Type:    "TASK_ASSIGNED",
		Title:   "Task Assigned",
		Message: fmt.Sprintf("You have been assigned to task: %s", taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId":    taskID,
			"projectId": projectID,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendTaskUpdated sends a notification when a task is updated
func (s *Service) SendTaskUpdated(ctx context.Context, userID, taskTitle, taskID string) error {
	notification := &repository.Notification{
		UserID:  userID,
		Type:    "TASK_UPDATED",
		Title:   "Task Updated",
		Message: fmt.Sprintf("Task has been updated: %s", taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId": taskID,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendTaskCommented sends a notification when a comment is added
func (s *Service) SendTaskCommented(ctx context.Context, userID, taskTitle, taskID string) error {
	notification := &repository.Notification{
		UserID:  userID,
		Type:    "TASK_COMMENTED",
		Title:   "New Comment",
		Message: fmt.Sprintf("New comment on task: %s", taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId": taskID,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendSprintStarted sends a notification when a sprint starts
func (s *Service) SendSprintStarted(ctx context.Context, userID, sprintName, sprintID string) error {
	notification := &repository.Notification{
		UserID:  userID,
		Type:    "SPRINT_STARTED",
		Title:   "Sprint Started",
		Message: fmt.Sprintf("Sprint has started: %s", sprintName),
		Read:    false,
		Data: map[string]interface{}{
			"sprintId": sprintID,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendSprintCompleted sends a notification when a sprint is completed
func (s *Service) SendSprintCompleted(ctx context.Context, userID, sprintName, sprintID string) error {
	notification := &repository.Notification{
		UserID:  userID,
		Type:    "SPRINT_COMPLETED",
		Title:   "Sprint Completed",
		Message: fmt.Sprintf("Sprint has been completed: %s", sprintName),
		Read:    false,
		Data: map[string]interface{}{
			"sprintId": sprintID,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendMention sends a notification when user is mentioned
func (s *Service) SendMention(ctx context.Context, userID, mentionedBy, taskTitle, taskID string) error {
	notification := &repository.Notification{
		UserID:  userID,
		Type:    "MENTION",
		Title:   "You were mentioned",
		Message: fmt.Sprintf("%s mentioned you in: %s", mentionedBy, taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId": taskID,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendDueDateReminder sends a reminder for upcoming due dates
func (s *Service) SendDueDateReminder(ctx context.Context, userID, taskTitle, taskID string, daysUntilDue int) error {
	var message string
	if daysUntilDue == 0 {
		message = fmt.Sprintf("Task is due today: %s", taskTitle)
	} else if daysUntilDue == 1 {
		message = fmt.Sprintf("Task is due tomorrow: %s", taskTitle)
	} else {
		message = fmt.Sprintf("Task is due in %d days: %s", daysUntilDue, taskTitle)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    "DUE_DATE_REMINDER",
		Title:   "Due Date Reminder",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"taskId": taskID,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendProjectInvitation sends a notification when invited to a project
func (s *Service) SendProjectInvitation(ctx context.Context, userID, projectName, projectID string) error {
	notification := &repository.Notification{
		UserID:  userID,
		Type:    "PROJECT_INVITATION",
		Title:   "Project Invitation",
		Message: fmt.Sprintf("You have been invited to project: %s", projectName),
		Read:    false,
		Data: map[string]interface{}{
			"projectId": projectID,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendWorkspaceInvitation sends a notification when invited to a workspace
func (s *Service) SendWorkspaceInvitation(ctx context.Context, userID, workspaceName, workspaceID string) error {
	notification := &repository.Notification{
		UserID:  userID,
		Type:    "WORKSPACE_INVITATION",
		Title:   "Workspace Invitation",
		Message: fmt.Sprintf("You have been invited to workspace: %s", workspaceName),
		Read:    false,
		Data: map[string]interface{}{
			"workspaceId": workspaceID,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendOverdueTaskReminder sends a reminder for overdue tasks
func (s *Service) SendOverdueTaskReminder(ctx context.Context, userID, taskTitle, taskID string) error {
	notification := &repository.Notification{
		UserID:  userID,
		Type:    "DUE_DATE_REMINDER",
		Title:   "Overdue Task",
		Message: fmt.Sprintf("Task is overdue: %s", taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId":   taskID,
			"isOverdue": true,
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}
