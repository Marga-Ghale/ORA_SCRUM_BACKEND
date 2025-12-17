-- ============================================
-- TASK MANAGEMENT DATABASE SCHEMA
-- ============================================
-- This migration creates all tables needed for the task management system

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- PROJECTS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    key VARCHAR(20) NOT NULL UNIQUE, -- Short identifier like "PROJ"
    type VARCHAR(50) NOT NULL DEFAULT 'scrum', -- scrum, kanban, basic
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, archived, on_hold
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_workspace_id ON projects(workspace_id);
CREATE INDEX idx_projects_key ON projects(key);
CREATE INDEX idx_projects_status ON projects(status);

-- ============================================
-- SPRINTS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS sprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    goal TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'planning', -- planning, active, completed
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sprints_project_id ON sprints(project_id);
CREATE INDEX idx_sprints_status ON sprints(status);
CREATE INDEX idx_sprints_dates ON sprints(start_date, end_date);

-- ============================================
-- TASKS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sprint_id UUID REFERENCES sprints(id) ON DELETE SET NULL,
    parent_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'todo', -- todo, in_progress, in_review, done, blocked
    priority VARCHAR(50) NOT NULL DEFAULT 'medium', -- low, medium, high, urgent
    assignee_ids UUID[] DEFAULT '{}',
    watcher_ids UUID[] DEFAULT '{}',
    label_ids UUID[] DEFAULT '{}',
    estimated_hours DECIMAL(10,2),
    actual_hours DECIMAL(10,2),
    story_points INTEGER,
    start_date TIMESTAMP,
    due_date TIMESTAMP,
    completed_at TIMESTAMP,
    blocked BOOLEAN DEFAULT FALSE,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_sprint_id ON tasks(sprint_id);
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_assignee_ids ON tasks USING GIN(assignee_ids);
CREATE INDEX idx_tasks_watcher_ids ON tasks USING GIN(watcher_ids);
CREATE INDEX idx_tasks_label_ids ON tasks USING GIN(label_ids);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_position ON tasks(project_id, position);

-- ============================================
-- TASK COMMENTS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS task_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    mentioned_users UUID[] DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_comments_task_id ON task_comments(task_id);
CREATE INDEX idx_task_comments_user_id ON task_comments(user_id);
CREATE INDEX idx_task_comments_created_at ON task_comments(created_at);

-- ============================================
-- TASK ATTACHMENTS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS task_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    filename VARCHAR(500) NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_attachments_task_id ON task_attachments(task_id);
CREATE INDEX idx_task_attachments_user_id ON task_attachments(user_id);

-- ============================================
-- TIME ENTRIES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_seconds INTEGER,
    description TEXT,
    is_manual BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_time_entries_task_id ON time_entries(task_id);
CREATE INDEX idx_time_entries_user_id ON time_entries(user_id);
CREATE INDEX idx_time_entries_start_time ON time_entries(start_time);
CREATE INDEX idx_time_entries_active ON time_entries(user_id, end_time) WHERE end_time IS NULL;

-- ============================================
-- TASK DEPENDENCIES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS task_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    dependency_type VARCHAR(50) NOT NULL DEFAULT 'blocks', -- blocks, is_blocked_by, relates_to
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(task_id, depends_on_task_id)
);

CREATE INDEX idx_task_dependencies_task_id ON task_dependencies(task_id);
CREATE INDEX idx_task_dependencies_depends_on ON task_dependencies(depends_on_task_id);

-- ============================================
-- TASK CHECKLISTS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS task_checklists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_checklists_task_id ON task_checklists(task_id);

-- ============================================
-- CHECKLIST ITEMS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS checklist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    checklist_id UUID NOT NULL REFERENCES task_checklists(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    assignee_id UUID REFERENCES users(id),
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_checklist_items_checklist_id ON checklist_items(checklist_id);
CREATE INDEX idx_checklist_items_assignee_id ON checklist_items(assignee_id);
CREATE INDEX idx_checklist_items_position ON checklist_items(checklist_id, position);

-- ============================================
-- TASK ACTIVITIES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS task_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id), -- Nullable for system actions
    action VARCHAR(100) NOT NULL, -- created, updated, status_changed, assigned, commented, etc.
    field_name VARCHAR(100),
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_activities_task_id ON task_activities(task_id);
CREATE INDEX idx_task_activities_user_id ON task_activities(user_id);
CREATE INDEX idx_task_activities_created_at ON task_activities(created_at DESC);
CREATE INDEX idx_task_activities_action ON task_activities(action);

-- ============================================
-- TRIGGERS FOR AUTOMATIC ACTIVITY LOGGING
-- ============================================

-- Function to log task creation
CREATE OR REPLACE FUNCTION log_task_creation()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO task_activities (task_id, user_id, action)
    VALUES (NEW.id, NEW.assignee_ids[1], 'created');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_log_task_creation
AFTER INSERT ON tasks
FOR EACH ROW
EXECUTE FUNCTION log_task_creation();

-- Function to log task updates
CREATE OR REPLACE FUNCTION log_task_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Log status changes
    IF OLD.status != NEW.status THEN
        INSERT INTO task_activities (task_id, action, field_name, old_value, new_value)
        VALUES (NEW.id, 'status_changed', 'status', OLD.status, NEW.status);
    END IF;

    -- Log priority changes
    IF OLD.priority != NEW.priority THEN
        INSERT INTO task_activities (task_id, action, field_name, old_value, new_value)
        VALUES (NEW.id, 'priority_changed', 'priority', OLD.priority, NEW.priority);
    END IF;

    -- Log assignee changes
    IF OLD.assignee_ids != NEW.assignee_ids THEN
        INSERT INTO task_activities (task_id, action, field_name, new_value)
        VALUES (NEW.id, 'assignees_changed', 'assignee_ids', array_to_string(NEW.assignee_ids, ','));
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_log_task_update
AFTER UPDATE ON tasks
FOR EACH ROW
EXECUTE FUNCTION log_task_update();

-- ============================================
-- VIEWS FOR COMMON QUERIES
-- ============================================

-- View for task summary with counts
CREATE OR REPLACE VIEW task_summary AS
SELECT 
    t.id,
    t.project_id,
    t.sprint_id,
    t.title,
    t.status,
    t.priority,
    t.story_points,
    t.due_date,
    CARDINALITY(t.assignee_ids) as assignee_count,
    (SELECT COUNT(*) FROM task_comments WHERE task_id = t.id) as comment_count,
    (SELECT COUNT(*) FROM task_attachments WHERE task_id = t.id) as attachment_count,
    (SELECT COUNT(*) FROM tasks WHERE parent_task_id = t.id) as subtask_count,
    t.created_at,
    t.updated_at
FROM tasks t;

-- View for sprint progress
CREATE OR REPLACE VIEW sprint_progress AS
SELECT 
    s.id as sprint_id,
    s.name as sprint_name,
    s.project_id,
    s.status as sprint_status,
    s.start_date,
    s.end_date,
    COUNT(t.id) as total_tasks,
    COUNT(CASE WHEN t.status = 'done' THEN 1 END) as completed_tasks,
    COALESCE(SUM(t.story_points), 0) as total_points,
    COALESCE(SUM(CASE WHEN t.status = 'done' THEN t.story_points ELSE 0 END), 0) as completed_points
FROM sprints s
LEFT JOIN tasks t ON t.sprint_id = s.id
GROUP BY s.id, s.name, s.project_id, s.status, s.start_date, s.end_date;

-- ============================================
-- COMMENTS
-- ============================================
COMMENT ON TABLE tasks IS 'Main tasks/issues table for project management';
COMMENT ON TABLE task_comments IS 'Comments on tasks';
COMMENT ON TABLE task_attachments IS 'File attachments for tasks';
COMMENT ON TABLE time_entries IS 'Time tracking entries for tasks';
COMMENT ON TABLE task_dependencies IS 'Task dependencies and blockers';
COMMENT ON TABLE task_checklists IS 'Checklists within tasks';
COMMENT ON TABLE checklist_items IS 'Individual checklist items';
COMMENT ON TABLE task_activities IS 'Audit log of all task activities';
COMMENT ON TABLE sprints IS 'Sprint/iteration information';
COMMENT ON TABLE projects IS 'Projects/boards';