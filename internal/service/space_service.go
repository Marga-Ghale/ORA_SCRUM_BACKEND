package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/socket"
)

type SpaceService interface {
	// Space CRUD
	Create(ctx context.Context, workspaceID, creatorID, name string, description, icon, color *string) (*repository.Space, error)
	GetByID(ctx context.Context, id string) (*repository.Space, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*repository.Space, error)
	ListByUser(ctx context.Context, userID string) ([]*repository.Space, error)
	Update(ctx context.Context, id string, name, description, icon, color, visibility *string, allowedUsers, allowedTeams *[]string) (*repository.Space, error)
	Delete(ctx context.Context, id string) error

	// Space-specific operations (not member management)
	UpdateVisibility(ctx context.Context, spaceID, visibility string, allowedUsers, allowedTeams []string) error
}

type spaceService struct {
	spaceRepo     repository.SpaceRepository
	workspaceRepo repository.WorkspaceRepository
	memberService MemberService
	broadcaster   *socket.Broadcaster // ✅ NEW: Added broadcaster
}

func NewSpaceService(
	spaceRepo repository.SpaceRepository,
	workspaceRepo repository.WorkspaceRepository,
	memberService MemberService,
	broadcaster *socket.Broadcaster, // ✅ ADD

) SpaceService {
	return &spaceService{
		spaceRepo:     spaceRepo,
		workspaceRepo: workspaceRepo,
		memberService: memberService,
		broadcaster:   broadcaster,
	}
}

// ✅ NEW: SetBroadcaster sets the broadcaster for real-time updates
func (s *spaceService) SetBroadcaster(b *socket.Broadcaster) {
	s.broadcaster = b
}

func (s *spaceService) Create(ctx context.Context, workspaceID, creatorID, name string, description, icon, color *string) (*repository.Space, error) {
	// ✅ Verify workspace exists
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return nil, ErrNotFound
	}

	// ✅ Verify creator has access to workspace
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeWorkspace, workspaceID, creatorID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Set default visibility
	defaultVisibility := "private"

	space := &repository.Space{
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		Icon:        icon,
		Color:       color,
		OwnerID:     creatorID,
		Visibility:  &defaultVisibility,
	}

	if err := s.spaceRepo.Create(ctx, space); err != nil {
		return nil, err
	}

	// ✅ Use MemberService to add creator as space owner
	if err := s.memberService.AddMember(ctx, EntityTypeSpace, space.ID, creatorID, "owner", creatorID); err != nil {
		// If member add fails, rollback space creation
		s.spaceRepo.Delete(ctx, space.ID)
		return nil, err
	}

	// ✅ NEW: Broadcast space creation to workspace members
	if s.broadcaster != nil {
		s.broadcaster.BroadcastSpaceCreated(workspaceID, map[string]interface{}{
			"id":          space.ID,
			"workspaceId": space.WorkspaceID,
			"name":        space.Name,
			"description": space.Description,
			"icon":        space.Icon,
			"color":       space.Color,
			"ownerId":     space.OwnerID,
			"visibility":  space.Visibility,
			"createdAt":   space.CreatedAt,
			"updatedAt":   space.UpdatedAt,
		}, creatorID)
	}

	return space, nil
}

func (s *spaceService) GetByID(ctx context.Context, id string) (*repository.Space, error) {
	space, err := s.spaceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if space == nil {
		return nil, ErrNotFound
	}
	return space, nil
}

func (s *spaceService) ListByWorkspace(ctx context.Context, workspaceID string) ([]*repository.Space, error) {
	return s.spaceRepo.FindByWorkspaceID(ctx, workspaceID)
}

func (s *spaceService) ListByUser(ctx context.Context, userID string) ([]*repository.Space, error) {
	return s.spaceRepo.FindByUserID(ctx, userID)
}

func (s *spaceService) Update(ctx context.Context, id string, name, description, icon, color, visibility *string, allowedUsers, allowedTeams *[]string) (*repository.Space, error) {
	space, err := s.spaceRepo.FindByID(ctx, id)
	if err != nil || space == nil {
		return nil, ErrNotFound
	}

	// Update name if provided
	if name != nil {
		space.Name = *name
	}

	// Nullable fields - always update to allow clearing
	space.Description = description
	space.Icon = icon
	space.Color = color
	space.Visibility = visibility

	if allowedUsers != nil {
		space.AllowedUsers = *allowedUsers
	}
	if allowedTeams != nil {
		space.AllowedTeams = *allowedTeams
	}

	if err := s.spaceRepo.Update(ctx, space); err != nil {
		return nil, err
	}

	// ✅ NEW: Broadcast space update to workspace members
	if s.broadcaster != nil {
		s.broadcaster.BroadcastSpaceUpdated(space.WorkspaceID, map[string]interface{}{
			"id":          space.ID,
			"workspaceId": space.WorkspaceID,
			"name":        space.Name,
			"description": space.Description,
			"icon":        space.Icon,
			"color":       space.Color,
			"ownerId":     space.OwnerID,
			"visibility":  space.Visibility,
			"updatedAt":   space.UpdatedAt,
		}, "") // Don't exclude anyone on updates - everyone should see
	}

	return space, nil
}

func (s *spaceService) Delete(ctx context.Context, id string) error {
	// ✅ Get space first to know workspace ID for broadcasting
	space, err := s.spaceRepo.FindByID(ctx, id)
	if err != nil || space == nil {
		return ErrNotFound
	}

	workspaceID := space.WorkspaceID

	if err := s.spaceRepo.Delete(ctx, id); err != nil {
		return err
	}

	// ✅ NEW: Broadcast space deletion to workspace members
	if s.broadcaster != nil {
		s.broadcaster.BroadcastSpaceDeleted(workspaceID, id, "")
	}

	return nil
}

func (s *spaceService) UpdateVisibility(ctx context.Context, spaceID, visibility string, allowedUsers, allowedTeams []string) error {
	space, err := s.spaceRepo.FindByID(ctx, spaceID)
	if err != nil || space == nil {
		return ErrNotFound
	}

	space.Visibility = &visibility
	space.AllowedUsers = allowedUsers
	space.AllowedTeams = allowedTeams

	if err := s.spaceRepo.Update(ctx, space); err != nil {
		return err
	}

	// ✅ NEW: Broadcast visibility update
	if s.broadcaster != nil {
		s.broadcaster.BroadcastSpaceUpdated(space.WorkspaceID, map[string]interface{}{
			"id":          space.ID,
			"workspaceId": space.WorkspaceID,
			"name":        space.Name,
			"visibility":  space.Visibility,
			"updatedAt":   space.UpdatedAt,
		}, "")
	}

	return nil
}