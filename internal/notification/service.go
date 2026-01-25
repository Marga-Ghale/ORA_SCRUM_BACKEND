// ✅ COMPLETE REPLACEMENT: internal/notification/service.go
package notification

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
)

// Notification types
const (
	TypeTaskAssigned          = "TASK_ASSIGNED"
	TypeTaskUpdated           = "TASK_UPDATED"
	TypeTaskCommented         = "TASK_COMMENTED"
	TypeTaskStatusChanged     = "TASK_STATUS_CHANGED"
	TypeTaskDueSoon           = "TASK_DUE_SOON"
	TypeTaskOverdue           = "TASK_OVERDUE"
	TypeSprintStarted         = "SPRINT_STARTED"
	TypeSprintCompleted       = "SPRINT_COMPLETED"
	TypeSprintEnding          = "SPRINT_ENDING"
	TypeMention               = "MENTION"
	TypeProjectInvitation     = "PROJECT_INVITATION"
	TypeWorkspaceInvitation   = "WORKSPACE_INVITATION"
	TypeTaskCreated           = "TASK_CREATED"
	TypeTaskDeleted           = "TASK_DELETED"
	TypeTaskAttachmentAdded   = "TASK_ATTACHMENT_ADDED"
	TypeTaskAttachmentDeleted = "TASK_ATTACHMENT_DELETED"
	TypeChecklistItemComplete = "CHECKLIST_ITEM_COMPLETED"
	TypeDependencyAdded       = "DEPENDENCY_ADDED"
	TypeDependencyBlocking    = "DEPENDENCY_BLOCKING"
	TypeTimeLoggedToTask      = "TIME_LOGGED_TO_TASK"
	TypeSpaceInvitation       = "SPACE_INVITATION"
	TypeFolderInvitation = "FOLDER_INVITATION"

	TypeWorkspaceRoleUpdated = "WORKSPACE_ROLE_UPDATED"
	TypeSpaceRoleUpdated     = "SPACE_ROLE_UPDATED"
	TypeFolderRoleUpdated    = "FOLDER_ROLE_UPDATED"
	TypeProjectRoleUpdated   = "PROJECT_ROLE_UPDATED"


	// ✅ NEW: Chat-related notification types
	TypeChatAddedToChannel   = "CHAT_ADDED_TO_CHANNEL"
	TypeChatRemovedFromChannel = "CHAT_REMOVED_FROM_CHANNEL"
	TypeChatDirectMessage    = "CHAT_DIRECT_MESSAGE"
	TypeChatMention          = "CHAT_MENTION"
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
// Helper: Get User Name by ID
// ============================================
func (s *Service) getUserName(ctx context.Context, userID string) string {
	if s.userRepo == nil || userID == "" {
		return "Someone"
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		log.Printf("⚠️ Failed to fetch user name for ID %s: %v", userID, err)
		return "Someone"
	}

	return user.Name
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
// Task Notifications - ENHANCED
// ============================================

// ✅ ENHANCED: SendTaskCreated with creator name
func (s *Service) SendTaskCreated(ctx context.Context, userIDs []string, creatorID, taskTitle, taskKey, taskID, projectID string) error {
	// ✅ Fetch creator name
	creatorName := s.getUserName(ctx, creatorID)

	var errs []error

	for _, userID := range userIDs {
		if userID == "" || userID == creatorID {
			continue
		}

		notification := &repository.Notification{
			UserID:  userID,
			Type:    TypeTaskCreated,
			Title:   "New Task Created",
			Message: fmt.Sprintf("%s created: %s", creatorName, taskTitle),
			Read:    false,
			Data: map[string]interface{}{
				"taskId":        taskID,
				"taskKey":       taskKey,
				"taskTitle":     taskTitle,
				"projectId":     projectID,
				"createdBy":     creatorID,
				"createdByName": creatorName, // ✅ Added
				"action":        "view_task",
			},
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			errs = append(errs, fmt.Errorf("failed to notify user %s: %w", userID, err))
		} else {
			s.sendWebSocketNotification(notification)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending task created notifications: %v", errs)
	}
	return nil
}

// ✅ ENHANCED: SendTaskAssigned (backward compatible)
func (s *Service) SendTaskAssigned(ctx context.Context, userID, taskTitle, taskID, projectID string) error {
	return s.SendTaskAssignedBy(ctx, userID, "", taskTitle, taskID, projectID)
}

// ✅ NEW: SendTaskAssignedBy with assigner info
func (s *Service) SendTaskAssignedBy(ctx context.Context, userID, assignedByID, taskTitle, taskID, projectID string) error {
	if userID == "" {
		return nil
	}

	// ✅ Fetch assigner name
	assignedByName := s.getUserName(ctx, assignedByID)

	message := fmt.Sprintf("You've been assigned to: %s", taskTitle)
	if assignedByID != "" && assignedByName != "Someone" {
		message = fmt.Sprintf("%s assigned you to: %s", assignedByName, taskTitle)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskAssigned,
		Title:   "Task Assigned",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"taskId":         taskID,
			"taskTitle":      taskTitle,
			"projectId":      projectID,
			"assignedBy":     assignedByID,
			"assignedByName": assignedByName, // ✅ Added
			"action":         "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// ✅ ENHANCED: SendTaskUpdated (backward compatible)
func (s *Service) SendTaskUpdated(ctx context.Context, userID, taskTitle, taskID, projectID string, changes []string) error {
	return s.SendTaskUpdatedBy(ctx, userID, "", taskTitle, taskID, projectID, changes)
}

// ✅ NEW: SendTaskUpdatedBy with updater info
func (s *Service) SendTaskUpdatedBy(ctx context.Context, userID, updatedByID, taskTitle, taskID, projectID string, changes []string) error {
	if userID == "" {
		return nil
	}

	updatedByName := s.getUserName(ctx, updatedByID)

	changeText := "updated"
	if len(changes) > 0 {
		changeText = strings.Join(changes, ", ")
	}

	message := fmt.Sprintf("Task '%s' - %s", taskTitle, changeText)
	if updatedByID != "" && updatedByName != "Someone" {
		message = fmt.Sprintf("%s updated '%s': %s", updatedByName, taskTitle, changeText)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskUpdated,
		Title:   "Task Updated",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"taskId":        taskID,
			"taskTitle":     taskTitle,
			"projectId":     projectID,
			"changes":       changes,
			"updatedBy":     updatedByID,
			"updatedByName": updatedByName, // ✅ Added
			"action":        "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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

		if err := s.SendTaskUpdatedBy(ctx, userID, excludeUserID, taskTitle, taskID, projectID, changes); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending task update notifications: %v", errs)
	}
	return nil
}

// ✅ ENHANCED: SendTaskStatusChanged with changer info
func (s *Service) SendTaskStatusChanged(ctx context.Context, userID, taskTitle, taskID, projectID, oldStatus, newStatus string) error {
	return s.SendTaskStatusChangedBy(ctx, userID, "", taskTitle, taskID, projectID, oldStatus, newStatus)
}

// ✅ NEW: SendTaskStatusChangedBy with changer info
func (s *Service) SendTaskStatusChangedBy(ctx context.Context, userID, changedByID, taskTitle, taskID, projectID, oldStatus, newStatus string) error {
	if userID == "" {
		return nil
	}

	changedByName := s.getUserName(ctx, changedByID)

	message := fmt.Sprintf("'%s' moved from %s to %s", taskTitle, formatStatus(oldStatus), formatStatus(newStatus))
	if changedByID != "" && changedByName != "Someone" {
		message = fmt.Sprintf("%s moved '%s' from %s to %s", changedByName, taskTitle, formatStatus(oldStatus), formatStatus(newStatus))
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskStatusChanged,
		Title:   "Task Status Changed",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"taskId":          taskID,
			"taskTitle":       taskTitle,
			"projectId":       projectID,
			"oldStatus":       oldStatus,
			"newStatus":       newStatus,
			"changedBy":       changedByID,
			"changedByName":   changedByName, // ✅ Added
			"action":          "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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

		if err := s.SendTaskStatusChangedBy(ctx, userID, excludeUserID, taskTitle, taskID, projectID, oldStatus, newStatus); err != nil {
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
		Message: fmt.Sprintf("%s commented on: %s", commenterName, taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId":       taskID,
			"taskTitle":    taskTitle,
			"projectId":    projectID,
			"commentedBy":  commenterName,
			"action":       "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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
			"taskTitle": taskTitle,
			"taskKey":   taskKey,
			"projectId": projectID,
			"action":    "view_project",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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
			"sprintId":   sprintID,
			"sprintName": sprintName,
			"projectId":  projectID,
			"action":     "view_sprint",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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
				"sprintId":   sprintID,
				"sprintName": sprintName,
				"projectId":  projectID,
				"action":     "view_sprint",
			},
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			errs = append(errs, fmt.Errorf("failed to notify user %s: %w", userID, err))
		} else {
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
			"sprintId":   sprintID,
			"sprintName": sprintName,
			"projectId":  projectID,
			"action":     "view_sprint",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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
				"sprintName":     sprintName,
				"projectId":      projectID,
				"completedTasks": completedTasks,
				"totalTasks":     totalTasks,
				"action":         "view_sprint",
			},
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			errs = append(errs, fmt.Errorf("failed to notify user %s: %w", userID, err))
		} else {
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
			"sprintName":    sprintName,
			"projectId":     projectID,
			"daysRemaining": daysRemaining,
			"action":        "view_sprint",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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
		Message: fmt.Sprintf("%s mentioned you in: %s", mentionedBy, taskTitle),
		Read:    false,
		Data: map[string]interface{}{
			"taskId":      taskID,
			"taskTitle":   taskTitle,
			"projectId":   projectID,
			"mentionedBy": mentionedBy,
			"action":      "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// ParseAndSendMentions parses text for @mentions and sends notifications
func (s *Service) ParseAndSendMentions(ctx context.Context, content, authorName, taskTitle, taskID, projectID, authorID string) error {
	if s.userRepo == nil {
		return nil
	}

	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9._]+(?:@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})?)`)
	matches := mentionRegex.FindAllStringSubmatch(content, -1)

	mentionedUsers := make(map[string]bool)
	var errs []error

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		mention := match[1]
		var user *repository.User
		var err error

		if strings.Contains(mention, "@") {
			user, err = s.userRepo.FindByEmail(ctx, mention)
		} else {
			user, err = s.userRepo.FindByName(ctx, mention)
		}

		if err != nil || user == nil {
			continue
		}

		if user.ID == authorID {
			continue
		}

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
		message = fmt.Sprintf("'%s' is due today", taskTitle)
	case 1:
		title = "Task Due Tomorrow"
		message = fmt.Sprintf("'%s' is due tomorrow", taskTitle)
	default:
		title = "Upcoming Due Date"
		message = fmt.Sprintf("'%s' is due in %d days", taskTitle, daysUntilDue)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskDueSoon,
		Title:   title,
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"taskId":       taskID,
			"taskTitle":    taskTitle,
			"projectId":    projectID,
			"daysUntilDue": daysUntilDue,
			"action":       "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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
		message = fmt.Sprintf("'%s' is 1 day overdue", taskTitle)
	} else {
		message = fmt.Sprintf("'%s' is %d days overdue", taskTitle, daysOverdue)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeTaskOverdue,
		Title:   "⚠️ Overdue Task",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"taskId":      taskID,
			"taskTitle":   taskTitle,
			"projectId":   projectID,
			"daysOverdue": daysOverdue,
			"isOverdue":   true,
			"action":      "view_task",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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
			"projectId":   projectID,
			"projectName": projectName,
			"action":      "view_project",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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
			"workspaceId":   workspaceID,
			"workspaceName": workspaceName,
			"action":        "view_workspace",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

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
			s.sendWebSocketNotification(notification)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending batch notifications: %v", errs)
	}
	return nil
}



// SendSpaceInvitation sends a notification when invited to a space
func (s *Service) SendSpaceInvitation(ctx context.Context, userID, spaceName, spaceID, inviterName string) error {
	if userID == "" {
		return nil
	}

	message := fmt.Sprintf("You have been added to space: %s", spaceName)
	if inviterName != "" {
		message = fmt.Sprintf("%s added you to space: %s", inviterName, spaceName)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeSpaceInvitation,
		Title:   "Space Invitation",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"spaceId":   spaceID,
			"spaceName": spaceName,
			"action":    "view_space",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// SendFolderInvitation sends a notification when invited to a folder
func (s *Service) SendFolderInvitation(ctx context.Context, userID, folderName, folderID, inviterName string) error {
	if userID == "" {
		return nil
	}

	message := fmt.Sprintf("You have been added to folder: %s", folderName)
	if inviterName != "" {
		message = fmt.Sprintf("%s added you to folder: %s", inviterName, folderName)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeFolderInvitation,
		Title:   "Folder Invitation",
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"folderId":   folderID,
			"folderName": folderName,
			"action":     "view_folder",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// ============================================
// Helper Functions
// ============================================

// formatStatus converts status codes to human-readable text
func formatStatus(status string) string {
	statusMap := map[string]string{
		"backlog":     "Backlog",
		"todo":        "To Do",
		"in_progress": "In Progress",
		"in_review":   "In Review",
		"done":        "Done",
		"cancelled":   "Cancelled",
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
	return strings.Title(strings.ReplaceAll(status, "_", " "))
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


// ============================================
// Removal Notifications
// ============================================

// SendWorkspaceRemoval notifies user they were removed from workspace
func (s *Service) SendWorkspaceRemoval(
	ctx context.Context,
	userID, workspaceName, workspaceID, removerName string,
) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    "WORKSPACE_REMOVAL",
		Title:   "Removed from Workspace",
		Message: fmt.Sprintf("%s removed you from workspace '%s'", removerName, workspaceName),
		Read:    false,
		Data: map[string]interface{}{
			"workspaceId":   workspaceID,
			"workspaceName": workspaceName,
			"removerName":   removerName,
			"action":        "view_workspace",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// SendSpaceRemoval notifies user they were removed from space
func (s *Service) SendSpaceRemoval(
	ctx context.Context,
	userID, spaceName, spaceID, removerName string,
) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    "SPACE_REMOVAL",
		Title:   "Removed from Space",
		Message: fmt.Sprintf("%s removed you from space '%s'", removerName, spaceName),
		Read:    false,
		Data: map[string]interface{}{
			"spaceId":     spaceID,
			"spaceName":   spaceName,
			"removerName": removerName,
			"action":      "view_space",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// SendFolderRemoval notifies user they were removed from folder
func (s *Service) SendFolderRemoval(
	ctx context.Context,
	userID, folderName, folderID, removerName string,
) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    "FOLDER_REMOVAL",
		Title:   "Removed from Folder",
		Message: fmt.Sprintf("%s removed you from folder '%s'", removerName, folderName),
		Read:    false,
		Data: map[string]interface{}{
			"folderId":    folderID,
			"folderName":  folderName,
			"removerName": removerName,
			"action":      "view_folder",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// SendProjectRemoval notifies user they were removed from project
func (s *Service) SendProjectRemoval(
	ctx context.Context,
	userID, projectName, projectID, removerName string,
) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    "PROJECT_REMOVAL",
		Title:   "Removed from Project",
		Message: fmt.Sprintf("%s removed you from project '%s'", removerName, projectName),
		Read:    false,
		Data: map[string]interface{}{
			"projectId":   projectID,
			"projectName": projectName,
			"removerName": removerName,
			"action":      "view_project",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}




// SendWorkspaceRoleUpdate notifies user their role was updated
func (s *Service) SendWorkspaceRoleUpdate(
	ctx context.Context,
	userID, workspaceName, workspaceID, oldRole, newRole, updaterName string,
) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeWorkspaceRoleUpdated,
		Title:   "Role Updated",
		Message: fmt.Sprintf("%s changed your role in workspace '%s' from %s to %s", updaterName, workspaceName, oldRole, newRole),
		Read:    false,
		Data: map[string]interface{}{
			"workspaceId":   workspaceID,
			"workspaceName": workspaceName,
			"oldRole":       oldRole,
			"newRole":       newRole,
			"updaterName":   updaterName,
			"action":        "view_workspace",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// SendSpaceRoleUpdate notifies user their role was updated
func (s *Service) SendSpaceRoleUpdate(
	ctx context.Context,
	userID, spaceName, spaceID, oldRole, newRole, updaterName string,
) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeSpaceRoleUpdated,
		Title:   "Role Updated",
		Message: fmt.Sprintf("%s changed your role in space '%s' from %s to %s", updaterName, spaceName, oldRole, newRole),
		Read:    false,
		Data: map[string]interface{}{
			"spaceId":     spaceID,
			"spaceName":   spaceName,
			"oldRole":     oldRole,
			"newRole":     newRole,
			"updaterName": updaterName,
			"action":      "view_space",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// SendFolderRoleUpdate notifies user their role was updated
func (s *Service) SendFolderRoleUpdate(
	ctx context.Context,
	userID, folderName, folderID, oldRole, newRole, updaterName string,
) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeFolderRoleUpdated,
		Title:   "Role Updated",
		Message: fmt.Sprintf("%s changed your role in folder '%s' from %s to %s", updaterName, folderName, oldRole, newRole),
		Read:    false,
		Data: map[string]interface{}{
			"folderId":    folderID,
			"folderName":  folderName,
			"oldRole":     oldRole,
			"newRole":     newRole,
			"updaterName": updaterName,
			"action":      "view_folder",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// SendProjectRoleUpdate notifies user their role was updated
func (s *Service) SendProjectRoleUpdate(
	ctx context.Context,
	userID, projectName, projectID, oldRole, newRole, updaterName string,
) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeProjectRoleUpdated,
		Title:   "Role Updated",
		Message: fmt.Sprintf("%s changed your role in project '%s' from %s to %s", updaterName, projectName, oldRole, newRole),
		Read:    false,
		Data: map[string]interface{}{
			"projectId":   projectID,
			"projectName": projectName,
			"oldRole":     oldRole,
			"newRole":     newRole,
			"updaterName": updaterName,
			"action":      "view_project",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}



// ============================================
// Chat Notifications (Slack-like)
// ============================================

// SendChatAddedToChannel notifies user they were added to a channel
func (s *Service) SendChatAddedToChannel(ctx context.Context, userID, channelID, channelName, addedByName, workspaceID string, isDirect bool) error {
	if userID == "" {
		return nil
	}

	var title, message string
	if isDirect {
		title = "New Conversation"
		message = fmt.Sprintf("%s started a conversation with you", addedByName)
	} else {
		title = "Added to Channel"
		message = fmt.Sprintf("%s added you to #%s", addedByName, channelName)
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeChatAddedToChannel,
		Title:   title,
		Message: message,
		Read:    false,
		Data: map[string]interface{}{
			"channelId":   channelID,
			"channelName": channelName,
			"addedBy":     addedByName,
			"isDirect":    isDirect,
			"workspaceId": workspaceID,
			"action":      "view_chat",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// SendChatRemovedFromChannel notifies user they were removed from a channel
func (s *Service) SendChatRemovedFromChannel(ctx context.Context, userID, channelName, removedByName string) error {
	if userID == "" {
		return nil
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeChatRemovedFromChannel,
		Title:   "Removed from Channel",
		Message: fmt.Sprintf("%s removed you from #%s", removedByName, channelName),
		Read:    false,
		Data: map[string]interface{}{
			"channelName": channelName,
			"removedBy":   removedByName,
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// SendChatMention notifies user they were mentioned in chat
func (s *Service) SendChatMention(ctx context.Context, userID, mentionedByName, channelID, channelName, messagePreview string, isDirect bool) error {
	if userID == "" {
		return nil
	}

	var title string
	if isDirect {
		title = "Mentioned in DM"
	} else {
		title = fmt.Sprintf("Mentioned in #%s", channelName)
	}

	// Truncate message preview
	if len(messagePreview) > 100 {
		messagePreview = messagePreview[:97] + "..."
	}

	notification := &repository.Notification{
		UserID:  userID,
		Type:    TypeChatMention,
		Title:   title,
		Message: fmt.Sprintf("%s: %s", mentionedByName, messagePreview),
		Read:    false,
		Data: map[string]interface{}{
			"channelId":   channelID,
			"channelName": channelName,
			"mentionedBy": mentionedByName,
			"isDirect":    isDirect,
			"action":      "view_chat",
		},
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.sendWebSocketNotification(notification)
	return nil
}

// ParseChatMentions parses message for @mentions and sends notifications
func (s *Service) ParseChatMentions(ctx context.Context, content, authorID, authorName, channelID, channelName string, isDirect bool) error {
	if s.userRepo == nil {
		return nil
	}

	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9._]+(?:@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})?)`)
	matches := mentionRegex.FindAllStringSubmatch(content, -1)

	mentionedUsers := make(map[string]bool)
	var errs []error

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		mention := match[1]
		var user *repository.User
		var err error

		if strings.Contains(mention, "@") {
			user, err = s.userRepo.FindByEmail(ctx, mention)
		} else {
			user, err = s.userRepo.FindByName(ctx, mention)
		}

		if err != nil || user == nil {
			continue
		}

		// Don't notify yourself
		if user.ID == authorID {
			continue
		}

		// Don't notify same user twice
		if mentionedUsers[user.ID] {
			continue
		}
		mentionedUsers[user.ID] = true

		if err := s.SendChatMention(ctx, user.ID, authorName, channelID, channelName, content, isDirect); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors sending chat mentions: %v", errs)
	}
	return nil
}