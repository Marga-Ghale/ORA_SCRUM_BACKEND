package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// ============================================
// SPRINT ANALYTICS MODELS
// ============================================

type SprintReport struct {
	ID                  string    `json:"id" db:"id"`
	SprintID            string    `json:"sprintId" db:"sprint_id"`
	CommittedTasks      int       `json:"committedTasks" db:"committed_tasks"`
	CommittedPoints     int       `json:"committedPoints" db:"committed_points"`
	CompletedTasks      int       `json:"completedTasks" db:"completed_tasks"`
	CompletedPoints     int       `json:"completedPoints" db:"completed_points"`
	IncompleteTasks     int       `json:"incompleteTasks" db:"incomplete_tasks"`
	IncompletePoints    int       `json:"incompletePoints" db:"incomplete_points"`
	AddedTasks          int       `json:"addedTasks" db:"added_tasks"`
	AddedPoints         int       `json:"addedPoints" db:"added_points"`
	RemovedTasks        int       `json:"removedTasks" db:"removed_tasks"`
	RemovedPoints       int       `json:"removedPoints" db:"removed_points"`
	CarryoverTasks      int       `json:"carryoverTasks" db:"carryover_tasks"`
	CarryoverPoints     int       `json:"carryoverPoints" db:"carryover_points"`
	TotalEstimatedHours float64   `json:"totalEstimatedHours" db:"total_estimated_hours"`
	TotalLoggedHours    float64   `json:"totalLoggedHours" db:"total_logged_hours"`
	AvgCycleTimeHours   *float64  `json:"avgCycleTimeHours,omitempty" db:"avg_cycle_time_hours"`
	AvgLeadTimeHours    *float64  `json:"avgLeadTimeHours,omitempty" db:"avg_lead_time_hours"`
	Velocity            int       `json:"velocity" db:"velocity"`
	GoalsCompleted      int       `json:"goalsCompleted" db:"goals_completed"`
	GoalsTotal          int       `json:"goalsTotal" db:"goals_total"`
	GeneratedAt         time.Time `json:"generatedAt" db:"generated_at"`
}

type VelocityHistory struct {
	ID              string    `json:"id" db:"id"`
	ProjectID       string    `json:"projectId" db:"project_id"`
	SprintID        string    `json:"sprintId" db:"sprint_id"`
	SprintName      string    `json:"sprintName" db:"sprint_name"`
	SprintNumber    int       `json:"sprintNumber" db:"sprint_number"`
	CommittedPoints int       `json:"committedPoints" db:"committed_points"`
	CompletedPoints int       `json:"completedPoints" db:"completed_points"`
	StartDate       time.Time `json:"startDate" db:"start_date"`
	EndDate         time.Time `json:"endDate" db:"end_date"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

type TaskStatusHistory struct {
	ID         string    `json:"id" db:"id"`
	TaskID     string    `json:"taskId" db:"task_id"`
	FromStatus *string   `json:"fromStatus,omitempty" db:"from_status"`
	ToStatus   string    `json:"toStatus" db:"to_status"`
	ChangedBy  *string   `json:"changedBy,omitempty" db:"changed_by"`
	ChangedAt  time.Time `json:"changedAt" db:"changed_at"`
}

type CycleTimeStats struct {
	TaskID           string   `json:"taskId"`
	TaskTitle        string   `json:"taskTitle"`
	CycleTimeSeconds *int     `json:"cycleTimeSeconds,omitempty"` // in_progress -> done
	LeadTimeSeconds  *int     `json:"leadTimeSeconds,omitempty"`  // created -> done
	StartedAt        *time.Time `json:"startedAt,omitempty"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type VelocityTrend struct {
	Sprints         []VelocityHistory `json:"sprints"`
	AverageVelocity float64           `json:"averageVelocity"`
	TrendDirection  string            `json:"trendDirection"` // up, down, stable
	TrendPercentage float64           `json:"trendPercentage"`
}

// ============================================
// GANTT CHART DATA
// ============================================

type GanttTask struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	StartDate     *time.Time `json:"startDate,omitempty"`
	DueDate       *time.Time `json:"dueDate,omitempty"`
	EndDate       *time.Time `json:"endDate,omitempty"` // actual end (completed_at)
	Status        string     `json:"status"`
	Priority      string     `json:"priority"`
	Progress      float64    `json:"progress"` // 0-100 based on subtasks or status
	AssigneeIDs   []string   `json:"assigneeIds"`
	ParentTaskID  *string    `json:"parentTaskId,omitempty"`
	Dependencies  []string   `json:"dependencies"` // task IDs this depends on
	SprintID      *string    `json:"sprintId,omitempty"`
	StoryPoints   *int       `json:"storyPoints,omitempty"`
	Type          *string    `json:"type,omitempty"`
	Color         string     `json:"color"` // derived from priority/status
}

type GanttData struct {
	Tasks      []GanttTask `json:"tasks"`
	StartDate  time.Time   `json:"startDate"`  // earliest task start
	EndDate    time.Time   `json:"endDate"`    // latest task end
	TotalTasks int         `json:"totalTasks"`
}

// ============================================
// REPOSITORY INTERFACE
// ============================================

type SprintAnalyticsRepository interface {
	// Sprint Reports
	GenerateSprintReport(ctx context.Context, sprintID string) (*SprintReport, error)
	GetSprintReport(ctx context.Context, sprintID string) (*SprintReport, error)
	SaveSprintReport(ctx context.Context, report *SprintReport) error

	// Velocity
	GetVelocityHistory(ctx context.Context, projectID string, limit int) ([]*VelocityHistory, error)
	GetVelocityTrend(ctx context.Context, projectID string, sprintCount int) (*VelocityTrend, error)
	SaveVelocityHistory(ctx context.Context, vh *VelocityHistory) error

	// Cycle Time
	GetCycleTimeStats(ctx context.Context, sprintID string) ([]*CycleTimeStats, error)
	GetAverageCycleTime(ctx context.Context, projectID string, days int) (float64, error)
	GetAverageLeadTime(ctx context.Context, projectID string, days int) (float64, error)
	GetTaskStatusHistory(ctx context.Context, taskID string) ([]*TaskStatusHistory, error)

	// Gantt Chart
	GetGanttData(ctx context.Context, projectID string, sprintID *string) (*GanttData, error)
}

// ============================================
// IMPLEMENTATION
// ============================================

type sprintAnalyticsRepository struct {
	db *sql.DB
}

func NewSprintAnalyticsRepository(db *sql.DB) SprintAnalyticsRepository {
	return &sprintAnalyticsRepository{db: db}
}

// ============================================
// SPRINT REPORTS
// ============================================

func (r *sprintAnalyticsRepository) GenerateSprintReport(ctx context.Context, sprintID string) (*SprintReport, error) {
	report := &SprintReport{SprintID: sprintID}

	// Get task stats
	taskStatsQuery := `
		SELECT 
			COUNT(*) as total_tasks,
			COUNT(*) FILTER (WHERE status = 'done') as completed_tasks,
			COUNT(*) FILTER (WHERE status != 'done') as incomplete_tasks,
			COALESCE(SUM(story_points), 0) as total_points,
			COALESCE(SUM(story_points) FILTER (WHERE status = 'done'), 0) as completed_points,
			COALESCE(SUM(story_points) FILTER (WHERE status != 'done'), 0) as incomplete_points,
			COALESCE(SUM(estimated_hours), 0) as estimated_hours,
			COALESCE(SUM(actual_hours), 0) as logged_hours
		FROM tasks
		WHERE sprint_id = $1 AND parent_task_id IS NULL`

	var totalTasks, totalPoints int
	err := r.db.QueryRowContext(ctx, taskStatsQuery, sprintID).Scan(
		&totalTasks,
		&report.CompletedTasks,
		&report.IncompleteTasks,
		&totalPoints,
		&report.CompletedPoints,
		&report.IncompletePoints,
		&report.TotalEstimatedHours,
		&report.TotalLoggedHours,
	)
	if err != nil {
		return nil, err
	}

	report.CommittedTasks = totalTasks
	report.CommittedPoints = totalPoints
	report.Velocity = report.CompletedPoints

	// Get cycle time averages
	cycleTimeQuery := `
		SELECT 
			AVG(cycle_time_seconds) / 3600.0 as avg_cycle_hours,
			AVG(lead_time_seconds) / 3600.0 as avg_lead_hours
		FROM tasks
		WHERE sprint_id = $1 AND status = 'done' AND cycle_time_seconds IS NOT NULL`

	r.db.QueryRowContext(ctx, cycleTimeQuery, sprintID).Scan(
		&report.AvgCycleTimeHours,
		&report.AvgLeadTimeHours,
	)

	// Get goals summary
	goalsQuery := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) as total
		FROM goals
		WHERE sprint_id = $1`

	r.db.QueryRowContext(ctx, goalsQuery, sprintID).Scan(
		&report.GoalsCompleted,
		&report.GoalsTotal,
	)

	report.GeneratedAt = time.Now()

	return report, nil
}

func (r *sprintAnalyticsRepository) GetSprintReport(ctx context.Context, sprintID string) (*SprintReport, error) {
	query := `
		SELECT id, sprint_id, committed_tasks, committed_points, completed_tasks,
			   completed_points, incomplete_tasks, incomplete_points, added_tasks,
			   added_points, removed_tasks, removed_points, carryover_tasks,
			   carryover_points, total_estimated_hours, total_logged_hours,
			   avg_cycle_time_hours, avg_lead_time_hours, velocity,
			   goals_completed, goals_total, generated_at
		FROM sprint_reports
		WHERE sprint_id = $1`

	report := &SprintReport{}
	err := r.db.QueryRowContext(ctx, query, sprintID).Scan(
		&report.ID, &report.SprintID, &report.CommittedTasks, &report.CommittedPoints,
		&report.CompletedTasks, &report.CompletedPoints, &report.IncompleteTasks,
		&report.IncompletePoints, &report.AddedTasks, &report.AddedPoints,
		&report.RemovedTasks, &report.RemovedPoints, &report.CarryoverTasks,
		&report.CarryoverPoints, &report.TotalEstimatedHours, &report.TotalLoggedHours,
		&report.AvgCycleTimeHours, &report.AvgLeadTimeHours, &report.Velocity,
		&report.GoalsCompleted, &report.GoalsTotal, &report.GeneratedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return report, err
}

func (r *sprintAnalyticsRepository) SaveSprintReport(ctx context.Context, report *SprintReport) error {
	query := `
		INSERT INTO sprint_reports (
			sprint_id, committed_tasks, committed_points, completed_tasks,
			completed_points, incomplete_tasks, incomplete_points, added_tasks,
			added_points, removed_tasks, removed_points, carryover_tasks,
			carryover_points, total_estimated_hours, total_logged_hours,
			avg_cycle_time_hours, avg_lead_time_hours, velocity,
			goals_completed, goals_total, generated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW()
		)
		ON CONFLICT (sprint_id) DO UPDATE SET
			committed_tasks = EXCLUDED.committed_tasks,
			committed_points = EXCLUDED.committed_points,
			completed_tasks = EXCLUDED.completed_tasks,
			completed_points = EXCLUDED.completed_points,
			incomplete_tasks = EXCLUDED.incomplete_tasks,
			incomplete_points = EXCLUDED.incomplete_points,
			added_tasks = EXCLUDED.added_tasks,
			added_points = EXCLUDED.added_points,
			removed_tasks = EXCLUDED.removed_tasks,
			removed_points = EXCLUDED.removed_points,
			carryover_tasks = EXCLUDED.carryover_tasks,
			carryover_points = EXCLUDED.carryover_points,
			total_estimated_hours = EXCLUDED.total_estimated_hours,
			total_logged_hours = EXCLUDED.total_logged_hours,
			avg_cycle_time_hours = EXCLUDED.avg_cycle_time_hours,
			avg_lead_time_hours = EXCLUDED.avg_lead_time_hours,
			velocity = EXCLUDED.velocity,
			goals_completed = EXCLUDED.goals_completed,
			goals_total = EXCLUDED.goals_total,
			generated_at = NOW()
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		report.SprintID, report.CommittedTasks, report.CommittedPoints,
		report.CompletedTasks, report.CompletedPoints, report.IncompleteTasks,
		report.IncompletePoints, report.AddedTasks, report.AddedPoints,
		report.RemovedTasks, report.RemovedPoints, report.CarryoverTasks,
		report.CarryoverPoints, report.TotalEstimatedHours, report.TotalLoggedHours,
		report.AvgCycleTimeHours, report.AvgLeadTimeHours, report.Velocity,
		report.GoalsCompleted, report.GoalsTotal,
	).Scan(&report.ID)
}

// ============================================
// VELOCITY
// ============================================

func (r *sprintAnalyticsRepository) GetVelocityHistory(ctx context.Context, projectID string, limit int) ([]*VelocityHistory, error) {
	query := `
		SELECT 
			vh.id, vh.project_id, vh.sprint_id, vh.sprint_name, vh.sprint_number,
			vh.committed_points, vh.completed_points, vh.start_date, vh.end_date, vh.created_at
		FROM velocity_history vh
		WHERE vh.project_id = $1
		ORDER BY vh.end_date DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*VelocityHistory
	for rows.Next() {
		vh := &VelocityHistory{}
		err := rows.Scan(
			&vh.ID, &vh.ProjectID, &vh.SprintID, &vh.SprintName, &vh.SprintNumber,
			&vh.CommittedPoints, &vh.CompletedPoints, &vh.StartDate, &vh.EndDate, &vh.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, vh)
	}

	return history, rows.Err()
}

func (r *sprintAnalyticsRepository) GetVelocityTrend(ctx context.Context, projectID string, sprintCount int) (*VelocityTrend, error) {
	history, err := r.GetVelocityHistory(ctx, projectID, sprintCount)
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		return &VelocityTrend{
			Sprints:         []VelocityHistory{},
			AverageVelocity: 0,
			TrendDirection:  "stable",
			TrendPercentage: 0,
		}, nil
	}

	// Convert to non-pointer slice and reverse to chronological order
	sprints := make([]VelocityHistory, len(history))
	for i, h := range history {
		sprints[len(history)-1-i] = *h
	}

	// Calculate average
	var totalVelocity float64
	for _, s := range sprints {
		totalVelocity += float64(s.CompletedPoints)
	}
	avgVelocity := totalVelocity / float64(len(sprints))

	// Calculate trend (compare first half vs second half)
	trend := &VelocityTrend{
		Sprints:         sprints,
		AverageVelocity: avgVelocity,
		TrendDirection:  "stable",
		TrendPercentage: 0,
	}

	if len(sprints) >= 2 {
		mid := len(sprints) / 2
		var firstHalf, secondHalf float64
		for i, s := range sprints {
			if i < mid {
				firstHalf += float64(s.CompletedPoints)
			} else {
				secondHalf += float64(s.CompletedPoints)
			}
		}
		firstAvg := firstHalf / float64(mid)
		secondAvg := secondHalf / float64(len(sprints)-mid)

		if firstAvg > 0 {
			trend.TrendPercentage = ((secondAvg - firstAvg) / firstAvg) * 100
			if trend.TrendPercentage > 5 {
				trend.TrendDirection = "up"
			} else if trend.TrendPercentage < -5 {
				trend.TrendDirection = "down"
			}
		}
	}

	return trend, nil
}

func (r *sprintAnalyticsRepository) SaveVelocityHistory(ctx context.Context, vh *VelocityHistory) error {
	query := `
		INSERT INTO velocity_history (
			project_id, sprint_id, sprint_name, sprint_number,
			committed_points, completed_points, start_date, end_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (sprint_id) DO UPDATE SET
			committed_points = EXCLUDED.committed_points,
			completed_points = EXCLUDED.completed_points
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		vh.ProjectID, vh.SprintID, vh.SprintName, vh.SprintNumber,
		vh.CommittedPoints, vh.CompletedPoints, vh.StartDate, vh.EndDate,
	).Scan(&vh.ID, &vh.CreatedAt)
}

// ============================================
// CYCLE TIME
// ============================================

func (r *sprintAnalyticsRepository) GetCycleTimeStats(ctx context.Context, sprintID string) ([]*CycleTimeStats, error) {
	query := `
		SELECT 
			t.id, t.title, t.cycle_time_seconds, t.lead_time_seconds,
			t.started_at, t.completed_at, t.created_at
		FROM tasks t
		WHERE t.sprint_id = $1 AND t.status = 'done'
		ORDER BY t.completed_at DESC`

	rows, err := r.db.QueryContext(ctx, query, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*CycleTimeStats
	for rows.Next() {
		s := &CycleTimeStats{}
		err := rows.Scan(
			&s.TaskID, &s.TaskTitle, &s.CycleTimeSeconds, &s.LeadTimeSeconds,
			&s.StartedAt, &s.CompletedAt, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}

func (r *sprintAnalyticsRepository) GetAverageCycleTime(ctx context.Context, projectID string, days int) (float64, error) {
	query := `
		SELECT COALESCE(AVG(cycle_time_seconds), 0) / 3600.0
		FROM tasks
		WHERE project_id = $1 
		  AND status = 'done' 
		  AND cycle_time_seconds IS NOT NULL
		  AND completed_at >= NOW() - INTERVAL '1 day' * $2`

	var avgHours float64
	err := r.db.QueryRowContext(ctx, query, projectID, days).Scan(&avgHours)
	return avgHours, err
}

func (r *sprintAnalyticsRepository) GetAverageLeadTime(ctx context.Context, projectID string, days int) (float64, error) {
	query := `
		SELECT COALESCE(AVG(lead_time_seconds), 0) / 3600.0
		FROM tasks
		WHERE project_id = $1 
		  AND status = 'done' 
		  AND lead_time_seconds IS NOT NULL
		  AND completed_at >= NOW() - INTERVAL '1 day' * $2`

	var avgHours float64
	err := r.db.QueryRowContext(ctx, query, projectID, days).Scan(&avgHours)
	return avgHours, err
}

func (r *sprintAnalyticsRepository) GetTaskStatusHistory(ctx context.Context, taskID string) ([]*TaskStatusHistory, error) {
	query := `
		SELECT id, task_id, from_status, to_status, changed_by, changed_at
		FROM task_status_history
		WHERE task_id = $1
		ORDER BY changed_at ASC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*TaskStatusHistory
	for rows.Next() {
		h := &TaskStatusHistory{}
		err := rows.Scan(&h.ID, &h.TaskID, &h.FromStatus, &h.ToStatus, &h.ChangedBy, &h.ChangedAt)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

// ============================================
// GANTT CHART
// ============================================

func (r *sprintAnalyticsRepository) GetGanttData(ctx context.Context, projectID string, sprintID *string) (*GanttData, error) {
	query := `
		SELECT 
			t.id, t.title, 
			COALESCE(t.start_date, t.created_at) as start_date,
			t.due_date, t.completed_at,
			t.status, t.priority, t.assignee_ids, t.parent_task_id,
			t.sprint_id, t.story_points, t.type,
			ARRAY(
				SELECT depends_on_task_id 
				FROM task_dependencies 
				WHERE task_id = t.id
			) as dependencies
		FROM tasks t
		WHERE t.project_id = $1`

	args := []interface{}{projectID}

	if sprintID != nil {
		query += ` AND t.sprint_id = $2`
		args = append(args, *sprintID)
	}

	query += ` ORDER BY COALESCE(t.start_date, t.created_at) ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []GanttTask
	var minDate, maxDate *time.Time

	for rows.Next() {
		var t GanttTask
		var deps []string
		var startDate time.Time  // ✅ Non-pointer - COALESCE guarantees non-null

		err := rows.Scan(
			&t.ID, &t.Title, &startDate, &t.DueDate, &t.EndDate,  // ✅ startDate is now non-pointer
			&t.Status, &t.Priority, pq.Array(&t.AssigneeIDs), &t.ParentTaskID,
			&t.SprintID, &t.StoryPoints, &t.Type, pq.Array(&deps),
		)
		if err != nil {
			return nil, err
		}

		t.StartDate = &startDate  // ✅ Assign to pointer field
		t.Dependencies = deps
		if t.Dependencies == nil {
			t.Dependencies = []string{}
		}
		if t.AssigneeIDs == nil {
			t.AssigneeIDs = []string{}
		}
		t.Progress = r.calculateTaskProgress(t.Status)
		t.Color = r.getTaskColor(t.Priority, t.Status)

		tasks = append(tasks, t)

		// Track date range - startDate is always set now
		if minDate == nil || startDate.Before(*minDate) {
			sd := startDate  // Create copy for pointer
			minDate = &sd
		}
		if t.DueDate != nil {
			if maxDate == nil || t.DueDate.After(*maxDate) {
				maxDate = t.DueDate
			}
		}
		if t.EndDate != nil {
			if maxDate == nil || t.EndDate.After(*maxDate) {
				maxDate = t.EndDate
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	data := &GanttData{
		Tasks:      tasks,
		TotalTasks: len(tasks),
	}

	if tasks == nil {
		data.Tasks = []GanttTask{}
	}

	if minDate != nil {
		data.StartDate = *minDate
	} else {
		data.StartDate = time.Now()
	}

	if maxDate != nil {
		data.EndDate = *maxDate
	} else {
		data.EndDate = time.Now().AddDate(0, 1, 0)
	}

	return data, nil
}

func (r *sprintAnalyticsRepository) calculateTaskProgress(status string) float64 {
	switch status {
	case "done":
		return 100
	case "in_review":
		return 80
	case "in_progress":
		return 50
	case "todo":
		return 10
	default:
		return 0
	}
}

func (r *sprintAnalyticsRepository) getTaskColor(priority, status string) string {
	if status == "done" {
		return "#22c55e" // green
	}
	switch priority {
	case "urgent":
		return "#ef4444" // red
	case "high":
		return "#f97316" // orange
	case "medium":
		return "#3b82f6" // blue
	case "low":
		return "#6b7280" // gray
	default:
		return "#3b82f6"
	}
}