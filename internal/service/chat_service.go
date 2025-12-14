package service

import (
	"context"
	"fmt"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
)

// ============================================
// Chat Service - ClickUp-like Chat
// ============================================

// ChatService handles real-time chat functionality
type ChatService interface {
	// Channel management
	CreateChannel(ctx context.Context, name, channelType, targetID, workspaceID, creatorID string, isPrivate bool) (*repository.ChatChannel, error)
	GetChannel(ctx context.Context, id string) (*repository.ChatChannel, error)
	GetChannelByTarget(ctx context.Context, targetType, targetID string) (*repository.ChatChannel, error)
	GetOrCreateChannelForTarget(ctx context.Context, name, targetType, targetID, workspaceID, userID string) (*repository.ChatChannel, error)
	ListChannels(ctx context.Context, userID string) ([]*repository.ChatChannel, error)
	ListWorkspaceChannels(ctx context.Context, workspaceID string) ([]*repository.ChatChannel, error)
	DeleteChannel(ctx context.Context, id, userID string) error

	// Direct messages
	CreateDirectChannel(ctx context.Context, user1ID, user2ID, workspaceID string) (*repository.ChatChannel, error)
	GetDirectChannel(ctx context.Context, user1ID, user2ID string) (*repository.ChatChannel, error)

	// Membership
	JoinChannel(ctx context.Context, channelID, userID string) error
	LeaveChannel(ctx context.Context, channelID, userID string) error
	GetChannelMembers(ctx context.Context, channelID string) ([]*repository.ChatChannelMember, error)
	MarkChannelAsRead(ctx context.Context, channelID, userID string) error

	// Messages
	SendMessage(ctx context.Context, channelID, userID, content, messageType string, parentID *string) (*repository.ChatMessage, error)
	GetMessages(ctx context.Context, channelID string, limit, offset int) ([]*repository.ChatMessage, error)
	GetThreadMessages(ctx context.Context, parentID string) ([]*repository.ChatMessage, error)
	EditMessage(ctx context.Context, messageID, userID, content string) (*repository.ChatMessage, error)
	DeleteMessage(ctx context.Context, messageID, userID string) error

	// Reactions
	AddReaction(ctx context.Context, messageID, userID, emoji string) error
	RemoveReaction(ctx context.Context, messageID, userID, emoji string) error
	GetReactions(ctx context.Context, messageID string) ([]*repository.ChatReaction, error)

	// Unread counts
	GetUnreadCount(ctx context.Context, channelID, userID string) (int, error)
	GetAllUnreadCounts(ctx context.Context, userID string) (map[string]int, error)
}

type chatService struct {
	chatRepo    repository.ChatRepository
	userRepo    repository.UserRepository
	notifSvc    *notification.Service
	broadcaster *socket.Broadcaster
}

// NewChatService creates a new chat service
func NewChatService(
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	notifSvc *notification.Service,
	broadcaster *socket.Broadcaster,
) ChatService {
	return &chatService{
		chatRepo:    chatRepo,
		userRepo:    userRepo,
		notifSvc:    notifSvc,
		broadcaster: broadcaster,
	}
}

// ============================================
// Channel Management
// ============================================

func (s *chatService) CreateChannel(ctx context.Context, name, channelType, targetID, workspaceID, creatorID string, isPrivate bool) (*repository.ChatChannel, error) {
	channel := &repository.ChatChannel{
		Name:        name,
		Type:        channelType,
		TargetID:    targetID,
		WorkspaceID: workspaceID,
		CreatedBy:   creatorID,
		IsPrivate:   isPrivate,
	}

	if err := s.chatRepo.CreateChannel(ctx, channel); err != nil {
		return nil, err
	}

	// Add creator as member
	member := &repository.ChatChannelMember{
		ChannelID: channel.ID,
		UserID:    creatorID,
	}
	s.chatRepo.AddMember(ctx, member)

	// Broadcast channel creation
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(workspaceID, socket.MessageType("chat_channel_created"), map[string]interface{}{
			"channel": channel,
		}, "")
	}

	return channel, nil
}

func (s *chatService) GetChannel(ctx context.Context, id string) (*repository.ChatChannel, error) {
	return s.chatRepo.GetChannelByID(ctx, id)
}

func (s *chatService) GetChannelByTarget(ctx context.Context, targetType, targetID string) (*repository.ChatChannel, error) {
	return s.chatRepo.GetChannelByTarget(ctx, targetType, targetID)
}

func (s *chatService) GetOrCreateChannelForTarget(ctx context.Context, name, targetType, targetID, workspaceID, userID string) (*repository.ChatChannel, error) {
	// Try to get existing channel
	channel, err := s.chatRepo.GetChannelByTarget(ctx, targetType, targetID)
	if err == nil && channel != nil {
		return channel, nil
	}

	// Create new channel
	return s.CreateChannel(ctx, name, targetType, targetID, workspaceID, userID, false)
}

func (s *chatService) ListChannels(ctx context.Context, userID string) ([]*repository.ChatChannel, error) {
	return s.chatRepo.ListChannelsByUser(ctx, userID)
}

func (s *chatService) ListWorkspaceChannels(ctx context.Context, workspaceID string) ([]*repository.ChatChannel, error) {
	return s.chatRepo.ListChannelsByWorkspace(ctx, workspaceID)
}

func (s *chatService) DeleteChannel(ctx context.Context, id, userID string) error {
	channel, err := s.chatRepo.GetChannelByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}

	// Only creator can delete
	if channel.CreatedBy != userID {
		return ErrForbidden
	}

	return s.chatRepo.DeleteChannel(ctx, id)
}

// ============================================
// Direct Messages
// ============================================

func (s *chatService) CreateDirectChannel(ctx context.Context, user1ID, user2ID, workspaceID string) (*repository.ChatChannel, error) {
	// Check if direct channel already exists
	existing, err := s.GetDirectChannel(ctx, user1ID, user2ID)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Get user names for channel name
	user1, _ := s.userRepo.FindByID(ctx, user1ID)
	user2, _ := s.userRepo.FindByID(ctx, user2ID)

	name := "Direct Message"
	if user1 != nil && user2 != nil {
		name = fmt.Sprintf("%s & %s", user1.Name, user2.Name)
	}

	// Create unique target ID for direct messages
	targetID := user1ID + "_" + user2ID
	if user1ID > user2ID {
		targetID = user2ID + "_" + user1ID
	}

	channel := &repository.ChatChannel{
		Name:        name,
		Type:        "direct",
		TargetID:    targetID,
		WorkspaceID: workspaceID,
		CreatedBy:   user1ID,
		IsPrivate:   true,
	}

	if err := s.chatRepo.CreateChannel(ctx, channel); err != nil {
		return nil, err
	}

	// Add both users as members
	s.chatRepo.AddMember(ctx, &repository.ChatChannelMember{ChannelID: channel.ID, UserID: user1ID})
	s.chatRepo.AddMember(ctx, &repository.ChatChannelMember{ChannelID: channel.ID, UserID: user2ID})

	return channel, nil
}

func (s *chatService) GetDirectChannel(ctx context.Context, user1ID, user2ID string) (*repository.ChatChannel, error) {
	targetID := user1ID + "_" + user2ID
	channel, err := s.chatRepo.GetChannelByTarget(ctx, "direct", targetID)
	if err == nil && channel != nil {
		return channel, nil
	}

	// Try reverse order
	targetID = user2ID + "_" + user1ID
	return s.chatRepo.GetChannelByTarget(ctx, "direct", targetID)
}

// ============================================
// Membership
// ============================================

func (s *chatService) JoinChannel(ctx context.Context, channelID, userID string) error {
	member := &repository.ChatChannelMember{
		ChannelID: channelID,
		UserID:    userID,
	}
	return s.chatRepo.AddMember(ctx, member)
}

func (s *chatService) LeaveChannel(ctx context.Context, channelID, userID string) error {
	return s.chatRepo.RemoveMember(ctx, channelID, userID)
}

func (s *chatService) GetChannelMembers(ctx context.Context, channelID string) ([]*repository.ChatChannelMember, error) {
	members, err := s.chatRepo.GetMembers(ctx, channelID)
	if err != nil {
		return nil, err
	}

	// Populate user info
	for _, m := range members {
		m.User, _ = s.userRepo.FindByID(ctx, m.UserID)
	}

	return members, nil
}

func (s *chatService) MarkChannelAsRead(ctx context.Context, channelID, userID string) error {
	return s.chatRepo.UpdateLastRead(ctx, channelID, userID)
}

// ============================================
// Messages
// ============================================

func (s *chatService) SendMessage(ctx context.Context, channelID, userID, content, messageType string, parentID *string) (*repository.ChatMessage, error) {
	if messageType == "" {
		messageType = "text"
	}

	message := &repository.ChatMessage{
		ChannelID:   channelID,
		UserID:      userID,
		Content:     content,
		MessageType: messageType,
		ParentID:    parentID,
	}

	if err := s.chatRepo.CreateMessage(ctx, message); err != nil {
		return nil, err
	}

	// Get user for response
	message.User, _ = s.userRepo.FindByID(ctx, userID)

	// Get channel for broadcasting
	channel, _ := s.chatRepo.GetChannelByID(ctx, channelID)

	// Broadcast message to channel members
	if s.broadcaster != nil && channel != nil {
		room := fmt.Sprintf("chat:%s", channelID)
		s.broadcaster.SendToUsers([]string{}, socket.MessageType("chat_message"), map[string]interface{}{
			"channelId": channelID,
			"message":   message,
		})

		// Also broadcast to the channel room
		s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_message"), map[string]interface{}{
			"channelId": channelID,
			"message":   message,
		}, userID)
	}

	// Send notifications to other channel members
	if s.notifSvc != nil {
		members, _ := s.chatRepo.GetMembers(ctx, channelID)
		var memberIDs []string
		for _, m := range members {
			if m.UserID != userID {
				memberIDs = append(memberIDs, m.UserID)
			}
		}

		if len(memberIDs) > 0 {
			user, _ := s.userRepo.FindByID(ctx, userID)
			userName := "Someone"
			if user != nil {
				userName = user.Name
			}

			channelName := "chat"
			if channel != nil {
				channelName = channel.Name
			}

			s.notifSvc.SendBatchNotifications(ctx, memberIDs, userID, "CHAT_MESSAGE", "New Message",
				fmt.Sprintf("%s sent a message in %s", userName, channelName),
				map[string]interface{}{
					"channelId": channelID,
					"messageId": message.ID,
					"action":    "view_chat",
				})
		}
	}

	return message, nil
}

func (s *chatService) GetMessages(ctx context.Context, channelID string, limit, offset int) ([]*repository.ChatMessage, error) {
	messages, err := s.chatRepo.GetMessages(ctx, channelID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Populate user info
	for _, m := range messages {
		m.User, _ = s.userRepo.FindByID(ctx, m.UserID)
	}

	return messages, nil
}

func (s *chatService) GetThreadMessages(ctx context.Context, parentID string) ([]*repository.ChatMessage, error) {
	messages, err := s.chatRepo.GetThreadMessages(ctx, parentID)
	if err != nil {
		return nil, err
	}

	// Populate user info
	for _, m := range messages {
		m.User, _ = s.userRepo.FindByID(ctx, m.UserID)
	}

	return messages, nil
}

func (s *chatService) EditMessage(ctx context.Context, messageID, userID, content string) (*repository.ChatMessage, error) {
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, ErrNotFound
	}

	if message.UserID != userID {
		return nil, ErrForbidden
	}

	message.Content = content
	if err := s.chatRepo.UpdateMessage(ctx, message); err != nil {
		return nil, err
	}

	message.User, _ = s.userRepo.FindByID(ctx, userID)

	// Broadcast update
	if s.broadcaster != nil {
		channel, _ := s.chatRepo.GetChannelByID(ctx, message.ChannelID)
		if channel != nil {
			s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_message_updated"), map[string]interface{}{
				"channelId": message.ChannelID,
				"message":   message,
			}, userID)
		}
	}

	return message, nil
}

func (s *chatService) DeleteMessage(ctx context.Context, messageID, userID string) error {
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return ErrNotFound
	}

	if message.UserID != userID {
		return ErrForbidden
	}

	channelID := message.ChannelID

	if err := s.chatRepo.DeleteMessage(ctx, messageID); err != nil {
		return err
	}

	// Broadcast deletion
	if s.broadcaster != nil {
		channel, _ := s.chatRepo.GetChannelByID(ctx, channelID)
		if channel != nil {
			s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_message_deleted"), map[string]interface{}{
				"channelId": channelID,
				"messageId": messageID,
			}, userID)
		}
	}

	return nil
}

// ============================================
// Reactions
// ============================================

func (s *chatService) AddReaction(ctx context.Context, messageID, userID, emoji string) error {
	reaction := &repository.ChatReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}
	return s.chatRepo.AddReaction(ctx, reaction)
}

func (s *chatService) RemoveReaction(ctx context.Context, messageID, userID, emoji string) error {
	return s.chatRepo.RemoveReaction(ctx, messageID, userID, emoji)
}

func (s *chatService) GetReactions(ctx context.Context, messageID string) ([]*repository.ChatReaction, error) {
	return s.chatRepo.GetReactions(ctx, messageID)
}

// ============================================
// Unread Counts
// ============================================

func (s *chatService) GetUnreadCount(ctx context.Context, channelID, userID string) (int, error) {
	return s.chatRepo.GetUnreadCount(ctx, channelID, userID)
}

func (s *chatService) GetAllUnreadCounts(ctx context.Context, userID string) (map[string]int, error) {
	channels, err := s.chatRepo.ListChannelsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, channel := range channels {
		count, err := s.chatRepo.GetUnreadCount(ctx, channel.ID, userID)
		if err == nil && count > 0 {
			counts[channel.ID] = count
		}
	}

	return counts, nil
}
