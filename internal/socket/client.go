package socket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// In production, restrict to your domains
		return true
	},
}

// ClientMessage represents an incoming message from a client
type ClientMessage struct {
	Action  string                 `json:"action"`
	Room    string                 `json:"room,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, userID string, conn *websocket.Conn) *Client {
	return &Client{
		ID:       uuid.New().String(),
		UserID:   userID,
		Conn:     conn,
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Rooms:    make(map[string]bool),
		lastPing: time.Now(),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.lastPing = time.Now()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse and handle the message
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	switch msg.Action {
	case "join":
		// Join a room (workspace, project, etc.)
		if msg.Room != "" {
			c.Hub.JoinRoom(c, msg.Room)
			c.sendAck("joined", msg.Room)
		}

	case "leave":
		// Leave a room
		if msg.Room != "" {
			c.Hub.LeaveRoom(c, msg.Room)
			c.sendAck("left", msg.Room)
		}

	case "typing":
		// Broadcast typing indicator
		if msg.Room != "" {
			c.Hub.SendToRoom(msg.Room, MessageUserTyping, map[string]interface{}{
				"userId": c.UserID,
				"room":   msg.Room,
			}, c.UserID)
		}

	case "pong":
		// Client responded to ping
		c.lastPing = time.Now()

	default:
		log.Printf("Unknown action: %s", msg.Action)
	}
}

func (c *Client) sendAck(action, room string) {
	msg := Message{
		Type: "ack",
		Payload: map[string]interface{}{
			"action": action,
			"room":   room,
		},
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(msg)
	c.Send <- data
}

// ============================================
// WebSocket Handler
// ============================================

// Handler handles WebSocket connections
type Handler struct {
	Hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{Hub: hub}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create new client
	client := NewClient(h.Hub, userID.(string), conn)

	// Register client with hub
	h.Hub.register <- client

	// Auto-join user's personal room for direct notifications
	h.Hub.JoinRoom(client, "user:"+userID.(string))

	// Start read/write goroutines
	go client.WritePump()
	go client.ReadPump()
}

// HandleWebSocketWithToken handles WebSocket with token in query params
// Use this when you can't set headers (e.g., browser WebSocket API)
func (h *Handler) HandleWebSocketWithToken(c *gin.Context) {
	// Token should be validated by the caller and userID set in context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create new client
	client := NewClient(h.Hub, userID.(string), conn)

	// Register client with hub
	h.Hub.register <- client

	// Auto-join user's personal room
	h.Hub.JoinRoom(client, "user:"+userID.(string))

	// Start read/write goroutines
	go client.WritePump()
	go client.ReadPump()
}
