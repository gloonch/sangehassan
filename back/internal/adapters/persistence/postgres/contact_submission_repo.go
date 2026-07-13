package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

const defaultContactSubmissionLimit = 50

type ContactSubmissionRepository struct {
	db *sql.DB
}

func NewContactSubmissionRepository(db *sql.DB) *ContactSubmissionRepository {
	return &ContactSubmissionRepository{db: db}
}

func (r *ContactSubmissionRepository) Create(ctx context.Context, submission domain.ContactSubmission) (domain.ContactSubmission, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO contact_submissions (
			full_name, email, country_iso, country_code, phone_number, phone_e164, message, source
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`,
		submission.FullName,
		submission.Email,
		submission.CountryISO,
		submission.CountryCode,
		submission.PhoneNumber,
		submission.PhoneE164,
		submission.Message,
		submission.Source,
	)
	if err := row.Scan(&submission.ID, &submission.CreatedAt); err != nil {
		return domain.ContactSubmission{}, err
	}
	return submission, nil
}

func (r *ContactSubmissionRepository) List(ctx context.Context, limit, offset int) ([]domain.ContactSubmission, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultContactSubmissionLimit
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, full_name, COALESCE(email, ''), country_iso, country_code, phone_number,
		       phone_e164, message, source, created_at
		FROM contact_submissions
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissions := make([]domain.ContactSubmission, 0, limit)
	for rows.Next() {
		submission, err := scanContactSubmission(rows)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return submissions, nil
}

func (r *ContactSubmissionRepository) GetByID(ctx context.Context, id int64) (domain.ContactSubmission, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, full_name, COALESCE(email, ''), country_iso, country_code, phone_number,
		       phone_e164, message, source, created_at
		FROM contact_submissions
		WHERE id = $1
	`, id)
	return scanContactSubmission(row)
}

type contactSubmissionScanner interface {
	Scan(dest ...any) error
}

func scanContactSubmission(scanner contactSubmissionScanner) (domain.ContactSubmission, error) {
	var submission domain.ContactSubmission
	err := scanner.Scan(
		&submission.ID,
		&submission.FullName,
		&submission.Email,
		&submission.CountryISO,
		&submission.CountryCode,
		&submission.PhoneNumber,
		&submission.PhoneE164,
		&submission.Message,
		&submission.Source,
		&submission.CreatedAt,
	)
	if err != nil {
		return domain.ContactSubmission{}, err
	}
	return submission, nil
}
