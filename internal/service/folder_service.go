package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

type FolderService interface {
	// Folder CRUD
	Create(ctx context.Context, spaceID, creatorID, name string, description, icon, color *string) (*repository.Folder, error)
	GetByID(ctx context.Context, id string) (*repository.Folder, error)
	ListBySpace(ctx context.Context, spaceID string) ([]*repository.Folder, error)
	ListByUser(ctx context.Context, userID string) ([]*repository.Folder, error)
	Update(ctx context.Context, id string, name, description, icon, color, visibility *string, allowedUsers, allowedTeams *[]string) (*repository.Folder, error)
	Delete(ctx context.Context, id string) error
	
	// Folder-specific operations (not member management)
	UpdateVisibility(ctx context.Context, folderID, visibility string, allowedUsers, allowedTeams []string) error
}

type folderService struct {
	folderRepo    repository.FolderRepository
	spaceRepo     repository.SpaceRepository
	memberService MemberService // ✅ Use MemberService for member operations
}

func NewFolderService(
	folderRepo repository.FolderRepository,
	spaceRepo repository.SpaceRepository,
	memberService MemberService,
) FolderService {
	return &folderService{
		folderRepo:    folderRepo,
		spaceRepo:     spaceRepo,
		memberService: memberService,
	}
}

func (s *folderService) Create(ctx context.Context, spaceID, creatorID, name string, description, icon, color *string) (*repository.Folder, error) {
	// ✅ Verify space exists
	space, err := s.spaceRepo.FindByID(ctx, spaceID)
	if err != nil || space == nil {
		return nil, ErrNotFound
	}

	// ✅ Verify creator has access to space
	hasAccess, _, err := s.memberService.HasEffectiveAccess(ctx, EntityTypeSpace, spaceID, creatorID)
	if err != nil || !hasAccess {
		return nil, ErrUnauthorized
	}

	// Set default visibility
	defaultVisibility := "private"

	folder := &repository.Folder{
		SpaceID:     spaceID, // ✅ CRITICAL - Must set parent space
		Name:        name,
		Description: description,
		Icon:        icon,
		Color:       color,
		OwnerID:     creatorID,
		Visibility:  &defaultVisibility,
	}

	if err := s.folderRepo.Create(ctx, folder); err != nil {
		return nil, err
	}

	// ✅ Use MemberService to add creator as folder owner
	if err := s.memberService.AddMember(ctx, EntityTypeFolder, folder.ID, creatorID, "owner", creatorID); err != nil {
		// If member add fails, rollback folder creation
		s.folderRepo.Delete(ctx, folder.ID)
		return nil, err
	}

	return folder, nil
}

func (s *folderService) GetByID(ctx context.Context, id string) (*repository.Folder, error) {
	folder, err := s.folderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if folder == nil {
		return nil, ErrNotFound
	}
	return folder, nil
}

func (s *folderService) ListBySpace(ctx context.Context, spaceID string) ([]*repository.Folder, error) {
	return s.folderRepo.FindBySpaceID(ctx, spaceID)
}

func (s *folderService) ListByUser(ctx context.Context, userID string) ([]*repository.Folder, error) {
	return s.folderRepo.FindByUserID(ctx, userID)
}

func (s *folderService) Update(ctx context.Context, id string, name, description, icon, color, visibility *string, allowedUsers, allowedTeams *[]string) (*repository.Folder, error) {
	folder, err := s.folderRepo.FindByID(ctx, id)
	if err != nil || folder == nil {
		return nil, ErrNotFound
	}

	// Update name if provided
	if name != nil {
		folder.Name = *name
	}

	// Nullable fields - always update to allow clearing
	folder.Description = description
	folder.Icon = icon
	folder.Color = color
	folder.Visibility = visibility

	if allowedUsers != nil {
		folder.AllowedUsers = *allowedUsers
	}
	if allowedTeams != nil {
		folder.AllowedTeams = *allowedTeams
	}

	if err := s.folderRepo.Update(ctx, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *folderService) Delete(ctx context.Context, id string) error {
	return s.folderRepo.Delete(ctx, id)
}

func (s *folderService) UpdateVisibility(ctx context.Context, folderID, visibility string, allowedUsers, allowedTeams []string) error {
	folder, err := s.folderRepo.FindByID(ctx, folderID)
	if err != nil || folder == nil {
		return ErrNotFound
	}

	folder.Visibility = &visibility
	folder.AllowedUsers = allowedUsers
	folder.AllowedTeams = allowedTeams

	return s.folderRepo.Update(ctx, folder)
}