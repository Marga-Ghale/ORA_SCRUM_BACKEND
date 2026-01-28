package service

import (
	"context"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// SPRINT ANALYTICS SERVICE INTERFACE
// ============================================

type SprintAnalyticsService interface {
	// Sprint Reports
	GenerateSprintReport(ctx context.Context, sprintID, userID string) (*repository.SprintReport, error)
	GetSprintReport(ctx context.Context, sprintID, userID string) (*repository.SprintReport, error)

	// Velocity
	GetVelocityHistory(ctx context.Context, projectID, userID string, limit int) ([]*repository.VelocityHistory, error)
	GetVelocityTrend(ctx context.Context, projectID, userID string, sprintCount int) (*repository.VelocityTrend, error)
	RecordSprintVelocity(ctx context.Context, sprintID string) error

	// Cycle Time
	GetCycleTimeStats(ctx context.Context, sprintID, userID string) ([]*repository.CycleTimeStats, error)
	GetProjectCycleTimeAvg(ctx context.Context, projectID, userID string, days int) (*CycleTimeAverage, error)
	GetTaskStatusHistory(ctx context.Context, taskID, userID string) ([]*repository.TaskStatusHistory, error)

	// Gantt Chart
	GetGanttData(ctx context.Context, projectID, userID string, sprintID *string) (*repository.GanttData, error)

	// Combined Analytics
	GetSprintAnalyticsDashboard(ctx context.Context, sprintID, userID string) (*SprintAnalyticsDashboard, error)
	GetProjectAnalyticsDashboard(ctx context.Context, projectID, userID string) (*ProjectAnalyticsDashboard, error)
}

// ============================================
// RESPONSE MODELS
// ============================================

type CycleTimeAverage struct {
	ProjectID        string  `json:"projectId"`
	Period           int     `json:"period"` // days
	AvgCycleTimeHrs  float64 `json:"avgCycleTimeHours"`
	AvgLeadTimeHrs   float64 `json:"avgLeadTimeHours"`
	TasksCompleted   int     `json:"tasksCompleted"`
}

type SprintAnalyticsDashboard struct {
	SprintID string `json:"sprintId"`
	
	// Report summary
	Report *repository.SprintReport `json:"report"`
	
	// Progress
	CompletionPercentage float64 `json:"completionPercentage"`
	DaysRemaining        int     `json:"daysRemaining"`
	DaysElapsed          int     `json:"daysElapsed"`
	
	// Velocity
	CurrentVelocity  int     `json:"currentVelocity"`
	ProjectedVelocity int    `json:"projectedVelocity"`
	
	// Cycle Time
	AvgCycleTimeHrs *float64 `json:"avgCycleTimeHours,omitempty"`
	
	// Burndown (already have in task_service, just reference)
	BurndownAvailable bool `json:"burndownAvailable"`
	
	// Goals
	GoalsCompleted int `json:"goalsCompleted"`
	GoalsTotal     int `json:"goalsTotal"`
	
	// Task breakdown
	TasksByStatus map[string]int `json:"tasksByStatus"`
	TasksByPriority map[string]int `json:"tasksByPriority"`
}

type ProjectAnalyticsDashboard struct {
	ProjectID string `json:"projectId"`
	
	// Velocity trend
	VelocityTrend *repository.VelocityTrend `json:"velocityTrend"`
	
	// Averages
	AvgCycleTimeHrs  float64 `json:"avgCycleTimeHours"`
	AvgLeadTimeHrs   float64 `json:"avgLeadTimeHours"`
	
	// Active sprint summary
	ActiveSprintID   *string `json:"activeSprintId,omitempty"`
	ActiveSprintName *string `json:"activeSprintName,omitempty"`
	
	// Overall stats
	TotalTasks       int `json:"totalTasks"`
	CompletedTasks   int `json:"completedTasks"`
	OpenTasks        int `json:"openTasks"`
	OverdueTasks     int `json:"overdueTasks"`
	
	// Team performance (last 30 days)
	TasksCompletedLast30Days int `json:"tasksCompletedLast30Days"`
	PointsCompletedLast30Days int `json:"pointsCompletedLast30Days"`
}

// ============================================
// IMPLEMENTATION
// ============================================

type sprintAnalyticsService struct {
	analyticsRepo repository.SprintAnalyticsRepository
	sprintRepo    repository.SprintRepository
	taskRepo      repository.TaskRepository
	projectRepo   repository.ProjectRepository
	goalRepo      repository.GoalRepository
	memberService MemberService
}

func NewSprintAnalyticsService(
	analyticsRepo repository.SprintAnalyticsRepository,
	sprintRepo repository.SprintRepository,
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
	goalRepo repository.GoalRepository,
	memberService MemberService,
) SprintAnalyticsService {
	return &sprintAnalyticsService{
		analyticsRepo: analyticsRepo,
		sprintRepo:    sprintRepo,
		taskRepo:      taskRepo,
		projectRepo:   projectRepo,
		goalRepo:      goalRepo,
		memberService: memberService,
	}
}

// ============================================
// SPRINT REPORTS
// ============================================

func (s *sprintAnalyticsService) GenerateSprintReport(ctx context.Context, sprintID, userID string) (*repository.SprintReport, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	report, err := s.analyticsRepo.GenerateSprintReport(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// Save the generated report
	if err := s.analyticsRepo.SaveSprintReport(ctx, report); err != nil {
		// Log but don't fail
	}

	return report, nil
}

func (s *sprintAnalyticsService) GetSprintReport(ctx context.Context, sprintID, userID string) (*repository.SprintReport, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Try to get cached report first
	report, err := s.analyticsRepo.GetSprintReport(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// If no cached report or sprint is still active, generate fresh
	if report == nil || sprint.Status == "active" {
		return s.GenerateSprintReport(ctx, sprintID, userID)
	}

	return report, nil
}

// ============================================
// VELOCITY
// ============================================

func (s *sprintAnalyticsService) GetVelocityHistory(ctx context.Context, projectID, userID string, limit int) ([]*repository.VelocityHistory, error) {
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	if limit <= 0 {
		limit = 10 // default
	}

	return s.analyticsRepo.GetVelocityHistory(ctx, projectID, limit)
}

func (s *sprintAnalyticsService) GetVelocityTrend(ctx context.Context, projectID, userID string, sprintCount int) (*repository.VelocityTrend, error) {
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	if sprintCount <= 0 {
		sprintCount = 6 // default to last 6 sprints
	}

	return s.analyticsRepo.GetVelocityTrend(ctx, projectID, sprintCount)
}

func (s *sprintAnalyticsService) RecordSprintVelocity(ctx context.Context, sprintID string) error {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return ErrNotFound
	}

	// Get velocity data
	totalPoints, _ := s.taskRepo.GetSprintVelocity(ctx, sprintID)
	completedPoints, _ := s.taskRepo.GetCompletedStoryPoints(ctx, sprintID)

	// Count sprints in project for sprint number
	sprints, _ := s.sprintRepo.FindByProjectID(ctx, sprint.ProjectID)
	sprintNumber := len(sprints)

	vh := &repository.VelocityHistory{
		ProjectID:       sprint.ProjectID,
		SprintID:        sprintID,
		SprintName:      sprint.Name,
		SprintNumber:    sprintNumber,
		CommittedPoints: totalPoints,
		CompletedPoints: completedPoints,
		StartDate:       sprint.StartDate,
		EndDate:         sprint.EndDate,
	}

	return s.analyticsRepo.SaveVelocityHistory(ctx, vh)
}

// ============================================
// CYCLE TIME
// ============================================

func (s *sprintAnalyticsService) GetCycleTimeStats(ctx context.Context, sprintID, userID string) ([]*repository.CycleTimeStats, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.analyticsRepo.GetCycleTimeStats(ctx, sprintID)
}

func (s *sprintAnalyticsService) GetProjectCycleTimeAvg(ctx context.Context, projectID, userID string, days int) (*CycleTimeAverage, error) {
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	if days <= 0 {
		days = 30 // default
	}

	avgCycle, err := s.analyticsRepo.GetAverageCycleTime(ctx, projectID, days)
	if err != nil {
		return nil, err
	}

	avgLead, err := s.analyticsRepo.GetAverageLeadTime(ctx, projectID, days)
	if err != nil {
		return nil, err
	}

	return &CycleTimeAverage{
		ProjectID:       projectID,
		Period:          days,
		AvgCycleTimeHrs: avgCycle,
		AvgLeadTimeHrs:  avgLead,
	}, nil
}

func (s *sprintAnalyticsService) GetTaskStatusHistory(ctx context.Context, taskID, userID string) ([]*repository.TaskStatusHistory, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, task.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.analyticsRepo.GetTaskStatusHistory(ctx, taskID)
}

// ============================================
// GANTT CHART
// ============================================

func (s *sprintAnalyticsService) GetGanttData(ctx context.Context, projectID, userID string, sprintID *string) (*repository.GanttData, error) {
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.analyticsRepo.GetGanttData(ctx, projectID, sprintID)
}

// ============================================
// COMBINED DASHBOARDS
// ============================================

func (s *sprintAnalyticsService) GetSprintAnalyticsDashboard(ctx context.Context, sprintID, userID string) (*SprintAnalyticsDashboard, error) {
	sprint, err := s.sprintRepo.FindByID(ctx, sprintID)
	if err != nil || sprint == nil {
		return nil, ErrNotFound
	}

	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, sprint.ProjectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Get report
	report, err := s.GetSprintReport(ctx, sprintID, userID)
	if err != nil {
		return nil, err
	}

	// Calculate days
	now := time.Now()
	totalDays := int(sprint.EndDate.Sub(sprint.StartDate).Hours() / 24)
	daysElapsed := int(now.Sub(sprint.StartDate).Hours() / 24)
	daysRemaining := int(sprint.EndDate.Sub(now).Hours() / 24)
	if daysElapsed < 0 {
		daysElapsed = 0
	}
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	// Calculate completion percentage
	completionPct := 0.0
	if report.CommittedPoints > 0 {
		completionPct = float64(report.CompletedPoints) / float64(report.CommittedPoints) * 100
	}

	// Calculate projected velocity based on current pace
	projectedVelocity := report.CompletedPoints
	if daysElapsed > 0 && totalDays > 0 {
		dailyRate := float64(report.CompletedPoints) / float64(daysElapsed)
		projectedVelocity = int(dailyRate * float64(totalDays))
	}

	// Get tasks by status
	tasks, _ := s.taskRepo.FindBySprintID(ctx, sprintID)
	tasksByStatus := make(map[string]int)
	tasksByPriority := make(map[string]int)
	for _, task := range tasks {
		tasksByStatus[task.Status]++
		tasksByPriority[task.Priority]++
	}

	// Get goals
	goalsCompleted, goalsTotal, _ := s.goalRepo.GetSprintGoalsSummary(ctx, sprintID)

	return &SprintAnalyticsDashboard{
		SprintID:             sprintID,
		Report:               report,
		CompletionPercentage: completionPct,
		DaysRemaining:        daysRemaining,
		DaysElapsed:          daysElapsed,
		CurrentVelocity:      report.CompletedPoints,
		ProjectedVelocity:    projectedVelocity,
		AvgCycleTimeHrs:      report.AvgCycleTimeHours,
		BurndownAvailable:    true,
		GoalsCompleted:       goalsCompleted,
		GoalsTotal:           goalsTotal,
		TasksByStatus:        tasksByStatus,
		TasksByPriority:      tasksByPriority,
	}, nil
}

func (s *sprintAnalyticsService) GetProjectAnalyticsDashboard(ctx context.Context, projectID, userID string) (*ProjectAnalyticsDashboard, error) {
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, userID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Get velocity trend
	velocityTrend, _ := s.analyticsRepo.GetVelocityTrend(ctx, projectID, 6)

	// Get cycle time averages
	avgCycle, _ := s.analyticsRepo.GetAverageCycleTime(ctx, projectID, 30)
	avgLead, _ := s.analyticsRepo.GetAverageLeadTime(ctx, projectID, 30)

	// Get active sprint
	activeSprint, _ := s.sprintRepo.FindActiveSprint(ctx, projectID)
	var activeSprintID, activeSprintName *string
	if activeSprint != nil {
		activeSprintID = &activeSprint.ID
		activeSprintName = &activeSprint.Name
	}

	// Get all tasks for stats
	allTasks, _ := s.taskRepo.FindByProjectID(ctx, projectID)
	totalTasks := len(allTasks)
	completedTasks := 0
	openTasks := 0
	overdueTasks := 0
	now := time.Now()

	for _, task := range allTasks {
		if task.Status == "done" {
			completedTasks++
		} else {
			openTasks++
			if task.DueDate != nil && task.DueDate.Before(now) {
				overdueTasks++
			}
		}
	}

	// Last 30 days completed
	tasksLast30 := 0
	pointsLast30 := 0
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	for _, task := range allTasks {
		if task.Status == "done" && task.CompletedAt != nil && task.CompletedAt.After(thirtyDaysAgo) {
			tasksLast30++
			if task.StoryPoints != nil {
				pointsLast30 += *task.StoryPoints
			}
		}
	}

	return &ProjectAnalyticsDashboard{
		ProjectID:                 projectID,
		VelocityTrend:             velocityTrend,
		AvgCycleTimeHrs:           avgCycle,
		AvgLeadTimeHrs:            avgLead,
		ActiveSprintID:            activeSprintID,
		ActiveSprintName:          activeSprintName,
		TotalTasks:                totalTasks,
		CompletedTasks:            completedTasks,
		OpenTasks:                 openTasks,
		OverdueTasks:              overdueTasks,
		TasksCompletedLast30Days:  tasksLast30,
		PointsCompletedLast30Days: pointsLast30,
	}, nil
}