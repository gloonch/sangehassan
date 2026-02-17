package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type ProductTermRepository interface {
	List(ctx context.Context, taxonomy string) ([]domain.ProductTerm, error)
	Upsert(ctx context.Context, term domain.ProductTerm) (domain.ProductTerm, error)
	Delete(ctx context.Context, id int64) error
}

