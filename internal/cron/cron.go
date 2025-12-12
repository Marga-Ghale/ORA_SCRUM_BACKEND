package cron

import (
	"context"
	"log"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/robfig/cron/v3"
)

// Scheduler handles all cron jobs
type Scheduler struct {
	cron     *cron.Cron
	services *service.Services
	notifSvc *notification.Service
}

// NewScheduler creates a new cron scheduler
func NewScheduler(services *service.Services, notifSvc *notification.Service) *Scheduler {
	return &Scheduler{
		cron:     cron.New(cron.WithSeconds()),
		services: services,
		notifSvc: notifSvc,
	}
}

// Start initializes and starts all cron jobs
func (s *Scheduler) Start() {
	log.Println("Starting cron scheduler...")

	// Due date reminders - Run every day at 9:00 AM
	_, err := s.cron.AddFunc("0 0 9 * * *", s.sendDueDateReminders)
	if err != nil {
		log.Printf("Error adding due date reminder cron: %v", err)
	}

	// Overdue task notifications - Run every day at 10:00 AM
	_, err = s.cron.AddFunc("0 0 10 * * *", s.sendOverdueNotifications)
	if err != nil {
		log.Printf("Error adding overdue notification cron: %v", err)
	}

	// Sprint ending reminders - Run every day at 9:00 AM
	_, err = s.cron.AddFunc("0 0 9 * * *", s.sendSprintEndingReminders)
	if err != nil {
		log.Printf("Error adding sprint ending reminder cron: %v", err)
	}

	// Clean up old notifications - Run every Sunday at midnight
	_, err = s.cron.AddFunc("0 0 0 * * 0", s.cleanupOldNotifications)
	if err != nil {
		log.Printf("Error adding notification cleanup cron: %v", err)
	}

	// Auto-complete expired sprints - Run every hour
	_, err = s.cron.AddFunc("0 0 * * * *", s.autoCompleteExpiredSprints)
	if err != nil {
		log.Printf("Error adding sprint auto-complete cron: %v", err)
	}

	// Update user status to away - Run every 30 minutes
	_, err = s.cron.AddFunc("0 */30 * * * *", s.updateInactiveUserStatus)
	if err != nil {
		log.Printf("Error adding user status update cron: %v", err)
	}

	s.cron.Start()
	log.Println("Cron scheduler started successfully")
}

// Stop gracefully stops the cron scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping cron scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Cron scheduler stopped")
}

// sendDueDateReminders sends reminders for tasks due soon
func (s *Scheduler) sendDueDateReminders() {
	log.Println("Running due date reminder job...")
	ctx := context.Background()

	// This would need access to task repository directly
	// For now, we'll log that the job ran
	// In production, you'd inject the task repository and fetch tasks due within 24-48 hours

	// Example implementation:
	// tasks, err := s.taskRepo.FindDueSoon(ctx, 48*time.Hour)
	// if err != nil {
	//     log.Printf("Error fetching tasks due soon: %v", err)
	//     return
	// }
	//
	// for _, task := range tasks {
	//     if task.AssigneeID != nil {
	//         daysUntilDue := int(task.DueDate.Sub(time.Now()).Hours() / 24)
	//         s.notifSvc.SendDueDateReminder(ctx, *task.AssigneeID, task.Title, task.ID, daysUntilDue)
	//     }
	// }

	log.Printf("Due date reminder job completed at %s", time.Now().Format(time.RFC3339))
	_ = ctx // suppress unused warning
}

// sendOverdueNotifications sends notifications for overdue tasks
func (s *Scheduler) sendOverdueNotifications() {
	log.Println("Running overdue notification job...")
	ctx := context.Background()

	// Example implementation:
	// tasks, err := s.taskRepo.FindOverdue(ctx)
	// if err != nil {
	//     log.Printf("Error fetching overdue tasks: %v", err)
	//     return
	// }
	//
	// for _, task := range tasks {
	//     if task.AssigneeID != nil {
	//         s.notifSvc.SendOverdueTaskReminder(ctx, *task.AssigneeID, task.Title, task.ID)
	//     }
	// }

	log.Printf("Overdue notification job completed at %s", time.Now().Format(time.RFC3339))
	_ = ctx
}

// sendSprintEndingReminders sends reminders for sprints ending soon
func (s *Scheduler) sendSprintEndingReminders() {
	log.Println("Running sprint ending reminder job...")
	ctx := context.Background()

	// Example implementation:
	// sprints, err := s.sprintRepo.FindEndingSoon(ctx, 24*time.Hour)
	// if err != nil {
	//     log.Printf("Error fetching sprints ending soon: %v", err)
	//     return
	// }
	//
	// for _, sprint := range sprints {
	//     // Get project members and notify them
	//     members, _ := s.projectRepo.FindMembers(ctx, sprint.ProjectID)
	//     for _, member := range members {
	//         s.notifSvc.SendSprintEndingReminder(ctx, member.UserID, sprint.Name, sprint.ID)
	//     }
	// }

	log.Printf("Sprint ending reminder job completed at %s", time.Now().Format(time.RFC3339))
	_ = ctx
}

// cleanupOldNotifications removes old read notifications
func (s *Scheduler) cleanupOldNotifications() {
	log.Println("Running notification cleanup job...")
	ctx := context.Background()

	// Example implementation:
	// Delete read notifications older than 30 days
	// cutoff := time.Now().AddDate(0, 0, -30)
	// err := s.notificationRepo.DeleteOlderThan(ctx, cutoff)
	// if err != nil {
	//     log.Printf("Error cleaning up notifications: %v", err)
	//     return
	// }

	log.Printf("Notification cleanup job completed at %s", time.Now().Format(time.RFC3339))
	_ = ctx
}

// autoCompleteExpiredSprints automatically completes sprints past their end date
func (s *Scheduler) autoCompleteExpiredSprints() {
	log.Println("Running auto-complete expired sprints job...")
	ctx := context.Background()

	// Example implementation:
	// sprints, err := s.sprintRepo.FindExpired(ctx)
	// if err != nil {
	//     log.Printf("Error fetching expired sprints: %v", err)
	//     return
	// }
	//
	// for _, sprint := range sprints {
	//     if sprint.Status == "ACTIVE" {
	//         // Auto-complete with tasks moved to backlog
	//         s.services.Sprint.Complete(ctx, sprint.ID, "backlog")
	//     }
	// }

	log.Printf("Auto-complete expired sprints job completed at %s", time.Now().Format(time.RFC3339))
	_ = ctx
}

// updateInactiveUserStatus marks users as away if inactive
func (s *Scheduler) updateInactiveUserStatus() {
	log.Println("Running user status update job...")
	ctx := context.Background()

	// Example implementation:
	// Update users who haven't made requests in 30 minutes to "AWAY" status
	// inactiveThreshold := time.Now().Add(-30 * time.Minute)
	// err := s.userRepo.UpdateStatusForInactive(ctx, inactiveThreshold, "AWAY")
	// if err != nil {
	//     log.Printf("Error updating user status: %v", err)
	//     return
	// }

	log.Printf("User status update job completed at %s", time.Now().Format(time.RFC3339))
	_ = ctx
}

// Additional cron job functions can be added here:

// Daily digest email - sends summary of activity
func (s *Scheduler) sendDailyDigest() {
	log.Println("Running daily digest job...")
	// Implementation would gather daily activity and send email summary
}

// Weekly report generation
func (s *Scheduler) generateWeeklyReports() {
	log.Println("Running weekly report generation job...")
	// Implementation would generate sprint/project reports
}

// Database maintenance
func (s *Scheduler) performDatabaseMaintenance() {
	log.Println("Running database maintenance job...")
	// Implementation would run VACUUM, analyze tables, etc.
}
