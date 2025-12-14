-- ============================================
-- Teams Table (Like ClickUp Teams)
-- ============================================
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    avatar TEXT,
    color VARCHAR(50),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_teams_workspace_id ON teams(workspace_id);
CREATE INDEX idx_teams_created_by ON teams(created_by);

-- ============================================
-- Team Members Table
-- ============================================
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(team_id, user_id)
);

CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);

-- ============================================
-- Add Privacy to Spaces (like ClickUp Spaces)
-- ============================================
ALTER TABLE spaces ADD COLUMN IF NOT EXISTS visibility VARCHAR(50) DEFAULT 'workspace';
ALTER TABLE spaces ADD COLUMN IF NOT EXISTS allowed_teams UUID[] DEFAULT '{}';
ALTER TABLE spaces ADD COLUMN IF NOT EXISTS allowed_users UUID[] DEFAULT '{}';

-- visibility options: 'workspace' (all workspace members), 'private' (only allowed), 'public' (anyone with link)
COMMENT ON COLUMN spaces.visibility IS 'workspace=all members, private=allowed only, public=anyone';

-- ============================================
-- Add Privacy to Projects (like ClickUp Lists)
-- ============================================
ALTER TABLE projects ADD COLUMN IF NOT EXISTS visibility VARCHAR(50) DEFAULT 'space';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS allowed_teams UUID[] DEFAULT '{}';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS allowed_users UUID[] DEFAULT '{}';

-- visibility options: 'space' (inherits from space), 'private' (only allowed), 'public' (anyone with link)
COMMENT ON COLUMN projects.visibility IS 'space=inherit, private=allowed only, public=anyone';

-- ============================================
-- Task Watchers (like ClickUp watchers)
-- ============================================
CREATE TABLE task_watchers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(task_id, user_id)
);

CREATE INDEX idx_task_watchers_task_id ON task_watchers(task_id);
CREATE INDEX idx_task_watchers_user_id ON task_watchers(user_id);

-- ============================================
-- Activity Log (audit trail)
-- ============================================
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    changes JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_activities_entity ON activities(entity_type, entity_id);
CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activities_created_at ON activities(created_at);

-- ============================================
-- Invitations Table (for email invitations)
-- ============================================
CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    target_id UUID NOT NULL,
    role VARCHAR(50) DEFAULT 'member',
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_invitations_email ON invitations(email);
CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_invitations_status ON invitations(status);

-- Apply updated_at trigger
CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
