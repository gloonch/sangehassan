package domain

import "time"

type Block struct {
	ID          int64     `json:"id"`
	TitleEN     string    `json:"title_en"`
	TitleFA     string    `json:"title_fa"`
	TitleAR     string    `json:"title_ar"`
	Slug        string    `json:"slug"`
	StoneType   string    `json:"stone_type"`
	Quarry      string    `json:"quarry"`
	Dimensions  string    `json:"dimensions"`
	WeightTon   float64   `json:"weight_ton"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Images      []string  `json:"images,omitempty"`
	ImageCount  int       `json:"image_count,omitempty"`
	IsFeatured  bool      `json:"is_featured"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
