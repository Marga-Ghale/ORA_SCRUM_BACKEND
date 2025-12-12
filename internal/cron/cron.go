package cron

import (
	"context"
	"log"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/robfig/cron/v3"
)

// Scheduler handles scheduled tasks
type Scheduler struct {
	cron        *cron.Cron
	services    *service.Services
	notifSvc    *notification.Service
	taskRepo    repository.TaskRepository
	sprintRepo  repository.SprintRepository
	projectRepo repository.ProjectRepository
}

// NewScheduler creates a new scheduler
func NewScheduler(services *service.Services, notifSvc *notification.Service) *Scheduler {
	return &Scheduler{
		cron:     cron.New(),
		services: services,
		notifSvc: notifSvc,
	}
}

// NewSchedulerWithRepos creates a scheduler with direct repository access
func NewSchedulerWithRepos(services *service.Services, notifSvc *notification.Service, taskRepo repository.TaskRepository, sprintRepo repository.SprintRepository, projectRepo repository.ProjectRepository) *Scheduler {
	return &Scheduler{
		cron:        cron.New(),
		services:    services,
		notifSvc:    notifSvc,
		taskRepo:    taskRepo,
		sprintRepo:  sprintRepo,
		projectRepo: projectRepo,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	// Run every day at 9 AM - Due date reminders
	s.cron.AddFunc("0 9 * * *", func() {
		log.Println("[Cron] Running due date reminder check...")
		s.checkDueDateReminders()
	})

	// Run every day at 9 AM - Overdue task check
	s.cron.AddFunc("0 9 * * *", func() {
		log.Println("[Cron] Running overdue task check...")
		s.checkOverdueTasks()
	})

	// Run every day at 9 AM - Sprint ending reminders
	s.cron.AddFunc("0 9 * * *", func() {
		log.Println("[Cron] Running sprint ending check...")
		s.checkSprintDeadlines()
	})

	// Run every hour - Check for tasks due today (more urgent)
	s.cron.AddFunc("0 * * * *", func() {
		log.Println("[Cron] Running hourly due today check...")
		s.checkTasksDueToday()
	})

	// Clean up old notifications - Run every Sunday at midnight
	s.cron.AddFunc("0 0 * * 0", func() {
		log.Println("[Cron] Running notification cleanup...")
		s.cleanupOldNotifications()
	})

	// Auto-complete expired sprints - Run every hour
	s.cron.AddFunc("0 * * * *", func() {
		log.Println("[Cron] Running auto-complete expired sprints...")
		s.autoCompleteExpiredSprints()
	})

	// Update user status to away - Run every 30 minutes
	s.cron.AddFunc("*/30 * * * *", func() {
		log.Println("[Cron] Running user status update...")
		s.updateInactiveUserStatus()
	})

	s.cron.Start()
	log.Println("[Cron] Scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
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
	for _, task := range tasks {
		if task.AssigneeID == nil || task.DueDate == nil {
			continue
		}

		daysUntilDue := int(task.DueDate.Sub(now).Hours() / 24)

		// Send reminder for 0, 1, 2, 3 days before due
		if daysUntilDue >= 0 && daysUntilDue <= 3 {
			if err := s.notifSvc.SendDueDateReminder(ctx, *task.AssigneeID, task.Title, task.ID, daysUntilDue); err != nil {
				log.Printf("[Cron] Error sending due date reminder for task %s: %v", task.ID, err)
			} else {
				log.Printf("[Cron] Sent due date reminder for task %s (due in %d days)", task.Key, daysUntilDue)
			}
		}
	}
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
	for _, task := range tasks {
		if task.AssigneeID == nil || task.DueDate == nil {
			continue
		}

		daysOverdue := int(now.Sub(*task.DueDate).Hours() / 24)

		// Only send reminders for recently overdue tasks (1-7 days)
		if daysOverdue >= 1 && daysOverdue <= 7 {
			if err := s.notifSvc.SendOverdueTaskReminder(ctx, *task.AssigneeID, task.Title, task.ID, daysOverdue); err != nil {
				log.Printf("[Cron] Error sending overdue reminder for task %s: %v", task.ID, err)
			} else {
				log.Printf("[Cron] Sent overdue reminder for task %s (%d days overdue)", task.Key, daysOverdue)
			}
		}
	}
}

// checkSprintDeadlines checks for sprints ending soon
func (s *Scheduler) checkSprintDeadlines() {
	ctx := context.Background()

	if s.sprintRepo == nil || s.projectRepo == nil {
		log.Println("[Cron] Sprint/Project repository not available for sprint deadline check")
		return
	}

	// This would need a new method in sprintRepo: FindEndingSoon
	// For now, we'll log that this feature needs the method
	log.Println("[Cron] Sprint deadline check: requires FindEndingSoon method in sprint repository")
	_ = ctx
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
	for _, task := range tasks {
		if task.AssigneeID == nil || task.DueDate == nil {
			continue
		}

		// Check if due today (same calendar day)
		if task.DueDate.Year() == now.Year() && task.DueDate.YearDay() == now.YearDay() {
			hoursRemaining := int(task.DueDate.Sub(now).Hours())

			// Send reminder if less than 4 hours remaining and task not done
			if hoursRemaining <= 4 && hoursRemaining > 0 && task.Status != "DONE" {
				log.Printf("[Cron] Task %s due in %d hours - would send urgent reminder", task.Key, hoursRemaining)
			}
		}
	}
}

// cleanupOldNotifications removes old read notifications
func (s *Scheduler) cleanupOldNotifications() {
	ctx := context.Background()

	// Delete read notifications older than 30 days
	// This requires a method in notification repository: DeleteOlderThan
	log.Println("[Cron] Notification cleanup: requires DeleteOlderThan method in notification repository")
	_ = ctx
}

// autoCompleteExpiredSprints automatically completes sprints past their end date
func (s *Scheduler) autoCompleteExpiredSprints() {
	ctx := context.Background()

	if s.sprintRepo == nil {
		log.Println("[Cron] Sprint repository not available for auto-complete check")
		return
	}

	// This requires FindExpired method in sprint repository
	log.Println("[Cron] Auto-complete expired sprints: requires FindExpired method in sprint repository")
	_ = ctx
}

// updateInactiveUserStatus marks users as away if inactive
func (s *Scheduler) updateInactiveUserStatus() {
	ctx := context.Background()

	// This requires UpdateStatusForInactive method in user repository
	log.Println("[Cron] User status update: requires UpdateStatusForInactive method in user repository")
	_ = ctx
}

// ManualTrigger allows manual triggering of notification checks (for testing)
func (s *Scheduler) ManualTrigger(checkType string) {
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
	}
}
