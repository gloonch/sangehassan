package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type BlogService struct {
	repo ports.BlogRepository
}

func NewBlogService(repo ports.BlogRepository) *BlogService {
	return &BlogService{repo: repo}
}

func (s *BlogService) List(ctx context.Context) ([]domain.Blog, error) {
	return s.repo.List(ctx)
}

func (s *BlogService) GetByID(ctx context.Context, id int64) (domain.Blog, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BlogService) Create(ctx context.Context, blog domain.Blog) (domain.Blog, error) {
	return s.repo.Create(ctx, blog)
}

func (s *BlogService) Update(ctx context.Context, blog domain.Blog) (domain.Blog, error) {
	return s.repo.Update(ctx, blog)
}

func (s *BlogService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
