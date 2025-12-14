package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================
// Chat Models
// ============================================

// ChatChannel represents a chat channel (like ClickUp)
type ChatChannel struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"` // "project", "space", "team", "direct"
	TargetID    string     `json:"targetId"`
	WorkspaceID string     `json:"workspaceId"`
	CreatedBy   string     `json:"createdBy"`
	IsPrivate   bool       `json:"isPrivate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	LastMessage *time.Time `json:"lastMessage,omitempty"`
	// Computed fields (not stored in DB)
	OtherUser   *User `json:"otherUser,omitempty"`   // For direct messages
	MemberCount int   `json:"memberCount,omitempty"` // Number of members
}

// ChatMessage represents a message in a channel
type ChatMessage struct {
	ID          string                 `json:"id"`
	ChannelID   string                 `json:"channelId"`
	UserID      string                 `json:"userId"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"messageType"` // "text", "file", "system"
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ParentID    *string                `json:"parentId,omitempty"` // For threads
	IsEdited    bool                   `json:"isEdited"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	User        *User                  `json:"user,omitempty"`
	Reactions   []*ChatReaction        `json:"reactions,omitempty"`
	ReplyCount  int                    `json:"replyCount,omitempty"`
}

// ChatChannelMember represents channel membership
type ChatChannelMember struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channelId"`
	UserID    string    `json:"userId"`
	JoinedAt  time.Time `json:"joinedAt"`
	LastRead  time.Time `json:"lastRead"`
	User      *User     `json:"user,omitempty"`
}

// ChatReaction represents a message reaction
type ChatReaction struct {
	ID        string    `json:"id"`
	MessageID string    `json:"messageId"`
	UserID    string    `json:"userId"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"createdAt"`
	User      *User     `json:"user,omitempty"`
}

// ============================================
// Chat Repository Interface
// ============================================

// ChatRepository defines chat operations
type ChatRepository interface {
	// Channel operations
	CreateChannel(ctx context.Context, channel *ChatChannel) error
	GetChannelByID(ctx context.Context, id string) (*ChatChannel, error)
	GetChannelByTarget(ctx context.Context, targetType, targetID string) (*ChatChannel, error)
	ListChannelsByWorkspace(ctx context.Context, workspaceID string) ([]*ChatChannel, error)
	ListChannelsByUser(ctx context.Context, userID string) ([]*ChatChannel, error)
	UpdateChannel(ctx context.Context, channel *ChatChannel) error
	DeleteChannel(ctx context.Context, id string) error

	// Member operations
	AddMember(ctx context.Context, member *ChatChannelMember) error
	RemoveMember(ctx context.Context, channelID, userID string) error
	GetMembers(ctx context.Context, channelID string) ([]*ChatChannelMember, error)
	GetMemberCount(ctx context.Context, channelID string) (int, error)
	IsMember(ctx context.Context, channelID, userID string) (bool, error)
	UpdateLastRead(ctx context.Context, channelID, userID string) error

	// Message operations
	CreateMessage(ctx context.Context, message *ChatMessage) error
	GetMessageByID(ctx context.Context, id string) (*ChatMessage, error)
	GetMessages(ctx context.Context, channelID string, limit, offset int) ([]*ChatMessage, error)
	GetThreadMessages(ctx context.Context, parentID string) ([]*ChatMessage, error)
	UpdateMessage(ctx context.Context, message *ChatMessage) error
	DeleteMessage(ctx context.Context, id string) error

	// Reaction operations
	AddReaction(ctx context.Context, reaction *ChatReaction) error
	RemoveReaction(ctx context.Context, messageID, userID, emoji string) error
	GetReactions(ctx context.Context, messageID string) ([]*ChatReaction, error)

	// Unread count
	GetUnreadCount(ctx context.Context, channelID, userID string) (int, error)
}

// ============================================
// PostgreSQL Implementation
// ============================================

type chatRepository struct {
	pool *pgxpool.Pool
}

// NewChatRepository creates a new chat repository
func NewChatRepository(pool *pgxpool.Pool) ChatRepository {
	return &chatRepository{pool: pool}
}

// ============================================
// Channel Operations
// ============================================

func (r *chatRepository) CreateChannel(ctx context.Context, channel *ChatChannel) error {
	channel.ID = uuid.New().String()
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO chat_channels (id, name, type, target_id, workspace_id, created_by, is_private, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, channel.ID, channel.Name, channel.Type, channel.TargetID, channel.WorkspaceID, channel.CreatedBy, channel.IsPrivate, channel.CreatedAt, channel.UpdatedAt)

	return err
}

func (r *chatRepository) GetChannelByID(ctx context.Context, id string) (*ChatChannel, error) {
	channel := &ChatChannel{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, type, target_id, workspace_id, created_by, is_private, created_at, updated_at, last_message
		FROM chat_channels WHERE id = $1
	`, id).Scan(&channel.ID, &channel.Name, &channel.Type, &channel.TargetID, &channel.WorkspaceID, &channel.CreatedBy, &channel.IsPrivate, &channel.CreatedAt, &channel.UpdatedAt, &channel.LastMessage)

	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (r *chatRepository) GetChannelByTarget(ctx context.Context, targetType, targetID string) (*ChatChannel, error) {
	channel := &ChatChannel{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, type, target_id, workspace_id, created_by, is_private, created_at, updated_at, last_message
		FROM chat_channels WHERE type = $1 AND target_id = $2
	`, targetType, targetID).Scan(&channel.ID, &channel.Name, &channel.Type, &channel.TargetID, &channel.WorkspaceID, &channel.CreatedBy, &channel.IsPrivate, &channel.CreatedAt, &channel.UpdatedAt, &channel.LastMessage)

	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (r *chatRepository) ListChannelsByWorkspace(ctx context.Context, workspaceID string) ([]*ChatChannel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, type, target_id, workspace_id, created_by, is_private, created_at, updated_at, last_message
		FROM chat_channels WHERE workspace_id = $1
		ORDER BY COALESCE(last_message, created_at) DESC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*ChatChannel
	for rows.Next() {
		channel := &ChatChannel{}
		if err := rows.Scan(&channel.ID, &channel.Name, &channel.Type, &channel.TargetID, &channel.WorkspaceID, &channel.CreatedBy, &channel.IsPrivate, &channel.CreatedAt, &channel.UpdatedAt, &channel.LastMessage); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

func (r *chatRepository) ListChannelsByUser(ctx context.Context, userID string) ([]*ChatChannel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name, c.type, c.target_id, c.workspace_id, c.created_by, c.is_private, c.created_at, c.updated_at, c.last_message
		FROM chat_channels c
		INNER JOIN chat_channel_members m ON c.id = m.channel_id
		WHERE m.user_id = $1
		ORDER BY COALESCE(c.last_message, c.created_at) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*ChatChannel
	for rows.Next() {
		channel := &ChatChannel{}
		if err := rows.Scan(&channel.ID, &channel.Name, &channel.Type, &channel.TargetID, &channel.WorkspaceID, &channel.CreatedBy, &channel.IsPrivate, &channel.CreatedAt, &channel.UpdatedAt, &channel.LastMessage); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

func (r *chatRepository) UpdateChannel(ctx context.Context, channel *ChatChannel) error {
	channel.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE chat_channels SET name = $2, type = $3, is_private = $4, updated_at = $5
		WHERE id = $1
	`, channel.ID, channel.Name, channel.Type, channel.IsPrivate, channel.UpdatedAt)
	return err
}

func (r *chatRepository) DeleteChannel(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM chat_channels WHERE id = $1`, id)
	return err
}

// ============================================
// Member Operations
// ============================================

func (r *chatRepository) AddMember(ctx context.Context, member *ChatChannelMember) error {
	member.ID = uuid.New().String()
	member.JoinedAt = time.Now()
	member.LastRead = time.Now()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO chat_channel_members (id, channel_id, user_id, joined_at, last_read)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (channel_id, user_id) DO NOTHING
	`, member.ID, member.ChannelID, member.UserID, member.JoinedAt, member.LastRead)
	return err
}

func (r *chatRepository) RemoveMember(ctx context.Context, channelID, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM chat_channel_members WHERE channel_id = $1 AND user_id = $2`, channelID, userID)
	return err
}

func (r *chatRepository) GetMembers(ctx context.Context, channelID string) ([]*ChatChannelMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			m.id, m.channel_id, m.user_id, m.joined_at, m.last_read,
			u.id, u.name, u.email, u.avatar
		FROM chat_channel_members m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.channel_id = $1
		ORDER BY m.joined_at ASC
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*ChatChannelMember
	for rows.Next() {
		member := &ChatChannelMember{}
		var userID, userName, userEmail, userAvatar *string

		if err := rows.Scan(
			&member.ID, &member.ChannelID, &member.UserID, &member.JoinedAt, &member.LastRead,
			&userID, &userName, &userEmail, &userAvatar,
		); err != nil {
			return nil, err
		}

		// Populate user if exists
		if userID != nil && userName != nil {
			member.User = &User{
				ID:   *userID,
				Name: *userName,
			}
			if userEmail != nil {
				member.User.Email = *userEmail
			}
			if userAvatar != nil {
				member.User.Avatar = userAvatar
			}
		}

		members = append(members, member)
	}

	return members, nil
}

func (r *chatRepository) GetMemberCount(ctx context.Context, channelID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM chat_channel_members WHERE channel_id = $1
	`, channelID).Scan(&count)
	return count, err
}

func (r *chatRepository) IsMember(ctx context.Context, channelID, userID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM chat_channel_members WHERE channel_id = $1 AND user_id = $2
	`, channelID, userID).Scan(&count)
	return count > 0, err
}

func (r *chatRepository) UpdateLastRead(ctx context.Context, channelID, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE chat_channel_members SET last_read = NOW() WHERE channel_id = $1 AND user_id = $2
	`, channelID, userID)
	return err
}

// ============================================
// Message Operations
// ============================================

func (r *chatRepository) CreateMessage(ctx context.Context, message *ChatMessage) error {
	message.ID = uuid.New().String()
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO chat_messages (id, channel_id, user_id, content, message_type, metadata, parent_id, is_edited, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, message.ID, message.ChannelID, message.UserID, message.Content, message.MessageType, message.Metadata, message.ParentID, message.IsEdited, message.CreatedAt, message.UpdatedAt)

	if err == nil {
		// Update channel last_message timestamp
		r.pool.Exec(ctx, `UPDATE chat_channels SET last_message = NOW() WHERE id = $1`, message.ChannelID)
	}

	return err
}

func (r *chatRepository) GetMessageByID(ctx context.Context, id string) (*ChatMessage, error) {
	message := &ChatMessage{}
	var userID, userName, userEmail, userAvatar *string

	err := r.pool.QueryRow(ctx, `
		SELECT 
			m.id, m.channel_id, m.user_id, m.content, m.message_type, 
			m.metadata, m.parent_id, m.is_edited, m.created_at, m.updated_at,
			u.id, u.name, u.email, u.avatar
		FROM chat_messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.id = $1
	`, id).Scan(
		&message.ID, &message.ChannelID, &message.UserID, &message.Content,
		&message.MessageType, &message.Metadata, &message.ParentID,
		&message.IsEdited, &message.CreatedAt, &message.UpdatedAt,
		&userID, &userName, &userEmail, &userAvatar,
	)

	if err != nil {
		return nil, err
	}

	// Populate user if exists
	if userID != nil && userName != nil {
		message.User = &User{
			ID:   *userID,
			Name: *userName,
		}
		if userEmail != nil {
			message.User.Email = *userEmail
		}
		if userAvatar != nil {
			message.User.Avatar = userAvatar

		}
	}

	return message, nil
}

func (r *chatRepository) GetMessages(ctx context.Context, channelID string, limit, offset int) ([]*ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.pool.Query(ctx, `
		SELECT 
			m.id, m.channel_id, m.user_id, m.content, m.message_type, 
			m.metadata, m.parent_id, m.is_edited, m.created_at, m.updated_at,
			u.id, u.name, u.email, u.avatar,
			(SELECT COUNT(*) FROM chat_messages replies WHERE replies.parent_id = m.id) as reply_count
		FROM chat_messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.channel_id = $1 AND m.parent_id IS NULL
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`, channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*ChatMessage
	for rows.Next() {
		message := &ChatMessage{}
		var userID, userName, userEmail, userAvatar *string

		if err := rows.Scan(
			&message.ID, &message.ChannelID, &message.UserID, &message.Content,
			&message.MessageType, &message.Metadata, &message.ParentID,
			&message.IsEdited, &message.CreatedAt, &message.UpdatedAt,
			&userID, &userName, &userEmail, &userAvatar,
			&message.ReplyCount,
		); err != nil {
			return nil, err
		}

		// Populate user if exists
		if userID != nil && userName != nil {
			message.User = &User{
				ID:   *userID,
				Name: *userName,
			}
			if userEmail != nil {
				message.User.Email = *userEmail
			}
			if userAvatar != nil {
				message.User.Avatar = userAvatar
			}
		}

		messages = append(messages, message)
	}

	// Load reactions for each message
	for _, msg := range messages {
		reactions, err := r.GetReactions(ctx, msg.ID)
		if err == nil {
			msg.Reactions = reactions
		}
	}

	return messages, nil
}

func (r *chatRepository) GetThreadMessages(ctx context.Context, parentID string) ([]*ChatMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			m.id, m.channel_id, m.user_id, m.content, m.message_type, 
			m.metadata, m.parent_id, m.is_edited, m.created_at, m.updated_at,
			u.id, u.name, u.email, u.avatar
		FROM chat_messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.parent_id = $1
		ORDER BY m.created_at ASC
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*ChatMessage
	for rows.Next() {
		message := &ChatMessage{}
		var userID, userName, userEmail, userAvatar *string

		if err := rows.Scan(
			&message.ID, &message.ChannelID, &message.UserID, &message.Content,
			&message.MessageType, &message.Metadata, &message.ParentID,
			&message.IsEdited, &message.CreatedAt, &message.UpdatedAt,
			&userID, &userName, &userEmail, &userAvatar,
		); err != nil {
			return nil, err
		}

		// Populate user if exists
		if userID != nil && userName != nil {
			message.User = &User{
				ID:   *userID,
				Name: *userName,
			}
			if userEmail != nil {
				message.User.Email = *userEmail
			}
			if userAvatar != nil {
				message.User.Avatar = userAvatar
			}
		}

		messages = append(messages, message)
	}

	// Load reactions for each message
	for _, msg := range messages {
		reactions, err := r.GetReactions(ctx, msg.ID)
		if err == nil {
			msg.Reactions = reactions
		}
	}

	return messages, nil
}

func (r *chatRepository) UpdateMessage(ctx context.Context, message *ChatMessage) error {
	message.UpdatedAt = time.Now()
	message.IsEdited = true
	_, err := r.pool.Exec(ctx, `
		UPDATE chat_messages SET content = $2, is_edited = $3, updated_at = $4 WHERE id = $1
	`, message.ID, message.Content, message.IsEdited, message.UpdatedAt)
	return err
}

func (r *chatRepository) DeleteMessage(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM chat_messages WHERE id = $1`, id)
	return err
}

// ============================================
// Reaction Operations
// ============================================

func (r *chatRepository) AddReaction(ctx context.Context, reaction *ChatReaction) error {
	reaction.ID = uuid.New().String()
	reaction.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO chat_reactions (id, message_id, user_id, emoji, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
	`, reaction.ID, reaction.MessageID, reaction.UserID, reaction.Emoji, reaction.CreatedAt)
	return err
}

func (r *chatRepository) RemoveReaction(ctx context.Context, messageID, userID, emoji string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM chat_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3
	`, messageID, userID, emoji)
	return err
}

func (r *chatRepository) GetReactions(ctx context.Context, messageID string) ([]*ChatReaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.message_id, r.user_id, r.emoji, r.created_at,
			   u.id, u.name, u.email, u.avatar
		FROM chat_reactions r
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.message_id = $1
		ORDER BY r.created_at ASC
	`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []*ChatReaction
	for rows.Next() {
		reaction := &ChatReaction{}
		var userID, userName, userEmail, userAvatar *string

		if err := rows.Scan(
			&reaction.ID, &reaction.MessageID, &reaction.UserID, &reaction.Emoji, &reaction.CreatedAt,
			&userID, &userName, &userEmail, &userAvatar,
		); err != nil {
			return nil, err
		}

		// Populate user if exists
		if userID != nil && userName != nil {
			reaction.User = &User{
				ID:   *userID,
				Name: *userName,
			}
			if userEmail != nil {
				reaction.User.Email = *userEmail
			}
			if userAvatar != nil {
				reaction.User.Avatar = userAvatar
			}
		}

		reactions = append(reactions, reaction)
	}

	return reactions, nil
}

// ============================================
// Unread Count
// ============================================

func (r *chatRepository) GetUnreadCount(ctx context.Context, channelID, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM chat_messages m
		INNER JOIN chat_channel_members cm ON m.channel_id = cm.channel_id
		WHERE m.channel_id = $1 AND cm.user_id = $2 AND m.created_at > cm.last_read AND m.user_id != $2
	`, channelID, userID).Scan(&count)
	return count, err
}
