package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"sangehassan/back/internal/domain"
)

type BlogRepository struct {
	db *sql.DB
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func NewBlogRepository(db *sql.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

const publicBlogColumns = `
	b.id, b.status, b.author_name, COALESCE(b.cover_image_url, ''), COALESCE(b.og_image_url, ''),
	COALESCE(b.category_slug, ''), b.tags, b.is_featured, b.scheduled_at, b.published_at,
	b.created_at, COALESCE(b.updated_at, b.created_at),
	t.locale, t.title, t.slug, COALESCE(t.excerpt, ''), t.content_json, t.content_html,
	COALESCE(t.seo_title, ''), COALESCE(t.seo_description, ''), COALESCE(t.canonical_url, ''),
	t.robots, t.translation_status, COALESCE(t.featured_image_alt, ''), COALESCE(t.og_image_alt, '')`

func scanPublicBlog(scanner rowScanner) (domain.Blog, error) {
	var blog domain.Blog
	var contentJSON []byte
	err := scanner.Scan(
		&blog.ID, &blog.Status, &blog.AuthorName, &blog.CoverImageURL, &blog.OGImageURL,
		&blog.CategorySlug, pq.Array(&blog.Tags), &blog.IsFeatured, &blog.ScheduledAt, &blog.PublishedAt,
		&blog.CreatedAt, &blog.UpdatedAt,
		&blog.Locale, &blog.Title, &blog.Slug, &blog.Excerpt, &contentJSON, &blog.ContentHTML,
		&blog.SEOTitle, &blog.SEODescription, &blog.CanonicalURL, &blog.Robots,
		&blog.TranslationStatus, &blog.FeaturedImageAlt, &blog.OGImageAlt,
	)
	blog.ContentJSON = json.RawMessage(contentJSON)
	return blog, err
}

func publicPublicationWhere() string {
	return `t.translation_status = 'published'
		AND (b.status = 'published' OR (b.status = 'scheduled' AND b.scheduled_at IS NOT NULL AND b.scheduled_at <= NOW()))
		AND COALESCE(b.published_at, b.scheduled_at, b.created_at) <= NOW()`
}

func (r *BlogRepository) ListPublic(ctx context.Context, locale string) ([]domain.Blog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+publicBlogColumns+`
		FROM blogs b
		JOIN blog_translations t ON t.blog_id = b.id
		WHERE t.locale = $1 AND `+publicPublicationWhere()+`
		ORDER BY b.is_featured DESC, COALESCE(b.published_at, b.scheduled_at, b.created_at) DESC, b.id DESC
	`, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blogs := make([]domain.Blog, 0)
	for rows.Next() {
		blog, scanErr := scanPublicBlog(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		blogs = append(blogs, blog)
	}
	return blogs, rows.Err()
}

func (r *BlogRepository) ListAdmin(ctx context.Context) ([]domain.Blog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, status, author_name, COALESCE(cover_image_url, ''), COALESCE(og_image_url, ''),
			COALESCE(category_slug, ''), tags, is_featured, scheduled_at, published_at,
			created_at, COALESCE(updated_at, created_at)
		FROM blogs
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blogs := make([]domain.Blog, 0)
	ids := make([]int64, 0)
	byID := make(map[int64]int)
	for rows.Next() {
		var blog domain.Blog
		if err := rows.Scan(
			&blog.ID, &blog.Status, &blog.AuthorName, &blog.CoverImageURL, &blog.OGImageURL,
			&blog.CategorySlug, pq.Array(&blog.Tags), &blog.IsFeatured, &blog.ScheduledAt,
			&blog.PublishedAt, &blog.CreatedAt, &blog.UpdatedAt,
		); err != nil {
			return nil, err
		}
		blog.Translations = make([]domain.BlogTranslation, 0)
		byID[blog.ID] = len(blogs)
		ids = append(ids, blog.ID)
		blogs = append(blogs, blog)
	}
	if err := rows.Err(); err != nil || len(ids) == 0 {
		return blogs, err
	}

	translationRows, err := r.db.QueryContext(ctx, `
		SELECT id, blog_id, locale, title, slug, COALESCE(excerpt, ''), content_json, content_html,
			COALESCE(seo_title, ''), COALESCE(seo_description, ''), COALESCE(canonical_url, ''),
			robots, translation_status, COALESCE(featured_image_alt, ''), COALESCE(og_image_alt, ''),
			created_at, COALESCE(updated_at, created_at)
		FROM blog_translations
		WHERE blog_id = ANY($1)
		ORDER BY blog_id, CASE locale WHEN 'fa' THEN 1 WHEN 'en' THEN 2 ELSE 3 END
	`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer translationRows.Close()
	for translationRows.Next() {
		translation, scanErr := scanTranslation(translationRows)
		if scanErr != nil {
			return nil, scanErr
		}
		if index, ok := byID[translation.BlogID]; ok {
			blogs[index].Translations = append(blogs[index].Translations, translation)
		}
	}
	return blogs, translationRows.Err()
}

func (r *BlogRepository) GetPublicBySlug(ctx context.Context, locale, slug string) (domain.Blog, error) {
	query := `SELECT ` + publicBlogColumns + `
		FROM blogs b JOIN blog_translations t ON t.blog_id = b.id
		WHERE t.locale = $1 AND t.slug = $2 AND ` + publicPublicationWhere()
	blog, err := scanPublicBlog(r.db.QueryRowContext(ctx, query, locale, slug))
	if err == nil {
		blog.Translations, err = r.listPublicTranslationLinks(ctx, blog.ID)
		return blog, err
	}
	if err != sql.ErrNoRows {
		return blog, err
	}

	var blogID int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT blog_id FROM blog_slug_redirects WHERE locale = $1 AND old_slug = $2
	`, locale, slug).Scan(&blogID); err != nil {
		return domain.Blog{}, err
	}
	redirectQuery := `SELECT ` + publicBlogColumns + `
		FROM blogs b JOIN blog_translations t ON t.blog_id = b.id
		WHERE b.id = $1 AND t.locale = $2 AND ` + publicPublicationWhere()
	blog, err = scanPublicBlog(r.db.QueryRowContext(ctx, redirectQuery, blogID, locale))
	if err == nil {
		blog.RedirectedFrom = slug
		blog.Translations, err = r.listPublicTranslationLinks(ctx, blog.ID)
	}
	return blog, err
}

func (r *BlogRepository) listPublicTranslationLinks(ctx context.Context, blogID int64) ([]domain.BlogTranslation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.blog_id, t.locale, t.title, t.slug, COALESCE(t.excerpt, ''),
			t.content_json, t.content_html, COALESCE(t.seo_title, ''),
			COALESCE(t.seo_description, ''), COALESCE(t.canonical_url, ''), t.robots,
			t.translation_status, COALESCE(t.featured_image_alt, ''), COALESCE(t.og_image_alt, ''),
			t.created_at, COALESCE(t.updated_at, t.created_at)
		FROM blog_translations t
		JOIN blogs b ON b.id = t.blog_id
		WHERE t.blog_id = $1 AND `+publicPublicationWhere()+`
		ORDER BY CASE t.locale WHEN 'fa' THEN 1 WHEN 'en' THEN 2 ELSE 3 END
	`, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	translations := make([]domain.BlogTranslation, 0)
	for rows.Next() {
		translation, scanErr := scanTranslation(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		translation.ContentHTML = ""
		translation.ContentJSON = nil
		translations = append(translations, translation)
	}
	return translations, rows.Err()
}

func (r *BlogRepository) GetByID(ctx context.Context, id int64) (domain.Blog, error) {
	var blog domain.Blog
	err := r.db.QueryRowContext(ctx, `
		SELECT id, status, author_name, COALESCE(cover_image_url, ''), COALESCE(og_image_url, ''),
			COALESCE(category_slug, ''), tags, is_featured, scheduled_at, published_at,
			created_at, COALESCE(updated_at, created_at)
		FROM blogs WHERE id = $1
	`, id).Scan(
		&blog.ID, &blog.Status, &blog.AuthorName, &blog.CoverImageURL, &blog.OGImageURL,
		&blog.CategorySlug, pq.Array(&blog.Tags), &blog.IsFeatured, &blog.ScheduledAt,
		&blog.PublishedAt, &blog.CreatedAt, &blog.UpdatedAt,
	)
	if err != nil {
		return domain.Blog{}, err
	}
	blog.Translations, err = r.listTranslations(ctx, r.db, id)
	return blog, err
}

func scanTranslation(scanner rowScanner) (domain.BlogTranslation, error) {
	var translation domain.BlogTranslation
	var contentJSON []byte
	err := scanner.Scan(
		&translation.ID, &translation.BlogID, &translation.Locale, &translation.Title,
		&translation.Slug, &translation.Excerpt, &contentJSON, &translation.ContentHTML,
		&translation.SEOTitle, &translation.SEODescription, &translation.CanonicalURL,
		&translation.Robots, &translation.TranslationStatus, &translation.FeaturedImageAlt,
		&translation.OGImageAlt, &translation.CreatedAt, &translation.UpdatedAt,
	)
	translation.ContentJSON = json.RawMessage(contentJSON)
	return translation, err
}

func (r *BlogRepository) listTranslations(ctx context.Context, queryer interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}, blogID int64) ([]domain.BlogTranslation, error) {
	rows, err := queryer.QueryContext(ctx, `
		SELECT id, blog_id, locale, title, slug, COALESCE(excerpt, ''), content_json, content_html,
			COALESCE(seo_title, ''), COALESCE(seo_description, ''), COALESCE(canonical_url, ''),
			robots, translation_status, COALESCE(featured_image_alt, ''), COALESCE(og_image_alt, ''),
			created_at, COALESCE(updated_at, created_at)
		FROM blog_translations WHERE blog_id = $1
		ORDER BY CASE locale WHEN 'fa' THEN 1 WHEN 'en' THEN 2 ELSE 3 END
	`, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	translations := make([]domain.BlogTranslation, 0)
	for rows.Next() {
		translation, scanErr := scanTranslation(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		translations = append(translations, translation)
	}
	return translations, rows.Err()
}

func (r *BlogRepository) Create(ctx context.Context, blog domain.Blog) (domain.Blog, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Blog{}, err
	}
	defer tx.Rollback()

	legacyTitle := firstTranslationTitle(blog.Translations)
	err = tx.QueryRowContext(ctx, `
		INSERT INTO blogs (
			title, excerpt, content, cover_image_url, status, author_name, og_image_url,
			category_slug, tags, is_featured, scheduled_at, published_at
		) VALUES ($1, NULL, NULL, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`, legacyTitle, nullableString(blog.CoverImageURL), blog.Status, blog.AuthorName,
		nullableString(blog.OGImageURL), nullableString(blog.CategorySlug), pq.Array(blog.Tags),
		blog.IsFeatured, blog.ScheduledAt, blog.PublishedAt,
	).Scan(&blog.ID, &blog.CreatedAt, &blog.UpdatedAt)
	if err != nil {
		return domain.Blog{}, err
	}

	for i := range blog.Translations {
		blog.Translations[i].BlogID = blog.ID
		if err := r.upsertTranslation(ctx, tx, &blog.Translations[i], 0); err != nil {
			return domain.Blog{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.Blog{}, err
	}
	return blog, nil
}

func (r *BlogRepository) Update(ctx context.Context, blog domain.Blog) (domain.Blog, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Blog{}, err
	}
	defer tx.Rollback()

	existing, err := r.getByIDTx(ctx, tx, blog.ID)
	if err != nil {
		return domain.Blog{}, err
	}
	legacyTitle := firstTranslationTitle(blog.Translations)
	err = tx.QueryRowContext(ctx, `
		UPDATE blogs SET title = $1, cover_image_url = $2, status = $3, author_name = $4,
			og_image_url = $5, category_slug = $6, tags = $7, is_featured = $8,
			scheduled_at = $9, published_at = $10, updated_at = NOW()
		WHERE id = $11
		RETURNING created_at, COALESCE(updated_at, created_at)
	`, legacyTitle, nullableString(blog.CoverImageURL), blog.Status, blog.AuthorName,
		nullableString(blog.OGImageURL), nullableString(blog.CategorySlug), pq.Array(blog.Tags),
		blog.IsFeatured, blog.ScheduledAt, blog.PublishedAt, blog.ID,
	).Scan(&blog.CreatedAt, &blog.UpdatedAt)
	if err != nil {
		return domain.Blog{}, err
	}

	existingByLocale := make(map[string]domain.BlogTranslation)
	for _, translation := range existing.Translations {
		existingByLocale[translation.Locale] = translation
	}
	locales := make([]string, 0, len(blog.Translations))
	for i := range blog.Translations {
		translation := &blog.Translations[i]
		translation.BlogID = blog.ID
		locales = append(locales, translation.Locale)
		old := existingByLocale[translation.Locale]
		if err := r.upsertTranslation(ctx, tx, translation, blog.ID); err != nil {
			return domain.Blog{}, err
		}
		if old.Slug != "" && old.Slug != translation.Slug {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO blog_slug_redirects (blog_id, locale, old_slug)
				VALUES ($1, $2, $3)
				ON CONFLICT (locale, old_slug) DO UPDATE SET blog_id = EXCLUDED.blog_id
			`, blog.ID, translation.Locale, old.Slug); err != nil {
				return domain.Blog{}, err
			}
		}
	}
	if len(locales) == 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM blog_translations WHERE blog_id = $1`, blog.ID); err != nil {
			return domain.Blog{}, err
		}
	} else if _, err := tx.ExecContext(ctx, `
		DELETE FROM blog_translations WHERE blog_id = $1 AND NOT (locale = ANY($2))
	`, blog.ID, pq.Array(locales)); err != nil {
		return domain.Blog{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Blog{}, err
	}
	return blog, nil
}

func (r *BlogRepository) getByIDTx(ctx context.Context, tx *sql.Tx, id int64) (domain.Blog, error) {
	var blog domain.Blog
	err := tx.QueryRowContext(ctx, `
		SELECT id, status, author_name, COALESCE(cover_image_url, ''), COALESCE(og_image_url, ''),
			COALESCE(category_slug, ''), tags, is_featured, scheduled_at, published_at,
			created_at, COALESCE(updated_at, created_at)
		FROM blogs WHERE id = $1
	`, id).Scan(
		&blog.ID, &blog.Status, &blog.AuthorName, &blog.CoverImageURL, &blog.OGImageURL,
		&blog.CategorySlug, pq.Array(&blog.Tags), &blog.IsFeatured, &blog.ScheduledAt,
		&blog.PublishedAt, &blog.CreatedAt, &blog.UpdatedAt,
	)
	if err != nil {
		return domain.Blog{}, err
	}
	blog.Translations, err = r.listTranslations(ctx, tx, id)
	return blog, err
}

func (r *BlogRepository) upsertTranslation(ctx context.Context, tx *sql.Tx, translation *domain.BlogTranslation, excludingBlogID int64) error {
	slug, err := r.uniqueSlug(ctx, tx, translation.Locale, translation.Slug, excludingBlogID)
	if err != nil {
		return err
	}
	translation.Slug = slug
	contentJSON := translation.ContentJSON
	if len(contentJSON) == 0 || !json.Valid(contentJSON) {
		contentJSON = json.RawMessage(`{"type":"doc","content":[]}`)
	}
	return tx.QueryRowContext(ctx, `
		INSERT INTO blog_translations (
			blog_id, locale, title, slug, excerpt, content_json, content_html,
			seo_title, seo_description, canonical_url, robots, translation_status,
			featured_image_alt, og_image_alt
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (blog_id, locale) DO UPDATE SET
			title = EXCLUDED.title, slug = EXCLUDED.slug, excerpt = EXCLUDED.excerpt,
			content_json = EXCLUDED.content_json, content_html = EXCLUDED.content_html,
			seo_title = EXCLUDED.seo_title, seo_description = EXCLUDED.seo_description,
			canonical_url = EXCLUDED.canonical_url, robots = EXCLUDED.robots,
			translation_status = EXCLUDED.translation_status,
			featured_image_alt = EXCLUDED.featured_image_alt, og_image_alt = EXCLUDED.og_image_alt,
			updated_at = NOW()
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`, translation.BlogID, translation.Locale, translation.Title, translation.Slug,
		nullableString(translation.Excerpt), contentJSON, translation.ContentHTML,
		nullableString(translation.SEOTitle), nullableString(translation.SEODescription),
		nullableString(translation.CanonicalURL), translation.Robots, translation.TranslationStatus,
		nullableString(translation.FeaturedImageAlt), nullableString(translation.OGImageAlt),
	).Scan(&translation.ID, &translation.CreatedAt, &translation.UpdatedAt)
}

func (r *BlogRepository) uniqueSlug(ctx context.Context, tx *sql.Tx, locale, desired string, excludingBlogID int64) (string, error) {
	base := strings.Trim(desired, "-")
	if base == "" {
		base = "article"
	}
	for suffix := 1; suffix < 10000; suffix++ {
		candidate := base
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}
		var exists bool
		err := tx.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM blog_translations WHERE locale = $1 AND slug = $2 AND blog_id <> $3
				UNION ALL
				SELECT 1 FROM blog_slug_redirects WHERE locale = $1 AND old_slug = $2 AND blog_id <> $3
			)
		`, locale, candidate, excludingBlogID).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not create a unique blog slug")
}

func firstTranslationTitle(translations []domain.BlogTranslation) string {
	for _, translation := range translations {
		if strings.TrimSpace(translation.Title) != "" {
			return strings.TrimSpace(translation.Title)
		}
	}
	return "Untitled article"
}

func (r *BlogRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM blogs WHERE id = $1`, id)
	return err
}
