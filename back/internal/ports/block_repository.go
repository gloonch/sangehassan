package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type BlockRepository interface {
	List(ctx context.Context) ([]domain.Block, error)
	ListFeatured(ctx context.Context) ([]domain.Block, error)
	GetByID(ctx context.Context, id int64) (domain.Block, error)
	GetBySlug(ctx context.Context, slug string) (domain.Block, error)
	Create(ctx context.Context, block domain.Block) (domain.Block, error)
	Update(ctx context.Context, block domain.Block) (domain.Block, error)
	Delete(ctx context.Context, id int64) error
	ReplaceImages(ctx context.Context, blockID int64, images []string) error
}
