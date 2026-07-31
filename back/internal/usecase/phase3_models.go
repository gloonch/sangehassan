package usecase

import (
	"encoding/json"
	"time"
)

const quantityScale = 4

var quantityUnits = map[string]bool{
	"TON": true, "KILOGRAM": true, "CUBIC_METER": true, "SQUARE_METER": true,
	"BLOCK": true, "SLAB": true, "TILE": true, "PIECE": true, "PACKAGE": true,
	"BUNDLE": true, "CONTAINER": true,
}

type OperationConflict struct {
	Code    string
	Message string
}

func (e *OperationConflict) Error() string { return e.Message }

func conflict(code, message string) error { return &OperationConflict{Code: code, Message: message} }

type OrderItemPayload struct {
	ProductID          *int64  `json:"product_id"`
	StoneCategory      string  `json:"stone_category"`
	StoneName          string  `json:"stone_name"`
	StoneVariant       string  `json:"stone_variant"`
	FinishType         string  `json:"finish_type"`
	CutType            string  `json:"cut_type"`
	ThicknessValue     *string `json:"thickness_value"`
	ThicknessUnit      string  `json:"thickness_unit"`
	WidthValue         *string `json:"width_value"`
	LengthValue        *string `json:"length_value"`
	DimensionUnit      string  `json:"dimension_unit"`
	OrderedQuantity    string  `json:"ordered_quantity"`
	QuantityUnit       string  `json:"quantity_unit"`
	QualityGrade       string  `json:"quality_grade"`
	Color              string  `json:"color"`
	Pattern            string  `json:"pattern"`
	ProgressWeight     string  `json:"progress_weight"`
	RequiresProduction *bool   `json:"requires_production"`
	RequiresPackaging  *bool   `json:"requires_packaging"`
	Notes              string  `json:"notes"`
}

type OrderItem struct {
	ID                 string    `json:"id"`
	OrderID            string    `json:"order_id"`
	ProductID          *int64    `json:"product_id,omitempty"`
	StoneCategory      string    `json:"stone_category"`
	StoneName          string    `json:"stone_name"`
	StoneVariant       string    `json:"stone_variant"`
	FinishType         string    `json:"finish_type"`
	CutType            string    `json:"cut_type"`
	OrderedQuantity    string    `json:"ordered_quantity"`
	QuantityUnit       string    `json:"quantity_unit"`
	QualityGrade       string    `json:"quality_grade"`
	ProgressWeight     string    `json:"progress_weight"`
	RequiresProduction bool      `json:"requires_production"`
	RequiresPackaging  bool      `json:"requires_packaging"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
}

type QuantityConversionPayload struct {
	FromQuantity string `json:"from_quantity"`
	FromUnit     string `json:"from_unit"`
	ToQuantity   string `json:"to_quantity"`
	ToUnit       string `json:"to_unit"`
	Reason       string `json:"reason"`
}

type BatchPayload struct {
	OrderItemID        string  `json:"order_item_id"`
	ParentBatchID      *string `json:"parent_batch_id"`
	SourceType         string  `json:"source_type"`
	SourceLocationID   *string `json:"source_location_id"`
	TargetLocationID   *string `json:"target_location_id"`
	StoneCategory      string  `json:"stone_category"`
	StoneName          string  `json:"stone_name"`
	StoneVariant       string  `json:"stone_variant"`
	FinishType         string  `json:"finish_type"`
	CutType            string  `json:"cut_type"`
	ThicknessValue     *string `json:"thickness_value"`
	ThicknessUnit      string  `json:"thickness_unit"`
	PlannedQuantity    string  `json:"planned_quantity"`
	QuantityUnit       string  `json:"quantity_unit"`
	Priority           string  `json:"priority"`
	IsRequired         *bool   `json:"is_required"`
	WorkflowTemplateID *int64  `json:"workflow_template_id"`
	ParentWorkflowID   *string `json:"parent_workflow_instance_id"`
	ParentStepID       *string `json:"parent_step_instance_id"`
	Override           bool    `json:"override"`
	Reason             string  `json:"reason"`
}

type Batch struct {
	ID                 string     `json:"id"`
	BatchNumber        string     `json:"batch_number"`
	OrderID            string     `json:"order_id"`
	OrderNumber        string     `json:"order_number"`
	OrderItemID        string     `json:"order_item_id"`
	ParentBatchID      *string    `json:"parent_batch_id,omitempty"`
	SourceType         string     `json:"source_type"`
	SourceLocationID   *string    `json:"source_location_id,omitempty"`
	TargetLocationID   *string    `json:"target_location_id,omitempty"`
	StoneName          string     `json:"stone_name"`
	StoneVariant       string     `json:"stone_variant"`
	PlannedQuantity    string     `json:"planned_quantity"`
	ActualQuantity     string     `json:"actual_quantity"`
	QuantityUnit       string     `json:"quantity_unit"`
	Status             string     `json:"status"`
	Priority           string     `json:"priority"`
	IsRequired         bool       `json:"is_required"`
	WorkflowInstanceID *string    `json:"workflow_instance_id,omitempty"`
	EstimatedReadyAt   *time.Time `json:"estimated_ready_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

type BatchSplitChild struct {
	PlannedQuantity        string                  `json:"planned_quantity"`
	SourceType             string                  `json:"source_type"`
	SourceLocationID       *string                 `json:"source_location_id"`
	TargetLocationID       *string                 `json:"target_location_id"`
	WorkflowTemplateID     *int64                  `json:"workflow_template_id"`
	ReservationAllocations []ReservationAllocation `json:"reservation_allocations"`
}
type ReservationAllocation struct {
	ReservationID string `json:"reservation_id"`
	Quantity      string `json:"quantity"`
}
type BatchSplitPayload struct {
	Children []BatchSplitChild `json:"children"`
	Reason   string            `json:"reason"`
}
type BatchMergePayload struct {
	BatchIDs           []string `json:"batch_ids"`
	WorkflowTemplateID *int64   `json:"workflow_template_id"`
	Reason             string   `json:"reason"`
}

type LocationPayload struct {
	Code         string  `json:"code"`
	NameFA       string  `json:"name_fa"`
	LocationType string  `json:"location_type"`
	Address      string  `json:"address"`
	City         string  `json:"city"`
	Province     string  `json:"province"`
	CountryCode  string  `json:"country_code"`
	Latitude     *string `json:"latitude"`
	Longitude    *string `json:"longitude"`
	IsActive     *bool   `json:"is_active"`
}
type InventoryLocation struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	NameFA       string `json:"name_fa"`
	LocationType string `json:"location_type"`
	Address      string `json:"address"`
	City         string `json:"city"`
	Province     string `json:"province"`
	CountryCode  string `json:"country_code"`
	IsActive     bool   `json:"is_active"`
}

type LotPayload struct {
	ParentLotID       *string `json:"parent_lot_id"`
	OriginType        string  `json:"origin_type"`
	OriginReferenceID string  `json:"origin_reference_id"`
	LocationID        string  `json:"current_location_id"`
	StoneCategory     string  `json:"stone_category"`
	StoneName         string  `json:"stone_name"`
	StoneVariant      string  `json:"stone_variant"`
	QualityGrade      string  `json:"quality_grade"`
	FinishType        string  `json:"finish_type"`
	CutType           string  `json:"cut_type"`
	Quantity          string  `json:"quantity"`
	QuantityUnit      string  `json:"quantity_unit"`
	SecondaryQuantity *string `json:"secondary_quantity"`
	SecondaryUnit     string  `json:"secondary_unit"`
}
type InventoryLot struct {
	ID                string    `json:"id"`
	LotNumber         string    `json:"lot_number"`
	ParentLotID       *string   `json:"parent_lot_id,omitempty"`
	CurrentLocationID string    `json:"current_location_id"`
	LocationName      string    `json:"location_name"`
	OriginType        string    `json:"origin_type"`
	StoneName         string    `json:"stone_name"`
	StoneVariant      string    `json:"stone_variant"`
	AvailableQuantity string    `json:"available_quantity"`
	ReservedQuantity  string    `json:"reserved_quantity"`
	QuantityUnit      string    `json:"quantity_unit"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}
type ReservationPayload struct {
	OrderID                string     `json:"order_id"`
	OrderItemID            string     `json:"order_item_id"`
	BatchID                string     `json:"batch_id"`
	Quantity               string     `json:"quantity"`
	QuantityUnit           string     `json:"quantity_unit"`
	ExpiresAt              *time.Time `json:"expires_at"`
	WorkflowStepInstanceID *string    `json:"workflow_step_instance_id"`
}
type InventoryReservation struct {
	ID               string     `json:"id"`
	InventoryLotID   string     `json:"inventory_lot_id"`
	BatchID          string     `json:"batch_id"`
	ReservedQuantity string     `json:"reserved_quantity"`
	ConsumedQuantity string     `json:"consumed_quantity"`
	QuantityUnit     string     `json:"quantity_unit"`
	Status           string     `json:"status"`
	ReservedAt       time.Time  `json:"reserved_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

type ReceiptPayload struct {
	Lot                    LotPayload `json:"lot"`
	OrderID                *string    `json:"order_id"`
	BatchID                *string    `json:"batch_id"`
	Reason                 string     `json:"reason"`
	WorkflowStepInstanceID *string    `json:"workflow_step_instance_id"`
}
type TransferPayload struct {
	LotID                 string `json:"lot_id"`
	DestinationLocationID string `json:"destination_location_id"`
	Quantity              string `json:"quantity"`
	Reason                string `json:"reason"`
}
type AdjustmentPayload struct {
	LotID                string  `json:"lot_id"`
	QuantityDelta        string  `json:"quantity_delta"`
	Reason               string  `json:"reason"`
	ReversalOfMovementID *string `json:"reversal_of_movement_id"`
}
type ConversionPayload struct {
	BatchID                string     `json:"batch_id"`
	WorkflowStepInstanceID *string    `json:"workflow_step_instance_id"`
	InputLotID             string     `json:"input_lot_id"`
	InputQuantity          string     `json:"input_quantity"`
	InputUnit              string     `json:"input_unit"`
	OutputQuantity         string     `json:"output_quantity"`
	OutputUnit             string     `json:"output_unit"`
	WasteQuantity          string     `json:"waste_quantity"`
	WasteUnit              string     `json:"waste_unit"`
	ConversionType         string     `json:"conversion_type"`
	OrderItemConversionID  *string    `json:"order_item_conversion_id"`
	OutputLot              LotPayload `json:"output_lot"`
}
type InventoryMovement struct {
	ID               string    `json:"id"`
	MovementNumber   string    `json:"movement_number"`
	OperationGroupID string    `json:"operation_group_id"`
	MovementType     string    `json:"movement_type"`
	InventoryLotID   string    `json:"inventory_lot_id"`
	Quantity         string    `json:"quantity"`
	QuantityUnit     string    `json:"quantity_unit"`
	BeforeAvailable  string    `json:"before_available_quantity"`
	AfterAvailable   string    `json:"after_available_quantity"`
	BeforeReserved   string    `json:"before_reserved_quantity"`
	AfterReserved    string    `json:"after_reserved_quantity"`
	Reason           string    `json:"reason"`
	OccurredAt       time.Time `json:"occurred_at"`
}

type VehiclePayload struct {
	VehicleType   string  `json:"vehicle_type"`
	PlateNumber   string  `json:"plate_number"`
	TrailerNumber string  `json:"trailer_number"`
	CapacityValue *string `json:"capacity_value"`
	CapacityUnit  string  `json:"capacity_unit"`
	OwnerName     string  `json:"owner_name"`
	CarrierName   string  `json:"carrier_name"`
	DriverUserID  *string `json:"driver_user_id"`
	IsActive      *bool   `json:"is_active"`
}
type Vehicle struct {
	ID              string  `json:"id"`
	VehicleType     string  `json:"vehicle_type"`
	PlateNumber     string  `json:"plate_number"`
	PlateNormalized string  `json:"plate_normalized"`
	CapacityValue   *string `json:"capacity_value,omitempty"`
	CapacityUnit    string  `json:"capacity_unit"`
	DriverUserID    *string `json:"driver_user_id,omitempty"`
	IsActive        bool    `json:"is_active"`
}

type ShipmentPayload struct {
	ShipmentType          string     `json:"shipment_type"`
	OriginLocationID      string     `json:"origin_location_id"`
	DestinationLocationID *string    `json:"destination_location_id"`
	DriverUserID          *string    `json:"driver_user_id"`
	ExternalDriverName    string     `json:"external_driver_name"`
	ExternalDriverPhone   string     `json:"external_driver_phone"`
	CarrierName           string     `json:"carrier_name"`
	VehicleID             *string    `json:"vehicle_id"`
	PlannedDepartureAt    *time.Time `json:"planned_departure_at"`
	EstimatedArrivalAt    *time.Time `json:"estimated_arrival_at"`
	DeliveryContactName   string     `json:"delivery_contact_name"`
	DeliveryContactPhone  string     `json:"delivery_contact_phone"`
	DeliveryAddress       string     `json:"delivery_address"`
	WorkflowTemplateID    *int64     `json:"workflow_template_id"`
	ParentWorkflowID      *string    `json:"parent_workflow_instance_id"`
	ParentStepID          *string    `json:"parent_step_instance_id"`
	CustomerTitleFA       string     `json:"customer_title_fa"`
	CustomerVisible       *bool      `json:"customer_visible"`
	Notes                 string     `json:"notes"`
}
type Shipment struct {
	ID                    string         `json:"id"`
	ShipmentNumber        string         `json:"shipment_number"`
	OrderID               string         `json:"order_id"`
	OrderNumber           string         `json:"order_number"`
	ShipmentType          string         `json:"shipment_type"`
	Status                string         `json:"status"`
	OriginLocationID      string         `json:"origin_location_id"`
	DestinationLocationID *string        `json:"destination_location_id,omitempty"`
	DriverUserID          *string        `json:"driver_user_id,omitempty"`
	VehicleID             *string        `json:"vehicle_id,omitempty"`
	WorkflowInstanceID    *string        `json:"workflow_instance_id,omitempty"`
	CustomerVisible       bool           `json:"customer_visible"`
	CustomerTitleFA       string         `json:"customer_title_fa"`
	PlannedDepartureAt    *time.Time     `json:"planned_departure_at,omitempty"`
	EstimatedArrivalAt    *time.Time     `json:"estimated_arrival_at,omitempty"`
	ActualDepartureAt     *time.Time     `json:"actual_departure_at,omitempty"`
	ActualArrivalAt       *time.Time     `json:"actual_arrival_at,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	Items                 []ShipmentItem `json:"items,omitempty"`
}
type ShipmentItemPayload struct {
	BatchID         string `json:"batch_id"`
	InventoryLotID  string `json:"inventory_lot_id"`
	PlannedQuantity string `json:"planned_quantity"`
	QuantityUnit    string `json:"quantity_unit"`
	PackageCount    int    `json:"package_count"`
	BundleCount     int    `json:"bundle_count"`
	Notes           string `json:"notes"`
}
type ShipmentItem struct {
	ID                string `json:"id"`
	ShipmentID        string `json:"shipment_id"`
	BatchID           string `json:"batch_id"`
	BatchNumber       string `json:"batch_number"`
	InventoryLotID    string `json:"inventory_lot_id"`
	PlannedQuantity   string `json:"planned_quantity"`
	LoadedQuantity    string `json:"loaded_quantity"`
	DeliveredQuantity string `json:"delivered_quantity"`
	QuantityUnit      string `json:"quantity_unit"`
	PackageCount      int    `json:"package_count"`
	BundleCount       int    `json:"bundle_count"`
}
type ShipmentQuantity struct {
	ShipmentItemID string `json:"shipment_item_id"`
	Quantity       string `json:"quantity"`
}
type ShipmentOperationPayload struct {
	Items                  []ShipmentQuantity `json:"items"`
	FinalizeLoading        bool               `json:"finalize_loading"`
	FinalizeDelivery       bool               `json:"finalize_delivery"`
	Reason                 string             `json:"reason"`
	ReceiverName           string             `json:"receiver_name"`
	ReceiverPhone          string             `json:"receiver_phone"`
	ProofFileID            *string            `json:"proof_file_id"`
	WorkflowStepInstanceID *string            `json:"workflow_step_instance_id"`
}

type PackagingPayload struct {
	InventoryLotID         string  `json:"inventory_lot_id"`
	PackageType            string  `json:"package_type"`
	Quantity               string  `json:"quantity"`
	QuantityUnit           string  `json:"quantity_unit"`
	GrossWeight            *string `json:"gross_weight"`
	NetWeight              *string `json:"net_weight"`
	WeightUnit             string  `json:"weight_unit"`
	CustomerVisible        bool    `json:"customer_visible"`
	WorkflowStepInstanceID *string `json:"workflow_step_instance_id"`
}
type ContainerPayload struct {
	ContainerNumber string  `json:"container_number"`
	ContainerType   string  `json:"container_type"`
	SealNumber      string  `json:"seal_number"`
	TareWeight      *string `json:"tare_weight"`
	GrossWeight     *string `json:"gross_weight"`
	NetWeight       *string `json:"net_weight"`
	WeightUnit      string  `json:"weight_unit"`
	PackageCount    int     `json:"package_count"`
}
type ContainerItemPayload struct {
	ShipmentItemID string `json:"shipment_item_id"`
	Quantity       string `json:"quantity"`
	QuantityUnit   string `json:"quantity_unit"`
	PackageCount   int    `json:"package_count"`
}

type CostPayload struct {
	EntityType string     `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	CostType   string     `json:"cost_type"`
	Amount     string     `json:"amount"`
	Currency   string     `json:"currency"`
	Notes      string     `json:"notes"`
	IncurredAt *time.Time `json:"incurred_at"`
}

type ProgressStage struct {
	Quantity string  `json:"quantity"`
	Percent  float64 `json:"percent"`
}
type OrderItemProgress struct {
	OrderItemID         string        `json:"order_item_id"`
	StoneName           string        `json:"stone_name"`
	OrderedQuantity     string        `json:"ordered_quantity"`
	QuantityUnit        string        `json:"quantity_unit"`
	Planned             ProgressStage `json:"planned"`
	Reserved            ProgressStage `json:"reserved"`
	InProduction        ProgressStage `json:"in_production"`
	Produced            ProgressStage `json:"produced"`
	Packaged            ProgressStage `json:"packaged"`
	Shipped             ProgressStage `json:"shipped"`
	Delivered           ProgressStage `json:"delivered"`
	RemainingQuantity   string        `json:"remaining_quantity"`
	ProcurementProgress float64       `json:"procurement_progress"`
	ProductionProgress  float64       `json:"production_progress"`
	PackagingProgress   float64       `json:"packaging_progress"`
	ShippingProgress    float64       `json:"shipping_progress"`
	DeliveryProgress    float64       `json:"delivery_progress"`
	OverallProgress     float64       `json:"overall_progress"`
	ProgressWeight      string        `json:"progress_weight"`
}
type OrderProgress struct {
	OrderID         string              `json:"order_id"`
	OrderNumber     string              `json:"order_number"`
	OverallProgress float64             `json:"overall_progress"`
	Items           []OrderItemProgress `json:"items"`
}

type WorkflowTransitionPayload struct {
	SourceStepID           int64   `json:"source_step_id"`
	TargetStepID           int64   `json:"target_step_id"`
	TransitionCode         string  `json:"transition_code"`
	LabelFA                string  `json:"label_fa"`
	TransitionType         string  `json:"transition_type"`
	ResultCode             *string `json:"result_code"`
	IsDefault              bool    `json:"is_default"`
	RequiresPermissionCode *string `json:"requires_permission_code"`
	RequiresReason         bool    `json:"requires_reason"`
	SortOrder              int     `json:"sort_order"`
}
type WorkflowTransitionDefinition struct {
	ID                     int64   `json:"id"`
	WorkflowTemplateID     int64   `json:"workflow_template_id"`
	SourceStepID           int64   `json:"source_step_id"`
	TargetStepID           int64   `json:"target_step_id"`
	TransitionCode         string  `json:"transition_code"`
	LabelFA                string  `json:"label_fa"`
	TransitionType         string  `json:"transition_type"`
	ResultCode             *string `json:"result_code,omitempty"`
	IsDefault              bool    `json:"is_default"`
	RequiresPermissionCode *string `json:"requires_permission_code,omitempty"`
	RequiresReason         bool    `json:"requires_reason"`
	SortOrder              int     `json:"sort_order"`
}
type RuntimeTransition struct {
	ID             string  `json:"id"`
	TransitionCode string  `json:"transition_code"`
	LabelFA        string  `json:"label_fa"`
	TransitionType string  `json:"transition_type"`
	ResultCode     *string `json:"result_code,omitempty"`
	RequiresReason bool    `json:"requires_reason"`
	TargetStepCode string  `json:"target_step_code"`
}
type SelectTransitionPayload struct {
	TransitionCode string  `json:"transition_code"`
	ResultCode     *string `json:"result_code"`
	Reason         string  `json:"reason"`
	Override       bool    `json:"override"`
}

type IdempotentResult struct {
	Existing bool
	Response json.RawMessage
}
