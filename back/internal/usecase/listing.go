package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

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

func (s *ListingService) GetByID(ctx context.Context, id int64) (domain.Listing, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ListingService) Create(ctx context.Context, listing domain.Listing) (domain.Listing, error) {
	if listing.Status == "" {
		listing.Status = domain.ListingStatusActive
	}
	if listing.ExtraProps == nil {
		listing.ExtraProps = map[string]any{}
	}

	images := normalizeListingImages(listing.Images)

	created, err := s.repo.Create(ctx, listing)
	if err != nil {
		return domain.Listing{}, err
	}
	if err := s.repo.ReplaceImages(ctx, created.ID, images); err != nil {
		return domain.Listing{}, err
	}
	created.Images = images
	return created, nil
}

func (s *ListingService) Update(ctx context.Context, listing domain.Listing) (domain.Listing, error) {
	if listing.Status == "" {
		listing.Status = domain.ListingStatusActive
	}
	if listing.ExtraProps == nil {
		listing.ExtraProps = map[string]any{}
	}

	images := normalizeListingImages(listing.Images)

	updated, err := s.repo.Update(ctx, listing, listing.CreatedBy)
	if err != nil {
		return domain.Listing{}, err
	}
	if err := s.repo.ReplaceImages(ctx, updated.ID, images); err != nil {
		return domain.Listing{}, err
	}
	updated.Images = images
	return updated, nil
}

func (s *ListingService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id, nil)
}

func (s *ListingService) DeleteOwned(ctx context.Context, id int64, ownerID *string) error {
	return s.repo.Delete(ctx, id, ownerID)
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
