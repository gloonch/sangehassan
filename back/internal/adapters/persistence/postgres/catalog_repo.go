package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"sangehassan/back/internal/domain"
)

type CatalogRepository struct {
	db *sql.DB
}

func NewCatalogRepository(db *sql.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) ListCategories(ctx context.Context) ([]domain.CatalogCategory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id,
		       COALESCE(c.image_url, ''), COALESCE(c.intro_en, ''), COALESCE(c.intro_fa, ''), COALESCE(c.intro_ar, ''),
		       COALESCE(c.seo_title_en, ''), COALESCE(c.seo_title_fa, ''), COALESCE(c.seo_title_ar, ''),
		       COALESCE(c.seo_description_en, ''), COALESCE(c.seo_description_fa, ''), COALESCE(c.seo_description_ar, ''),
		       c.is_active, c.is_indexable, c.created_at, COALESCE(c.updated_at, c.created_at),
		       COUNT(DISTINCT p.id)::int,
		       COALESCE(NULLIF(c.image_url, ''), (ARRAY_AGG(p.image_url ORDER BY p.is_popular DESC, p.id) FILTER (WHERE COALESCE(p.image_url, '') <> ''))[1], '')
		FROM categories c
		LEFT JOIN products p ON p.is_active = TRUE AND p.is_indexable = TRUE AND (
			p.main_category_id = c.id OR EXISTS (
				SELECT 1 FROM product_categories pc WHERE pc.product_id = p.id AND pc.category_id = c.id
			)
		)
		WHERE c.is_active = TRUE AND c.parent_id IS NULL
		GROUP BY c.id
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY c.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.CatalogCategory
	for rows.Next() {
		var category domain.CatalogCategory
		if err := scanCatalogCategory(rows, &category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (r *CatalogRepository) GetCategory(ctx context.Context, slug string) (domain.CatalogCategory, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id,
		       COALESCE(c.image_url, ''), COALESCE(c.intro_en, ''), COALESCE(c.intro_fa, ''), COALESCE(c.intro_ar, ''),
		       COALESCE(c.seo_title_en, ''), COALESCE(c.seo_title_fa, ''), COALESCE(c.seo_title_ar, ''),
		       COALESCE(c.seo_description_en, ''), COALESCE(c.seo_description_fa, ''), COALESCE(c.seo_description_ar, ''),
		       c.is_active, c.is_indexable, c.created_at, COALESCE(c.updated_at, c.created_at),
		       COUNT(DISTINCT p.id)::int,
		       COALESCE(NULLIF(c.image_url, ''), (ARRAY_AGG(p.image_url ORDER BY p.is_popular DESC, p.id) FILTER (WHERE COALESCE(p.image_url, '') <> ''))[1], '')
		FROM categories c
		LEFT JOIN products p ON p.is_active = TRUE AND p.is_indexable = TRUE AND (
			p.main_category_id = c.id OR EXISTS (
				SELECT 1 FROM product_categories pc WHERE pc.product_id = p.id AND pc.category_id = c.id
			)
		)
		WHERE c.slug = $1 AND c.is_active = TRUE
		GROUP BY c.id
		HAVING COUNT(DISTINCT p.id) > 0
	`, slug)
	var category domain.CatalogCategory
	if err := scanCatalogCategory(row, &category); err != nil {
		return domain.CatalogCategory{}, err
	}
	return category, nil
}

func (r *CatalogRepository) ListProducts(ctx context.Context, categoryID int64, filters map[string][]string, limit, offset int) ([]domain.Product, int, error) {
	where, args := catalogProductWhere(categoryID, filters, "p")
	countQuery := `SELECT COUNT(DISTINCT p.id) FROM products p WHERE ` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT DISTINCT p.id, p.title_en, p.title_fa, p.title_ar, p.slug,
		       COALESCE(p.short_description_html_en, p.short_description_html, ''),
		       COALESCE(p.short_description_html_fa, ''), COALESCE(p.short_description_html_ar, ''),
		       COALESCE(p.price, 0), COALESCE(p.price_html, ''), COALESCE(p.image_url, ''),
		       p.main_category_id, p.is_popular, p.is_active, p.is_indexable,
		       p.created_at, COALESCE(p.updated_at, p.created_at)
		FROM products p
		WHERE %s
		ORDER BY p.is_popular DESC, p.id
		LIMIT $%d OFFSET $%d
	`, where, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		if err := rows.Scan(
			&product.ID, &product.TitleEN, &product.TitleFA, &product.TitleAR, &product.Slug,
			&product.ShortDescriptionHTMLEn, &product.ShortDescriptionHTMLFa, &product.ShortDescriptionHTMLAr,
			&product.Price, &product.PriceHTML, &product.ImageURL, &product.MainCategoryID,
			&product.IsPopular, &product.IsActive, &product.IsIndexable, &product.CreatedAt, &product.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		product.ShortDescriptionHTML = product.ShortDescriptionHTMLEn
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := r.attachCatalogTerms(ctx, products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *CatalogRepository) ListFacetValues(ctx context.Context, categoryID int64, taxonomy string, filters map[string][]string) ([]domain.CatalogFacetValue, error) {
	otherFilters := make(map[string][]string, len(filters))
	for key, values := range filters {
		if key != taxonomy {
			otherFilters[key] = values
		}
	}
	where, args := catalogProductWhere(categoryID, otherFilters, "p")
	args = append(args, taxonomy)
	query := fmt.Sprintf(`
		SELECT t.id, t.term_key, t.label_en, t.label_fa, t.label_ar, COUNT(DISTINCT p.id)::int, t.is_indexable
		FROM products p
		JOIN product_term_links ptl ON ptl.product_id = p.id
		JOIN product_terms t ON t.id = ptl.term_id AND t.is_active = TRUE
		WHERE %s AND t.taxonomy = $%d
		GROUP BY t.id
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY COUNT(DISTINCT p.id) DESC, t.id
	`, where, len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []domain.CatalogFacetValue
	for rows.Next() {
		var value domain.CatalogFacetValue
		if err := rows.Scan(&value.ID, &value.Key, &value.LabelEN, &value.LabelFA, &value.LabelAR, &value.Count, &value.IsIndexable); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func (r *CatalogRepository) GetFacetValue(ctx context.Context, categoryID int64, taxonomy, key string) (domain.CatalogFacetValue, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT t.id, t.term_key, t.label_en, t.label_fa, t.label_ar, COUNT(DISTINCT p.id)::int, t.is_indexable
		FROM product_terms t
		JOIN product_term_links ptl ON ptl.term_id = t.id
		JOIN products p ON p.id = ptl.product_id AND p.is_active = TRUE AND p.is_indexable = TRUE
		WHERE t.taxonomy = $1 AND t.term_key = $2 AND t.is_active = TRUE
		  AND (p.main_category_id = $3 OR EXISTS (
			SELECT 1 FROM product_categories pc WHERE pc.product_id = p.id AND pc.category_id = $3
		  ))
		GROUP BY t.id
	`, taxonomy, key, categoryID)
	var value domain.CatalogFacetValue
	if err := row.Scan(&value.ID, &value.Key, &value.LabelEN, &value.LabelFA, &value.LabelAR, &value.Count, &value.IsIndexable); err != nil {
		return domain.CatalogFacetValue{}, err
	}
	return value, nil
}

func (r *CatalogRepository) GetFacetPage(ctx context.Context, categoryID, termID int64) (domain.CatalogFacetPage, error) {
	var page domain.CatalogFacetPage
	err := r.db.QueryRowContext(ctx, `
		SELECT id, category_id, term_id,
		       COALESCE(title_en, ''), COALESCE(title_fa, ''), COALESCE(title_ar, ''),
		       COALESCE(description_en, ''), COALESCE(description_fa, ''), COALESCE(description_ar, ''),
		       COALESCE(h1_en, ''), COALESCE(h1_fa, ''), COALESCE(h1_ar, ''),
		       COALESCE(intro_en, ''), COALESCE(intro_fa, ''), COALESCE(intro_ar, ''),
		       is_active, is_indexable
		FROM catalog_facet_pages
		WHERE category_id = $1 AND term_id = $2
	`, categoryID, termID).Scan(
		&page.ID, &page.CategoryID, &page.TermID,
		&page.TitleEN, &page.TitleFA, &page.TitleAR,
		&page.DescriptionEN, &page.DescriptionFA, &page.DescriptionAR,
		&page.H1EN, &page.H1FA, &page.H1AR,
		&page.IntroEN, &page.IntroFA, &page.IntroAR,
		&page.IsActive, &page.IsIndexable,
	)
	return page, err
}

func (r *CatalogRepository) ListFacetPages(ctx context.Context) ([]domain.CatalogFacetPage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT fp.id, fp.category_id, fp.term_id, c.slug, t.taxonomy, t.term_key,
		       COALESCE(fp.title_en, ''), COALESCE(fp.title_fa, ''), COALESCE(fp.title_ar, ''),
		       COALESCE(fp.description_en, ''), COALESCE(fp.description_fa, ''), COALESCE(fp.description_ar, ''),
		       COALESCE(fp.h1_en, ''), COALESCE(fp.h1_fa, ''), COALESCE(fp.h1_ar, ''),
		       COALESCE(fp.intro_en, ''), COALESCE(fp.intro_fa, ''), COALESCE(fp.intro_ar, ''),
		       fp.is_active, fp.is_indexable
		FROM catalog_facet_pages fp
		JOIN categories c ON c.id = fp.category_id
		JOIN product_terms t ON t.id = fp.term_id
		ORDER BY c.id, t.taxonomy, t.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := make([]domain.CatalogFacetPage, 0)
	for rows.Next() {
		var page domain.CatalogFacetPage
		if err := rows.Scan(
			&page.ID, &page.CategoryID, &page.TermID, &page.CategorySlug, &page.Taxonomy, &page.TermKey,
			&page.TitleEN, &page.TitleFA, &page.TitleAR,
			&page.DescriptionEN, &page.DescriptionFA, &page.DescriptionAR,
			&page.H1EN, &page.H1FA, &page.H1AR,
			&page.IntroEN, &page.IntroFA, &page.IntroAR,
			&page.IsActive, &page.IsIndexable,
		); err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, rows.Err()
}

func (r *CatalogRepository) UpsertFacetPage(ctx context.Context, page domain.CatalogFacetPage) (domain.CatalogFacetPage, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO catalog_facet_pages (
			category_id, term_id, title_en, title_fa, title_ar,
			description_en, description_fa, description_ar,
			h1_en, h1_fa, h1_ar, intro_en, intro_fa, intro_ar, is_active, is_indexable
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (category_id, term_id) DO UPDATE SET
			title_en = EXCLUDED.title_en, title_fa = EXCLUDED.title_fa, title_ar = EXCLUDED.title_ar,
			description_en = EXCLUDED.description_en, description_fa = EXCLUDED.description_fa, description_ar = EXCLUDED.description_ar,
			h1_en = EXCLUDED.h1_en, h1_fa = EXCLUDED.h1_fa, h1_ar = EXCLUDED.h1_ar,
			intro_en = EXCLUDED.intro_en, intro_fa = EXCLUDED.intro_fa, intro_ar = EXCLUDED.intro_ar,
			is_active = EXCLUDED.is_active, is_indexable = EXCLUDED.is_indexable, updated_at = NOW()
		RETURNING id
	`, page.CategoryID, page.TermID,
		nullableString(page.TitleEN), nullableString(page.TitleFA), nullableString(page.TitleAR),
		nullableString(page.DescriptionEN), nullableString(page.DescriptionFA), nullableString(page.DescriptionAR),
		nullableString(page.H1EN), nullableString(page.H1FA), nullableString(page.H1AR),
		nullableString(page.IntroEN), nullableString(page.IntroFA), nullableString(page.IntroAR),
		page.IsActive, page.IsIndexable)
	if err := row.Scan(&page.ID); err != nil {
		return domain.CatalogFacetPage{}, err
	}
	return page, nil
}

func (r *CatalogRepository) DeleteFacetPage(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM catalog_facet_pages WHERE id = $1`, id)
	return err
}

func (r *CatalogRepository) ListRelatedProjects(ctx context.Context, categoryID int64, limit int) ([]domain.ProjectCard, error) {
	if limit < 1 {
		limit = 6
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT pr.id, COALESCE(pr.cover_image_url, ''), pr.sort_order
		FROM projects pr
		JOIN project_products pp ON pp.project_id = pr.id
		JOIN products p ON p.id = pp.product_id AND p.is_active = TRUE AND p.is_indexable = TRUE
		WHERE p.main_category_id = $1 OR EXISTS (
			SELECT 1 FROM product_categories pc WHERE pc.product_id = p.id AND pc.category_id = $1
		)
		ORDER BY pr.sort_order ASC, pr.id DESC
		LIMIT $2
	`, categoryID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]domain.ProjectCard, 0)
	for rows.Next() {
		var project domain.ProjectCard
		if err := rows.Scan(&project.ID, &project.CoverImageURL, &project.SortOrder); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (r *CatalogRepository) ListIndexableRoutes(ctx context.Context, minimumProducts int) ([]domain.CatalogRoute, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.slug, COUNT(DISTINCT p.id)::int, c.is_indexable
		FROM categories c
		JOIN products p ON p.is_active = TRUE AND p.is_indexable = TRUE AND (
			p.main_category_id = c.id OR EXISTS (
				SELECT 1 FROM product_categories pc WHERE pc.product_id = p.id AND pc.category_id = c.id
			)
		)
		WHERE c.is_active = TRUE AND c.parent_id IS NULL
		GROUP BY c.id
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY c.id
	`)
	if err != nil {
		return nil, err
	}
	var routes []domain.CatalogRoute
	for rows.Next() {
		var route domain.CatalogRoute
		if err := rows.Scan(&route.CategorySlug, &route.ProductCount, &route.Indexable); err != nil {
			rows.Close()
			return nil, err
		}
		route.Type = "category"
		route.Path = "/fa/products/" + route.CategorySlug
		routes = append(routes, route)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	facetRows, err := r.db.QueryContext(ctx, `
		SELECT c.slug, t.taxonomy, t.term_key, COUNT(DISTINCT p.id)::int,
		       c.is_indexable, (t.is_indexable AND COALESCE(cfp.is_indexable, TRUE))
		FROM categories c
		JOIN products p ON p.is_active = TRUE AND p.is_indexable = TRUE AND (
			p.main_category_id = c.id OR EXISTS (
				SELECT 1 FROM product_categories pc WHERE pc.product_id = p.id AND pc.category_id = c.id
			)
		)
		JOIN product_term_links ptl ON ptl.product_id = p.id
		JOIN product_terms t ON t.id = ptl.term_id AND t.is_active = TRUE
		LEFT JOIN catalog_facet_pages cfp ON cfp.category_id = c.id AND cfp.term_id = t.id
		WHERE c.is_active = TRUE AND c.parent_id IS NULL
		  AND t.taxonomy = ANY($1)
		  AND COALESCE(cfp.is_active, TRUE) = TRUE
		GROUP BY c.id, t.id, cfp.is_indexable
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY c.id, t.taxonomy, t.id
	`, pq.Array(catalogTaxonomies()))
	if err != nil {
		return nil, err
	}
	defer facetRows.Close()
	for facetRows.Next() {
		var route domain.CatalogRoute
		var taxonomy string
		var categoryIndexable, termIndexable bool
		if err := facetRows.Scan(&route.CategorySlug, &taxonomy, &route.Value, &route.ProductCount, &categoryIndexable, &termIndexable); err != nil {
			return nil, err
		}
		route.Facet = facetRouteKey(taxonomy)
		if route.Facet == "" {
			continue
		}
		route.Type = "facet"
		route.Indexable = categoryIndexable && termIndexable && route.ProductCount >= minimumProducts
		route.Path = fmt.Sprintf("/fa/products/%s/%s/%s", route.CategorySlug, route.Facet, route.Value)
		routes = append(routes, route)
	}
	return routes, facetRows.Err()
}

func catalogProductWhere(categoryID int64, filters map[string][]string, alias string) (string, []any) {
	args := []any{categoryID}
	parts := []string{
		fmt.Sprintf("%s.is_active = TRUE", alias),
		fmt.Sprintf("%s.is_indexable = TRUE", alias),
		fmt.Sprintf("(%s.main_category_id = $1 OR EXISTS (SELECT 1 FROM product_categories pc WHERE pc.product_id = %s.id AND pc.category_id = $1))", alias, alias),
	}
	for taxonomy, values := range filters {
		if len(values) == 0 {
			continue
		}
		if taxonomy == "__search" {
			args = append(args, "%"+strings.Join(values, " ")+"%")
			parts = append(parts, fmt.Sprintf(`(
				%s.title_en ILIKE $%d OR %s.title_fa ILIKE $%d OR %s.title_ar ILIKE $%d OR %s.slug ILIKE $%d
			)`, alias, len(args), alias, len(args), alias, len(args), alias, len(args)))
			continue
		}
		args = append(args, taxonomy, pq.Array(values))
		parts = append(parts, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM product_term_links selected_ptl
			JOIN product_terms selected_t ON selected_t.id = selected_ptl.term_id AND selected_t.is_active = TRUE
			WHERE selected_ptl.product_id = %s.id AND selected_t.taxonomy = $%d AND selected_t.term_key = ANY($%d)
		)`, alias, len(args)-1, len(args)))
	}
	return strings.Join(parts, " AND "), args
}

func (r *CatalogRepository) attachCatalogTerms(ctx context.Context, products []domain.Product) error {
	if len(products) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(products))
	for _, product := range products {
		ids = append(ids, product.ID)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT ptl.product_id, t.id, t.taxonomy, t.term_key, t.label_en, t.label_fa, t.label_ar,
		       COALESCE(t.link_url, ''), t.is_active, t.is_indexable
		FROM product_term_links ptl
		JOIN product_terms t ON t.id = ptl.term_id AND t.is_active = TRUE
		WHERE ptl.product_id = ANY($1)
		ORDER BY ptl.product_id, t.taxonomy, t.id
	`, pq.Array(ids))
	if err != nil {
		return err
	}
	defer rows.Close()
	byID := make(map[int64][]domain.ProductTerm)
	for rows.Next() {
		var productID int64
		var term domain.ProductTerm
		if err := rows.Scan(&productID, &term.ID, &term.Taxonomy, &term.Key, &term.LabelEN, &term.LabelFA, &term.LabelAR, &term.LinkURL, &term.IsActive, &term.IsIndexable); err != nil {
			return err
		}
		byID[productID] = append(byID[productID], term)
	}
	for i := range products {
		products[i].Terms = byID[products[i].ID]
	}
	return rows.Err()
}

type catalogCategoryScanner interface {
	Scan(dest ...any) error
}

func scanCatalogCategory(scanner catalogCategoryScanner, category *domain.CatalogCategory) error {
	return scanner.Scan(
		&category.ID, &category.TitleEN, &category.TitleFA, &category.TitleAR, &category.Slug, &category.ParentID,
		&category.ImageURL, &category.IntroEN, &category.IntroFA, &category.IntroAR,
		&category.SEOTitleEN, &category.SEOTitleFA, &category.SEOTitleAR,
		&category.SEODescriptionEN, &category.SEODescriptionFA, &category.SEODescriptionAR,
		&category.IsActive, &category.IsIndexable, &category.CreatedAt, &category.UpdatedAt,
		&category.ProductCount, &category.PreviewImage,
	)
}

func catalogTaxonomies() []string {
	return []string{"tone", "use_case_application", "finishes", "use_case_form", "mines", "pattern", "availability"}
}

func facetRouteKey(taxonomy string) string {
	switch taxonomy {
	case "tone":
		return "color"
	case "use_case_application":
		return "application"
	case "finishes":
		return "finish"
	case "use_case_form":
		return "form"
	case "mines":
		return "origin"
	case "pattern":
		return "pattern"
	case "availability":
		return "availability"
	default:
		return ""
	}
}
