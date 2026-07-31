package usecase

import (
	"encoding/json"
	"time"
)

var fieldTypes = map[string]bool{"SHORT_TEXT": true, "LONG_TEXT": true, "INTEGER": true, "DECIMAL": true, "BOOLEAN": true, "DATE": true, "TIME": true, "DATETIME": true, "MONEY": true, "SELECT": true, "MULTI_SELECT": true, "PHONE": true, "ADDRESS": true, "WEIGHT": true, "AREA": true, "VOLUME": true, "QUANTITY": true, "IMAGE": true, "FILE": true, "SIGNATURE": true}
var stepTriggerTypes = map[string]bool{"ON_STEP_OPEN": true, "ON_STEP_START": true, "ON_STEP_SUBMIT": true, "ON_STEP_APPROVE": true, "ON_STEP_COMPLETE": true}

type WorkflowStepCatalogueItem struct {
	ID                    int64  `json:"id"`
	StepCode              string `json:"step_code"`
	TitleFA               string `json:"title_fa"`
	CustomerTitleFA       string `json:"customer_title_fa"`
	DescriptionFA         string `json:"description_fa"`
	DefaultRoleCode       string `json:"default_role_code"`
	DefaultPermissionCode string `json:"default_permission_code"`
	DefaultDurationHours  int    `json:"default_duration_hours"`
}

type WorkflowTemplateVersion struct {
	ID                    int64                          `json:"id"`
	TemplateGroupCode     string                         `json:"template_group_code"`
	VersionNumber         int                            `json:"version_number"`
	Code                  string                         `json:"code"`
	NameFA                string                         `json:"name_fa"`
	DescriptionFA         string                         `json:"description_fa"`
	IconKey               string                         `json:"icon_key"`
	Status                string                         `json:"status"`
	ScopeType             string                         `json:"scope_type"`
	MaxIterations         int                            `json:"max_iterations"`
	StartPermissionCode   string                         `json:"start_permission_code"`
	IsActive              bool                           `json:"is_active"`
	CreatedFromTemplateID *int64                         `json:"created_from_template_id,omitempty"`
	PublishedAt           *time.Time                     `json:"published_at,omitempty"`
	CreatedAt             time.Time                      `json:"created_at"`
	UpdatedAt             time.Time                      `json:"updated_at"`
	Steps                 []WorkflowTemplateStepV2       `json:"steps,omitempty"`
	Metrics               []HandoffMetricDefinition      `json:"handoff_metrics,omitempty"`
	Tasks                 []WorkflowLevelTask            `json:"workflow_tasks,omitempty"`
	Transitions           []WorkflowTransitionDefinition `json:"transitions,omitempty"`
}
type WorkflowTemplateStepV2 struct {
	ID                     int64                     `json:"id"`
	WorkflowTemplateID     int64                     `json:"workflow_template_id"`
	StepCode               string                    `json:"step_code"`
	InternalTitleFA        string                    `json:"internal_title_fa"`
	InternalDescriptionFA  string                    `json:"internal_description_fa"`
	CustomerTitleFA        string                    `json:"customer_title_fa"`
	CustomerDescriptionFA  string                    `json:"customer_description_fa"`
	SequenceNumber         int                       `json:"sequence_number"`
	ResponsibleRoleID      *int64                    `json:"responsible_role_id,omitempty"`
	ResponsibleRoleCode    string                    `json:"responsible_role_code,omitempty"`
	RequiredPermissionCode string                    `json:"required_permission_code"`
	CustomerVisible        bool                      `json:"customer_visible"`
	RequiresApproval       bool                      `json:"requires_approval"`
	ApprovalRoleID         *int64                    `json:"approval_role_id,omitempty"`
	IsOptional             bool                      `json:"is_optional"`
	IsSkippable            bool                      `json:"is_skippable"`
	IsActive               bool                      `json:"is_active"`
	DefaultDurationHours   int                       `json:"default_duration_hours"`
	StartsAutomatically    bool                      `json:"starts_automatically"`
	IsEntry                bool                      `json:"is_entry"`
	DomainEventCode        *string                   `json:"domain_event_code,omitempty"`
	Fields                 []WorkflowFieldDefinition `json:"fields"`
	Tasks                  []WorkflowTaskTemplate    `json:"tasks"`
}
type WorkflowFieldDefinition struct {
	ID                     int64           `json:"id"`
	WorkflowTemplateStepID int64           `json:"workflow_template_step_id"`
	FieldKey               string          `json:"field_key"`
	LabelFA                string          `json:"label_fa"`
	DescriptionFA          string          `json:"description_fa"`
	FieldType              string          `json:"field_type"`
	IsRequired             bool            `json:"is_required"`
	IsCustomerVisible      bool            `json:"is_customer_visible"`
	IsSalesVisible         bool            `json:"is_sales_visible"`
	IsInternalCost         bool            `json:"is_internal_cost"`
	UnitCode               *string         `json:"unit_code,omitempty"`
	CurrencyCode           *string         `json:"currency_code,omitempty"`
	PlaceholderFA          *string         `json:"placeholder_fa,omitempty"`
	DefaultValue           json.RawMessage `json:"default_value,omitempty"`
	OptionsJSON            json.RawMessage `json:"options_json,omitempty"`
	ValidationJSON         json.RawMessage `json:"validation_json"`
	SortOrder              int             `json:"sort_order"`
	HandoffMetricKey       *string         `json:"handoff_metric_key,omitempty"`
	HandoffDirection       *string         `json:"handoff_direction,omitempty"`
}
type WorkflowTaskTemplate struct {
	ID                     int64  `json:"id"`
	WorkflowTemplateStepID int64  `json:"workflow_template_step_id"`
	TriggerType            string `json:"trigger_type"`
	TitleFA                string `json:"title_fa"`
	DescriptionFA          string `json:"description_fa"`
	AssignedRoleID         *int64 `json:"assigned_role_id,omitempty"`
	RequiredPermissionCode string `json:"required_permission_code"`
	Priority               string `json:"priority"`
	DueOffsetHours         *int   `json:"due_offset_hours,omitempty"`
	BlocksStepCompletion   bool   `json:"blocks_step_completion"`
}
type WorkflowLevelTask struct {
	ID                     int64  `json:"id"`
	WorkflowTemplateID     int64  `json:"workflow_template_id"`
	TriggerType            string `json:"trigger_type"`
	TitleFA                string `json:"title_fa"`
	DescriptionFA          string `json:"description_fa"`
	AssignedRoleID         *int64 `json:"assigned_role_id,omitempty"`
	RequiredPermissionCode string `json:"required_permission_code"`
	Priority               string `json:"priority"`
	DueOffsetHours         *int   `json:"due_offset_hours,omitempty"`
	BlocksWorkflowProgress bool   `json:"blocks_workflow_progress"`
}
type HandoffMetricDefinition struct {
	ID                  int64    `json:"id"`
	WorkflowTemplateID  int64    `json:"workflow_template_id"`
	MetricKey           string   `json:"metric_key"`
	LabelFA             string   `json:"label_fa"`
	UnitCode            string   `json:"unit_code"`
	AbsoluteTolerance   *float64 `json:"absolute_tolerance,omitempty"`
	PercentageTolerance *float64 `json:"percentage_tolerance,omitempty"`
	BlockingOnMismatch  bool     `json:"blocking_on_mismatch"`
}

type WorkflowTemplatePayload struct {
	TemplateGroupCode   string `json:"template_group_code"`
	Code                string `json:"code"`
	NameFA              string `json:"name_fa"`
	DescriptionFA       string `json:"description_fa"`
	IconKey             string `json:"icon_key"`
	StartPermissionCode string `json:"start_permission_code"`
	IsActive            bool   `json:"is_active"`
	ScopeType           string `json:"scope_type"`
	MaxIterations       int    `json:"max_iterations"`
}
type WorkflowStepPayload struct {
	StepCode               string  `json:"step_code"`
	InternalTitleFA        string  `json:"internal_title_fa"`
	InternalDescriptionFA  string  `json:"internal_description_fa"`
	CustomerTitleFA        string  `json:"customer_title_fa"`
	CustomerDescriptionFA  string  `json:"customer_description_fa"`
	ResponsibleRoleID      *int64  `json:"responsible_role_id"`
	RequiredPermissionCode string  `json:"required_permission_code"`
	CustomerVisible        bool    `json:"customer_visible"`
	RequiresApproval       bool    `json:"requires_approval"`
	ApprovalRoleID         *int64  `json:"approval_role_id"`
	IsOptional             bool    `json:"is_optional"`
	IsSkippable            bool    `json:"is_skippable"`
	IsActive               bool    `json:"is_active"`
	DefaultDurationHours   int     `json:"default_duration_hours"`
	StartsAutomatically    bool    `json:"starts_automatically"`
	IsEntry                bool    `json:"is_entry"`
	DomainEventCode        *string `json:"domain_event_code"`
}
type WorkflowFieldPayload struct {
	FieldKey          string          `json:"field_key"`
	LabelFA           string          `json:"label_fa"`
	DescriptionFA     string          `json:"description_fa"`
	FieldType         string          `json:"field_type"`
	IsRequired        bool            `json:"is_required"`
	IsCustomerVisible bool            `json:"is_customer_visible"`
	IsSalesVisible    bool            `json:"is_sales_visible"`
	IsInternalCost    bool            `json:"is_internal_cost"`
	UnitCode          *string         `json:"unit_code"`
	CurrencyCode      *string         `json:"currency_code"`
	PlaceholderFA     *string         `json:"placeholder_fa"`
	DefaultValue      json.RawMessage `json:"default_value"`
	OptionsJSON       json.RawMessage `json:"options_json"`
	ValidationJSON    json.RawMessage `json:"validation_json"`
	SortOrder         int             `json:"sort_order"`
	HandoffMetricKey  *string         `json:"handoff_metric_key"`
	HandoffDirection  *string         `json:"handoff_direction"`
}
type WorkflowTaskPayload struct {
	TriggerType            string `json:"trigger_type"`
	TitleFA                string `json:"title_fa"`
	DescriptionFA          string `json:"description_fa"`
	AssignedRoleID         *int64 `json:"assigned_role_id"`
	RequiredPermissionCode string `json:"required_permission_code"`
	Priority               string `json:"priority"`
	DueOffsetHours         *int   `json:"due_offset_hours"`
	BlocksStepCompletion   bool   `json:"blocks_step_completion"`
}
type WorkflowLevelTaskPayload struct {
	TriggerType            string `json:"trigger_type"`
	TitleFA                string `json:"title_fa"`
	DescriptionFA          string `json:"description_fa"`
	AssignedRoleID         *int64 `json:"assigned_role_id"`
	RequiredPermissionCode string `json:"required_permission_code"`
	Priority               string `json:"priority"`
	DueOffsetHours         *int   `json:"due_offset_hours"`
	BlocksWorkflowProgress bool   `json:"blocks_workflow_progress"`
}
type HandoffMetricPayload struct {
	MetricKey           string   `json:"metric_key"`
	LabelFA             string   `json:"label_fa"`
	UnitCode            string   `json:"unit_code"`
	AbsoluteTolerance   *float64 `json:"absolute_tolerance"`
	PercentageTolerance *float64 `json:"percentage_tolerance"`
	BlockingOnMismatch  bool     `json:"blocking_on_mismatch"`
}
