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
	// users, _ := repos.UserRepo.FindAll(ctx)
	// if len(users) > 0 {
	// 	log.Println("[Seed] Data already exists, skipping...")
	// 	return
	// }

	log.Println("[Seed] 🌱 Creating initial data with real scenarios...")

	// ============================================
	// CREATE USERS (4 real team members)
	// ============================================
	password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// 1. MARGA - Workspace Owner (Admin)
	marga := &repository.User{
		Email:    "marga.ghale@oratechnologies.io",
		Password: string(password),
		Name:     "Marga Ghale",
		Status:   "online",
	}
	repos.UserRepo.Create(ctx, marga)

	// 2. BIPIN - Full Workspace Member
	bipin := &repository.User{
		Email:    "bipin.dhimal@oratechnologies.io",
		Password: string(password),
		Name:     "Bipin Dhimal",
		Status:   "online",
	}
	repos.UserRepo.Create(ctx, bipin)

	// 3. KRITIM - Space-Only Member (Limited Access)
	kritim := &repository.User{
		Email:    "kritim.kafle@oratechnologies.io",
		Password: string(password),
		Name:     "Kritim Kafle",
		Status:   "away",
	}
	repos.UserRepo.Create(ctx, kritim)

	// 4. PRERAK - Project-Only Contractor
	prerak := &repository.User{
		Email:    "prerak.khadka@oratechnologies.io",
		Password: string(password),
		Name:     "Prerak Khadka",
		Status:   "online",
	}
	repos.UserRepo.Create(ctx, prerak)

	log.Printf("✅ Created 4 users: Marga (admin), Bipin (member), Kritim (space-only), Prerak (contractor)")

	// ============================================
	// SCENARIO 1: CREATE WORKSPACE
	// Marga creates "ORA Technologies" workspace
	// ============================================
	defaultVisibility := "private"
	workspace := &repository.Workspace{
		Name:        "ORA Technologies",
		Description: stringPtr("Main company workspace for all projects"),
		OwnerID:     marga.ID,
		Visibility:  &defaultVisibility,
	}
	repos.WorkspaceRepo.Create(ctx, workspace)

	// Add workspace members
	// Marga = OWNER (full control)
	repos.WorkspaceRepo.AddMember(ctx, &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      marga.ID,
		Role:        "owner",
	})

	// Bipin = MEMBER (can see everything, but limited permissions)
	repos.WorkspaceRepo.AddMember(ctx, &repository.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      bipin.ID,
		Role:        "member",
	})

	// ❌ Kritim NOT added to workspace (will only access Design space)
	// ❌ Prerak NOT added to workspace (will only access one project)

	log.Printf("✅ Created workspace: ORA Technologies")
	log.Printf("   └─ Members: Marga (owner), Bipin (member)")
	log.Printf("   └─ NOT in workspace: Kritim, Prerak")

	// ============================================
	// SCENARIO 2: CREATE SPACES
	// Bipin creates "Engineering" space (he's workspace member)
	// Marga creates "Design" space
	// ============================================

	// Space 1: Engineering (created by Bipin)
	engineering := &repository.Space{
		Name:        "Engineering",
		Description: stringPtr("All engineering projects"),
		WorkspaceID: workspace.ID,
		OwnerID:     bipin.ID,
		Visibility:  &defaultVisibility,
	}
	repos.SpaceRepo.Create(ctx, engineering)

	// Add Bipin as space owner
	repos.SpaceRepo.AddMember(ctx, &repository.SpaceMember{
		SpaceID: engineering.ID,
		UserID:  bipin.ID,
		Role:    "owner",
	})

	// Space 2: Design (created by Marga)
	design := &repository.Space{
		Name:        "Design",
		Description: stringPtr("Design and UI/UX projects"),
		WorkspaceID: workspace.ID,
		OwnerID:     marga.ID,
		Visibility:  &defaultVisibility,
	}
	repos.SpaceRepo.Create(ctx, design)

	// Add Marga as space owner
	repos.SpaceRepo.AddMember(ctx, &repository.SpaceMember{
		SpaceID: design.ID,
		UserID:  marga.ID,
		Role:    "owner",
	})

	// ✅ Add Kritim ONLY to Design space (she's a designer, not full workspace member)
	repos.SpaceRepo.AddMember(ctx, &repository.SpaceMember{
		SpaceID: design.ID,
		UserID:  kritim.ID,
		Role:    "member",
	})

	log.Printf("✅ Created 2 spaces:")
	log.Printf("   ├─ Engineering (Bipin created)")
	log.Printf("   │  └─ Direct members: Bipin (owner)")
	log.Printf("   │  └─ Inherited: Marga (from workspace)")
	log.Printf("   └─ Design (Marga created)")
	log.Printf("      └─ Direct members: Marga (owner), Kritim (member)")
	log.Printf("      └─ Inherited: Bipin (from workspace)")

	// ============================================
	// SCENARIO 3: CREATE FOLDERS
	// Bipin creates "Backend" folder in Engineering
	// ============================================
	backendFolder := &repository.Folder{
		Name:        "Backend Projects",
		Description: stringPtr("All backend microservices"),
		SpaceID:     engineering.ID,
		OwnerID:     bipin.ID,
		Visibility:  &defaultVisibility,
	}
	repos.FolderRepo.Create(ctx, backendFolder)

	repos.FolderRepo.AddMember(ctx, &repository.FolderMember{
		FolderID: backendFolder.ID,
		UserID:   bipin.ID,
		Role:     "owner",
	})

	log.Printf("✅ Created folder: Backend Projects")
	log.Printf("   └─ Direct: Bipin (owner)")
	log.Printf("   └─ Inherited: Marga (workspace)")

	// ============================================
	// SCENARIO 4: CREATE PROJECTS
	// Project 1: ORA Scrum (in Backend folder, Engineering space)
	// Project 2: Design System (in Design space, NO folder)
	// Project 3: Mobile App (Prerak is contractor, project-only access)
	// ============================================

	// Project 1: ORA Scrum Backend (Bipin leads)
	oraScrum := &repository.Project{
		Name:        "ORA Scrum Backend",
		Key:         "ORA",
		Description: stringPtr("Main scrum management system backend"),
		SpaceID:     engineering.ID,
		FolderID:    &backendFolder.ID,
		LeadID:      &bipin.ID,
		Visibility:  &defaultVisibility,
	}
	repos.ProjectRepo.Create(ctx, oraScrum)

	// Add project members
	repos.ProjectRepo.AddMember(ctx, &repository.ProjectMember{
		ProjectID: oraScrum.ID,
		UserID:    bipin.ID,
		Role:      "lead", // Bipin is project lead
	})
	// Marga has access via workspace (inherited)

	// Project 2: Design System (Kritim + Marga)
	designSystem := &repository.Project{
		Name:        "Design System",
		Key:         "DS",
		Description: stringPtr("Company-wide design system and components"),
		SpaceID:     design.ID,
		FolderID:    nil, // NO folder - direct in space
		LeadID:      &kritim.ID,
		Visibility:  &defaultVisibility,
	}
	repos.ProjectRepo.Create(ctx, designSystem)

	repos.ProjectRepo.AddMember(ctx, &repository.ProjectMember{
		ProjectID: designSystem.ID,
		UserID:    kritim.ID,
		Role:      "lead", // Kritim leads design
	})
	// Marga & Bipin have access via workspace/space

	// Project 3: Mobile App (Prerak is CONTRACTOR - project-only access)
	mobileApp := &repository.Project{
		Name:        "Mobile App",
		Key:         "MOB",
		Description: stringPtr("Customer-facing mobile application"),
		SpaceID:     engineering.ID,
		FolderID:    nil, // Direct in space
		LeadID:      &marga.ID,
		Visibility:  &defaultVisibility,
	}
	repos.ProjectRepo.Create(ctx, mobileApp)

	repos.ProjectRepo.AddMember(ctx, &repository.ProjectMember{
		ProjectID: mobileApp.ID,
		UserID:    marga.ID,
		Role:      "lead",
	})

	// ✅ Add Prerak ONLY to this project (he's a contractor)
	repos.ProjectRepo.AddMember(ctx, &repository.ProjectMember{
		ProjectID: mobileApp.ID,
		UserID:    prerak.ID,
		Role:      "member", // Contractor = member only
	})

	log.Printf("✅ Created 3 projects:")
	log.Printf("   ├─ ORA Scrum Backend (Engineering/Backend)")
	log.Printf("   │  └─ Lead: Bipin | Access: Marga (workspace), Bipin (direct)")
	log.Printf("   ├─ Design System (Design, no folder)")
	log.Printf("   │  └─ Lead: Kritim | Access: Marga (workspace), Bipin (workspace), Kritim (direct)")
	log.Printf("   └─ Mobile App (Engineering, no folder)")
	log.Printf("      └─ Lead: Marga | Access: Marga (direct), Bipin (workspace), Prerak (contractor)")

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
		{"urgent", "#DC2626"},
	}

	for _, l := range labels {
		repos.LabelRepo.Create(ctx, &repository.Label{
			Name:      l.Name,
			Color:     l.Color,
			ProjectID: oraScrum.ID,
		})
	}

	// ============================================
	// CREATE SPRINT
	// ============================================
	now := time.Now()
	sprintStart := now.AddDate(0, 0, -7)
	sprintEnd := now.AddDate(0, 0, 7)

	sprint := &repository.Sprint{
		Name:      "Sprint 1 - December",
		ProjectID: oraScrum.ID,
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
		ProjectID   string
	}{
		{"Setup project structure", "done", "high", []string{bipin.ID}, &sprint.ID, oraScrum.ID},
		{"Implement authentication", "done", "urgent", []string{bipin.ID}, &sprint.ID, oraScrum.ID},
		{"Create dashboard UI", "in_progress", "high", []string{kritim.ID}, &sprint.ID, oraScrum.ID},
		{"API integration", "in_progress", "medium", []string{bipin.ID}, &sprint.ID, oraScrum.ID},
		{"Fix login bug", "todo", "urgent", []string{marga.ID}, &sprint.ID, oraScrum.ID},
		{"Add dark mode", "todo", "low", []string{kritim.ID}, &sprint.ID, oraScrum.ID},
		{"Write documentation", "backlog", "low", []string{}, nil, oraScrum.ID},
		{"Performance optimization", "backlog", "medium", []string{prerak.ID}, nil, oraScrum.ID},
	}

	for i, t := range tasks {
		task := &repository.Task{
			Title:       t.Title,
			Status:      t.Status,
			Priority:    t.Priority,
			ProjectID:   t.ProjectID,
			SprintID:    t.SprintID,
			AssigneeIDs: t.AssigneeIDs,
			LabelIDs:    []string{},
			CreatedBy:   &marga.ID,
			Position:    i,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		repos.TaskRepo.Create(ctx, task)
	}

	// ============================================
	// CREATE SAMPLE NOTIFICATIONS
	// ============================================
	seedNotifications(ctx, repos, marga.ID, bipin.ID, kritim.ID, prerak.ID, oraScrum.ID, workspace.ID, sprint.ID)

	// ============================================
	// FINAL SUMMARY
	// ============================================
	log.Println("")
	log.Println("🎉 ============================================")
	log.Println("🎉 SEED COMPLETE - ACCESS SUMMARY")
	log.Println("🎉 ============================================")
	log.Println("")
	log.Println("👤 MARGA GHALE (marga.ghale@oratechnologies.io)")
	log.Println("   Role: WORKSPACE OWNER")
	log.Println("   Access:")
	log.Println("   ✅ Workspace: ORA Technologies (owner)")
	log.Println("   ✅ Space: Engineering (inherited)")
	log.Println("   ✅ Space: Design (owner)")
	log.Println("   ✅ Folder: Backend Projects (inherited)")
	log.Println("   ✅ Project: ORA Scrum Backend (inherited)")
	log.Println("   ✅ Project: Design System (inherited)")
	log.Println("   ✅ Project: Mobile App (lead)")
	log.Println("")
	log.Println("👤 BIPIN DHIMAL (bipin.dhimal@oratechnologies.io)")
	log.Println("   Role: WORKSPACE MEMBER")
	log.Println("   Access:")
	log.Println("   ✅ Workspace: ORA Technologies (member)")
	log.Println("   ✅ Space: Engineering (owner)")
	log.Println("   ✅ Space: Design (inherited)")
	log.Println("   ✅ Folder: Backend Projects (owner)")
	log.Println("   ✅ Project: ORA Scrum Backend (lead)")
	log.Println("   ✅ Project: Design System (inherited)")
	log.Println("   ✅ Project: Mobile App (inherited)")
	log.Println("")
	log.Println("👤 KRITIM KAFLE (kritim.kafle@oratechnologies.io)")
	log.Println("   Role: SPACE-ONLY MEMBER (Designer)")
	log.Println("   Access:")
	log.Println("   ❌ Workspace: ORA Technologies (no access)")
	log.Println("   ❌ Space: Engineering (no access)")
	log.Println("   ✅ Space: Design (member)")
	log.Println("   ❌ Folder: Backend Projects (no access)")
	log.Println("   ❌ Project: ORA Scrum Backend (no access)")
	log.Println("   ✅ Project: Design System (lead)")
	log.Println("   ❌ Project: Mobile App (no access)")
	log.Println("")
	log.Println("👤 PRERAK KHADKA (prerak.khadka@oratechnologies.io)")
	log.Println("   Role: PROJECT-ONLY CONTRACTOR")
	log.Println("   Access:")
	log.Println("   ❌ Workspace: ORA Technologies (no access)")
	log.Println("   ❌ Space: Engineering (no access)")
	log.Println("   ❌ Space: Design (no access)")
	log.Println("   ❌ Folder: Backend Projects (no access)")
	log.Println("   ❌ Project: ORA Scrum Backend (no access)")
	log.Println("   ❌ Project: Design System (no access)")
	log.Println("   ✅ Project: Mobile App (member)")
	log.Println("")
	log.Println("🎯 Test Credentials:")
	log.Println("   Email: any of the above")
	log.Println("   Password: password123")
	log.Println("")
}

// seedNotifications creates sample notifications for all users
func seedNotifications(ctx context.Context, repos *repository.Repositories, margaID, bipinID, kritimID, prerakID, projectID, workspaceID, sprintID string) {
	now := time.Now()

	notifications := []repository.Notification{
		// Marga's notifications (workspace owner)
		{
			UserID:    margaID,
			Type:      "WORKSPACE_CREATED",
			Title:     "Workspace Created",
			Message:   "You created workspace: ORA Technologies",
			Read:      true,
			Data:      map[string]interface{}{"workspaceId": workspaceID},
			CreatedAt: now.Add(-10 * 24 * time.Hour),
		},
		{
			UserID:    margaID,
			Type:      "TASK_ASSIGNED",
			Title:     "Task Assigned",
			Message:   "You have been assigned to task: Fix login bug",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-5", "projectId": projectID, "taskKey": "ORA-5"},
			CreatedAt: now.Add(-1 * time.Hour),
		},
		{
			UserID:    margaID,
			Type:      "SPRINT_ENDING",
			Title:     "Sprint Ending Soon",
			Message:   "Sprint 'Sprint 1 - December' ends in 7 days",
			Read:      false,
			Data:      map[string]interface{}{"sprintId": sprintID, "projectId": projectID, "daysRemaining": 7},
			CreatedAt: now.Add(-3 * time.Hour),
		},

		// Bipin's notifications (workspace member, space owner)
		{
			UserID:    bipinID,
			Type:      "WORKSPACE_INVITATION",
			Title:     "Workspace Invitation",
			Message:   "Marga Ghale added you to workspace: ORA Technologies",
			Read:      true,
			Data:      map[string]interface{}{"workspaceId": workspaceID},
			CreatedAt: now.Add(-9 * 24 * time.Hour),
		},
		{
			UserID:    bipinID,
			Type:      "SPACE_CREATED",
			Title:     "Space Created",
			Message:   "You created space: Engineering",
			Read:      true,
			Data:      map[string]interface{}{"spaceId": "space-eng"},
			CreatedAt: now.Add(-8 * 24 * time.Hour),
		},
		{
			UserID:    bipinID,
			Type:      "TASK_ASSIGNED",
			Title:     "Task Assigned",
			Message:   "You have been assigned to task: API integration",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-4", "projectId": projectID, "taskKey": "ORA-4"},
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			UserID:    bipinID,
			Type:      "TASK_COMPLETED",
			Title:     "Task Completed",
			Message:   "You completed task: Implement authentication",
			Read:      true,
			Data:      map[string]interface{}{"taskId": "task-2", "projectId": projectID, "taskKey": "ORA-2"},
			CreatedAt: now.Add(-5 * 24 * time.Hour),
		},

		// Kritim's notifications (space-only member)
		{
			UserID:    kritimID,
			Type:      "SPACE_INVITATION",
			Title:     "Space Invitation",
			Message:   "Marga Ghale added you to space: Design",
			Read:      true,
			Data:      map[string]interface{}{"spaceId": "space-design"},
			CreatedAt: now.Add(-7 * 24 * time.Hour),
		},
		{
			UserID:    kritimID,
			Type:      "PROJECT_INVITATION",
			Title:     "Project Lead",
			Message:   "You are now lead of project: Design System",
			Read:      false,
			Data:      map[string]interface{}{"projectId": "project-ds"},
			CreatedAt: now.Add(-6 * 24 * time.Hour),
		},
		{
			UserID:    kritimID,
			Type:      "TASK_ASSIGNED",
			Title:     "Task Assigned",
			Message:   "You have been assigned to task: Create dashboard UI",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-3", "projectId": projectID, "taskKey": "ORA-3"},
			CreatedAt: now.Add(-4 * time.Hour),
		},

		// Prerak's notifications (contractor - project only)
		{
			UserID:    prerakID,
			Type:      "PROJECT_INVITATION",
			Title:     "Project Invitation",
			Message:   "Marga Ghale added you to project: Mobile App",
			Read:      false,
			Data:      map[string]interface{}{"projectId": "project-mobile"},
			CreatedAt: now.Add(-2 * 24 * time.Hour),
		},
		{
			UserID:    prerakID,
			Type:      "TASK_ASSIGNED",
			Title:     "Task Assigned",
			Message:   "You have been assigned to task: Performance optimization",
			Read:      false,
			Data:      map[string]interface{}{"taskId": "task-8", "projectId": projectID, "taskKey": "ORA-8"},
			CreatedAt: now.Add(-1 * 24 * time.Hour),
		},
	}

	for _, n := range notifications {
		notif := n // Create a copy to avoid pointer issues
		repos.NotificationRepo.Create(ctx, &notif)
	}

	log.Printf("✅ Created %d notifications for all users", len(notifications))
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}