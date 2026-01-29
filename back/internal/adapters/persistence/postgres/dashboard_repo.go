package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetStats(ctx context.Context) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	row := r.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM products) AS products_count,
			(SELECT COUNT(*) FROM categories) AS categories_count,
			(SELECT COUNT(*) FROM templates) AS templates_count,
			(SELECT COUNT(*) FROM blogs) AS blogs_count
	`)
	if err := row.Scan(&stats.ProductsCount, &stats.CategoriesCount, &stats.TemplatesCount, &stats.BlogsCount); err != nil {
		return domain.DashboardStats{}, err
	}

	usageRows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.title_en, c.title_fa, c.title_ar, c.slug, COALESCE(COUNT(pc.product_id), 0) AS product_count
		FROM categories c
		LEFT JOIN product_categories pc ON pc.category_id = c.id
		GROUP BY c.id
		ORDER BY product_count DESC, c.id
	`)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	defer usageRows.Close()

	for usageRows.Next() {
		var item domain.CategoryUsage
		if err := usageRows.Scan(&item.ID, &item.TitleEN, &item.TitleFA, &item.TitleAR, &item.Slug, &item.ProductCount); err != nil {
			return domain.DashboardStats{}, err
		}
		stats.CategoryUsage = append(stats.CategoryUsage, item)
	}
	if err := usageRows.Err(); err != nil {
		return domain.DashboardStats{}, err
	}

	templateRows, err := r.db.QueryContext(ctx, `
		SELECT id, name, image_url, is_active, created_at, COALESCE(updated_at, created_at)
		FROM templates
		ORDER BY id
	`)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	defer templateRows.Close()

	for templateRows.Next() {
		var template domain.Template
		if err := templateRows.Scan(&template.ID, &template.Name, &template.ImageURL, &template.IsActive, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return domain.DashboardStats{}, err
		}
		stats.Templates = append(stats.Templates, template)
	}
	if err := templateRows.Err(); err != nil {
		return domain.DashboardStats{}, err
	}

	return stats, nil
}
