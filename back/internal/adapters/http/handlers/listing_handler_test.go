package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
	"sangehassan/back/internal/usecase"
)

type liveFeedListingRepoStub struct {
	limit int
	items []domain.ListingLiveFeedItem
}

func (r *liveFeedListingRepoStub) List(context.Context, ports.ListingFilter) ([]domain.Listing, error) {
	return nil, nil
}

func (r *liveFeedListingRepoStub) ListLiveFeed(_ context.Context, limit int) ([]domain.ListingLiveFeedItem, error) {
	r.limit = limit
	return r.items, nil
}

func (r *liveFeedListingRepoStub) GetByID(context.Context, int64) (domain.Listing, error) {
	return domain.Listing{}, nil
}

func (r *liveFeedListingRepoStub) Create(_ context.Context, listing domain.Listing) (domain.Listing, error) {
	return listing, nil
}

func (r *liveFeedListingRepoStub) Update(_ context.Context, listing domain.Listing, _ *string) (domain.Listing, error) {
	return listing, nil
}

func (r *liveFeedListingRepoStub) ReplaceImages(context.Context, int64, []domain.ListingImage) error {
	return nil
}

func (r *liveFeedListingRepoStub) Delete(context.Context, int64, *string) error { return nil }

func (r *liveFeedListingRepoStub) CreateProductRequest(_ context.Context, request domain.ListingProductRequest) (domain.ListingProductRequest, error) {
	return request, nil
}

func (r *liveFeedListingRepoStub) ListProductRequests(context.Context, int, int) ([]domain.ListingProductRequest, error) {
	return nil, nil
}

func TestLiveFeedReturnsLocalizedMinimalPayloadAndCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	quantity := 24.0
	repo := &liveFeedListingRepoStub{items: []domain.ListingLiveFeedItem{{
		ID:        42,
		StoneType: "travertine",
		Form:      "block",
		Quantity:  &quantity,
		Product: domain.ListingProduct{
			ID:      7,
			TitleEN: "Abbasabad Travertine",
			TitleFA: "تراورتن عباس‌آباد",
			TitleAR: "ترافرتين عباس آباد",
		},
		PublishedAt: time.Date(2026, time.August, 2, 8, 30, 0, 0, time.UTC),
	}}}
	handler := NewListingHandler(usecase.NewListingService(repo), nil, nil)
	router := gin.New()
	router.GET("/api/ads/live-feed", handler.LiveFeed)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/ads/live-feed?limit=99&locale=fa", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if repo.limit != 10 {
		t.Fatalf("expected service to cap limit at 10, got %d", repo.limit)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=30, s-maxage=30, stale-while-revalidate=60" {
		t.Fatalf("unexpected cache header: %q", got)
	}

	var response struct {
		Data struct {
			Items []liveFeedItemResponse `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Data.Items) != 1 {
		t.Fatalf("expected one response item, got %d", len(response.Data.Items))
	}
	item := response.Data.Items[0]
	if item.StoneName != "تراورتن عباس‌آباد" || item.Unit != "ton" || item.ProductType != "block" {
		t.Fatalf("unexpected response item: %#v", item)
	}
}
