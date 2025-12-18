package models

import "time"

// ============================================
// TASK REQUESTS & RESPONSES
// ============================================

// ============================================
// COMMENT REQUESTS & RESPONSES
// ============================================

type CreateCommentRequest struct {
	Content        string   `json:"content" binding:"required"`
	MentionedUsers []string `json:"mentionedUsers"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type CommentResponse struct {
	ID              string        `json:"id"`
	TaskID          string        `json:"taskId"`
	UserID          string        `json:"userId"`
	ParentCommentID *string       `json:"parentCommentId"`
	Content         string        `json:"content"`
	MentionedUsers  []string      `json:"mentionedUsers"`
	User            *UserResponse `json:"user"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

// ============================================
// ATTACHMENT REQUESTS & RESPONSES
// ============================================

type CreateAttachmentRequest struct {
	Filename string `json:"filename" binding:"required"`
	FileURL  string `json:"fileUrl" binding:"required"`
	FileSize int64  `json:"fileSize" binding:"required"`
	MimeType string `json:"mimeType" binding:"required"`
}

type AttachmentResponse struct {
	ID         string        `json:"id"`
	TaskID     string        `json:"taskId"`
	UserID     string        `json:"userId"`
	Filename   string        `json:"filename"`
	FileURL    string        `json:"fileUrl"`
	FileSize   int64         `json:"fileSize"`
	MimeType   string        `json:"mimeType"`
	User       *UserResponse `json:"user"`
	CreatedAt  time.Time     `json:"createdAt"`
}

// ============================================
// TIME TRACKING REQUESTS & RESPONSES
// ============================================

type LogTimeRequest struct {
	DurationSeconds int     `json:"durationSeconds" binding:"required"`
	Description     *string `json:"description"`
}

type TimeEntryResponse struct {
	ID              string        `json:"id"`
	TaskID          string        `json:"taskId"`
	UserID          string        `json:"userId"`
	StartTime       time.Time     `json:"startTime"`
	EndTime         *time.Time    `json:"endTime"`
	DurationSeconds *int          `json:"durationSeconds"`
	Description     *string       `json:"description"`
	IsManual        bool          `json:"isManual"`
	User            *UserResponse `json:"user"`
	CreatedAt       time.Time     `json:"createdAt"`
}

// ============================================
// DEPENDENCY REQUESTS & RESPONSES
// ============================================

type CreateDependencyRequest struct {
	DependsOnTaskID string `json:"dependsOnTaskId" binding:"required"`
	DependencyType  string `json:"dependencyType" binding:"required"` // 'blocks', 'depends_on', 'related'
}

type DependencyResponse struct {
	ID              string       `json:"id"`
	TaskID          string       `json:"taskId"`
	DependsOnTaskID string       `json:"dependsOnTaskId"`
	DependencyType  string       `json:"dependencyType"`
	DependsOnTask   *TaskSummary `json:"dependsOnTask"`
	CreatedAt       time.Time    `json:"createdAt"`
}

type TaskSummary struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Priority  string `json:"priority"`
}

// ============================================
// CHECKLIST REQUESTS & RESPONSES
// ============================================

type CreateChecklistRequest struct {
	Title string `json:"title" binding:"required"`
}

type CreateChecklistItemRequest struct {
	Content    string  `json:"content" binding:"required"`
	AssigneeID *string `json:"assigneeId"`
}

type ChecklistResponse struct {
	ID        string                  `json:"id"`
	TaskID    string                  `json:"taskId"`
	Title     string                  `json:"title"`
	Items     []ChecklistItemResponse `json:"items"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`  
}

type ChecklistItemResponse struct {
	ID          string     `json:"id"`
	ChecklistID string     `json:"checklistId"`
	Content     string     `json:"content"`
	Completed   bool       `json:"completed"`      // ✅ Fixed: was IsCompleted
	AssigneeID  *string    `json:"assigneeId"`
	Position    int        `json:"position"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
// ============================================
// ACTIVITY RESPONSES
// ============================================

type ActivityResponse struct {
	ID        string        `json:"id"`
	TaskID    string        `json:"taskId"`
	UserID    *string       `json:"userId"`
	Action    string        `json:"action"`
	FieldName *string       `json:"fieldName"`
	OldValue  *string       `json:"oldValue"`
	NewValue  *string       `json:"newValue"`
	User      *UserResponse `json:"user"`
	CreatedAt time.Time     `json:"createdAt"`
}

// ============================================
// BULK OPERATION REQUESTS
// ============================================

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

// ============================================
// SCRUM SPECIFIC RESPONSES
// ============================================

type SprintBurndownResponse struct {
	SprintID         string               `json:"sprintId"`
	StartDate        time.Time            `json:"startDate"`
	EndDate          time.Time            `json:"endDate"`
	TotalStoryPoints int                  `json:"totalStoryPoints"`
	CompletedPoints  int                  `json:"completedPoints"`
	RemainingPoints  int                  `json:"remainingPoints"`
	IdealBurndown    []BurndownPoint      `json:"idealBurndown"`
	ActualBurndown   []BurndownPoint      `json:"actualBurndown"`
	CompletionRate   float64              `json:"completionRate"`
}

type BurndownPoint struct {
	Date   time.Time `json:"date"`
	Points int       `json:"points"`
}

type SprintBoardResponse struct {
	Todo       []TaskResponse `json:"todo"`
	InProgress []TaskResponse `json:"inProgress"`
	InReview   []TaskResponse `json:"inReview"`
	Done       []TaskResponse `json:"done"`
}

// ============================================
// HELPER: USER RESPONSE
// ============================================


// internal/models/task_model.go - TaskResponse
type TaskResponse struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"projectId"`
	SprintID       *string    `json:"sprintId,omitempty"`
	ParentTaskID   *string    `json:"parentTaskId,omitempty"`
	Title          string     `json:"title"`
	Description    *string    `json:"description,omitempty"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	Type           *string    `json:"type,omitempty"` // ← ADD THIS
	AssigneeIDs    []string   `json:"assigneeIds"`
	WatcherIDs     []string   `json:"watcherIds"`
	LabelIDs       []string   `json:"labelIds"`
	EstimatedHours *float64   `json:"estimatedHours,omitempty"`
	ActualHours    *float64   `json:"actualHours,omitempty"`
	StoryPoints    *int       `json:"storyPoints,omitempty"`
	StartDate      *time.Time `json:"startDate,omitempty"`
	DueDate        *time.Time `json:"dueDate,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	Blocked        bool       `json:"blocked"`
	Position       int        `json:"position"`
	CreatedBy      *string    `json:"createdBy,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}