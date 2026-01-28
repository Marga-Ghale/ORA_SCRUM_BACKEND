package models

import "time"

// TaskResponse is the API response model
type TaskResponse struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Description    *string    `json:"description,omitempty"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	Type           *string    `json:"type,omitempty"`
	ProjectID      string     `json:"projectId"`
	SprintID       *string    `json:"sprintId,omitempty"`
	ParentTaskID   *string    `json:"parentTaskId,omitempty"`
	AssigneeIDs    []string   `json:"assigneeIds"`
	WatcherIDs     []string   `json:"watcherIds"`
	LabelIDs       []string   `json:"labelIds"`
	StoryPoints    *int       `json:"storyPoints,omitempty"`
	EstimatedHours *float64   `json:"estimatedHours,omitempty"`
	ActualHours    *float64   `json:"actualHours,omitempty"`
	StartDate      *time.Time `json:"startDate,omitempty"`
	DueDate        *time.Time `json:"dueDate,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	Blocked        bool       `json:"blocked"`
	Position       int        `json:"position"`
	CreatedBy      *string    `json:"createdBy,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	SubtaskCount   int        `json:"subtaskCount"`
	Subtasks       []TaskResponse `json:"subtasks,omitempty"`
	
	// ✅ Cycle Time Tracking Fields
	StartedAt        *time.Time `json:"startedAt,omitempty"`
	CycleTimeSeconds *int       `json:"cycleTimeSeconds,omitempty"`  // Changed from *int64 to *int
	LeadTimeSeconds  *int       `json:"leadTimeSeconds,omitempty"`   // Changed from *int64 to *int
}

// CreateTaskRequest for creating tasks
type CreateTaskRequest struct {
	ProjectID      string
	SprintID       *string    `json:"sprintId,omitempty"`
	ParentTaskID   *string    `json:"parentTaskId,omitempty"`
	Title          string     `json:"title" binding:"required"`
	Description    *string    `json:"description,omitempty"`
	Status         string     `json:"status,omitempty"`
	Priority       string     `json:"priority,omitempty"`
	Type           *string    `json:"type,omitempty"`
	AssigneeIDs    []string   `json:"assigneeIds,omitempty"`
	LabelIDs       []string   `json:"labelIds,omitempty"`
	EstimatedHours *float64   `json:"estimatedHours,omitempty"`
	StoryPoints    *int       `json:"storyPoints,omitempty"`
	StartDate      *time.Time `json:"startDate,omitempty"`
	DueDate        *time.Time `json:"dueDate,omitempty"`
	CreatedBy      *string
	Subtasks       []SubtaskRequest `json:"subtasks,omitempty"` 
}

type SubtaskRequest struct {
	Title          string   `json:"title" binding:"required"`
	Description    *string  `json:"description"`
	Status         string   `json:"status"`
	Priority       string   `json:"priority"`
	AssigneeIDs    []string `json:"assigneeIds"`
	EstimatedHours *float64 `json:"estimatedHours"`
	StoryPoints    *int     `json:"storyPoints"`
}
// UpdateTaskRequest for updating tasks
type UpdateTaskRequest struct {
	Title          *string    `json:"title,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Status         *string    `json:"status,omitempty"`
	Priority       *string    `json:"priority,omitempty"`
	Type           *string    `json:"type,omitempty"`
	SprintID       *string    `json:"sprintId,omitempty"`
	AssigneeIDs    *[]string  `json:"assigneeIds,omitempty"`
	LabelIDs       *[]string  `json:"labelIds,omitempty"`
	EstimatedHours *float64   `json:"estimatedHours,omitempty"`
	ActualHours    *float64   `json:"actualHours,omitempty"`
	StoryPoints    *int       `json:"storyPoints,omitempty"`
	StartDate      *time.Time `json:"startDate,omitempty"`
	DueDate        *time.Time `json:"dueDate,omitempty"`
}

// Comment models
type CreateCommentRequest struct {
	Content        string   `json:"content" binding:"required"`
	MentionedUsers []string `json:"mentionedUsers,omitempty"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type CommentResponse struct {
	ID             string    `json:"id"`
	TaskID         string    `json:"taskId"`
	UserID         string    `json:"userId"`
	Content        string    `json:"content"`
	MentionedUsers []string  `json:"mentionedUsers"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// Attachment models
type CreateAttachmentRequest struct {
	Filename string `json:"filename" binding:"required"`
	FileURL  string `json:"fileUrl" binding:"required"`
	FileSize int64  `json:"fileSize" binding:"required"`
	MimeType string `json:"mimeType" binding:"required"`
}

type AttachmentResponse struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"taskId"`
	UserID    string    `json:"userId"`
	Filename  string    `json:"filename"`
	FileURL   string    `json:"fileUrl"`
	FileSize  int64     `json:"fileSize"`
	MimeType  string    `json:"mimeType"`
	CreatedAt time.Time `json:"createdAt"`
}

// Time tracking models
type LogTimeRequest struct {
	DurationSeconds int     `json:"durationSeconds" binding:"required"`
	Description     *string `json:"description,omitempty"`
}

type TimeEntryResponse struct {
	ID              string     `json:"id"`
	TaskID          string     `json:"taskId"`
	UserID          string     `json:"userId"`
	StartTime       time.Time  `json:"startTime"`
	EndTime         *time.Time `json:"endTime,omitempty"`
	DurationSeconds *int       `json:"durationSeconds,omitempty"`
	Description     *string    `json:"description,omitempty"`
	IsManual        bool       `json:"isManual"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// Dependency models
type CreateDependencyRequest struct {
	DependsOnTaskID string `json:"dependsOnTaskId" binding:"required"`
	DependencyType  string `json:"dependencyType" binding:"required"`
}

type DependencyResponse struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"taskId"`
	DependsOnTaskID string    `json:"dependsOnTaskId"`
	DependencyType  string    `json:"dependencyType"`
	CreatedAt       time.Time `json:"createdAt"`
}

// Checklist models
type CreateChecklistRequest struct {
	Title string `json:"title" binding:"required"`
}

type CreateChecklistItemRequest struct {
	Content    string  `json:"content" binding:"required"`
	AssigneeID *string `json:"assigneeId,omitempty"`
}

type ChecklistItemResponse struct {
	ID          string    `json:"id"`
	ChecklistID string    `json:"checklistId"`
	Content     string    `json:"content"`
	Completed   bool      `json:"completed"`
	Position    int       `json:"position"`
	AssigneeID  *string   `json:"assigneeId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ChecklistResponse struct {
	ID        string                  `json:"id"`
	TaskID    string                  `json:"taskId"`
	Title     string                  `json:"title"`
	Items     []ChecklistItemResponse `json:"items"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

// Activity models
type ActivityResponse struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"taskId"`
	UserID    *string   `json:"userId,omitempty"`
	Action    string    `json:"action"`
	FieldName *string   `json:"fieldName,omitempty"`
	OldValue  *string   `json:"oldValue,omitempty"`
	NewValue  *string   `json:"newValue,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// Filter models
type TaskFiltersRequest struct {
	ProjectID   string     `json:"projectId" binding:"required"`
	SprintID    *string    `json:"sprintId,omitempty"`
	AssigneeIDs []string   `json:"assigneeIds,omitempty"`
	Statuses    []string   `json:"statuses,omitempty"`
	Priorities  []string   `json:"priorities,omitempty"`
	LabelIDs    []string   `json:"labelIds,omitempty"`
	SearchQuery *string    `json:"searchQuery,omitempty"`
	DueBefore   *time.Time `json:"dueBefore,omitempty"`
	DueAfter    *time.Time `json:"dueAfter,omitempty"`
	Overdue     *bool      `json:"overdue,omitempty"`
	Blocked     *bool      `json:"blocked,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// Bulk operation models
type BulkUpdateStatusRequest struct {
	TaskIDs []string `json:"taskIds" binding:"required"`
	Status  string   `json:"status" binding:"required"`
}

type BulkAssignRequest struct {
	TaskIDs    []string `json:"taskIds" binding:"required"`
	AssigneeID string   `json:"assigneeId" binding:"required"`
}

type BulkMoveToSprintRequest struct {
	TaskIDs  []string `json:"taskIds" binding:"required"`
	SprintID string   `json:"sprintId" binding:"required"`
}

// Sprint burndown models
type BurndownPoint struct {
	Date   time.Time `json:"date"`
	Points int       `json:"points"`
}

type SprintBurndownResponse struct {
	SprintID         string          `json:"sprintId"`
	StartDate        time.Time       `json:"startDate"`
	EndDate          time.Time       `json:"endDate"`
	TotalStoryPoints int             `json:"totalStoryPoints"`
	CompletedPoints  int             `json:"completedPoints"`
	RemainingPoints  int             `json:"remainingPoints"`
	IdealBurndown    []BurndownPoint `json:"idealBurndown"`
	ActualBurndown   []BurndownPoint `json:"actualBurndown"`
	CompletionRate   float64         `json:"completionRate"`
}



// ============================================
// GOAL MODELS (if not using repository.Goal directly)
// ============================================

type GoalResponse struct {
	ID           string     `json:"id"`
	WorkspaceID  string     `json:"workspaceId"`
	ProjectID    *string    `json:"projectId,omitempty"`
	SprintID     *string    `json:"sprintId,omitempty"`
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	GoalType     string     `json:"goalType"`
	Status       string     `json:"status"`
	TargetValue  *float64   `json:"targetValue,omitempty"`
	CurrentValue float64    `json:"currentValue"`
	Progress     float64    `json:"progress"`
	Unit         *string    `json:"unit,omitempty"`
	StartDate    *time.Time `json:"startDate,omitempty"`
	TargetDate   *time.Time `json:"targetDate,omitempty"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
	OwnerID      *string    `json:"ownerId,omitempty"`
	CreatedBy    *string    `json:"createdBy,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	KeyResults   []KeyResultResponse `json:"keyResults,omitempty"`
	LinkedTasks  []string   `json:"linkedTasks,omitempty"`
}

type KeyResultResponse struct {
	ID           string    `json:"id"`
	GoalID       string    `json:"goalId"`
	Title        string    `json:"title"`
	Description  *string   `json:"description,omitempty"`
	TargetValue  float64   `json:"targetValue"`
	CurrentValue float64   `json:"currentValue"`
	Progress     float64   `json:"progress"`
	Unit         *string   `json:"unit,omitempty"`
	Status       string    `json:"status"`
	Weight       float64   `json:"weight"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ============================================
// VELOCITY/ANALYTICS MODELS
// ============================================

type VelocityHistoryResponse struct {
	SprintID        string    `json:"sprintId"`
	SprintName      string    `json:"sprintName"`
	SprintNumber    int       `json:"sprintNumber"`
	CommittedPoints int       `json:"committedPoints"`
	CompletedPoints int       `json:"completedPoints"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
}

type VelocityTrendResponse struct {
	Sprints         []VelocityHistoryResponse `json:"sprints"`
	AverageVelocity float64                   `json:"averageVelocity"`
	TrendDirection  string                    `json:"trendDirection"`
	TrendPercentage float64                   `json:"trendPercentage"`
}

type CycleTimeResponse struct {
	TaskID           string     `json:"taskId"`
	TaskTitle        string     `json:"taskTitle"`
	CycleTimeHours   *float64   `json:"cycleTimeHours,omitempty"`
	LeadTimeHours    *float64   `json:"leadTimeHours,omitempty"`
	StartedAt        *time.Time `json:"startedAt,omitempty"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
}

// ============================================
// GANTT CHART MODELS
// ============================================

type GanttTaskResponse struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	StartDate     *time.Time `json:"startDate,omitempty"`
	DueDate       *time.Time `json:"dueDate,omitempty"`
	EndDate       *time.Time `json:"endDate,omitempty"`
	Status        string     `json:"status"`
	Priority      string     `json:"priority"`
	Progress      float64    `json:"progress"`
	AssigneeIDs   []string   `json:"assigneeIds"`
	ParentTaskID  *string    `json:"parentTaskId,omitempty"`
	Dependencies  []string   `json:"dependencies"`
	SprintID      *string    `json:"sprintId,omitempty"`
	StoryPoints   *int       `json:"storyPoints,omitempty"`
	Type          *string    `json:"type,omitempty"`
	Color         string     `json:"color"`
}

type GanttDataResponse struct {
	Tasks      []GanttTaskResponse `json:"tasks"`
	StartDate  time.Time           `json:"startDate"`
	EndDate    time.Time           `json:"endDate"`
	TotalTasks int                 `json:"totalTasks"`
}

// ============================================
// SPRINT REPORT MODELS
// ============================================

type SprintReportResponse struct {
	SprintID            string   `json:"sprintId"`
	CommittedTasks      int      `json:"committedTasks"`
	CommittedPoints     int      `json:"committedPoints"`
	CompletedTasks      int      `json:"completedTasks"`
	CompletedPoints     int      `json:"completedPoints"`
	IncompleteTasks     int      `json:"incompleteTasks"`
	IncompletePoints    int      `json:"incompletePoints"`
	CarryoverTasks      int      `json:"carryoverTasks"`
	CarryoverPoints     int      `json:"carryoverPoints"`
	TotalEstimatedHours float64  `json:"totalEstimatedHours"`
	TotalLoggedHours    float64  `json:"totalLoggedHours"`
	AvgCycleTimeHours   *float64 `json:"avgCycleTimeHours,omitempty"`
	AvgLeadTimeHours    *float64 `json:"avgLeadTimeHours,omitempty"`
	Velocity            int      `json:"velocity"`
	GoalsCompleted      int      `json:"goalsCompleted"`
	GoalsTotal          int      `json:"goalsTotal"`
}

// ============================================
// DASHBOARD MODELS
// ============================================

type SprintAnalyticsDashboardResponse struct {
	SprintID             string              `json:"sprintId"`
	Report               SprintReportResponse `json:"report"`
	CompletionPercentage float64             `json:"completionPercentage"`
	DaysRemaining        int                 `json:"daysRemaining"`
	DaysElapsed          int                 `json:"daysElapsed"`
	CurrentVelocity      int                 `json:"currentVelocity"`
	ProjectedVelocity    int                 `json:"projectedVelocity"`
	AvgCycleTimeHours    *float64            `json:"avgCycleTimeHours,omitempty"`
	BurndownAvailable    bool                `json:"burndownAvailable"`
	GoalsCompleted       int                 `json:"goalsCompleted"`
	GoalsTotal           int                 `json:"goalsTotal"`
	TasksByStatus        map[string]int      `json:"tasksByStatus"`
	TasksByPriority      map[string]int      `json:"tasksByPriority"`
}

type ProjectAnalyticsDashboardResponse struct {
	ProjectID                 string                `json:"projectId"`
	VelocityTrend             VelocityTrendResponse `json:"velocityTrend"`
	AvgCycleTimeHours         float64               `json:"avgCycleTimeHours"`
	AvgLeadTimeHours          float64               `json:"avgLeadTimeHours"`
	ActiveSprintID            *string               `json:"activeSprintId,omitempty"`
	ActiveSprintName          *string               `json:"activeSprintName,omitempty"`
	TotalTasks                int                   `json:"totalTasks"`
	CompletedTasks            int                   `json:"completedTasks"`
	OpenTasks                 int                   `json:"openTasks"`
	OverdueTasks              int                   `json:"overdueTasks"`
	TasksCompletedLast30Days  int                   `json:"tasksCompletedLast30Days"`
	PointsCompletedLast30Days int                   `json:"pointsCompletedLast30Days"`
}