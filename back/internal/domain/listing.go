package domain

import "time"

const (
	ListingStatusDraft   = "DRAFT"
	ListingStatusActive  = "ACTIVE"
	ListingStatusPaused  = "PAUSED"
	ListingStatusSold    = "SOLD"
	ListingStatusDeleted = "DELETED"

	ListingFormBlock    = "block"
	ListingFormFinished = "finished"

	PriceUnitPerTon     = "per_ton"
	PriceUnitTotal      = "total"
	PriceUnitNegotiable = "negotiable"
)

// Listing represents a stone advertisement.
type Listing struct {
	ID          int64           `json:"id"`
	CreatedBy   *string         `json:"created_by,omitempty"`
	Author      *UserInfo       `json:"author,omitempty"`
	ProductID   int64           `json:"product_id,omitempty"`
	Product     *ListingProduct `json:"product,omitempty"`
	Title       string          `json:"title,omitempty"`
	StoneType   string          `json:"stone_type,omitempty"`
	Form        string          `json:"form,omitempty"`
	Tonnage     *float64        `json:"tonnage,omitempty"`
	Province    string          `json:"province,omitempty"`
	City        string          `json:"city,omitempty"`
	PriceAmount *float64        `json:"price_amount,omitempty"`
	PriceUnit   string          `json:"price_unit,omitempty"`
	Description string          `json:"description,omitempty"`
	ExtraProps  map[string]any  `json:"extra_props,omitempty"`
	Status      string          `json:"status"`
	Images      []ListingImage  `json:"images,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ListingProduct struct {
	ID       int64  `json:"id"`
	TitleEN  string `json:"title_en"`
	TitleFA  string `json:"title_fa"`
	TitleAR  string `json:"title_ar"`
	Slug     string `json:"slug"`
	ImageURL string `json:"image_url,omitempty"`
}

// ListingLiveFeedItem is the minimal public projection used by the navbar feed.
type ListingLiveFeedItem struct {
	ID          int64
	Title       string
	StoneType   string
	Form        string
	Quantity    *float64
	Product     ListingProduct
	PublishedAt time.Time
}

// ListingImage is intentionally kept for future listing-specific/admin media.
// User-submitted offers currently render the selected product cover instead.
type ListingImage struct {
	ID        int64     `json:"id"`
	ListingID int64     `json:"listing_id"`
	ImageURL  string    `json:"image_url"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

type ListingProductRequest struct {
	ID        int64     `json:"id"`
	UserID    *string   `json:"user_id,omitempty"`
	User      *UserInfo `json:"user,omitempty"`
	Query     string    `json:"query"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
