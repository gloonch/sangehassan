package domain

import (
	"encoding/json"
	"time"
)

type BlogTranslation struct {
	ID                int64           `json:"id,omitempty"`
	BlogID            int64           `json:"blog_id,omitempty"`
	Locale            string          `json:"locale"`
	Title             string          `json:"title"`
	Slug              string          `json:"slug"`
	Excerpt           string          `json:"excerpt"`
	ContentJSON       json.RawMessage `json:"content_json,omitempty"`
	ContentHTML       string          `json:"content_html"`
	SEOTitle          string          `json:"seo_title"`
	SEODescription    string          `json:"seo_description"`
	CanonicalURL      string          `json:"canonical_url"`
	Robots            string          `json:"robots"`
	TranslationStatus string          `json:"translation_status"`
	FeaturedImageAlt  string          `json:"featured_image_alt"`
	OGImageAlt        string          `json:"og_image_alt"`
	CreatedAt         time.Time       `json:"created_at,omitempty"`
	UpdatedAt         time.Time       `json:"updated_at,omitempty"`
}

type Blog struct {
	ID            int64             `json:"id"`
	Status        string            `json:"status"`
	AuthorName    string            `json:"author_name"`
	CoverImageURL string            `json:"cover_image_url"`
	OGImageURL    string            `json:"og_image_url"`
	CategorySlug  string            `json:"category_slug"`
	Tags          []string          `json:"tags"`
	IsFeatured    bool              `json:"is_featured"`
	ScheduledAt   *time.Time        `json:"scheduled_at,omitempty"`
	PublishedAt   *time.Time        `json:"published_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Translations  []BlogTranslation `json:"translations,omitempty"`

	Locale             string          `json:"locale,omitempty"`
	Title              string          `json:"title,omitempty"`
	Slug               string          `json:"slug,omitempty"`
	Excerpt            string          `json:"excerpt,omitempty"`
	ContentJSON        json.RawMessage `json:"content_json,omitempty"`
	ContentHTML        string          `json:"content_html,omitempty"`
	SEOTitle           string          `json:"seo_title,omitempty"`
	SEODescription     string          `json:"seo_description,omitempty"`
	CanonicalURL       string          `json:"canonical_url,omitempty"`
	Robots             string          `json:"robots,omitempty"`
	TranslationStatus  string          `json:"translation_status,omitempty"`
	FeaturedImageAlt   string          `json:"featured_image_alt,omitempty"`
	OGImageAlt         string          `json:"og_image_alt,omitempty"`
	ReadingTimeMinutes int             `json:"reading_time_minutes,omitempty"`
	RedirectedFrom     string          `json:"redirected_from,omitempty"`
}
