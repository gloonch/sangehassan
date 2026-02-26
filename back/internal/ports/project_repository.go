package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type ProjectRepository interface {
	ListPublic(ctx context.Context) ([]domain.ProjectCard, error)
	List(ctx context.Context) ([]domain.Project, error)
	GetByID(ctx context.Context, id int64) (domain.Project, error)
	Create(ctx context.Context, project domain.Project) (domain.Project, error)
	Update(ctx context.Context, project domain.Project) (domain.Project, error)
	Delete(ctx context.Context, id int64) error
	ReplaceGalleryImages(ctx context.Context, projectID int64, images []string) error
}
