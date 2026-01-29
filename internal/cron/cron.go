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
	cronJob            *cronlib.Cron
	services           *service.Services
	notifSvc           *notification.Service
	taskRepo           repository.TaskRepository
	sprintRepo         repository.SprintRepository
	projectRepo        repository.ProjectRepository
	userRepo           repository.UserRepository
	notificationRepo   repository.NotificationRepository
	sprintAnalyticsSvc service.SprintAnalyticsService
}

// NewSchedulerWithRepos creates a scheduler with repositories
func NewSchedulerWithRepos(
	services *service.Services,
	notifSvc *notification.Service,
	taskRepo repository.TaskRepository,
	sprintRepo repository.SprintRepository,
	projectRepo repository.ProjectRepository,
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepository,
	sprintAnalyticsSvc service.SprintAnalyticsService,
) *Scheduler {
	return &Scheduler{
		cronJob:            cronlib.New(),
		services:           services,
		notifSvc:           notifSvc,
		taskRepo:           taskRepo,
		sprintRepo:         sprintRepo,
		projectRepo:        projectRepo,
		userRepo:           userRepo,
		notificationRepo:   notificationRepo,
		sprintAnalyticsSvc: sprintAnalyticsSvc,
	}
}

// Start runs the cron scheduler
func (s *Scheduler) Start() {
	// Daily 9 AM
	s.cronJob.AddFunc("0 9 * * *", func() {
		log.Println("[Cron] Daily checks starting...")
		s.checkDueDateReminders()
		s.checkOverdueTasks()
		s.checkSprintDeadlines()
	})

	// Hourly
	s.cronJob.AddFunc("0 * * * *", func() {
		log.Println("[Cron] Hourly checks starting...")
		s.checkTasksDueToday()
		s.autoCompleteExpiredSprints()
	})

	// Every 30 minutes: inactive user update
	s.cronJob.AddFunc("*/30 * * * *", func() {
		log.Println("[Cron] Updating user status...")
		s.updateInactiveUserStatus()
	})

	// Weekly Sunday midnight: clean notifications
	s.cronJob.AddFunc("0 0 * * 0", func() {
		log.Println("[Cron] Cleaning up old notifications...")
		s.cleanupOldNotifications()
	})

	// Optional: Daily at 1 AM - generate sprint reports (cached for performance)
	s.cronJob.AddFunc("0 1 * * *", func() {
		log.Println("[Cron] Generating sprint reports...")
		s.generateActiveSprintReports()
	})

	s.cronJob.Start()
	log.Println("[Cron] Scheduler started")
}

func (s *Scheduler) Stop() {
	s.cronJob.Stop()
	log.Println("[Cron] Scheduler stopped")
}

// ------------------- TASK METHODS -------------------

// checkDueDateReminders sends reminders for tasks due in 3 days
func (s *Scheduler) checkDueDateReminders() {
	ctx := context.Background()
	due := time.Now().Add(72 * time.Hour)
	tasks, _, err := s.taskRepo.FindWithFilters(ctx, &repository.TaskFilters{
		DueBefore: &due,
		Overdue:   new(bool), // false
	})
	if err != nil {
		log.Printf("[Cron] Error finding tasks due soon: %v", err)
		return
	}

	now := time.Now()
	sent := 0
	for _, t := range tasks {
		if t.AssigneeIDs == nil || t.DueDate == nil {
			continue
		}
		daysLeft := int(t.DueDate.Sub(now).Hours() / 24)
		if daysLeft >= 0 && daysLeft <= 3 {
			for _, uid := range t.AssigneeIDs {
				if err := s.notifSvc.SendDueDateReminder(ctx, uid, t.Title, t.ID, t.ProjectID, daysLeft); err == nil {
					sent++
				}
			}
		}
	}
	log.Printf("[Cron] Due date reminders sent: %d", sent)
}

// checkOverdueTasks sends reminders for overdue tasks (1-7 days)
func (s *Scheduler) checkOverdueTasks() {
	ctx := context.Background()
	tasks, err := s.taskRepo.FindOverdue(ctx, "")
	if err != nil {
		log.Printf("[Cron] Error finding overdue tasks: %v", err)
		return
	}

	now := time.Now()
	sent := 0
	for _, t := range tasks {
		if t.AssigneeIDs == nil || t.DueDate == nil {
			continue
		}
		daysOverdue := int(now.Sub(*t.DueDate).Hours() / 24)
		if daysOverdue >= 1 && daysOverdue <= 7 {
			for _, uid := range t.AssigneeIDs {
				if err := s.notifSvc.SendOverdueTaskReminder(ctx, uid, t.Title, t.ID, t.ProjectID, daysOverdue); err == nil {
					sent++
				}
			}
		}
	}
	log.Printf("[Cron] Overdue reminders sent: %d", sent)
}

// checkSprintDeadlines sends reminders for sprints ending in 3 days
func (s *Scheduler) checkSprintDeadlines() {
	ctx := context.Background()
	sprints, err := s.sprintRepo.FindSprintsEndingSoon(ctx, 72*time.Hour)
	if err != nil {
		log.Printf("[Cron] Error fetching sprints ending soon: %v", err)
		return
	}

	sent := 0
	for _, sp := range sprints {
		memberIDs, _ := s.projectRepo.FindMemberUserIDs(ctx, sp.ProjectID)
		if err := s.notifSvc.SendSprintEndingToMembers(ctx, memberIDs, sp.Name, sp.ID, sp.ProjectID, int(sp.EndDate.Sub(time.Now()).Hours()/24)); err == nil {
			sent++
		}
	}
	log.Printf("[Cron] Sprint ending notifications sent: %d", sent)
}

// checkTasksDueToday sends urgent reminders for tasks due in 4 hours
func (s *Scheduler) checkTasksDueToday() {
	ctx := context.Background()
	due := time.Now().Add(24 * time.Hour)
	tasks, _, err := s.taskRepo.FindWithFilters(ctx, &repository.TaskFilters{
		DueBefore: &due,
	})
	if err != nil {
		log.Printf("[Cron] Error finding tasks due today: %v", err)
		return
	}

	now := time.Now()
	sent := 0
	for _, t := range tasks {
		if t.AssigneeIDs == nil || t.DueDate == nil || t.Status == "done" || t.Status == "cancelled" {
			continue
		}
		if t.DueDate.Year() == now.Year() && t.DueDate.YearDay() == now.YearDay() {
			hoursLeft := int(t.DueDate.Sub(now).Hours())
			if hoursLeft <= 4 && hoursLeft > 0 {
				for _, uid := range t.AssigneeIDs {
					if err := s.notifSvc.SendDueDateReminder(ctx, uid, t.Title, t.ID, t.ProjectID, 0); err == nil {
						sent++
					}
				}
			}
		}
	}
	log.Printf("[Cron] Hourly due today reminders sent: %d", sent)
}

// autoCompleteExpiredSprints marks expired sprints as completed and records velocity
func (s *Scheduler) autoCompleteExpiredSprints() {
	ctx := context.Background()
	sprints, err := s.sprintRepo.FindExpiredSprints(ctx)
	if err != nil {
		log.Printf("[Cron] Error fetching expired sprints: %v", err)
		return
	}

	for _, sp := range sprints {
		totalPoints, _ := s.taskRepo.GetSprintVelocity(ctx, sp.ID)
		completedPoints, _ := s.taskRepo.GetCompletedStoryPoints(ctx, sp.ID)

		// ✅ Record velocity history BEFORE completing sprint
		if s.sprintAnalyticsSvc != nil {
			if err := s.sprintAnalyticsSvc.RecordSprintVelocity(ctx, sp.ID); err != nil {
				log.Printf("[Cron] Failed to record velocity for sprint %s: %v", sp.ID, err)
			} else {
				log.Printf("[Cron] Recorded velocity for sprint %s: %d/%d points", sp.Name, completedPoints, totalPoints)
			}
		}

		// Update sprint status to completed
		sp.Status = "completed"
		sp.EndDate = time.Now()
		if err := s.sprintRepo.Update(ctx, sp); err != nil {
			log.Printf("[Cron] Error completing sprint %s: %v", sp.ID, err)
			continue
		}

		// Notify project members
		memberIDs, _ := s.projectRepo.FindMemberUserIDs(ctx, sp.ProjectID)
		if len(memberIDs) > 0 {
			s.notifSvc.SendSprintCompletedToMembers(ctx, memberIDs, sp.Name, sp.ID, sp.ProjectID, completedPoints, totalPoints)
		}
		log.Printf("[Cron] Auto-completed sprint %s (%d/%d story points done)", sp.Name, completedPoints, totalPoints)
	}
}

// cleanupOldNotifications deletes read notifications older than 30 days
func (s *Scheduler) cleanupOldNotifications() {
	ctx := context.Background()
	threshold := time.Now().AddDate(0, 0, -30)
	deleted, err := s.notificationRepo.DeleteOlderThan(ctx, threshold, true)
	if err != nil {
		log.Printf("[Cron] Error cleaning notifications: %v", err)
		return
	}
	log.Printf("[Cron] Old notifications deleted: %d", deleted)
}

// updateInactiveUserStatus sets inactive users to away
func (s *Scheduler) updateInactiveUserStatus() {
	ctx := context.Background()
	if err := s.userRepo.UpdateStatusForInactive(ctx, 30*time.Minute); err != nil {
		log.Printf("[Cron] Error updating inactive users: %v", err)
		return
	}
	log.Println("[Cron] User status update complete")
}

// generateActiveSprintReports generates cached reports for active sprints
// This is optional - reports are generated on-demand, but caching them nightly improves dashboard performance
func (s *Scheduler) generateActiveSprintReports() {
	ctx := context.Background()
	
	// Get all active sprints
	sprints, err := s.sprintRepo.FindActiveSprints(ctx)
	if err != nil {
		log.Printf("[Cron] Error fetching active sprints: %v", err)
		return
	}

	generated := 0
	for _, sprint := range sprints {
		// Generate and cache the sprint report
		if s.sprintAnalyticsSvc != nil {
			_, err := s.sprintAnalyticsSvc.GetSprintReport(ctx, sprint.ID, "")
			if err != nil {
				log.Printf("[Cron] Failed to generate report for sprint %s: %v", sprint.ID, err)
				continue
			}
			generated++
		}
	}
	
	log.Printf("[Cron] Generated %d sprint reports", generated)
}