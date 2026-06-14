package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title_en, title_fa, title_ar, slug, parent_id,
		       COALESCE(image_url, ''), COALESCE(intro_en, ''), COALESCE(intro_fa, ''), COALESCE(intro_ar, ''),
		       COALESCE(seo_title_en, ''), COALESCE(seo_title_fa, ''), COALESCE(seo_title_ar, ''),
		       COALESCE(seo_description_en, ''), COALESCE(seo_description_fa, ''), COALESCE(seo_description_ar, ''),
		       is_active, is_indexable, created_at, COALESCE(updated_at, created_at)
		FROM categories
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var category domain.Category
		if err := scanCategory(rows, &category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (domain.Category, error) {
	var category domain.Category
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title_en, title_fa, title_ar, slug, parent_id,
		       COALESCE(image_url, ''), COALESCE(intro_en, ''), COALESCE(intro_fa, ''), COALESCE(intro_ar, ''),
		       COALESCE(seo_title_en, ''), COALESCE(seo_title_fa, ''), COALESCE(seo_title_ar, ''),
		       COALESCE(seo_description_en, ''), COALESCE(seo_description_fa, ''), COALESCE(seo_description_ar, ''),
		       is_active, is_indexable, created_at, COALESCE(updated_at, created_at)
		FROM categories
		WHERE id = $1
	`, id)
	if err := scanCategory(row, &category); err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

func (r *CategoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO categories (
			title_en, title_fa, title_ar, slug, parent_id, image_url,
			intro_en, intro_fa, intro_ar, seo_title_en, seo_title_fa, seo_title_ar,
			seo_description_en, seo_description_fa, seo_description_ar, is_active, is_indexable
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`, category.TitleEN, category.TitleFA, category.TitleAR, category.Slug, category.ParentID, nullableString(category.ImageURL),
		category.IntroEN, category.IntroFA, category.IntroAR, category.SEOTitleEN, category.SEOTitleFA, category.SEOTitleAR,
		category.SEODescriptionEN, category.SEODescriptionFA, category.SEODescriptionAR, category.IsActive, category.IsIndexable)

	if err := row.Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt); err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category domain.Category) (domain.Category, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE categories
		SET title_en = $1, title_fa = $2, title_ar = $3, slug = $4, parent_id = $5, image_url = $6,
		    intro_en = $7, intro_fa = $8, intro_ar = $9,
		    seo_title_en = $10, seo_title_fa = $11, seo_title_ar = $12,
		    seo_description_en = $13, seo_description_fa = $14, seo_description_ar = $15,
		    is_active = $16, is_indexable = $17, updated_at = NOW()
		WHERE id = $18
		RETURNING created_at, COALESCE(updated_at, created_at)
	`, category.TitleEN, category.TitleFA, category.TitleAR, category.Slug, category.ParentID, nullableString(category.ImageURL),
		category.IntroEN, category.IntroFA, category.IntroAR, category.SEOTitleEN, category.SEOTitleFA, category.SEOTitleAR,
		category.SEODescriptionEN, category.SEODescriptionFA, category.SEODescriptionAR,
		category.IsActive, category.IsIndexable, category.ID)

	if err := row.Scan(&category.CreatedAt, &category.UpdatedAt); err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

type categoryScanner interface {
	Scan(dest ...any) error
}

func scanCategory(scanner categoryScanner, category *domain.Category) error {
	return scanner.Scan(
		&category.ID, &category.TitleEN, &category.TitleFA, &category.TitleAR, &category.Slug, &category.ParentID,
		&category.ImageURL, &category.IntroEN, &category.IntroFA, &category.IntroAR,
		&category.SEOTitleEN, &category.SEOTitleFA, &category.SEOTitleAR,
		&category.SEODescriptionEN, &category.SEODescriptionFA, &category.SEODescriptionAR,
		&category.IsActive, &category.IsIndexable, &category.CreatedAt, &category.UpdatedAt,
	)
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}
