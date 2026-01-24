package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) GetByUsername(ctx context.Context, username string) (domain.AdminUser, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash
		FROM admin_users
		WHERE username = $1
	`, username)

	var user domain.AdminUser
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash); err != nil {
		return domain.AdminUser{}, err
	}
	return user, nil
}
