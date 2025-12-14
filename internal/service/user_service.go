package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// User Service
// ============================================

type UserService interface {
	GetByID(ctx context.Context, id string) (*repository.User, error)
	GetByEmail(ctx context.Context, email string) (*repository.User, error)
	Update(ctx context.Context, id string, name, avatar *string) (*repository.User, error)
	UpdateLastActive(ctx context.Context, id string) error
	Search(ctx context.Context, query string) ([]*repository.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetByID(ctx context.Context, id string) (*repository.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*repository.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, id string, name, avatar *string) (*repository.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if name != nil {
		user.Name = *name
	}
	if avatar != nil {
		user.Avatar = avatar
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateLastActive(ctx context.Context, id string) error {
	return s.userRepo.UpdateLastActive(ctx, id)
}

func (s *userService) Search(ctx context.Context, query string) ([]*repository.User, error) {
	return s.userRepo.Search(ctx, query)
}
