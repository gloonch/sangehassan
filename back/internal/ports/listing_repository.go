package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type ListingFilter struct {
	StoneType  string
	Form       string
	Province   string
	City       string
	OwnerID    *string
	MinTonnage *float64
	MaxTonnage *float64
	MinPrice   *float64
	MaxPrice   *float64
	Status     []string
	Query      string
	Sort       string // newest | price_asc | price_desc | tonnage_desc | tonnage_asc
	Limit      int
	Offset     int
}

type ListingRepository interface {
	List(ctx context.Context, filter ListingFilter) ([]domain.Listing, error)
	GetByID(ctx context.Context, id int64) (domain.Listing, error)
	Create(ctx context.Context, listing domain.Listing) (domain.Listing, error)
	Update(ctx context.Context, listing domain.Listing, ownerID *string) (domain.Listing, error)
	ReplaceImages(ctx context.Context, listingID int64, images []domain.ListingImage) error
	Delete(ctx context.Context, id int64, ownerID *string) error
}
