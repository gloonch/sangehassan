package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"sangehassan/back/internal/domain"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.db.QueryContext(ctx, `
			SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.slug,
			       p.description_html, p.short_description_html,
			       p.description_html_en, p.description_html_fa, p.description_html_ar,
			       p.short_description_html_en, p.short_description_html_fa, p.short_description_html_ar,
			       p.price, p.price_html,
			       p.image_url, p.main_category_id, p.is_popular,
			       (SELECT COUNT(*) FROM product_images pi WHERE pi.product_id = p.id) AS image_count,
			       p.created_at, COALESCE(p.updated_at, p.created_at),
		       c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		ORDER BY p.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		var description sql.NullString
		var shortDescription sql.NullString
		var descriptionEN sql.NullString
		var descriptionFA sql.NullString
		var descriptionAR sql.NullString
		var shortDescriptionEN sql.NullString
		var shortDescriptionFA sql.NullString
		var shortDescriptionAR sql.NullString
		var price sql.NullFloat64
		var priceHTML sql.NullString
		var imageURL sql.NullString
		var mainCategoryID sql.NullInt64
		var imageCount int64
		var categoryID sql.NullInt64
		var categoryTitleEN sql.NullString
		var categoryTitleFA sql.NullString
		var categoryTitleAR sql.NullString
		var categorySlug sql.NullString
		var categoryParentID sql.NullInt64
		if err := rows.Scan(
			&product.ID,
			&product.TitleEN,
			&product.TitleFA,
			&product.TitleAR,
			&product.Slug,
			&description,
			&shortDescription,
			&descriptionEN,
			&descriptionFA,
			&descriptionAR,
			&shortDescriptionEN,
			&shortDescriptionFA,
			&shortDescriptionAR,
			&price,
			&priceHTML,
			&imageURL,
			&mainCategoryID,
			&product.IsPopular,
			&imageCount,
			&product.CreatedAt,
			&product.UpdatedAt,
			&categoryID,
			&categoryTitleEN,
			&categoryTitleFA,
			&categoryTitleAR,
			&categorySlug,
			&categoryParentID,
		); err != nil {
			return nil, err
		}
		product.DescriptionHTML = description.String
		product.Description = product.DescriptionHTML
		product.ShortDescriptionHTML = shortDescription.String
		product.DescriptionHTMLEn = descriptionEN.String
		product.DescriptionHTMLFa = descriptionFA.String
		product.DescriptionHTMLAr = descriptionAR.String
		product.ShortDescriptionHTMLEn = shortDescriptionEN.String
		product.ShortDescriptionHTMLFa = shortDescriptionFA.String
		product.ShortDescriptionHTMLAr = shortDescriptionAR.String
		applyDescriptionFallbacks(&product)
		if price.Valid {
			product.Price = price.Float64
		}
		product.PriceHTML = priceHTML.String
		product.ImageURL = imageURL.String
		product.ImageCount = int(imageCount)
		if product.ImageCount == 0 && product.ImageURL != "" {
			product.ImageCount = 1
		}
		if mainCategoryID.Valid {
			product.MainCategoryID = &mainCategoryID.Int64
		}
		if categoryID.Valid {
			category := domain.Category{
				ID:      categoryID.Int64,
				TitleEN: categoryTitleEN.String,
				TitleFA: categoryTitleFA.String,
				TitleAR: categoryTitleAR.String,
				Slug:    categorySlug.String,
			}
			if categoryParentID.Valid {
				category.ParentID = &categoryParentID.Int64
			}
			product.Category = &category
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.attachTermsToProducts(ctx, products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) ListPopular(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.db.QueryContext(ctx, `
			SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.slug,
			       p.description_html, p.short_description_html,
			       p.description_html_en, p.description_html_fa, p.description_html_ar,
			       p.short_description_html_en, p.short_description_html_fa, p.short_description_html_ar,
			       p.price, p.price_html,
			       p.image_url, p.main_category_id, p.is_popular,
			       (SELECT COUNT(*) FROM product_images pi WHERE pi.product_id = p.id) AS image_count,
			       p.created_at, COALESCE(p.updated_at, p.created_at),
		       c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE p.is_popular = TRUE
		ORDER BY p.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		var description sql.NullString
		var shortDescription sql.NullString
		var descriptionEN sql.NullString
		var descriptionFA sql.NullString
		var descriptionAR sql.NullString
		var shortDescriptionEN sql.NullString
		var shortDescriptionFA sql.NullString
		var shortDescriptionAR sql.NullString
		var price sql.NullFloat64
		var priceHTML sql.NullString
		var imageURL sql.NullString
		var mainCategoryID sql.NullInt64
		var imageCount int64
		var categoryID sql.NullInt64
		var categoryTitleEN sql.NullString
		var categoryTitleFA sql.NullString
		var categoryTitleAR sql.NullString
		var categorySlug sql.NullString
		var categoryParentID sql.NullInt64
		if err := rows.Scan(
			&product.ID,
			&product.TitleEN,
			&product.TitleFA,
			&product.TitleAR,
			&product.Slug,
			&description,
			&shortDescription,
			&descriptionEN,
			&descriptionFA,
			&descriptionAR,
			&shortDescriptionEN,
			&shortDescriptionFA,
			&shortDescriptionAR,
			&price,
			&priceHTML,
			&imageURL,
			&mainCategoryID,
			&product.IsPopular,
			&imageCount,
			&product.CreatedAt,
			&product.UpdatedAt,
			&categoryID,
			&categoryTitleEN,
			&categoryTitleFA,
			&categoryTitleAR,
			&categorySlug,
			&categoryParentID,
		); err != nil {
			return nil, err
		}
		product.DescriptionHTML = description.String
		product.Description = product.DescriptionHTML
		product.ShortDescriptionHTML = shortDescription.String
		product.DescriptionHTMLEn = descriptionEN.String
		product.DescriptionHTMLFa = descriptionFA.String
		product.DescriptionHTMLAr = descriptionAR.String
		product.ShortDescriptionHTMLEn = shortDescriptionEN.String
		product.ShortDescriptionHTMLFa = shortDescriptionFA.String
		product.ShortDescriptionHTMLAr = shortDescriptionAR.String
		applyDescriptionFallbacks(&product)
		if price.Valid {
			product.Price = price.Float64
		}
		product.PriceHTML = priceHTML.String
		product.ImageURL = imageURL.String
		product.ImageCount = int(imageCount)
		if product.ImageCount == 0 && product.ImageURL != "" {
			product.ImageCount = 1
		}
		if mainCategoryID.Valid {
			product.MainCategoryID = &mainCategoryID.Int64
		}
		if categoryID.Valid {
			category := domain.Category{
				ID:      categoryID.Int64,
				TitleEN: categoryTitleEN.String,
				TitleFA: categoryTitleFA.String,
				TitleAR: categoryTitleAR.String,
				Slug:    categorySlug.String,
			}
			if categoryParentID.Valid {
				category.ParentID = &categoryParentID.Int64
			}
			product.Category = &category
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.attachTermsToProducts(ctx, products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (domain.Product, error) {
	row := r.db.QueryRowContext(ctx, `
			SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.slug,
			       p.description_html, p.short_description_html,
			       p.description_html_en, p.description_html_fa, p.description_html_ar,
			       p.short_description_html_en, p.short_description_html_fa, p.short_description_html_ar,
			       p.price, p.price_html,
			       p.image_url, p.main_category_id, p.is_popular,
			       (SELECT COUNT(*) FROM product_images pi WHERE pi.product_id = p.id) AS image_count,
			       p.created_at, COALESCE(p.updated_at, p.created_at),
		       c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE p.id = $1
	`, id)

	var product domain.Product
	var description sql.NullString
	var shortDescription sql.NullString
	var descriptionEN sql.NullString
	var descriptionFA sql.NullString
	var descriptionAR sql.NullString
	var shortDescriptionEN sql.NullString
	var shortDescriptionFA sql.NullString
	var shortDescriptionAR sql.NullString
	var price sql.NullFloat64
	var priceHTML sql.NullString
	var imageURL sql.NullString
	var mainCategoryID sql.NullInt64
	var imageCount int64
	var categoryID sql.NullInt64
	var categoryTitleEN sql.NullString
	var categoryTitleFA sql.NullString
	var categoryTitleAR sql.NullString
	var categorySlug sql.NullString
	var categoryParentID sql.NullInt64

	if err := row.Scan(
		&product.ID,
		&product.TitleEN,
		&product.TitleFA,
		&product.TitleAR,
		&product.Slug,
		&description,
		&shortDescription,
		&descriptionEN,
		&descriptionFA,
		&descriptionAR,
		&shortDescriptionEN,
		&shortDescriptionFA,
		&shortDescriptionAR,
		&price,
		&priceHTML,
		&imageURL,
		&mainCategoryID,
		&product.IsPopular,
		&imageCount,
		&product.CreatedAt,
		&product.UpdatedAt,
		&categoryID,
		&categoryTitleEN,
		&categoryTitleFA,
		&categoryTitleAR,
		&categorySlug,
		&categoryParentID,
	); err != nil {
		return domain.Product{}, err
	}

	product.DescriptionHTML = description.String
	product.Description = product.DescriptionHTML
	product.ShortDescriptionHTML = shortDescription.String
	product.DescriptionHTMLEn = descriptionEN.String
	product.DescriptionHTMLFa = descriptionFA.String
	product.DescriptionHTMLAr = descriptionAR.String
	product.ShortDescriptionHTMLEn = shortDescriptionEN.String
	product.ShortDescriptionHTMLFa = shortDescriptionFA.String
	product.ShortDescriptionHTMLAr = shortDescriptionAR.String
	applyDescriptionFallbacks(&product)
	if price.Valid {
		product.Price = price.Float64
	}
	product.PriceHTML = priceHTML.String
	product.ImageURL = imageURL.String
	product.ImageCount = int(imageCount)
	if product.ImageCount == 0 && product.ImageURL != "" {
		product.ImageCount = 1
	}
	if mainCategoryID.Valid {
		product.MainCategoryID = &mainCategoryID.Int64
	}
	if categoryID.Valid {
		category := domain.Category{
			ID:      categoryID.Int64,
			TitleEN: categoryTitleEN.String,
			TitleFA: categoryTitleFA.String,
			TitleAR: categoryTitleAR.String,
			Slug:    categorySlug.String,
		}
		if categoryParentID.Valid {
			category.ParentID = &categoryParentID.Int64
		}
		product.Category = &category
	}
	if err := r.loadProductRelations(ctx, &product); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (domain.Product, error) {
	escapedSlug := url.PathEscape(slug)
	row := r.db.QueryRowContext(ctx, `
			SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.slug,
			       p.description_html, p.short_description_html,
			       p.description_html_en, p.description_html_fa, p.description_html_ar,
			       p.short_description_html_en, p.short_description_html_fa, p.short_description_html_ar,
			       p.price, p.price_html,
			       p.image_url, p.main_category_id, p.is_popular,
			       (SELECT COUNT(*) FROM product_images pi WHERE pi.product_id = p.id) AS image_count,
			       p.created_at, COALESCE(p.updated_at, p.created_at),
		       c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE p.slug = $1 OR p.slug = $2
	`, slug, escapedSlug)

	var product domain.Product
	var description sql.NullString
	var shortDescription sql.NullString
	var descriptionEN sql.NullString
	var descriptionFA sql.NullString
	var descriptionAR sql.NullString
	var shortDescriptionEN sql.NullString
	var shortDescriptionFA sql.NullString
	var shortDescriptionAR sql.NullString
	var price sql.NullFloat64
	var priceHTML sql.NullString
	var imageURL sql.NullString
	var mainCategoryID sql.NullInt64
	var imageCount int64
	var categoryID sql.NullInt64
	var categoryTitleEN sql.NullString
	var categoryTitleFA sql.NullString
	var categoryTitleAR sql.NullString
	var categorySlug sql.NullString
	var categoryParentID sql.NullInt64

	if err := row.Scan(
		&product.ID,
		&product.TitleEN,
		&product.TitleFA,
		&product.TitleAR,
		&product.Slug,
		&description,
		&shortDescription,
		&descriptionEN,
		&descriptionFA,
		&descriptionAR,
		&shortDescriptionEN,
		&shortDescriptionFA,
		&shortDescriptionAR,
		&price,
		&priceHTML,
		&imageURL,
		&mainCategoryID,
		&product.IsPopular,
		&imageCount,
		&product.CreatedAt,
		&product.UpdatedAt,
		&categoryID,
		&categoryTitleEN,
		&categoryTitleFA,
		&categoryTitleAR,
		&categorySlug,
		&categoryParentID,
	); err != nil {
		return domain.Product{}, err
	}

	product.DescriptionHTML = description.String
	product.Description = product.DescriptionHTML
	product.ShortDescriptionHTML = shortDescription.String
	product.DescriptionHTMLEn = descriptionEN.String
	product.DescriptionHTMLFa = descriptionFA.String
	product.DescriptionHTMLAr = descriptionAR.String
	product.ShortDescriptionHTMLEn = shortDescriptionEN.String
	product.ShortDescriptionHTMLFa = shortDescriptionFA.String
	product.ShortDescriptionHTMLAr = shortDescriptionAR.String
	applyDescriptionFallbacks(&product)
	if price.Valid {
		product.Price = price.Float64
	}
	product.PriceHTML = priceHTML.String
	product.ImageURL = imageURL.String
	product.ImageCount = int(imageCount)
	if product.ImageCount == 0 && product.ImageURL != "" {
		product.ImageCount = 1
	}
	if mainCategoryID.Valid {
		product.MainCategoryID = &mainCategoryID.Int64
	}
	if categoryID.Valid {
		category := domain.Category{
			ID:      categoryID.Int64,
			TitleEN: categoryTitleEN.String,
			TitleFA: categoryTitleFA.String,
			TitleAR: categoryTitleAR.String,
			Slug:    categorySlug.String,
		}
		if categoryParentID.Valid {
			category.ParentID = &categoryParentID.Int64
		}
		product.Category = &category
	}

	if err := r.loadProductRelations(ctx, &product); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

func (r *ProductRepository) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	applyDescriptionFallbacks(&product)

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO products (
		  title_en, title_fa, title_ar, slug,
		  description_html, short_description_html,
		  description_html_en, description_html_fa, description_html_ar,
		  short_description_html_en, short_description_html_fa, short_description_html_ar,
		  price, price_html, image_url, main_category_id, is_popular
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`,
		product.TitleEN,
		product.TitleFA,
		product.TitleAR,
		product.Slug,
		nullableString(product.DescriptionHTML),
		nullableString(product.ShortDescriptionHTML),
		nullableString(product.DescriptionHTMLEn),
		nullableString(product.DescriptionHTMLFa),
		nullableString(product.DescriptionHTMLAr),
		nullableString(product.ShortDescriptionHTMLEn),
		nullableString(product.ShortDescriptionHTMLFa),
		nullableString(product.ShortDescriptionHTMLAr),
		nullableFloat(product.Price),
		nullableString(product.PriceHTML),
		nullableString(product.ImageURL),
		nullableInt64(product.MainCategoryID),
		product.IsPopular,
	)

	if err := row.Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

func (r *ProductRepository) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	applyDescriptionFallbacks(&product)

	row := r.db.QueryRowContext(ctx, `
		UPDATE products
		SET title_en = $1,
		    title_fa = $2,
		    title_ar = $3,
		    slug = $4,
		    description_html = $5,
		    short_description_html = $6,
		    description_html_en = $7,
		    description_html_fa = $8,
		    description_html_ar = $9,
		    short_description_html_en = $10,
		    short_description_html_fa = $11,
		    short_description_html_ar = $12,
		    price = $13,
		    price_html = $14,
		    image_url = $15,
		    main_category_id = $16,
		    is_popular = $17,
		    updated_at = NOW()
		WHERE id = $18
		RETURNING created_at, COALESCE(updated_at, created_at)
	`,
		product.TitleEN,
		product.TitleFA,
		product.TitleAR,
		product.Slug,
		nullableString(product.DescriptionHTML),
		nullableString(product.ShortDescriptionHTML),
		nullableString(product.DescriptionHTMLEn),
		nullableString(product.DescriptionHTMLFa),
		nullableString(product.DescriptionHTMLAr),
		nullableString(product.ShortDescriptionHTMLEn),
		nullableString(product.ShortDescriptionHTMLFa),
		nullableString(product.ShortDescriptionHTMLAr),
		nullableFloat(product.Price),
		nullableString(product.PriceHTML),
		nullableString(product.ImageURL),
		nullableInt64(product.MainCategoryID),
		product.IsPopular,
		product.ID,
	)

	if err := row.Scan(&product.CreatedAt, &product.UpdatedAt); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

func (r *ProductRepository) ReplaceImages(ctx context.Context, productID int64, images []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM product_images WHERE product_id = $1`, productID); err != nil {
		return err
	}
	for index, url := range images {
		if url == "" {
			continue
		}
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO product_images (product_id, image_url, position)
			VALUES ($1, $2, $3)
			ON CONFLICT (product_id, image_url) DO UPDATE SET position = EXCLUDED.position
		`, productID, url, index)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepository) ReplaceCategories(ctx context.Context, productID int64, categoryIDs []int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM product_categories WHERE product_id = $1`, productID); err != nil {
		return err
	}
	for _, categoryID := range categoryIDs {
		if categoryID == 0 {
			continue
		}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO product_categories (product_id, category_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, productID, categoryID); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepository) ReplaceTerms(ctx context.Context, productID int64, termIDs []int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM product_term_links WHERE product_id = $1`, productID); err != nil {
		return err
	}
	seen := make(map[int64]struct{}, len(termIDs))
	for _, termID := range termIDs {
		if termID == 0 {
			continue
		}
		if _, ok := seen[termID]; ok {
			continue
		}
		seen[termID] = struct{}{}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO product_term_links (product_id, term_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, productID, termID); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepository) loadProductRelations(ctx context.Context, product *domain.Product) error {
	images, err := r.loadImages(ctx, product.ID)
	if err != nil {
		return err
	}
	product.Images = images
	if product.ImageCount == 0 {
		product.ImageCount = len(images)
	}

	categories, err := r.loadCategories(ctx, product.ID)
	if err != nil {
		return err
	}
	product.Categories = categories

	attributes, err := r.loadAttributes(ctx, product.ID)
	if err != nil {
		return err
	}
	product.Attributes = attributes

	terms, err := r.loadTerms(ctx, product.ID)
	if err != nil {
		return err
	}
	product.Terms = terms
	if len(terms) > 0 {
		product.TermIDs = make([]int64, 0, len(terms))
		for _, term := range terms {
			product.TermIDs = append(product.TermIDs, term.ID)
		}
	}
	return nil
}

func (r *ProductRepository) loadImages(ctx context.Context, productID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT image_url
		FROM product_images
		WHERE product_id = $1
		ORDER BY position, id
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		images = append(images, url)
	}
	return images, rows.Err()
}

func (r *ProductRepository) loadCategories(ctx context.Context, productID int64) ([]domain.Category, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id
		FROM product_categories pc
		JOIN categories c ON c.id = pc.category_id
		WHERE pc.product_id = $1
		ORDER BY c.id
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var category domain.Category
		var parentID sql.NullInt64
		if err := rows.Scan(&category.ID, &category.TitleEN, &category.TitleFA, &category.TitleAR, &category.Slug, &parentID); err != nil {
			return nil, err
		}
		if parentID.Valid {
			category.ParentID = &parentID.Int64
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (r *ProductRepository) loadAttributes(ctx context.Context, productID int64) (map[string][]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.name, t.name
		FROM product_attribute_terms pat
		JOIN attribute_terms t ON t.id = pat.attribute_term_id
		JOIN attributes a ON a.id = t.attribute_id
		WHERE pat.product_id = $1
		ORDER BY a.id, t.id
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attributes := make(map[string][]string)
	for rows.Next() {
		var attrName string
		var termName string
		if err := rows.Scan(&attrName, &termName); err != nil {
			return nil, err
		}
		attributes[attrName] = append(attributes[attrName], termName)
	}
	return attributes, rows.Err()
}

func (r *ProductRepository) loadTerms(ctx context.Context, productID int64) ([]domain.ProductTerm, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.taxonomy, t.term_key, t.label_en, t.label_fa, t.label_ar
		FROM product_term_links ptl
		JOIN product_terms t ON t.id = ptl.term_id
		WHERE ptl.product_id = $1
		ORDER BY t.taxonomy, t.id
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []domain.ProductTerm
	for rows.Next() {
		var term domain.ProductTerm
		if err := rows.Scan(&term.ID, &term.Taxonomy, &term.Key, &term.LabelEN, &term.LabelFA, &term.LabelAR); err != nil {
			return nil, err
		}
		terms = append(terms, term)
	}
	return terms, rows.Err()
}

func (r *ProductRepository) attachTermsToProducts(ctx context.Context, products []domain.Product) error {
	if len(products) == 0 {
		return nil
	}

	productIDs := make([]int64, 0, len(products))
	for _, product := range products {
		if product.ID == 0 {
			continue
		}
		productIDs = append(productIDs, product.ID)
	}

	termsByProductID, err := r.loadTermsForProductIDs(ctx, productIDs)
	if err != nil {
		return err
	}

	for i := range products {
		terms := termsByProductID[products[i].ID]
		if len(terms) == 0 {
			continue
		}
		products[i].Terms = terms
		products[i].TermIDs = make([]int64, 0, len(terms))
		for _, term := range terms {
			products[i].TermIDs = append(products[i].TermIDs, term.ID)
		}
	}

	return nil
}

func (r *ProductRepository) loadTermsForProductIDs(ctx context.Context, productIDs []int64) (map[int64][]domain.ProductTerm, error) {
	out := make(map[int64][]domain.ProductTerm)
	if len(productIDs) == 0 {
		return out, nil
	}

	args := make([]any, 0, len(productIDs))
	placeholders := make([]string, 0, len(productIDs))
	for idx, id := range productIDs {
		args = append(args, id)
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx+1))
	}

	query := fmt.Sprintf(`
		SELECT ptl.product_id, t.id, t.taxonomy, t.term_key, t.label_en, t.label_fa, t.label_ar
		FROM product_term_links ptl
		JOIN product_terms t ON t.id = ptl.term_id
		WHERE ptl.product_id IN (%s)
		ORDER BY ptl.product_id, t.taxonomy, t.id
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var productID int64
		var term domain.ProductTerm
		if err := rows.Scan(&productID, &term.ID, &term.Taxonomy, &term.Key, &term.LabelEN, &term.LabelFA, &term.LabelAR); err != nil {
			return nil, err
		}
		out[productID] = append(out[productID], term)
	}

	return out, rows.Err()
}

func applyDescriptionFallbacks(product *domain.Product) {
	// Backward compatibility + graceful migration:
	// - Legacy fields: description_html / short_description_html
	// - New localized fields: *_en/fa/ar
	//
	// Prefer explicit *_en, otherwise fall back to legacy, and keep legacy in sync as EN.
	if strings.TrimSpace(product.DescriptionHTMLEn) == "" {
		product.DescriptionHTMLEn = product.DescriptionHTML
	}
	if strings.TrimSpace(product.ShortDescriptionHTMLEn) == "" {
		product.ShortDescriptionHTMLEn = product.ShortDescriptionHTML
	}

	if strings.TrimSpace(product.DescriptionHTML) == "" {
		product.DescriptionHTML = product.DescriptionHTMLEn
	}
	product.Description = product.DescriptionHTML

	if strings.TrimSpace(product.ShortDescriptionHTML) == "" {
		product.ShortDescriptionHTML = product.ShortDescriptionHTMLEn
	}
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func nullableFloat(value float64) interface{} {
	if value == 0 {
		return nil
	}
	return value
}

func nullableInt64(ptr *int64) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}
