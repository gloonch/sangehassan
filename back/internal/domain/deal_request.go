package domain

import "time"

type DealRequestType string

const (
	DealRequestTypeInspection DealRequestType = "INSPECTION"
	DealRequestTypePurchase   DealRequestType = "PURCHASE"
	DealRequestTypeBoth       DealRequestType = "BOTH"
)

type DealRequestStatus string

const (
	DealStatusPendingReview    DealRequestStatus = "PENDING_REVIEW"
	DealStatusApproved         DealRequestStatus = "APPROVED"
	DealStatusRejected         DealRequestStatus = "REJECTED"
	DealStatusMeetingScheduled DealRequestStatus = "MEETING_SCHEDULED"
	DealStatusInProgress       DealRequestStatus = "IN_PROGRESS"
	DealStatusCompleted        DealRequestStatus = "COMPLETED"
	DealStatusCanceled         DealRequestStatus = "CANCELED"
)

type DealRequest struct {
	ID            int64               `json:"id"`
	ListingID     int64               `json:"listing_id"`
	BuyerID       *string             `json:"buyer_id,omitempty"`
	SellerID      *string             `json:"seller_id,omitempty"`
	RequestType   DealRequestType     `json:"request_type"`
	BuyerNote     string              `json:"buyer_note,omitempty"`
	Status        DealRequestStatus   `json:"status"`
	MeetingAt     *time.Time          `json:"meeting_at,omitempty"`
	AdminNote     string              `json:"admin_note,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	StatusHistory []DealStatusHistory `json:"status_history,omitempty"`
}

type DealStatusHistory struct {
	ID         int64              `json:"id"`
	RequestID  int64              `json:"request_id"`
	FromStatus *DealRequestStatus `json:"from_status,omitempty"`
	ToStatus   DealRequestStatus  `json:"to_status"`
	CreatedBy  *string            `json:"created_by,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
}
