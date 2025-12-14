// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/handlers"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/config"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/cron"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/db"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/email"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/seed"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ============================================
	// Initialize Database Connection
	// ============================================
	postgresDB, err := db.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()
	log.Println("✅ Connected to PostgreSQL")

	// Initialize Repositories
	repos := repository.NewRepositories(postgresDB.Pool)
	log.Println("📦 Repositories initialized")

	// ============================================
	// Initialize Redis (optional cache)
	// ============================================
	var redisDB *db.RedisDB
	if cfg.RedisURL != "" {
		redisDB, err = db.NewRedisDB(cfg.RedisURL)
		if err != nil {
			log.Printf("⚠️  Failed to connect to Redis: %v (continuing without cache)", err)
		} else {
			defer redisDB.Close()
			log.Println("⚡ Redis cache enabled")
		}
	}

	// ============================================
	// Initialize Email Service (optional)
	// ============================================
	var emailSvc *email.Service
	if cfg.SMTPHost != "" {
		emailSvc = email.NewService(&email.Config{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			User:     cfg.SMTPUser,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			FromName: cfg.SMTPFromName,
			UseTLS:   cfg.SMTPUseTLS,
		})
		log.Println("📧 Email service initialized")
	} else {
		log.Println("⚠️  Email not configured (SMTP_HOST not set)")
	}

	// ============================================
	// Initialize WebSocket Hub
	// ============================================
	hub := socket.NewHub()
	go hub.Run()
	broadcaster := socket.NewBroadcaster(hub)

	// WebSocket handler with JWT secret for self-authentication
	wsHandler := socket.NewHandler(hub, cfg.JWTSecret)
	log.Println("🔌 WebSocket hub initialized")

	// ============================================
	// Seed Data (for development)
	// ============================================
	if cfg.Environment == "development" {
		seed.SeedData(repos)
	}

	// ============================================
	// Initialize Services
	// ============================================
	notificationSvc := notification.NewServiceWithRepos(
		repos.NotificationRepo,
		repos.UserRepo,
		repos.ProjectRepo,
	)

	// Set broadcaster on notification service for real-time updates
	notificationSvc.SetBroadcaster(broadcaster)

	services := service.NewServices(cfg, repos, notificationSvc)

	// Initialize handlers
	h := handlers.NewHandlers(services)

	// Initialize cron scheduler
	cronScheduler := cron.NewSchedulerWithRepos(
		services,
		notificationSvc,
		repos.TaskRepo,
		repos.SprintRepo,
		repos.ProjectRepo,
		repos.UserRepo,
		repos.NotificationRepo,
	)
	cronScheduler.Start()
	defer cronScheduler.Stop()

	// ============================================
	// Create Gin Router
	// ============================================
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"timestamp":  time.Now(),
			"database":   "connected",
			"cache":      getCacheStatus(redisDB),
			"websocket":  "active",
			"ws_clients": hub.GetConnectedClientsCount(),
			"email":      getEmailStatus(emailSvc),
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// ============================================
		// Public routes (no auth required)
		// ============================================

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Auth.Register)
			auth.POST("/login", h.Auth.Login)
			auth.POST("/refresh", h.Auth.RefreshToken)
			auth.POST("/logout", h.Auth.Logout)
		}

		// ============================================
		// WebSocket route (handles its own auth via query param)
		// IMPORTANT: Must be BEFORE the protected middleware group
		// ============================================
		api.GET("/ws", wsHandler.HandleWebSocket)

		// ============================================
		// Protected routes (require auth middleware)
		// ============================================
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(services.Auth))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", h.User.GetCurrentUser)
				users.PUT("/me", h.User.UpdateCurrentUser)
			}

			// Workspace routes
			workspaces := protected.Group("/workspaces")
			{
				workspaces.GET("", h.Workspace.List)
				workspaces.POST("", h.Workspace.Create)
				workspaces.GET("/:id", h.Workspace.Get)
				workspaces.PUT("/:id", h.Workspace.Update)
				workspaces.DELETE("/:id", h.Workspace.Delete)
				workspaces.GET("/:id/members", h.Workspace.ListMembers)
				workspaces.POST("/:id/members", h.Workspace.AddMember)
				workspaces.PUT("/:id/members/:userId", h.Workspace.UpdateMemberRole)
				workspaces.DELETE("/:id/members/:userId", h.Workspace.RemoveMember)
				workspaces.GET("/:id/spaces", h.Space.ListByWorkspace)
				workspaces.POST("/:id/spaces", h.Space.Create)
			}

			// Space routes
			spaces := protected.Group("/spaces")
			{
				spaces.GET("/:id", h.Space.Get)
				spaces.PUT("/:id", h.Space.Update)
				spaces.DELETE("/:id", h.Space.Delete)
				spaces.GET("/:id/projects", h.Project.ListBySpace)
				spaces.POST("/:id/projects", h.Project.Create)
			}

			// Project routes
			projects := protected.Group("/projects")
			{
				projects.GET("/:id", h.Project.Get)
				projects.PUT("/:id", h.Project.Update)
				projects.DELETE("/:id", h.Project.Delete)
				projects.GET("/:id/members", h.Project.ListMembers)
				projects.POST("/:id/members", h.Project.AddMember)
				projects.DELETE("/:id/members/:userId", h.Project.RemoveMember)
				projects.GET("/:id/sprints", h.Sprint.ListByProject)
				projects.POST("/:id/sprints", h.Sprint.Create)
				projects.GET("/:id/tasks", h.Task.ListByProject)
				projects.POST("/:id/tasks", h.Task.Create)
				projects.GET("/:id/labels", h.Label.ListByProject)
				projects.POST("/:id/labels", h.Label.Create)
			}

			// Sprint routes
			sprints := protected.Group("/sprints")
			{
				sprints.GET("/:id", h.Sprint.Get)
				sprints.PUT("/:id", h.Sprint.Update)
				sprints.DELETE("/:id", h.Sprint.Delete)
				sprints.POST("/:id/start", h.Sprint.Start)
				sprints.POST("/:id/complete", h.Sprint.Complete)
				sprints.GET("/:id/tasks", h.Task.ListBySprint)
			}

			// Task routes
			tasks := protected.Group("/tasks")
			{
				tasks.GET("/:id", h.Task.Get)
				tasks.PUT("/:id", h.Task.Update)
				tasks.PATCH("/:id", h.Task.PartialUpdate)
				tasks.DELETE("/:id", h.Task.Delete)
				tasks.PUT("/bulk", h.Task.BulkUpdate)
				tasks.GET("/:id/comments", h.Comment.ListByTask)
				tasks.POST("/:id/comments", h.Comment.Create)
			}

			// Comment routes
			comments := protected.Group("/comments")
			{
				comments.PUT("/:id", h.Comment.Update)
				comments.DELETE("/:id", h.Comment.Delete)
			}

			// Label routes
			labels := protected.Group("/labels")
			{
				labels.PUT("/:id", h.Label.Update)
				labels.DELETE("/:id", h.Label.Delete)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", h.Notification.List)
				notifications.GET("/count", h.Notification.Count)
				notifications.PUT("/:id/read", h.Notification.MarkRead)
				notifications.PUT("/read-all", h.Notification.MarkAllRead)
				notifications.DELETE("/:id", h.Notification.Delete)
				notifications.DELETE("", h.Notification.DeleteAll)
			}
		}
	}

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getCacheStatus(redisDB *db.RedisDB) string {
	if redisDB != nil {
		return "connected"
	}
	return "disabled"
}

func getEmailStatus(emailSvc *email.Service) string {
	if emailSvc != nil {
		return "configured"
	}
	return "disabled"
}
