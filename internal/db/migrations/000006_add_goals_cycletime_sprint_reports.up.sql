-- Migration: Add Goals, Cycle Time Tracking, Sprint Reports
-- Run this migration to add the new tables

-- ============================================
-- GOALS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    sprint_id UUID REFERENCES sprints(id) ON DELETE CASCADE,
    
    title VARCHAR(255) NOT NULL,
    description TEXT,
    goal_type VARCHAR(50) NOT NULL DEFAULT 'sprint', -- 'sprint', 'project', 'quarterly', 'annual'
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'completed', 'cancelled', 'at_risk'
    
    target_value DECIMAL(10,2),
    current_value DECIMAL(10,2) DEFAULT 0,
    unit VARCHAR(50), -- 'story_points', 'tasks', 'percentage', 'custom'
    
    start_date DATE,
    target_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_goals_workspace ON goals(workspace_id);
CREATE INDEX idx_goals_project ON goals(project_id);
CREATE INDEX idx_goals_sprint ON goals(sprint_id);
CREATE INDEX idx_goals_status ON goals(status);

-- ============================================
-- GOAL KEY RESULTS (OKR Style)
-- ============================================
CREATE TABLE IF NOT EXISTS goal_key_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    target_value DECIMAL(10,2) NOT NULL,
    current_value DECIMAL(10,2) DEFAULT 0,
    unit VARCHAR(50),
    
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'missed'
    weight DECIMAL(5,2) DEFAULT 1.0, -- for weighted progress calculation
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_goal_key_results_goal ON goal_key_results(goal_id);

-- ============================================
-- GOAL LINKED TASKS
-- ============================================
CREATE TABLE IF NOT EXISTS goal_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(goal_id, task_id)
);

-- ============================================
-- TASK STATUS HISTORY (For Cycle Time)
-- ============================================
CREATE TABLE IF NOT EXISTS task_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_task_status_history_task ON task_status_history(task_id);
CREATE INDEX idx_task_status_history_changed_at ON task_status_history(changed_at);
CREATE INDEX idx_task_status_history_to_status ON task_status_history(to_status);

-- ============================================
-- SPRINT REPORTS (Cached/Stored)
-- ============================================
CREATE TABLE IF NOT EXISTS sprint_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    
    -- Commitment
    committed_tasks INT DEFAULT 0,
    committed_points INT DEFAULT 0,
    
    -- Completed
    completed_tasks INT DEFAULT 0,
    completed_points INT DEFAULT 0,
    
    -- Incomplete
    incomplete_tasks INT DEFAULT 0,
    incomplete_points INT DEFAULT 0,
    
    -- Added mid-sprint
    added_tasks INT DEFAULT 0,
    added_points INT DEFAULT 0,
    
    -- Removed mid-sprint
    removed_tasks INT DEFAULT 0,
    removed_points INT DEFAULT 0,
    
    -- Carryover (moved to next sprint)
    carryover_tasks INT DEFAULT 0,
    carryover_points INT DEFAULT 0,
    
    -- Time metrics (in hours)
    total_estimated_hours DECIMAL(10,2) DEFAULT 0,
    total_logged_hours DECIMAL(10,2) DEFAULT 0,
    
    -- Averages
    avg_cycle_time_hours DECIMAL(10,2), -- avg time from in_progress to done
    avg_lead_time_hours DECIMAL(10,2),  -- avg time from created to done
    
    -- Velocity
    velocity INT DEFAULT 0, -- completed story points
    
    -- Goals
    goals_completed INT DEFAULT 0,
    goals_total INT DEFAULT 0,
    
    -- Generated at
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(sprint_id)
);

-- ============================================
-- VELOCITY HISTORY (For Trend)
-- ============================================
CREATE TABLE IF NOT EXISTS velocity_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    
    sprint_name VARCHAR(255),
    sprint_number INT,
    
    committed_points INT DEFAULT 0,
    completed_points INT DEFAULT 0,
    
    start_date DATE,
    end_date DATE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(sprint_id)
);

CREATE INDEX idx_velocity_history_project ON velocity_history(project_id);

-- ============================================
-- ADD CYCLE TIME FIELDS TO TASKS
-- ============================================
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS started_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS cycle_time_seconds INT;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS lead_time_seconds INT;

-- ============================================
-- TRIGGER: Auto-track status changes
-- ============================================
CREATE OR REPLACE FUNCTION track_task_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Only track if status actually changed
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO task_status_history (task_id, from_status, to_status, changed_at)
        VALUES (NEW.id, OLD.status, NEW.status, NOW());
        
        -- Track started_at when moving to in_progress
        IF NEW.status = 'in_progress' AND OLD.status != 'in_progress' AND NEW.started_at IS NULL THEN
            NEW.started_at = NOW();
        END IF;
        
        -- Calculate cycle time when completed
        IF NEW.status = 'done' AND OLD.status != 'done' THEN
            IF NEW.started_at IS NOT NULL THEN
                NEW.cycle_time_seconds = EXTRACT(EPOCH FROM (NOW() - NEW.started_at))::INT;
            END IF;
            NEW.lead_time_seconds = EXTRACT(EPOCH FROM (NOW() - NEW.created_at))::INT;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS task_status_change_trigger ON tasks;
CREATE TRIGGER task_status_change_trigger
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION track_task_status_change();