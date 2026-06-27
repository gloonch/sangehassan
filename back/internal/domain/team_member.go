package domain

import "time"

type TeamMember struct {
	ID          int64     `json:"id"`
	NameEN      string    `json:"name_en"`
	NameFA      string    `json:"name_fa"`
	NameAR      string    `json:"name_ar"`
	RoleEN      string    `json:"role_en"`
	RoleFA      string    `json:"role_fa"`
	RoleAR      string    `json:"role_ar"`
	BioEN       string    `json:"bio_en"`
	BioFA       string    `json:"bio_fa"`
	BioAR       string    `json:"bio_ar"`
	PhotoURL    string    `json:"photo_url"`
	LinkedInURL string    `json:"linkedin_url"`
	OrderIndex  int       `json:"order_index"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
