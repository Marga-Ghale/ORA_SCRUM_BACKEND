package service

import (
	"context"
	"fmt"
	"strings"

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

	// Add to ChatService interface:
	ArchiveChannel(ctx context.Context, channelID, userID string) error

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

// Channel types (Slack-like)
const (
	ChannelTypePublic  = "public"  // Anyone can browse and join
	ChannelTypePrivate = "private" // Must be invited
	ChannelTypeDM      = "dm"      // 1:1 direct message (cannot leave)
	ChannelTypeGroupDM = "group_dm" // 3+ people (can leave, cannot remove others)
)

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
func (s *chatService) convertDMToGroupDM(ctx context.Context, channel *repository.ChatChannel) error {
	if channel.Type != ChannelTypeDM && channel.Type != "direct" {
		return nil
	}

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

	newName := "Group Conversation"
	if len(names) > 0 {
		if len(names) <= 3 {
			newName = strings.Join(names, ", ")
		} else {
			newName = fmt.Sprintf("%s, %s, %s +%d", names[0], names[1], names[2], len(names)-3)
		}
	}

	channel.Name = newName
	channel.Type = ChannelTypeGroupDM
	return s.chatRepo.UpdateChannel(ctx, channel)
}

// ============================================
// Channel Management
// ============================================

func (s *chatService) CreateChannel(ctx context.Context, name, channelType, targetID, workspaceID, creatorID string, isPrivate bool) (*repository.ChatChannel, error) {
	// Normalize channel type
	if channelType == "" {
		if isPrivate {
			channelType = ChannelTypePrivate
		} else {
			channelType = ChannelTypePublic
		}
	}

	channel := &repository.ChatChannel{
		Name:        name,
		Type:        channelType,
		TargetID:    targetID,
		WorkspaceID: workspaceID,
		CreatedBy:   creatorID,
		IsPrivate:   isPrivate || channelType == ChannelTypePrivate,
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

	// Broadcast
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

// DeleteChannel - Only allow deleting non-DM channels by creator
func (s *chatService) DeleteChannel(ctx context.Context, id, userID string) error {
	channel, err := s.chatRepo.GetChannelByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}

	switch channel.Type {
	case ChannelTypeDM, "direct", ChannelTypeGroupDM, "group":
		// Cannot delete DMs or group DMs - they persist forever
		return fmt.Errorf("conversations cannot be deleted")

	case ChannelTypePrivate, ChannelTypePublic:
		// Only creator can delete channels
		// TODO: Also allow workspace admins
		if channel.CreatedBy != userID {
			return ErrForbidden
		}
	}

	// Broadcast before deletion
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_channel_deleted"), map[string]interface{}{
			"channelId": id,
		}, "")
	}

	return s.chatRepo.DeleteChannel(ctx, id)
}



// ArchiveChannel archives a channel (Slack-like soft delete)
func (s *chatService) ArchiveChannel(ctx context.Context, channelID, userID string) error {
	channel, err := s.chatRepo.GetChannelByID(ctx, channelID)
	if err != nil {
		return ErrNotFound
	}

	switch channel.Type {
	case ChannelTypeDM, "direct", ChannelTypeGroupDM, "group":
		// Cannot archive DMs
		return fmt.Errorf("conversations cannot be archived")

	case ChannelTypePrivate, ChannelTypePublic:
		// Any member can archive (Slack default)
		isMember, _ := s.chatRepo.IsMember(ctx, channelID, userID)
		if !isMember {
			return ErrForbidden
		}
	}

	channel.IsArchived = true
	if err := s.chatRepo.UpdateChannel(ctx, channel); err != nil {
		return err
	}

	// System message
	user, _ := s.userRepo.FindByID(ctx, userID)
	s.sendSystemMessage(ctx, channelID, fmt.Sprintf("%s archived this channel", getNameOrUnknown(user)))

	// Broadcast
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_channel_archived"), map[string]interface{}{
			"channelId": channelID,
		}, "")
	}

	return nil
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
		Type:        ChannelTypeDM,
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
	channel, err := s.chatRepo.GetChannelByID(ctx, channelID)
	if err != nil {
		return ErrNotFound
	}

	switch channel.Type {
	case ChannelTypeDM, "direct", ChannelTypeGroupDM, "group":
		// Cannot self-join DMs or group DMs - must be added
		return fmt.Errorf("you must be added to this conversation")

	case ChannelTypePrivate:
		// Cannot self-join private channels - must be added
		isMember, _ := s.chatRepo.IsMember(ctx, channelID, userID)
		if !isMember {
			return fmt.Errorf("you must be invited to join this private channel")
		}
		return nil // Already a member

	case ChannelTypePublic:
		// Public channels - anyone can join
		member := &repository.ChatChannelMember{
			ChannelID: channelID,
			UserID:    userID,
		}
		return s.chatRepo.AddMember(ctx, member)
	}

	return fmt.Errorf("unknown channel type")
}

func (s *chatService) AddMemberToChannel(ctx context.Context, channelID, userID, addedByID string) error {
	channel, err := s.chatRepo.GetChannelByID(ctx, channelID)
	if err != nil {
		return ErrNotFound
	}

	// Check if adder is a member (required for all types except public)
	isMember, _ := s.chatRepo.IsMember(ctx, channelID, addedByID)
	
	switch channel.Type {
	case ChannelTypeDM, "direct":
		// Adding to 1:1 DM converts it to group DM
		if !isMember {
			return ErrForbidden
		}
		
	case ChannelTypeGroupDM, "group":
		// Anyone in group DM can add members
		if !isMember {
			return ErrForbidden
		}

	case ChannelTypePrivate:
		// Private channel - only members can add
		if !isMember {
			return ErrForbidden
		}

	case ChannelTypePublic:
		// Public channel - anyone in workspace can add (or user joins themselves)
		// No membership check needed
	}

	// Check if already a member
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

	// Convert 1:1 DM to Group DM when adding third person
	memberCount, _ := s.chatRepo.GetMemberCount(ctx, channelID)
	if (channel.Type == ChannelTypeDM || channel.Type == "direct") && memberCount > 2 {
		s.convertDMToGroupDM(ctx, channel)
	}

	addedUser, _ := s.userRepo.FindByID(ctx, userID)
	adderUser, _ := s.userRepo.FindByID(ctx, addedByID)

	// Broadcast
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

	// Notification
	if s.notifSvc != nil {
		s.notifSvc.SendChatAddedToChannel(
			ctx, userID, channelID, channel.Name,
			getNameOrUnknown(adderUser), channel.WorkspaceID,
			channel.Type == ChannelTypeDM || channel.Type == "direct",
		)
	}

	return nil
}

func (s *chatService) LeaveChannel(ctx context.Context, channelID, userID string) error {
	channel, err := s.chatRepo.GetChannelByID(ctx, channelID)
	if err != nil {
		return ErrNotFound
	}

	switch channel.Type {
	case ChannelTypeDM, "direct":
		// Cannot leave 1:1 DM
		return fmt.Errorf("cannot leave a direct message conversation")

	case ChannelTypeGroupDM, "group":
		// Can leave group DMs
		break

	case ChannelTypePublic, ChannelTypePrivate:
		// Can leave channels
		break
	}

	return s.RemoveMemberFromChannel(ctx, channelID, userID, userID)
}

func (s *chatService) RemoveMemberFromChannel(ctx context.Context, channelID, userID, removedByID string) error {
	channel, err := s.chatRepo.GetChannelByID(ctx, channelID)
	if err != nil {
		return ErrNotFound
	}

	memberCount, _ := s.chatRepo.GetMemberCount(ctx, channelID)

	switch channel.Type {
	case ChannelTypeDM, "direct":
		// 1:1 DM - CANNOT leave or remove
		return fmt.Errorf("cannot leave a direct message conversation")

	case ChannelTypeGroupDM, "group":
		// Group DM - can only leave yourself, cannot remove others
		if userID != removedByID {
			return fmt.Errorf("you can only leave group conversations yourself, not remove others")
		}

	case ChannelTypePrivate:
		// Private channel - any member can remove others
		isMember, _ := s.chatRepo.IsMember(ctx, channelID, removedByID)
		if !isMember {
			return ErrForbidden
		}

	case ChannelTypePublic:
		// Public channel - only creator/admin can remove others (self-leave always allowed)
		if userID != removedByID && channel.CreatedBy != removedByID {
			// TODO: Check workspace admin role here
			return ErrForbidden
		}
	}

	// Prevent last member from leaving (except group DMs which can become empty)
	if memberCount <= 1 && channel.Type != ChannelTypeGroupDM && channel.Type != "group" {
		return fmt.Errorf("cannot leave: you are the last member")
	}

	if err := s.chatRepo.RemoveMember(ctx, channelID, userID); err != nil {
		return err
	}

	removedUser, _ := s.userRepo.FindByID(ctx, userID)
	removerUser, _ := s.userRepo.FindByID(ctx, removedByID)

	// Broadcast
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
			s.notifSvc.SendChatRemovedFromChannel(ctx, userID, channel.Name, getNameOrUnknown(removerUser))
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
	// Broadcast message - EXCLUDE the sender
if s.broadcaster != nil && channel != nil {
    s.broadcaster.BroadcastToWorkspace(channel.WorkspaceID, socket.MessageType("chat_message"), map[string]interface{}{
        "channelId": channelID,
        "messageId": message.ID,
        "message":   message,
    }, userID) // ✅ CHANGE: was "" now userID - excludes sender
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
