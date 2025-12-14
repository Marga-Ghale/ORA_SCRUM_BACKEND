package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Activity Service
// ============================================

// ActivityService defines activity log operations
type ActivityService interface {
	LogActivity(ctx context.Context, activityType, entityType, entityID, userID string, changes, metadata map[string]interface{}) error
	GetEntityActivities(ctx context.Context, entityType, entityID string, limit int) ([]*repository.Activity, error)
	GetUserActivities(ctx context.Context, userID string, limit int) ([]*repository.Activity, error)
}

type activityService struct {
	activityRepo repository.ActivityRepository
}

// NewActivityService creates a new activity service
func NewActivityService(activityRepo repository.ActivityRepository) ActivityService {
	return &activityService{activityRepo: activityRepo}
}

func (s *activityService) LogActivity(ctx context.Context, activityType, entityType, entityID, userID string, changes, metadata map[string]interface{}) error {
	activity := &repository.Activity{
		Type:       activityType,
		EntityType: entityType,
		EntityID:   entityID,
		UserID:     userID,
		Changes:    changes,
		Metadata:   metadata,
	}
	return s.activityRepo.Create(ctx, activity)
}

func (s *activityService) GetEntityActivities(ctx context.Context, entityType, entityID string, limit int) ([]*repository.Activity, error) {
	return s.activityRepo.FindByEntity(ctx, entityType, entityID, limit)
}

func (s *activityService) GetUserActivities(ctx context.Context, userID string, limit int) ([]*repository.Activity, error) {
	return s.activityRepo.FindByUser(ctx, userID, limit)
}
