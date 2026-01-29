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
		SELECT id, title_en, title_fa, title_ar, slug, parent_id, created_at, COALESCE(updated_at, created_at)
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
		if err := rows.Scan(&category.ID, &category.TitleEN, &category.TitleFA, &category.TitleAR, &category.Slug, &category.ParentID, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (domain.Category, error) {
	var category domain.Category
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title_en, title_fa, title_ar, slug, parent_id, created_at, COALESCE(updated_at, created_at)
		FROM categories
		WHERE id = $1
	`, id)
	if err := row.Scan(&category.ID, &category.TitleEN, &category.TitleFA, &category.TitleAR, &category.Slug, &category.ParentID, &category.CreatedAt, &category.UpdatedAt); err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

func (r *CategoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO categories (title_en, title_fa, title_ar, slug, parent_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`, category.TitleEN, category.TitleFA, category.TitleAR, category.Slug, category.ParentID)

	if err := row.Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt); err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category domain.Category) (domain.Category, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE categories
		SET title_en = $1, title_fa = $2, title_ar = $3, slug = $4, parent_id = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING created_at, COALESCE(updated_at, created_at)
	`, category.TitleEN, category.TitleFA, category.TitleAR, category.Slug, category.ParentID, category.ID)

	if err := row.Scan(&category.CreatedAt, &category.UpdatedAt); err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}
