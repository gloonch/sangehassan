package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type ProductTermService struct {
	repo ports.ProductTermRepository
}

func NewProductTermService(repo ports.ProductTermRepository) *ProductTermService {
	return &ProductTermService{repo: repo}
}

func (s *ProductTermService) List(ctx context.Context, taxonomy string) ([]domain.ProductTerm, error) {
	return s.repo.List(ctx, taxonomy)
}

func (s *ProductTermService) Upsert(ctx context.Context, term domain.ProductTerm) (domain.ProductTerm, error) {
	return s.repo.Upsert(ctx, term)
}

func (s *ProductTermService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

