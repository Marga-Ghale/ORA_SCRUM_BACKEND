-- ---------------------------------------------------------------------------
-- DROP TRIGGER & FUNCTION
-- ---------------------------------------------------------------------------
DROP TRIGGER IF EXISTS trigger_record_sprint_velocity ON sprints;
DROP FUNCTION IF EXISTS record_sprint_velocity();

-- ---------------------------------------------------------------------------
-- OPTIONAL CLEANUP (KEEP DATA BY DEFAULT)
-- Uncomment ONLY if you want full rollback
-- ---------------------------------------------------------------------------
-- DELETE FROM sprint_reports;
-- DELETE FROM velocity_history;
