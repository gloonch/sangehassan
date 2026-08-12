package usecase

import (
	"encoding/json"
	"time"
)

type QualityInspectionPayload struct {
	OrderID                string                     `json:"order_id"`
	BatchID                *string                    `json:"batch_id"`
	InventoryLotID         *string                    `json:"inventory_lot_id"`
	WorkflowStepInstanceID *string                    `json:"workflow_step_instance_id"`
	InspectionType         string                     `json:"inspection_type"`
	AssignedUserID         *string                    `json:"assigned_user_id"`
	AssignedRoleID         *int64                     `json:"assigned_role_id"`
	Notes                  string                     `json:"notes"`
	Values                 map[string]json.RawMessage `json:"values"`
}

type QualityDecisionPayload struct {
	Values map[string]json.RawMessage `json:"values"`
	Notes  string                     `json:"notes"`
	Reason string                     `json:"reason"`
}

type QualityInspection struct {
	ID                     string                     `json:"id"`
	InspectionNumber       string                     `json:"inspection_number"`
	OrderID                string                     `json:"order_id"`
	OrderNumber            string                     `json:"order_number"`
	BatchID                *string                    `json:"batch_id,omitempty"`
	BatchNumber            *string                    `json:"batch_number,omitempty"`
	InventoryLotID         *string                    `json:"inventory_lot_id,omitempty"`
	LotNumber              *string                    `json:"lot_number,omitempty"`
	WorkflowStepInstanceID *string                    `json:"workflow_step_instance_id,omitempty"`
	InspectionType         string                     `json:"inspection_type"`
	Status                 string                     `json:"status"`
	AssignedUserID         *string                    `json:"assigned_user_id,omitempty"`
	AssignedRoleID         *int64                     `json:"assigned_role_id,omitempty"`
	InspectedByUserID      *string                    `json:"inspected_by_user_id,omitempty"`
	InspectedAt            *time.Time                 `json:"inspected_at,omitempty"`
	Notes                  string                     `json:"notes"`
	DecisionReason         string                     `json:"decision_reason"`
	Values                 map[string]json.RawMessage `json:"values,omitempty"`
	CreatedAt              time.Time                  `json:"created_at"`
	UpdatedAt              time.Time                  `json:"updated_at"`
}

type InstallationPayload struct {
	ProjectName            string     `json:"project_name"`
	ProjectAddress         string     `json:"project_address"`
	ContactName            string     `json:"contact_name"`
	ContactPhone           string     `json:"contact_phone"`
	InstallationLeadUserID *string    `json:"installation_lead_user_id"`
	SupplierID             *string    `json:"supplier_id"`
	PlannedStartAt         *time.Time `json:"planned_start_at"`
	EstimatedEndAt         *time.Time `json:"estimated_end_at"`
	PlannedArea            *string    `json:"planned_area"`
	AreaUnit               string     `json:"area_unit"`
	Notes                  string     `json:"notes"`
	StartWorkflow          bool       `json:"start_workflow"`
	WorkflowTemplateID     *int64     `json:"workflow_template_id"`
	ParentWorkflowID       *string    `json:"parent_workflow_instance_id"`
	ParentStepID           *string    `json:"parent_step_instance_id"`
	ExcludedOptionalSteps  []string   `json:"excluded_optional_step_codes"`
}

type InstallationMemberPayload struct {
	UserID       *string `json:"user_id"`
	NameOverride string  `json:"name_override"`
	RoleLabel    string  `json:"role_label"`
}

type InstallationUpdatePayload struct {
	UpdateDate             *time.Time `json:"update_date"`
	InstalledQuantity      string     `json:"installed_quantity"`
	QuantityUnit           string     `json:"quantity_unit"`
	Status                 string     `json:"status"`
	Description            string     `json:"description"`
	CustomerVisible        bool       `json:"customer_visible"`
	Activities             []string   `json:"activities"`
	WorkflowStepInstanceID *string    `json:"workflow_step_instance_id"`
}

type InstallationIssuePayload struct {
	IssueType       string `json:"issue_type"`
	Severity        string `json:"severity"`
	Description     string `json:"description"`
	CustomerVisible bool   `json:"customer_visible"`
	ResolutionNote  string `json:"resolution_note"`
}

type InstallationMaterialPayload struct {
	MaterialName string  `json:"material_name"`
	Quantity     string  `json:"quantity"`
	Unit         string  `json:"unit"`
	CostEntryID  *string `json:"cost_entry_id"`
	Notes        string  `json:"notes"`
}

type InstallationFlowPayload struct {
	Reason                 string  `json:"reason"`
	Force                  bool    `json:"force"`
	WorkflowStepInstanceID *string `json:"workflow_step_instance_id"`
}

type InstallationJob struct {
	ID                     string           `json:"id"`
	InstallationNumber     string           `json:"installation_number"`
	OrderID                string           `json:"order_id"`
	OrderNumber            string           `json:"order_number"`
	CustomerUserID         string           `json:"customer_user_id"`
	ProjectName            string           `json:"project_name"`
	ProjectAddress         string           `json:"project_address"`
	ContactName            string           `json:"contact_name"`
	ContactPhone           string           `json:"contact_phone"`
	Status                 string           `json:"status"`
	InstallationLeadUserID *string          `json:"installation_lead_user_id,omitempty"`
	SupplierID             *string          `json:"supplier_id,omitempty"`
	SupplierName           *string          `json:"supplier_name,omitempty"`
	PlannedStartAt         *time.Time       `json:"planned_start_at,omitempty"`
	ActualStartAt          *time.Time       `json:"actual_start_at,omitempty"`
	EstimatedEndAt         *time.Time       `json:"estimated_end_at,omitempty"`
	ActualEndAt            *time.Time       `json:"actual_end_at,omitempty"`
	PlannedArea            *string          `json:"planned_area,omitempty"`
	InstalledArea          string           `json:"installed_area"`
	AreaUnit               string           `json:"area_unit"`
	Notes                  string           `json:"notes,omitempty"`
	WorkflowInstanceID     *string          `json:"workflow_instance_id,omitempty"`
	Members                []map[string]any `json:"members,omitempty"`
	Updates                []map[string]any `json:"updates,omitempty"`
	Issues                 []map[string]any `json:"issues,omitempty"`
	Materials              []map[string]any `json:"materials,omitempty"`
	Files                  []map[string]any `json:"files,omitempty"`
	Acceptances            []map[string]any `json:"acceptances,omitempty"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
}

type CustomerAcceptancePayload struct {
	InstallationJobID      *string `json:"installation_job_id"`
	ShipmentID             *string `json:"shipment_id"`
	WorkflowStepInstanceID *string `json:"workflow_step_instance_id"`
	CustomerName           string  `json:"customer_name"`
	Accepted               bool    `json:"accepted"`
	Comment                string  `json:"comment"`
	SignatureFileID        *string `json:"signature_file_id"`
	PhotoFileID            *string `json:"photo_file_id"`
}

type OrderClosurePayload struct {
	Force  bool   `json:"force"`
	Reason string `json:"reason"`
}
