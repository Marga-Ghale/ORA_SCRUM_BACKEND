// internal/socket/handler.go
package socket

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

// Handler handles WebSocket connections
type Handler struct {
	Hub       *Hub
	JWTSecret string
}

// NewHandler creates a new WebSocket handler
// jwtSecret is required for validating tokens from query params
func NewHandler(hub *Hub, jwtSecret string) *Handler {
	return &Handler{
		Hub:       hub,
		JWTSecret: jwtSecret,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
// This handler validates JWT from query parameter because browser WebSocket API cannot set custom headers
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Get token from query parameter
	tokenString := c.Query("token")
	if tokenString == "" {
		// Also try Authorization header as fallback
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if tokenString == "" {
		log.Println("[WebSocket] No token provided")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}

	// Parse and validate JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.JWTSecret), nil
	})

	if err != nil {
		log.Printf("[WebSocket] Token parse error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	if !token.Valid {
		log.Println("[WebSocket] Token is not valid")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Extract user ID from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("[WebSocket] Invalid token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Check token expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			log.Println("[WebSocket] Token expired")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		log.Println("[WebSocket] No user ID in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user ID in token"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade error: %v", err)
		return
	}

	log.Printf("[WebSocket] ✅ Client connected: userID=%s", userID)

	// Create new client
	client := NewClient(h.Hub, userID, conn)

	// Register client with hub
	h.Hub.register <- client

	// Auto-join user's personal room for direct notifications
	h.Hub.JoinRoom(client, "user:"+userID)

	// Start read/write goroutines
	go client.WritePump()
	go client.ReadPump()
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
