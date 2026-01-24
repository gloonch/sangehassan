package domain

type AdminUser struct {
	ID           int64
	Username     string
	PasswordHash string
}
