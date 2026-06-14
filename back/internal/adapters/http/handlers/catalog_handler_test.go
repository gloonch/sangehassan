package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/usecase"
)

type catalogHandlerRepoStub struct{}

func (catalogHandlerRepoStub) ListCategories(context.Context) ([]domain.CatalogCategory, error) {
	return []domain.CatalogCategory{{Category: domain.Category{ID: 1, Slug: "travertine", TitleEN: "Travertine", TitleFA: "تراورتن", TitleAR: "ترافرتين", IsIndexable: true}, ProductCount: 2}}, nil
}

func (catalogHandlerRepoStub) GetCategory(_ context.Context, slug string) (domain.CatalogCategory, error) {
	if slug != "travertine" {
		return domain.CatalogCategory{}, sql.ErrNoRows
	}
	return domain.CatalogCategory{Category: domain.Category{ID: 1, Slug: slug, TitleEN: "Travertine", TitleFA: "تراورتن", TitleAR: "ترافرتين", IsIndexable: true}, ProductCount: 2}, nil
}

func (catalogHandlerRepoStub) ListProducts(context.Context, int64, map[string][]string, int, int) ([]domain.Product, int, error) {
	return []domain.Product{{ID: 1, Slug: "sample", TitleFA: "نمونه", IsActive: true, IsIndexable: true}}, 2, nil
}

func (catalogHandlerRepoStub) ListFacetValues(context.Context, int64, string, map[string][]string) ([]domain.CatalogFacetValue, error) {
	return nil, nil
}

func (catalogHandlerRepoStub) GetFacetValue(_ context.Context, _ int64, taxonomy, key string) (domain.CatalogFacetValue, error) {
	if taxonomy != "tone" || key != "white" {
		return domain.CatalogFacetValue{}, sql.ErrNoRows
	}
	return domain.CatalogFacetValue{ID: 10, Key: key, LabelEN: "White", LabelFA: "سفید", LabelAR: "أبيض", Count: 2, IsIndexable: true}, nil
}

func (catalogHandlerRepoStub) GetFacetPage(context.Context, int64, int64) (domain.CatalogFacetPage, error) {
	return domain.CatalogFacetPage{}, sql.ErrNoRows
}

func (catalogHandlerRepoStub) ListFacetPages(context.Context) ([]domain.CatalogFacetPage, error) {
	return nil, nil
}

func (catalogHandlerRepoStub) UpsertFacetPage(_ context.Context, page domain.CatalogFacetPage) (domain.CatalogFacetPage, error) {
	return page, nil
}

func (catalogHandlerRepoStub) DeleteFacetPage(context.Context, int64) error { return nil }

func (catalogHandlerRepoStub) ListRelatedProjects(context.Context, int64, int) ([]domain.ProjectCard, error) {
	return nil, nil
}

func (catalogHandlerRepoStub) ListIndexableRoutes(context.Context, int) ([]domain.CatalogRoute, error) {
	return nil, nil
}

func catalogTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	service := usecase.NewCatalogService(catalogHandlerRepoStub{}, 2)
	handler := NewCatalogHandler(service)
	router := gin.New()
	router.GET("/api/catalog/categories", handler.Hub)
	router.GET("/api/catalog/categories/:categorySlug", handler.Page)
	router.GET("/api/catalog/categories/:categorySlug/:facet/:value", handler.Page)
	return router
}

func TestCatalogHTTPRoutes(t *testing.T) {
	router := catalogTestRouter()
	for _, test := range []struct {
		path string
		want int
	}{
		{path: "/api/catalog/categories?locale=fa", want: http.StatusOK},
		{path: "/api/catalog/categories/travertine?locale=fa", want: http.StatusOK},
		{path: "/api/catalog/categories/travertine/color/white?locale=fa", want: http.StatusOK},
		{path: "/api/catalog/categories/missing?locale=fa", want: http.StatusNotFound},
		{path: "/api/catalog/categories/travertine/color/missing?locale=fa", want: http.StatusNotFound},
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != test.want {
			t.Errorf("GET %s status=%d want=%d body=%s", test.path, recorder.Code, test.want, recorder.Body.String())
		}
	}
}

func TestCatalogMultiFilterHTTPResponseIsNoindex(t *testing.T) {
	recorder := httptest.NewRecorder()
	catalogTestRouter().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/catalog/categories/travertine?locale=fa&color=white&finish=polished", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Data domain.CatalogPage `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.SEO.Robots != "noindex,follow" || response.Data.Indexable {
		t.Fatalf("expected noindex response, got indexable=%v robots=%q", response.Data.Indexable, response.Data.SEO.Robots)
	}
}
