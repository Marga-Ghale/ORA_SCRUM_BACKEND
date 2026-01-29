package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("❌ [Auth] Missing Authorization header - Path: %s", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("❌ [Auth] Invalid header format - Path: %s", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		token, err := authService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			log.Printf("❌ [Auth] Invalid token - Path: %s, Error: %v", c.Request.URL.Path, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract user ID from token
		userID, err := authService.GetUserIDFromToken(token)
		if err != nil {
			log.Printf("❌ [Auth] Failed to extract userID - Path: %s, Error: %v", c.Request.URL.Path, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user ID in context for handlers
		c.Set("userID", userID)
		log.Printf("✅ [Auth] User authenticated - UserID: %s, Path: %s", userID, c.Request.URL.Path)
		c.Next()
	}
}

// OptionalAuthMiddleware allows requests without authentication but sets user context if present
func OptionalAuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		token, err := authService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.Next()
			return
		}

		userID, err := authService.GetUserIDFromToken(token)
		if err != nil {
			c.Next()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// RequestLogger logs all incoming requests with details
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log after request
		duration := time.Since(start)
		status := c.Writer.Status()
		
		// Color code based on status
		statusEmoji := "✅"
		if status >= 400 && status < 500 {
			statusEmoji = "⚠️"
		} else if status >= 500 {
			statusEmoji = "❌"
		}

		log.Printf("%s [%s] %s %d - %v", statusEmoji, method, path, status, duration)
		
		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("❌ [Error] %v", e.Err)
			}
		}
	}
}

// ErrorLogger logs detailed error information
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors after request processing
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("❌ [Error Details] Path: %s, Method: %s, Error: %v, Type: %v", 
					c.Request.URL.Path, 
					c.Request.Method, 
					err.Err, 
					err.Type,
				)
			}
		}
	}
}

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	return userID.(string)
}

// RequireUserID returns error if user ID is not in context
func RequireUserID(c *gin.Context) (string, bool) {
	userID := GetUserID(c)
	if userID == "" {
		log.Printf("❌ [Auth] User not authenticated - Path: %s", c.Request.URL.Path)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return "", false
	}
	return userID, true
}