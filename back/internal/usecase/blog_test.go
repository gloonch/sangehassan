package usecase

import (
	"context"
	"strings"
	"testing"

	"sangehassan/back/internal/domain"
)

type blogRepoStub struct {
	existing domain.Blog
	saved    domain.Blog
}

func (r *blogRepoStub) ListPublic(context.Context, string) ([]domain.Blog, error) { return nil, nil }
func (r *blogRepoStub) ListAdmin(context.Context) ([]domain.Blog, error)          { return nil, nil }
func (r *blogRepoStub) GetPublicBySlug(context.Context, string, string) (domain.Blog, error) {
	return domain.Blog{}, nil
}
func (r *blogRepoStub) GetByID(context.Context, int64) (domain.Blog, error) { return r.existing, nil }
func (r *blogRepoStub) Create(_ context.Context, blog domain.Blog) (domain.Blog, error) {
	r.saved = blog
	return blog, nil
}
func (r *blogRepoStub) Update(_ context.Context, blog domain.Blog) (domain.Blog, error) {
	r.saved = blog
	return blog, nil
}
func (r *blogRepoStub) Delete(context.Context, int64) error { return nil }

func TestBlogCreateGeneratesSlugAndSanitizesHTML(t *testing.T) {
	repo := &blogRepoStub{}
	service := NewBlogService(repo)
	_, err := service.Create(context.Background(), domain.Blog{
		Status: "published",
		Translations: []domain.BlogTranslation{{
			Locale:            "fa",
			Title:             "راهنمای خرید سنگ",
			ContentHTML:       `<h1>bad</h1><h2>good</h2><script>alert(1)</script><p><a href="/fa/products">link</a></p>`,
			TranslationStatus: "published",
		}},
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	translation := repo.saved.Translations[0]
	if translation.Slug != "rahnmay-khryd-sng" {
		t.Fatalf("unexpected generated slug: %q", translation.Slug)
	}
	if strings.Contains(translation.ContentHTML, "<script") || strings.Contains(translation.ContentHTML, "<h1") {
		t.Fatalf("unsafe or unsupported HTML survived: %s", translation.ContentHTML)
	}
	if !strings.Contains(translation.ContentHTML, `<a href="/fa/products"`) {
		t.Fatalf("safe relative link was removed: %s", translation.ContentHTML)
	}
}

func TestBlogUpdateKeepsSlugWhenInputIsBlank(t *testing.T) {
	repo := &blogRepoStub{existing: domain.Blog{
		ID: 7,
		Translations: []domain.BlogTranslation{{
			Locale: "en", Title: "Old title", Slug: "stable-slug",
		}},
	}}
	service := NewBlogService(repo)
	_, err := service.Update(context.Background(), domain.Blog{
		ID:     7,
		Status: "draft",
		Translations: []domain.BlogTranslation{{
			Locale: "en", Title: "A completely new title", ContentHTML: "<p>Draft</p>",
		}},
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}
	if got := repo.saved.Translations[0].Slug; got != "stable-slug" {
		t.Fatalf("expected stable slug, got %q", got)
	}
}
