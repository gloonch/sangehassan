package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type ProductService struct {
	repo ports.ProductRepository
}

func NewProductService(repo ports.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(ctx context.Context) ([]domain.Product, error) {
	return s.repo.List(ctx)
}

func (s *ProductService) ListPopular(ctx context.Context) ([]domain.Product, error) {
	return s.repo.ListPopular(ctx)
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (domain.Product, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *ProductService) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	if product.Slug == "" {
		product.Slug = slugify(product.TitleEN)
	}
	images := normalizeProductImages(product.ImageURL, product.Images)
	if product.ImageURL == "" && len(images) > 0 {
		product.ImageURL = images[0]
	}

	created, err := s.repo.Create(ctx, product)
	if err != nil {
		return domain.Product{}, err
	}
	if len(images) > 0 {
		if err := s.repo.ReplaceImages(ctx, created.ID, images); err != nil {
			return domain.Product{}, err
		}
		created.Images = images
		created.ImageCount = len(images)
	}
	if err := s.repo.ReplaceCategories(ctx, created.ID, categoryIDsFromProduct(product)); err != nil {
		return domain.Product{}, err
	}
	return created, nil
}

func (s *ProductService) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	if product.Slug == "" {
		product.Slug = slugify(product.TitleEN)
	}
	images := normalizeProductImages(product.ImageURL, product.Images)
	if product.ImageURL == "" && len(images) > 0 {
		product.ImageURL = images[0]
	}

	updated, err := s.repo.Update(ctx, product)
	if err != nil {
		return domain.Product{}, err
	}
	if err := s.repo.ReplaceImages(ctx, updated.ID, images); err != nil {
		return domain.Product{}, err
	}
	updated.Images = images
	updated.ImageCount = len(images)
	if err := s.repo.ReplaceCategories(ctx, updated.ID, categoryIDsFromProduct(product)); err != nil {
		return domain.Product{}, err
	}
	return updated, nil
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeProductImages(primary string, images []string) []string {
	if len(images) == 0 && primary != "" {
		return []string{primary}
	}
	unique := make([]string, 0, len(images))
	seen := make(map[string]struct{})
	for _, img := range images {
		if img == "" {
			continue
		}
		if _, ok := seen[img]; ok {
			continue
		}
		seen[img] = struct{}{}
		unique = append(unique, img)
	}
	return unique
}

func categoryIDsFromProduct(product domain.Product) []int64 {
	if product.MainCategoryID == nil {
		return nil
	}
	return []int64{*product.MainCategoryID}
}
