package cron

import (
	"context"
	"log"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	cronlib "github.com/robfig/cron/v3"
)

// Scheduler handles scheduled tasks
type Scheduler struct {
	cronJob          *cronlib.Cron
	services         *service.Services
	notifSvc         *notification.Service
	taskRepo         repository.TaskRepository
	sprintRepo       repository.SprintRepository
	projectRepo      repository.ProjectRepository
	userRepo         repository.UserRepository
	notificationRepo repository.NotificationRepository
}

// NewScheduler creates a new scheduler
func NewScheduler(services *service.Services, notifSvc *notification.Service) *Scheduler {
	return &Scheduler{
		cronJob:  cronlib.New(),
		services: services,
		notifSvc: notifSvc,
	}
}

// NewSchedulerWithRepos creates a scheduler with direct repository access
func NewSchedulerWithRepos(
	services *service.Services,
	notifSvc *notification.Service,
	taskRepo repository.TaskRepository,
	sprintRepo repository.SprintRepository,
	projectRepo repository.ProjectRepository,
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepository,
) *Scheduler {
	return &Scheduler{
		cronJob:          cronlib.New(),
		services:         services,
		notifSvc:         notifSvc,
		taskRepo:         taskRepo,
		sprintRepo:       sprintRepo,
		projectRepo:      projectRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	// Run every day at 9 AM - Due date reminders
	s.cronJob.AddFunc("0 9 * * *", func() {
		log.Println("[Cron] Running due date reminder check...")
		s.checkDueDateReminders()
	})

	// Run every day at 9 AM - Overdue task check
	s.cronJob.AddFunc("0 9 * * *", func() {
		log.Println("[Cron] Running overdue task check...")
		s.checkOverdueTasks()
	})

	// Run every day at 9 AM - Sprint ending reminders
	s.cronJob.AddFunc("0 9 * * *", func() {
		log.Println("[Cron] Running sprint ending check...")
		s.checkSprintDeadlines()
	})

	// Run every hour - Check for tasks due today (more urgent)
	s.cronJob.AddFunc("0 * * * *", func() {
		log.Println("[Cron] Running hourly due today check...")
		s.checkTasksDueToday()
	})

	// Clean up old notifications - Run every Sunday at midnight
	s.cronJob.AddFunc("0 0 * * 0", func() {
		log.Println("[Cron] Running notification cleanup...")
		s.cleanupOldNotifications()
	})

	// Auto-complete expired sprints - Run every hour
	s.cronJob.AddFunc("0 * * * *", func() {
		log.Println("[Cron] Running auto-complete expired sprints...")
		s.autoCompleteExpiredSprints()
	})

	// Update user status to away - Run every 30 minutes
	s.cronJob.AddFunc("*/30 * * * *", func() {
		log.Println("[Cron] Running user status update...")
		s.updateInactiveUserStatus()
	})

	s.cronJob.Start()
	log.Println("[Cron] Scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cronJob.Stop()
	log.Println("[Cron] Scheduler stopped")
}

// checkDueDateReminders checks for tasks due soon and sends reminders
func (s *Scheduler) checkDueDateReminders() {
	ctx := context.Background()

	if s.taskRepo == nil {
		log.Println("[Cron] Task repository not available for due date check")
		return
	}

	// Get tasks due within 3 days
	tasks, err := s.taskRepo.FindDueSoon(ctx, 72*time.Hour)
	if err != nil {
		log.Printf("[Cron] Error finding tasks due soon: %v", err)
		return
	}

	now := time.Now()
	sentCount := 0

	for _, task := range tasks {
		if task.AssigneeID == nil || task.DueDate == nil {
			continue
		}

		daysUntilDue := int(task.DueDate.Sub(now).Hours() / 24)

		// Send reminder for 0, 1, 2, 3 days before due
		if daysUntilDue >= 0 && daysUntilDue <= 3 {
			if err := s.notifSvc.SendDueDateReminder(ctx, *task.AssigneeID, task.Title, task.ID, task.ProjectID, daysUntilDue); err != nil {
				log.Printf("[Cron] Error sending due date reminder for task %s: %v", task.ID, err)
			} else {
				sentCount++
				log.Printf("[Cron] Sent due date reminder for task %s (due in %d days)", task.Key, daysUntilDue)
			}
		}
	}

	log.Printf("[Cron] Due date check complete: sent %d reminders", sentCount)
}

// checkOverdueTasks checks for overdue tasks and sends reminders
func (s *Scheduler) checkOverdueTasks() {
	ctx := context.Background()

	if s.taskRepo == nil {
		log.Println("[Cron] Task repository not available for overdue check")
		return
	}

	tasks, err := s.taskRepo.FindOverdue(ctx)
	if err != nil {
		log.Printf("[Cron] Error finding overdue tasks: %v", err)
		return
	}

	now := time.Now()
	sentCount := 0

	for _, task := range tasks {
		if task.AssigneeID == nil || task.DueDate == nil {
			continue
		}

		daysOverdue := int(now.Sub(*task.DueDate).Hours() / 24)

		// Only send reminders for recently overdue tasks (1-7 days)
		if daysOverdue >= 1 && daysOverdue <= 7 {
			if err := s.notifSvc.SendOverdueTaskReminder(ctx, *task.AssigneeID, task.Title, task.ID, task.ProjectID, daysOverdue); err != nil {
				log.Printf("[Cron] Error sending overdue reminder for task %s: %v", task.ID, err)
			} else {
				sentCount++
				log.Printf("[Cron] Sent overdue reminder for task %s (%d days overdue)", task.Key, daysOverdue)
			}
		}
	}

	log.Printf("[Cron] Overdue check complete: sent %d reminders", sentCount)
}

// checkSprintDeadlines checks for sprints ending soon
func (s *Scheduler) checkSprintDeadlines() {
	ctx := context.Background()

	if s.sprintRepo == nil || s.projectRepo == nil {
		log.Println("[Cron] Sprint/Project repository not available for sprint deadline check")
		return
	}

	// Get sprints ending within 3 days
	sprints, err := s.sprintRepo.FindEndingSoon(ctx, 72*time.Hour)
	if err != nil {
		log.Printf("[Cron] Error finding sprints ending soon: %v", err)
		return
	}

	now := time.Now()
	sentCount := 0

	for _, sprint := range sprints {
		if sprint.EndDate == nil {
			continue
		}

		daysRemaining := int(sprint.EndDate.Sub(now).Hours() / 24)

		// Get project members
		memberIDs, err := s.projectRepo.FindMemberUserIDs(ctx, sprint.ProjectID)
		if err != nil {
			log.Printf("[Cron] Error getting project members for sprint %s: %v", sprint.ID, err)
			continue
		}

		// Send notification to all members
		if err := s.notifSvc.SendSprintEndingToMembers(ctx, memberIDs, sprint.Name, sprint.ID, sprint.ProjectID, daysRemaining); err != nil {
			log.Printf("[Cron] Error sending sprint ending notification for sprint %s: %v", sprint.ID, err)
		} else {
			sentCount += len(memberIDs)
			log.Printf("[Cron] Sent sprint ending notification for sprint %s (%d days remaining) to %d members", sprint.Name, daysRemaining, len(memberIDs))
		}
	}

	log.Printf("[Cron] Sprint deadline check complete: sent %d notifications", sentCount)
}

// checkTasksDueToday checks for tasks due today and sends urgent reminders
func (s *Scheduler) checkTasksDueToday() {
	ctx := context.Background()

	if s.taskRepo == nil {
		return
	}

	tasks, err := s.taskRepo.FindDueSoon(ctx, 24*time.Hour)
	if err != nil {
		log.Printf("[Cron] Error finding tasks due today: %v", err)
		return
	}

	now := time.Now()
	urgentCount := 0

	for _, task := range tasks {
		if task.AssigneeID == nil || task.DueDate == nil {
			continue
		}

		// Check if task is due today
		if task.DueDate.Year() == now.Year() && task.DueDate.YearDay() == now.YearDay() {
			hoursRemaining := int(task.DueDate.Sub(now).Hours())

			// Send urgent reminder for tasks due in 4 hours or less
			if hoursRemaining <= 4 && hoursRemaining > 0 && task.Status != "DONE" && task.Status != "CANCELLED" {
				if err := s.notifSvc.SendDueDateReminder(ctx, *task.AssigneeID, task.Title, task.ID, task.ProjectID, 0); err != nil {
					log.Printf("[Cron] Error sending urgent reminder for task %s: %v", task.ID, err)
				} else {
					urgentCount++
					log.Printf("[Cron] Sent urgent reminder for task %s (due in %d hours)", task.Key, hoursRemaining)
				}
			}
		}
	}

	if urgentCount > 0 {
		log.Printf("[Cron] Hourly due today check complete: sent %d urgent reminders", urgentCount)
	}
}

// cleanupOldNotifications removes old read notifications
func (s *Scheduler) cleanupOldNotifications() {
	ctx := context.Background()

	if s.notificationRepo == nil {
		log.Println("[Cron] Notification repository not available for cleanup")
		return
	}

	// Delete read notifications older than 30 days
	threshold := time.Now().AddDate(0, 0, -30)
	deleted, err := s.notificationRepo.DeleteOlderThan(ctx, threshold, true)
	if err != nil {
		log.Printf("[Cron] Error cleaning up old notifications: %v", err)
		return
	}

	log.Printf("[Cron] Notification cleanup complete: deleted %d old notifications", deleted)
}

// autoCompleteExpiredSprints automatically completes sprints past their end date
func (s *Scheduler) autoCompleteExpiredSprints() {
	ctx := context.Background()

	if s.sprintRepo == nil || s.projectRepo == nil || s.taskRepo == nil {
		log.Println("[Cron] Repositories not available for auto-complete check")
		return
	}

	sprints, err := s.sprintRepo.FindExpired(ctx)
	if err != nil {
		log.Printf("[Cron] Error finding expired sprints: %v", err)
		return
	}

	completedCount := 0

	for _, sprint := range sprints {
		// Count tasks
		totalTasks, completedTasks, _ := s.taskRepo.CountBySprintID(ctx, sprint.ID)

		// Mark sprint as completed
		now := time.Now()
		sprint.Status = "COMPLETED"
		sprint.EndDate = &now

		if err := s.sprintRepo.Update(ctx, sprint); err != nil {
			log.Printf("[Cron] Error auto-completing sprint %s: %v", sprint.ID, err)
			continue
		}

		completedCount++

		// Notify project members
		memberIDs, _ := s.projectRepo.FindMemberUserIDs(ctx, sprint.ProjectID)
		if len(memberIDs) > 0 {
			s.notifSvc.SendSprintCompletedToMembers(ctx, memberIDs, sprint.Name, sprint.ID, sprint.ProjectID, completedTasks, totalTasks)
		}

		log.Printf("[Cron] Auto-completed expired sprint %s with %d/%d tasks done", sprint.Name, completedTasks, totalTasks)
	}

	if completedCount > 0 {
		log.Printf("[Cron] Auto-complete check complete: completed %d sprints", completedCount)
	}
}

// updateInactiveUserStatus marks users as away if inactive
func (s *Scheduler) updateInactiveUserStatus() {
	ctx := context.Background()

	if s.userRepo == nil {
		log.Println("[Cron] User repository not available for status update")
		return
	}

	// Mark users as AWAY if inactive for more than 30 minutes
	if err := s.userRepo.UpdateStatusForInactive(ctx, 30*time.Minute); err != nil {
		log.Printf("[Cron] Error updating inactive user status: %v", err)
		return
	}

	log.Println("[Cron] User status update complete")
}

// ManualTrigger allows manual triggering of notification checks (for testing)
func (s *Scheduler) ManualTrigger(checkType string) {
	log.Printf("[Cron] Manual trigger: %s", checkType)

	switch checkType {
	case "due_date":
		s.checkDueDateReminders()
	case "overdue":
		s.checkOverdueTasks()
	case "sprint":
		s.checkSprintDeadlines()
	case "due_today":
		s.checkTasksDueToday()
	case "cleanup":
		s.cleanupOldNotifications()
	case "auto_complete":
		s.autoCompleteExpiredSprints()
	case "user_status":
		s.updateInactiveUserStatus()
	case "all":
		s.checkDueDateReminders()
		s.checkOverdueTasks()
		s.checkSprintDeadlines()
		s.checkTasksDueToday()
		s.cleanupOldNotifications()
		s.autoCompleteExpiredSprints()
		s.updateInactiveUserStatus()
	default:
		log.Printf("[Cron] Unknown check type: %s", checkType)
	}
}

// RunOnce runs all checks once immediately (useful for testing or manual runs)
func (s *Scheduler) RunOnce() {
	log.Println("[Cron] Running all checks once...")
	s.ManualTrigger("all")
	log.Println("[Cron] All checks complete")
}
