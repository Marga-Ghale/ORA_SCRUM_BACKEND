// internal/socket/hub.go
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
    MessageTaskPositionChanged MessageType = "task_position_changed"


	// Sprint messages
	MessageSprintStarted   MessageType = "sprint_started"
	MessageSprintCompleted MessageType = "sprint_completed"

	// Project messages
	MessageProjectUpdated MessageType = "project_updated"
	

	MessageMemberAdded       MessageType = "member_added"
	MessageMemberRemoved     MessageType = "member_removed"
	MessageMemberRoleUpdated MessageType = "member_role_updated"

	// Team messages
	MessageTeamCreated       MessageType = "team_created"
	MessageTeamUpdated       MessageType = "team_updated"
	MessageTeamDeleted       MessageType = "team_deleted"
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
	MessageAck  MessageType = "ack"

	// ✅ NEW: Workspace CRUD messages
	MessageWorkspaceCreated MessageType = "workspace_created"
	MessageWorkspaceUpdated MessageType = "workspace_updated"
	MessageWorkspaceDeleted MessageType = "workspace_deleted"

	// ✅ NEW: Space CRUD messages
	MessageSpaceCreated MessageType = "space_created"
	MessageSpaceUpdated MessageType = "space_updated"
	MessageSpaceDeleted MessageType = "space_deleted"

	// ✅ NEW: Folder CRUD messages
	MessageFolderCreated MessageType = "folder_created"
	MessageFolderUpdated MessageType = "folder_updated"
	MessageFolderDeleted MessageType = "folder_deleted"

	// ✅ NEW: Project CRUD messages
	MessageProjectCreated MessageType = "project_created"
	MessageProjectDeleted MessageType = "project_deleted"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType            `json:"type"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID       string
	UserID   string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte
	Rooms    map[string]bool // Subscribed rooms (workspace:id, project:id, etc.)
	mu       sync.Mutex
	lastPing time.Time
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
	log.Println("[Hub] WebSocket hub started")

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

	log.Printf("[Hub] ✅ Client registered: user=%s, id=%s, total_clients=%d",
		client.UserID, client.ID, len(h.clients))

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
				// User went offline (no more connections)
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
		log.Printf("[Hub] ❌ Client disconnected: user=%s, id=%s, total_clients=%d",
			client.UserID, client.ID, len(h.clients))
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

	clients, ok := h.roomClients[rm.Room]
	if !ok {
		log.Printf("[Hub] Room not found: %s", rm.Room)
		return
	}

	sentCount := 0
	for client := range clients {
		// Skip excluded user
		if rm.Exclude != "" && client.UserID == rm.Exclude {
			continue
		}
		select {
		case client.Send <- rm.Message:
			sentCount++
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
	log.Printf("[Hub] Broadcast to room %s: sent to %d clients", rm.Room, sentCount)
}

func (h *Hub) sendToUser(dm *DirectMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.userClients[dm.UserID]
	if !ok {
		log.Printf("[Hub] User not connected: %s", dm.UserID)
		return
	}

	sentCount := 0
	for client := range clients {
		select {
		case client.Send <- dm.Message:
			sentCount++
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
	log.Printf("[Hub] Direct message to user %s: sent to %d clients", dm.UserID, sentCount)
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
// Public Methods for Room Management
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

	log.Printf("[Hub] 👥 Client joined room: user=%s, room=%s", client.UserID, room)
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

	log.Printf("[Hub] 👋 Client left room: user=%s, room=%s", client.UserID, room)
}

// ============================================
// Public Methods for Sending Messages
// ============================================

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID string, msgType MessageType, payload map[string]interface{}) {
	msg := Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Hub] Error marshaling message: %v", err)
		return
	}

	log.Printf("[Hub] 📤 SendToUser: user=%s, type=%s", userID, msgType)

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
		log.Printf("[Hub] Error marshaling message: %v", err)
		return
	}

	log.Printf("[Hub] 📤 SendToRoom: room=%s, type=%s, exclude=%s", room, msgType, excludeUserID)

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

// ============================================
// Query Methods
// ============================================

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

// GetConnectedClientsCount returns total connected clients
func (h *Hub) GetConnectedClientsCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}