// internal/socket/client.go
package socket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket connection constants
const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (4KB)
	maxMessageSize int64 = 4096
)

// ClientMessage represents an incoming message from a client
type ClientMessage struct {
	Action  string                 `json:"action"`
	Room    string                 `json:"room,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
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
				log.Printf("[Client] WebSocket error for user %s: %v", c.UserID, err)
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
		log.Printf("[Client] Error parsing message from user %s: %v", c.UserID, err)
		return
	}

	log.Printf("[Client] Received action=%s room=%s from user=%s", msg.Action, msg.Room, c.UserID)

	switch msg.Action {
	case "join":
		if msg.Room != "" {
			c.Hub.JoinRoom(c, msg.Room)
			c.sendAck("joined", msg.Room)
		}

	case "leave":
		if msg.Room != "" {
			c.Hub.LeaveRoom(c, msg.Room)
			c.sendAck("left", msg.Room)
		}

	case "typing":
		if msg.Room != "" {
			c.Hub.SendToRoom(msg.Room, MessageUserTyping, map[string]interface{}{
				"userId": c.UserID,
				"room":   msg.Room,
			}, c.UserID)
		}

	case "ping":
		c.lastPing = time.Now()
		c.sendPong()

	case "pong":
		c.lastPing = time.Now()

	default:
		log.Printf("[Client] Unknown action: %s from user: %s", msg.Action, c.UserID)
	}
}

func (c *Client) sendAck(action, room string) {
	msg := Message{
		Type: MessageAck,
		Payload: map[string]interface{}{
			"action": action,
			"room":   room,
		},
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(msg)

	select {
	case c.Send <- data:
	default:
		log.Printf("[Client] Failed to send ack to user %s", c.UserID)
	}
}

func (c *Client) sendPong() {
	msg := Message{
		Type: MessagePong,
		Payload: map[string]interface{}{
			"time": time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(msg)

	select {
	case c.Send <- data:
	default:
		log.Printf("[Client] Failed to send pong to user %s", c.UserID)
	}
}
