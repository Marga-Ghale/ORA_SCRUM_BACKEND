-- Add Privacy to Spaces
ALTER TABLE spaces ADD COLUMN IF NOT EXISTS visibility VARCHAR(50) DEFAULT 'workspace';
ALTER TABLE spaces ADD COLUMN IF NOT EXISTS allowed_teams UUID[] DEFAULT '{}';
ALTER TABLE spaces ADD COLUMN IF NOT EXISTS allowed_users UUID[] DEFAULT '{}';

-- Add Privacy to Projects
ALTER TABLE projects ADD COLUMN IF NOT EXISTS visibility VARCHAR(50) DEFAULT 'space';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS allowed_teams UUID[] DEFAULT '{}';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS allowed_users UUID[] DEFAULT '{}';

-- Add extra columns to teams if missing
ALTER TABLE teams ADD COLUMN IF NOT EXISTS avatar TEXT;
ALTER TABLE teams ADD COLUMN IF NOT EXISTS color VARCHAR(50);
