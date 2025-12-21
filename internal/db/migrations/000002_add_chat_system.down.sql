-- ============================================
-- DROP MIGRATION 000002 - Chat System
-- ============================================

DROP TRIGGER IF EXISTS update_chat_channels_updated_at ON chat_channels;
DROP TRIGGER IF EXISTS update_chat_messages_updated_at ON chat_messages;

DROP TABLE IF EXISTS chat_reactions CASCADE;
DROP TABLE IF EXISTS chat_messages CASCADE;
DROP TABLE IF EXISTS chat_channel_members CASCADE;
DROP TABLE IF EXISTS chat_channels CASCADE;