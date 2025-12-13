// internal/seed/seed.go
package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// SeedData creates initial test data for development
func SeedData(repos *repository.Repositories) {
	ctx := context.Background()

	// Check if test user already exists
	existingUser, _ := repos.UserRepo.FindByEmail(ctx, "test@example.com")
	if existingUser != nil {
		log.Println("[Seed] Data already exists, skipping...")
		return
	}

	log.Println("[Seed] Creating initial data...")

	// 1. Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[Seed] Failed to hash password: %v", err)
		return
	}

	testUser := &repository.User{
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Name:     "Test User",
		Status:   "online", // lowercase
	}
	if err := repos.UserRepo.Create(ctx, testUser); err != nil {
		log.Printf("[Seed] Failed to create test user: %v", err)
		return
	}
	log.Printf("[Seed] Created user: %s (ID: %s)", testUser.Email, testUser.ID)

	// Create a second user for assignment testing
	hashedPassword2, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	devUser := &repository.User{
		Email:    "dev@example.com",
		Password: string(hashedPassword2),
		Name:     "Dev User",
		Status:   "online", // lowercase
	}
	if err := repos.UserRepo.Create(ctx, devUser); err != nil {
		log.Printf("[Seed] Failed to create dev user: %v", err)
	} else {
		log.Printf("[Seed] Created user: %s (ID: %s)", devUser.Email, devUser.ID)
	}

	// 2. Create default workspace
	workspace := &repository.Workspace{
		Name:    "My Workspace",
		OwnerID: testUser.ID,
	}
	if err := repos.WorkspaceRepo.Create(ctx, workspace); err != nil {
		log.Printf("[Seed] Failed to create workspace: %v", err)
		return
	}
	log.Printf("[Seed] Created workspace: %s (ID: %s)", workspace.Name, workspace.ID)

	// 3. Add users as workspace members
	workspaceMember := &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      testUser.ID,
		Role:        "owner", // lowercase
	}
	if err := repos.WorkspaceRepo.AddMember(ctx, workspaceMember); err != nil {
		log.Printf("[Seed] Failed to add workspace member: %v", err)
	}

	// Add dev user as member
	devMember := &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      devUser.ID,
		Role:        "member", // lowercase
	}
	repos.WorkspaceRepo.AddMember(ctx, devMember)

	// 4. Create default space
	space := &repository.Space{
		Name:        "Engineering",
		WorkspaceID: workspace.ID,
	}
	if err := repos.SpaceRepo.Create(ctx, space); err != nil {
		log.Printf("[Seed] Failed to create space: %v", err)
		return
	}
	log.Printf("[Seed] Created space: %s (ID: %s)", space.Name, space.ID)

	// 5. Create default project
	project := &repository.Project{
		Name:    "My Project",
		Key:     "PRJ",
		SpaceID: space.ID,
		LeadID:  &testUser.ID,
	}
	if err := repos.ProjectRepo.Create(ctx, project); err != nil {
		log.Printf("[Seed] Failed to create project: %v", err)
		return
	}
	log.Printf("[Seed] Created project: %s (ID: %s, Key: %s)", project.Name, project.ID, project.Key)

	// Add users as project members
	projectMember := &repository.ProjectMember{
		ProjectID: project.ID,
		UserID:    testUser.ID,
		Role:      "lead", // lowercase
	}
	repos.ProjectRepo.AddMember(ctx, projectMember)

	devProjectMember := &repository.ProjectMember{
		ProjectID: project.ID,
		UserID:    devUser.ID,
		Role:      "member", // lowercase
	}
	repos.ProjectRepo.AddMember(ctx, devProjectMember)

	// 6. Create sample labels
	labels := []struct {
		name  string
		color string
	}{
		{"frontend", "#3B82F6"},
		{"backend", "#10B981"},
		{"bug", "#EF4444"},
		{"feature", "#8B5CF6"},
		{"documentation", "#F59E0B"},
	}

	for _, l := range labels {
		label := &repository.Label{
			Name:      l.name,
			Color:     l.color,
			ProjectID: project.ID,
		}
		repos.LabelRepo.Create(ctx, label)
	}
	log.Printf("[Seed] Created %d labels", len(labels))

	// 7. Create sample tasks with LOWERCASE values
	sampleTasks := []struct {
		title       string
		description string
		status      string
		priority    string
		taskType    string
		assigneeID  *string
		labels      []string
	}{
		{
			title:       "Welcome to ORA Scrum!",
			description: "This is your first task. Get started by exploring the board.",
			status:      "todo",   // lowercase
			priority:    "medium", // lowercase
			taskType:    "task",   // lowercase
			assigneeID:  &testUser.ID,
			labels:      []string{"documentation"},
		},
		{
			title:       "Set up development environment",
			description: "Install Node.js, Go, and Docker for local development.",
			status:      "in_progress", // lowercase
			priority:    "high",        // lowercase
			taskType:    "task",        // lowercase
			assigneeID:  &devUser.ID,
			labels:      []string{"backend", "frontend"},
		},
		{
			title:       "Review project requirements",
			description: "Go through the PRD and ensure all requirements are captured.",
			status:      "done",   // lowercase
			priority:    "medium", // lowercase
			taskType:    "task",   // lowercase
			assigneeID:  &testUser.ID,
			labels:      []string{"documentation"},
		},
		{
			title:       "Implement user authentication",
			description: "Create login, register, and password reset functionality.",
			status:      "backlog", // lowercase
			priority:    "high",    // lowercase
			taskType:    "story",   // lowercase
			assigneeID:  nil,
			labels:      []string{"backend", "feature"},
		},
		{
			title:       "Fix login button not responding",
			description: "The login button sometimes doesn't respond on mobile devices.",
			status:      "todo",   // lowercase
			priority:    "urgent", // lowercase
			taskType:    "bug",    // lowercase
			assigneeID:  &devUser.ID,
			labels:      []string{"frontend", "bug"},
		},
		{
			title:       "Design system improvements",
			description: "Update color palette and typography for better accessibility.",
			status:      "in_review", // lowercase
			priority:    "low",       // lowercase
			taskType:    "story",     // lowercase
			assigneeID:  &testUser.ID,
			labels:      []string{"frontend", "feature"},
		},
		{
			title:       "API performance optimization",
			description: "Reduce response times for the task listing endpoint.",
			status:      "backlog", // lowercase
			priority:    "medium",  // lowercase
			taskType:    "task",    // lowercase
			assigneeID:  nil,
			labels:      []string{"backend"},
		},
		{
			title:       "Mobile app epic",
			description: "Plan and implement the mobile application for iOS and Android.",
			status:      "backlog", // lowercase
			priority:    "none",    // lowercase
			taskType:    "epic",    // lowercase
			assigneeID:  nil,
			labels:      []string{"feature"},
		},
	}

	for i, t := range sampleTasks {
		taskKey := fmt.Sprintf("%s-%d", project.Key, i+1)
		task := &repository.Task{
			Key:         taskKey,
			Title:       t.title,
			Description: &t.description,
			Status:      t.status,
			Priority:    t.priority,
			Type:        t.taskType,
			ProjectID:   project.ID,
			ReporterID:  testUser.ID,
			AssigneeID:  t.assigneeID,
			Labels:      t.labels,
			OrderIndex:  i,
		}
		if err := repos.TaskRepo.Create(ctx, task); err != nil {
			log.Printf("[Seed] Failed to create task %s: %v", taskKey, err)
		}
	}
	log.Printf("[Seed] Created %d tasks", len(sampleTasks))

	// 8. Create a sample sprint
	sprint := &repository.Sprint{
		Name:      "Sprint 1",
		ProjectID: project.ID,
		Status:    "planning", // lowercase
	}
	if err := repos.SprintRepo.Create(ctx, sprint); err != nil {
		log.Printf("[Seed] Failed to create sprint: %v", err)
	} else {
		log.Printf("[Seed] Created sprint: %s (ID: %s)", sprint.Name, sprint.ID)
	}

	log.Println("[Seed] ✅ Initial data created successfully!")
	log.Println("==========================================")
	log.Printf("  📧 Test User: test@example.com")
	log.Printf("  📧 Dev User:  dev@example.com")
	log.Printf("  🔑 Password:  password123 (for both)")
	log.Printf("  🏢 Workspace: %s", workspace.ID)
	log.Printf("  📁 Space:     %s", space.ID)
	log.Printf("  📋 Project:   %s (Key: %s)", project.ID, project.Key)
	log.Printf("  🏃 Sprint:    %s", sprint.ID)
	log.Println("==========================================")
}
