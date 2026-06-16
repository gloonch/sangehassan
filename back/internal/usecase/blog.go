package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/microcosm-cc/bluemonday"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

var ErrInvalidBlog = errors.New("invalid blog")

var blogReservedSlugs = map[string]bool{
	"admin": true, "api": true, "category": true, "tag": true, "search": true,
	"new": true, "edit": true, "preview": true, "feed": true, "rss": true,
}

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

type BlogService struct {
	repo      ports.BlogRepository
	sanitizer *bluemonday.Policy
}

func NewBlogService(repo ports.BlogRepository) *BlogService {
	policy := bluemonday.NewPolicy()
	policy.AllowStandardURLs()
	policy.AllowRelativeURLs(true)
	policy.AllowURLSchemes("http", "https", "mailto", "tel")
	policy.AllowElements(
		"p", "br", "strong", "b", "em", "i", "s", "blockquote", "ul", "ol", "li",
		"h2", "h3", "h4", "pre", "code", "hr", "figure", "figcaption",
		"table", "thead", "tbody", "tfoot", "tr", "th", "td", "img",
	)
	policy.AllowAttrs("href", "title", "target", "rel").OnElements("a")
	policy.AllowAttrs("src", "alt", "title", "width", "height", "loading").OnElements("img")
	policy.AllowAttrs("colspan", "rowspan", "scope").OnElements("th", "td")
	return &BlogService{repo: repo, sanitizer: policy}
}

func (s *BlogService) ListPublic(ctx context.Context, locale string) ([]domain.Blog, error) {
	if !validBlogLocale(locale) {
		return nil, fmt.Errorf("%w: unsupported locale", ErrInvalidBlog)
	}
	blogs, err := s.repo.ListPublic(ctx, locale)
	return decorateBlogs(blogs), err
}

func (s *BlogService) ListAdmin(ctx context.Context) ([]domain.Blog, error) {
	blogs, err := s.repo.ListAdmin(ctx)
	if err != nil {
		return nil, err
	}
	for i := range blogs {
		decorateAdminBlog(&blogs[i])
	}
	return blogs, nil
}

func (s *BlogService) GetPublicBySlug(ctx context.Context, locale, slug string) (domain.Blog, error) {
	if !validBlogLocale(locale) || strings.TrimSpace(slug) == "" {
		return domain.Blog{}, fmt.Errorf("%w: invalid locale or slug", ErrInvalidBlog)
	}
	blog, err := s.repo.GetPublicBySlug(ctx, locale, slug)
	return decorateBlog(blog), err
}

func (s *BlogService) GetByID(ctx context.Context, id int64) (domain.Blog, error) {
	blog, err := s.repo.GetByID(ctx, id)
	if err == nil {
		decorateAdminBlog(&blog)
	}
	return blog, err
}

func (s *BlogService) Create(ctx context.Context, blog domain.Blog) (domain.Blog, error) {
	if err := s.prepare(&blog, nil); err != nil {
		return domain.Blog{}, err
	}
	created, err := s.repo.Create(ctx, blog)
	if err == nil {
		decorateAdminBlog(&created)
	}
	return created, err
}

func (s *BlogService) Update(ctx context.Context, blog domain.Blog) (domain.Blog, error) {
	existing, err := s.repo.GetByID(ctx, blog.ID)
	if err != nil {
		return domain.Blog{}, err
	}
	if err := s.prepare(&blog, &existing); err != nil {
		return domain.Blog{}, err
	}
	updated, err := s.repo.Update(ctx, blog)
	if err == nil {
		decorateAdminBlog(&updated)
	}
	return updated, err
}

func (s *BlogService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *BlogService) prepare(blog *domain.Blog, existing *domain.Blog) error {
	blog.Status = strings.ToLower(strings.TrimSpace(blog.Status))
	if blog.Status == "" {
		blog.Status = "draft"
	}
	if blog.Status != "draft" && blog.Status != "scheduled" && blog.Status != "published" && blog.Status != "archived" {
		return fmt.Errorf("%w: unsupported status", ErrInvalidBlog)
	}
	blog.AuthorName = strings.TrimSpace(blog.AuthorName)
	if blog.AuthorName == "" {
		blog.AuthorName = "SangeHassan"
	}
	blog.CategorySlug = slugify(blog.CategorySlug)
	blog.Tags = normalizeTags(blog.Tags)

	if blog.Status == "scheduled" {
		if blog.ScheduledAt == nil {
			return fmt.Errorf("%w: scheduled_at is required", ErrInvalidBlog)
		}
		if blog.ScheduledAt.Before(time.Now().Add(-time.Minute)) {
			return fmt.Errorf("%w: scheduled_at must be in the future", ErrInvalidBlog)
		}
	}
	if blog.Status == "published" && blog.PublishedAt == nil {
		now := time.Now().UTC()
		if existing != nil && existing.PublishedAt != nil {
			now = existing.PublishedAt.UTC()
		}
		blog.PublishedAt = &now
	}

	existingTranslations := make(map[string]domain.BlogTranslation)
	if existing != nil {
		for _, translation := range existing.Translations {
			existingTranslations[translation.Locale] = translation
		}
	}
	seenLocales := make(map[string]bool)
	prepared := make([]domain.BlogTranslation, 0, len(blog.Translations))
	for _, translation := range blog.Translations {
		translation.Locale = strings.ToLower(strings.TrimSpace(translation.Locale))
		translation.Title = strings.TrimSpace(translation.Title)
		if translation.Title == "" {
			continue
		}
		if !validBlogLocale(translation.Locale) || seenLocales[translation.Locale] {
			return fmt.Errorf("%w: invalid or duplicate locale", ErrInvalidBlog)
		}
		seenLocales[translation.Locale] = true
		translation.TranslationStatus = strings.ToLower(strings.TrimSpace(translation.TranslationStatus))
		if translation.TranslationStatus == "" {
			translation.TranslationStatus = "draft"
		}
		if translation.TranslationStatus != "draft" && translation.TranslationStatus != "published" {
			return fmt.Errorf("%w: invalid translation status", ErrInvalidBlog)
		}
		old := existingTranslations[translation.Locale]
		requestedSlug := strings.TrimSpace(translation.Slug)
		if requestedSlug == "" && old.Slug != "" {
			requestedSlug = old.Slug
		}
		if requestedSlug == "" {
			requestedSlug = translation.Title
		}
		translation.Slug = slugify(requestedSlug)
		if translation.Slug == "" {
			translation.Slug = fmt.Sprintf("article-%s", translation.Locale)
		}
		if blogReservedSlugs[translation.Slug] {
			translation.Slug += "-article"
		}
		translation.ContentHTML = s.sanitizer.Sanitize(translation.ContentHTML)
		if len(translation.ContentJSON) == 0 || !json.Valid(translation.ContentJSON) {
			translation.ContentJSON = json.RawMessage(`{"type":"doc","content":[]}`)
		}
		translation.Robots = normalizeRobots(translation.Robots)
		prepared = append(prepared, translation)
	}
	if len(prepared) == 0 {
		return fmt.Errorf("%w: at least one translation is required", ErrInvalidBlog)
	}
	if blog.Status == "published" || blog.Status == "scheduled" {
		hasPublishedTranslation := false
		for _, translation := range prepared {
			if translation.TranslationStatus == "published" && strings.TrimSpace(stripHTML(translation.ContentHTML)) != "" {
				hasPublishedTranslation = true
				break
			}
		}
		if !hasPublishedTranslation {
			return fmt.Errorf("%w: a published translation with content is required", ErrInvalidBlog)
		}
	}
	blog.Translations = prepared
	return nil
}

func validBlogLocale(locale string) bool {
	return locale == "fa" || locale == "en" || locale == "ar"
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		key := strings.ToLower(tag)
		if tag == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, tag)
	}
	return result
}

func normalizeRobots(value string) string {
	value = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
	if value == "noindex,follow" || value == "noindex,nofollow" || value == "index,nofollow" {
		return value
	}
	return "index,follow"
}

func stripHTML(value string) string {
	return strings.Join(strings.Fields(htmlTagPattern.ReplaceAllString(value, " ")), " ")
}

func decorateBlog(blog domain.Blog) domain.Blog {
	words := 0
	inWord := false
	for _, char := range stripHTML(blog.ContentHTML) {
		if unicode.IsSpace(char) {
			inWord = false
		} else if !inWord {
			words++
			inWord = true
		}
	}
	if words > 0 {
		blog.ReadingTimeMinutes = (words + 199) / 200
	}
	return blog
}

func decorateBlogs(blogs []domain.Blog) []domain.Blog {
	for i := range blogs {
		blogs[i] = decorateBlog(blogs[i])
	}
	return blogs
}

func decorateAdminBlog(blog *domain.Blog) {
	for _, translation := range blog.Translations {
		if blog.Title == "" || translation.Locale == "en" {
			blog.Title = translation.Title
			blog.Slug = translation.Slug
			blog.Excerpt = translation.Excerpt
			blog.ContentHTML = translation.ContentHTML
		}
	}
}
