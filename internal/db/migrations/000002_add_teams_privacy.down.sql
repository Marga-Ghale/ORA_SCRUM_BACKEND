-- Drop triggers
DROP TRIGGER IF EXISTS update_teams_updated_at ON teams;

-- Drop invitations table
DROP TABLE IF EXISTS invitations;

-- Drop activities table
DROP TABLE IF EXISTS activities;

-- Drop task watchers table
DROP TABLE IF EXISTS task_watchers;

-- Remove columns from projects
ALTER TABLE projects DROP COLUMN IF EXISTS visibility;
ALTER TABLE projects DROP COLUMN IF EXISTS allowed_teams;
ALTER TABLE projects DROP COLUMN IF EXISTS allowed_users;

-- Remove columns from spaces
ALTER TABLE spaces DROP COLUMN IF EXISTS visibility;
ALTER TABLE spaces DROP COLUMN IF EXISTS allowed_teams;
ALTER TABLE spaces DROP COLUMN IF EXISTS allowed_users;

-- Drop team members table
DROP TABLE IF EXISTS team_members;

-- Drop teams table
DROP TABLE IF EXISTS teams;
