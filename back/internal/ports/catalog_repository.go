package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type CatalogRepository interface {
	ListCategories(ctx context.Context) ([]domain.CatalogCategory, error)
	GetCategory(ctx context.Context, slug string) (domain.CatalogCategory, error)
	ListProducts(ctx context.Context, categoryID int64, filters map[string][]string, limit, offset int) ([]domain.Product, int, error)
	ListFacetValues(ctx context.Context, categoryID int64, taxonomy string, filters map[string][]string) ([]domain.CatalogFacetValue, error)
	GetFacetValue(ctx context.Context, categoryID int64, taxonomy, key string) (domain.CatalogFacetValue, error)
	GetFacetPage(ctx context.Context, categoryID, termID int64) (domain.CatalogFacetPage, error)
	ListFacetPages(ctx context.Context) ([]domain.CatalogFacetPage, error)
	UpsertFacetPage(ctx context.Context, page domain.CatalogFacetPage) (domain.CatalogFacetPage, error)
	DeleteFacetPage(ctx context.Context, id int64) error
	ListRelatedProjects(ctx context.Context, categoryID int64, limit int) ([]domain.ProjectCard, error)
	ListIndexableRoutes(ctx context.Context, minimumProducts int) ([]domain.CatalogRoute, error)
}
