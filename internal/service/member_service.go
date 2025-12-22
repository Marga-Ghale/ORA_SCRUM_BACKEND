package service

import (
	"context"
	"log"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// MemberService handles member operations across all entity types
type MemberService interface {
	// Direct member operations
	AddMember(ctx context.Context, entityType, entityID, userID, role, inviterID string) error
	RemoveMember(ctx context.Context, entityType, entityID, userID string) error
	UpdateMemberRole(ctx context.Context, entityType, entityID, userID, role string) error
	GetMember(ctx context.Context, entityType, entityID, userID string) (*UnifiedMember, error)
	
	// Listing members
	ListDirectMembers(ctx context.Context, entityType, entityID string) ([]*UnifiedMember, error)
	ListEffectiveMembers(ctx context.Context, entityType, entityID string) ([]*UnifiedMember, error)
	
	// Access control
	HasDirectAccess(ctx context.Context, entityType, entityID, userID string) (bool, error)
	HasEffectiveAccess(ctx context.Context, entityType, entityID, userID string) (bool, string, error)
	GetAccessLevel(ctx context.Context, entityType, entityID, userID string) (string, string, error)
	
	// Invitation
	InviteMemberByEmail(ctx context.Context, entityType, entityID, email, role, inviterID string) error
	InviteMemberByID(ctx context.Context, entityType, entityID, userID, role, inviterID string) error
	
	// User memberships
	GetUserMemberships(ctx context.Context, userID string) (map[string][]string, error)
	GetUserAllAccess(ctx context.Context, userID string) (*UserAccessMap, error)

	// NEW: Accessible entities
	GetAccessibleWorkspaces(ctx context.Context, userID string) ([]*repository.Workspace, error)
	GetAccessibleSpaces(ctx context.Context, userID string) ([]*repository.Space, error)
	GetAccessibleFolders(ctx context.Context, userID string) ([]*repository.Folder, error)
	GetAccessibleProjects(ctx context.Context, userID string) ([]*repository.Project, error)
}

// EntityType constants
const (
	EntityTypeWorkspace = "workspace"
	EntityTypeSpace     = "space"
	EntityTypeFolder    = "folder"
	EntityTypeProject   = "project"
)

// UnifiedMember represents a member across any entity type
type UnifiedMember struct {
	ID           string
	EntityType   string
	EntityID     string
	UserID       string
	Role         string
	JoinedAt     time.Time
	User         *repository.User
	IsInherited  bool   // True if access comes from parent entity
	InheritedFrom string // EntityType where membership originates
}


type UserAccessMap struct {
	Workspaces []string `json:"workspaces"`
	Spaces     []string `json:"spaces"`
	Folders    []string `json:"folders"`
	Projects   []string `json:"projects"`
}

type memberService struct {
	workspaceRepo repository.WorkspaceRepository
	spaceRepo     repository.SpaceRepository
	folderRepo    repository.FolderRepository
	projectRepo   repository.ProjectRepository
	userRepo      repository.UserRepository
	notifSvc      *notification.Service
}

func NewMemberService(
	workspaceRepo repository.WorkspaceRepository,
	spaceRepo repository.SpaceRepository,
	folderRepo repository.FolderRepository,
	projectRepo repository.ProjectRepository,
	userRepo repository.UserRepository,
	notifSvc *notification.Service,
) MemberService {
	return &memberService{
		workspaceRepo: workspaceRepo,
		spaceRepo:     spaceRepo,
		folderRepo:    folderRepo,
		projectRepo:   projectRepo,
		userRepo:      userRepo,
		notifSvc:      notifSvc,
	}
}

// ============================================
// DIRECT MEMBER OPERATIONS
// ============================================


// func (s *memberService) AddMember(ctx context.Context, entityType, entityID, userID, role, inviterID string) error {
// 	// Verify user exists first
// 	user, err := s.userRepo.FindByID(ctx, userID)
// 	if err != nil || user == nil {
// 		return ErrUserNotFound
// 	}

// 	// Check if already a direct member
// 	existing, _ := s.GetMember(ctx, entityType, entityID, userID)
// 	if existing != nil && !existing.IsInherited {
// 		return ErrConflict
// 	}

// 	// ✅ FIXED: Skip permission check if creator is adding themselves
// 	// This happens when creating new workspace/space/folder/project
// 	if inviterID != userID {
// 		// Only check permissions if someone else is adding this user
// 		hasPermission := false
		
// 		switch entityType {
// 		case EntityTypeWorkspace:
// 			member, _ := s.workspaceRepo.FindMember(ctx, entityID, inviterID)
// 			if member != nil {
// 				roleVal := getRoleLevel(member.Role)
// 				hasPermission = roleVal >= 4  // admin or owner
// 			}
// 		case EntityTypeSpace:
// 			space, _ := s.spaceRepo.FindByID(ctx, entityID)
// 			if space != nil {
// 				wsMember, _ := s.workspaceRepo.FindMember(ctx, space.WorkspaceID, inviterID)
// 				if wsMember != nil {
// 					roleVal := getRoleLevel(wsMember.Role)
// 					hasPermission = roleVal >= 4  // admin or owner
// 				}
// 			}
// 		case EntityTypeFolder:
// 			folder, _ := s.folderRepo.FindByID(ctx, entityID)
// 			if folder != nil {
// 				spaceMember, _ := s.spaceRepo.FindMember(ctx, folder.SpaceID, inviterID)
// 				if spaceMember != nil {
// 					roleVal := getRoleLevel(spaceMember.Role)
// 					hasPermission = roleVal >= 4  // admin or owner
// 				}
// 			}
// 		case EntityTypeProject:
// 			projMember, _ := s.projectRepo.FindMember(ctx, entityID, inviterID)
// 			if projMember != nil {
// 				roleVal := getRoleLevel(projMember.Role)
// 				hasPermission = roleVal >= 3  // ✅ FIXED: lead, admin, or owner
// 			}
// 		}

// 		if !hasPermission {
// 			return ErrUnauthorized
// 		}
// 	}

// 	// Delegate to appropriate repository
// 	switch entityType {
// 	case EntityTypeWorkspace:
// 		member := &repository.WorkspaceMember{
// 			WorkspaceID: entityID,
// 			UserID:      userID,
// 			Role:        role,
// 		}
// 		if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
// 			return err
// 		}
// 		s.sendNotification(ctx, entityType, entityID, userID, inviterID)
// 		return nil

// 	case EntityTypeSpace:
// 		member := &repository.SpaceMember{
// 			SpaceID: entityID,
// 			UserID:  userID,
// 			Role:    role,
// 		}
// 		if err := s.spaceRepo.AddMember(ctx, member); err != nil {
// 			return err
// 		}
// 		s.sendNotification(ctx, entityType, entityID, userID, inviterID)
// 		return nil

// 	case EntityTypeFolder:
// 		member := &repository.FolderMember{
// 			FolderID: entityID,
// 			UserID:   userID,
// 			Role:     role,
// 		}
// 		if err := s.folderRepo.AddMember(ctx, member); err != nil {
// 			return err
// 		}
// 		s.sendNotification(ctx, entityType, entityID, userID, inviterID)
// 		return nil

// 	case EntityTypeProject:
// 		member := &repository.ProjectMember{
// 			ProjectID: entityID,
// 			UserID:    userID,
// 			Role:      role,
// 		}
// 		if err := s.projectRepo.AddMember(ctx, member); err != nil {
// 			return err
// 		}
// 		s.sendNotification(ctx, entityType, entityID, userID, inviterID)
// 		return nil

// 	default:
// 		return ErrInvalidEntityType
// 	}
// }



// ✅ FIXED: AddMember with proper permission checks
func (s *memberService) AddMember(ctx context.Context, entityType, entityID, userID, role, inviterID string) error {
	// Verify user exists first
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	// Check if already a direct member
	existing, _ := s.GetMember(ctx, entityType, entityID, userID)
	if existing != nil && !existing.IsInherited {
		return ErrConflict
	}

	// ✅ FIXED: Permission check logic
	hasPermission := false
	isCreatorAddingSelf := inviterID == userID

	// ✅ CRITICAL FIX: For workspace creation, skip permission check
	if entityType == EntityTypeWorkspace && isCreatorAddingSelf {
		// When creating workspace, creator adds themselves - ALWAYS ALLOW
		hasPermission = true
	} else {
		// Check permissions for adding OTHER users
		switch entityType {
		case EntityTypeWorkspace:
			// Only workspace admin/owner can add members
			member, _ := s.workspaceRepo.FindMember(ctx, entityID, inviterID)
			if member != nil {
				roleVal := getRoleLevel(member.Role)
				hasPermission = roleVal >= 4 // admin or owner
			}

		case EntityTypeSpace:
			// Space creation: creator adds themselves
			if isCreatorAddingSelf {
				// ✅ Check workspace permission
				space, _ := s.spaceRepo.FindByID(ctx, entityID)
				if space != nil {
					// Check if inviter has workspace access
					wsMember, _ := s.workspaceRepo.FindMember(ctx, space.WorkspaceID, inviterID)
					hasPermission = wsMember != nil // Any workspace member can create space
				}
			} else {
				// Adding other users: need admin/owner
				space, _ := s.spaceRepo.FindByID(ctx, entityID)
				if space != nil {
					wsMember, _ := s.workspaceRepo.FindMember(ctx, space.WorkspaceID, inviterID)
					if wsMember != nil {
						roleVal := getRoleLevel(wsMember.Role)
						hasPermission = roleVal >= 4
					}
				}
			}

		case EntityTypeFolder:
			// Folder creation: creator adds themselves
			if isCreatorAddingSelf {
				folder, _ := s.folderRepo.FindByID(ctx, entityID)
				if folder != nil {
					// Check space access
					spaceMember, _ := s.spaceRepo.FindMember(ctx, folder.SpaceID, inviterID)
					if spaceMember != nil {
						hasPermission = true
					} else {
						// Check workspace access
						space, _ := s.spaceRepo.FindByID(ctx, folder.SpaceID)
						if space != nil {
							wsMember, _ := s.workspaceRepo.FindMember(ctx, space.WorkspaceID, inviterID)
							hasPermission = wsMember != nil
						}
					}
				}
			} else {
				// Adding other users
				folder, _ := s.folderRepo.FindByID(ctx, entityID)
				if folder != nil {
					spaceMember, _ := s.spaceRepo.FindMember(ctx, folder.SpaceID, inviterID)
					if spaceMember != nil {
						roleVal := getRoleLevel(spaceMember.Role)
						hasPermission = roleVal >= 4
					}
				}
			}

		case EntityTypeProject:
			// Project creation: creator adds themselves
			if isCreatorAddingSelf {
				project, _ := s.projectRepo.FindByID(ctx, entityID)
				if project != nil {
					// Check folder/space/workspace access
					if project.FolderID != nil {
						folderMember, _ := s.folderRepo.FindMember(ctx, *project.FolderID, inviterID)
						if folderMember != nil {
							hasPermission = true
						}
					}
					if !hasPermission {
						spaceMember, _ := s.spaceRepo.FindMember(ctx, project.SpaceID, inviterID)
						if spaceMember != nil {
							hasPermission = true
						}
					}
					if !hasPermission {
						space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
						if space != nil {
							wsMember, _ := s.workspaceRepo.FindMember(ctx, space.WorkspaceID, inviterID)
							hasPermission = wsMember != nil
						}
					}
				}
			} else {
				// Adding other users: need lead/admin/owner
				projMember, _ := s.projectRepo.FindMember(ctx, entityID, inviterID)
				if projMember != nil {
					roleVal := getRoleLevel(projMember.Role)
					hasPermission = roleVal >= 3 // lead, admin, or owner
				}
			}
		}
	}

	if !hasPermission {
		log.Printf("[AddMember] DENIED: entityType=%s entityID=%s userID=%s inviterID=%s", 
			entityType, entityID, userID, inviterID)
		return ErrUnauthorized
	}

	// Delegate to appropriate repository
	switch entityType {
	case EntityTypeWorkspace:
		member := &repository.WorkspaceMember{
			WorkspaceID: entityID,
			UserID:      userID,
			Role:        role,
		}
		if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
			return err
		}
		s.sendNotification(ctx, entityType, entityID, userID, inviterID)
		return nil

	case EntityTypeSpace:
		member := &repository.SpaceMember{
			SpaceID: entityID,
			UserID:  userID,
			Role:    role,
		}
		if err := s.spaceRepo.AddMember(ctx, member); err != nil {
			return err
		}
		s.sendNotification(ctx, entityType, entityID, userID, inviterID)
		return nil

	case EntityTypeFolder:
		member := &repository.FolderMember{
			FolderID: entityID,
			UserID:   userID,
			Role:     role,
		}
		if err := s.folderRepo.AddMember(ctx, member); err != nil {
			return err
		}
		s.sendNotification(ctx, entityType, entityID, userID, inviterID)
		return nil

	case EntityTypeProject:
		member := &repository.ProjectMember{
			ProjectID: entityID,
			UserID:    userID,
			Role:      role,
		}
		if err := s.projectRepo.AddMember(ctx, member); err != nil {
			return err
		}
		s.sendNotification(ctx, entityType, entityID, userID, inviterID)
		return nil

	default:
		return ErrInvalidEntityType
	}
}
func getRoleLevel(role string) int {
	roleMap := map[string]int{
		"owner":  5,
		"admin":  4,
		"lead":   3,
		"member": 2,
		"viewer": 1,
	}
	if level, ok := roleMap[role]; ok {
		return level
	}
	return 0
}

func (s *memberService) RemoveMember(ctx context.Context, entityType, entityID, userID string) error {
	switch entityType {
	case EntityTypeWorkspace:
		return s.workspaceRepo.RemoveMember(ctx, entityID, userID)
	case EntityTypeSpace:
		return s.spaceRepo.RemoveMember(ctx, entityID, userID)
	case EntityTypeFolder:
		return s.folderRepo.RemoveMember(ctx, entityID, userID)
	case EntityTypeProject:
		return s.projectRepo.RemoveMember(ctx, entityID, userID)
	default:
		return ErrInvalidEntityType
	}
}

func (s *memberService) UpdateMemberRole(ctx context.Context, entityType, entityID, userID, role string) error {
	switch entityType {
	case EntityTypeWorkspace:
		return s.workspaceRepo.UpdateMemberRole(ctx, entityID, userID, role)
	case EntityTypeSpace:
		return s.spaceRepo.UpdateMemberRole(ctx, entityID, userID, role)
	case EntityTypeFolder:
		return s.folderRepo.UpdateMemberRole(ctx, entityID, userID, role)
	case EntityTypeProject:
		return s.projectRepo.UpdateMemberRole(ctx, entityID, userID, role)
	default:
		return ErrInvalidEntityType
	}
}

func (s *memberService) GetMember(ctx context.Context, entityType, entityID, userID string) (*UnifiedMember, error) {
	switch entityType {
	case EntityTypeWorkspace:
		member, err := s.workspaceRepo.FindMember(ctx, entityID, userID)
		if err != nil || member == nil {
			return nil, err
		}
		return &UnifiedMember{
			ID:          member.ID,
			EntityType:  EntityTypeWorkspace,
			EntityID:    entityID,
			UserID:      member.UserID,
			Role:        member.Role,
			JoinedAt:    member.JoinedAt,
			User:        member.User,
			IsInherited: false,
		}, nil

	case EntityTypeSpace:
		member, err := s.spaceRepo.FindMember(ctx, entityID, userID)
		if err != nil || member == nil {
			return nil, err
		}
		return &UnifiedMember{
			ID:          member.ID,
			EntityType:  EntityTypeSpace,
			EntityID:    entityID,
			UserID:      member.UserID,
			Role:        member.Role,
			JoinedAt:    member.JoinedAt,
			User:        member.User,
			IsInherited: false,
		}, nil

	case EntityTypeFolder:
		member, err := s.folderRepo.FindMember(ctx, entityID, userID)
		if err != nil || member == nil {
			return nil, err
		}
		return &UnifiedMember{
			ID:          member.ID,
			EntityType:  EntityTypeFolder,
			EntityID:    entityID,
			UserID:      member.UserID,
			Role:        member.Role,
			JoinedAt:    member.JoinedAt,
			User:        member.User,
			IsInherited: false,
		}, nil

	case EntityTypeProject:
		member, err := s.projectRepo.FindMember(ctx, entityID, userID)
		if err != nil || member == nil {
			return nil, err
		}
		return &UnifiedMember{
			ID:          member.ID,
			EntityType:  EntityTypeProject,
			EntityID:    entityID,
			UserID:      member.UserID,
			Role:        member.Role,
			JoinedAt:    member.JoinedAt,
			User:        member.User,
			IsInherited: false,
		}, nil

	default:
		return nil, ErrInvalidEntityType
	}
}

// ============================================
// LISTING MEMBERS
// ============================================

// ListDirectMembers returns only direct members of the entity
func (s *memberService) ListDirectMembers(ctx context.Context, entityType, entityID string) ([]*UnifiedMember, error) {
	switch entityType {
	case EntityTypeWorkspace:
		members, err := s.workspaceRepo.FindMembers(ctx, entityID)
		if err != nil {
			return nil, err
		}
		return s.convertWorkspaceMembers(members, entityID, false), nil

	case EntityTypeSpace:
		members, err := s.spaceRepo.FindMembers(ctx, entityID)
		if err != nil {
			return nil, err
		}
		return s.convertSpaceMembers(members, entityID, false), nil

	case EntityTypeFolder:
		members, err := s.folderRepo.FindMembers(ctx, entityID)
		if err != nil {
			return nil, err
		}
		return s.convertFolderMembers(members, entityID, false), nil

	case EntityTypeProject:
		members, err := s.projectRepo.FindMembers(ctx, entityID)
		if err != nil {
			return nil, err
		}
		return s.convertProjectMembers(members, entityID, false), nil

	default:
		return nil, ErrInvalidEntityType
	}
}

// ListEffectiveMembers returns direct + inherited members (from parent entities)
func (s *memberService) ListEffectiveMembers(ctx context.Context, entityType, entityID string) ([]*UnifiedMember, error) {
	// Start with direct members
	directMembers, err := s.ListDirectMembers(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}

	// Create map to track unique users
	memberMap := make(map[string]*UnifiedMember)
	for _, m := range directMembers {
		memberMap[m.UserID] = m
	}

	// Add inherited members based on entity type
	switch entityType {
	case EntityTypeWorkspace:
		// Workspace is top-level, no inheritance
		return directMembers, nil

	case EntityTypeSpace:
		// Inherit from workspace
		space, err := s.spaceRepo.FindByID(ctx, entityID)
		if err != nil || space == nil {
			return directMembers, nil
		}
		
		workspaceMembers, err := s.workspaceRepo.FindMembers(ctx, space.WorkspaceID)
		if err == nil {
			for _, wm := range workspaceMembers {
				if _, exists := memberMap[wm.UserID]; !exists {
					memberMap[wm.UserID] = &UnifiedMember{
						ID:            wm.ID,
						EntityType:    EntityTypeSpace,
						EntityID:      entityID,
						UserID:        wm.UserID,
						Role:          wm.Role,
						JoinedAt:      wm.JoinedAt,
						User:          wm.User,
						IsInherited:   true,
						InheritedFrom: EntityTypeWorkspace,
					}
				}
			}
		}

	case EntityTypeFolder:
		// Inherit from space → workspace
		folder, err := s.folderRepo.FindByID(ctx, entityID)
		if err != nil || folder == nil {
			return directMembers, nil
		}

		// Space members
		spaceMembers, err := s.spaceRepo.FindMembers(ctx, folder.SpaceID)
		if err == nil {
			for _, sm := range spaceMembers {
				if _, exists := memberMap[sm.UserID]; !exists {
					memberMap[sm.UserID] = &UnifiedMember{
						ID:            sm.ID,
						EntityType:    EntityTypeFolder,
						EntityID:      entityID,
						UserID:        sm.UserID,
						Role:          sm.Role,
						JoinedAt:      sm.JoinedAt,
						User:          sm.User,
						IsInherited:   true,
						InheritedFrom: EntityTypeSpace,
					}
				}
			}
		}

		// Workspace members
		space, _ := s.spaceRepo.FindByID(ctx, folder.SpaceID)
		if space != nil {
			workspaceMembers, err := s.workspaceRepo.FindMembers(ctx, space.WorkspaceID)
			if err == nil {
				for _, wm := range workspaceMembers {
					if _, exists := memberMap[wm.UserID]; !exists {
						memberMap[wm.UserID] = &UnifiedMember{
							ID:            wm.ID,
							EntityType:    EntityTypeFolder,
							EntityID:      entityID,
							UserID:        wm.UserID,
							Role:          wm.Role,
							JoinedAt:      wm.JoinedAt,
							User:          wm.User,
							IsInherited:   true,
							InheritedFrom: EntityTypeWorkspace,
						}
					}
				}
			}
		}

	case EntityTypeProject:
		// Inherit from folder → space → workspace
		project, err := s.projectRepo.FindByID(ctx, entityID)
		if err != nil || project == nil {
			return directMembers, nil
		}

		// Folder members (if project in folder)
		if project.FolderID != nil {
			folderMembers, err := s.folderRepo.FindMembers(ctx, *project.FolderID)
			if err == nil {
				for _, fm := range folderMembers {
					if _, exists := memberMap[fm.UserID]; !exists {
						memberMap[fm.UserID] = &UnifiedMember{
							ID:            fm.ID,
							EntityType:    EntityTypeProject,
							EntityID:      entityID,
							UserID:        fm.UserID,
							Role:          fm.Role,
							JoinedAt:      fm.JoinedAt,
							User:          fm.User,
							IsInherited:   true,
							InheritedFrom: EntityTypeFolder,
						}
					}
				}
			}
		}

		// Space members
		spaceMembers, err := s.spaceRepo.FindMembers(ctx, project.SpaceID)
		if err == nil {
			for _, sm := range spaceMembers {
				if _, exists := memberMap[sm.UserID]; !exists {
					memberMap[sm.UserID] = &UnifiedMember{
						ID:            sm.ID,
						EntityType:    EntityTypeProject,
						EntityID:      entityID,
						UserID:        sm.UserID,
						Role:          sm.Role,
						JoinedAt:      sm.JoinedAt,
						User:          sm.User,
						IsInherited:   true,
						InheritedFrom: EntityTypeSpace,
					}
				}
			}
		}

		// Workspace members
		space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
		if space != nil {
			workspaceMembers, err := s.workspaceRepo.FindMembers(ctx, space.WorkspaceID)
			if err == nil {
				for _, wm := range workspaceMembers {
					if _, exists := memberMap[wm.UserID]; !exists {
						memberMap[wm.UserID] = &UnifiedMember{
							ID:            wm.ID,
							EntityType:    EntityTypeProject,
							EntityID:      entityID,
							UserID:        wm.UserID,
							Role:          wm.Role,
							JoinedAt:      wm.JoinedAt,
							User:          wm.User,
							IsInherited:   true,
							InheritedFrom: EntityTypeWorkspace,
						}
					}
				}
			}
		}
	}

	// Convert map back to slice
	result := make([]*UnifiedMember, 0, len(memberMap))
	for _, m := range memberMap {
		result = append(result, m)
	}

	return result, nil
}

// ============================================
// ACCESS CONTROL (CASCADE CHECKING)
// ============================================

// HasDirectAccess checks if user is a direct member
func (s *memberService) HasDirectAccess(ctx context.Context, entityType, entityID, userID string) (bool, error) {
	switch entityType {
	case EntityTypeWorkspace:
		return s.workspaceRepo.HasAccess(ctx, entityID, userID)
	case EntityTypeSpace:
		return s.spaceRepo.HasAccess(ctx, entityID, userID)
	case EntityTypeFolder:
		return s.folderRepo.HasAccess(ctx, entityID, userID)
	case EntityTypeProject:
		return s.projectRepo.HasAccess(ctx, entityID, userID)
	default:
		return false, ErrInvalidEntityType
	}
}

// HasEffectiveAccess checks if user has access (direct or inherited)
// Returns: (hasAccess, inheritedFrom, error)
func (s *memberService) HasEffectiveAccess(ctx context.Context, entityType, entityID, userID string) (bool, string, error) {
	// Check direct access first
	hasDirect, err := s.HasDirectAccess(ctx, entityType, entityID, userID)
	if err != nil {
		return false, "", err
	}
	if hasDirect {
		return true, "", nil // Direct access, not inherited
	}

	// Check inherited access based on entity type
	switch entityType {
	case EntityTypeWorkspace:
		// Workspace is top-level, no inheritance
		return false, "", nil

	case EntityTypeSpace:
		// Check workspace membership
		space, err := s.spaceRepo.FindByID(ctx, entityID)
		if err != nil || space == nil {
			return false, "", err
		}
		hasWorkspaceAccess, _ := s.workspaceRepo.HasAccess(ctx, space.WorkspaceID, userID)
		if hasWorkspaceAccess {
			return true, EntityTypeWorkspace, nil
		}
		return false, "", nil

	case EntityTypeFolder:
		// Check space → workspace
		folder, err := s.folderRepo.FindByID(ctx, entityID)
		if err != nil || folder == nil {
			return false, "", err
		}

		// Check space
		hasSpaceAccess, _ := s.spaceRepo.HasAccess(ctx, folder.SpaceID, userID)
		if hasSpaceAccess {
			return true, EntityTypeSpace, nil
		}

		// Check workspace
		space, _ := s.spaceRepo.FindByID(ctx, folder.SpaceID)
		if space != nil {
			hasWorkspaceAccess, _ := s.workspaceRepo.HasAccess(ctx, space.WorkspaceID, userID)
			if hasWorkspaceAccess {
				return true, EntityTypeWorkspace, nil
			}
		}
		return false, "", nil

	case EntityTypeProject:
		// Check folder → space → workspace
		project, err := s.projectRepo.FindByID(ctx, entityID)
		if err != nil || project == nil {
			return false, "", err
		}

		// Check folder (if exists)
		if project.FolderID != nil {
			hasFolderAccess, _ := s.folderRepo.HasAccess(ctx, *project.FolderID, userID)
			if hasFolderAccess {
				return true, EntityTypeFolder, nil
			}
		}

		// Check space
		hasSpaceAccess, _ := s.spaceRepo.HasAccess(ctx, project.SpaceID, userID)
		if hasSpaceAccess {
			return true, EntityTypeSpace, nil
		}

		// Check workspace
		space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
		if space != nil {
			hasWorkspaceAccess, _ := s.workspaceRepo.HasAccess(ctx, space.WorkspaceID, userID)
			if hasWorkspaceAccess {
				return true, EntityTypeWorkspace, nil
			}
		}
		return false, "", nil

	default:
		return false, "", ErrInvalidEntityType
	}
}

// GetAccessLevel returns user's role and where it comes from
// Returns: (role, inheritedFrom, error)
func (s *memberService) GetAccessLevel(ctx context.Context, entityType, entityID, userID string) (string, string, error) {
	// Check direct membership first
	member, err := s.GetMember(ctx, entityType, entityID, userID)
	if err == nil && member != nil {
		return member.Role, "", nil // Direct member
	}

	// Check inherited access
	hasAccess, inheritedFrom, err := s.HasEffectiveAccess(ctx, entityType, entityID, userID)
	if err != nil || !hasAccess {
		return "", "", ErrUnauthorized
	}

	// Get role from parent entity
	switch inheritedFrom {
	case EntityTypeWorkspace:
		var workspaceID string
		switch entityType {
		case EntityTypeSpace:
			space, _ := s.spaceRepo.FindByID(ctx, entityID)
			if space != nil {
				workspaceID = space.WorkspaceID
			}
		case EntityTypeFolder:
			folder, _ := s.folderRepo.FindByID(ctx, entityID)
			if folder != nil {
				space, _ := s.spaceRepo.FindByID(ctx, folder.SpaceID)
				if space != nil {
					workspaceID = space.WorkspaceID
				}
			}
		case EntityTypeProject:
			project, _ := s.projectRepo.FindByID(ctx, entityID)
			if project != nil {
				space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
				if space != nil {
					workspaceID = space.WorkspaceID
				}
			}
		}
		if workspaceID != "" {
			wsMember, _ := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
			if wsMember != nil {
				return wsMember.Role, EntityTypeWorkspace, nil
			}
		}

	case EntityTypeSpace:
		var spaceID string
		switch entityType {
		case EntityTypeFolder:
			folder, _ := s.folderRepo.FindByID(ctx, entityID)
			if folder != nil {
				spaceID = folder.SpaceID
			}
		case EntityTypeProject:
			project, _ := s.projectRepo.FindByID(ctx, entityID)
			if project != nil {
				spaceID = project.SpaceID
			}
		}
		if spaceID != "" {
			spaceMember, _ := s.spaceRepo.FindMember(ctx, spaceID, userID)
			if spaceMember != nil {
				return spaceMember.Role, EntityTypeSpace, nil
			}
		}

	case EntityTypeFolder:
		project, _ := s.projectRepo.FindByID(ctx, entityID)
		if project != nil && project.FolderID != nil {
			folderMember, _ := s.folderRepo.FindMember(ctx, *project.FolderID, userID)
			if folderMember != nil {
				return folderMember.Role, EntityTypeFolder, nil
			}
		}
	}

	return "", "", ErrUnauthorized
}

// ============================================
// INVITATION
// ============================================

func (s *memberService) InviteMemberByEmail(ctx context.Context, entityType, entityID, email, role, inviterID string) error {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	return s.AddMember(ctx, entityType, entityID, user.ID, role, inviterID)
}

func (s *memberService) InviteMemberByID(ctx context.Context, entityType, entityID, userID, role, inviterID string) error {
	return s.AddMember(ctx, entityType, entityID, userID, role, inviterID)
}

// ============================================
// USER MEMBERSHIPS
// ============================================

func (s *memberService) GetUserMemberships(ctx context.Context, userID string) (map[string][]string, error) {
	result := make(map[string][]string)

	// Get workspaces
	workspaces, _ := s.workspaceRepo.FindByUserID(ctx, userID)
	workspaceIDs := make([]string, len(workspaces))
	for i, ws := range workspaces {
		workspaceIDs[i] = ws.ID
	}
	result[EntityTypeWorkspace] = workspaceIDs

	// Get spaces
	spaces, _ := s.spaceRepo.FindByUserID(ctx, userID)
	spaceIDs := make([]string, len(spaces))
	for i, sp := range spaces {
		spaceIDs[i] = sp.ID
	}
	result[EntityTypeSpace] = spaceIDs

	// Get folders
	folders, _ := s.folderRepo.FindByUserID(ctx, userID)
	folderIDs := make([]string, len(folders))
	for i, f := range folders {
		folderIDs[i] = f.ID
	}
	result[EntityTypeFolder] = folderIDs

	// Get projects
	projects, _ := s.projectRepo.FindByUserID(ctx, userID)
	projectIDs := make([]string, len(projects))
	for i, p := range projects {
		projectIDs[i] = p.ID
	}
	result[EntityTypeProject] = projectIDs

	return result, nil
}

func (s *memberService) GetUserAllAccess(ctx context.Context, userID string) (*UserAccessMap, error) {
	accessMap := &UserAccessMap{
		Workspaces: []string{},
		Spaces:     []string{},
		Folders:    []string{},
		Projects:   []string{},
	}

	// Direct memberships
	memberships, err := s.GetUserMemberships(ctx, userID)
	if err != nil {
		return nil, err
	}

	accessMap.Workspaces = memberships[EntityTypeWorkspace]
	accessMap.Spaces = memberships[EntityTypeSpace]
	accessMap.Folders = memberships[EntityTypeFolder]
	accessMap.Projects = memberships[EntityTypeProject]

	// TODO: Add inherited access (spaces from workspace membership, etc.)
	// This would require iterating through all entities which could be expensive
	// Better to check access on-demand when needed

	return accessMap, nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

func (s *memberService) sendNotification(ctx context.Context, entityType, entityID, userID, inviterID string) {
	if s.notifSvc == nil {
		return
	}

	inviterName := ""
	if inviter, _ := s.userRepo.FindByID(ctx, inviterID); inviter != nil {
		inviterName = inviter.Name
	}

	var entityName string
	switch entityType {
	case EntityTypeWorkspace:
		if ws, _ := s.workspaceRepo.FindByID(ctx, entityID); ws != nil {
			entityName = ws.Name
		}
	case EntityTypeSpace:
		if sp, _ := s.spaceRepo.FindByID(ctx, entityID); sp != nil {
			entityName = sp.Name
		}
	case EntityTypeFolder:
		if f, _ := s.folderRepo.FindByID(ctx, entityID); f != nil {
			entityName = f.Name
		}
	case EntityTypeProject:
		if p, _ := s.projectRepo.FindByID(ctx, entityID); p != nil {
			entityName = p.Name
		}
	}

	// Send appropriate notification based on entity type
	switch entityType {
	case EntityTypeWorkspace:
		s.notifSvc.SendWorkspaceInvitation(ctx, userID, entityName, entityID, inviterName)
	case EntityTypeProject:
		s.notifSvc.SendProjectInvitation(ctx, userID, entityName, entityID, inviterName)
	// Add other notification types as needed
	}
}

// Converter functions
func (s *memberService) convertWorkspaceMembers(members []*repository.WorkspaceMember, entityID string, inherited bool) []*UnifiedMember {
	result := make([]*UnifiedMember, len(members))
	for i, m := range members {
		result[i] = &UnifiedMember{
			ID:          m.ID,
			EntityType:  EntityTypeWorkspace,
			EntityID:    entityID,
			UserID:      m.UserID,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt,
			User:        m.User,
			IsInherited: inherited,
		}
	}
	return result
}

func (s *memberService) convertSpaceMembers(members []*repository.SpaceMember, entityID string, inherited bool) []*UnifiedMember {
	result := make([]*UnifiedMember, len(members))
	for i, m := range members {
		result[i] = &UnifiedMember{
			ID:          m.ID,
			EntityType:  EntityTypeSpace,
			EntityID:    entityID,
			UserID:      m.UserID,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt,
			User:        m.User,
			IsInherited: inherited,
		}
	}
	return result
}

func (s *memberService) convertFolderMembers(members []*repository.FolderMember, entityID string, inherited bool) []*UnifiedMember {
	result := make([]*UnifiedMember, len(members))
	for i, m := range members {
		result[i] = &UnifiedMember{
			ID:          m.ID,
			EntityType:  EntityTypeFolder,
			EntityID:    entityID,
			UserID:      m.UserID,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt,
			User:        m.User,
			IsInherited: inherited,
		}
	}
	return result
}

func (s *memberService) convertProjectMembers(members []*repository.ProjectMember, entityID string, inherited bool) []*UnifiedMember {
	result := make([]*UnifiedMember, len(members))
	for i, m := range members {
		result[i] = &UnifiedMember{
			ID:          m.ID,
			EntityType:  EntityTypeProject,
			EntityID:    entityID,
			UserID:      m.UserID,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt,
			User:        m.User,
			IsInherited: inherited,
		}
	}
	return result
}



// GetAccessibleWorkspaces returns all workspaces user can access (direct members)
func (s *memberService) GetAccessibleWorkspaces(ctx context.Context, userID string) ([]*repository.Workspace, error) {
	return s.workspaceRepo.FindByUserID(ctx, userID)
}

// GetAccessibleSpaces returns all spaces user can access (direct + from workspace)
func (s *memberService) GetAccessibleSpaces(ctx context.Context, userID string) ([]*repository.Space, error) {
	spaceMap := make(map[string]*repository.Space)
	
	// 1. Get spaces where user is direct member
	directSpaces, _ := s.spaceRepo.FindByUserID(ctx, userID)
	for _, space := range directSpaces {
		spaceMap[space.ID] = space
	}
	
	// 2. Get spaces from workspace membership
	workspaces, _ := s.workspaceRepo.FindByUserID(ctx, userID)
	for _, ws := range workspaces {
		workspaceSpaces, _ := s.spaceRepo.FindByWorkspaceID(ctx, ws.ID)
		for _, space := range workspaceSpaces {
			if _, exists := spaceMap[space.ID]; !exists {
				spaceMap[space.ID] = space
			}
		}
	}
	
	// Convert map to slice
	result := make([]*repository.Space, 0, len(spaceMap))
	for _, space := range spaceMap {
		result = append(result, space)
	}
	
	return result, nil
}

// GetAccessibleFolders returns all folders user can access
func (s *memberService) GetAccessibleFolders(ctx context.Context, userID string) ([]*repository.Folder, error) {
	folderMap := make(map[string]*repository.Folder)
	
	// 1. Direct folder members
	directFolders, _ := s.folderRepo.FindByUserID(ctx, userID)
	for _, folder := range directFolders {
		folderMap[folder.ID] = folder
	}
	
	// 2. Folders from space membership
	spaces, _ := s.GetAccessibleSpaces(ctx, userID)
	for _, space := range spaces {
		spaceFolders, _ := s.folderRepo.FindBySpaceID(ctx, space.ID)
		for _, folder := range spaceFolders {
			if _, exists := folderMap[folder.ID]; !exists {
				folderMap[folder.ID] = folder
			}
		}
	}
	
	result := make([]*repository.Folder, 0, len(folderMap))
	for _, folder := range folderMap {
		result = append(result, folder)
	}
	
	return result, nil
}

// GetAccessibleProjects returns all projects user can access (direct + inherited)
func (s *memberService) GetAccessibleProjects(ctx context.Context, userID string) ([]*repository.Project, error) {
	projectMap := make(map[string]*repository.Project)
	
	// 1. Direct project members
	directProjects, _ := s.projectRepo.FindByUserID(ctx, userID)
	for _, proj := range directProjects {
		projectMap[proj.ID] = proj
	}
	
	// 2. Projects from folder membership
	folders, _ := s.GetAccessibleFolders(ctx, userID)
	for _, folder := range folders {
		folderProjects, _ := s.projectRepo.FindByFolderID(ctx, folder.ID)
		for _, proj := range folderProjects {
			if _, exists := projectMap[proj.ID]; !exists {
				projectMap[proj.ID] = proj
			}
		}
	}
	
	// 3. Projects from space membership
	spaces, _ := s.GetAccessibleSpaces(ctx, userID)
	for _, space := range spaces {
		spaceProjects, _ := s.projectRepo.FindBySpaceID(ctx, space.ID)
		for _, proj := range spaceProjects {
			if _, exists := projectMap[proj.ID]; !exists {
				projectMap[proj.ID] = proj
			}
		}
	}
	
	result := make([]*repository.Project, 0, len(projectMap))
	for _, proj := range projectMap {
		result = append(result, proj)
	}
	
	return result, nil
}
