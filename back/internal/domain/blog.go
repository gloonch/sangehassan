package domain

import "time"

type Blog struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Excerpt       string    `json:"excerpt"`
	Content       string    `json:"content"`
	CoverImageURL string    `json:"cover_image_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}
