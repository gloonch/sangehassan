package domain

import "time"

type Project struct {
	ID            int64     `json:"id"`
	Description   string    `json:"description"`
	DescriptionEN string    `json:"description_en"`
	DescriptionFA string    `json:"description_fa"`
	DescriptionAR string    `json:"description_ar"`
	CoverImageURL string    `json:"cover_image_url"`
	VideoURL      string    `json:"video_url,omitempty"`
	GalleryImages []string  `json:"gallery_images,omitempty"`
	GalleryCount  int       `json:"gallery_count,omitempty"`
	ProductIDs    []int64   `json:"product_ids,omitempty"`
	UsedProducts  []Product `json:"used_products,omitempty"`
	SortOrder     int       `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type ProjectCard struct {
	ID            int64  `json:"id"`
	CoverImageURL string `json:"cover_image_url"`
	SortOrder     int    `json:"sort_order"`
}
