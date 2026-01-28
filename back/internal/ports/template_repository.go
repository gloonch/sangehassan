package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type TemplateRepository interface {
	List(ctx context.Context) ([]domain.Template, error)
	ListActive(ctx context.Context) ([]domain.Template, error)
	GetByID(ctx context.Context, id int64) (domain.Template, error)
	Create(ctx context.Context, template domain.Template) (domain.Template, error)
	Update(ctx context.Context, template domain.Template) (domain.Template, error)
	Delete(ctx context.Context, id int64) error
}
