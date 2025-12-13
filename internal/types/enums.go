package types

// Task Status values
const (
	StatusBacklog    = "backlog"
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusInReview   = "in_review"
	StatusDone       = "done"
	StatusCancelled  = "cancelled"
)

// Task Priority values
const (
	PriorityUrgent = "urgent"
	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
	PriorityNone   = "none"
)

// Task Type values
const (
	TypeEpic    = "epic"
	TypeStory   = "story"
	TypeTask    = "task"
	TypeBug     = "bug"
	TypeSubtask = "subtask"
)

// Sprint Status values
const (
	SprintPlanning  = "planning"
	SprintActive    = "active"
	SprintCompleted = "completed"
)

// User Status values
const (
	UserOnline  = "online"
	UserOffline = "offline"
	UserAway    = "away"
	UserBusy    = "busy"
)

// Workspace/Project Member Roles
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleLead   = "lead"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// Valid status values for validation
var ValidTaskStatuses = []string{
	StatusBacklog, StatusTodo, StatusInProgress,
	StatusInReview, StatusDone, StatusCancelled,
}

var ValidPriorities = []string{
	PriorityUrgent, PriorityHigh, PriorityMedium,
	PriorityLow, PriorityNone,
}

var ValidTaskTypes = []string{
	TypeEpic, TypeStory, TypeTask, TypeBug, TypeSubtask,
}

// Helper functions for validation
func IsValidTaskStatus(status string) bool {
	for _, s := range ValidTaskStatuses {
		if s == status {
			return true
		}
	}
	return false
}

func IsValidPriority(priority string) bool {
	for _, p := range ValidPriorities {
		if p == priority {
			return true
		}
	}
	return false
}

func IsValidTaskType(taskType string) bool {
	for _, t := range ValidTaskTypes {
		if t == taskType {
			return true
		}
	}
	return false
}
