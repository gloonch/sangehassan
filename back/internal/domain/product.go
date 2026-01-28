package domain

import "time"

type Product struct {
	ID          int64     `json:"id"`
	TitleEN     string    `json:"title_en"`
	TitleFA     string    `json:"title_fa"`
	TitleAR     string    `json:"title_ar"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url"`
	CategoryID  int64     `json:"category_id"`
	IsPopular   bool      `json:"is_popular"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Category    *Category `json:"category,omitempty"`
}
