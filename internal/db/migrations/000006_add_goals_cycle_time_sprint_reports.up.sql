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
    goal_type VARCHAR(50) NOT NULL DEFAULT 'sprint',
    status VARCHAR(50) NOT NULL DEFAULT 'active',

    target_value DECIMAL(10,2),
    current_value DECIMAL(10,2) DEFAULT 0,
    unit VARCHAR(50),

    start_date DATE,
    target_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,

    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_goals_workspace ON goals(workspace_id);
CREATE INDEX IF NOT EXISTS idx_goals_project ON goals(project_id);
CREATE INDEX IF NOT EXISTS idx_goals_sprint ON goals(sprint_id);
CREATE INDEX IF NOT EXISTS idx_goals_status ON goals(status);

-- ============================================
-- GOAL KEY RESULTS
-- ============================================
CREATE TABLE IF NOT EXISTS goal_key_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,

    title VARCHAR(255) NOT NULL,
    description TEXT,

    target_value DECIMAL(10,2) NOT NULL,
    current_value DECIMAL(10,2) DEFAULT 0,
    unit VARCHAR(50),

    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    weight DECIMAL(5,2) DEFAULT 1.0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_goal_key_results_goal ON goal_key_results(goal_id);

-- ============================================
-- GOAL TASKS
-- ============================================
CREATE TABLE IF NOT EXISTS goal_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(goal_id, task_id)
);

-- ============================================
-- TASK STATUS HISTORY
-- ============================================
CREATE TABLE IF NOT EXISTS task_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,

    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_status_history_task ON task_status_history(task_id);
CREATE INDEX IF NOT EXISTS idx_task_status_history_changed_at ON task_status_history(changed_at);
CREATE INDEX IF NOT EXISTS idx_task_status_history_to_status ON task_status_history(to_status);

-- ============================================
-- SPRINT REPORTS
-- ============================================
CREATE TABLE IF NOT EXISTS sprint_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,

    committed_tasks INT DEFAULT 0,
    committed_points INT DEFAULT 0,

    completed_tasks INT DEFAULT 0,
    completed_points INT DEFAULT 0,

    incomplete_tasks INT DEFAULT 0,
    incomplete_points INT DEFAULT 0,

    added_tasks INT DEFAULT 0,
    added_points INT DEFAULT 0,

    removed_tasks INT DEFAULT 0,
    removed_points INT DEFAULT 0,

    carryover_tasks INT DEFAULT 0,
    carryover_points INT DEFAULT 0,

    total_estimated_hours DECIMAL(10,2) DEFAULT 0,
    total_logged_hours DECIMAL(10,2) DEFAULT 0,

    avg_cycle_time_hours DECIMAL(10,2),
    avg_lead_time_hours DECIMAL(10,2),

    velocity INT DEFAULT 0,

    goals_completed INT DEFAULT 0,
    goals_total INT DEFAULT 0,

    generated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(sprint_id)
);

-- ============================================
-- VELOCITY HISTORY
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

CREATE INDEX IF NOT EXISTS idx_velocity_history_project ON velocity_history(project_id);

-- ============================================
-- TASK METRICS COLUMNS
-- ============================================
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS started_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS cycle_time_seconds INT;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS lead_time_seconds INT;

-- ============================================
-- FUNCTION: Track Task Status Changes
-- ============================================
CREATE OR REPLACE FUNCTION track_task_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO task_status_history (task_id, from_status, to_status, changed_at)
        VALUES (NEW.id, OLD.status, NEW.status, NOW());

        IF NEW.status = 'in_progress'
           AND OLD.status != 'in_progress'
           AND NEW.started_at IS NULL THEN
            NEW.started_at = NOW();
        END IF;

        IF NEW.status = 'done' AND OLD.status != 'done' THEN
            IF NEW.started_at IS NOT NULL THEN
                NEW.cycle_time_seconds =
                    EXTRACT(EPOCH FROM (NOW() - NEW.started_at))::INT;
            END IF;

            NEW.lead_time_seconds =
                EXTRACT(EPOCH FROM (NOW() - NEW.created_at))::INT;
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

-- ============================================
-- FUNCTION: Record Sprint Velocity
-- ============================================
CREATE OR REPLACE FUNCTION record_sprint_velocity()
RETURNS TRIGGER AS $$
DECLARE
    v_committed_points INTEGER;
    v_completed_points INTEGER;
    v_project_id UUID;
    v_sprint_number INTEGER;
    v_committed_tasks INTEGER;
    v_completed_tasks INTEGER;
    v_incomplete_tasks INTEGER;
    v_incomplete_points INTEGER;
    v_goals_completed INTEGER;
    v_goals_total INTEGER;
    v_avg_cycle_time DECIMAL(10,2);
    v_avg_lead_time DECIMAL(10,2);
BEGIN
    -- Only proceed if sprint is being marked as 'completed'
    IF NEW.status = 'completed' AND (OLD.status IS NULL OR OLD.status != 'completed') THEN
        
        -- Get project_id from the sprint
        v_project_id := NEW.project_id;
        
        -- Calculate sprint number (count of all sprints for this project up to this one)
        SELECT COUNT(*) INTO v_sprint_number
        FROM sprints
        WHERE project_id = v_project_id
        AND created_at <= NEW.created_at;
        
        -- Calculate task metrics
        SELECT 
            COUNT(*),
            COALESCE(SUM(story_points), 0),
            COUNT(*) FILTER (WHERE status = 'done'),
            COALESCE(SUM(story_points) FILTER (WHERE status = 'done'), 0),
            COUNT(*) FILTER (WHERE status != 'done'),
            COALESCE(SUM(story_points) FILTER (WHERE status != 'done'), 0),
            AVG(cycle_time_seconds / 3600.0) FILTER (WHERE cycle_time_seconds IS NOT NULL),
            AVG(lead_time_seconds / 3600.0) FILTER (WHERE lead_time_seconds IS NOT NULL)
        INTO 
            v_committed_tasks,
            v_committed_points,
            v_completed_tasks,
            v_completed_points,
            v_incomplete_tasks,
            v_incomplete_points,
            v_avg_cycle_time,
            v_avg_lead_time
        FROM tasks
        WHERE sprint_id = NEW.id;
        
        -- Calculate goal metrics
        SELECT 
            COUNT(*) FILTER (WHERE status = 'completed'),
            COUNT(*)
        INTO 
            v_goals_completed,
            v_goals_total
        FROM goals
        WHERE sprint_id = NEW.id;
        
        -- Insert into velocity_history (if not already exists)
        INSERT INTO velocity_history (
            project_id,
            sprint_id,
            sprint_name,
            sprint_number,
            committed_points,
            completed_points,
            start_date,
            end_date,
            created_at
        ) VALUES (
            v_project_id,
            NEW.id,
            NEW.name,
            v_sprint_number,
            v_committed_points,
            v_completed_points,
            NEW.start_date::date,
            NEW.end_date::date,
            NOW()
        )
        ON CONFLICT (sprint_id) DO UPDATE SET
            committed_points = EXCLUDED.committed_points,
            completed_points = EXCLUDED.completed_points,
            sprint_number = EXCLUDED.sprint_number,
            created_at = NOW();
        
        -- Generate comprehensive sprint report
        INSERT INTO sprint_reports (
            sprint_id,
            committed_tasks,
            committed_points,
            completed_tasks,
            completed_points,
            incomplete_tasks,
            incomplete_points,
            velocity,
            avg_cycle_time_hours,
            avg_lead_time_hours,
            goals_completed,
            goals_total,
            generated_at
        ) VALUES (
            NEW.id,
            v_committed_tasks,
            v_committed_points,
            v_completed_tasks,
            v_completed_points,
            v_incomplete_tasks,
            v_incomplete_points,
            v_completed_points, -- velocity = completed points
            v_avg_cycle_time,
            v_avg_lead_time,
            v_goals_completed,
            v_goals_total,
            NOW()
        )
        ON CONFLICT (sprint_id) DO UPDATE SET
            committed_tasks = EXCLUDED.committed_tasks,
            committed_points = EXCLUDED.committed_points,
            completed_tasks = EXCLUDED.completed_tasks,
            completed_points = EXCLUDED.completed_points,
            incomplete_tasks = EXCLUDED.incomplete_tasks,
            incomplete_points = EXCLUDED.incomplete_points,
            velocity = EXCLUDED.velocity,
            avg_cycle_time_hours = EXCLUDED.avg_cycle_time_hours,
            avg_lead_time_hours = EXCLUDED.avg_lead_time_hours,
            goals_completed = EXCLUDED.goals_completed,
            goals_total = EXCLUDED.goals_total,
            generated_at = NOW();
        
        RAISE NOTICE '✅ Velocity recorded for sprint "%": % / % points completed', 
            NEW.name, 
            v_completed_points, 
            v_committed_points;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_record_sprint_velocity ON sprints;

CREATE TRIGGER trigger_record_sprint_velocity
AFTER UPDATE OF status ON sprints
FOR EACH ROW
EXECUTE FUNCTION record_sprint_velocity();