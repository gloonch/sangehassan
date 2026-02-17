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
	DescriptionHTMLEn    string              `json:"description_html_en,omitempty"`
	DescriptionHTMLFa    string              `json:"description_html_fa,omitempty"`
	DescriptionHTMLAr    string              `json:"description_html_ar,omitempty"`
	ShortDescriptionHTML string              `json:"short_description_html"`
	ShortDescriptionHTMLEn string            `json:"short_description_html_en,omitempty"`
	ShortDescriptionHTMLFa string            `json:"short_description_html_fa,omitempty"`
	ShortDescriptionHTMLAr string            `json:"short_description_html_ar,omitempty"`
	Price                float64             `json:"price"`
	PriceHTML            string              `json:"price_html"`
	ImageURL             string              `json:"image_url"`
	Images               []string            `json:"images,omitempty"`
	ImageCount           int                 `json:"image_count,omitempty"`
	MainCategoryID       *int64              `json:"category_id"`
	IsPopular            bool                `json:"is_popular"`
	TermIDs              []int64             `json:"term_ids,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at,omitempty"`
	Category             *Category           `json:"category,omitempty"`
	Categories           []Category          `json:"categories,omitempty"`
	Attributes           map[string][]string `json:"attributes,omitempty"`
	Terms                []ProductTerm       `json:"terms,omitempty"`
}
