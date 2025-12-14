package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Task Watcher Service
// ============================================

// TaskWatcherService defines task watcher operations
type TaskWatcherService interface {
	Watch(ctx context.Context, taskID, userID string) error
	Unwatch(ctx context.Context, taskID, userID string) error
	GetWatchers(ctx context.Context, taskID string) ([]string, error)
	IsWatching(ctx context.Context, taskID, userID string) (bool, error)
}

type taskWatcherService struct {
	watcherRepo repository.TaskWatcherRepository
}

// NewTaskWatcherService creates a new task watcher service
func NewTaskWatcherService(watcherRepo repository.TaskWatcherRepository) TaskWatcherService {
	return &taskWatcherService{watcherRepo: watcherRepo}
}

func (s *taskWatcherService) Watch(ctx context.Context, taskID, userID string) error {
	watcher := &repository.TaskWatcher{
		TaskID: taskID,
		UserID: userID,
	}
	return s.watcherRepo.Add(ctx, watcher)
}

func (s *taskWatcherService) Unwatch(ctx context.Context, taskID, userID string) error {
	return s.watcherRepo.Remove(ctx, taskID, userID)
}

func (s *taskWatcherService) GetWatchers(ctx context.Context, taskID string) ([]string, error) {
	return s.watcherRepo.GetWatcherUserIDs(ctx, taskID)
}

func (s *taskWatcherService) IsWatching(ctx context.Context, taskID, userID string) (bool, error) {
	return s.watcherRepo.IsWatching(ctx, taskID, userID)
}
