package ports

import (
	"context"
	"time"

	"sangehassan/back/internal/domain"
)

type DealRequestFilter struct {
	BuyerID   *string
	Status    []domain.DealRequestStatus
	ListingID *int64
	Limit     int
	Offset    int
}

type DealRequestRepository interface {
	Create(ctx context.Context, req domain.DealRequest) (domain.DealRequest, error)
	GetByID(ctx context.Context, id int64) (domain.DealRequest, error)
	ListByBuyer(ctx context.Context, buyerID string, limit, offset int) ([]domain.DealRequest, error)
	ListAdmin(ctx context.Context, filter DealRequestFilter) ([]domain.DealRequest, error)
	UpdateStatus(ctx context.Context, requestID int64, toStatus domain.DealRequestStatus, meetingAt *time.Time, adminNote *string) error
	AddStatusHistory(ctx context.Context, history domain.DealStatusHistory) error
}
