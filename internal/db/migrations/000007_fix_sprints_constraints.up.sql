-- ============================================
-- ADD created_by COLUMN
-- ============================================
ALTER TABLE sprints
ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);

-- NOTE:
-- We do NOT add NOT NULL immediately to avoid breaking existing rows.
-- If you want NOT NULL, backfill first, then add constraint.

-- ============================================
-- ADD STATUS CHECK CONSTRAINT
-- ============================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'sprints_status_check'
    ) THEN
        ALTER TABLE sprints
        ADD CONSTRAINT sprints_status_check
        CHECK (status IN ('planning', 'active', 'completed'));
    END IF;
END$$;

-- ============================================
-- ADD DATE VALIDATION CONSTRAINT
-- ============================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'sprints_dates_check'
    ) THEN
        ALTER TABLE sprints
        ADD CONSTRAINT sprints_dates_check
        CHECK (end_date > start_date);
    END IF;
END$$;
