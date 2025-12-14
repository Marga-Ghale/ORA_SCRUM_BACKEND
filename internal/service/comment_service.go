package service

import (
	"context"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/notification"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Comment Service
// ============================================

type CommentService interface {
	Create(ctx context.Context, taskID, userID, content string) (*repository.Comment, error)
	GetByID(ctx context.Context, id string) (*repository.Comment, error)
	ListByTask(ctx context.Context, taskID string) ([]*repository.Comment, error)
	Update(ctx context.Context, id, userID, content string) (*repository.Comment, error)
	Delete(ctx context.Context, id, userID string) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	taskRepo    repository.TaskRepository
	userRepo    repository.UserRepository
	notifSvc    *notification.Service
}

func NewCommentService(commentRepo repository.CommentRepository, taskRepo repository.TaskRepository, userRepo repository.UserRepository, notifSvc *notification.Service) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		taskRepo:    taskRepo,
		userRepo:    userRepo,
		notifSvc:    notifSvc,
	}
}

func (s *commentService) Create(ctx context.Context, taskID, userID, content string) (*repository.Comment, error) {
	comment := &repository.Comment{
		TaskID:  taskID,
		UserID:  userID,
		Content: content,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// Get commenter info for notifications
	commenter, _ := s.userRepo.FindByID(ctx, userID)
	commenterName := "Someone"
	if commenter != nil {
		commenterName = commenter.Name
	}

	// Send notifications
	task, _ := s.taskRepo.FindByID(ctx, taskID)
	if task != nil && s.notifSvc != nil {
		// Notify assignee
		if task.AssigneeID != nil && *task.AssigneeID != userID {
			s.notifSvc.SendTaskCommented(ctx, *task.AssigneeID, commenterName, task.Title, task.ID, task.ProjectID)
		}

		// Notify reporter (if different from assignee and commenter)
		if task.ReporterID != userID {
			if task.AssigneeID == nil || *task.AssigneeID != task.ReporterID {
				s.notifSvc.SendTaskCommented(ctx, task.ReporterID, commenterName, task.Title, task.ID, task.ProjectID)
			}
		}

		// Parse and send mention notifications
		s.notifSvc.ParseAndSendMentions(ctx, content, commenterName, task.Title, task.ID, task.ProjectID, userID)
	}

	comment.User = commenter
	return comment, nil
}

func (s *commentService) GetByID(ctx context.Context, id string) (*repository.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, ErrNotFound
	}
	comment.User, _ = s.userRepo.FindByID(ctx, comment.UserID)
	return comment, nil
}

func (s *commentService) ListByTask(ctx context.Context, taskID string) ([]*repository.Comment, error) {
	comments, err := s.commentRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	for _, c := range comments {
		c.User, _ = s.userRepo.FindByID(ctx, c.UserID)
	}

	return comments, nil
}

func (s *commentService) Update(ctx context.Context, id, userID, content string) (*repository.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil || comment == nil {
		return nil, ErrNotFound
	}

	if comment.UserID != userID {
		return nil, ErrUnauthorized
	}

	comment.Content = content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}

	comment.User, _ = s.userRepo.FindByID(ctx, comment.UserID)
	return comment, nil
}

func (s *commentService) Delete(ctx context.Context, id, userID string) error {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil || comment == nil {
		return ErrNotFound
	}

	if comment.UserID != userID {
		return ErrUnauthorized
	}

	return s.commentRepo.Delete(ctx, id)
}
