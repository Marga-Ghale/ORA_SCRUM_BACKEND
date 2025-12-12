// internal/seed/seed.go
package seed

import (
	"context"
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
		Status:   "ONLINE",
	}
	if err := repos.UserRepo.Create(ctx, testUser); err != nil {
		log.Printf("[Seed] Failed to create test user: %v", err)
		return
	}
	log.Printf("[Seed] Created user: %s (ID: %s)", testUser.Email, testUser.ID)

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

	// 3. Add user as workspace member (OWNER)
	workspaceMember := &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      testUser.ID,
		Role:        "OWNER",
	}
	if err := repos.WorkspaceRepo.AddMember(ctx, workspaceMember); err != nil {
		log.Printf("[Seed] Failed to add workspace member: %v", err)
	}

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
	}
	if err := repos.ProjectRepo.Create(ctx, project); err != nil {
		log.Printf("[Seed] Failed to create project: %v", err)
		return
	}
	log.Printf("[Seed] Created project: %s (ID: %s, Key: %s)", project.Name, project.ID, project.Key)

	// 6. Create sample tasks
	sampleTasks := []struct {
		title    string
		status   string
		priority string
		taskType string
	}{
		{"Welcome to ORA Scrum!", "TODO", "MEDIUM", "TASK"},
		{"Set up development environment", "IN_PROGRESS", "HIGH", "TASK"},
		{"Review project requirements", "DONE", "MEDIUM", "TASK"},
		{"Create user authentication", "BACKLOG", "HIGH", "STORY"},
		{"Fix login bug", "TODO", "HIGHEST", "BUG"},
	}

	for i, t := range sampleTasks {
		task := &repository.Task{
			Key:        project.Key + "-" + string(rune('1'+i)),
			Title:      t.title,
			Status:     t.status,
			Priority:   t.priority,
			Type:       t.taskType,
			ProjectID:  project.ID,
			ReporterID: testUser.ID,
			Labels:     []string{},
		}
		if err := repos.TaskRepo.Create(ctx, task); err != nil {
			log.Printf("[Seed] Failed to create task: %v", err)
		} else {
			// Fix the key after creation
			task.Key = project.Key + "-" + string(rune('1'+i))
			repos.TaskRepo.Update(ctx, task)
		}
	}

	log.Println("[Seed] ✅ Initial data created successfully!")
	log.Println("==========================================")
	log.Printf("  📧 Test User: test@example.com")
	log.Printf("  🔑 Password: password123")
	log.Printf("  🏢 Workspace ID: %s", workspace.ID)
	log.Printf("  📁 Space ID: %s", space.ID)
	log.Printf("  📋 Project ID: %s", project.ID)
	log.Println("==========================================")
}
