-- ============================================
-- Chat System Migration
-- ClickUp-like Real-time Chat
-- ============================================

-- Chat Channels Table
CREATE TABLE IF NOT EXISTS chat_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'project', 'space', 'team', 'direct'
    target_id VARCHAR(255) NOT NULL, -- ID of project/space/team or user pair for direct
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_private BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_message TIMESTAMPTZ,
    UNIQUE(type, target_id)
);

-- Chat Channel Members Table
CREATE TABLE IF NOT EXISTS chat_channel_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES chat_channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    last_read TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(channel_id, user_id)
);

-- Chat Messages Table
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES chat_channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type VARCHAR(50) DEFAULT 'text', -- 'text', 'file', 'system'
    metadata JSONB,
    parent_id UUID REFERENCES chat_messages(id) ON DELETE CASCADE, -- For threads
    is_edited BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Chat Reactions Table
CREATE TABLE IF NOT EXISTS chat_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(message_id, user_id, emoji)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_chat_channels_workspace ON chat_channels(workspace_id);
CREATE INDEX IF NOT EXISTS idx_chat_channels_type_target ON chat_channels(type, target_id);
CREATE INDEX IF NOT EXISTS idx_chat_channel_members_user ON chat_channel_members(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_channel_members_channel ON chat_channel_members(channel_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_channel ON chat_messages(channel_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created ON chat_messages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_parent ON chat_messages(parent_id);
CREATE INDEX IF NOT EXISTS idx_chat_reactions_message ON chat_reactions(message_id);

-- Add notification preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type VARCHAR(100) NOT NULL,
    channel VARCHAR(50) NOT NULL, -- 'push', 'email', 'in_app'
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, notification_type, channel)
);

-- Add index for notification preferences
CREATE INDEX IF NOT EXISTS idx_notification_preferences_user ON notification_preferences(user_id);
