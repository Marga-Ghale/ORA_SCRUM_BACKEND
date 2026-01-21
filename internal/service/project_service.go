package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
)

// ============================================
// Project Service
// ============================================

type ProjectService interface {
	// Project CRUD
	Create(ctx context.Context, spaceID string, folderID *string, creatorID, name, key string, description, icon, color, leadID *string) (*repository.Project, error)
	GetByID(ctx context.Context, id string) (*repository.Project, error)
	GetByKey(ctx context.Context, spaceID, key string) (*repository.Project, error)
	ListBySpace(ctx context.Context, spaceID string) ([]*repository.Project, error)
	ListByFolder(ctx context.Context, folderID string) ([]*repository.Project, error)
	Update(ctx context.Context, id string, name, key, description, icon, color, leadID *string, folderID *string) (*repository.Project, error)
	Delete(ctx context.Context, id string) error

	// Project-specific operations (not member management)
	MoveToFolder(ctx context.Context, projectID string, folderID *string) error
	SetLead(ctx context.Context, projectID, leadID string) error
	UpdateVisibility(ctx context.Context, projectID, visibility string, allowedUsers, allowedTeams []string) error
}

type projectService struct {
	projectRepo   repository.ProjectRepository
	spaceRepo     repository.SpaceRepository
	folderRepo    repository.FolderRepository
	memberService MemberService
	broadcaster   *socket.Broadcaster // ✅ NEW: Added broadcaster
}

func NewProjectService(
	projectRepo repository.ProjectRepository,
	spaceRepo repository.SpaceRepository,
	folderRepo repository.FolderRepository,
	memberService MemberService,
) ProjectService {
	return &projectService{
		projectRepo:   projectRepo,
		spaceRepo:     spaceRepo,
		folderRepo:    folderRepo,
		memberService: memberService,
	}
}

// ✅ NEW: SetBroadcaster sets the broadcaster for real-time updates
func (s *projectService) SetBroadcaster(b *socket.Broadcaster) {
	s.broadcaster = b
}

func (s *projectService) Create(ctx context.Context, spaceID string, folderID *string, creatorID, name, key string, description, icon, color, leadID *string) (*repository.Project, error) {
	// Verify space exists
	space, err := s.spaceRepo.FindByID(ctx, spaceID)
	if err != nil || space == nil {
		return nil, ErrNotFound
	}

	// Verify folder exists if provided
	if folderID != nil && *folderID != "" {
		folder, err := s.folderRepo.FindByID(ctx, *folderID)
		if err != nil {
			return nil, err
		}
		if folder == nil {
			return nil, ErrNotFound
		}
		// Verify folder belongs to the same space
		if folder.SpaceID != spaceID {
			return nil, ErrInvalidInput
		}
	}

	// Check if key already exists in this space
	projects, _ := s.projectRepo.FindBySpaceID(ctx, spaceID)
	for _, p := range projects {
		if p.Key == key {
			return nil, ErrConflict
		}
	}

	// ✅ Set default lead to creator if not provided
	finalLeadID := leadID
	if finalLeadID == nil {
		finalLeadID = &creatorID
	}

	// Set default visibility
	defaultVisibility := "private"

	project := &repository.Project{
		SpaceID:     spaceID,
		FolderID:    folderID,
		Name:        name,
		Key:         key,
		Description: description,
		Icon:        icon,
		Color:       color,
		LeadID:      finalLeadID,
		Visibility:  &defaultVisibility,
		CreatedBy:   &creatorID,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	// ✅ Use MemberService to add creator as project lead
	if err := s.memberService.AddMember(ctx, EntityTypeProject, project.ID, creatorID, "lead", creatorID); err != nil {
		// If member add fails, rollback project creation
		s.projectRepo.Delete(ctx, project.ID)
		return nil, err
	}

	// ✅ NEW: Broadcast project creation to workspace members
	if s.broadcaster != nil {
		s.broadcaster.BroadcastProjectCreated(space.WorkspaceID, spaceID, folderID, map[string]interface{}{
			"id":          project.ID,
			"spaceId":     project.SpaceID,
			"folderId":    project.FolderID,
			"name":        project.Name,
			"key":         project.Key,
			"description": project.Description,
			"icon":        project.Icon,
			"color":       project.Color,
			"leadId":      project.LeadID,
			"visibility":  project.Visibility,
			"createdBy":   project.CreatedBy,
			"createdAt":   project.CreatedAt,
			"updatedAt":   project.UpdatedAt,
		}, creatorID)
	}

	return project, nil
}

func (s *projectService) GetByID(ctx context.Context, id string) (*repository.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrNotFound
	}
	return project, nil
}

func (s *projectService) GetByKey(ctx context.Context, spaceID, key string) (*repository.Project, error) {
	projects, err := s.projectRepo.FindBySpaceID(ctx, spaceID)
	if err != nil {
		return nil, err
	}

	for _, p := range projects {
		if p.Key == key {
			return p, nil
		}
	}

	return nil, ErrNotFound
}

func (s *projectService) ListBySpace(ctx context.Context, spaceID string) ([]*repository.Project, error) {
	return s.projectRepo.FindBySpaceID(ctx, spaceID)
}

func (s *projectService) ListByFolder(ctx context.Context, folderID string) ([]*repository.Project, error) {
	return s.projectRepo.FindByFolderID(ctx, folderID)
}

func (s *projectService) Update(ctx context.Context, id string, name, key, description, icon, color, leadID *string, folderID *string) (*repository.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil || project == nil {
		return nil, ErrNotFound
	}

	// Update name if provided
	if name != nil {
		project.Name = *name
	}

	// Update key if provided (check uniqueness in space)
	if key != nil && *key != project.Key {
		projects, _ := s.projectRepo.FindBySpaceID(ctx, project.SpaceID)
		for _, p := range projects {
			if p.Key == *key && p.ID != id {
				return nil, ErrConflict
			}
		}
		project.Key = *key
	}

	// Update folder if provided (verify it belongs to same space)
	if folderID != nil {
		if *folderID != "" {
			folder, err := s.folderRepo.FindByID(ctx, *folderID)
			if err != nil || folder == nil {
				return nil, ErrNotFound
			}
			if folder.SpaceID != project.SpaceID {
				return nil, ErrInvalidInput
			}
		}
		project.FolderID = folderID
	}

	// Nullable fields - always update to allow clearing
	project.Description = description
	project.Icon = icon
	project.Color = color
	project.LeadID = leadID

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	// ✅ NEW: Broadcast project update to workspace members
	if s.broadcaster != nil {
		space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
		if space != nil {
			s.broadcaster.BroadcastProjectUpdated(space.WorkspaceID, map[string]interface{}{
				"id":          project.ID,
				"spaceId":     project.SpaceID,
				"folderId":    project.FolderID,
				"name":        project.Name,
				"key":         project.Key,
				"description": project.Description,
				"icon":        project.Icon,
				"color":       project.Color,
				"leadId":      project.LeadID,
				"visibility":  project.Visibility,
				"updatedAt":   project.UpdatedAt,
			}, "")
		}
	}

	return project, nil
}

func (s *projectService) Delete(ctx context.Context, id string) error {
	// ✅ Get project first to know space ID for broadcasting
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil || project == nil {
		return ErrNotFound
	}

	spaceID := project.SpaceID
	folderID := project.FolderID

	// Get space to find workspace ID
	space, _ := s.spaceRepo.FindByID(ctx, spaceID)

	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return err
	}

	// ✅ NEW: Broadcast project deletion to workspace members
	if s.broadcaster != nil && space != nil {
		s.broadcaster.BroadcastProjectDeleted(space.WorkspaceID, spaceID, id, folderID, "")
	}

	return nil
}

func (s *projectService) MoveToFolder(ctx context.Context, projectID string, folderID *string) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrNotFound
	}

	// If moving to a folder, verify it exists and belongs to same space
	if folderID != nil && *folderID != "" {
		folder, err := s.folderRepo.FindByID(ctx, *folderID)
		if err != nil || folder == nil {
			return ErrNotFound
		}
		if folder.SpaceID != project.SpaceID {
			return ErrInvalidInput
		}
	}

	project.FolderID = folderID

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return err
	}

	// ✅ NEW: Broadcast project move
	if s.broadcaster != nil {
		space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
		if space != nil {
			s.broadcaster.BroadcastProjectUpdated(space.WorkspaceID, map[string]interface{}{
				"id":        project.ID,
				"spaceId":   project.SpaceID,
				"folderId":  project.FolderID,
				"name":      project.Name,
				"key":       project.Key,
				"updatedAt": project.UpdatedAt,
			}, "")
		}
	}

	return nil
}

func (s *projectService) SetLead(ctx context.Context, projectID, leadID string) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrNotFound
	}

	// Verify user is a member of the project
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeProject, projectID, leadID)
	if err != nil || !hasAccess {
		return ErrUnauthorized
	}

	project.LeadID = &leadID

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return err
	}

	// ✅ NEW: Broadcast lead change
	if s.broadcaster != nil {
		space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
		if space != nil {
			s.broadcaster.BroadcastProjectUpdated(space.WorkspaceID, map[string]interface{}{
				"id":        project.ID,
				"spaceId":   project.SpaceID,
				"leadId":    project.LeadID,
				"name":      project.Name,
				"key":       project.Key,
				"updatedAt": project.UpdatedAt,
			}, "")
		}
	}

	return nil
}

func (s *projectService) UpdateVisibility(ctx context.Context, projectID, visibility string, allowedUsers, allowedTeams []string) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrNotFound
	}

	project.Visibility = &visibility
	project.AllowedUsers = allowedUsers
	project.AllowedTeams = allowedTeams

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return err
	}

	// ✅ NEW: Broadcast visibility update
	if s.broadcaster != nil {
		space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
		if space != nil {
			s.broadcaster.BroadcastProjectUpdated(space.WorkspaceID, map[string]interface{}{
				"id":         project.ID,
				"spaceId":    project.SpaceID,
				"name":       project.Name,
				"key":        project.Key,
				"visibility": project.Visibility,
				"updatedAt":  project.UpdatedAt,
			}, "")
		}
	}

	return nil
}