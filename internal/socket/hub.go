// Package socket provides WebSocket functionality for real-time notifications
package socket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Notification messages
	MessageNotification      MessageType = "notification"
	MessageNotificationRead  MessageType = "notification_read"
	MessageNotificationCount MessageType = "notification_count"

	// Task messages
	MessageTaskCreated       MessageType = "task_created"
	MessageTaskUpdated       MessageType = "task_updated"
	MessageTaskDeleted       MessageType = "task_deleted"
	MessageTaskStatusChanged MessageType = "task_status_changed"
	MessageTaskAssigned      MessageType = "task_assigned"

	// Sprint messages
	MessageSprintStarted   MessageType = "sprint_started"
	MessageSprintCompleted MessageType = "sprint_completed"

	// Project messages
	MessageProjectUpdated MessageType = "project_updated"
	MessageMemberAdded    MessageType = "member_added"
	MessageMemberRemoved  MessageType = "member_removed"

	// Team messages
	MessageTeamCreated MessageType = "team_created"
	MessageTeamUpdated MessageType = "team_updated"
	MessageTeamDeleted MessageType = "team_deleted"
	MessageTeamMemberAdded   MessageType = "team_member_added"
	MessageTeamMemberRemoved MessageType = "team_member_removed"

	// User presence
	MessageUserOnline  MessageType = "user_online"
	MessageUserOffline MessageType = "user_offline"
	MessageUserTyping  MessageType = "user_typing"

	// Comment messages
	MessageCommentAdded   MessageType = "comment_added"
	MessageCommentUpdated MessageType = "comment_updated"
	MessageCommentDeleted MessageType = "comment_deleted"

	// System messages
	MessagePing MessageType = "ping"
	MessagePong MessageType = "pong"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType            `json:"type"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID        string
	UserID    string
	Conn      *websocket.Conn
	Hub       *Hub
	Send      chan []byte
	Rooms     map[string]bool // Subscribed rooms (workspace:id, project:id, etc.)
	mu        sync.Mutex
	lastPing  time.Time
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients indexed by user ID for direct messaging
	userClients map[string]map[*Client]bool

	// Clients indexed by room for broadcasting
	roomClients map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast to all clients
	broadcast chan []byte

	// Broadcast to specific room
	roomBroadcast chan *RoomMessage

	// Direct message to specific user
	directMessage chan *DirectMessage

	mu sync.RWMutex
}

// RoomMessage represents a message to be sent to a specific room
type RoomMessage struct {
	Room    string
	Message []byte
	Exclude string // User ID to exclude from broadcast
}

// DirectMessage represents a message to be sent to a specific user
type DirectMessage struct {
	UserID  string
	Message []byte
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		userClients:   make(map[string]map[*Client]bool),
		roomClients:   make(map[string]map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan []byte, 256),
		roomBroadcast: make(chan *RoomMessage, 256),
		directMessage: make(chan *DirectMessage, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	// Start ping ticker
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAll(message)

		case rm := <-h.roomBroadcast:
			h.broadcastToRoom(rm)

		case dm := <-h.directMessage:
			h.sendToUser(dm)

		case <-pingTicker.C:
			h.pingClients()
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// Index by user ID
	if h.userClients[client.UserID] == nil {
		h.userClients[client.UserID] = make(map[*Client]bool)
	}
	h.userClients[client.UserID][client] = true

	log.Printf("🔌 Client connected: user=%s, id=%s", client.UserID, client.ID)

	// Broadcast user online status
	go h.BroadcastUserStatus(client.UserID, true)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)

		// Remove from user index
		if clients, ok := h.userClients[client.UserID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.userClients, client.UserID)
				// User went offline
				go h.BroadcastUserStatus(client.UserID, false)
			}
		}

		// Remove from all rooms
		for room := range client.Rooms {
			if clients, ok := h.roomClients[room]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.roomClients, room)
				}
			}
		}

		close(client.Send)
		log.Printf("🔌 Client disconnected: user=%s, id=%s", client.UserID, client.ID)
	}
}

func (h *Hub) broadcastToAll(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

func (h *Hub) broadcastToRoom(rm *RoomMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.roomClients[rm.Room]; ok {
		for client := range clients {
			// Skip excluded user
			if rm.Exclude != "" && client.UserID == rm.Exclude {
				continue
			}
			select {
			case client.Send <- rm.Message:
			default:
				go func(c *Client) {
					h.unregister <- c
				}(client)
			}
		}
	}
}

func (h *Hub) sendToUser(dm *DirectMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.userClients[dm.UserID]; ok {
		for client := range clients {
			select {
			case client.Send <- dm.Message:
			default:
				go func(c *Client) {
					h.unregister <- c
				}(client)
			}
		}
	}
}

func (h *Hub) pingClients() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := Message{
		Type:      MessagePing,
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(msg)

	for client := range h.clients {
		select {
		case client.Send <- data:
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// ============================================
// Public Methods for Broadcasting
// ============================================

// JoinRoom adds a client to a room
func (h *Hub) JoinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	client.Rooms[room] = true
	client.mu.Unlock()

	if h.roomClients[room] == nil {
		h.roomClients[room] = make(map[*Client]bool)
	}
	h.roomClients[room][client] = true

	log.Printf("👥 Client joined room: user=%s, room=%s", client.UserID, room)
}

// LeaveRoom removes a client from a room
func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	delete(client.Rooms, room)
	client.mu.Unlock()

	if clients, ok := h.roomClients[room]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.roomClients, room)
		}
	}

	log.Printf("👋 Client left room: user=%s, room=%s", client.UserID, room)
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID string, msgType MessageType, payload map[string]interface{}) {
	msg := Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.directMessage <- &DirectMessage{
		UserID:  userID,
		Message: data,
	}
}

// SendToRoom broadcasts a message to all clients in a room
func (h *Hub) SendToRoom(room string, msgType MessageType, payload map[string]interface{}, excludeUserID string) {
	msg := Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.roomBroadcast <- &RoomMessage{
		Room:    room,
		Message: data,
		Exclude: excludeUserID,
	}
}

// BroadcastUserStatus broadcasts user online/offline status
func (h *Hub) BroadcastUserStatus(userID string, online bool) {
	msgType := MessageUserOffline
	if online {
		msgType = MessageUserOnline
	}

	msg := Message{
		Type: msgType,
		Payload: map[string]interface{}{
			"userId": userID,
			"online": online,
		},
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(msg)
	h.broadcast <- data
}

// GetOnlineUsers returns a list of online user IDs
func (h *Hub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.userClients))
	for userID := range h.userClients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline checks if a user is currently connected
func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, ok := h.userClients[userID]
	return ok
}

// GetRoomClients returns the number of clients in a room
func (h *Hub) GetRoomClients(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.roomClients[room]; ok {
		return len(clients)
	}
	return 0
}
