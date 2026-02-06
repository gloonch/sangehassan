package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token domain.RefreshToken) (domain.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at
	`, token.UserID, token.TokenHash, token.ExpiresAt, token.RevokedAt)
	return scanRefreshToken(row)
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (domain.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, hash)
	return scanRefreshToken(row)
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *RefreshTokenRepository) DeleteByUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM refresh_tokens WHERE user_id = $1
	`, userID)
	return err
}

func scanRefreshToken(row *sql.Row) (domain.RefreshToken, error) {
	var t domain.RefreshToken
	if err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt); err != nil {
		return domain.RefreshToken{}, err
	}
	return t, nil
}
