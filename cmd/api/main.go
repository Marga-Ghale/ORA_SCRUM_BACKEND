// main.go
package main

import (
	"context"
	"database/sql"
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
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type Postgres struct {
	Pool *pgxpool.Pool
	DB   *sql.DB
}

func main() {
	// ============================================
	// Load environment variables
	// ============================================
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// ============================================
	// Load configuration
	// ============================================
	cfg := config.Load()

	// ============================================
	// Set Gin mode
	// ============================================
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ============================================
	// Run Database Migrations FIRST
	// ============================================
	log.Println("🔄 Running database migrations...")
	migrationsPath := "./internal/db/migrations"
	if err := db.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Database migrations completed")

	// ============================================
	// Initialize PostgreSQL (pgxpool + sql.DB)
	// ============================================
	ctx := context.Background()

	pgPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to create pgx pool: %v", err)
	}
	defer pgPool.Close()

	sqlDB, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to open sql DB: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping sql DB: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	// ============================================
	// Initialize Repositories
	// ============================================
	repos := repository.NewRepositories(pgPool, sqlDB)
	log.Println("📦 Repositories initialized")

	// ============================================
	// Initialize Redis (optional)
	// ============================================
	var redisDB *db.RedisDB
	if cfg.RedisURL != "" {
		redisDB, err = db.NewRedisDB(cfg.RedisURL)
		if err != nil {
			log.Printf("⚠️ Failed to connect to Redis: %v (continuing without cache)", err)
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
	if cfg.Environment != "production" {
		log.Println("🌱 Seeding development data...")
		seed.SeedData(repos)
	}

	// ============================================
	// Initialize Notification Service
	// ============================================
	notificationSvc := notification.NewServiceWithRepos(
		repos.NotificationRepo,
		repos.UserRepo,
		repos.ProjectRepo,
	)
	notificationSvc.SetBroadcaster(broadcaster)

	// ============================================
	// Initialize All Services
	// ============================================
	services := service.NewServices(&service.ServiceDeps{
		Config:      cfg,
		Repos:       repos,
		NotifSvc:    notificationSvc,
		EmailSvc:    emailSvc,
		Broadcaster: broadcaster,
	})
	log.Println("✨ All services initialized")

	// ============================================
	// Initialize Handlers
	// ============================================
	h := handlers.NewHandlers(services)
	teamHandler := handlers.NewTeamHandler(services.Team)
	activityHandler := handlers.NewActivityHandler(services.Activity)
	chatHandler := handlers.NewChatHandler(services.Chat)
	invitationHandler := handlers.NewInvitationHandler(services.Invitation)

	// ============================================
	// Initialize Cron Scheduler
	// ============================================
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
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Auth.Register)
			auth.POST("/login", h.Auth.Login)
			auth.POST("/refresh", h.Auth.RefreshToken)
			auth.POST("/logout", h.Auth.Logout)
		}

		// Public invitation routes (for accepting without login)
		publicInvitations := api.Group("/invitations")
		{
			publicInvitations.GET("/link/:token", invitationHandler.GetLinkInvitation)
		}

		// WebSocket route
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
				users.GET("/search", h.User.SearchUsers)
			}

			// Workspace routes
			workspaces := protected.Group("/workspaces")
			{
				workspaces.GET("", h.Workspace.List)
				workspaces.POST("", h.Workspace.Create)
				workspaces.GET("/:id", h.Workspace.Get)
				workspaces.PUT("/:id", h.Workspace.Update)
				workspaces.DELETE("/:id", h.Workspace.Delete)

				// Invitations
				workspaces.POST("/:id/invitations", invitationHandler.CreateWorkspaceInvitation)
				workspaces.GET("/:id/invitations", invitationHandler.GetWorkspaceInvitations)

				// Spaces
				workspaces.GET("/:id/spaces", h.Space.ListByWorkspace)
				workspaces.POST("/:id/spaces", h.Space.Create)

				// Teams
				workspaces.GET("/:id/teams", teamHandler.ListByWorkspace)

				// Chat channels
				workspaces.GET("/:id/chat/channels", chatHandler.ListWorkspaceChannels)
			}

			// Space routes
			spaces := protected.Group("/spaces")
			{
				spaces.GET("/:id", h.Space.Get)
				spaces.PUT("/:id", h.Space.Update)
				spaces.DELETE("/:id", h.Space.Delete)
				
				// Folder routes
				spaces.GET("/:id/folders", h.Folder.ListBySpace)
				spaces.POST("/:id/folders", h.Folder.Create)
				
				// Project routes
				spaces.GET("/:id/projects", h.Project.ListBySpace)
				spaces.POST("/:id/projects", h.Project.Create)
			}

			// Folder routes
			folders := protected.Group("/folders")
			{
				folders.GET("/my", h.Folder.ListByUser)
				folders.GET("/:id", h.Folder.Get)
				folders.PUT("/:id", h.Folder.Update)
				folders.DELETE("/:id", h.Folder.Delete)
				folders.PATCH("/:id/visibility", h.Folder.UpdateVisibility)
			}

			// Project routes
			projects := protected.Group("/projects")
			{
				projects.GET("/:id", h.Project.Get)
				projects.PUT("/:id", h.Project.Update)
				projects.DELETE("/:id", h.Project.Delete)

				// Invitations
				projects.POST("/:id/invitations", invitationHandler.CreateProjectInvitation)
				projects.GET("/:id/invitations", invitationHandler.GetProjectInvitations)

				// Tasks
				projects.GET("/:id/tasks", h.Task.ListByProject)
				projects.POST("/:id/tasks", h.Task.Create)

				// Labels
				projects.GET("/:id/labels", h.Label.ListByProject)
				projects.POST("/:id/labels", h.Label.Create)

				// Activities
				projects.GET("/:id/activities", activityHandler.GetProjectActivities)
			}

			// Task routes
			tasks := protected.Group("/tasks")
			{
				// Core CRUD
				tasks.GET("/:id", h.Task.Get)
				tasks.PUT("/:id", h.Task.Update)
				tasks.DELETE("/:id", h.Task.Delete)

				// Listing
				tasks.GET("/my", h.Task.ListMyTasks)
				tasks.GET("/status/:projectId", h.Task.ListByStatus)
				tasks.GET("/:id/subtasks", h.Task.ListSubtasks)

				// Status & Priority
				tasks.PATCH("/:id/status", h.Task.UpdateStatus)
				tasks.PATCH("/:id/priority", h.Task.UpdatePriority)

				// Assignment
				tasks.POST("/:id/assign", h.Task.AssignTask)
				tasks.DELETE("/:id/assign/:assigneeId", h.Task.UnassignTask)

				// Watchers
				tasks.POST("/:id/watchers", h.Task.AddWatcher)
				tasks.DELETE("/:id/watchers/:watcherId", h.Task.RemoveWatcher)

				// Sprint & hierarchy
				tasks.POST("/:id/move-sprint", h.Task.MoveToSprint)
				tasks.POST("/:id/convert-subtask", h.Task.ConvertToSubtask)

				// Completion
				tasks.POST("/:id/complete", h.Task.MarkComplete)

				// Comments
				tasks.GET("/:id/comments", h.Task.ListComments)
				tasks.POST("/:id/comments", h.Task.AddComment)
				tasks.PUT("/comments/:commentId", h.Task.UpdateComment)
				tasks.DELETE("/comments/:commentId", h.Task.DeleteComment)

				// Attachments
				tasks.GET("/:id/attachments", h.Task.ListAttachments)
				tasks.POST("/:id/attachments", h.Task.AddAttachment)
				tasks.DELETE("/attachments/:attachmentId", h.Task.DeleteAttachment)

				// Time tracking
				tasks.POST("/:id/timer/start", h.Task.StartTimer)
				tasks.POST("/timer/stop", h.Task.StopTimer)
				tasks.GET("/timer/active", h.Task.GetActiveTimer)
				tasks.POST("/:id/time", h.Task.LogTime)
				tasks.GET("/:id/time", h.Task.GetTimeEntries)
				tasks.GET("/:id/time/total", h.Task.GetTotalTime)

				// Dependencies
				tasks.GET("/:id/dependencies", h.Task.ListDependencies)
				tasks.GET("/:id/blocked-by", h.Task.ListBlockedBy)
				tasks.POST("/:id/dependencies", h.Task.AddDependency)
				tasks.DELETE("/:id/dependencies/:dependsOnTaskId", h.Task.RemoveDependency)

				// Checklists
				tasks.GET("/:id/checklists", h.Task.ListChecklists)
				tasks.POST("/:id/checklists", h.Task.CreateChecklist)
				tasks.POST("/checklists/:checklistId/items", h.Task.AddChecklistItem)
				tasks.PATCH("/checklists/items/:itemId", h.Task.ToggleChecklistItem)
				tasks.DELETE("/checklists/items/:itemId", h.Task.DeleteChecklistItem)

				// Activity
				tasks.GET("/:id/activity", h.Task.GetActivity)

				// Advanced filtering
				tasks.POST("/filter", h.Task.FilterTasks)

				// Bulk operations
				tasks.POST("/bulk/status", h.Task.BulkUpdateStatus)
				tasks.POST("/bulk/assign", h.Task.BulkAssign)
				tasks.POST("/bulk/move-sprint", h.Task.BulkMoveToSprint)
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

			// Team routes
			teams := protected.Group("/teams")
			{
				teams.POST("", teamHandler.Create)
				teams.GET("/:id", teamHandler.Get)
				teams.PUT("/:id", teamHandler.Update)
				teams.DELETE("/:id", teamHandler.Delete)
				teams.GET("/:id/members", teamHandler.ListMembers)
				teams.POST("/:id/members", teamHandler.AddMember)
				teams.PUT("/:id/members/:userId", teamHandler.UpdateMemberRole)
				teams.DELETE("/:id/members/:userId", teamHandler.RemoveMember)
			}

			// Chat routes
			chat := protected.Group("/chat")
			{
				chat.GET("/channels", chatHandler.ListChannels)
				chat.POST("/channels", chatHandler.CreateChannel)
				chat.GET("/channels/find", chatHandler.GetChannelByTarget)
				chat.GET("/channels/:id", chatHandler.GetChannel)
				chat.DELETE("/channels/:id", chatHandler.DeleteChannel)

				chat.POST("/channels/:id/join", chatHandler.JoinChannel)
				chat.POST("/channels/:id/leave", chatHandler.LeaveChannel)
				chat.GET("/channels/:id/members", chatHandler.GetChannelMembers)
				chat.POST("/channels/:id/members/add", chatHandler.AddMember)
				chat.POST("/channels/:id/read", chatHandler.MarkAsRead)
				chat.GET("/channels/:id/unread", chatHandler.GetUnreadCount)

				chat.GET("/channels/:id/messages", chatHandler.GetMessages)
				chat.POST("/channels/:id/messages", chatHandler.SendMessage)
				chat.GET("/messages/:messageId/thread", chatHandler.GetThreadMessages)
				chat.PUT("/messages/:messageId", chatHandler.UpdateMessage)
				chat.DELETE("/messages/:messageId", chatHandler.DeleteMessage)

				chat.POST("/messages/:messageId/reactions", chatHandler.AddReaction)
				chat.DELETE("/messages/:messageId/reactions", chatHandler.RemoveReaction)
				chat.GET("/messages/:messageId/reactions", chatHandler.GetReactions)

				chat.POST("/direct", chatHandler.CreateDirectChannel)
				chat.GET("/unread", chatHandler.GetAllUnreadCounts)
			}

			// Invitation routes (protected)
			invitations := protected.Group("/invitations")
			{
				invitations.GET("/pending", invitationHandler.GetMyInvitations)
				invitations.POST("/accept/:token", invitationHandler.AcceptInvitation)
				invitations.POST("/accept-link", invitationHandler.AcceptInvitationByLink)
				invitations.POST("/resend/:id", invitationHandler.ResendInvitation)
				invitations.DELETE("/:id", invitationHandler.CancelInvitation)
				invitations.POST("/link", invitationHandler.CreateLinkInvitation)
				invitations.GET("/stats", invitationHandler.GetInvitationStats)
			}

			// Activity routes
			activities := protected.Group("/activities")
			{
				activities.GET("/me", activityHandler.GetMyActivities)
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