package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/lib/pq"

	"sangehassan/back/internal/domain"
)

type ProductRepository struct {
	db *sql.DB
}

const (
	relatedProductsLimit = 6
	relatedProjectsLimit = 6
)

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]domain.Product, error) {
	query := `
			SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.slug,
			       p.aliases, p.variants, p.mines, p.finishes,
			       p.description_html, p.short_description_html,
			       p.description_html_en, p.description_html_fa, p.description_html_ar,
			       p.short_description_html_en, p.short_description_html_fa, p.short_description_html_ar,
			       p.price, p.price_html,
			       p.image_url, p.video_url, p.main_category_id, p.is_popular, p.is_active, p.is_indexable,
			       (SELECT COUNT(*) FROM product_images pi WHERE pi.product_id = p.id) AS image_count,
			       p.created_at, COALESCE(p.updated_at, p.created_at),
		       c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		ORDER BY p.is_popular DESC, p.id
	`
	args := make([]any, 0, 2)
	if limit > 0 {
		args = append(args, limit)
		query = fmt.Sprintf("%s LIMIT $%d", query, len(args))
		if offset > 0 {
			args = append(args, offset)
			query = fmt.Sprintf("%s OFFSET $%d", query, len(args))
		}
	} else if offset > 0 {
		args = append(args, offset)
		query = fmt.Sprintf("%s OFFSET $%d", query, len(args))
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		var aliases pq.StringArray
		var variants pq.StringArray
		var mines pq.StringArray
		var finishes pq.StringArray
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
		var videoURL sql.NullString
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
			&aliases,
			&variants,
			&mines,
			&finishes,
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
			&videoURL,
			&mainCategoryID,
			&product.IsPopular,
			&product.IsActive,
			&product.IsIndexable,
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
		product.Aliases = aliases
		product.Variants = variants
		product.Mines = mines
		product.Finishes = finishes
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
		product.VideoURL = videoURL.String
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
			       p.aliases, p.variants, p.mines, p.finishes,
			       p.description_html, p.short_description_html,
			       p.description_html_en, p.description_html_fa, p.description_html_ar,
			       p.short_description_html_en, p.short_description_html_fa, p.short_description_html_ar,
			       p.price, p.price_html,
			       p.image_url, p.video_url, p.main_category_id, p.is_popular, p.is_active, p.is_indexable,
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
		var aliases pq.StringArray
		var variants pq.StringArray
		var mines pq.StringArray
		var finishes pq.StringArray
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
		var videoURL sql.NullString
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
			&aliases,
			&variants,
			&mines,
			&finishes,
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
			&videoURL,
			&mainCategoryID,
			&product.IsPopular,
			&product.IsActive,
			&product.IsIndexable,
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
		product.Aliases = aliases
		product.Variants = variants
		product.Mines = mines
		product.Finishes = finishes
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
		product.VideoURL = videoURL.String
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
			       p.aliases, p.variants, p.mines, p.finishes,
			       p.description_html, p.short_description_html,
			       p.description_html_en, p.description_html_fa, p.description_html_ar,
			       p.short_description_html_en, p.short_description_html_fa, p.short_description_html_ar,
			       p.price, p.price_html,
			       p.image_url, p.video_url, p.main_category_id, p.is_popular, p.is_active, p.is_indexable,
			       (SELECT COUNT(*) FROM product_images pi WHERE pi.product_id = p.id) AS image_count,
			       p.created_at, COALESCE(p.updated_at, p.created_at),
		       c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE p.id = $1
	`, id)

	var product domain.Product
	var aliases pq.StringArray
	var variants pq.StringArray
	var mines pq.StringArray
	var finishes pq.StringArray
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
	var videoURL sql.NullString
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
		&aliases,
		&variants,
		&mines,
		&finishes,
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
		&videoURL,
		&mainCategoryID,
		&product.IsPopular,
		&product.IsActive,
		&product.IsIndexable,
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
	product.Aliases = aliases
	product.Variants = variants
	product.Mines = mines
	product.Finishes = finishes
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
	product.VideoURL = videoURL.String
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
			       p.aliases, p.variants, p.mines, p.finishes,
			       p.description_html, p.short_description_html,
			       p.description_html_en, p.description_html_fa, p.description_html_ar,
			       p.short_description_html_en, p.short_description_html_fa, p.short_description_html_ar,
			       p.price, p.price_html,
			       p.image_url, p.video_url, p.main_category_id, p.is_popular, p.is_active, p.is_indexable,
			       (SELECT COUNT(*) FROM product_images pi WHERE pi.product_id = p.id) AS image_count,
			       p.created_at, COALESCE(p.updated_at, p.created_at),
		       c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE (p.slug = $1 OR p.slug = $2) AND p.is_active = TRUE
	`, slug, escapedSlug)

	var product domain.Product
	var aliases pq.StringArray
	var variants pq.StringArray
	var mines pq.StringArray
	var finishes pq.StringArray
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
	var videoURL sql.NullString
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
		&aliases,
		&variants,
		&mines,
		&finishes,
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
		&videoURL,
		&mainCategoryID,
		&product.IsPopular,
		&product.IsActive,
		&product.IsIndexable,
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
	product.Aliases = aliases
	product.Variants = variants
	product.Mines = mines
	product.Finishes = finishes
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
	product.VideoURL = videoURL.String
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
		  aliases, variants, mines, finishes,
		  description_html, short_description_html,
		  description_html_en, description_html_fa, description_html_ar,
		  short_description_html_en, short_description_html_fa, short_description_html_ar,
		  price, price_html, image_url, video_url, main_category_id, is_popular, is_active, is_indexable
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		ON CONFLICT (slug) DO UPDATE SET
		  title_en = EXCLUDED.title_en,
		  title_fa = EXCLUDED.title_fa,
		  title_ar = EXCLUDED.title_ar,
		  aliases = EXCLUDED.aliases,
		  variants = EXCLUDED.variants,
		  mines = EXCLUDED.mines,
		  finishes = EXCLUDED.finishes,
		  description_html = EXCLUDED.description_html,
		  short_description_html = EXCLUDED.short_description_html,
		  description_html_en = EXCLUDED.description_html_en,
		  description_html_fa = EXCLUDED.description_html_fa,
		  description_html_ar = EXCLUDED.description_html_ar,
		  short_description_html_en = EXCLUDED.short_description_html_en,
		  short_description_html_fa = EXCLUDED.short_description_html_fa,
		  short_description_html_ar = EXCLUDED.short_description_html_ar,
		  price = EXCLUDED.price,
		  price_html = EXCLUDED.price_html,
		  image_url = EXCLUDED.image_url,
		  video_url = EXCLUDED.video_url,
		  main_category_id = EXCLUDED.main_category_id,
		  is_popular = EXCLUDED.is_popular,
		  is_active = EXCLUDED.is_active,
		  is_indexable = EXCLUDED.is_indexable,
		  updated_at = NOW()
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`,
		product.TitleEN,
		product.TitleFA,
		product.TitleAR,
		product.Slug,
		pq.Array(product.Aliases),
		pq.Array(product.Variants),
		pq.Array(product.Mines),
		pq.Array(product.Finishes),
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
		nullableString(product.VideoURL),
		nullableInt64(product.MainCategoryID),
		product.IsPopular,
		product.IsActive,
		product.IsIndexable,
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
		    aliases = $5,
		    variants = $6,
		    mines = $7,
		    finishes = $8,
		    description_html = $9,
		    short_description_html = $10,
		    description_html_en = $11,
		    description_html_fa = $12,
		    description_html_ar = $13,
		    short_description_html_en = $14,
		    short_description_html_fa = $15,
		    short_description_html_ar = $16,
		    price = $17,
		    price_html = $18,
		    image_url = $19,
		    video_url = $20,
		    main_category_id = $21,
		    is_popular = $22,
		    is_active = $23,
		    is_indexable = $24,
		    updated_at = NOW()
		WHERE id = $25
		RETURNING created_at, COALESCE(updated_at, created_at)
	`,
		product.TitleEN,
		product.TitleFA,
		product.TitleAR,
		product.Slug,
		pq.Array(product.Aliases),
		pq.Array(product.Variants),
		pq.Array(product.Mines),
		pq.Array(product.Finishes),
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
		nullableString(product.VideoURL),
		nullableInt64(product.MainCategoryID),
		product.IsPopular,
		product.IsActive,
		product.IsIndexable,
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
	relatedProducts, err := r.loadRelatedProducts(ctx, product)
	if err != nil {
		return err
	}
	product.RelatedProducts = relatedProducts

	relatedProjects, err := r.loadRelatedProjects(ctx, product.ID)
	if err != nil {
		return err
	}
	product.RelatedProjects = relatedProjects
	return nil
}

func (r *ProductRepository) loadRelatedProducts(ctx context.Context, product *domain.Product) ([]domain.Product, error) {
	whereClause := "p.is_popular = TRUE"
	args := []any{product.ID, relatedProductsLimit}
	if product.MainCategoryID != nil {
		whereClause = "p.main_category_id = $3"
		args = append(args, *product.MainCategoryID)
	}

	query := fmt.Sprintf(`
		SELECT p.id,
		       p.title_en,
		       p.title_fa,
		       p.title_ar,
		       p.slug,
		       COALESCE(p.image_url, ''),
		       p.main_category_id,
		       p.is_popular,
		       (SELECT COUNT(*) FROM product_images pi WHERE pi.product_id = p.id) AS image_count,
		       c.id,
		       c.title_en,
		       c.title_fa,
		       c.title_ar,
		       c.slug,
		       c.parent_id
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE p.id <> $1 AND %s
		ORDER BY
		  CASE WHEN COALESCE(p.image_url, '') <> '' THEN 0 ELSE 1 END,
		  p.is_popular DESC,
		  CASE WHEN COALESCE(p.short_description_html_en, p.short_description_html, p.description_html_en, p.description_html, '') <> '' THEN 0 ELSE 1 END,
		  p.id DESC
		LIMIT $2
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	for rows.Next() {
		product, err := scanProductCard(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func (r *ProductRepository) loadRelatedProjects(ctx context.Context, productID int64) ([]domain.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id,
		       COALESCE(p.description_en, p.description, ''),
		       COALESCE(p.description_fa, ''),
		       COALESCE(p.description_ar, ''),
		       COALESCE(p.cover_image_url, ''),
		       COALESCE(p.video_url, ''),
		       p.sort_order,
		       p.created_at,
		       COALESCE(p.updated_at, p.created_at)
		FROM project_products pp
		JOIN projects p ON p.id = pp.project_id
		WHERE pp.product_id = $1
		ORDER BY p.sort_order ASC, p.id DESC
		LIMIT $2
	`, productID, relatedProjectsLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]domain.Project, 0)
	for rows.Next() {
		var project domain.Project
		if err := rows.Scan(
			&project.ID,
			&project.DescriptionEN,
			&project.DescriptionFA,
			&project.DescriptionAR,
			&project.CoverImageURL,
			&project.VideoURL,
			&project.SortOrder,
			&project.CreatedAt,
			&project.UpdatedAt,
		); err != nil {
			return nil, err
		}
		project.Description = project.DescriptionEN
		if project.Description == "" {
			project.Description = project.DescriptionFA
		}
		if project.Description == "" {
			project.Description = project.DescriptionAR
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

type productCardScanner interface {
	Scan(dest ...any) error
}

func scanProductCard(scanner productCardScanner) (domain.Product, error) {
	var product domain.Product
	var mainCategoryID sql.NullInt64
	var categoryID sql.NullInt64
	var categoryTitleEN sql.NullString
	var categoryTitleFA sql.NullString
	var categoryTitleAR sql.NullString
	var categorySlug sql.NullString
	var categoryParentID sql.NullInt64

	if err := scanner.Scan(
		&product.ID,
		&product.TitleEN,
		&product.TitleFA,
		&product.TitleAR,
		&product.Slug,
		&product.ImageURL,
		&mainCategoryID,
		&product.IsPopular,
		&product.ImageCount,
		&categoryID,
		&categoryTitleEN,
		&categoryTitleFA,
		&categoryTitleAR,
		&categorySlug,
		&categoryParentID,
	); err != nil {
		return domain.Product{}, err
	}

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
	return product, nil
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
		SELECT t.id, t.taxonomy, t.term_key, t.label_en, t.label_fa, t.label_ar, COALESCE(t.link_url, ''), t.is_active, t.is_indexable
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
		if err := rows.Scan(&term.ID, &term.Taxonomy, &term.Key, &term.LabelEN, &term.LabelFA, &term.LabelAR, &term.LinkURL, &term.IsActive, &term.IsIndexable); err != nil {
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
		SELECT ptl.product_id, t.id, t.taxonomy, t.term_key, t.label_en, t.label_fa, t.label_ar, COALESCE(t.link_url, ''), t.is_active, t.is_indexable
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
		if err := rows.Scan(&productID, &term.ID, &term.Taxonomy, &term.Key, &term.LabelEN, &term.LabelFA, &term.LabelAR, &term.LinkURL, &term.IsActive, &term.IsIndexable); err != nil {
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
