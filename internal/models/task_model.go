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
}

// CreateTaskRequest for creating tasks
type CreateTaskRequest struct {
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