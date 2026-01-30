package domain

import "time"

type ContentSection struct {
	ID            int64     `json:"id"`
	Page          string    `json:"page"`
	Key           string    `json:"key"`
	TitleEN       string    `json:"title_en"`
	TitleFA       string    `json:"title_fa"`
	TitleAR       string    `json:"title_ar"`
	SubtitleEN    string    `json:"subtitle_en"`
	SubtitleFA    string    `json:"subtitle_fa"`
	SubtitleAR    string    `json:"subtitle_ar"`
	DescriptionEN string    `json:"description_en"`
	DescriptionFA string    `json:"description_fa"`
	DescriptionAR string    `json:"description_ar"`
	CTALabelEN    string    `json:"cta_label_en"`
	CTALabelFA    string    `json:"cta_label_fa"`
	CTALabelAR    string    `json:"cta_label_ar"`
	CTAHref       string    `json:"cta_href"`
	OrderIndex    int       `json:"order_index"`
	IsActive      bool      `json:"is_active"`
	Images        []string  `json:"images,omitempty"`
	ImageCount    int       `json:"image_count,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}
