// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	// Initialize Repositories
	// ============================================
	var repos *repository.Repositories
	var pool *pgxpool.Pool

	// Try to connect to PostgreSQL
	postgresDB, err := db.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Printf("⚠️  Failed to connect to PostgreSQL: %v", err)
		log.Println("📦 Falling back to in-memory storage")
		repos = repository.NewRepositories() // In-memory fallback
	} else {
		defer postgresDB.Close()
		pool = postgresDB.Pool
		repos = repository.NewPgRepositories(pool)
		log.Println("📦 Using PostgreSQL storage")
	}

	// Try to connect to Redis (optional)
	redisDB, err := db.NewRedisDB(cfg.RedisURL)
	if err != nil {
		log.Printf("⚠️  Failed to connect to Redis: %v (continuing without cache)", err)
	} else {
		defer redisDB.Close()
		log.Println("⚡ Redis cache enabled")
	}

	// ============================================
	// Seed Data (for development)
	// ============================================
	if cfg.Environment == "development" {
		seed.SeedData(repos)
	}

	// ============================================
	// Initialize WebSocket Hub
	// ============================================
	socketHub := socket.NewHub()
	go socketHub.Run()
	log.Println("🔌 WebSocket hub started")

	// Create broadcaster for real-time events
	broadcaster := socket.NewBroadcaster(socketHub)

	// ============================================
	// Initialize Email Service
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
		log.Println("📧 Email service configured")
	} else {
		log.Println("📧 Email service not configured (SMTP_HOST not set)")
	}

	// ============================================
	// Initialize Additional Repositories (for new features)
	// ============================================
	var teamRepo repository.TeamRepository
	var invitationRepo repository.InvitationRepository
	var activityRepo repository.ActivityRepository
	var taskWatcherRepo repository.TaskWatcherRepository

	if pool != nil {
		teamRepo = repository.NewPgTeamRepository(pool)
		invitationRepo = repository.NewPgInvitationRepository(pool)
		activityRepo = repository.NewPgActivityRepository(pool)
		taskWatcherRepo = repository.NewPgTaskWatcherRepository(pool)
		log.Println("📦 Extended repositories initialized")
	}

	// Initialize notification service
	notificationSvc := notification.NewServiceWithRepos(
		repos.NotificationRepo,
		repos.UserRepo,
		repos.ProjectRepo,
	)

	// Initialize services
	services := service.NewServices(cfg, repos, notificationSvc)

	// ============================================
	// Initialize New Services (Teams, Invitations, etc.)
	// ============================================
	var teamSvc service.TeamService
	var invitationSvc service.InvitationService
	var activitySvc service.ActivityService
	var taskWatcherSvc service.TaskWatcherService

	if teamRepo != nil {
		teamSvc = service.NewTeamService(
			teamRepo,
			repos.UserRepo,
			repos.WorkspaceRepo,
			notificationSvc,
			emailSvc,
			broadcaster,
		)
		log.Println("👥 Team service initialized")
	}

	if invitationRepo != nil {
		invitationSvc = service.NewInvitationService(
			invitationRepo,
			repos.WorkspaceRepo,
			teamRepo,
			repos.ProjectRepo,
			repos.UserRepo,
			emailSvc,
		)
		log.Println("📨 Invitation service initialized")
	}

	if activityRepo != nil {
		activitySvc = service.NewActivityService(activityRepo)
		log.Println("📝 Activity service initialized")
	}

	if taskWatcherRepo != nil {
		taskWatcherSvc = service.NewTaskWatcherService(taskWatcherRepo)
		log.Println("👀 Task watcher service initialized")
	}

	// Initialize handlers
	h := handlers.NewHandlers(services)

	// Initialize new handlers
	var teamHandler *handlers.TeamHandler
	var invitationHandler *handlers.InvitationHandler
	var activityHandler *handlers.ActivityHandler
	var taskWatcherHandler *handlers.TaskWatcherHandler

	if teamSvc != nil {
		teamHandler = handlers.NewTeamHandler(teamSvc)
	}
	if invitationSvc != nil {
		invitationHandler = handlers.NewInvitationHandler(invitationSvc)
	}
	if activitySvc != nil {
		activityHandler = handlers.NewActivityHandler(activitySvc)
	}
	if taskWatcherSvc != nil {
		taskWatcherHandler = handlers.NewTaskWatcherHandler(taskWatcherSvc)
	}

	// Initialize WebSocket handler
	socketHandler := socket.NewHandler(socketHub)

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

	// Create Gin router
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
		status := gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"features": gin.H{
				"websocket": true,
				"teams":     teamRepo != nil,
				"email":     emailSvc != nil,
			},
		}
		if postgresDB != nil {
			status["database"] = "connected"
		} else {
			status["database"] = "in-memory"
		}
		if redisDB != nil {
			status["cache"] = "connected"
		} else {
			status["cache"] = "disabled"
		}
		c.JSON(http.StatusOK, status)
	})

	// ============================================
	// WebSocket Route
	// ============================================
	r.GET("/ws", func(c *gin.Context) {
		// Support token in query string for WebSocket
		token := c.Query("token")
		if token == "" {
			// Try Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			return
		}

		// Validate token
		jwtToken, err := services.Auth.ValidateToken(token)
		if err != nil || !jwtToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		userID, err := services.Auth.GetUserIDFromToken(jwtToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("userID", userID)
		socketHandler.HandleWebSocket(c)
	})

	// API routes
	api := r.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Auth.Register)
			auth.POST("/login", h.Auth.Login)
			auth.POST("/refresh", h.Auth.RefreshToken)
			auth.POST("/logout", h.Auth.Logout)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(services.Auth))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", h.User.GetCurrentUser)
				users.PUT("/me", h.User.UpdateCurrentUser)
				users.GET("/me/teams", func(c *gin.Context) {
					if teamHandler != nil {
						teamHandler.ListMyTeams(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})
				users.GET("/me/activities", func(c *gin.Context) {
					if activityHandler != nil {
						activityHandler.GetMyActivities(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "activities not available"})
					}
				})
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

				// Team routes within workspace
				workspaces.GET("/:id/teams", func(c *gin.Context) {
					if teamHandler != nil {
						teamHandler.ListByWorkspace(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})
				workspaces.POST("/:id/teams", func(c *gin.Context) {
					if teamHandler != nil {
						c.Params = append(c.Params, gin.Param{Key: "workspaceId", Value: c.Param("id")})
						teamHandler.Create(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})

				// Invitation routes within workspace
				workspaces.POST("/:id/invitations", func(c *gin.Context) {
					if invitationHandler != nil {
						invitationHandler.CreateWorkspaceInvitation(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "invitations not available"})
					}
				})
			}

			// Team routes
			teams := protected.Group("/teams")
			{
				teams.GET("/:id", func(c *gin.Context) {
					if teamHandler != nil {
						teamHandler.Get(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})
				teams.PUT("/:id", func(c *gin.Context) {
					if teamHandler != nil {
						teamHandler.Update(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})
				teams.DELETE("/:id", func(c *gin.Context) {
					if teamHandler != nil {
						teamHandler.Delete(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})
				teams.GET("/:id/members", func(c *gin.Context) {
					if teamHandler != nil {
						teamHandler.ListMembers(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})
				teams.POST("/:id/members", func(c *gin.Context) {
					if teamHandler != nil {
						teamHandler.AddMember(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})
				teams.PUT("/:id/members/:userId", func(c *gin.Context) {
					if teamHandler != nil {
						teamHandler.UpdateMemberRole(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})
				teams.DELETE("/:id/members/:userId", func(c *gin.Context) {
					if teamHandler != nil {
						teamHandler.RemoveMember(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "teams not available"})
					}
				})
			}

			// Invitation routes
			invitations := protected.Group("/invitations")
			{
				invitations.GET("", func(c *gin.Context) {
					if invitationHandler != nil {
						invitationHandler.GetMyInvitations(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "invitations not available"})
					}
				})
				invitations.POST("/:token/accept", func(c *gin.Context) {
					if invitationHandler != nil {
						invitationHandler.AcceptInvitation(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "invitations not available"})
					}
				})
				invitations.DELETE("/:id", func(c *gin.Context) {
					if invitationHandler != nil {
						invitationHandler.CancelInvitation(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "invitations not available"})
					}
				})
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

				// Activity routes for project
				projects.GET("/:id/activities", func(c *gin.Context) {
					if activityHandler != nil {
						activityHandler.GetProjectActivities(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "activities not available"})
					}
				})

				// Invitation routes within project
				projects.POST("/:id/invitations", func(c *gin.Context) {
					if invitationHandler != nil {
						invitationHandler.CreateProjectInvitation(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "invitations not available"})
					}
				})
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

				// Task watcher routes
				tasks.POST("/:id/watch", func(c *gin.Context) {
					if taskWatcherHandler != nil {
						taskWatcherHandler.WatchTask(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "watchers not available"})
					}
				})
				tasks.DELETE("/:id/watch", func(c *gin.Context) {
					if taskWatcherHandler != nil {
						taskWatcherHandler.UnwatchTask(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "watchers not available"})
					}
				})
				tasks.GET("/:id/watchers", func(c *gin.Context) {
					if taskWatcherHandler != nil {
						taskWatcherHandler.GetWatchers(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "watchers not available"})
					}
				})
				tasks.GET("/:id/watching", func(c *gin.Context) {
					if taskWatcherHandler != nil {
						taskWatcherHandler.IsWatching(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "watchers not available"})
					}
				})

				// Activity routes for task
				tasks.GET("/:id/activities", func(c *gin.Context) {
					if activityHandler != nil {
						activityHandler.GetTaskActivities(c)
					} else {
						c.JSON(http.StatusNotImplemented, gin.H{"error": "activities not available"})
					}
				})
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

			// Online users endpoint
			protected.GET("/online-users", func(c *gin.Context) {
				users := socketHub.GetOnlineUsers()
				c.JSON(http.StatusOK, gin.H{"users": users, "count": len(users)})
			})
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
		log.Printf("🔌 WebSocket available at ws://localhost:%s/ws", cfg.Port)
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
