package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Permission Levels (ClickUp-like hierarchy)
// ============================================

// Permission levels from highest to lowest
const (
	PermissionOwner  = "owner"
	PermissionAdmin  = "admin"
	PermissionLead   = "lead"
	PermissionMember = "member"
	PermissionViewer = "viewer"
)

// Action types
const (
	ActionView   = "view"
	ActionEdit   = "edit"
	ActionDelete = "delete"
	ActionManage = "manage" // manage members, settings
	ActionCreate = "create"
)

// PermissionService handles hierarchical permission checking like ClickUp
type PermissionService interface {
	// Workspace permissions
	CanAccessWorkspace(ctx context.Context, userID, workspaceID string) bool
	CanManageWorkspace(ctx context.Context, userID, workspaceID string) bool
	GetWorkspaceRole(ctx context.Context, userID, workspaceID string) string

	// Space permissions
	CanAccessSpace(ctx context.Context, userID, spaceID string) bool
	CanManageSpace(ctx context.Context, userID, spaceID string) bool
	GetSpaceRole(ctx context.Context, userID, spaceID string) string

	// Project permissions
	CanAccessProject(ctx context.Context, userID, projectID string) bool
	CanManageProject(ctx context.Context, userID, projectID string) bool
	CanEditProject(ctx context.Context, userID, projectID string) bool
	GetProjectRole(ctx context.Context, userID, projectID string) string

	// Task permissions
	CanAccessTask(ctx context.Context, userID, taskID string) bool
	CanEditTask(ctx context.Context, userID, taskID string) bool
	CanDeleteTask(ctx context.Context, userID, taskID string) bool

	// Team permissions
	CanAccessTeam(ctx context.Context, userID, teamID string) bool
	CanManageTeam(ctx context.Context, userID, teamID string) bool

	// General permission check
	CheckPermission(ctx context.Context, userID, entityType, entityID, action string) bool
}

type permissionService struct {
	workspaceRepo repository.WorkspaceRepository
	spaceRepo     repository.SpaceRepository
	projectRepo   repository.ProjectRepository
	taskRepo      repository.TaskRepository
	teamRepo      repository.TeamRepository
	folderRepo    repository.FolderRepository
}

// NewPermissionService creates a new permission service
func NewPermissionService(
    workspaceRepo repository.WorkspaceRepository,
    spaceRepo repository.SpaceRepository,
    projectRepo repository.ProjectRepository,
    taskRepo repository.TaskRepository,
    teamRepo repository.TeamRepository,
    folderRepo repository.FolderRepository,  // ✅ Add this parameter
) PermissionService {
    return &permissionService{
        workspaceRepo: workspaceRepo,
        spaceRepo:     spaceRepo,
        projectRepo:   projectRepo,
        taskRepo:      taskRepo,
        teamRepo:      teamRepo,
        folderRepo:    folderRepo,  // ✅ Add this
    }
}
// ============================================
// Permission Hierarchy Helpers
// ============================================

// roleLevel returns numeric level for role comparison (higher = more permissions)
func roleLevel(role string) int {
	switch role {
	case PermissionOwner:
		return 5
	case PermissionAdmin:
		return 4
	case PermissionLead:
		return 3
	case PermissionMember:
		return 2
	case PermissionViewer:
		return 1
	default:
		return 0
	}
}

// hasMinimumRole checks if user has at least the minimum required role
func hasMinimumRole(userRole, minRole string) bool {
	return roleLevel(userRole) >= roleLevel(minRole)
}

// ============================================
// Workspace Permissions
// ============================================

func (s *permissionService) CanAccessWorkspace(ctx context.Context, userID, workspaceID string) bool {
	member, err := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	return err == nil && member != nil
}

func (s *permissionService) CanManageWorkspace(ctx context.Context, userID, workspaceID string) bool {
	role := s.GetWorkspaceRole(ctx, userID, workspaceID)
	return hasMinimumRole(role, PermissionAdmin)
}

func (s *permissionService) GetWorkspaceRole(ctx context.Context, userID, workspaceID string) string {
	member, err := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return ""
	}
	return normalizeRole(member.Role)
}

// ============================================
// Space Permissions
// ============================================

func (s *permissionService) CanAccessSpace(ctx context.Context, userID, spaceID string) bool {
	space, err := s.spaceRepo.FindByID(ctx, spaceID)
	if err != nil || space == nil {
		return false
	}

	// Check workspace membership first
	if !s.CanAccessWorkspace(ctx, userID, space.WorkspaceID) {
		return false
	}

	// Check space visibility
	if space.Visibility != nil {
		switch *space.Visibility {
		case "private":
			// Check if user is in allowed_users or allowed_teams
			return s.isAllowedInSpace(ctx, userID, space)
		case "workspace":
			return true // Workspace members can access
		case "public":
			return true
		}
	}

	return true // Default: workspace members can access
}

func (s *permissionService) CanManageSpace(ctx context.Context, userID, spaceID string) bool {
	space, err := s.spaceRepo.FindByID(ctx, spaceID)
	if err != nil || space == nil {
		return false
	}

	// Workspace admins can manage spaces
	return s.CanManageWorkspace(ctx, userID, space.WorkspaceID)
}

func (s *permissionService) GetSpaceRole(ctx context.Context, userID, spaceID string) string {
    // ✅ Check direct space membership first
    member, err := s.spaceRepo.FindMember(ctx, spaceID, userID)
    if err == nil && member != nil {
        return normalizeRole(member.Role)
    }

    // Fall back to workspace role
    space, err := s.spaceRepo.FindByID(ctx, spaceID)
    if err != nil || space == nil {
        return ""
    }
    return s.GetWorkspaceRole(ctx, userID, space.WorkspaceID)
}

func (s *permissionService) isAllowedInSpace(ctx context.Context, userID string, space *repository.Space) bool {
	// Check allowed users
	if space.AllowedUsers != nil {
		for _, allowedID := range space.AllowedUsers {
			if allowedID == userID {
				return true
			}
		}
	}

	// Check allowed teams
	if space.AllowedTeams != nil && s.teamRepo != nil {
		for _, teamID := range space.AllowedTeams {
			isMember, _ := s.teamRepo.IsMember(ctx, teamID, userID)
			if isMember {
				return true
			}
		}
	}

	return false
}

// ============================================
// Project Permissions
// ============================================

func (s *permissionService) CanAccessProject(ctx context.Context, userID, projectID string) bool {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil || project == nil {
		return false
	}

	// Check project membership first
	member, _ := s.projectRepo.FindMember(ctx, projectID, userID)
	if member != nil {
		return true
	}

	// Check space access
	return s.CanAccessSpace(ctx, userID, project.SpaceID)
}

func (s *permissionService) CanManageProject(ctx context.Context, userID, projectID string) bool {
	role := s.GetProjectRole(ctx, userID, projectID)
	return hasMinimumRole(role, PermissionAdmin)
}

func (s *permissionService) CanEditProject(ctx context.Context, userID, projectID string) bool {
	role := s.GetProjectRole(ctx, userID, projectID)
	return hasMinimumRole(role, PermissionMember)
}

func (s *permissionService) GetProjectRole(ctx context.Context, userID, projectID string) string {
    // ✅ Check direct project membership
    member, err := s.projectRepo.FindMember(ctx, projectID, userID)
    if err == nil && member != nil {
        return normalizeRole(member.Role)
    }

    project, err := s.projectRepo.FindByID(ctx, projectID)
    if err != nil || project == nil {
        return ""
    }

    // ✅ Check folder membership if project is in a folder
    if project.FolderID != nil {
        folderMember, _ := s.folderRepo.FindMember(ctx, *project.FolderID, userID)
        if folderMember != nil {
            return normalizeRole(folderMember.Role)
        }
    }

    return s.GetSpaceRole(ctx, userID, project.SpaceID)
}
// ============================================
// Task Permissions
// ============================================

func (s *permissionService) CanAccessTask(ctx context.Context, userID, taskID string) bool {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || task == nil {
		return false
	}

	return s.CanAccessProject(ctx, userID, task.ProjectID)
}

func (s *permissionService) CanEditTask(ctx context.Context, userID, taskID string) bool {
    task, err := s.taskRepo.FindByID(ctx, taskID)
    if err != nil || task == nil {
        return false
    }

    // ✅ Check if user is one of the assignees (it's an array!)
    for _, assigneeID := range task.AssigneeIDs {
        if assigneeID == userID {
            return true
        }
    }

    // ✅ NO ReporterID field exists - remove this check or add CreatedBy field
    // You might want to add a CreatedBy/ReporterID field to Task struct

    // Check project-level edit permission
    return s.CanEditProject(ctx, userID, task.ProjectID)
}


func (s *permissionService) CanDeleteTask(ctx context.Context, userID, taskID string) bool {
    task, err := s.taskRepo.FindByID(ctx, taskID)
    if err != nil || task == nil {
        return false
    }

    // ✅ NO ReporterID field exists - remove or add to struct
    // Project managers can delete
    return s.CanManageProject(ctx, userID, task.ProjectID)
}


// ============================================
// Team Permissions
// ============================================

func (s *permissionService) CanAccessTeam(ctx context.Context, userID, teamID string) bool {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil || team == nil {
		return false
	}

	// Check team membership
	isMember, _ := s.teamRepo.IsMember(ctx, teamID, userID)
	if isMember {
		return true
	}

	// Workspace admins can access all teams
	return s.CanManageWorkspace(ctx, userID, team.WorkspaceID)
}

func (s *permissionService) CanManageTeam(ctx context.Context, userID, teamID string) bool {
	member, err := s.teamRepo.FindMember(ctx, teamID, userID)
	if err == nil && member != nil {
		return hasMinimumRole(normalizeRole(member.Role), PermissionAdmin)
	}

	// Workspace admins can manage teams
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil || team == nil {
		return false
	}

	return s.CanManageWorkspace(ctx, userID, team.WorkspaceID)
}

// ============================================
// Generic Permission Check
// ============================================

func (s *permissionService) CheckPermission(ctx context.Context, userID, entityType, entityID, action string) bool {
	switch entityType {
	case "workspace":
		switch action {
		case ActionView:
			return s.CanAccessWorkspace(ctx, userID, entityID)
		case ActionEdit, ActionManage, ActionDelete:
			return s.CanManageWorkspace(ctx, userID, entityID)
		}

	case "space":
		switch action {
		case ActionView:
			return s.CanAccessSpace(ctx, userID, entityID)
		case ActionEdit, ActionManage, ActionDelete:
			return s.CanManageSpace(ctx, userID, entityID)
		}

	case "project":
		switch action {
		case ActionView:
			return s.CanAccessProject(ctx, userID, entityID)
		case ActionEdit:
			return s.CanEditProject(ctx, userID, entityID)
		case ActionManage, ActionDelete:
			return s.CanManageProject(ctx, userID, entityID)
		}

	case "task":
		switch action {
		case ActionView:
			return s.CanAccessTask(ctx, userID, entityID)
		case ActionEdit:
			return s.CanEditTask(ctx, userID, entityID)
		case ActionDelete:
			return s.CanDeleteTask(ctx, userID, entityID)
		}

	case "team":
		switch action {
		case ActionView:
			return s.CanAccessTeam(ctx, userID, entityID)
		case ActionEdit, ActionManage, ActionDelete:
			return s.CanManageTeam(ctx, userID, entityID)
		}
	}

	return false
}

// normalizeRole converts various role formats to standard lowercase
func normalizeRole(role string) string {
	switch role {
	case "OWNER", "owner":
		return PermissionOwner
	case "ADMIN", "admin":
		return PermissionAdmin
	case "LEAD", "lead":
		return PermissionLead
	case "MEMBER", "member":
		return PermissionMember
	case "VIEWER", "viewer":
		return PermissionViewer
	default:
		return role
	}
}
