package usecase

import (
	"context"
	"database/sql"
	"testing"

	"sangehassan/back/internal/domain"
)

type catalogRepoStub struct {
	category  domain.CatalogCategory
	value     domain.CatalogFacetValue
	facetPage *domain.CatalogFacetPage
}

func (s catalogRepoStub) ListCategories(context.Context) ([]domain.CatalogCategory, error) {
	return []domain.CatalogCategory{s.category}, nil
}
func (s catalogRepoStub) GetCategory(context.Context, string) (domain.CatalogCategory, error) {
	if s.category.ID == 0 {
		return domain.CatalogCategory{}, sql.ErrNoRows
	}
	return s.category, nil
}
func (s catalogRepoStub) ListProducts(context.Context, int64, map[string][]string, int, int) ([]domain.Product, int, error) {
	return []domain.Product{{ID: 1, Slug: "sample", TitleFA: "نمونه"}}, 1, nil
}
func (s catalogRepoStub) ListFacetValues(context.Context, int64, string, map[string][]string) ([]domain.CatalogFacetValue, error) {
	return nil, nil
}
func (s catalogRepoStub) GetFacetValue(context.Context, int64, string, string) (domain.CatalogFacetValue, error) {
	if s.value.Key == "" {
		return domain.CatalogFacetValue{}, sql.ErrNoRows
	}
	return s.value, nil
}
func (s catalogRepoStub) GetFacetPage(context.Context, int64, int64) (domain.CatalogFacetPage, error) {
	if s.facetPage == nil {
		return domain.CatalogFacetPage{}, sql.ErrNoRows
	}
	return *s.facetPage, nil
}
func (s catalogRepoStub) ListFacetPages(context.Context) ([]domain.CatalogFacetPage, error) {
	return nil, nil
}
func (s catalogRepoStub) UpsertFacetPage(_ context.Context, page domain.CatalogFacetPage) (domain.CatalogFacetPage, error) {
	return page, nil
}
func (s catalogRepoStub) DeleteFacetPage(context.Context, int64) error { return nil }
func (s catalogRepoStub) ListRelatedProjects(context.Context, int64, int) ([]domain.ProjectCard, error) {
	return nil, nil
}
func (s catalogRepoStub) ListIndexableRoutes(context.Context, int) ([]domain.CatalogRoute, error) {
	return nil, nil
}

func TestCatalogSingleFacetIndexabilityThreshold(t *testing.T) {
	repo := catalogRepoStub{
		category: domain.CatalogCategory{Category: domain.Category{ID: 1, Slug: "travertine", TitleFA: "تراورتن", IsIndexable: true}},
		value:    domain.CatalogFacetValue{ID: 10, Key: "white", LabelFA: "سفید", Count: 2, IsIndexable: true},
	}
	service := NewCatalogService(repo, 2)
	page, err := service.Page(context.Background(), "fa", "travertine", "color", "white", nil, 24, 0)
	if err != nil {
		t.Fatalf("Page returned error: %v", err)
	}
	if !page.Indexable || page.SEO.Robots != "index,follow" {
		t.Fatalf("expected indexable page, got indexable=%v robots=%q", page.Indexable, page.SEO.Robots)
	}
	if page.SEO.Canonical != "/fa/products/travertine/color/white" {
		t.Fatalf("unexpected canonical: %s", page.SEO.Canonical)
	}
}

func TestCatalogQueryFiltersAreNoindex(t *testing.T) {
	repo := catalogRepoStub{category: domain.CatalogCategory{Category: domain.Category{ID: 1, Slug: "granite", TitleFA: "گرانیت", IsIndexable: true}}}
	service := NewCatalogService(repo, 2)
	page, err := service.Page(context.Background(), "fa", "granite", "", "", map[string][]string{
		"application": {"facade"},
		"finish":      {"polished"},
	}, 24, 0)
	if err != nil {
		t.Fatalf("Page returned error: %v", err)
	}
	if page.Indexable || page.SEO.Robots != "noindex,follow" {
		t.Fatalf("expected noindex query page, got indexable=%v robots=%q", page.Indexable, page.SEO.Robots)
	}
}

func TestCatalogThinFacetIsNoindex(t *testing.T) {
	repo := catalogRepoStub{
		category: domain.CatalogCategory{Category: domain.Category{ID: 1, Slug: "onyx", TitleFA: "مرمر", IsIndexable: true}},
		value:    domain.CatalogFacetValue{ID: 11, Key: "green", LabelFA: "سبز", Count: 1, IsIndexable: true},
	}
	service := NewCatalogService(repo, 2)
	page, err := service.Page(context.Background(), "fa", "onyx", "color", "green", nil, 24, 0)
	if err != nil {
		t.Fatalf("Page returned error: %v", err)
	}
	if page.Indexable || page.SEO.Robots != "noindex,follow" {
		t.Fatalf("expected thin facet to be noindex, got indexable=%v robots=%q", page.Indexable, page.SEO.Robots)
	}
}

func TestCatalogRejectsUnknownFacetRoute(t *testing.T) {
	repo := catalogRepoStub{category: domain.CatalogCategory{Category: domain.Category{ID: 1, Slug: "marble", TitleFA: "مرمریت", IsIndexable: true}}}
	service := NewCatalogService(repo, 2)
	_, err := service.Page(context.Background(), "fa", "marble", "unknown", "value", nil, 24, 0)
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestCatalogLocalizedCanonicalPaths(t *testing.T) {
	repo := catalogRepoStub{category: domain.CatalogCategory{Category: domain.Category{
		ID: 1, Slug: "granite", TitleEN: "Granite", TitleFA: "گرانیت", TitleAR: "جرانيت", IsIndexable: true,
	}}}
	service := NewCatalogService(repo, 2)
	for _, test := range []struct {
		locale, canonical string
	}{
		{locale: "en", canonical: "/en/products/granite"},
		{locale: "fa", canonical: "/fa/products/granite"},
		{locale: "ar", canonical: "/ar/products/granite"},
	} {
		page, err := service.Page(context.Background(), test.locale, "granite", "", "", nil, 24, 0)
		if err != nil {
			t.Fatalf("Page(%s) returned error: %v", test.locale, err)
		}
		if page.Locale != test.locale || page.SEO.Canonical != test.canonical {
			t.Fatalf("Page(%s) locale=%q canonical=%q", test.locale, page.Locale, page.SEO.Canonical)
		}
	}
}

func TestCatalogFacetSEOOverride(t *testing.T) {
	repo := catalogRepoStub{
		category: domain.CatalogCategory{Category: domain.Category{ID: 1, Slug: "travertine", TitleFA: "تراورتن", IsIndexable: true}},
		value:    domain.CatalogFacetValue{ID: 10, Key: "white", LabelFA: "سفید", Count: 3, IsIndexable: true},
		facetPage: &domain.CatalogFacetPage{
			IsActive: true, IsIndexable: true, TitleFA: "عنوان اختصاصی", H1FA: "تیتر اختصاصی", DescriptionFA: "توضیحات اختصاصی", IntroFA: "معرفی اختصاصی",
		},
	}
	page, err := NewCatalogService(repo, 2).Page(context.Background(), "fa", "travertine", "color", "white", nil, 24, 0)
	if err != nil {
		t.Fatalf("Page returned error: %v", err)
	}
	if page.SEO.Title != "عنوان اختصاصی" || page.SEO.H1 != "تیتر اختصاصی" || page.SEO.Description != "توضیحات اختصاصی" || page.SEO.Intro != "معرفی اختصاصی" {
		t.Fatalf("facet override was not applied: %+v", page.SEO)
	}
}
