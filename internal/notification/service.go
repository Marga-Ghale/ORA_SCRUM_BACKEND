package notification

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
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
	broadcaster      *socket.Broadcaster
}

func (s *Service) SetBroadcaster(b *socket.Broadcaster) {
	s.broadcaster = b
}

// NewService creates a new notification service
func NewService(notificationRepo repository.NotificationRepository) *Service {
	return &Service{
		notificationRepo: notificationRepo,
	}
}

// NewServiceWithRepos creates a notification service with all repositories for advanced features
func NewServiceWithRepos(
	notificationRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	projectRepo repository.ProjectRepository,
) *Service {
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
// WebSocket Helper
// ============================================

// sendWebSocketNotification sends real-time notification via WebSocket
func (s *Service) sendWebSocketNotification(notification *repository.Notification) {
	if s.broadcaster == nil || notification == nil {
		return
	}

	s.broadcaster.SendNotification(notification.UserID, map[string]interface{}{
		"id":        notification.ID,
		"type":      notification.Type,
		"title":     notification.Title,
		"message":   notification.Message,
		"data":      notification.Data,
		"read":      notification.Read,
		"createdAt": notification.CreatedAt,
	})
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

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// SendTaskCreated sends a notification when a task is created in a project
func (s *Service) SendTaskCreated(ctx context.Context, userIDs []string, creatorID, taskTitle, taskKey, taskID, projectID string) error {
	var errs []error

	for _, userID := range userIDs {
		if userID == "" || userID == creatorID {
			continue // Don't notify the creator or empty IDs
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
			errs = append(errs, fmt.Errorf("failed to notify user %s: %w", userID, err))
		} else {
			// Send real-time WebSocket notification
			s.sendWebSocketNotification(notification)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending task created notifications: %v", errs)
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

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// SendTaskUpdatedToUsers sends task update notifications to multiple users
func (s *Service) SendTaskUpdatedToUsers(ctx context.Context, userIDs []string, excludeUserID, taskTitle, taskID, projectID string, changes []string) error {
	var errs []error

	for _, userID := range userIDs {
		if userID == "" || userID == excludeUserID {
			continue
		}

		if err := s.SendTaskUpdated(ctx, userID, taskTitle, taskID, projectID, changes); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending task update notifications: %v", errs)
	}
	return nil
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

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// SendTaskStatusChangedToUsers sends status change notifications to multiple users
func (s *Service) SendTaskStatusChangedToUsers(ctx context.Context, userIDs []string, excludeUserID, taskTitle, taskID, projectID, oldStatus, newStatus string) error {
	var errs []error

	for _, userID := range userIDs {
		if userID == "" || userID == excludeUserID {
			continue
		}

		if err := s.SendTaskStatusChanged(ctx, userID, taskTitle, taskID, projectID, oldStatus, newStatus); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending status change notifications: %v", errs)
	}
	return nil
}

// SendTaskCommented sends a notification when a comment is added
func (s *Service) SendTaskCommented(ctx context.Context, userID, commenterName, taskTitle, taskID, projectID string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskCommented,
		Title:   "New Comment",
		Message: fmt.Sprintf("%s commented on task: %s", commenterName, taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId":    taskID,
			"projectId": projectID,
			"action":    "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
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

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// SendTaskDeletedToUsers sends task deleted notifications to multiple users
func (s *Service) SendTaskDeletedToUsers(ctx context.Context, userIDs []string, excludeUserID, taskTitle, taskKey, projectID string) error {
	var errs []error

	for _, userID := range userIDs {
		if userID == "" || userID == excludeUserID {
			continue
		}

		if err := s.SendTaskDeleted(ctx, userID, taskTitle, taskKey, projectID); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending task deleted notifications: %v", errs)
	}
	return nil
}

// ============================================
// Sprint Notifications
// ============================================

// SendSprintStarted sends a notification when a sprint starts
func (s *Service) SendSprintStarted(ctx context.Context, userID, sprintName, sprintID, projectID string) error {
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
			"sprintId":  sprintID,
			"projectId": projectID,
			"action":    "view_sprint",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// SendSprintStartedToMembers sends sprint started notification to all project members
func (s *Service) SendSprintStartedToMembers(ctx context.Context, members []string, sprintName, sprintID, projectID string) error {
	var errs []error

	for _, userID := range members {
		if userID == "" {
			continue
		}

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
			errs = append(errs, fmt.Errorf("failed to notify user %s: %w", userID, err))
		} else {
			// Send real-time WebSocket notification
			s.sendWebSocketNotification(notification)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending sprint started notifications: %v", errs)
	}
	return nil
}

// SendSprintCompleted sends a notification when a sprint is completed
func (s *Service) SendSprintCompleted(ctx context.Context, userID, sprintName, sprintID, projectID string) error {
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
			"sprintId":  sprintID,
			"projectId": projectID,
			"action":    "view_sprint",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// SendSprintCompletedToMembers sends sprint completed notification to all project members
func (s *Service) SendSprintCompletedToMembers(ctx context.Context, members []string, sprintName, sprintID, projectID string, completedTasks, totalTasks int) error {
	var errs []error

	for _, userID := range members {
		if userID == "" {
			continue
		}

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
			errs = append(errs, fmt.Errorf("failed to notify user %s: %w", userID, err))
		} else {
			// Send real-time WebSocket notification
			s.sendWebSocketNotification(notification)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending sprint completed notifications: %v", errs)
	}
	return nil
}

// SendSprintEnding sends a notification when a sprint is about to end
func (s *Service) SendSprintEnding(ctx context.Context, userID, sprintName, sprintID, projectID string, daysRemaining int) error {
	if userID == "" {
		return nil
	}

	var message string
	switch daysRemaining {
	case 0:
		message = fmt.Sprintf("Sprint '%s' ends today!", sprintName)
	case 1:
		message = fmt.Sprintf("Sprint '%s' ends tomorrow!", sprintName)
	default:
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
			"projectId":     projectID,
			"daysRemaining": daysRemaining,
			"action":        "view_sprint",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// SendSprintEndingToMembers sends sprint ending notification to all project members
func (s *Service) SendSprintEndingToMembers(ctx context.Context, members []string, sprintName, sprintID, projectID string, daysRemaining int) error {
	var errs []error

	for _, userID := range members {
		if userID == "" {
			continue
		}

		if err := s.SendSprintEnding(ctx, userID, sprintName, sprintID, projectID, daysRemaining); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending sprint ending notifications: %v", errs)
	}
	return nil
}

// ============================================
// Mention Notifications
// ============================================

// SendMention sends a notification when user is mentioned
func (s *Service) SendMention(ctx context.Context, userID, mentionedBy, taskTitle, taskID, projectID string) error {
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
			"projectId":   projectID,
			"mentionedBy": mentionedBy,
			"action":      "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// ParseAndSendMentions parses text for @mentions and sends notifications
func (s *Service) ParseAndSendMentions(ctx context.Context, content, authorName, taskTitle, taskID, projectID, authorID string) error {
	if s.userRepo == nil {
		return nil // Can't look up users without repo
	}

	// Find all @mentions (e.g., @john.doe or @john or @user@email.com)
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9._]+(?:@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})?)`)
	matches := mentionRegex.FindAllStringSubmatch(content, -1)

	mentionedUsers := make(map[string]bool) // Avoid duplicate notifications
	var errs []error

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		mention := match[1]
		var user *repository.User
		var err error

		// Check if it's an email mention
		if strings.Contains(mention, "@") {
			user, err = s.userRepo.FindByEmail(ctx, mention)
		} else {
			user, err = s.userRepo.FindByName(ctx, mention)
		}

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

		if err := s.SendMention(ctx, user.ID, authorName, taskTitle, taskID, projectID); err != nil {
			errs = append(errs, fmt.Errorf("failed to send mention to user %s: %w", user.ID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending mentions: %v", errs)
	}
	return nil
}

// ============================================
// Due Date Notifications
// ============================================

// SendDueDateReminder sends a reminder for upcoming due dates
func (s *Service) SendDueDateReminder(ctx context.Context, userID, taskTitle, taskID, projectID string, daysUntilDue int) error {
	if userID == "" {
		return nil
	}

	var message string
	var title string
	switch daysUntilDue {
	case 0:
		title = "Task Due Today!"
		message = fmt.Sprintf("Task '%s' is due today", taskTitle)
	case 1:
		title = "Task Due Tomorrow"
		message = fmt.Sprintf("Task '%s' is due tomorrow", taskTitle)
	default:
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
			"projectId":    projectID,
			"daysUntilDue": daysUntilDue,
			"action":       "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// SendOverdueTaskReminder sends a reminder for overdue tasks
func (s *Service) SendOverdueTaskReminder(ctx context.Context, userID, taskTitle, taskID, projectID string, daysOverdue int) error {
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
			"projectId":   projectID,
			"daysOverdue": daysOverdue,
			"isOverdue":   true,
			"action":      "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// ============================================
// Invitation Notifications
// ============================================

// SendProjectInvitation sends a notification when invited to a project
func (s *Service) SendProjectInvitation(ctx context.Context, userID, projectName, projectID, inviterName string) error {
	if userID == "" {
		return nil
	}

	message := fmt.Sprintf("You have been added to project: %s", projectName)
	if inviterName != "" {
		message = fmt.Sprintf("%s added you to project: %s", inviterName, projectName)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeProjectInvitation,
		Title:   "Project Invitation",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"projectId": projectID,
			"action":    "view_project",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// SendWorkspaceInvitation sends a notification when invited to a workspace
func (s *Service) SendWorkspaceInvitation(ctx context.Context, userID, workspaceName, workspaceID, inviterName string) error {
	if userID == "" {
		return nil
	}

	message := fmt.Sprintf("You have been added to workspace: %s", workspaceName)
	if inviterName != "" {
		message = fmt.Sprintf("%s added you to workspace: %s", inviterName, workspaceName)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeWorkspaceInvitation,
		Title:   "Workspace Invitation",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"workspaceId": workspaceID,
			"action":      "view_workspace",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send real-time WebSocket notification
	s.sendWebSocketNotification(notification)

	return nil
}

// ============================================
// Batch Notifications
// ============================================

// SendBatchNotifications sends the same notification to multiple users
func (s *Service) SendBatchNotifications(ctx context.Context, userIDs []string, excludeUserID, notificationType, title, message string, data map[string]interface{}) error {
	var errs []error

	for _, userID := range userIDs {
		if userID == "" || userID == excludeUserID {
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
			errs = append(errs, fmt.Errorf("failed to notify user %s: %w", userID, err))
		} else {
			// Send real-time WebSocket notification
			s.sendWebSocketNotification(notification)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending batch notifications: %v", errs)
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

// GetProjectMemberIDs returns user IDs of project members (helper for services)
func (s *Service) GetProjectMemberIDs(ctx context.Context, projectID string) ([]string, error) {
	if s.projectRepo == nil {
		return nil, fmt.Errorf("project repository not available")
	}

	members, err := s.projectRepo.FindMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	return userIDs, nil
}
