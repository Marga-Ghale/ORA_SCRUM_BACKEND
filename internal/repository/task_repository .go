package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	ID          string
	Key         string
	Title       string
	Description *string
	Status      string
	Priority    string
	Type        string
	ProjectID   string
	SprintID    *string
	AssigneeID  *string
	ReporterID  string
	ParentID    *string
	StoryPoints *int
	DueDate     *time.Time
	OrderIndex  int
	Labels      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Assignee    *User
	Reporter    *User
}

type TaskFilters struct {
	Status   []string
	Priority []string
	Type     []string
	SprintID *string
	Search   string
	Limit    int
	Offset   int
}

type BulkTaskUpdate struct {
	ID         string
	Status     *string
	SprintID   *string
	OrderIndex *int
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindByKey(ctx context.Context, key string) (*Task, error)
	FindByProjectID(ctx context.Context, projectID string, filters *TaskFilters) ([]*Task, error)
	FindBySprintID(ctx context.Context, sprintID string) ([]*Task, error)
	FindBacklog(ctx context.Context, projectID string) ([]*Task, error)
	FindOverdue(ctx context.Context) ([]*Task, error)
	FindDueSoon(ctx context.Context, within time.Duration) ([]*Task, error)
	FindByAssignee(ctx context.Context, assigneeID string) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
	BulkUpdate(ctx context.Context, updates []BulkTaskUpdate) error
	CountBySprintID(ctx context.Context, sprintID string) (total int, completed int, err error)
}

type pgTaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) TaskRepository {
	return &pgTaskRepository{pool: pool}
}

func (r *pgTaskRepository) Create(ctx context.Context, task *Task) error {
	if task.Status == "" {
		task.Status = "backlog"
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}
	if task.Type == "" {
		task.Type = "task"
	}
	if task.Labels == nil {
		task.Labels = []string{}
	}

	query := `
		INSERT INTO tasks (key, title, description, status, priority, type, project_id, sprint_id, 
		                   assignee_id, reporter_id, parent_id, story_points, due_date, order_index, labels)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		task.Key, task.Title, task.Description, task.Status, task.Priority, task.Type,
		task.ProjectID, task.SprintID, task.AssigneeID, task.ReporterID, task.ParentID,
		task.StoryPoints, task.DueDate, task.OrderIndex, task.Labels,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *pgTaskRepository) FindByID(ctx context.Context, id string) (*Task, error) {
	query := `
		SELECT t.id, t.key, t.title, t.description, t.status, t.priority, t.type,
		       t.project_id, t.sprint_id, t.assignee_id, t.reporter_id, t.parent_id,
		       t.story_points, t.due_date, t.order_index, t.labels, t.created_at, t.updated_at,
		       a.id, a.name, a.email, a.avatar,
		       rep.id, rep.name, rep.email, rep.avatar
		FROM tasks t
		LEFT JOIN users a ON t.assignee_id = a.id
		LEFT JOIN users rep ON t.reporter_id = rep.id
		WHERE t.id = $1
	`
	task := &Task{}
	var assigneeID, assigneeName, assigneeEmail, assigneeAvatar sql.NullString
	var reporterID, reporterName, reporterEmail, reporterAvatar sql.NullString

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.Key, &task.Title, &task.Description, &task.Status, &task.Priority, &task.Type,
		&task.ProjectID, &task.SprintID, &task.AssigneeID, &task.ReporterID, &task.ParentID,
		&task.StoryPoints, &task.DueDate, &task.OrderIndex, &task.Labels, &task.CreatedAt, &task.UpdatedAt,
		&assigneeID, &assigneeName, &assigneeEmail, &assigneeAvatar,
		&reporterID, &reporterName, &reporterEmail, &reporterAvatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if assigneeID.Valid {
		task.Assignee = &User{ID: assigneeID.String, Name: assigneeName.String, Email: assigneeEmail.String}
		if assigneeAvatar.Valid {
			task.Assignee.Avatar = &assigneeAvatar.String
		}
	}
	if reporterID.Valid {
		task.Reporter = &User{ID: reporterID.String, Name: reporterName.String, Email: reporterEmail.String}
		if reporterAvatar.Valid {
			task.Reporter.Avatar = &reporterAvatar.String
		}
	}

	return task, nil
}

func (r *pgTaskRepository) FindByKey(ctx context.Context, key string) (*Task, error) {
	query := `SELECT id FROM tasks WHERE UPPER(key) = UPPER($1)`
	var id string
	err := r.pool.QueryRow(ctx, query, key).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *pgTaskRepository) FindByProjectID(ctx context.Context, projectID string, filters *TaskFilters) ([]*Task, error) {
	query := `
		SELECT t.id, t.key, t.title, t.description, t.status, t.priority, t.type,
		       t.project_id, t.sprint_id, t.assignee_id, t.reporter_id, t.parent_id,
		       t.story_points, t.due_date, t.order_index, t.labels, t.created_at, t.updated_at,
		       a.id, a.name, a.email, a.avatar,
		       rep.id, rep.name, rep.email, rep.avatar
		FROM tasks t
		LEFT JOIN users a ON t.assignee_id = a.id
		LEFT JOIN users rep ON t.reporter_id = rep.id
		WHERE t.project_id = $1
	`
	args := []interface{}{projectID}
	argNum := 1

	if filters != nil {
		if len(filters.Status) > 0 {
			argNum++
			query += fmt.Sprintf(" AND t.status = ANY($%d)", argNum)
			args = append(args, filters.Status)
		}
		if len(filters.Priority) > 0 {
			argNum++
			query += fmt.Sprintf(" AND t.priority = ANY($%d)", argNum)
			args = append(args, filters.Priority)
		}
		if len(filters.Type) > 0 {
			argNum++
			query += fmt.Sprintf(" AND t.type = ANY($%d)", argNum)
			args = append(args, filters.Type)
		}
		if filters.SprintID != nil {
			argNum++
			query += fmt.Sprintf(" AND t.sprint_id = $%d", argNum)
			args = append(args, *filters.SprintID)
		}
		if filters.Search != "" {
			argNum++
			query += fmt.Sprintf(" AND (LOWER(t.title) LIKE LOWER($%d) OR LOWER(t.key) LIKE LOWER($%d))", argNum, argNum)
			args = append(args, "%"+filters.Search+"%")
		}
	}

	query += " ORDER BY t.order_index, t.created_at DESC"

	if filters != nil && filters.Limit > 0 {
		argNum++
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filters.Limit)
	}
	if filters != nil && filters.Offset > 0 {
		argNum++
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var assigneeID, assigneeName, assigneeEmail, assigneeAvatar sql.NullString
		var reporterID, reporterName, reporterEmail, reporterAvatar sql.NullString

		if err := rows.Scan(
			&task.ID, &task.Key, &task.Title, &task.Description, &task.Status, &task.Priority, &task.Type,
			&task.ProjectID, &task.SprintID, &task.AssigneeID, &task.ReporterID, &task.ParentID,
			&task.StoryPoints, &task.DueDate, &task.OrderIndex, &task.Labels, &task.CreatedAt, &task.UpdatedAt,
			&assigneeID, &assigneeName, &assigneeEmail, &assigneeAvatar,
			&reporterID, &reporterName, &reporterEmail, &reporterAvatar,
		); err != nil {
			return nil, err
		}

		if assigneeID.Valid {
			task.Assignee = &User{ID: assigneeID.String, Name: assigneeName.String, Email: assigneeEmail.String}
			if assigneeAvatar.Valid {
				task.Assignee.Avatar = &assigneeAvatar.String
			}
		}
		if reporterID.Valid {
			task.Reporter = &User{ID: reporterID.String, Name: reporterName.String, Email: reporterEmail.String}
			if reporterAvatar.Valid {
				task.Reporter.Avatar = &reporterAvatar.String
			}
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindBySprintID(ctx context.Context, sprintID string) ([]*Task, error) {
	query := `
		SELECT t.id, t.key, t.title, t.description, t.status, t.priority, t.type,
		       t.project_id, t.sprint_id, t.assignee_id, t.reporter_id, t.parent_id,
		       t.story_points, t.due_date, t.order_index, t.labels, t.created_at, t.updated_at,
		       a.id, a.name, a.email, a.avatar,
		       rep.id, rep.name, rep.email, rep.avatar
		FROM tasks t
		LEFT JOIN users a ON t.assignee_id = a.id
		LEFT JOIN users rep ON t.reporter_id = rep.id
		WHERE t.sprint_id = $1
		ORDER BY t.order_index
	`
	rows, err := r.pool.Query(ctx, query, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var assigneeID, assigneeName, assigneeEmail, assigneeAvatar sql.NullString
		var reporterID, reporterName, reporterEmail, reporterAvatar sql.NullString

		if err := rows.Scan(
			&task.ID, &task.Key, &task.Title, &task.Description, &task.Status, &task.Priority, &task.Type,
			&task.ProjectID, &task.SprintID, &task.AssigneeID, &task.ReporterID, &task.ParentID,
			&task.StoryPoints, &task.DueDate, &task.OrderIndex, &task.Labels, &task.CreatedAt, &task.UpdatedAt,
			&assigneeID, &assigneeName, &assigneeEmail, &assigneeAvatar,
			&reporterID, &reporterName, &reporterEmail, &reporterAvatar,
		); err != nil {
			return nil, err
		}

		if assigneeID.Valid {
			task.Assignee = &User{ID: assigneeID.String, Name: assigneeName.String, Email: assigneeEmail.String}
			if assigneeAvatar.Valid {
				task.Assignee.Avatar = &assigneeAvatar.String
			}
		}
		if reporterID.Valid {
			task.Reporter = &User{ID: reporterID.String, Name: reporterName.String, Email: reporterEmail.String}
			if reporterAvatar.Valid {
				task.Reporter.Avatar = &reporterAvatar.String
			}
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindBacklog(ctx context.Context, projectID string) ([]*Task, error) {
	query := `SELECT id FROM tasks WHERE project_id = $1 AND sprint_id IS NULL ORDER BY order_index`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		task, _ := r.FindByID(ctx, id)
		if task != nil {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindOverdue(ctx context.Context) ([]*Task, error) {
	query := `SELECT id FROM tasks WHERE due_date < NOW() AND status NOT IN ('done', 'cancelled')`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		task, _ := r.FindByID(ctx, id)
		if task != nil {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindDueSoon(ctx context.Context, within time.Duration) ([]*Task, error) {
	query := `SELECT id FROM tasks WHERE due_date > NOW() AND due_date < $1 AND status NOT IN ('done', 'cancelled')`
	deadline := time.Now().Add(within)
	rows, err := r.pool.Query(ctx, query, deadline)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		task, _ := r.FindByID(ctx, id)
		if task != nil {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *pgTaskRepository) FindByAssignee(ctx context.Context, assigneeID string) ([]*Task, error) {
	query := `SELECT id FROM tasks WHERE assignee_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, assigneeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		task, _ := r.FindByID(ctx, id)
		if task != nil {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *pgTaskRepository) Update(ctx context.Context, task *Task) error {
	query := `
		UPDATE tasks SET 
			title = $2, description = $3, status = $4, priority = $5, type = $6,
			sprint_id = $7, assignee_id = $8, story_points = $9, due_date = $10,
			order_index = $11, labels = $12, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.Priority, task.Type,
		task.SprintID, task.AssigneeID, task.StoryPoints, task.DueDate,
		task.OrderIndex, task.Labels,
	)
	return err
}

func (r *pgTaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgTaskRepository) BulkUpdate(ctx context.Context, updates []BulkTaskUpdate) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, u := range updates {
		if u.Status != nil {
			_, err = tx.Exec(ctx, `UPDATE tasks SET status = $2, updated_at = NOW() WHERE id = $1`, u.ID, *u.Status)
			if err != nil {
				return err
			}
		}
		if u.SprintID != nil {
			_, err = tx.Exec(ctx, `UPDATE tasks SET sprint_id = $2, updated_at = NOW() WHERE id = $1`, u.ID, *u.SprintID)
			if err != nil {
				return err
			}
		}
		if u.OrderIndex != nil {
			_, err = tx.Exec(ctx, `UPDATE tasks SET order_index = $2, updated_at = NOW() WHERE id = $1`, u.ID, *u.OrderIndex)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *pgTaskRepository) CountBySprintID(ctx context.Context, sprintID string) (total int, completed int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'done') as completed
		FROM tasks WHERE sprint_id = $1
	`
	err = r.pool.QueryRow(ctx, query, sprintID).Scan(&total, &completed)
	return
}
