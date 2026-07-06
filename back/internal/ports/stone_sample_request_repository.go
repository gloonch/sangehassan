package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type StoneSampleRequestFilter struct {
	UserID *string
	Status []domain.StoneSampleRequestStatus
	Limit  int
	Offset int
}

type StoneSampleRequestRepository interface {
	ListSampleCategories(ctx context.Context) ([]domain.CatalogCategory, error)
	ListSampleProducts(ctx context.Context, filter domain.StoneSampleCatalogFilter) ([]domain.StoneSampleProduct, int, error)
	GetSampleProductsByID(ctx context.Context, ids []int64) (map[int64]domain.StoneSampleProduct, error)

	ListAddresses(ctx context.Context, userID string) ([]domain.UserAddress, error)
	GetAddress(ctx context.Context, userID string, id int64) (domain.UserAddress, error)
	SaveAddress(ctx context.Context, address domain.UserAddress) (domain.UserAddress, error)
	ListContactNumbers(ctx context.Context, userID string) ([]domain.UserContactNumber, error)
	GetContactNumber(ctx context.Context, userID string, id int64) (domain.UserContactNumber, error)
	SaveContactNumber(ctx context.Context, phone domain.UserContactNumber) (domain.UserContactNumber, error)

	Create(ctx context.Context, req domain.StoneSampleRequest, items []domain.StoneSampleRequestItem) (domain.StoneSampleRequest, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.StoneSampleRequest, error)
	GetByUser(ctx context.Context, userID string, id int64) (domain.StoneSampleRequest, error)
	ListAdmin(ctx context.Context, filter StoneSampleRequestFilter) ([]domain.StoneSampleRequest, error)
	GetAdmin(ctx context.Context, id int64) (domain.StoneSampleRequest, error)
	UpdateStatus(ctx context.Context, requestID int64, status domain.StoneSampleRequestStatus, adminNote *string) error
	AddStatusHistory(ctx context.Context, history domain.StoneSampleStatusHistory) error
}
