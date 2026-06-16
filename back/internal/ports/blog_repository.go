package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type BlogRepository interface {
	ListPublic(ctx context.Context, locale string) ([]domain.Blog, error)
	ListAdmin(ctx context.Context) ([]domain.Blog, error)
	GetPublicBySlug(ctx context.Context, locale, slug string) (domain.Blog, error)
	GetByID(ctx context.Context, id int64) (domain.Blog, error)
	Create(ctx context.Context, blog domain.Blog) (domain.Blog, error)
	Update(ctx context.Context, blog domain.Blog) (domain.Blog, error)
	Delete(ctx context.Context, id int64) error
}
