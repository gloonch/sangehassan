package domain

import "time"

type Product struct {
	ID                   int64               `json:"id"`
	TitleEN              string              `json:"title_en"`
	TitleFA              string              `json:"title_fa"`
	TitleAR              string              `json:"title_ar"`
	Slug                 string              `json:"slug"`
	Description          string              `json:"description"` // plain alias for HTML field
	DescriptionHTML      string              `json:"description_html"`
	ShortDescriptionHTML string              `json:"short_description_html"`
	Price                float64             `json:"price"`
	PriceHTML            string              `json:"price_html"`
	ImageURL             string              `json:"image_url"`
	Images               []string            `json:"images,omitempty"`
	ImageCount           int                 `json:"image_count,omitempty"`
	MainCategoryID       *int64              `json:"category_id"`
	IsPopular            bool                `json:"is_popular"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at,omitempty"`
	Category             *Category           `json:"category,omitempty"`
	Categories           []Category          `json:"categories,omitempty"`
	Attributes           map[string][]string `json:"attributes,omitempty"`
}
