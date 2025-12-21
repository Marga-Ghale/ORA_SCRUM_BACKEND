
-- ============================================
-- 000007_add_performance_indexes.up.sql
-- ============================================
-- Add GIN indexes for array columns and other performance indexes

-- Task array columns for searching assignees, watchers, labels
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_ids ON tasks USING GIN(assignee_ids);
CREATE INDEX IF NOT EXISTS idx_tasks_watcher_ids ON tasks USING GIN(watcher_ids);
CREATE INDEX IF NOT EXISTS idx_tasks_label_ids ON tasks USING GIN(label_ids);

-- Task date queries for filtering
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_start_date ON tasks(start_date) WHERE start_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_completed_at ON tasks(completed_at) WHERE completed_at IS NOT NULL;

-- Sprint date range queries
CREATE INDEX IF NOT EXISTS idx_sprints_dates ON sprints(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_sprints_project_status ON sprints(project_id, status);

-- Active timer lookup (end_time IS NULL means timer is running)
CREATE INDEX IF NOT EXISTS idx_time_entries_active_timer ON time_entries(user_id, end_time) WHERE end_time IS NULL;

-- Notification queries (unread notifications by user)
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, read, created_at DESC);

-- Activity log queries
CREATE INDEX IF NOT EXISTS idx_task_activity_user_id ON task_activity(user_id);
CREATE INDEX IF NOT EXISTS idx_task_activity_task_created ON task_activity(task_id, created_at DESC);

-- Workspace/Space/Folder visibility queries
CREATE INDEX IF NOT EXISTS idx_workspaces_allowed_users ON workspaces USING GIN(allowed_users);
CREATE INDEX IF NOT EXISTS idx_workspaces_allowed_teams ON workspaces USING GIN(allowed_teams);
CREATE INDEX IF NOT EXISTS idx_spaces_allowed_users ON spaces USING GIN(allowed_users);
CREATE INDEX IF NOT EXISTS idx_spaces_allowed_teams ON spaces USING GIN(allowed_teams);
CREATE INDEX IF NOT EXISTS idx_folders_allowed_users ON folders USING GIN(allowed_users);
CREATE INDEX IF NOT EXISTS idx_folders_allowed_teams ON folders USING GIN(allowed_teams);
CREATE INDEX IF NOT EXISTS idx_projects_allowed_users ON projects USING GIN(allowed_users);
CREATE INDEX IF NOT EXISTS idx_projects_allowed_teams ON projects USING GIN(allowed_teams);

-- Invitation lookup by token and status
CREATE INDEX IF NOT EXISTS idx_invitations_token_status ON invitations(token, status);

-- Chat message metadata (JSONB)
CREATE INDEX IF NOT EXISTS idx_chat_messages_metadata ON chat_messages USING GIN(metadata);

-- Composite index for task filtering
CREATE INDEX IF NOT EXISTS idx_tasks_project_status_priority ON tasks(project_id, status, priority);
