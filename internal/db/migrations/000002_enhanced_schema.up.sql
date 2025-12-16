-- ============================================================================
-- ORA SCRUM BACKEND - ENHANCED DATABASE SCHEMA
-- Based on ClickUp, Jira, and Project Management Best Practices
-- ============================================================================
-- 
-- HIERARCHY: Workspace > Space > Folder (optional) > Project > Sprint > Task > Subtask
-- 
-- Key Improvements:
-- 1. Folders (optional layer between Space and Project like ClickUp)
-- 2. Multiple Assignees support
-- 3. Task Dependencies (blocking, waiting_on, related_to)
-- 4. Custom Fields system
-- 5. Checklists with items
-- 6. Attachments for tasks
-- 7. Time Tracking entries
-- 8. Enhanced invitation flow for registered/unregistered users
-- 9. Comprehensive notification system
-- 10. Comment mentions tracking
-- 11. Task templates
-- 12. User presence/status tracking
-- ============================================================================

-- Enable UUID extension (if not already enabled)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- 1. FOLDERS TABLE (Optional hierarchy level between Space and Project)
-- ============================================================================
-- ClickUp has: Workspace > Space > Folder > List > Task
-- We'll add Folders as optional containers for Projects

CREATE TABLE IF NOT EXISTS folders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(50),
    space_id UUID NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_archived BOOLEAN DEFAULT false,
    order_index INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_folders_space_id ON folders(space_id);
CREATE INDEX idx_folders_created_by ON folders(created_by);

-- Add folder_id to projects (nullable for projects directly under spaces)
ALTER TABLE projects ADD COLUMN IF NOT EXISTS folder_id UUID REFERENCES folders(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_projects_folder_id ON projects(folder_id);

-- ============================================================================
-- 2. MULTIPLE ASSIGNEES TABLE
-- ============================================================================
-- Support multiple people assigned to the same task

CREATE TABLE IF NOT EXISTS task_assignees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    is_primary BOOLEAN DEFAULT false, -- Primary assignee for notifications
    UNIQUE(task_id, user_id)
);
CREATE INDEX idx_task_assignees_task_id ON task_assignees(task_id);
CREATE INDEX idx_task_assignees_user_id ON task_assignees(user_id);

-- ============================================================================
-- 3. TASK DEPENDENCIES TABLE
-- ============================================================================
-- Types: blocking, waiting_on, related_to, duplicates, parent_of

CREATE TABLE IF NOT EXISTS task_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,           -- The dependent task
    depends_on_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE, -- The task it depends on
    dependency_type VARCHAR(50) NOT NULL DEFAULT 'blocking',
    -- Types:
    -- 'blocking' - This task blocks the dependent task
    -- 'waiting_on' - This task is waiting on another task
    -- 'related_to' - Tasks are related but not blocking
    -- 'duplicates' - This task duplicates another
    -- 'finish_to_start' - B cannot start until A finishes
    -- 'start_to_start' - B cannot start until A starts
    -- 'finish_to_finish' - B cannot finish until A finishes
    -- 'start_to_finish' - B cannot finish until A starts
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(task_id, depends_on_task_id),
    CHECK (task_id != depends_on_task_id) -- Prevent self-reference
);
CREATE INDEX idx_task_dependencies_task_id ON task_dependencies(task_id);
CREATE INDEX idx_task_dependencies_depends_on ON task_dependencies(depends_on_task_id);
CREATE INDEX idx_task_dependencies_type ON task_dependencies(dependency_type);

-- ============================================================================
-- 4. CUSTOM FIELDS SYSTEM
-- ============================================================================
-- Define custom fields at project/space level, store values per task

-- Custom field definitions
CREATE TABLE IF NOT EXISTS custom_field_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    field_type VARCHAR(50) NOT NULL,
    -- Types: text, number, date, dropdown, checkbox, url, email, phone, 
    --        currency, percentage, rating, people, labels, formula
    options JSONB DEFAULT '{}', -- For dropdown options, formula config, etc.
    default_value TEXT,
    is_required BOOLEAN DEFAULT false,
    -- Scope: where this field is available
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    space_id UUID REFERENCES spaces(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    -- At least one scope must be set
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_index INT DEFAULT 0,
    is_archived BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_custom_field_def_workspace ON custom_field_definitions(workspace_id);
CREATE INDEX idx_custom_field_def_space ON custom_field_definitions(space_id);
CREATE INDEX idx_custom_field_def_project ON custom_field_definitions(project_id);
CREATE INDEX idx_custom_field_def_type ON custom_field_definitions(field_type);

-- Custom field values per task
CREATE TABLE IF NOT EXISTS custom_field_values (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
    -- Store different value types
    text_value TEXT,
    number_value DECIMAL(20, 4),
    date_value TIMESTAMPTZ,
    boolean_value BOOLEAN,
    json_value JSONB, -- For arrays, objects, people lists, etc.
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(task_id, field_id)
);
CREATE INDEX idx_custom_field_values_task ON custom_field_values(task_id);
CREATE INDEX idx_custom_field_values_field ON custom_field_values(field_id);

-- ============================================================================
-- 5. CHECKLISTS AND CHECKLIST ITEMS
-- ============================================================================
-- Tasks can have multiple checklists, each with multiple items

CREATE TABLE IF NOT EXISTS checklists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL DEFAULT 'Checklist',
    order_index INT DEFAULT 0,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_checklists_task_id ON checklists(task_id);

CREATE TABLE IF NOT EXISTS checklist_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    checklist_id UUID NOT NULL REFERENCES checklists(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_completed BOOLEAN DEFAULT false,
    completed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    completed_at TIMESTAMPTZ,
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    due_date TIMESTAMPTZ,
    order_index INT DEFAULT 0,
    parent_id UUID REFERENCES checklist_items(id) ON DELETE CASCADE, -- Nested items
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_checklist_items_checklist ON checklist_items(checklist_id);
CREATE INDEX idx_checklist_items_assignee ON checklist_items(assignee_id);
CREATE INDEX idx_checklist_items_parent ON checklist_items(parent_id);

-- ============================================================================
-- 6. ATTACHMENTS TABLE
-- ============================================================================
-- Files attached to tasks, comments, or messages

CREATE TABLE IF NOT EXISTS attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    filename VARCHAR(500) NOT NULL,
    original_filename VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    storage_url TEXT NOT NULL,
    thumbnail_url TEXT,
    -- Polymorphic attachment (can attach to different entities)
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    chat_message_id UUID REFERENCES chat_messages(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_attachments_task ON attachments(task_id);
CREATE INDEX idx_attachments_comment ON attachments(comment_id);
CREATE INDEX idx_attachments_message ON attachments(chat_message_id);
CREATE INDEX idx_attachments_uploader ON attachments(uploaded_by);

-- ============================================================================
-- 7. TIME TRACKING / TIME ENTRIES
-- ============================================================================
-- Track time spent on tasks

CREATE TABLE IF NOT EXISTS time_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    description TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    duration_minutes INT, -- Can be calculated or manually entered
    is_billable BOOLEAN DEFAULT false,
    is_running BOOLEAN DEFAULT false, -- For active timers
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_time_entries_task ON time_entries(task_id);
CREATE INDEX idx_time_entries_user ON time_entries(user_id);
CREATE INDEX idx_time_entries_started ON time_entries(started_at);
CREATE INDEX idx_time_entries_running ON time_entries(is_running) WHERE is_running = true;

-- ============================================================================
-- 8. ENHANCED INVITATIONS TABLE
-- ============================================================================
-- Better support for registered vs unregistered user flows

-- Drop and recreate with enhancements (if safe to do so)
-- Or add columns to existing table:

ALTER TABLE invitations 
    ADD COLUMN IF NOT EXISTS invitee_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS declined_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS message TEXT, -- Personal invitation message
    ADD COLUMN IF NOT EXISTS invitation_type VARCHAR(50) DEFAULT 'email',
    -- Types: 'email', 'link', 'direct' (for existing users)
    ADD COLUMN IF NOT EXISTS reminder_sent_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reminder_count INT DEFAULT 0;

-- Status values: 'pending', 'accepted', 'declined', 'expired', 'cancelled'
COMMENT ON COLUMN invitations.status IS 'Status: pending, accepted, declined, expired, cancelled';
COMMENT ON COLUMN invitations.invitation_type IS 'Type: email (sent via email), link (shareable link), direct (existing user)';

-- ============================================================================
-- 9. COMMENT MENTIONS TABLE
-- ============================================================================
-- Track @mentions in comments for notifications

CREATE TABLE IF NOT EXISTS comment_mentions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    mentioned_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(comment_id, mentioned_user_id)
);
CREATE INDEX idx_comment_mentions_comment ON comment_mentions(comment_id);
CREATE INDEX idx_comment_mentions_user ON comment_mentions(mentioned_user_id);

-- ============================================================================
-- 10. CHAT MESSAGE MENTIONS TABLE  
-- ============================================================================
-- Track @mentions in chat messages

CREATE TABLE IF NOT EXISTS chat_message_mentions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    mentioned_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(message_id, mentioned_user_id)
);
CREATE INDEX idx_chat_mentions_message ON chat_message_mentions(message_id);
CREATE INDEX idx_chat_mentions_user ON chat_message_mentions(mentioned_user_id);

-- ============================================================================
-- 11. TASK TEMPLATES
-- ============================================================================
-- Save task configurations as reusable templates

CREATE TABLE IF NOT EXISTS task_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    -- Template content
    title_template VARCHAR(500),
    description_template TEXT,
    default_status VARCHAR(50),
    default_priority VARCHAR(50),
    default_type VARCHAR(50),
    default_story_points INT,
    default_labels TEXT[] DEFAULT '{}',
    checklist_template JSONB DEFAULT '[]', -- Array of checklist items
    custom_fields_template JSONB DEFAULT '{}', -- Pre-filled custom field values
    -- Scope
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT false, -- Visible to all project members
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_task_templates_project ON task_templates(project_id);
CREATE INDEX idx_task_templates_workspace ON task_templates(workspace_id);
CREATE INDEX idx_task_templates_creator ON task_templates(created_by);

-- ============================================================================
-- 12. USER PREFERENCES TABLE (Enhanced)
-- ============================================================================
-- Store user-specific preferences

CREATE TABLE IF NOT EXISTS user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    -- UI Preferences
    theme VARCHAR(20) DEFAULT 'light', -- light, dark, system
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    date_format VARCHAR(20) DEFAULT 'YYYY-MM-DD',
    time_format VARCHAR(10) DEFAULT '24h', -- 12h, 24h
    first_day_of_week INT DEFAULT 0, -- 0=Sunday, 1=Monday
    -- Default views
    default_workspace_id UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    default_project_view VARCHAR(50) DEFAULT 'board', -- board, list, table, timeline
    -- Notification preferences
    email_notifications BOOLEAN DEFAULT true,
    push_notifications BOOLEAN DEFAULT true,
    desktop_notifications BOOLEAN DEFAULT true,
    notification_sound BOOLEAN DEFAULT true,
    -- Working hours (for smart notifications)
    work_start_time TIME DEFAULT '09:00',
    work_end_time TIME DEFAULT '17:00',
    work_days INT[] DEFAULT '{1,2,3,4,5}', -- 0=Sun, 1=Mon, etc.
    -- Other
    compact_mode BOOLEAN DEFAULT false,
    show_completed_tasks BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_user_preferences_user ON user_preferences(user_id);

-- ============================================================================
-- 13. ENHANCED NOTIFICATIONS TABLE
-- ============================================================================
-- Add more notification types and delivery tracking

ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS entity_type VARCHAR(50), -- task, comment, project, sprint, etc.
    ADD COLUMN IF NOT EXISTS entity_id UUID,
    ADD COLUMN IF NOT EXISTS action VARCHAR(50), -- created, updated, deleted, mentioned, assigned, etc.
    ADD COLUMN IF NOT EXISTS actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS is_email_sent BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS email_sent_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS is_push_sent BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS push_sent_at TIMESTAMPTZ;

-- Common notification types:
-- task_assigned, task_unassigned
-- task_updated, task_status_changed
-- task_comment_added, task_mentioned
-- task_due_soon, task_overdue
-- sprint_started, sprint_ended
-- project_member_added, project_member_removed
-- invitation_received, invitation_accepted
-- chat_message, chat_mentioned

-- ============================================================================
-- 14. SPRINT CAPACITY PLANNING
-- ============================================================================
-- Track team capacity per sprint

CREATE TABLE IF NOT EXISTS sprint_capacity (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capacity_hours DECIMAL(5, 2) DEFAULT 40, -- Available hours for the sprint
    planned_hours DECIMAL(5, 2) DEFAULT 0, -- Planned work hours
    logged_hours DECIMAL(5, 2) DEFAULT 0, -- Actual logged hours
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(sprint_id, user_id)
);
CREATE INDEX idx_sprint_capacity_sprint ON sprint_capacity(sprint_id);
CREATE INDEX idx_sprint_capacity_user ON sprint_capacity(user_id);

-- ============================================================================
-- 15. TASK HISTORY / AUDIT LOG
-- ============================================================================
-- Track all changes to tasks for audit purposes

CREATE TABLE IF NOT EXISTS task_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    field_name VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    change_type VARCHAR(50) NOT NULL, -- created, updated, deleted
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_task_history_task ON task_history(task_id);
CREATE INDEX idx_task_history_user ON task_history(user_id);
CREATE INDEX idx_task_history_created ON task_history(created_at);

-- ============================================================================
-- 16. SAVED FILTERS / VIEWS
-- ============================================================================
-- Save custom filters/views for projects

CREATE TABLE IF NOT EXISTS saved_views (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    view_type VARCHAR(50) NOT NULL DEFAULT 'list', -- list, board, table, timeline, calendar
    -- Scope
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    space_id UUID REFERENCES spaces(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    -- Filter/sort configuration
    filters JSONB DEFAULT '{}', -- Filter criteria
    sort_by VARCHAR(100),
    sort_order VARCHAR(10) DEFAULT 'asc',
    group_by VARCHAR(100),
    columns JSONB DEFAULT '[]', -- Visible columns for table view
    -- Permissions
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_default BOOLEAN DEFAULT false,
    is_public BOOLEAN DEFAULT false, -- Visible to all members
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_saved_views_project ON saved_views(project_id);
CREATE INDEX idx_saved_views_space ON saved_views(space_id);
CREATE INDEX idx_saved_views_workspace ON saved_views(workspace_id);
CREATE INDEX idx_saved_views_creator ON saved_views(created_by);

-- ============================================================================
-- 17. WEBHOOKS TABLE
-- ============================================================================
-- For integrations and external notifications

CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    secret VARCHAR(255), -- For signature verification
    events TEXT[] NOT NULL, -- Array of event types to trigger
    -- Scope
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    -- Status
    is_active BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMPTZ,
    failure_count INT DEFAULT 0,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_webhooks_workspace ON webhooks(workspace_id);
CREATE INDEX idx_webhooks_project ON webhooks(project_id);
CREATE INDEX idx_webhooks_active ON webhooks(is_active) WHERE is_active = true;

-- ============================================================================
-- 18. ENHANCE EXISTING TABLES
-- ============================================================================

-- Add epic support to tasks (tasks can be epics containing other tasks)
ALTER TABLE tasks 
    ADD COLUMN IF NOT EXISTS is_epic BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS epic_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS estimated_hours DECIMAL(8, 2),
    ADD COLUMN IF NOT EXISTS actual_hours DECIMAL(8, 2),
    ADD COLUMN IF NOT EXISTS start_date TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_tasks_epic_id ON tasks(epic_id);
CREATE INDEX IF NOT EXISTS idx_tasks_is_epic ON tasks(is_epic) WHERE is_epic = true;
CREATE INDEX IF NOT EXISTS idx_tasks_is_archived ON tasks(is_archived);

-- Add more user tracking fields
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone VARCHAR(50),
    ADD COLUMN IF NOT EXISTS job_title VARCHAR(255),
    ADD COLUMN IF NOT EXISTS department VARCHAR(255),
    ADD COLUMN IF NOT EXISTS bio TEXT,
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMPTZ;

-- Add sprint enhancements
ALTER TABLE sprints
    ADD COLUMN IF NOT EXISTS velocity_planned INT,
    ADD COLUMN IF NOT EXISTS velocity_completed INT,
    ADD COLUMN IF NOT EXISTS retrospective_notes TEXT,
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT false;

-- Add project enhancements  
ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS default_assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS default_status VARCHAR(50) DEFAULT 'backlog',
    ADD COLUMN IF NOT EXISTS default_priority VARCHAR(50) DEFAULT 'medium';

-- Add workspace enhancements
ALTER TABLE workspaces
    ADD COLUMN IF NOT EXISTS logo_url TEXT,
    ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT false;

-- ============================================================================
-- 19. PROJECT STATUSES (Custom per project)
-- ============================================================================

CREATE TABLE IF NOT EXISTS project_statuses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(50) DEFAULT '#808080',
    category VARCHAR(50) NOT NULL DEFAULT 'custom',
    -- Categories: backlog, todo, in_progress, review, done, cancelled
    order_index INT DEFAULT 0,
    is_default BOOLEAN DEFAULT false,
    is_done BOOLEAN DEFAULT false, -- Marks task as completed
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, name)
);
CREATE INDEX idx_project_statuses_project ON project_statuses(project_id);

-- ============================================================================
-- 20. UPDATE TRIGGERS FOR NEW TABLES
-- ============================================================================

CREATE TRIGGER update_folders_updated_at 
    BEFORE UPDATE ON folders 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_custom_field_def_updated_at 
    BEFORE UPDATE ON custom_field_definitions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_custom_field_values_updated_at 
    BEFORE UPDATE ON custom_field_values 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_checklists_updated_at 
    BEFORE UPDATE ON checklists 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_checklist_items_updated_at 
    BEFORE UPDATE ON checklist_items 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_time_entries_updated_at 
    BEFORE UPDATE ON time_entries 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_task_templates_updated_at 
    BEFORE UPDATE ON task_templates 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_preferences_updated_at 
    BEFORE UPDATE ON user_preferences 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sprint_capacity_updated_at 
    BEFORE UPDATE ON sprint_capacity 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_saved_views_updated_at 
    BEFORE UPDATE ON saved_views 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webhooks_updated_at 
    BEFORE UPDATE ON webhooks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- NOTIFICATION TYPE REFERENCE
-- ============================================================================
-- 
-- Here's a comprehensive list of notification types your app should support:
--
-- TASK NOTIFICATIONS:
-- - task_created: When a task is created in a project you're watching
-- - task_assigned: When you're assigned to a task
-- - task_unassigned: When you're removed from a task
-- - task_updated: When a task you're watching is updated
-- - task_status_changed: When task status changes
-- - task_priority_changed: When task priority changes
-- - task_due_soon: Task due date approaching (24h, 1h before)
-- - task_overdue: Task is past due date
-- - task_completed: Task you created/watching is completed
-- - task_mentioned: When you're @mentioned in task description
--
-- COMMENT NOTIFICATIONS:
-- - comment_added: New comment on a task you're watching
-- - comment_mentioned: You're @mentioned in a comment
-- - comment_reply: Reply to your comment
--
-- SPRINT NOTIFICATIONS:
-- - sprint_started: Sprint has started
-- - sprint_ending_soon: Sprint ending within 24h
-- - sprint_completed: Sprint has ended
-- - sprint_task_added: Task added to active sprint
--
-- PROJECT NOTIFICATIONS:
-- - project_member_added: You're added to a project
-- - project_member_removed: You're removed from a project
-- - project_role_changed: Your role in project changed
--
-- INVITATION NOTIFICATIONS:
-- - invitation_received: You received an invitation
-- - invitation_accepted: Your invitation was accepted
-- - invitation_declined: Your invitation was declined
-- - invitation_expired: Your invitation expired
--
-- TEAM NOTIFICATIONS:
-- - team_member_added: Added to a team
-- - team_member_removed: Removed from a team
--
-- CHAT NOTIFICATIONS:
-- - chat_message: New message in channel you follow
-- - chat_mentioned: @mentioned in chat
-- - chat_reaction: Reaction to your message
-- - direct_message: New direct message
--
-- ============================================================================

-- ============================================================================
-- END OF ENHANCED SCHEMA
-- ============================================================================