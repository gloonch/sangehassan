package usecase

import (
	"context"
	"errors"
	"strings"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

var ErrListingProductRequired = errors.New("product_id is required")
var ErrListingProductRequestRequired = errors.New("product request query is required")

type ListingService struct {
	repo ports.ListingRepository
}

func NewListingService(repo ports.ListingRepository) *ListingService {
	return &ListingService{repo: repo}
}

func (s *ListingService) List(ctx context.Context, filter ports.ListingFilter) ([]domain.Listing, error) {
	if len(filter.Status) == 0 {
		filter.Status = []string{domain.ListingStatusActive}
	}
	return s.repo.List(ctx, filter)
}

func (s *ListingService) ListAdmin(ctx context.Context, filter ports.ListingFilter) ([]domain.Listing, error) {
	return s.repo.List(ctx, filter)
}

func (s *ListingService) ListLiveFeed(ctx context.Context, limit int) ([]domain.ListingLiveFeedItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 10 {
		limit = 10
	}
	return s.repo.ListLiveFeed(ctx, limit)
}

func (s *ListingService) GetByID(ctx context.Context, id int64) (domain.Listing, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ListingService) Create(ctx context.Context, listing domain.Listing) (domain.Listing, error) {
	if listing.ProductID <= 0 {
		return domain.Listing{}, ErrListingProductRequired
	}
	if listing.Status == "" {
		listing.Status = domain.ListingStatusActive
	}
	if listing.ExtraProps == nil {
		listing.ExtraProps = map[string]any{}
	}
	listing.ExtraProps = sanitizeListingExtraProps(listing.ExtraProps)

	created, err := s.repo.Create(ctx, listing)
	if err != nil {
		return domain.Listing{}, err
	}
	return created, nil
}

func (s *ListingService) Update(ctx context.Context, listing domain.Listing) (domain.Listing, error) {
	if listing.ProductID <= 0 {
		return domain.Listing{}, ErrListingProductRequired
	}
	if listing.Status == "" {
		listing.Status = domain.ListingStatusActive
	}
	if listing.ExtraProps == nil {
		listing.ExtraProps = map[string]any{}
	}
	listing.ExtraProps = sanitizeListingExtraProps(listing.ExtraProps)

	updated, err := s.repo.Update(ctx, listing, listing.CreatedBy)
	if err != nil {
		return domain.Listing{}, err
	}
	return updated, nil
}

func (s *ListingService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id, nil)
}

func (s *ListingService) DeleteOwned(ctx context.Context, id int64, ownerID *string) error {
	return s.repo.Delete(ctx, id, ownerID)
}

func (s *ListingService) CreateProductRequest(ctx context.Context, request domain.ListingProductRequest) (domain.ListingProductRequest, error) {
	request.Query = strings.TrimSpace(request.Query)
	if request.Query == "" {
		return domain.ListingProductRequest{}, ErrListingProductRequestRequired
	}
	if request.Status == "" {
		request.Status = "NEW"
	}
	return s.repo.CreateProductRequest(ctx, request)
}

func (s *ListingService) ListProductRequests(ctx context.Context, limit, offset int) ([]domain.ListingProductRequest, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListProductRequests(ctx, limit, offset)
}

func normalizeListingImages(images []domain.ListingImage) []domain.ListingImage {
	if len(images) == 0 {
		return images
	}
	unique := make([]domain.ListingImage, 0, len(images))
	seen := make(map[string]struct{})
	for idx, img := range images {
		if img.ImageURL == "" {
			continue
		}
		if _, ok := seen[img.ImageURL]; ok {
			continue
		}
		seen[img.ImageURL] = struct{}{}
		if img.Position == 0 {
			img.Position = idx
		}
		unique = append(unique, img)
	}
	return unique
}

func sanitizeListingExtraProps(extra map[string]any) map[string]any {
	if len(extra) == 0 {
		return extra
	}
	for key := range extra {
		if strings.EqualFold(strings.TrimSpace(key), "recommended_use") {
			delete(extra, key)
		}
	}
	return extra
}
