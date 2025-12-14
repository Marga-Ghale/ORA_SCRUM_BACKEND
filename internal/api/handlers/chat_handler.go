package handlers

import (
	"net/http"
	"strconv"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	chatSvc service.ChatService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatSvc service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

// ============================================
// Request/Response DTOs
// ============================================

type CreateChannelRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=project space team direct"`
	TargetID    string `json:"targetId" binding:"required"`
	WorkspaceID string `json:"workspaceId" binding:"required"`
	IsPrivate   bool   `json:"isPrivate"`
}

type CreateDirectChannelRequest struct {
	UserID      string `json:"userId" binding:"required"`
	WorkspaceID string `json:"workspaceId" binding:"required"`
}

type SendMessageRequest struct {
	Content     string  `json:"content" binding:"required,min=1,max=10000"`
	MessageType string  `json:"messageType,omitempty"`
	ParentID    *string `json:"parentId,omitempty"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=10000"`
}

type ReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

// ============================================
// Channel Endpoints
// ============================================

// CreateChannel creates a new chat channel
func (h *ChatHandler) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	channel, err := h.chatSvc.CreateChannel(c.Request.Context(), req.Name, req.Type, req.TargetID, req.WorkspaceID, userID, req.IsPrivate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

// GetChannel retrieves a channel by ID
func (h *ChatHandler) GetChannel(c *gin.Context) {
	channelID := c.Param("id")

	channel, err := h.chatSvc.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// GetChannelByTarget retrieves a channel by target type and ID
func (h *ChatHandler) GetChannelByTarget(c *gin.Context) {
	targetType := c.Query("type")
	targetID := c.Query("targetId")

	if targetType == "" || targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type and targetId are required"})
		return
	}

	channel, err := h.chatSvc.GetChannelByTarget(c.Request.Context(), targetType, targetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// ListChannels lists all channels for the current user
func (h *ChatHandler) ListChannels(c *gin.Context) {
	userID := c.GetString("userID")

	channels, err := h.chatSvc.ListChannels(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channels)
}

// ListWorkspaceChannels lists all channels in a workspace
func (h *ChatHandler) ListWorkspaceChannels(c *gin.Context) {
	workspaceID := c.Param("id")

	channels, err := h.chatSvc.ListWorkspaceChannels(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channels)
}

// DeleteChannel deletes a channel
func (h *ChatHandler) DeleteChannel(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.GetString("userID")

	if err := h.chatSvc.DeleteChannel(c.Request.Context(), channelID, userID); err != nil {
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this channel"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================
// Direct Message Endpoints
// ============================================

// CreateDirectChannel creates or gets a direct message channel
func (h *ChatHandler) CreateDirectChannel(c *gin.Context) {
	var req CreateDirectChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	channel, err := h.chatSvc.CreateDirectChannel(c.Request.Context(), userID, req.UserID, req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// ============================================
// Membership Endpoints
// ============================================

// JoinChannel joins a channel
func (h *ChatHandler) JoinChannel(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.GetString("userID")

	if err := h.chatSvc.JoinChannel(c.Request.Context(), channelID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// LeaveChannel leaves a channel
func (h *ChatHandler) LeaveChannel(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.GetString("userID")

	if err := h.chatSvc.LeaveChannel(c.Request.Context(), channelID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetChannelMembers gets channel members
func (h *ChatHandler) GetChannelMembers(c *gin.Context) {
	channelID := c.Param("id")

	members, err := h.chatSvc.GetChannelMembers(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}

// MarkAsRead marks channel as read
func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.GetString("userID")

	if err := h.chatSvc.MarkChannelAsRead(c.Request.Context(), channelID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================
// Message Endpoints
// ============================================

// SendMessage sends a message to a channel
func (h *ChatHandler) SendMessage(c *gin.Context) {
	channelID := c.Param("id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	message, err := h.chatSvc.SendMessage(c.Request.Context(), channelID, userID, req.Content, req.MessageType, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetMessages gets messages from a channel
func (h *ChatHandler) GetMessages(c *gin.Context) {
	channelID := c.Param("id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.chatSvc.GetMessages(c.Request.Context(), channelID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// GetThreadMessages gets thread replies
func (h *ChatHandler) GetThreadMessages(c *gin.Context) {
	messageID := c.Param("messageId")

	messages, err := h.chatSvc.GetThreadMessages(c.Request.Context(), messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// UpdateMessage edits a message
func (h *ChatHandler) UpdateMessage(c *gin.Context) {
	messageID := c.Param("messageId")

	var req UpdateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	message, err := h.chatSvc.EditMessage(c.Request.Context(), messageID, userID, req.Content)
	if err != nil {
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own messages"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, message)
}

// DeleteMessage deletes a message
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	messageID := c.Param("messageId")
	userID := c.GetString("userID")

	if err := h.chatSvc.DeleteMessage(c.Request.Context(), messageID, userID); err != nil {
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own messages"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================
// Reaction Endpoints
// ============================================

// AddReaction adds a reaction to a message
func (h *ChatHandler) AddReaction(c *gin.Context) {
	messageID := c.Param("messageId")

	var req ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	if err := h.chatSvc.AddReaction(c.Request.Context(), messageID, userID, req.Emoji); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RemoveReaction removes a reaction from a message
func (h *ChatHandler) RemoveReaction(c *gin.Context) {
	messageID := c.Param("messageId")
	emoji := c.Query("emoji")
	userID := c.GetString("userID")

	if err := h.chatSvc.RemoveReaction(c.Request.Context(), messageID, userID, emoji); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetReactions gets reactions for a message
func (h *ChatHandler) GetReactions(c *gin.Context) {
	messageID := c.Param("messageId")

	reactions, err := h.chatSvc.GetReactions(c.Request.Context(), messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reactions)
}

// ============================================
// Unread Count Endpoints
// ============================================

// GetUnreadCount gets unread count for a channel
func (h *ChatHandler) GetUnreadCount(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.GetString("userID")

	count, err := h.chatSvc.GetUnreadCount(c.Request.Context(), channelID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread": count})
}

// GetAllUnreadCounts gets unread counts for all channels
func (h *ChatHandler) GetAllUnreadCounts(c *gin.Context) {
	userID := c.GetString("userID")

	counts, err := h.chatSvc.GetAllUnreadCounts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// AddMemberRequest for adding a member to a channel
type AddMemberRequest struct {
	UserID string `json:"userId" binding:"required"`
}

// AddMember adds a member to a channel
func (h *ChatHandler) AddMember(c *gin.Context) {
	channelID := c.Param("id")

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.GetString("userID")

	if err := h.chatSvc.AddMemberToChannel(c.Request.Context(), channelID, req.UserID, currentUserID); err != nil {
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add members"})
			return
		}
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
