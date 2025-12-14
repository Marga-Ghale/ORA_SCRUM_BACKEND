package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Notification Service (for handlers)
// ============================================

type NotificationService interface {
	List(ctx context.Context, userID string, unreadOnly bool) ([]*repository.Notification, error)
	Count(ctx context.Context, userID string) (total int, unread int, err error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
	DeleteAll(ctx context.Context, userID string) error
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
}

func NewNotificationService(notificationRepo repository.NotificationRepository) NotificationService {
	return &notificationService{notificationRepo: notificationRepo}
}

func (s *notificationService) List(ctx context.Context, userID string, unreadOnly bool) ([]*repository.Notification, error) {
	return s.notificationRepo.FindByUserID(ctx, userID, unreadOnly)
}

func (s *notificationService) Count(ctx context.Context, userID string) (total int, unread int, err error) {
	return s.notificationRepo.CountByUserID(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id string) error {
	return s.notificationRepo.MarkAsRead(ctx, id)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) Delete(ctx context.Context, id string) error {
	return s.notificationRepo.Delete(ctx, id)
}

func (s *notificationService) DeleteAll(ctx context.Context, userID string) error {
	return s.notificationRepo.DeleteAll(ctx, userID)
}
