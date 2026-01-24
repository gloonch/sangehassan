package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type BlogRepository struct {
	db *sql.DB
}

func NewBlogRepository(db *sql.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

func (r *BlogRepository) List(ctx context.Context) ([]domain.Blog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, COALESCE(excerpt, ''), COALESCE(content, ''), COALESCE(cover_image_url, ''), created_at, COALESCE(updated_at, created_at)
		FROM blogs
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []domain.Blog
	for rows.Next() {
		var blog domain.Blog
		if err := rows.Scan(&blog.ID, &blog.Title, &blog.Excerpt, &blog.Content, &blog.CoverImageURL, &blog.CreatedAt, &blog.UpdatedAt); err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}
	return blogs, rows.Err()
}

func (r *BlogRepository) GetByID(ctx context.Context, id int64) (domain.Blog, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, COALESCE(excerpt, ''), COALESCE(content, ''), COALESCE(cover_image_url, ''), created_at, COALESCE(updated_at, created_at)
		FROM blogs
		WHERE id = $1
	`, id)
	var blog domain.Blog
	if err := row.Scan(&blog.ID, &blog.Title, &blog.Excerpt, &blog.Content, &blog.CoverImageURL, &blog.CreatedAt, &blog.UpdatedAt); err != nil {
		return domain.Blog{}, err
	}
	return blog, nil
}

func (r *BlogRepository) Create(ctx context.Context, blog domain.Blog) (domain.Blog, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO blogs (title, excerpt, content, cover_image_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`, blog.Title, nullableString(blog.Excerpt), nullableString(blog.Content), nullableString(blog.CoverImageURL))

	if err := row.Scan(&blog.ID, &blog.CreatedAt, &blog.UpdatedAt); err != nil {
		return domain.Blog{}, err
	}
	return blog, nil
}

func (r *BlogRepository) Update(ctx context.Context, blog domain.Blog) (domain.Blog, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE blogs
		SET title = $1, excerpt = $2, content = $3, cover_image_url = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING created_at, COALESCE(updated_at, created_at)
	`, blog.Title, nullableString(blog.Excerpt), nullableString(blog.Content), nullableString(blog.CoverImageURL), blog.ID)

	if err := row.Scan(&blog.CreatedAt, &blog.UpdatedAt); err != nil {
		return domain.Blog{}, err
	}
	return blog, nil
}

func (r *BlogRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM blogs WHERE id = $1`, id)
	return err
}
