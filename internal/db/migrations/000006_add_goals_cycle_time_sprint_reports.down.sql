-- ============================================
-- DROP TRIGGERS & FUNCTIONS (in reverse order)
-- ============================================
DROP TRIGGER IF EXISTS trigger_record_sprint_velocity ON sprints;
DROP FUNCTION IF EXISTS record_sprint_velocity();

DROP TRIGGER IF EXISTS task_status_change_trigger ON tasks;
DROP FUNCTION IF EXISTS track_task_status_change();

-- ============================================
-- REMOVE TASK COLUMNS
-- ============================================
ALTER TABLE tasks DROP COLUMN IF EXISTS started_at;
ALTER TABLE tasks DROP COLUMN IF EXISTS cycle_time_seconds;
ALTER TABLE tasks DROP COLUMN IF EXISTS lead_time_seconds;

-- ============================================
-- DROP TABLES (reverse dependency order)
-- ============================================
DROP TABLE IF EXISTS velocity_history;
DROP TABLE IF EXISTS sprint_reports;
DROP TABLE IF EXISTS task_status_history;
DROP TABLE IF EXISTS goal_tasks;
DROP TABLE IF EXISTS goal_key_results;
DROP TABLE IF EXISTS goals;