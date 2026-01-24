package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type CategoryService struct {
	repo ports.CategoryRepository
}

func NewCategoryService(repo ports.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context) ([]domain.Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) GetByID(ctx context.Context, id int64) (domain.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CategoryService) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	category.Slug = slugify(category.TitleEN)
	return s.repo.Create(ctx, category)
}

func (s *CategoryService) Update(ctx context.Context, category domain.Category) (domain.Category, error) {
	category.Slug = slugify(category.TitleEN)
	return s.repo.Update(ctx, category)
}

func (s *CategoryService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
