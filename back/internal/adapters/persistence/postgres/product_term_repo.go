package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type ProductTermRepository struct {
	db *sql.DB
}

func NewProductTermRepository(db *sql.DB) *ProductTermRepository {
	return &ProductTermRepository{db: db}
}

func (r *ProductTermRepository) List(ctx context.Context, taxonomy string) ([]domain.ProductTerm, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if taxonomy == "" {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, taxonomy, term_key, label_en, label_fa, label_ar, COALESCE(link_url, '')
			FROM product_terms
			ORDER BY taxonomy, id
		`)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, taxonomy, term_key, label_en, label_fa, label_ar, COALESCE(link_url, '')
			FROM product_terms
			WHERE taxonomy = $1
			ORDER BY id
		`, taxonomy)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []domain.ProductTerm
	for rows.Next() {
		var term domain.ProductTerm
		if err := rows.Scan(&term.ID, &term.Taxonomy, &term.Key, &term.LabelEN, &term.LabelFA, &term.LabelAR, &term.LinkURL); err != nil {
			return nil, err
		}
		terms = append(terms, term)
	}
	return terms, rows.Err()
}

func (r *ProductTermRepository) Upsert(ctx context.Context, term domain.ProductTerm) (domain.ProductTerm, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO product_terms (taxonomy, term_key, label_en, label_fa, label_ar, link_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (taxonomy, term_key)
		DO UPDATE SET
		  label_en = EXCLUDED.label_en,
		  label_fa = EXCLUDED.label_fa,
		  label_ar = EXCLUDED.label_ar,
		  link_url = EXCLUDED.link_url,
		  updated_at = NOW()
		RETURNING id, taxonomy, term_key, label_en, label_fa, label_ar, COALESCE(link_url, '')
	`, term.Taxonomy, term.Key, term.LabelEN, term.LabelFA, term.LabelAR, nullableString(term.LinkURL))

	var out domain.ProductTerm
	if err := row.Scan(&out.ID, &out.Taxonomy, &out.Key, &out.LabelEN, &out.LabelFA, &out.LabelAR, &out.LinkURL); err != nil {
		return domain.ProductTerm{}, err
	}
	return out, nil
}

func (r *ProductTermRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM product_terms WHERE id = $1`, id)
	return err
}
