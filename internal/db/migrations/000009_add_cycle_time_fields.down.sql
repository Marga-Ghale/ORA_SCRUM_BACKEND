ALTER TABLE tasks DROP COLUMN IF EXISTS started_at;
ALTER TABLE tasks DROP COLUMN IF EXISTS cycle_time_seconds;
ALTER TABLE tasks DROP COLUMN IF EXISTS lead_time_seconds;
DROP TABLE IF EXISTS task_status_history;
DROP TABLE IF EXISTS sprint_commitments;