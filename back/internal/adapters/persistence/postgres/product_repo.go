package postgres

import (
	"context"
	"database/sql"

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
		SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.description, p.price, p.image_url, p.category_id, p.is_popular,
			p.created_at, COALESCE(p.updated_at, p.created_at),
			c.id, c.title_en, c.title_fa, c.title_ar, c.slug
		FROM products p
		JOIN categories c ON c.id = p.category_id
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
		var price sql.NullFloat64
		var imageURL sql.NullString
		var category domain.Category
		if err := rows.Scan(
			&product.ID,
			&product.TitleEN,
			&product.TitleFA,
			&product.TitleAR,
			&description,
			&price,
			&imageURL,
			&product.CategoryID,
			&product.IsPopular,
			&product.CreatedAt,
			&product.UpdatedAt,
			&category.ID,
			&category.TitleEN,
			&category.TitleFA,
			&category.TitleAR,
			&category.Slug,
		); err != nil {
			return nil, err
		}
		product.Description = description.String
		if price.Valid {
			product.Price = price.Float64
		}
		product.ImageURL = imageURL.String
		product.Category = &category
		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *ProductRepository) ListPopular(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.description, p.price, p.image_url, p.category_id, p.is_popular,
			p.created_at, COALESCE(p.updated_at, p.created_at),
			c.id, c.title_en, c.title_fa, c.title_ar, c.slug
		FROM products p
		JOIN categories c ON c.id = p.category_id
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
		var price sql.NullFloat64
		var imageURL sql.NullString
		var category domain.Category
		if err := rows.Scan(
			&product.ID,
			&product.TitleEN,
			&product.TitleFA,
			&product.TitleAR,
			&description,
			&price,
			&imageURL,
			&product.CategoryID,
			&product.IsPopular,
			&product.CreatedAt,
			&product.UpdatedAt,
			&category.ID,
			&category.TitleEN,
			&category.TitleFA,
			&category.TitleAR,
			&category.Slug,
		); err != nil {
			return nil, err
		}
		product.Description = description.String
		if price.Valid {
			product.Price = price.Float64
		}
		product.ImageURL = imageURL.String
		product.Category = &category
		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (domain.Product, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.description, p.price, p.image_url, p.category_id, p.is_popular,
			p.created_at, COALESCE(p.updated_at, p.created_at),
			c.id, c.title_en, c.title_fa, c.title_ar, c.slug
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1
	`, id)

	var product domain.Product
	var description sql.NullString
	var price sql.NullFloat64
	var imageURL sql.NullString
	var category domain.Category

	if err := row.Scan(
		&product.ID,
		&product.TitleEN,
		&product.TitleFA,
		&product.TitleAR,
		&description,
		&price,
		&imageURL,
		&product.CategoryID,
		&product.IsPopular,
		&product.CreatedAt,
		&product.UpdatedAt,
		&category.ID,
		&category.TitleEN,
		&category.TitleFA,
		&category.TitleAR,
		&category.Slug,
	); err != nil {
		return domain.Product{}, err
	}

	product.Description = description.String
	if price.Valid {
		product.Price = price.Float64
	}
	product.ImageURL = imageURL.String
	product.Category = &category
	return product, nil
}

func (r *ProductRepository) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO products (title_en, title_fa, title_ar, description, price, image_url, category_id, is_popular)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`, product.TitleEN, product.TitleFA, product.TitleAR, nullableString(product.Description), nullableFloat(product.Price), nullableString(product.ImageURL), product.CategoryID, product.IsPopular)

	if err := row.Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

func (r *ProductRepository) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE products
		SET title_en = $1, title_fa = $2, title_ar = $3, description = $4, price = $5, image_url = $6, category_id = $7, is_popular = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING created_at, COALESCE(updated_at, created_at)
	`, product.TitleEN, product.TitleFA, product.TitleAR, nullableString(product.Description), nullableFloat(product.Price), nullableString(product.ImageURL), product.CategoryID, product.IsPopular, product.ID)

	if err := row.Scan(&product.CreatedAt, &product.UpdatedAt); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
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
