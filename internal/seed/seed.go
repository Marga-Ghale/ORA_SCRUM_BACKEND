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

	// ============================================
	// CREATE USERS
	// ============================================
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

	// ============================================
	// CREATE WORKSPACE
	// ============================================
	defaultVisibility := "private"
	workspace := &repository.Workspace{
		Name:       "My Workspace",
		OwnerID:    user1.ID,
		Visibility: &defaultVisibility,
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

	// ============================================
	// CREATE SPACE (under workspace)
	// ============================================
	space := &repository.Space{
		Name:        "Engineering",
		WorkspaceID: workspace.ID, // ✅ Must have workspace
		OwnerID:     user1.ID,     // ✅ Must have owner
		Visibility:  &defaultVisibility,
	}
	repos.SpaceRepo.Create(ctx, space)

	// Add space members
	repos.SpaceRepo.AddMember(ctx, &repository.SpaceMember{
		SpaceID: space.ID,
		UserID:  user1.ID,
		Role:    "owner",
	})
	repos.SpaceRepo.AddMember(ctx, &repository.SpaceMember{
		SpaceID: space.ID,
		UserID:  user2.ID,
		Role:    "admin",
	})
	repos.SpaceRepo.AddMember(ctx, &repository.SpaceMember{
		SpaceID: space.ID,
		UserID:  user3.ID,
		Role:    "member",
	})

	// ============================================
	// CREATE FOLDER (optional, under space)
	// ============================================
	folder := &repository.Folder{
		Name:       "Backend Projects",
		SpaceID:    space.ID, // ✅ Must have space
		OwnerID:    user1.ID, // ✅ Must have owner
		Visibility: &defaultVisibility,
	}
	repos.FolderRepo.Create(ctx, folder)

	// Add folder members
	repos.FolderRepo.AddMember(ctx, &repository.FolderMember{
		FolderID: folder.ID,
		UserID:   user1.ID,
		Role:     "owner",
	})
	repos.FolderRepo.AddMember(ctx, &repository.FolderMember{
		FolderID: folder.ID,
		UserID:   user2.ID,
		Role:     "member",
	})

	// ============================================
	// CREATE PROJECT (under space, optionally in folder)
	// ============================================
	project := &repository.Project{
		Name:       "ORA Scrum",
		Key:        "ORA",
		SpaceID:    space.ID,   // ✅ Must have space
		FolderID:   &folder.ID, // ✅ Optional folder (can be nil)
		LeadID:     &user1.ID,
		Visibility: &defaultVisibility,
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

	// ============================================
	// CREATE PROJECT WITHOUT FOLDER (direct in space)
	// ============================================
	project2 := &repository.Project{
		Name:       "Mobile App",
		Key:        "MOB",
		SpaceID:    space.ID, // ✅ Must have space
		FolderID:   nil,      // ✅ No folder - direct in space
		LeadID:     &user2.ID,
		Visibility: &defaultVisibility,
	}
	repos.ProjectRepo.Create(ctx, project2)

	// Add project2 members
	repos.ProjectRepo.AddMember(ctx, &repository.ProjectMember{
		ProjectID: project2.ID,
		UserID:    user2.ID,
		Role:      "lead",
	})
	repos.ProjectRepo.AddMember(ctx, &repository.ProjectMember{
		ProjectID: project2.ID,
		UserID:    user1.ID,
		Role:      "member",
	})

	// ============================================
	// CREATE LABELS
	// ============================================
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

	// ============================================
	// CREATE SPRINT
	// ============================================
	now := time.Now()
	sprintStart := now.AddDate(0, 0, -7)
	sprintEnd := now.AddDate(0, 0, 7)

	sprint := &repository.Sprint{
		Name:      "Sprint 1",
		ProjectID: project.ID,
		Status:    "active",
		StartDate: sprintStart,
		EndDate:   sprintEnd,
	}
	repos.SprintRepo.Create(ctx, sprint)

	// ============================================
	// CREATE TASKS
	// ============================================
	tasks := []struct {
		Title       string
		Status      string
		Priority    string
		AssigneeIDs []string
		SprintID    *string
	}{
		{"Setup project structure", "done", "high", []string{user1.ID}, &sprint.ID},
		{"Implement authentication", "done", "urgent", []string{user1.ID}, &sprint.ID},
		{"Create dashboard UI", "in_progress", "high", []string{user2.ID}, &sprint.ID},
		{"API integration", "in_progress", "medium", []string{user2.ID}, &sprint.ID},
		{"Fix login bug", "todo", "urgent", []string{user1.ID}, &sprint.ID},
		{"Add dark mode", "todo", "low", []string{user3.ID}, &sprint.ID},
		{"Write documentation", "backlog", "low", []string{}, nil},
		{"Performance optimization", "backlog", "medium", []string{}, nil},
	}

	for i, t := range tasks {
		task := &repository.Task{
			Title:       t.Title,
			Status:      t.Status,
			Priority:    t.Priority,
			ProjectID:   project.ID,
			SprintID:    t.SprintID,
			AssigneeIDs: t.AssigneeIDs,
			LabelIDs:    []string{},
			CreatedBy:   &user1.ID,
			Position:    i,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		repos.TaskRepo.Create(ctx, task)
	}

	// ============================================
	// CREATE SAMPLE NOTIFICATIONS
	// ============================================
	seedNotifications(ctx, repos, user1.ID, user2.ID, user3.ID, project.ID, workspace.ID, sprint.ID)

	log.Println("[Seed] ✅ Initial data created successfully!")
	log.Printf("[Seed] Created hierarchy: Workspace → Space → Folder → Project")
	log.Printf("[Seed] Workspace ID: %s", workspace.ID)
	log.Printf("[Seed] Space ID: %s", space.ID)
	log.Printf("[Seed] Folder ID: %s", folder.ID)
	log.Printf("[Seed] Project 1 ID (in folder): %s", project.ID)
	log.Printf("[Seed] Project 2 ID (direct in space): %s", project2.ID)
}

// seedNotifications creates sample notifications for testing
func seedNotifications(ctx context.Context, repos *repository.Repositories, user1ID, user2ID, user3ID, projectID, workspaceID, sprintID string) {
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
			Data:      map[string]interface{}{"workspaceId": workspaceID},
			CreatedAt: now.Add(-10 * 24 * time.Hour),
		},
	}

	for _, n := range notifications {
		notif := n // Create a copy to avoid pointer issues
		repos.NotificationRepo.Create(ctx, &notif)
	}

	log.Printf("[Seed] Created %d sample notifications", len(notifications))
}