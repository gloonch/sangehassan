package domain

import "time"

type Category struct {
	ID        int64     `json:"id"`
	TitleEN   string    `json:"title_en"`
	TitleFA   string    `json:"title_fa"`
	TitleAR   string    `json:"title_ar"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
