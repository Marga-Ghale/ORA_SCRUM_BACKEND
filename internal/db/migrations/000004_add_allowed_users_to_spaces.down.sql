-- ============================================
-- 000007_add_performance_indexes.down.sql
-- ============================================

DROP INDEX IF EXISTS idx_tasks_assignee_ids;
DROP INDEX IF EXISTS idx_tasks_watcher_ids;
DROP INDEX IF EXISTS idx_tasks_label_ids;
DROP INDEX IF EXISTS idx_tasks_due_date;
DROP INDEX IF EXISTS idx_tasks_start_date;
DROP INDEX IF EXISTS idx_tasks_completed_at;
DROP INDEX IF EXISTS idx_sprints_dates;
DROP INDEX IF EXISTS idx_sprints_project_status;
DROP INDEX IF EXISTS idx_time_entries_active_timer;
DROP INDEX IF EXISTS idx_notifications_user_unread;
DROP INDEX IF EXISTS idx_task_activity_user_id;
DROP INDEX IF EXISTS idx_task_activity_task_created;
DROP INDEX IF EXISTS idx_workspaces_allowed_users;
DROP INDEX IF EXISTS idx_workspaces_allowed_teams;
DROP INDEX IF EXISTS idx_spaces_allowed_users;
DROP INDEX IF EXISTS idx_spaces_allowed_teams;
DROP INDEX IF EXISTS idx_folders_allowed_users;
DROP INDEX IF EXISTS idx_folders_allowed_teams;
DROP INDEX IF EXISTS idx_projects_allowed_users;
DROP INDEX IF EXISTS idx_projects_allowed_teams;
DROP INDEX IF EXISTS idx_invitations_token_status;
DROP INDEX IF EXISTS idx_chat_messages_metadata;
DROP INDEX IF EXISTS idx_tasks_project_status_priority;