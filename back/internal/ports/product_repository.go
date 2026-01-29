package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type ProductRepository interface {
	List(ctx context.Context) ([]domain.Product, error)
	ListPopular(ctx context.Context) ([]domain.Product, error)
	GetByID(ctx context.Context, id int64) (domain.Product, error)
	GetBySlug(ctx context.Context, slug string) (domain.Product, error)
	Create(ctx context.Context, product domain.Product) (domain.Product, error)
	Update(ctx context.Context, product domain.Product) (domain.Product, error)
	Delete(ctx context.Context, id int64) error
	ReplaceImages(ctx context.Context, productID int64, images []string) error
	ReplaceCategories(ctx context.Context, productID int64, categoryIDs []int64) error
}
