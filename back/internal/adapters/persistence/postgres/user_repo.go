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
		INSERT INTO users (email, password_hash, full_name, phone, phone_normalized, role, is_active, user_type, status)
		VALUES ($1, $2, $3, $4, $4, COALESCE($5, 'user'), $6, COALESCE(NULLIF($7,''),'CUSTOMER'), CASE WHEN $6 THEN 'ACTIVE' ELSE 'DISABLED' END)
		RETURNING id, email, password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at,
		phone_normalized, first_name, last_name, user_type, status, must_change_password, disabled_at
	`, user.Email, user.PasswordHash, user.FullName, user.Phone, user.Role, user.IsActive, user.UserType)

	return scanUser(row)
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(email,''), password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at,
		phone_normalized, first_name, last_name, user_type, status, must_change_password, disabled_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.User, 0, limit)
	for rows.Next() {
		var (
			u         domain.User
			fullName  sql.NullString
			phone     sql.NullString
			lastLogin sql.NullTime
		)
		var phoneNormalized, firstName, lastName sql.NullString
		var disabledAt sql.NullTime
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &fullName, &phone, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &lastLogin, &phoneNormalized, &firstName, &lastName, &u.UserType, &u.Status, &u.MustChangePassword, &disabledAt); err != nil {
			return nil, err
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
		setOperationalUserFields(&u, phoneNormalized, firstName, lastName, disabledAt)
		items = append(items, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	return total, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(email,''), password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at,
		phone_normalized, first_name, last_name, user_type, status, must_change_password, disabled_at
		FROM users
		WHERE email = $1
	`, email)
	return scanUser(row)
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(email,''), password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at,
		phone_normalized, first_name, last_name, user_type, status, must_change_password, disabled_at
		FROM users
		WHERE phone_normalized = $1 OR phone = $1
	`, phone)
	return scanUser(row)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(email,''), password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at,
		phone_normalized, first_name, last_name, user_type, status, must_change_password, disabled_at
		FROM users
		WHERE id = $1
	`, id)
	return scanUser(row)
}

func (r *UserRepository) UpdateProfile(ctx context.Context, user domain.User) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET email = $2, full_name = $3, phone = $4, phone_normalized = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, COALESCE(email,''), password_hash, full_name, phone, role, is_active, created_at, updated_at, last_login_at,
		phone_normalized, first_name, last_name, user_type, status, must_change_password, disabled_at
	`, user.ID, user.Email, user.FullName, user.Phone)
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
		u                                    domain.User
		fullName                             sql.NullString
		phone                                sql.NullString
		lastLogin                            sql.NullTime
		phoneNormalized, firstName, lastName sql.NullString
		disabledAt                           sql.NullTime
	)
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &fullName, &phone, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &lastLogin, &phoneNormalized, &firstName, &lastName, &u.UserType, &u.Status, &u.MustChangePassword, &disabledAt); err != nil {
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
	setOperationalUserFields(&u, phoneNormalized, firstName, lastName, disabledAt)
	return u, nil
}

func setOperationalUserFields(u *domain.User, phone, first, last sql.NullString, disabled sql.NullTime) {
	if phone.Valid {
		u.PhoneNormalized = &phone.String
	}
	if first.Valid {
		u.FirstName = &first.String
	}
	if last.Valid {
		u.LastName = &last.String
	}
	if disabled.Valid {
		t := disabled.Time
		u.DisabledAt = &t
	}
}

func (r *UserRepository) AccessAllowed(ctx context.Context, id string, issuedAt time.Time) (bool, error) {
	var allowed bool
	err := r.db.QueryRowContext(ctx, `SELECT is_active AND status='ACTIVE' AND (auth_invalid_before IS NULL OR auth_invalid_before<=$2) FROM users WHERE id=$1`, id, issuedAt).Scan(&allowed)
	return allowed, err
}

func (r *UserRepository) Authorization(ctx context.Context, id string) ([]string, []string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT r.code, COALESCE(p.code,'') FROM user_roles ur JOIN roles r ON r.id=ur.role_id AND r.is_active LEFT JOIN role_permissions rp ON rp.role_id=r.id LEFT JOIN permissions p ON p.id=rp.permission_id AND p.is_active WHERE ur.user_id=$1`, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	roleSet, permissionSet := map[string]bool{}, map[string]bool{}
	for rows.Next() {
		var role, permission string
		if err := rows.Scan(&role, &permission); err != nil {
			return nil, nil, err
		}
		roleSet[role] = true
		if permission != "" {
			permissionSet[permission] = true
		}
	}
	if roleSet["SUPER_ADMIN"] {
		all, err := r.db.QueryContext(ctx, `SELECT code FROM permissions WHERE is_active`)
		if err != nil {
			return nil, nil, err
		}
		defer all.Close()
		for all.Next() {
			var p string
			if err := all.Scan(&p); err != nil {
				return nil, nil, err
			}
			permissionSet[p] = true
		}
	}
	roles, permissions := make([]string, 0, len(roleSet)), make([]string, 0, len(permissionSet))
	for v := range roleSet {
		roles = append(roles, v)
	}
	for v := range permissionSet {
		permissions = append(permissions, v)
	}
	return roles, permissions, rows.Err()
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id, hash string, mustChange bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash=$2,must_change_password=$3,updated_at=NOW() WHERE id=$1`, id, hash, mustChange)
	return err
}
