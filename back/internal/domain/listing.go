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
	ID          int64          `json:"id"`
	CreatedBy   *string        `json:"created_by,omitempty"`
	Author      *UserInfo      `json:"author,omitempty"`
	Title       string         `json:"title,omitempty"`
	StoneType   string         `json:"stone_type,omitempty"`
	Form        string         `json:"form,omitempty"`
	Tonnage     *float64       `json:"tonnage,omitempty"`
	Province    string         `json:"province,omitempty"`
	City        string         `json:"city,omitempty"`
	PriceAmount *float64       `json:"price_amount,omitempty"`
	PriceUnit   string         `json:"price_unit,omitempty"`
	Description string         `json:"description,omitempty"`
	ExtraProps  map[string]any `json:"extra_props,omitempty"`
	Status      string         `json:"status"`
	Images      []ListingImage `json:"images,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ListingImage struct {
	ID        int64     `json:"id"`
	ListingID int64     `json:"listing_id"`
	ImageURL  string    `json:"image_url"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}
