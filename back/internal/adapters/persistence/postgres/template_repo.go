package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type TemplateRepository struct {
	db *sql.DB
}

func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) List(ctx context.Context) ([]domain.Template, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, image_url, is_active, created_at, COALESCE(updated_at, created_at)
		FROM templates
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.Template
	for rows.Next() {
		var template domain.Template
		if err := rows.Scan(&template.ID, &template.Name, &template.ImageURL, &template.IsActive, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, rows.Err()
}

func (r *TemplateRepository) ListActive(ctx context.Context) ([]domain.Template, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, image_url, is_active, created_at, COALESCE(updated_at, created_at)
		FROM templates
		WHERE is_active = TRUE
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.Template
	for rows.Next() {
		var template domain.Template
		if err := rows.Scan(&template.ID, &template.Name, &template.ImageURL, &template.IsActive, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, rows.Err()
}

func (r *TemplateRepository) GetByID(ctx context.Context, id int64) (domain.Template, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, image_url, is_active, created_at, COALESCE(updated_at, created_at)
		FROM templates
		WHERE id = $1
	`, id)

	var template domain.Template
	if err := row.Scan(&template.ID, &template.Name, &template.ImageURL, &template.IsActive, &template.CreatedAt, &template.UpdatedAt); err != nil {
		return domain.Template{}, err
	}
	return template, nil
}

func (r *TemplateRepository) Create(ctx context.Context, template domain.Template) (domain.Template, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO templates (name, image_url, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`, template.Name, template.ImageURL, template.IsActive)

	if err := row.Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt); err != nil {
		return domain.Template{}, err
	}
	return template, nil
}

func (r *TemplateRepository) Update(ctx context.Context, template domain.Template) (domain.Template, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE templates
		SET name = $1, image_url = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING created_at, COALESCE(updated_at, created_at)
	`, template.Name, template.ImageURL, template.IsActive, template.ID)

	if err := row.Scan(&template.CreatedAt, &template.UpdatedAt); err != nil {
		return domain.Template{}, err
	}
	return template, nil
}

func (r *TemplateRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM templates WHERE id = $1`, id)
	return err
}
