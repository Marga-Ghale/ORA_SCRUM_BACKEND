package notification

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// Notification types
const (
	TypeTaskAssigned        = "TASK_ASSIGNED"
	TypeTaskUpdated         = "TASK_UPDATED"
	TypeTaskCommented       = "TASK_COMMENTED"
	TypeTaskStatusChanged   = "TASK_STATUS_CHANGED"
	TypeTaskDueSoon         = "TASK_DUE_SOON"
	TypeTaskOverdue         = "TASK_OVERDUE"
	TypeSprintStarted       = "SPRINT_STARTED"
	TypeSprintCompleted     = "SPRINT_COMPLETED"
	TypeSprintEnding        = "SPRINT_ENDING"
	TypeMention             = "MENTION"
	TypeProjectInvitation   = "PROJECT_INVITATION"
	TypeWorkspaceInvitation = "WORKSPACE_INVITATION"
	TypeTaskCreated         = "TASK_CREATED"
	TypeTaskDeleted         = "TASK_DELETED"
)

// Service handles sending notifications
type Service struct {
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
	projectRepo      repository.ProjectRepository
}

// NewService creates a new notification service
func NewService(notificationRepo repository.NotificationRepository) *Service {
	return &Service{
		notificationRepo: notificationRepo,
	}
}

// NewServiceWithRepos creates a notification service with all repositories for advanced features
func NewServiceWithRepos(notificationRepo repository.NotificationRepository, userRepo repository.UserRepository, projectRepo repository.ProjectRepository) *Service {
	return &Service{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		projectRepo:      projectRepo,
	}
}

// SetUserRepo sets the user repository (for dependency injection)
func (s *Service) SetUserRepo(userRepo repository.UserRepository) {
	s.userRepo = userRepo
}

// SetProjectRepo sets the project repository (for dependency injection)
func (s *Service) SetProjectRepo(projectRepo repository.ProjectRepository) {
	s.projectRepo = projectRepo
}

// ============================================
// Task Notifications
// ============================================

// SendTaskAssigned sends a notification when a task is assigned
func (s *Service) SendTaskAssigned(ctx context.Context, userID, taskTitle, taskID, projectID string) error {
	if userID == "" {
		return nil // Skip if no user to notify
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskAssigned,
		Title:   "Task Assigned",
		Message: fmt.Sprintf("You have been assigned to task: %s", taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId":    taskID,
			"projectId": projectID,
			"action":    "view_task",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendTaskCreated sends a notification when a task is created in a project
func (s *Service) SendTaskCreated(ctx context.Context, userIDs []string, creatorID, taskTitle, taskKey, taskID, projectID string) error {
	for _, userID := range userIDs {
		if userID == creatorID {
			continue // Don't notify the creator
		}

		notification := &repository.Notification{
			UserID:  userID,
			Type:    TypeTaskCreated,
			Title:   "New Task Created",
			Message: fmt.Sprintf("New task created: %s (%s)", taskTitle, taskKey),
			Read:    false,
			Data: map[string]interface{}{
				"taskId":    taskID,
				"taskKey":   taskKey,
				"projectId": projectID,
				"action":    "view_task",
			},
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

// SendTaskUpdated sends a notification when a task is updated
func (s *Service) SendTaskUpdated(ctx context.Context, userID, taskTitle, taskID, projectID string, changes []string) error {
	if userID == "" {
		return nil
	}

	changeText := "updated"
	if len(changes) > 0 {
		changeText = strings.Join(changes, ", ") + " changed"
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskUpdated,
		Title:   "Task Updated",
		Message: fmt.Sprintf("Task %s: %s", changeText, taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId":    taskID,
			"projectId": projectID,
			"changes":   changes,
			"action":    "view_task",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendTaskStatusChanged sends a notification when task status changes
func (s *Service) SendTaskStatusChanged(ctx context.Context, userID, taskTitle, taskID, projectID, oldStatus, newStatus string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskStatusChanged,
		Title:   "Task Status Changed",
		Message: fmt.Sprintf("Task '%s' moved from %s to %s", taskTitle, formatStatus(oldStatus), formatStatus(newStatus)),
		Read:    false,
		Data: map[string]interface{}{
			"taskId":    taskID,
			"projectId": projectID,
			"oldStatus": oldStatus,
			"newStatus": newStatus,
			"action":    "view_task",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendTaskCommented sends a notification when a comment is added
func (s *Service) SendTaskCommented(ctx context.Context, userID, taskTitle, taskID string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskCommented,
		Title:   "New Comment",
		Message: fmt.Sprintf("New comment on task: %s", taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId": taskID,
			"action": "view_task",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendTaskDeleted sends a notification when a task is deleted
func (s *Service) SendTaskDeleted(ctx context.Context, userID, taskTitle, taskKey, projectID string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskDeleted,
		Title:   "Task Deleted",
		Message: fmt.Sprintf("Task '%s' (%s) has been deleted", taskTitle, taskKey),
		Read:    false,
		Data: map[string]interface{}{
			"taskKey":   taskKey,
			"projectId": projectID,
			"action":    "view_project",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// ============================================
// Sprint Notifications
// ============================================

// SendSprintStarted sends a notification when a sprint starts
func (s *Service) SendSprintStarted(ctx context.Context, userID, sprintName, sprintID string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeSprintStarted,
		Title:   "Sprint Started",
		Message: fmt.Sprintf("Sprint '%s' has started", sprintName),
		Read:    false,
		Data: map[string]interface{}{
			"sprintId": sprintID,
			"action":   "view_sprint",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendSprintStartedToMembers sends sprint started notification to all project members
func (s *Service) SendSprintStartedToMembers(ctx context.Context, members []string, sprintName, sprintID, projectID string) error {
	for _, userID := range members {
		notification := &repository.Notification{
			UserID:  userID,
			Type:    TypeSprintStarted,
			Title:   "Sprint Started",
			Message: fmt.Sprintf("Sprint '%s' has started! Time to get to work.", sprintName),
			Read:    false,
			Data: map[string]interface{}{
				"sprintId":  sprintID,
				"projectId": projectID,
				"action":    "view_sprint",
			},
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

// SendSprintCompleted sends a notification when a sprint is completed
func (s *Service) SendSprintCompleted(ctx context.Context, userID, sprintName, sprintID string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeSprintCompleted,
		Title:   "Sprint Completed",
		Message: fmt.Sprintf("Sprint '%s' has been completed", sprintName),
		Read:    false,
		Data: map[string]interface{}{
			"sprintId": sprintID,
			"action":   "view_sprint",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendSprintCompletedToMembers sends sprint completed notification to all project members
func (s *Service) SendSprintCompletedToMembers(ctx context.Context, members []string, sprintName, sprintID, projectID string, completedTasks, totalTasks int) error {
	for _, userID := range members {
		notification := &repository.Notification{
			UserID:  userID,
			Type:    TypeSprintCompleted,
			Title:   "Sprint Completed! 🎉",
			Message: fmt.Sprintf("Sprint '%s' completed with %d/%d tasks done", sprintName, completedTasks, totalTasks),
			Read:    false,
			Data: map[string]interface{}{
				"sprintId":       sprintID,
				"projectId":      projectID,
				"completedTasks": completedTasks,
				"totalTasks":     totalTasks,
				"action":         "view_sprint",
			},
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

// SendSprintEnding sends a notification when a sprint is about to end
func (s *Service) SendSprintEnding(ctx context.Context, userID, sprintName, sprintID string, daysRemaining int) error {
	if userID == "" {
		return nil
	}

	var message string
	if daysRemaining == 0 {
		message = fmt.Sprintf("Sprint '%s' ends today!", sprintName)
	} else if daysRemaining == 1 {
		message = fmt.Sprintf("Sprint '%s' ends tomorrow!", sprintName)
	} else {
		message = fmt.Sprintf("Sprint '%s' ends in %d days", sprintName, daysRemaining)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeSprintEnding,
		Title:   "Sprint Ending Soon",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"sprintId":      sprintID,
			"daysRemaining": daysRemaining,
			"action":        "view_sprint",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// ============================================
// Mention Notifications
// ============================================

// SendMention sends a notification when user is mentioned
func (s *Service) SendMention(ctx context.Context, userID, mentionedBy, taskTitle, taskID string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeMention,
		Title:   "You were mentioned",
		Message: fmt.Sprintf("%s mentioned you in task: %s", mentionedBy, taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId":      taskID,
			"mentionedBy": mentionedBy,
			"action":      "view_task",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// ParseAndSendMentions parses text for @mentions and sends notifications
func (s *Service) ParseAndSendMentions(ctx context.Context, content, authorName, taskTitle, taskID string, authorID string) error {
	if s.userRepo == nil {
		return nil // Can't look up users without repo
	}

	// Find all @mentions (e.g., @john.doe or @john)
	mentionRegex := regexp.MustCompile(`@(\w+(?:\.\w+)?)`)
	matches := mentionRegex.FindAllStringSubmatch(content, -1)

	mentionedUsers := make(map[string]bool) // Avoid duplicate notifications

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		username := match[1]

		// Try to find user by name (simplified - in production you'd want a proper username field)
		user, err := s.userRepo.FindByEmail(ctx, username+"@") // Partial match
		if err != nil || user == nil {
			continue
		}

		// Don't notify the author
		if user.ID == authorID {
			continue
		}

		// Avoid duplicate notifications
		if mentionedUsers[user.ID] {
			continue
		}
		mentionedUsers[user.ID] = true

		if err := s.SendMention(ctx, user.ID, authorName, taskTitle, taskID); err != nil {
			return err
		}
	}

	return nil
}

// ============================================
// Due Date Notifications
// ============================================

// SendDueDateReminder sends a reminder for upcoming due dates
func (s *Service) SendDueDateReminder(ctx context.Context, userID, taskTitle, taskID string, daysUntilDue int) error {
	if userID == "" {
		return nil
	}

	var message string
	var title string
	if daysUntilDue == 0 {
		title = "Task Due Today!"
		message = fmt.Sprintf("Task '%s' is due today", taskTitle)
	} else if daysUntilDue == 1 {
		title = "Task Due Tomorrow"
		message = fmt.Sprintf("Task '%s' is due tomorrow", taskTitle)
	} else {
		title = "Upcoming Due Date"
		message = fmt.Sprintf("Task '%s' is due in %d days", taskTitle, daysUntilDue)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskDueSoon,
		Title:   title,
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"taskId":       taskID,
			"daysUntilDue": daysUntilDue,
			"action":       "view_task",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendOverdueTaskReminder sends a reminder for overdue tasks
func (s *Service) SendOverdueTaskReminder(ctx context.Context, userID, taskTitle, taskID string, daysOverdue int) error {
	if userID == "" {
		return nil
	}

	var message string
	if daysOverdue == 1 {
		message = fmt.Sprintf("Task '%s' is 1 day overdue", taskTitle)
	} else {
		message = fmt.Sprintf("Task '%s' is %d days overdue", taskTitle, daysOverdue)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskOverdue,
		Title:   "⚠️ Overdue Task",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"taskId":      taskID,
			"daysOverdue": daysOverdue,
			"isOverdue":   true,
			"action":      "view_task",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// ============================================
// Invitation Notifications
// ============================================

// SendProjectInvitation sends a notification when invited to a project
func (s *Service) SendProjectInvitation(ctx context.Context, userID, projectName, projectID string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeProjectInvitation,
		Title:   "Project Invitation",
		Message: fmt.Sprintf("You have been added to project: %s", projectName),
		Read:    false,
		Data: map[string]interface{}{
			"projectId": projectID,
			"action":    "view_project",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendWorkspaceInvitation sends a notification when invited to a workspace
func (s *Service) SendWorkspaceInvitation(ctx context.Context, userID, workspaceName, workspaceID string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeWorkspaceInvitation,
		Title:   "Workspace Invitation",
		Message: fmt.Sprintf("You have been added to workspace: %s", workspaceName),
		Read:    false,
		Data: map[string]interface{}{
			"workspaceId": workspaceID,
			"action":      "view_workspace",
		},
	}

	return s.notificationRepo.Create(ctx, notification)
}

// ============================================
// Batch Notifications
// ============================================

// SendBatchNotifications sends the same notification to multiple users
func (s *Service) SendBatchNotifications(ctx context.Context, userIDs []string, notificationType, title, message string, data map[string]interface{}) error {
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}

		notification := &repository.Notification{
			UserID:  userID,
			Type:    notificationType,
			Title:   title,
			Message: message,
			Read:    false,
			Data:    data,
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

// ============================================
// Helper Functions
// ============================================

// formatStatus converts status codes to human-readable text
func formatStatus(status string) string {
	statusMap := map[string]string{
		"BACKLOG":     "Backlog",
		"TODO":        "To Do",
		"IN_PROGRESS": "In Progress",
		"IN_REVIEW":   "In Review",
		"DONE":        "Done",
		"CANCELLED":   "Cancelled",
	}

	if formatted, ok := statusMap[status]; ok {
		return formatted
	}
	return status
}
