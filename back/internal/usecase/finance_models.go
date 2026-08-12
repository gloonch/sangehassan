package usecase

import "time"

const (
	ErrPaymentOverAllocation      = "PAYMENT_OVER_ALLOCATION"
	ErrCurrencyMismatch           = "CURRENCY_MISMATCH"
	ErrPaymentRequired            = "PAYMENT_REQUIRED"
	ErrDocumentRequired           = "DOCUMENT_REQUIRED"
	ErrInvalidFinancialTransition = "INVALID_FINANCIAL_TRANSITION"
)

type CommercialTermsPayload struct {
	TermsType           string  `json:"terms_type"`
	Currency            string  `json:"currency"`
	Subtotal            string  `json:"subtotal"`
	DiscountAmount      string  `json:"discount_amount"`
	TaxAmount           string  `json:"tax_amount"`
	AdditionalCharge    string  `json:"additional_charge_amount"`
	FinalCustomerAmount string  `json:"final_customer_amount"`
	DepositPercentage   *string `json:"deposit_percentage"`
	DepositAmount       *string `json:"deposit_amount"`
	PaymentTermsText    string  `json:"payment_terms_text"`
	DeliveryTermsText   string  `json:"delivery_terms_text"`
	Reason              string  `json:"reason"`
}

type CommercialTerms struct {
	OrderID                string    `json:"order_id"`
	TermsType              string    `json:"terms_type"`
	Currency               string    `json:"currency"`
	Subtotal               string    `json:"subtotal"`
	DiscountAmount         string    `json:"discount_amount"`
	TaxAmount              string    `json:"tax_amount"`
	AdditionalChargeAmount string    `json:"additional_charge_amount"`
	FinalCustomerAmount    string    `json:"final_customer_amount"`
	DepositPercentage      *string   `json:"deposit_percentage,omitempty"`
	DepositAmount          *string   `json:"deposit_amount,omitempty"`
	PaymentTermsText       string    `json:"payment_terms_text"`
	DeliveryTermsText      string    `json:"delivery_terms_text"`
	VersionNumber          int       `json:"version_number"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type PaymentSchedulePayload struct {
	Items []PaymentScheduleItemPayload `json:"items"`
}

type PaymentScheduleItemPayload struct {
	TitleFA         string     `json:"title_fa"`
	PaymentType     string     `json:"payment_type"`
	DueAt           *time.Time `json:"due_at"`
	Amount          string     `json:"amount"`
	Percentage      *string    `json:"percentage_of_order"`
	Currency        string     `json:"currency"`
	TriggerType     string     `json:"trigger_type"`
	TriggerStepCode string     `json:"trigger_step_code"`
	CustomerVisible *bool      `json:"customer_visible"`
}

type PaymentScheduleItem struct {
	ID              string     `json:"id"`
	Sequence        int        `json:"sequence_number"`
	TitleFA         string     `json:"title_fa"`
	PaymentType     string     `json:"payment_type"`
	DueAt           *time.Time `json:"due_at,omitempty"`
	Amount          string     `json:"amount"`
	Percentage      *string    `json:"percentage_of_order,omitempty"`
	Currency        string     `json:"currency"`
	PaidAmount      string     `json:"paid_amount"`
	Status          string     `json:"status"`
	TriggerType     string     `json:"trigger_type"`
	TriggerStepCode *string    `json:"trigger_step_code,omitempty"`
	CustomerVisible bool       `json:"customer_visible"`
}

type PaymentPayload struct {
	Amount          string     `json:"amount"`
	Currency        string     `json:"currency"`
	PaymentMethod   string     `json:"payment_method"`
	ReferenceNumber string     `json:"reference_number"`
	BankName        string     `json:"bank_name"`
	ReceiptFileID   *string    `json:"receipt_file_id"`
	PaidAt          *time.Time `json:"paid_at"`
	Notes           string     `json:"notes"`
}

type PaymentAllocationPayload struct {
	ScheduleID string `json:"schedule_id"`
	Amount     string `json:"amount"`
}

type PaymentConfirmPayload struct {
	Allocations []PaymentAllocationPayload `json:"allocations"`
}

type PaymentRefundPayload struct {
	Amount          string                     `json:"amount"`
	Reason          string                     `json:"reason"`
	ReferenceNumber string                     `json:"reference_number"`
	Allocations     []PaymentAllocationPayload `json:"allocations"`
}

type Payment struct {
	ID              string     `json:"id"`
	PaymentNumber   string     `json:"payment_number"`
	OrderID         string     `json:"order_id"`
	Amount          string     `json:"amount"`
	Currency        string     `json:"currency"`
	PaymentMethod   string     `json:"payment_method"`
	Status          string     `json:"status"`
	ReferenceNumber *string    `json:"reference_number,omitempty"`
	PaidAt          time.Time  `json:"paid_at"`
	ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
	RefundedAmount  string     `json:"refunded_amount"`
}

type CostFlowPayload struct {
	Reason           string `json:"reason"`
	PaymentReference string `json:"payment_reference"`
}

type ExchangeRatePayload struct {
	BaseCurrency  string    `json:"base_currency"`
	QuoteCurrency string    `json:"quote_currency"`
	Rate          string    `json:"rate"`
	EffectiveAt   time.Time `json:"effective_at"`
	Notes         string    `json:"notes"`
}

type Notification struct {
	ID         string         `json:"id"`
	EventType  string         `json:"event_type"`
	Title      string         `json:"title"`
	Body       string         `json:"body"`
	EntityType *string        `json:"entity_type,omitempty"`
	EntityID   *string        `json:"entity_id,omitempty"`
	DeepLink   *string        `json:"deep_link,omitempty"`
	Data       map[string]any `json:"data"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
}

type NotificationPreferencePayload struct {
	EventType    string `json:"event_type"`
	InAppEnabled bool   `json:"in_app_enabled"`
	SMSEnabled   bool   `json:"sms_enabled"`
}

type DocumentGeneratePayload struct {
	DocumentType    string         `json:"document_type"`
	ScopeType       string         `json:"scope_type"`
	ScopeID         string         `json:"scope_id"`
	CustomerVisible bool           `json:"customer_visible"`
	Data            map[string]any `json:"data"`
}

type DocumentTemplateUpdate struct {
	NameFA   string         `json:"name_fa"`
	Template map[string]any `json:"template_json"`
	IsActive bool           `json:"is_active"`
}

type ContactLogPayload struct {
	ContactType string     `json:"contact_type"`
	Direction   string     `json:"direction"`
	ReasonCode  string     `json:"reason_code"`
	ResultCode  string     `json:"result_code"`
	Subject     string     `json:"subject"`
	Summary     string     `json:"summary"`
	FollowUpAt  *time.Time `json:"follow_up_at"`
	ContactedAt *time.Time `json:"contacted_at"`
}
