package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type ProductRepository interface {
	List(ctx context.Context) ([]domain.Product, error)
	ListPopular(ctx context.Context) ([]domain.Product, error)
	GetByID(ctx context.Context, id int64) (domain.Product, error)
	Create(ctx context.Context, product domain.Product) (domain.Product, error)
	Update(ctx context.Context, product domain.Product) (domain.Product, error)
	Delete(ctx context.Context, id int64) error
}
