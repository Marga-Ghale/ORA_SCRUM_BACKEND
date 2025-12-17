package repository

import (
	"context"
	"database/sql"
	"time"
)

// TaskChecklist model
type TaskChecklist struct {
	ID        string            `json:"id" db:"id"`
	TaskID    string            `json:"taskId" db:"task_id"`
	Title     string            `json:"title" db:"title"`
	Items     []*ChecklistItem  `json:"items,omitempty"`
	CreatedAt time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time         `json:"updatedAt" db:"updated_at"`
}

// ChecklistItem model
type ChecklistItem struct {
	ID          string     `json:"id" db:"id"`
	ChecklistID string     `json:"checklistId" db:"checklist_id"`
	Content     string     `json:"content" db:"content"`
	Completed   bool       `json:"completed" db:"completed"`
	AssigneeID  *string    `json:"assigneeId,omitempty" db:"assignee_id"`
	Position    int        `json:"position" db:"position"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
}

// TaskChecklistRepository interface
type TaskChecklistRepository interface {
	// Checklist operations
	CreateChecklist(ctx context.Context, checklist *TaskChecklist) error
	FindChecklistByID(ctx context.Context, id string) (*TaskChecklist, error)
	FindByTaskID(ctx context.Context, taskID string) ([]*TaskChecklist, error)
	UpdateChecklist(ctx context.Context, checklist *TaskChecklist) error
	DeleteChecklist(ctx context.Context, id string) error

	// Checklist item operations
	CreateItem(ctx context.Context, item *ChecklistItem) error
	FindItemByID(ctx context.Context, id string) (*ChecklistItem, error)
	FindItemsByChecklistID(ctx context.Context, checklistID string) ([]*ChecklistItem, error)
	UpdateItem(ctx context.Context, item *ChecklistItem) error
	ToggleItem(ctx context.Context, id string) error
	DeleteItem(ctx context.Context, id string) error
}

// taskChecklistRepository implementation
type taskChecklistRepository struct {
	db *sql.DB
}

// NewTaskChecklistRepository creates a new TaskChecklistRepository
func NewTaskChecklistRepository(db *sql.DB) TaskChecklistRepository {
	return &taskChecklistRepository{db: db}
}

// ============================================
// CHECKLIST OPERATIONS
// ============================================

// CreateChecklist inserts a new checklist
func (r *taskChecklistRepository) CreateChecklist(ctx context.Context, checklist *TaskChecklist) error {
	query := `
		INSERT INTO task_checklists (
			id, task_id, title, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, NOW(), NOW()
		) RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		checklist.TaskID,
		checklist.Title,
	).Scan(&checklist.ID, &checklist.CreatedAt, &checklist.UpdatedAt)
}

// FindChecklistByID retrieves a checklist by ID
func (r *taskChecklistRepository) FindChecklistByID(ctx context.Context, id string) (*TaskChecklist, error) {
	query := `SELECT * FROM task_checklists WHERE id = $1`

	checklist := &TaskChecklist{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&checklist.ID,
		&checklist.TaskID,
		&checklist.Title,
		&checklist.CreatedAt,
		&checklist.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load items
	items, err := r.FindItemsByChecklistID(ctx, id)
	if err != nil {
		return nil, err
	}
	checklist.Items = items

	return checklist, nil
}

// FindByTaskID retrieves all checklists for a task
func (r *taskChecklistRepository) FindByTaskID(ctx context.Context, taskID string) ([]*TaskChecklist, error) {
	query := `SELECT * FROM task_checklists WHERE task_id = $1 ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checklists []*TaskChecklist
	for rows.Next() {
		checklist := &TaskChecklist{}
		err := rows.Scan(
			&checklist.ID,
			&checklist.TaskID,
			&checklist.Title,
			&checklist.CreatedAt,
			&checklist.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Load items for each checklist
		items, err := r.FindItemsByChecklistID(ctx, checklist.ID)
		if err != nil {
			return nil, err
		}
		checklist.Items = items

		checklists = append(checklists, checklist)
	}

	return checklists, rows.Err()
}

// UpdateChecklist updates an existing checklist
func (r *taskChecklistRepository) UpdateChecklist(ctx context.Context, checklist *TaskChecklist) error {
	query := `
		UPDATE task_checklists SET
			title = $2,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		checklist.ID,
		checklist.Title,
	).Scan(&checklist.UpdatedAt)
}

// DeleteChecklist removes a checklist and all its items
func (r *taskChecklistRepository) DeleteChecklist(ctx context.Context, id string) error {
	// Delete items first (cascade should handle this, but explicit is better)
	_, err := r.db.ExecContext(ctx, `DELETE FROM checklist_items WHERE checklist_id = $1`, id)
	if err != nil {
		return err
	}

	// Delete checklist
	query := `DELETE FROM task_checklists WHERE id = $1`
	_, err = r.db.ExecContext(ctx, query, id)
	return err
}

// ============================================
// CHECKLIST ITEM OPERATIONS
// ============================================

// CreateItem inserts a new checklist item
func (r *taskChecklistRepository) CreateItem(ctx context.Context, item *ChecklistItem) error {
	query := `
		INSERT INTO checklist_items (
			id, checklist_id, content, completed, assignee_id, position, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, false, $3, 
			COALESCE((SELECT MAX(position) + 1 FROM checklist_items WHERE checklist_id = $1), 0),
			NOW(), NOW()
		) RETURNING id, position, created_at, updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		item.ChecklistID,
		item.Content,
		item.AssigneeID,
	).Scan(&item.ID, &item.Position, &item.CreatedAt, &item.UpdatedAt)
}

// FindItemByID retrieves a checklist item by ID
func (r *taskChecklistRepository) FindItemByID(ctx context.Context, id string) (*ChecklistItem, error) {
	query := `SELECT * FROM checklist_items WHERE id = $1`

	item := &ChecklistItem{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.ChecklistID,
		&item.Content,
		&item.Completed,
		&item.AssigneeID,
		&item.Position,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return item, nil
}

// FindItemsByChecklistID retrieves all items for a checklist
func (r *taskChecklistRepository) FindItemsByChecklistID(ctx context.Context, checklistID string) ([]*ChecklistItem, error) {
	query := `SELECT * FROM checklist_items WHERE checklist_id = $1 ORDER BY position ASC`

	rows, err := r.db.QueryContext(ctx, query, checklistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*ChecklistItem
	for rows.Next() {
		item := &ChecklistItem{}
		err := rows.Scan(
			&item.ID,
			&item.ChecklistID,
			&item.Content,
			&item.Completed,
			&item.AssigneeID,
			&item.Position,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// UpdateItem updates an existing checklist item
func (r *taskChecklistRepository) UpdateItem(ctx context.Context, item *ChecklistItem) error {
	query := `
		UPDATE checklist_items SET
			content = $2,
			completed = $3,
			assignee_id = $4,
			position = $5,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		item.ID,
		item.Content,
		item.Completed,
		item.AssigneeID,
		item.Position,
	).Scan(&item.UpdatedAt)
}

// ToggleItem toggles the completed status of a checklist item
func (r *taskChecklistRepository) ToggleItem(ctx context.Context, id string) error {
	query := `
		UPDATE checklist_items SET
			completed = NOT completed,
			updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DeleteItem removes a checklist item
func (r *taskChecklistRepository) DeleteItem(ctx context.Context, id string) error {
	query := `DELETE FROM checklist_items WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}