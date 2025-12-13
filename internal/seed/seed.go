// internal/seed/seed.go
package seed

import (
	"context"
	"log"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func SeedData(repos *repository.Repositories) {
	ctx := context.Background()

	// Check if data already exists
	users, _ := repos.UserRepo.FindAll(ctx)
	if len(users) > 0 {
		log.Println("[Seed] Data already exists, skipping...")
		return
	}

	log.Println("[Seed] Creating initial data...")

	// Create users
	password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user1 := &repository.User{
		Email:    "test@example.com",
		Password: string(password),
		Name:     "Test User",
		Status:   "online",
	}
	repos.UserRepo.Create(ctx, user1)

	user2 := &repository.User{
		Email:    "dev@example.com",
		Password: string(password),
		Name:     "Dev User",
		Status:   "online",
	}
	repos.UserRepo.Create(ctx, user2)

	user3 := &repository.User{
		Email:    "admin@example.com",
		Password: string(password),
		Name:     "Admin User",
		Status:   "away",
	}
	repos.UserRepo.Create(ctx, user3)

	// Create workspace
	workspace := &repository.Workspace{
		Name:    "My Workspace",
		OwnerID: user1.ID,
	}
	repos.WorkspaceRepo.Create(ctx, workspace)

	// Add members to workspace
	repos.WorkspaceRepo.AddMember(ctx, &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      user1.ID,
		Role:        "owner",
	})
	repos.WorkspaceRepo.AddMember(ctx, &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      user2.ID,
		Role:        "admin",
	})
	repos.WorkspaceRepo.AddMember(ctx, &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      user3.ID,
		Role:        "member",
	})

	// Create space
	space := &repository.Space{
		Name:        "Engineering",
		WorkspaceID: workspace.ID,
	}
	repos.SpaceRepo.Create(ctx, space)

	// Create project
	project := &repository.Project{
		Name:    "ORA Scrum",
		Key:     "ORA",
		SpaceID: space.ID,
		LeadID:  &user1.ID,
	}
	repos.ProjectRepo.Create(ctx, project)

	// Add project members
	repos.ProjectRepo.AddMember(ctx, &repository.ProjectMember{
		ProjectID: project.ID,
		UserID:    user1.ID,
		Role:      "lead",
	})
	repos.ProjectRepo.AddMember(ctx, &repository.ProjectMember{
		ProjectID: project.ID,
		UserID:    user2.ID,
		Role:      "member",
	})
	repos.ProjectRepo.AddMember(ctx, &repository.ProjectMember{
		ProjectID: project.ID,
		UserID:    user3.ID,
		Role:      "member",
	})

	// Create labels
	labels := []struct {
		Name  string
		Color string
	}{
		{"frontend", "#3B82F6"},
		{"backend", "#10B981"},
		{"bug", "#EF4444"},
		{"feature", "#8B5CF6"},
		{"documentation", "#F59E0B"},
	}

	for _, l := range labels {
		repos.LabelRepo.Create(ctx, &repository.Label{
			Name:      l.Name,
			Color:     l.Color,
			ProjectID: project.ID,
		})
	}

	// Create sprint
	now := time.Now()
	sprintStart := now.AddDate(0, 0, -7)
	sprintEnd := now.AddDate(0, 0, 7)
	sprint := &repository.Sprint{
		Name:      "Sprint 1",
		ProjectID: project.ID,
		Status:    "active",
		StartDate: &sprintStart,
		EndDate:   &sprintEnd,
	}
	repos.SprintRepo.Create(ctx, sprint)

	// Create tasks
	tasks := []struct {
		Title      string
		Status     string
		Priority   string
		Type       string
		AssigneeID *string
		SprintID   *string
	}{
		{"Setup project structure", "done", "high", "task", &user1.ID, &sprint.ID},
		{"Implement authentication", "done", "urgent", "task", &user1.ID, &sprint.ID},
		{"Create dashboard UI", "in_progress", "high", "task", &user2.ID, &sprint.ID},
		{"API integration", "in_progress", "medium", "task", &user2.ID, &sprint.ID},
		{"Fix login bug", "todo", "urgent", "bug", &user1.ID, &sprint.ID},
		{"Add dark mode", "todo", "low", "feature", &user3.ID, &sprint.ID},
		{"Write documentation", "backlog", "low", "task", nil, nil},
		{"Performance optimization", "backlog", "medium", "task", nil, nil},
	}

	for i, t := range tasks {
		task := &repository.Task{
			Key:        "ORA-" + string(rune('1'+i)),
			Title:      t.Title,
			Status:     t.Status,
			Priority:   t.Priority,
			Type:       t.Type,
			ProjectID:  project.ID,
			SprintID:   t.SprintID,
			AssigneeID: t.AssigneeID,
			ReporterID: user1.ID,
			OrderIndex: i,
			Labels:     []string{},
		}
		repos.TaskRepo.Create(ctx, task)
	}

	// ============================================
	// CREATE SAMPLE NOTIFICATIONS
	// ============================================
	seedNotifications(ctx, repos, user1.ID, user2.ID, user3.ID, project.ID, sprint.ID)

	log.Println("[Seed] ✅ Initial data created successfully!")
}

// seedNotifications creates sample notifications for testing
func seedNotifications(ctx context.Context, repos *repository.Repositories, user1ID, user2ID, user3ID, projectID, sprintID string) {
	now := time.Now()

	notifications := []repository.Notification{
		// Task notifications for user1 (test@example.com)
		{
			UserID:    user1ID,
			Type:      "TASK_ASSIGNED",
			Title:     "Task Assigned",
			Message:   "You have been assigned to task: Create dashboard UI",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-1", "projectId": projectID, "taskKey": "ORA-3"},
			CreatedAt: now.Add(-5 * time.Minute),
		},
		{
			UserID:    user1ID,
			Type:      "TASK_COMMENTED",
			Title:     "New Comment",
			Message:   "Dev User commented on task: Setup project structure",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-2", "projectId": projectID, "taskKey": "ORA-1"},
			CreatedAt: now.Add(-15 * time.Minute),
		},
		{
			UserID:    user1ID,
			Type:      "TASK_STATUS_CHANGED",
			Title:     "Task Status Changed",
			Message:   "Task 'API integration' moved from To Do to In Progress",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-3", "projectId": projectID, "taskKey": "ORA-4", "oldStatus": "todo", "newStatus": "in_progress"},
			CreatedAt: now.Add(-1 * time.Hour),
		},
		{
			UserID:    user1ID,
			Type:      "SPRINT_ENDING",
			Title:     "Sprint Ending Soon",
			Message:   "Sprint 'Sprint 1' ends in 7 days",
			Read:      false,
			Data:      map[string]interface{}{"sprintId": sprintID, "projectId": projectID, "daysRemaining": 7},
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			UserID:    user1ID,
			Type:      "MENTION",
			Title:     "You were mentioned",
			Message:   "Dev User mentioned you in task: Fix login bug",
			Read:      true,
			Data:      map[string]interface{}{"taskId": "task-4", "projectId": projectID, "taskKey": "ORA-5"},
			CreatedAt: now.Add(-1 * 24 * time.Hour),
		},
		{
			UserID:    user1ID,
			Type:      "TASK_DUE_SOON",
			Title:     "Task Due Tomorrow",
			Message:   "Task 'Create dashboard UI' is due tomorrow",
			Read:      true,
			Data:      map[string]interface{}{"taskId": "task-5", "projectId": projectID, "taskKey": "ORA-3", "daysUntilDue": 1},
			CreatedAt: now.Add(-2 * 24 * time.Hour),
		},
		{
			UserID:    user1ID,
			Type:      "PROJECT_INVITATION",
			Title:     "Project Invitation",
			Message:   "Admin User added you to project: Mobile App",
			Read:      true,
			Data:      map[string]interface{}{"projectId": "project-2"},
			CreatedAt: now.Add(-5 * 24 * time.Hour),
		},

		// Notifications for user2 (dev@example.com)
		{
			UserID:    user2ID,
			Type:      "TASK_ASSIGNED",
			Title:     "Task Assigned",
			Message:   "You have been assigned to task: API integration",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-6", "projectId": projectID, "taskKey": "ORA-4"},
			CreatedAt: now.Add(-30 * time.Minute),
		},
		{
			UserID:    user2ID,
			Type:      "TASK_CREATED",
			Title:     "New Task Created",
			Message:   "New task created: Performance optimization (ORA-8)",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-7", "projectId": projectID, "taskKey": "ORA-8"},
			CreatedAt: now.Add(-45 * time.Minute),
		},
		{
			UserID:    user2ID,
			Type:      "SPRINT_STARTED",
			Title:     "Sprint Started",
			Message:   "Sprint 'Sprint 1' has started! Time to get to work.",
			Read:      true,
			Data:      map[string]interface{}{"sprintId": sprintID, "projectId": projectID},
			CreatedAt: now.Add(-7 * 24 * time.Hour),
		},

		// Notifications for user3 (admin@example.com)
		{
			UserID:    user3ID,
			Type:      "TASK_ASSIGNED",
			Title:     "Task Assigned",
			Message:   "You have been assigned to task: Add dark mode",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-8", "projectId": projectID, "taskKey": "ORA-6"},
			CreatedAt: now.Add(-1 * time.Hour),
		},
		{
			UserID:    user3ID,
			Type:      "WORKSPACE_INVITATION",
			Title:     "Workspace Invitation",
			Message:   "Test User added you to workspace: My Workspace",
			Read:      true,
			Data:      map[string]interface{}{"workspaceId": "workspace-1"},
			CreatedAt: now.Add(-10 * 24 * time.Hour),
		},
	}

	for _, n := range notifications {
		notif := n // Create a copy to avoid pointer issues
		repos.NotificationRepo.Create(ctx, &notif)
	}

	log.Printf("[Seed] Created %d sample notifications", len(notifications))
}
