-- ============================================
-- ORA SCRUM - CHAT SYSTEM (Migration 000002)
-- ============================================

-- ============================================
-- CHAT CHANNELS TABLE
-- ============================================
CREATE TABLE chat_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    target_id VARCHAR(255) NOT NULL,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_private BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_message TIMESTAMPTZ,
    UNIQUE(workspace_id, type, target_id)
);
CREATE INDEX idx_chat_channels_workspace ON chat_channels(workspace_id);
CREATE INDEX idx_chat_channels_type_target ON chat_channels(type, target_id);
CREATE INDEX idx_chat_channels_created_by ON chat_channels(created_by);

-- ============================================
-- CHAT CHANNEL MEMBERS TABLE
-- ============================================
CREATE TABLE chat_channel_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id UUID NOT NULL REFERENCES chat_channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    last_read TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(channel_id, user_id)
);
CREATE INDEX idx_chat_channel_members_channel ON chat_channel_members(channel_id);
CREATE INDEX idx_chat_channel_members_user ON chat_channel_members(user_id);

-- ============================================
-- CHAT MESSAGES TABLE
-- ============================================
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id UUID NOT NULL REFERENCES chat_channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type VARCHAR(50) DEFAULT 'text',
    metadata JSONB,
    parent_id UUID REFERENCES chat_messages(id) ON DELETE CASCADE,
    is_edited BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_chat_messages_channel_created ON chat_messages(channel_id, created_at DESC);
CREATE INDEX idx_chat_messages_parent ON chat_messages(parent_id);
CREATE INDEX idx_chat_messages_user ON chat_messages(user_id);

-- ============================================
-- CHAT REACTIONS TABLE
-- ============================================
CREATE TABLE chat_reactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(message_id, user_id, emoji)
);
CREATE INDEX idx_chat_reactions_message ON chat_reactions(message_id);
CREATE INDEX idx_chat_reactions_user ON chat_reactions(user_id);

-- ============================================
-- TRIGGERS
-- ============================================

CREATE TRIGGER update_chat_channels_updated_at 
    BEFORE UPDATE ON chat_channels 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chat_messages_updated_at 
    BEFORE UPDATE ON chat_messages 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();