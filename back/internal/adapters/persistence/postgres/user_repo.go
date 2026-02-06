package postgres

import (
	"context"
	"database/sql"
	"time"

	"sangehassan/back/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, full_name, phone, role, is_active)
		VALUES ($1, $2, $3, $4, COALESCE($5, 'user'), $6)
		RETURNING id, email, password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at
	`, user.Email, user.PasswordHash, user.FullName, user.Phone, user.Role, user.IsActive)

	return scanUser(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`, email)
	return scanUser(row)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`, id)
	return scanUser(row)
}

func (r *UserRepository) UpdateProfile(ctx context.Context, user domain.User) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET full_name = $2, phone = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at
	`, user.ID, user.FullName, user.Phone)
	return scanUser(row)
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET last_login_at = $2, updated_at = NOW() WHERE id = $1
	`, id, at)
	return err
}

func scanUser(row *sql.Row) (domain.User, error) {
	var (
		u         domain.User
		fullName  sql.NullString
		phone     sql.NullString
		lastLogin sql.NullTime
	)
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &fullName, &phone, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &lastLogin); err != nil {
		return domain.User{}, err
	}
	if fullName.Valid {
		u.FullName = &fullName.String
	}
	if phone.Valid {
		u.Phone = &phone.String
	}
	if lastLogin.Valid {
		t := lastLogin.Time
		u.LastLoginAt = &t
	}
	return u, nil
}
