-- ============================================================================
-- FIX: Velocity Tracking System for ORA Scrum
-- ============================================================================

-- ---------------------------------------------------------------------------
-- CLEAN EXISTING OBJECTS (SAFE)
-- ---------------------------------------------------------------------------
DROP TRIGGER IF EXISTS trigger_record_sprint_velocity ON sprints;
DROP FUNCTION IF EXISTS record_sprint_velocity();

-- ---------------------------------------------------------------------------
-- FUNCTION: Calculate and Record Sprint Velocity
-- ---------------------------------------------------------------------------
CREATE FUNCTION record_sprint_velocity()
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
    IF NEW.status = 'completed'
       AND (OLD.status IS DISTINCT FROM 'completed') THEN

        v_project_id := NEW.project_id;

        -- Sprint number per project
        SELECT COUNT(*) INTO v_sprint_number
        FROM sprints
        WHERE project_id = v_project_id
          AND created_at <= NEW.created_at;

        -- Task metrics
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

        -- Goal metrics
        SELECT
            COUNT(*) FILTER (WHERE status = 'completed'),
            COUNT(*)
        INTO
            v_goals_completed,
            v_goals_total
        FROM goals
        WHERE sprint_id = NEW.id;

        -- Velocity history
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

        -- Sprint report
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
            v_completed_points,
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
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------------------------------------------------------------------------
-- TRIGGER
-- ---------------------------------------------------------------------------
CREATE TRIGGER trigger_record_sprint_velocity
AFTER UPDATE OF status ON sprints
FOR EACH ROW
EXECUTE FUNCTION record_sprint_velocity();

-- ---------------------------------------------------------------------------
-- BACKFILL EXISTING COMPLETED SPRINTS
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    s RECORD;
    v_committed_tasks INTEGER;
    v_committed_points INTEGER;
    v_completed_tasks INTEGER;
    v_completed_points INTEGER;
    v_incomplete_tasks INTEGER;
    v_incomplete_points INTEGER;
    v_sprint_number INTEGER;
    v_goals_completed INTEGER;
    v_goals_total INTEGER;
    v_avg_cycle_time DECIMAL(10,2);
    v_avg_lead_time DECIMAL(10,2);
BEGIN
    FOR s IN
        SELECT sp.*
        FROM sprints sp
        LEFT JOIN velocity_history vh ON vh.sprint_id = sp.id
        WHERE sp.status = 'completed'
          AND vh.id IS NULL
        ORDER BY sp.created_at
    LOOP
        SELECT COUNT(*) INTO v_sprint_number
        FROM sprints
        WHERE project_id = s.project_id
          AND created_at <= s.created_at;

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
        WHERE sprint_id = s.id;

        SELECT
            COUNT(*) FILTER (WHERE status = 'completed'),
            COUNT(*)
        INTO
            v_goals_completed,
            v_goals_total
        FROM goals
        WHERE sprint_id = s.id;

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
            s.project_id,
            s.id,
            s.name,
            v_sprint_number,
            v_committed_points,
            v_completed_points,
            s.start_date::date,
            s.end_date::date,
            s.updated_at
        );

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
            s.id,
            v_committed_tasks,
            v_committed_points,
            v_completed_tasks,
            v_completed_points,
            v_incomplete_tasks,
            v_incomplete_points,
            v_completed_points,
            v_avg_cycle_time,
            v_avg_lead_time,
            v_goals_completed,
            v_goals_total,
            s.updated_at
        );
    END LOOP;
END $$;
