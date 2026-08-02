package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListLiveFeedOnlyQueriesActiveListingsAndActiveProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	publishedAt := time.Date(2026, time.August, 2, 8, 30, 0, 0, time.UTC)
	query := regexp.QuoteMeta(`
		SELECT l.id, COALESCE(l.title, ''), COALESCE(l.stone_type, ''), COALESCE(l.form, ''),
		       l.tonnage, l.created_at,
		       p.id, COALESCE(p.title_en, ''), COALESCE(p.title_fa, ''),
		       COALESCE(p.title_ar, ''), COALESCE(p.slug, '')
		FROM listings l
		JOIN products p ON p.id = l.product_id
		WHERE l.status = $1
		  AND p.is_active = TRUE
		  AND COALESCE(
		        NULLIF(BTRIM(l.title), ''),
		        NULLIF(BTRIM(p.title_en), ''),
		        NULLIF(BTRIM(p.title_fa), ''),
		        NULLIF(BTRIM(p.title_ar), '')
		      ) IS NOT NULL
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT $2
	`)
	mock.ExpectQuery(query).
		WithArgs("ACTIVE", 10).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "stone_type", "form", "tonnage", "created_at",
			"product_id", "title_en", "title_fa", "title_ar", "slug",
		}).AddRow(
			42, "", "travertine", "block", 24.0, publishedAt,
			7, "Abbasabad Travertine", "تراورتن عباس‌آباد", "ترافرتين عباس آباد", "abbasabad-travertine",
		))

	items, err := NewListingRepository(db).ListLiveFeed(context.Background(), 10)
	if err != nil {
		t.Fatalf("list live feed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one item, got %d", len(items))
	}
	if items[0].ID != 42 || items[0].Quantity == nil || *items[0].Quantity != 24 {
		t.Fatalf("unexpected feed item: %#v", items[0])
	}
	if items[0].Product.TitleFA != "تراورتن عباس‌آباد" || !items[0].PublishedAt.Equal(publishedAt) {
		t.Fatalf("unexpected localized product or publication time: %#v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
