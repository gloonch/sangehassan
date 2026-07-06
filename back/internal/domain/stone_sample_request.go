package domain

import "time"

const (
	StoneSampleSamplesPerBox = 4
	StoneSampleMaxBoxes      = 4
	StoneSamplePricePerBox   = int64(4000000)
)

type StoneSampleRequestStatus string

const (
	StoneSampleStatusPending   StoneSampleRequestStatus = "PENDING"
	StoneSampleStatusConfirmed StoneSampleRequestStatus = "CONFIRMED"
	StoneSampleStatusRejected  StoneSampleRequestStatus = "REJECTED"
	StoneSampleStatusShipped   StoneSampleRequestStatus = "SHIPPED"
	StoneSampleStatusDelivered StoneSampleRequestStatus = "DELIVERED"
	StoneSampleStatusCanceled  StoneSampleRequestStatus = "CANCELED"
)

type StoneSampleShippingMethod string

const (
	StoneSampleShippingCourier              StoneSampleShippingMethod = "COURIER"
	StoneSampleShippingFreight              StoneSampleShippingMethod = "FREIGHT"
	StoneSampleShippingOperatorCoordination StoneSampleShippingMethod = "OPERATOR_COORDINATION"
)

type UserAddress struct {
	ID          int64     `json:"id"`
	UserID      string    `json:"user_id,omitempty"`
	Label       string    `json:"label,omitempty"`
	AddressText string    `json:"address_text"`
	City        string    `json:"city,omitempty"`
	Province    string    `json:"province,omitempty"`
	PostalCode  string    `json:"postal_code,omitempty"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserContactNumber struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	Label     string    `json:"label,omitempty"`
	Phone     string    `json:"phone"`
	IsDefault bool      `json:"is_default"`
	Source    string    `json:"source,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StoneSampleProduct struct {
	ID              int64  `json:"id"`
	TitleEN         string `json:"title_en"`
	TitleFA         string `json:"title_fa"`
	TitleAR         string `json:"title_ar"`
	Slug            string `json:"slug"`
	ImageURL        string `json:"image_url"`
	CategoryID      *int64 `json:"category_id,omitempty"`
	CategorySlug    string `json:"category_slug,omitempty"`
	CategoryTitleEN string `json:"category_title_en,omitempty"`
	CategoryTitleFA string `json:"category_title_fa,omitempty"`
	CategoryTitleAR string `json:"category_title_ar,omitempty"`
}

type StoneSampleShippingMethodOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type StoneSampleRequestOptions struct {
	Addresses        []UserAddress                     `json:"addresses"`
	Phones           []UserContactNumber               `json:"phones"`
	ShippingMethods  []StoneSampleShippingMethodOption `json:"shipping_methods"`
	PricePerBoxToman int64                             `json:"price_per_box_toman"`
	SamplesPerBox    int                               `json:"samples_per_box"`
	MaxBoxes         int                               `json:"max_boxes"`
}

type StoneSampleRequest struct {
	ID               int64                      `json:"id"`
	UserID           string                     `json:"user_id,omitempty"`
	Status           StoneSampleRequestStatus   `json:"status"`
	BoxCount         int                        `json:"box_count"`
	StoneCount       int                        `json:"stone_count"`
	PricePerBoxToman int64                      `json:"price_per_box_toman"`
	TotalPriceToman  int64                      `json:"total_price_toman"`
	ShippingMethod   string                     `json:"shipping_method"`
	AddressID        *int64                     `json:"address_id,omitempty"`
	PhoneID          *int64                     `json:"phone_id,omitempty"`
	AddressSnapshot  string                     `json:"address_snapshot"`
	PhoneSnapshot    string                     `json:"phone_snapshot"`
	AdminNote        string                     `json:"admin_note,omitempty"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
	Items            []StoneSampleRequestItem   `json:"items,omitempty"`
	StatusHistory    []StoneSampleStatusHistory `json:"status_history,omitempty"`
	User             *UserInfo                  `json:"user,omitempty"`
}

type StoneSampleRequestItem struct {
	ID              int64  `json:"id"`
	RequestID       int64  `json:"request_id,omitempty"`
	ProductID       *int64 `json:"product_id,omitempty"`
	BoxIndex        int    `json:"box_index"`
	SlotIndex       int    `json:"slot_index"`
	ProductTitleEN  string `json:"product_title_en"`
	ProductTitleFA  string `json:"product_title_fa"`
	ProductTitleAR  string `json:"product_title_ar"`
	ProductSlug     string `json:"product_slug"`
	ProductImageURL string `json:"product_image_url,omitempty"`
	CategoryTitleEN string `json:"category_title_en,omitempty"`
	CategoryTitleFA string `json:"category_title_fa,omitempty"`
	CategoryTitleAR string `json:"category_title_ar,omitempty"`
}

type StoneSampleStatusHistory struct {
	ID         int64                     `json:"id"`
	RequestID  int64                     `json:"request_id,omitempty"`
	FromStatus *StoneSampleRequestStatus `json:"from_status,omitempty"`
	ToStatus   StoneSampleRequestStatus  `json:"to_status"`
	CreatedBy  *string                   `json:"created_by,omitempty"`
	AdminNote  string                    `json:"admin_note,omitempty"`
	CreatedAt  time.Time                 `json:"created_at"`
}

type StoneSampleCatalogFilter struct {
	CategoryID *int64
	Query      string
	Limit      int
	Offset     int
}

type CreateStoneSampleRequestInput struct {
	UserID         string
	Boxes          [][]int64
	AddressID      *int64
	AddressText    string
	PhoneID        *int64
	Phone          string
	ShippingMethod string
}
