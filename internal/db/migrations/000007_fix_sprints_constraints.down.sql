-- ============================================
-- DROP CONSTRAINTS
-- ============================================
ALTER TABLE sprints
DROP CONSTRAINT IF EXISTS sprints_dates_check;

ALTER TABLE sprints
DROP CONSTRAINT IF EXISTS sprints_status_check;

-- ============================================
-- DROP COLUMN
-- ============================================
ALTER TABLE sprints
DROP COLUMN IF EXISTS created_by;
