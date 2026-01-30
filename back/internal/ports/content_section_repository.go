package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type ContentSectionRepository interface {
	List(ctx context.Context, page string) ([]domain.ContentSection, error)
	GetByID(ctx context.Context, id int64) (domain.ContentSection, error)
	Create(ctx context.Context, section domain.ContentSection) (domain.ContentSection, error)
	Update(ctx context.Context, section domain.ContentSection) (domain.ContentSection, error)
	Delete(ctx context.Context, id int64) error
	ReplaceImages(ctx context.Context, sectionID int64, images []string) error
}
