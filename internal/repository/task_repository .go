package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Task model

// Task struct - ADD Type field after Priority
type Task struct {
	ID             string     `json:"id" db:"id"`
	Title          string     `json:"title" db:"title"`
	Description    *string    `json:"description,omitempty" db:"description"`
	Status         string     `json:"status" db:"status"`
	Priority       string     `json:"priority" db:"priority"`
	Type           *string    `json:"type,omitempty" db:"type"` // ← ADD THIS
	ProjectID      string     `json:"projectId" db:"project_id"`
	SprintID       *string    `json:"sprintId,omitempty" db:"sprint_id"`
	ParentTaskID   *string    `json:"parentTaskId,omitempty" db:"parent_task_id"`
	AssigneeIDs    []string   `json:"assigneeIds" db:"assignee_ids"`
	WatcherIDs     []string   `json:"watcherIds" db:"watcher_ids"`
	LabelIDs       []string   `json:"labelIds" db:"label_ids"`
	StoryPoints    *int       `json:"storyPoints,omitempty" db:"story_points"`
	EstimatedHours *float64   `json:"estimatedHours,omitempty" db:"estimated_hours"`
	ActualHours    *float64   `json:"actualHours,omitempty" db:"actual_hours"`
	StartDate      *time.Time `json:"startDate,omitempty" db:"start_date"`
	DueDate        *time.Time `json:"dueDate,omitempty" db:"due_date"`
	CompletedAt    *time.Time `json:"completedAt,omitempty" db:"completed_at"`
	Blocked        bool       `json:"blocked" db:"blocked"`
	Position       int        `json:"position" db:"position"`
	CreatedBy      *string    `json:"createdBy,omitempty" db:"created_by"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
}

// TaskFilters for advanced filtering
type TaskFilters struct {
	ProjectID   string
	SprintID    *string
	AssigneeIDs []string
	Status      []string
	Priority    []string
	LabelIDs    []string
	Search      *string
	DueBefore   *time.Time
	DueAfter    *time.Time
	Overdue     *bool
	Blocked     *bool
	Limit       int
	Offset      int
}

// TaskRepository interface
type TaskRepository interface {
	// Basic CRUD
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error

	// Listing methods
	FindByProjectID(ctx context.Context, projectID string) ([]*Task, error)
	FindBySprintID(ctx context.Context, sprintID string) ([]*Task, error)
	FindByParentTaskID(ctx context.Context, parentTaskID string) ([]*Task, error)
	FindByAssigneeID(ctx context.Context, assigneeID string) ([]*Task, error)
	FindByStatus(ctx context.Context, projectID, status string) ([]*Task, error)
	FindBacklog(ctx context.Context, projectID string) ([]*Task, error)

	// Quick updates
	UpdateStatus(ctx context.Context, taskID, status string) error
	UpdatePriority(ctx context.Context, taskID, priority string) error
	MarkComplete(ctx context.Context, taskID string) error

	// Assignee/Watcher management
	AddAssignee(ctx context.Context, taskID, assigneeID string) error
	RemoveAssignee(ctx context.Context, taskID, assigneeID string) error
	AddWatcher(ctx context.Context, taskID, watcherID string) error
	RemoveWatcher(ctx context.Context, taskID, watcherID string) error

	// Advanced filtering
	FindWithFilters(ctx context.Context, filters *TaskFilters) ([]*Task, int, error)
	FindOverdue(ctx context.Context, projectID string) ([]*Task, error)
	FindBlocked(ctx context.Context, projectID string) ([]*Task, error)

	// Sprint/Scrum specific
	GetSprintVelocity(ctx context.Context, sprintID string) (int, error)
	GetCompletedStoryPoints(ctx context.Context, sprintID string) (int, error)

	// Bulk operations
	BulkUpdateStatus(ctx context.Context, taskIDs []string, status string) error
	BulkMoveToSprint(ctx context.Context, taskIDs []string, sprintID string) error
}

// taskRepository implementation
type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new TaskRepository
func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

// Create inserts a new task
func (r *taskRepository) Create(ctx context.Context, task *Task) error {
	query := `
		INSERT INTO tasks (
			id, project_id, sprint_id, parent_task_id, title, description,
			status, priority, assignee_ids, watcher_ids, label_ids,
			estimated_hours, actual_hours, story_points, start_date, due_date,
			blocked, position, created_by, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, COALESCE((SELECT MAX(position) + 1 FROM tasks WHERE project_id = $1), 0),
			$17, NOW(), NOW()
		) RETURNING id, created_at, updated_at, position`

	return r.db.QueryRowContext(
		ctx, query,
		task.ProjectID, task.SprintID, task.ParentTaskID, task.Title, task.Description,
		task.Status, task.Priority, pq.Array(task.AssigneeIDs), pq.Array(task.WatcherIDs),
		pq.Array(task.LabelIDs), task.EstimatedHours, task.ActualHours, task.StoryPoints,
		task.StartDate, task.DueDate, task.Blocked, task.CreatedBy,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt, &task.Position)
}

// FindByID retrieves a task by ID
func (r *taskRepository) FindByID(ctx context.Context, id string) (*Task, error) {
	query := `SELECT * FROM tasks WHERE id = $1`
	
	task := &Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority, &task.Type,
		&task.ProjectID, &task.SprintID, &task.ParentTaskID,
		pq.Array(&task.AssigneeIDs), pq.Array(&task.WatcherIDs), pq.Array(&task.LabelIDs),
		&task.StoryPoints, &task.EstimatedHours, &task.ActualHours,
		&task.StartDate, &task.DueDate, &task.CompletedAt, &task.Blocked,
		&task.Position, &task.CreatedBy, &task.CreatedAt, &task.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return task, nil
}

// Update updates an existing task
func (r *taskRepository) Update(ctx context.Context, task *Task) error {
	query := `
		UPDATE tasks SET
			sprint_id = $2, parent_task_id = $3, title = $4, description = $5,
			status = $6, priority = $7, assignee_ids = $8, watcher_ids = $9,
			label_ids = $10, estimated_hours = $11, actual_hours = $12,
			story_points = $13, start_date = $14, due_date = $15,
			completed_at = $16, blocked = $17, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		task.ID, task.SprintID, task.ParentTaskID, task.Title, task.Description,
		task.Status, task.Priority, pq.Array(task.AssigneeIDs), pq.Array(task.WatcherIDs),
		pq.Array(task.LabelIDs), task.EstimatedHours, task.ActualHours, task.StoryPoints,
		task.StartDate, task.DueDate, task.CompletedAt, task.Blocked,
	).Scan(&task.UpdatedAt)
}

// Delete removes a task
func (r *taskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// FindByProjectID retrieves all tasks for a project
func (r *taskRepository) FindByProjectID(ctx context.Context, projectID string) ([]*Task, error) {
	query := `SELECT * FROM tasks WHERE project_id = $1 ORDER BY position ASC, created_at DESC`
	return r.queryTasks(ctx, query, projectID)
}

// FindBySprintID retrieves all tasks for a sprint
func (r *taskRepository) FindBySprintID(ctx context.Context, sprintID string) ([]*Task, error) {
	query := `SELECT * FROM tasks WHERE sprint_id = $1 ORDER BY position ASC, created_at DESC`
	return r.queryTasks(ctx, query, sprintID)
}

// FindByParentTaskID retrieves all subtasks
func (r *taskRepository) FindByParentTaskID(ctx context.Context, parentTaskID string) ([]*Task, error) {
	query := `SELECT * FROM tasks WHERE parent_task_id = $1 ORDER BY position ASC, created_at DESC`
	return r.queryTasks(ctx, query, parentTaskID)
}

// FindByAssigneeID retrieves tasks assigned to a user
func (r *taskRepository) FindByAssigneeID(ctx context.Context, assigneeID string) ([]*Task, error) {
	query := `SELECT * FROM tasks WHERE $1 = ANY(assignee_ids) ORDER BY due_date ASC NULLS LAST, created_at DESC`
	return r.queryTasks(ctx, query, assigneeID)
}

// FindByStatus retrieves tasks by status in a project
func (r *taskRepository) FindByStatus(ctx context.Context, projectID, status string) ([]*Task, error) {
	query := `SELECT * FROM tasks WHERE project_id = $1 AND status = $2 ORDER BY position ASC`
	return r.queryTasks(ctx, query, projectID, status)
}

// FindBacklog retrieves backlog tasks (not in any sprint)
func (r *taskRepository) FindBacklog(ctx context.Context, projectID string) ([]*Task, error) {
	query := `SELECT * FROM tasks WHERE project_id = $1 AND sprint_id IS NULL AND parent_task_id IS NULL ORDER BY position ASC`
	return r.queryTasks(ctx, query, projectID)
}

// UpdateStatus updates task status
func (r *taskRepository) UpdateStatus(ctx context.Context, taskID, status string) error {
	query := `
		UPDATE tasks SET 
			status = $2, 
			completed_at = CASE WHEN $2 = 'done' THEN NOW() ELSE completed_at END,
			updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, taskID, status)
	return err
}

// UpdatePriority updates task priority
func (r *taskRepository) UpdatePriority(ctx context.Context, taskID, priority string) error {
	query := `UPDATE tasks SET priority = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, taskID, priority)
	return err
}

// MarkComplete marks a task as complete
func (r *taskRepository) MarkComplete(ctx context.Context, taskID string) error {
	query := `UPDATE tasks SET status = 'done', completed_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, taskID)
	return err
}

// AddAssignee adds an assignee to a task
func (r *taskRepository) AddAssignee(ctx context.Context, taskID, assigneeID string) error {
	query := `
		UPDATE tasks 
		SET assignee_ids = array_append(assignee_ids, $2),
		    updated_at = NOW()
		WHERE id = $1 AND NOT ($2 = ANY(assignee_ids))`
	_, err := r.db.ExecContext(ctx, query, taskID, assigneeID)
	return err
}

// RemoveAssignee removes an assignee from a task
func (r *taskRepository) RemoveAssignee(ctx context.Context, taskID, assigneeID string) error {
	query := `
		UPDATE tasks 
		SET assignee_ids = array_remove(assignee_ids, $2),
		    updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, taskID, assigneeID)
	return err
}

// AddWatcher adds a watcher to a task
func (r *taskRepository) AddWatcher(ctx context.Context, taskID, watcherID string) error {
	query := `
		UPDATE tasks 
		SET watcher_ids = array_append(watcher_ids, $2),
		    updated_at = NOW()
		WHERE id = $1 AND NOT ($2 = ANY(watcher_ids))`
	_, err := r.db.ExecContext(ctx, query, taskID, watcherID)
	return err
}

// RemoveWatcher removes a watcher from a task
func (r *taskRepository) RemoveWatcher(ctx context.Context, taskID, watcherID string) error {
	query := `
		UPDATE tasks 
		SET watcher_ids = array_remove(watcher_ids, $2),
		    updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, taskID, watcherID)
	return err
}

// FindWithFilters performs advanced filtering
func (r *taskRepository) FindWithFilters(ctx context.Context, filters *TaskFilters) ([]*Task, int, error) {
	// Build dynamic query based on filters
	baseQuery := `SELECT * FROM tasks WHERE project_id = $1`
	countQuery := `SELECT COUNT(*) FROM tasks WHERE project_id = $1`
	args := []interface{}{filters.ProjectID}
	argIndex := 2

	// Apply filters
	if filters.SprintID != nil {
		baseQuery += ` AND sprint_id = $` + string(rune(argIndex))
		countQuery += ` AND sprint_id = $` + string(rune(argIndex))
		args = append(args, *filters.SprintID)
		argIndex++
	}

	if len(filters.Status) > 0 {
		baseQuery += ` AND status = ANY($` + string(rune(argIndex)) + `)`
		countQuery += ` AND status = ANY($` + string(rune(argIndex)) + `)`
		args = append(args, pq.Array(filters.Status))
		argIndex++
	}

	if len(filters.Priority) > 0 {
		baseQuery += ` AND priority = ANY($` + string(rune(argIndex)) + `)`
		countQuery += ` AND priority = ANY($` + string(rune(argIndex)) + `)`
		args = append(args, pq.Array(filters.Priority))
		argIndex++
	}

	if filters.Overdue != nil && *filters.Overdue {
		baseQuery += ` AND due_date < NOW() AND status != 'done'`
		countQuery += ` AND due_date < NOW() AND status != 'done'`
	}

	if filters.Blocked != nil && *filters.Blocked {
		baseQuery += ` AND blocked = true`
		countQuery += ` AND blocked = true`
	}

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	baseQuery += ` ORDER BY position ASC LIMIT $` + string(rune(argIndex)) + ` OFFSET $` + string(rune(argIndex+1))
	args = append(args, filters.Limit, filters.Offset)

	tasks, err := r.queryTasks(ctx, baseQuery, args...)
	return tasks, total, err
}

// FindOverdue retrieves overdue tasks
func (r *taskRepository) FindOverdue(ctx context.Context, projectID string) ([]*Task, error) {
	query := `
		SELECT * FROM tasks 
		WHERE project_id = $1 
		  AND due_date < NOW() 
		  AND status != 'done'
		ORDER BY due_date ASC`
	return r.queryTasks(ctx, query, projectID)
}

// FindBlocked retrieves blocked tasks
func (r *taskRepository) FindBlocked(ctx context.Context, projectID string) ([]*Task, error) {
	query := `SELECT * FROM tasks WHERE project_id = $1 AND blocked = true ORDER BY created_at DESC`
	return r.queryTasks(ctx, query, projectID)
}

// GetSprintVelocity calculates total story points in a sprint
func (r *taskRepository) GetSprintVelocity(ctx context.Context, sprintID string) (int, error) {
	query := `SELECT COALESCE(SUM(story_points), 0) FROM tasks WHERE sprint_id = $1`
	var velocity int
	err := r.db.QueryRowContext(ctx, query, sprintID).Scan(&velocity)
	return velocity, err
}

// GetCompletedStoryPoints calculates completed story points in a sprint
func (r *taskRepository) GetCompletedStoryPoints(ctx context.Context, sprintID string) (int, error) {
	query := `SELECT COALESCE(SUM(story_points), 0) FROM tasks WHERE sprint_id = $1 AND status = 'done'`
	var points int
	err := r.db.QueryRowContext(ctx, query, sprintID).Scan(&points)
	return points, err
}

// BulkUpdateStatus updates status for multiple tasks
func (r *taskRepository) BulkUpdateStatus(ctx context.Context, taskIDs []string, status string) error {
	query := `
		UPDATE tasks SET 
			status = $2,
			completed_at = CASE WHEN $2 = 'done' THEN NOW() ELSE completed_at END,
			updated_at = NOW()
		WHERE id = ANY($1)`
	_, err := r.db.ExecContext(ctx, query, pq.Array(taskIDs), status)
	return err
}

// BulkMoveToSprint moves multiple tasks to a sprint
func (r *taskRepository) BulkMoveToSprint(ctx context.Context, taskIDs []string, sprintID string) error {
	query := `UPDATE tasks SET sprint_id = $2, updated_at = NOW() WHERE id = ANY($1)`
	_, err := r.db.ExecContext(ctx, query, pq.Array(taskIDs), sprintID)
	return err
}

// Helper method to query multiple tasks
// queryTasks helper - line 380
func (r *taskRepository) queryTasks(ctx context.Context, query string, args ...interface{}) ([]*Task, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority, &task.Type,
			&task.ProjectID, &task.SprintID, &task.ParentTaskID,
			pq.Array(&task.AssigneeIDs), pq.Array(&task.WatcherIDs), pq.Array(&task.LabelIDs),
			&task.StoryPoints, &task.EstimatedHours, &task.ActualHours,
			&task.StartDate, &task.DueDate, &task.CompletedAt, &task.Blocked,
			&task.Position, &task.CreatedBy, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}