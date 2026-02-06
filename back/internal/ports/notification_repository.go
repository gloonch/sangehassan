package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification domain.Notification) (domain.Notification, error)
	ListByUser(ctx context.Context, userID string, includeRead bool) ([]domain.Notification, error)
	MarkRead(ctx context.Context, id int64) error
}
