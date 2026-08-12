package usecase

import "time"

type SupplierPayload struct {
	Name           string `json:"name"`
	SupplierType   string `json:"supplier_type"`
	Phone          string `json:"phone"`
	SecondaryPhone string `json:"secondary_phone"`
	ContactName    string `json:"contact_name"`
	Address        string `json:"address"`
	City           string `json:"city"`
	Province       string `json:"province"`
	CountryCode    string `json:"country_code"`
	Notes          string `json:"notes"`
}

type Supplier struct {
	ID             string         `json:"id"`
	SupplierCode   string         `json:"supplier_code"`
	Name           string         `json:"name"`
	SupplierType   string         `json:"supplier_type"`
	Phone          string         `json:"phone"`
	SecondaryPhone string         `json:"secondary_phone"`
	ContactName    string         `json:"contact_name"`
	Address        string         `json:"address"`
	City           string         `json:"city"`
	Province       string         `json:"province"`
	CountryCode    string         `json:"country_code"`
	Notes          string         `json:"notes,omitempty"`
	IsActive       bool           `json:"is_active"`
	Statistics     map[string]any `json:"statistics,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type PurchasePayload struct {
	BatchID         *string    `json:"batch_id"`
	SupplierID      string     `json:"supplier_id"`
	AssignedUserID  *string    `json:"assigned_user_id"`
	AssignedRoleID  *int64     `json:"assigned_role_id"`
	StoneName       string     `json:"stone_name"`
	Description     string     `json:"description"`
	Quantity        string     `json:"quantity"`
	QuantityUnit    string     `json:"quantity_unit"`
	UnitPrice       string     `json:"unit_price"`
	CurrencyCode    string     `json:"currency_code"`
	ExpectedAt      *time.Time `json:"expected_at"`
	Notes           string     `json:"notes"`
	CreateCostEntry bool       `json:"create_cost_entry"`
}

type PurchaseReceiptPayload struct {
	Quantity         string          `json:"quantity"`
	ReceivedAt       *time.Time      `json:"received_at"`
	Notes            string          `json:"notes"`
	InventoryReceipt *ReceiptPayload `json:"inventory_receipt"`
}

type Purchase struct {
	ID               string            `json:"id"`
	PurchaseNumber   string            `json:"purchase_number"`
	OrderID          string            `json:"order_id"`
	OrderNumber      string            `json:"order_number"`
	BatchID          *string           `json:"batch_id,omitempty"`
	SupplierID       string            `json:"supplier_id"`
	SupplierName     string            `json:"supplier_name"`
	AssignedUserID   *string           `json:"assigned_user_id,omitempty"`
	AssignedRoleID   *int64            `json:"assigned_role_id,omitempty"`
	StoneName        string            `json:"stone_name"`
	Description      string            `json:"description"`
	Quantity         string            `json:"quantity"`
	ReceivedQuantity string            `json:"received_quantity"`
	QuantityUnit     string            `json:"quantity_unit"`
	UnitPrice        string            `json:"unit_price"`
	TotalAmount      string            `json:"total_amount"`
	CurrencyCode     string            `json:"currency_code"`
	Status           string            `json:"status"`
	ExpectedAt       *time.Time        `json:"expected_at,omitempty"`
	ReceivedAt       *time.Time        `json:"received_at,omitempty"`
	Notes            string            `json:"notes"`
	CostEntryID      *string           `json:"cost_entry_id,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	Receipts         []PurchaseReceipt `json:"receipts,omitempty"`
}

type PurchaseReceipt struct {
	ID                  string    `json:"id"`
	Quantity            string    `json:"quantity"`
	QuantityUnit        string    `json:"quantity_unit"`
	InventoryLotID      *string   `json:"inventory_lot_id,omitempty"`
	InventoryMovementID *string   `json:"inventory_movement_id,omitempty"`
	Notes               string    `json:"notes"`
	ReceivedAt          time.Time `json:"received_at"`
}
