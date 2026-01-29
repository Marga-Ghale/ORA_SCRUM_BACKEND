-- Add cycle time tracking fields if not exists
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS started_at TIMESTAMP;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS cycle_time_seconds INTEGER;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS lead_time_seconds INTEGER;

-- Create task_status_history table if not exists
CREATE TABLE IF NOT EXISTS task_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_by UUID REFERENCES users(id),
    changed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_status_history_task_id ON task_status_history(task_id);
CREATE INDEX IF NOT EXISTS idx_task_status_history_changed_at ON task_status_history(changed_at);

-- Create sprint commitment snapshot table
CREATE TABLE IF NOT EXISTS sprint_commitments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    committed_tasks INTEGER NOT NULL DEFAULT 0,
    committed_points INTEGER NOT NULL DEFAULT 0,
    snapshot_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(sprint_id)
);