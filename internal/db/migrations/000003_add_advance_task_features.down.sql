-- ============================================
-- DROP MIGRATION 000003 - Advanced Task Features
-- ============================================

DROP TRIGGER IF EXISTS update_checklists_updated_at ON checklists;
DROP TRIGGER IF EXISTS update_checklist_items_updated_at ON checklist_items;
DROP TRIGGER IF EXISTS update_time_entries_updated_at ON time_entries;

DROP TABLE IF EXISTS task_activity CASCADE;
DROP TABLE IF EXISTS checklist_items CASCADE;
DROP TABLE IF EXISTS checklists CASCADE;
DROP TABLE IF EXISTS task_dependencies CASCADE;
DROP TABLE IF EXISTS time_entries CASCADE;
DROP TABLE IF EXISTS attachments CASCADE;