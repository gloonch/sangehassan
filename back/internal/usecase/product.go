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

func (s *ProductService) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	return s.repo.Create(ctx, product)
}

func (s *ProductService) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	return s.repo.Update(ctx, product)
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
