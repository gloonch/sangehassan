package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type NotificationService struct {
	repo ports.NotificationRepository
}

func NewNotificationService(repo ports.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Create(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
	return s.repo.Create(ctx, notification)
}

func (s *NotificationService) ListByUser(ctx context.Context, userID string, includeRead bool) ([]domain.Notification, error) {
	return s.repo.ListByUser(ctx, userID, includeRead)
}

func (s *NotificationService) MarkRead(ctx context.Context, id int64) error {
	return s.repo.MarkRead(ctx, id)
}
