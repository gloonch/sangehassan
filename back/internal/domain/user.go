package domain

import "time"

type User struct {
	ID                 string
	Email              string
	PasswordHash       string
	FullName           *string
	Phone              *string
	Role               string
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	LastLoginAt        *time.Time
	PhoneNormalized    *string
	FirstName          *string
	LastName           *string
	UserType           string
	Status             string
	MustChangePassword bool
	DisabledAt         *time.Time
}

// UserInfo is safe to expose to clients.
type UserInfo struct {
	ID                 string     `json:"id"`
	Email              string     `json:"email"`
	FullName           *string    `json:"full_name,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	Role               string     `json:"role"`
	CreatedAt          time.Time  `json:"created_at"`
	FirstName          *string    `json:"first_name,omitempty"`
	LastName           *string    `json:"last_name,omitempty"`
	UserType           string     `json:"user_type"`
	Status             string     `json:"status"`
	MustChangePassword bool       `json:"must_change_password"`
	Roles              []string   `json:"roles"`
	Permissions        []string   `json:"permissions"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
}

func (u User) SafeInfo() UserInfo {
	phone := u.PhoneNormalized
	if phone == nil {
		phone = u.Phone
	}
	return UserInfo{
		ID:                 u.ID,
		Email:              u.Email,
		FullName:           u.FullName,
		Phone:              phone,
		Role:               u.Role,
		CreatedAt:          u.CreatedAt,
		FirstName:          u.FirstName,
		LastName:           u.LastName,
		UserType:           u.UserType,
		Status:             u.Status,
		MustChangePassword: u.MustChangePassword,
		LastLoginAt:        u.LastLoginAt,
	}
}
