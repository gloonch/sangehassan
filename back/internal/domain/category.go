package domain

import "time"

type Category struct {
	ID               int64     `json:"id"`
	TitleEN          string    `json:"title_en"`
	TitleFA          string    `json:"title_fa"`
	TitleAR          string    `json:"title_ar"`
	Slug             string    `json:"slug"`
	ParentID         *int64    `json:"parent_id,omitempty"`
	ImageURL         string    `json:"image_url,omitempty"`
	IntroEN          string    `json:"intro_en,omitempty"`
	IntroFA          string    `json:"intro_fa,omitempty"`
	IntroAR          string    `json:"intro_ar,omitempty"`
	SEOTitleEN       string    `json:"seo_title_en,omitempty"`
	SEOTitleFA       string    `json:"seo_title_fa,omitempty"`
	SEOTitleAR       string    `json:"seo_title_ar,omitempty"`
	SEODescriptionEN string    `json:"seo_description_en,omitempty"`
	SEODescriptionFA string    `json:"seo_description_fa,omitempty"`
	SEODescriptionAR string    `json:"seo_description_ar,omitempty"`
	IsActive         bool      `json:"is_active"`
	IsIndexable      bool      `json:"is_indexable"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}
