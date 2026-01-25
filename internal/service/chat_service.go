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
	GetChannelForUser(ctx context.Context, id, userID string) (*repository.ChatChannel, error)
	GetChannelByTarget(ctx context.Context, targetType, targetID string) (*repository.ChatChannel, error)
	GetOrCreateChannelForTarget(ctx context.Context, name, targetType, targetID, workspaceID, userID string) (*repository.ChatChannel, error)
	ListChannels(ctx context.Context, userID string) ([]*repository.ChatChannel, error)
	ListWorkspaceChannels(ctx context.Context, workspaceID string) ([]*repository.ChatChannel, error)
	UpdateChannel(ctx context.Context, id, name string, isPrivate bool) (*repository.ChatChannel, error)
	DeleteChannel(ctx context.Context, id, userID string) error

	// Direct messages
	CreateDirectChannel(ctx context.Context, user1ID, user2ID, workspaceID string) (*repository.ChatChannel, error)
	GetDirectChannel(ctx context.Context, user1ID, user2ID string) (*repository.ChatChannel, error)

	// Membership
	JoinChannel(ctx context.Context, channelID, userID string) error
	AddMemberToChannel(ctx context.Context, channelID, userID, addedByID string) error
	LeaveChannel(ctx context.Context, channelID, userID string) error
	RemoveMemberFromChannel(ctx context.Context, channelID, userID, removedByID string) error
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
// Helper Methods
// ============================================

// populateDirectChannelUser populates the OtherUser field for direct message channels
func (s *chatService) populateDirectChannelUser(ctx context.Context, channel *repository.ChatChannel, currentUserID string) {
	if channel.Type != "direct" {
		return
	}

	// Get channel members
	members, err := s.chatRepo.GetMembers(ctx, channel.ID)
	if err != nil {
		return
	}

	// Find the other user
	for _, member := range members {
		if member.UserID != currentUserID {
			if member.User != nil {
				channel.OtherUser = member.User
			} else {
				// Fallback: fetch user directly
				user, err := s.userRepo.FindByID(ctx, member.UserID)
				if err == nil && user != nil {
					channel.OtherUser = user
				}
			}
			break
		}
	}
}

// populateMemberCount populates the MemberCount field
func (s *chatService) populateMemberCount(ctx context.Context, channel *repository.ChatChannel) {
	count, err := s.chatRepo.GetMemberCount(ctx, channel.ID)
	if err == nil {
		channel.MemberCount = count
	}
}

// convertDMToGroup converts a direct message to a group chat when adding a third person
func (s *chatService) convertDMToGroup(ctx context.Context, channel *repository.ChatChannel) error {
	if channel.Type != "direct" {
		return nil
	}

	// Get current members to build new name
	members, err := s.chatRepo.GetMembers(ctx, channel.ID)
	if err != nil {
		return err
	}

	// Build group name from member names
	var names []string
	for _, m := range members {
		if m.User != nil {
			names = append(names, m.User.Name)
		}
	}

	newName := "Group Chat"
	if len(names) > 0 {
		newName = fmt.Sprintf("%s", names[0])
		for i := 1; i < len(names) && i < 3; i++ {
			newName += fmt.Sprintf(", %s", names[i])
		}
		if len(names) > 3 {
			newName += fmt.Sprintf(" +%d", len(names)-3)
		}
	}

	// Update channel type to "group"
	channel.Name = newName
	channel.Type = "group"
	return s.chatRepo.UpdateChannel(ctx, channel)
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
	channel.MemberCount = 1

	// Broadcast channel creation
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(workspaceID, socket.MessageType("chat_channel_created"), map[string]interface{}{
			"channel": channel,
		}, "")
	}

	return channel, nil
}


// ✅ NEW: CreateChannelWithMembers creates a channel and adds initial members
func (s *chatService) CreateChannelWithMembers(ctx context.Context, name, channelType, targetID, workspaceID, creatorID string, isPrivate bool, memberIDs []string) (*repository.ChatChannel, error) {
	channel, err := s.CreateChannel(ctx, name, channelType, targetID, workspaceID, creatorID, isPrivate)
	if err != nil {
		return nil, err
	}

	creator, _ := s.userRepo.FindByID(ctx, creatorID)
	creatorName := getNameOrUnknown(creator)

	// Add additional members
	for _, memberID := range memberIDs {
		if memberID == creatorID {
			continue
		}

		member := &repository.ChatChannelMember{
			ChannelID: channel.ID,
			UserID:    memberID,
		}
		if err := s.chatRepo.AddMember(ctx, member); err != nil {
			continue
		}

		// Notify added member
		if s.notifSvc != nil {
			s.notifSvc.SendChatAddedToChannel(ctx, memberID, channel.ID, name, creatorName, workspaceID, false)
		}
	}

	// Update member count
	count, _ := s.chatRepo.GetMemberCount(ctx, channel.ID)
	channel.MemberCount = count

	return channel, nil
}

func (s *chatService) GetChannel(ctx context.Context, id string) (*repository.ChatChannel, error) {
	channel, err := s.chatRepo.GetChannelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.populateMemberCount(ctx, channel)
	return channel, nil
}

// GetChannelForUser returns a channel with OtherUser populated for DMs
func (s *chatService) GetChannelForUser(ctx context.Context, id, userID string) (*repository.ChatChannel, error) {
	channel, err := s.chatRepo.GetChannelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Populate OtherUser for direct messages
	s.populateDirectChannelUser(ctx, channel, userID)
	s.populateMemberCount(ctx, channel)

	return channel, nil
}

func (s *chatService) GetChannelByTarget(ctx context.Context, targetType, targetID string) (*repository.ChatChannel, error) {
	channel, err := s.chatRepo.GetChannelByTarget(ctx, targetType, targetID)
	if err != nil {
		return nil, err
	}

	s.populateMemberCount(ctx, channel)
	return channel, nil
}

func (s *chatService) GetOrCreateChannelForTarget(ctx context.Context, name, targetType, targetID, workspaceID, userID string) (*repository.ChatChannel, error) {
	// Try to get existing channel
	channel, err := s.chatRepo.GetChannelByTarget(ctx, targetType, targetID)
	if err == nil && channel != nil {
		s.populateMemberCount(ctx, channel)
		return channel, nil
	}

	// Create new channel
	return s.CreateChannel(ctx, name, targetType, targetID, workspaceID, userID, false)
}

func (s *chatService) ListChannels(ctx context.Context, userID string) ([]*repository.ChatChannel, error) {
	channels, err := s.chatRepo.ListChannelsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Populate OtherUser for direct message channels and member counts
	for _, channel := range channels {
		if channel.Type == "direct" {
			s.populateDirectChannelUser(ctx, channel, userID)
		}
		s.populateMemberCount(ctx, channel)
	}

	return channels, nil
}

func (s *chatService) ListWorkspaceChannels(ctx context.Context, workspaceID string) ([]*repository.ChatChannel, error) {
	channels, err := s.chatRepo.ListChannelsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Populate member counts
	for _, channel := range channels {
		s.populateMemberCount(ctx, channel)
	}

	return channels, nil
}

func (s *chatService) UpdateChannel(ctx context.Context, id, name string, isPrivate bool) (*repository.ChatChannel, error) {
	channel, err := s.chatRepo.GetChannelByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	channel.Name = name
	channel.IsPrivate = isPrivate

	if err := s.chatRepo.UpdateChannel(ctx, channel); err != nil {
		return nil, err
	}

	s.populateMemberCount(ctx, channel)

	// Broadcast update
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_channel_updated"), map[string]interface{}{
			"channel": channel,
		}, "")
	}

	return channel, nil
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

	// Don't allow deleting direct messages
	if channel.Type == "direct" {
		return fmt.Errorf("cannot delete direct message channels")
	}

	// Broadcast before deletion
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_channel_deleted"), map[string]interface{}{
			"channelId": id,
		}, "")
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
		s.populateDirectChannelUser(ctx, existing, user1ID)
		s.populateMemberCount(ctx, existing)
		return existing, nil
	}

	// Get user names
	user1, _ := s.userRepo.FindByID(ctx, user1ID)
	user2, _ := s.userRepo.FindByID(ctx, user2ID)

	name := "Direct Message"
	if user1 != nil && user2 != nil {
		name = fmt.Sprintf("%s & %s", user1.Name, user2.Name)
	}

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

	// Populate OtherUser for response
	channel.OtherUser = user2
	channel.MemberCount = 2

	// ✅ NEW: Notify the other user about the new DM
	if s.notifSvc != nil && user1 != nil {
		s.notifSvc.SendChatAddedToChannel(ctx, user2ID, channel.ID, name, user1.Name, workspaceID, true)
	}

	return channel, nil
}

func (s *chatService) GetDirectChannel(ctx context.Context, user1ID, user2ID string) (*repository.ChatChannel, error) {
	// Try user1_user2 format
	targetID := user1ID + "_" + user2ID
	channel, err := s.chatRepo.GetChannelByTarget(ctx, "direct", targetID)
	if err == nil && channel != nil {
		return channel, nil
	}

	// Try user2_user1 format (reverse order)
	targetID = user2ID + "_" + user1ID
	return s.chatRepo.GetChannelByTarget(ctx, "direct", targetID)
}

// ============================================
// Membership
// ============================================

func (s *chatService) JoinChannel(ctx context.Context, channelID, userID string) error {
	// Check if channel exists
	channel, err := s.chatRepo.GetChannelByID(ctx, channelID)
	if err != nil {
		return ErrNotFound
	}

	// Don't allow joining private channels without invitation
	if channel.IsPrivate {
		isMember, _ := s.chatRepo.IsMember(ctx, channelID, userID)
		if !isMember {
			return ErrForbidden
		}
	}

	member := &repository.ChatChannelMember{
		ChannelID: channelID,
		UserID:    userID,
	}
	return s.chatRepo.AddMember(ctx, member)
}

func (s *chatService) AddMemberToChannel(ctx context.Context, channelID, userID, addedByID string) error {
	channel, err := s.chatRepo.GetChannelByID(ctx, channelID)
	if err != nil {
		return ErrNotFound
	}

	isMember, _ := s.chatRepo.IsMember(ctx, channelID, addedByID)
	if !isMember {
		return ErrForbidden
	}

	alreadyMember, _ := s.chatRepo.IsMember(ctx, channelID, userID)
	if alreadyMember {
		return nil
	}

	member := &repository.ChatChannelMember{
		ChannelID: channelID,
		UserID:    userID,
	}
	if err := s.chatRepo.AddMember(ctx, member); err != nil {
		return err
	}

	memberCount, _ := s.chatRepo.GetMemberCount(ctx, channelID)
	if channel.Type == "direct" && memberCount > 2 {
		s.convertDMToGroup(ctx, channel)
	}

	addedUser, _ := s.userRepo.FindByID(ctx, userID)
	adderUser, _ := s.userRepo.FindByID(ctx, addedByID)

	// Broadcast member added
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_member_added"), map[string]interface{}{
			"channelId":   channelID,
			"userId":      userID,
			"addedBy":     addedByID,
			"user":        addedUser,
			"addedByUser": adderUser,
		}, "")
	}

	// System message
	systemMessage := fmt.Sprintf("%s added %s to the conversation",
		getNameOrUnknown(adderUser),
		getNameOrUnknown(addedUser))
	s.sendSystemMessage(ctx, channelID, systemMessage)

	// ✅ NEW: Send notification to the added user
	if s.notifSvc != nil {
    s.notifSvc.SendChatAddedToChannel(
        ctx,
        userID,
        channelID,
        channel.Name,
        getNameOrUnknown(adderUser),  // ✅ This passes name correctly
        channel.WorkspaceID,
        channel.Type == "direct",
    )
}

	return nil
}

func (s *chatService) LeaveChannel(ctx context.Context, channelID, userID string) error {
	return s.chatRepo.RemoveMember(ctx, channelID, userID)
}

func (s *chatService) RemoveMemberFromChannel(ctx context.Context, channelID, userID, removedByID string) error {
	channel, err := s.chatRepo.GetChannelByID(ctx, channelID)
	if err != nil {
		return ErrNotFound
	}

	// Direct messages: don't allow removal
	if channel.Type == "direct" {
		return fmt.Errorf("cannot remove members from direct messages")
	}

	// Check permissions: only creator can remove others, anyone can remove themselves
	if userID != removedByID && channel.CreatedBy != removedByID {
		return ErrForbidden
	}

	// Don't allow creator to remove themselves if others exist
	if userID == channel.CreatedBy {
		count, _ := s.chatRepo.GetMemberCount(ctx, channelID)
		if count > 1 {
			return fmt.Errorf("channel creator cannot leave while other members exist")
		}
	}

	if err := s.chatRepo.RemoveMember(ctx, channelID, userID); err != nil {
		return err
	}

	removedUser, _ := s.userRepo.FindByID(ctx, userID)
	removerUser, _ := s.userRepo.FindByID(ctx, removedByID)

	// Broadcast member removed
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_member_removed"), map[string]interface{}{
			"channelId": channelID,
			"userId":    userID,
			"removedBy": removedByID,
		}, "")
	}

	// System message
	var systemMessage string
	if userID == removedByID {
		systemMessage = fmt.Sprintf("%s left the conversation", getNameOrUnknown(removedUser))
	} else {
		systemMessage = fmt.Sprintf("%s removed %s from the conversation",
			getNameOrUnknown(removerUser),
			getNameOrUnknown(removedUser))

		if s.notifSvc != nil {
    s.notifSvc.SendChatRemovedFromChannel(
        ctx,
        userID,
        channel.Name,
        getNameOrUnknown(removerUser),  // ✅ This passes name correctly
    )
}
	}
	s.sendSystemMessage(ctx, channelID, systemMessage)

	return nil
}


func (s *chatService) GetChannelMembers(ctx context.Context, channelID string) ([]*repository.ChatChannelMember, error) {
	return s.chatRepo.GetMembers(ctx, channelID)
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

	message.User, _ = s.userRepo.FindByID(ctx, userID)
	channel, _ := s.chatRepo.GetChannelByID(ctx, channelID)

	// Broadcast message
	if s.broadcaster != nil && channel != nil {
		s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_message"), map[string]interface{}{
			"channelId": channelID,
			"messageId": message.ID,
			"message":   message,
		}, "")
	}

	// ✅ NEW: Parse and send @mention notifications
	if s.notifSvc != nil && channel != nil && message.User != nil {
		s.notifSvc.ParseChatMentions(
			ctx,
			content,
			userID,
			message.User.Name,
			channelID,
			channel.Name,
			channel.Type == "direct",
		)
	}

	return message, nil
}

// sendSystemMessage sends a system message to the channel
func (s *chatService) sendSystemMessage(ctx context.Context, channelID, content string) {
	message := &repository.ChatMessage{
		ChannelID:   channelID,
		UserID:      "system",
		Content:     content,
		MessageType: "system",
	}
	s.chatRepo.CreateMessage(ctx, message)
}

func (s *chatService) GetMessages(ctx context.Context, channelID string, limit, offset int) ([]*repository.ChatMessage, error) {
	return s.chatRepo.GetMessages(ctx, channelID, limit, offset)
}

func (s *chatService) GetThreadMessages(ctx context.Context, parentID string) ([]*repository.ChatMessage, error) {
	return s.chatRepo.GetThreadMessages(ctx, parentID)
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

	if err := s.chatRepo.AddReaction(ctx, reaction); err != nil {
		return err
	}

	// Broadcast reaction
	if s.broadcaster != nil {
		message, _ := s.chatRepo.GetMessageByID(ctx, messageID)
		if message != nil {
			channel, _ := s.chatRepo.GetChannelByID(ctx, message.ChannelID)
			if channel != nil {
				s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_reaction_added"), map[string]interface{}{
					"channelId": message.ChannelID,
					"messageId": messageID,
					"userId":    userID,
					"emoji":     emoji,
				}, "")
			}
		}
	}

	return nil
}

func (s *chatService) RemoveReaction(ctx context.Context, messageID, userID, emoji string) error {
	if err := s.chatRepo.RemoveReaction(ctx, messageID, userID, emoji); err != nil {
		return err
	}

	// Broadcast reaction removal
	if s.broadcaster != nil {
		message, _ := s.chatRepo.GetMessageByID(ctx, messageID)
		if message != nil {
			channel, _ := s.chatRepo.GetChannelByID(ctx, message.ChannelID)
			if channel != nil {
				s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_reaction_removed"), map[string]interface{}{
					"channelId": message.ChannelID,
					"messageId": messageID,
					"userId":    userID,
					"emoji":     emoji,
				}, "")
			}
		}
	}

	return nil
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

// Helper function
func getNameOrUnknown(user *repository.User) string {
	if user != nil && user.Name != "" {
		return user.Name
	}
	return "Unknown User"
}
