package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

var ErrInvalidTransition = errors.New("invalid workflow transition")
var ErrValidation = errors.New("workflow field validation failed")

var workflowStepTransitions = map[string]map[string]bool{
	"NOT_STARTED":          {"WAITING_FOR_ASSIGNEE": true, "CANCELLED": true},
	"WAITING_FOR_ASSIGNEE": {"IN_PROGRESS": true, "SKIPPED": true, "CANCELLED": true},
	"IN_PROGRESS":          {"WAITING_FOR_APPROVAL": true, "HAS_MISMATCH": true, "BLOCKED": true, "COMPLETED": true, "SKIPPED": true, "CANCELLED": true},
	"WAITING_FOR_APPROVAL": {"COMPLETED": true, "NEEDS_CORRECTION": true, "CANCELLED": true},
	"NEEDS_CORRECTION":     {"IN_PROGRESS": true, "SKIPPED": true, "CANCELLED": true},
	"HAS_MISMATCH":         {"WAITING_FOR_APPROVAL": true, "NEEDS_CORRECTION": true, "BLOCKED": true, "COMPLETED": true, "CANCELLED": true},
	"BLOCKED":              {"WAITING_FOR_APPROVAL": true, "COMPLETED": true, "CANCELLED": true},
	"COMPLETED":            {"IN_PROGRESS": true},
	"SKIPPED":              {"IN_PROGRESS": true},
}

func validWorkflowStepTransition(from, to string) bool { return workflowStepTransitions[from][to] }

type RuntimeWorkflow struct {
	ID                    string               `json:"id"`
	OrderID               string               `json:"order_id"`
	OrderNumber           string               `json:"order_number"`
	CustomerUserID        string               `json:"customer_user_id"`
	CustomerName          string               `json:"customer_name"`
	TemplateGroupCode     string               `json:"template_group_code"`
	TemplateVersionNumber int                  `json:"template_version_number"`
	WorkflowName          string               `json:"workflow_name"`
	Status                string               `json:"status"`
	StartedAt             time.Time            `json:"started_at"`
	EstimatedEndAt        *time.Time           `json:"estimated_end_at,omitempty"`
	CurrentStepInstanceID *string              `json:"current_step_instance_id,omitempty"`
	Steps                 []RuntimeStep        `json:"steps"`
	ActionItems           []RuntimeAction      `json:"action_items"`
	Discrepancies         []RuntimeDiscrepancy `json:"discrepancies,omitempty"`
	Proformas             []Proforma           `json:"proformas,omitempty"`
	AuditSummary          []RuntimeAudit       `json:"audit_summary,omitempty"`
	ViewMode              string               `json:"view_mode"`
	ScopeType             string               `json:"scope_type"`
	ScopeID               string               `json:"scope_id"`
	ParentWorkflowID      *string              `json:"parent_workflow_instance_id,omitempty"`
}
type RuntimeAudit struct {
	Actor      string    `json:"actor"`
	ActionCode string    `json:"action_code"`
	EntityType string    `json:"entity_type"`
	CreatedAt  time.Time `json:"created_at"`
}
type RuntimeStep struct {
	ID                     string         `json:"id"`
	StepCode               string         `json:"step_code"`
	InternalTitleFA        string         `json:"internal_title_fa,omitempty"`
	CustomerTitleFA        string         `json:"customer_title_fa"`
	InternalDescriptionFA  string         `json:"internal_description_fa,omitempty"`
	CustomerDescriptionFA  string         `json:"customer_description_fa,omitempty"`
	SequenceNumber         int            `json:"sequence_number"`
	Status                 string         `json:"status"`
	CustomerStatus         string         `json:"customer_status"`
	ResponsibleRoleID      *int64         `json:"responsible_role_id,omitempty"`
	ResponsibleRoleName    string         `json:"responsible_role_name,omitempty"`
	AssignedUserID         *string        `json:"assigned_user_id,omitempty"`
	RequiredPermissionCode string         `json:"required_permission_code,omitempty"`
	RequiresApproval       bool           `json:"requires_approval"`
	ApprovalRoleID         *int64         `json:"approval_role_id,omitempty"`
	IsOptional             bool           `json:"is_optional"`
	IsSkippable            bool           `json:"is_skippable"`
	CustomerVisible        bool           `json:"customer_visible"`
	EstimatedStartAt       *time.Time     `json:"estimated_start_at,omitempty"`
	EstimatedEndAt         *time.Time     `json:"estimated_end_at,omitempty"`
	ActualStartAt          *time.Time     `json:"actual_start_at,omitempty"`
	ActualEndAt            *time.Time     `json:"actual_end_at,omitempty"`
	IsOverdue              bool           `json:"is_overdue"`
	DelayHours             int            `json:"delay_hours"`
	RejectionReason        *string        `json:"rejection_reason,omitempty"`
	Fields                 []RuntimeField `json:"fields"`
	OpenActionCount        int            `json:"open_action_count"`
	HasDiscrepancy         bool           `json:"has_discrepancy"`
	IterationNumber        int            `json:"iteration_number"`
	PathState              string         `json:"path_state"`
}
type RuntimeField struct {
	ID                int64           `json:"id"`
	FieldKey          string          `json:"field_key"`
	LabelFA           string          `json:"label_fa"`
	DescriptionFA     string          `json:"description_fa"`
	FieldType         string          `json:"field_type"`
	IsRequired        bool            `json:"is_required"`
	IsCustomerVisible bool            `json:"is_customer_visible"`
	IsSalesVisible    bool            `json:"is_sales_visible"`
	IsInternalCost    bool            `json:"is_internal_cost,omitempty"`
	UnitCode          *string         `json:"unit_code,omitempty"`
	CurrencyCode      *string         `json:"currency_code,omitempty"`
	PlaceholderFA     *string         `json:"placeholder_fa,omitempty"`
	DefaultValue      json.RawMessage `json:"default_value,omitempty"`
	OptionsJSON       json.RawMessage `json:"options_json,omitempty"`
	ValidationJSON    json.RawMessage `json:"validation_json,omitempty"`
	Value             json.RawMessage `json:"value,omitempty"`
	SortOrder         int             `json:"sort_order"`
}
type RuntimeAction struct {
	ID                string     `json:"id"`
	StepInstanceID    *string    `json:"step_instance_id,omitempty"`
	TitleFA           string     `json:"title_fa"`
	Status            string     `json:"status"`
	Priority          string     `json:"priority"`
	DueAt             *time.Time `json:"due_at,omitempty"`
	IsBlocking        bool       `json:"is_blocking"`
	SourceTriggerType string     `json:"source_trigger_type"`
}
type RuntimeDiscrepancy struct {
	ID                   string    `json:"id"`
	SourceStepInstanceID *string   `json:"source_step_instance_id,omitempty"`
	TargetStepInstanceID *string   `json:"target_step_instance_id,omitempty"`
	MetricKey            string    `json:"metric_key"`
	ExpectedValue        *float64  `json:"expected_value,omitempty"`
	ActualValue          *float64  `json:"actual_value,omitempty"`
	DifferenceValue      *float64  `json:"difference_value,omitempty"`
	DifferencePercentage *float64  `json:"difference_percentage,omitempty"`
	UnitCode             string    `json:"unit_code"`
	Severity             string    `json:"severity"`
	IsBlocking           bool      `json:"is_blocking"`
	Status               string    `json:"status"`
	ResolutionNote       *string   `json:"resolution_note,omitempty"`
	ReportedAt           time.Time `json:"reported_at"`
}
type CreateDiscrepancyPayload struct {
	TargetStepInstanceID string   `json:"target_step_instance_id"`
	MetricKey            string   `json:"metric_key"`
	ExpectedValue        *float64 `json:"expected_value"`
	ActualValue          *float64 `json:"actual_value"`
	UnitCode             string   `json:"unit_code"`
	Severity             string   `json:"severity"`
	IsBlocking           bool     `json:"is_blocking"`
	Explanation          string   `json:"explanation"`
}
type StepValuesPayload struct {
	Values     map[string]json.RawMessage `json:"values"`
	Reason     string                     `json:"reason,omitempty"`
	ResultCode string                     `json:"result_code,omitempty"`
}
type StepOperationPayload struct {
	Reason         string `json:"reason"`
	AssignedUserID string `json:"assigned_user_id,omitempty"`
	Override       bool   `json:"override,omitempty"`
}

type WorkflowDashboardSummary struct {
	ActiveWorkflows       int `json:"active_workflows"`
	WaitingSteps          int `json:"waiting_steps"`
	ApprovalSteps         int `json:"approval_steps"`
	OverdueSteps          int `json:"overdue_steps"`
	OpenDiscrepancies     int `json:"open_discrepancies"`
	CriticalDiscrepancies int `json:"critical_discrepancies"`
	BlockingTasks         int `json:"blocking_tasks"`
	UnassignedSteps       int `json:"unassigned_steps"`
	DueSoonWorkflows      int `json:"due_soon_workflows"`
}

func (s *OperationsService) WorkflowDashboard(ctx context.Context, actor string) (WorkflowDashboardSummary, error) {
	var out WorkflowDashboardSummary
	viewAll := s.HasPermission(ctx, actor, "workflow_instances.view_all")
	err := s.db.QueryRowContext(ctx, `WITH accessible AS (SELECT wi.id FROM workflow_instances wi WHERE $2 OR EXISTS(SELECT 1 FROM workflow_step_instances si WHERE si.workflow_instance_id=wi.id AND (si.assigned_user_id=$1 OR (si.assigned_user_id IS NULL AND EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$1 AND ur.role_id=si.responsible_role_id))))) SELECT
	(SELECT COUNT(*) FROM workflow_instances wi JOIN accessible a ON a.id=wi.id WHERE wi.status='IN_PROGRESS'),
	(SELECT COUNT(*) FROM workflow_step_instances si JOIN accessible a ON a.id=si.workflow_instance_id WHERE si.status IN ('WAITING_FOR_ASSIGNEE','NEEDS_CORRECTION')),
	(SELECT COUNT(*) FROM workflow_step_instances si JOIN accessible a ON a.id=si.workflow_instance_id WHERE si.status='WAITING_FOR_APPROVAL'),
	(SELECT COUNT(*) FROM workflow_step_instances si JOIN accessible a ON a.id=si.workflow_instance_id WHERE si.estimated_end_at<NOW() AND si.status NOT IN ('COMPLETED','SKIPPED','CANCELLED')),
	(SELECT COUNT(*) FROM workflow_discrepancies d JOIN accessible a ON a.id=d.workflow_instance_id WHERE d.status NOT IN ('ACCEPTED','RESOLVED','CANCELLED')),
	(SELECT COUNT(*) FROM workflow_discrepancies d JOIN accessible a ON a.id=d.workflow_instance_id WHERE d.severity='CRITICAL' AND d.status NOT IN ('ACCEPTED','RESOLVED','CANCELLED')),
	(SELECT COUNT(*) FROM action_items i JOIN accessible a ON a.id=i.workflow_instance_id WHERE i.is_blocking AND i.status NOT IN ('COMPLETED','CANCELLED')),
	(SELECT COUNT(*) FROM workflow_step_instances si JOIN accessible a ON a.id=si.workflow_instance_id WHERE si.status='WAITING_FOR_ASSIGNEE' AND si.assigned_user_id IS NULL AND si.responsible_role_id IS NULL),
	(SELECT COUNT(*) FROM workflow_instances wi JOIN accessible a ON a.id=wi.id WHERE wi.status='IN_PROGRESS' AND wi.estimated_end_at BETWEEN NOW() AND NOW()+INTERVAL '48 hours')`, actor, viewAll).Scan(&out.ActiveWorkflows, &out.WaitingSteps, &out.ApprovalSteps, &out.OverdueSteps, &out.OpenDiscrepancies, &out.CriticalDiscrepancies, &out.BlockingTasks, &out.UnassignedSteps, &out.DueSoonWorkflows)
	return out, err
}

func customerStatus(status string) string {
	switch status {
	case "NOT_STARTED", "WAITING_FOR_ASSIGNEE":
		return "در انتظار شروع"
	case "IN_PROGRESS":
		return "در حال انجام"
	case "SUBMITTED", "WAITING_FOR_APPROVAL":
		return "در حال بررسی"
	case "NEEDS_CORRECTION", "HAS_MISMATCH", "BLOCKED":
		return "در حال تکمیل"
	case "COMPLETED", "APPROVED":
		return "تکمیل شده"
	case "CANCELLED":
		return "لغو شده"
	case "SKIPPED":
		return "عبور شده"
	}
	return status
}

func (s *OperationsService) snapshotWorkflowTx(ctx context.Context, tx *sql.Tx, workflowID string, templateID int64, actor string, excluded []string) error {
	var transitionCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_step_transitions WHERE workflow_template_id=$1`, templateID).Scan(&transitionCount); err != nil {
		return err
	}
	excludedSet := map[string]bool{}
	for _, c := range excluded {
		excludedSet[strings.ToUpper(strings.TrimSpace(c))] = true
	}
	normalizedExcluded := make([]string, 0, len(excludedSet))
	for code := range excludedSet {
		if code != "" {
			normalizedExcluded = append(normalizedExcluded, code)
		}
	}
	if len(normalizedExcluded) > 0 {
		var valid int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_template_steps WHERE workflow_template_id=$1 AND is_active AND is_optional AND step_code=ANY($2)`, templateID, pq.Array(normalizedExcluded)).Scan(&valid); err != nil {
			return err
		}
		if valid != len(normalizedExcluded) {
			return errors.New("only optional steps may be excluded")
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT s.id,s.step_code,s.internal_title_fa,s.internal_description_fa,s.customer_title_fa,s.customer_description_fa,s.sequence_number,s.responsible_role_id,s.required_permission_code,s.customer_visible,s.requires_approval,s.approval_role_id,s.is_optional,s.is_skippable,s.default_duration_hours,s.starts_automatically,s.domain_event_code,s.is_entry FROM workflow_template_steps s WHERE s.workflow_template_id=$1 AND s.is_active ORDER BY s.sequence_number`, templateID)
	if err != nil {
		return err
	}
	type snapshotStep struct {
		templateStepID                                              int64
		code, it, idsc, ct, cdsc, perm                              string
		domainEvent                                                 sql.NullString
		seq, duration                                               int
		role, approval                                              sql.NullInt64
		visible, requiresApproval, optional, skippable, auto, entry bool
	}
	steps := []snapshotStep{}
	for rows.Next() {
		var step snapshotStep
		if err := rows.Scan(&step.templateStepID, &step.code, &step.it, &step.idsc, &step.ct, &step.cdsc, &step.seq, &step.role, &step.perm, &step.visible, &step.requiresApproval, &step.approval, &step.optional, &step.skippable, &step.duration, &step.auto, &step.domainEvent, &step.entry); err != nil {
			rows.Close()
			return err
		}
		if !excludedSet[step.code] {
			steps = append(steps, step)
		}
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	start := time.Now()
	cursor := start
	var firstID string
	var workflowEnd time.Time
	for _, step := range steps {
		templateStepID, code, it, idsc, ct, cdsc, perm := step.templateStepID, step.code, step.it, step.idsc, step.ct, step.cdsc, step.perm
		seq, duration := step.seq, step.duration
		role, approval := step.role, step.approval
		visible, requiresApproval, optional, skippable, auto := step.visible, step.requiresApproval, step.optional, step.skippable, step.auto
		estimatedStart := cursor
		cursor = cursor.Add(time.Duration(duration) * time.Hour)
		estimatedEnd := cursor
		workflowEnd = estimatedEnd
		status := "NOT_STARTED"
		var assigned any = nil
		var actualStart any = nil
		if (transitionCount == 0 && firstID == "") || (transitionCount > 0 && step.entry) {
			status = "WAITING_FOR_ASSIGNEE"
			if role.Valid {
				var actorHasRole bool
				if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id=$1 AND role_id=$2)`, actor, role.Int64).Scan(&actorHasRole); err != nil {
					return err
				}
				if actorHasRole {
					assigned = actor
				}
			}
			if auto {
				status = "IN_PROGRESS"
				actualStart = start
			}
		}
		var stepID string
		err = tx.QueryRowContext(ctx, `INSERT INTO workflow_step_instances(workflow_instance_id,workflow_template_step_id,template_step_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,status,assigned_role_id,responsible_role_id,assigned_user_id,required_permission_code,requires_approval,approval_role_id,is_optional,is_skippable,starts_automatically,customer_visible,estimated_start_at,estimated_end_at,actual_start_at,customer_status_text,domain_event_code) VALUES($1,$2,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23) RETURNING id`, workflowID, templateStepID, code, it, idsc, ct, cdsc, seq, status, nullableInt(role), assigned, perm, requiresApproval, nullableInt(approval), optional, skippable, auto, visible, estimatedStart, estimatedEnd, actualStart, customerStatus(status), step.domainEvent).Scan(&stepID)
		if err != nil {
			return err
		}
		if (transitionCount == 0 && firstID == "") || (transitionCount > 0 && step.entry) {
			firstID = stepID
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO workflow_instance_field_definitions(workflow_instance_id,workflow_step_instance_id,source_field_definition_id,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,is_internal_cost,unit_code,currency_code,placeholder_fa,default_value,options_json,validation_json,sort_order,handoff_metric_key,handoff_direction) SELECT $1,$2,id,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,is_internal_cost,unit_code,currency_code,placeholder_fa,default_value,options_json,validation_json,sort_order,handoff_metric_key,handoff_direction FROM workflow_step_field_definitions WHERE workflow_template_step_id=$3`, workflowID, stepID, templateStepID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO workflow_instance_step_task_templates(workflow_instance_id,workflow_step_instance_id,source_task_template_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_step_completion) SELECT $1,$2,id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_step_completion FROM workflow_step_task_templates WHERE workflow_template_step_id=$3`, workflowID, stepID, templateStepID)
		if err != nil {
			return err
		}
	}
	if firstID == "" {
		return errors.New("published template has no active steps")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_instance_handoff_metrics(workflow_instance_id,source_metric_definition_id,metric_key,label_fa,unit_code,absolute_tolerance,percentage_tolerance,blocking_on_mismatch) SELECT $1,id,metric_key,label_fa,unit_code,absolute_tolerance,percentage_tolerance,blocking_on_mismatch FROM workflow_handoff_metric_definitions WHERE workflow_template_id=$2`, workflowID, templateID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_instance_task_templates(workflow_instance_id,source_task_template_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_workflow_progress) SELECT $1,id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_workflow_progress FROM workflow_task_templates WHERE workflow_template_id=$2`, workflowID, templateID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_instance_step_transitions(workflow_instance_id,source_template_step_id,target_template_step_id,source_step_code,target_step_code,transition_code,label_fa,transition_type,result_code,is_default,requires_permission_code,requires_reason,sort_order) SELECT $1,t.source_step_id,t.target_step_id,s.step_code,d.step_code,t.transition_code,t.label_fa,t.transition_type,t.result_code,t.is_default,t.requires_permission_code,t.requires_reason,t.sort_order FROM workflow_step_transitions t JOIN workflow_template_steps s ON s.id=t.source_step_id JOIN workflow_template_steps d ON d.id=t.target_step_id WHERE t.workflow_template_id=$2`, workflowID, templateID)
	if err != nil {
		return err
	}
	if transitionCount > 0 {
		_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET path_state=CASE WHEN id=$2 THEN 'INCLUDED' ELSE 'NOT_SELECTED' END WHERE workflow_instance_id=$1`, workflowID, firstID)
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_instances SET current_step_instance_id=$2,estimated_end_at=$3 WHERE id=$1`, workflowID, firstID, workflowEnd)
	if err != nil {
		return err
	}
	if err = s.createMainStepActionTx(ctx, tx, workflowID, firstID); err != nil {
		return err
	}
	if err = s.runWorkflowTriggersTx(ctx, tx, workflowID, "ON_WORKFLOW_START"); err != nil {
		return err
	}
	return s.runStepTriggersTx(ctx, tx, workflowID, firstID, "ON_STEP_OPEN")
}

func nullableInt(v sql.NullInt64) any {
	if v.Valid {
		return v.Int64
	}
	return nil
}

func (s *OperationsService) createMainStepActionTx(ctx context.Context, tx *sql.Tx, workflowID, stepID string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,assigned_user_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT wi.id,si.id,wi.order_id,wi.customer_user_id,si.internal_title_fa,si.internal_description_fa,'OPEN','NORMAL',si.responsible_role_id,si.assigned_user_id,si.required_permission_code,si.estimated_end_at,'main:'||si.id,'MAIN_STEP' FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.id=$1 AND wi.id=$2 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, stepID, workflowID)
	return err
}
func (s *OperationsService) runStepTriggersTx(ctx context.Context, tx *sql.Tx, workflowID, stepID, event string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,required_permission_code,due_at,deduplication_key,source_trigger_type,is_blocking) SELECT wi.id,t.workflow_step_instance_id,wi.order_id,wi.customer_user_id,t.title_fa,t.description_fa,'OPEN',t.priority,t.assigned_role_id,t.required_permission_code,CASE WHEN t.due_offset_hours IS NULL THEN NULL ELSE NOW()+MAKE_INTERVAL(hours=>t.due_offset_hours) END,'step-task:'||t.id||':'||$3,t.trigger_type,t.blocks_step_completion FROM workflow_instance_step_task_templates t JOIN workflow_instances wi ON wi.id=t.workflow_instance_id WHERE t.workflow_instance_id=$1 AND t.workflow_step_instance_id=$2 AND t.trigger_type=$3 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, workflowID, stepID, event)
	return err
}
func (s *OperationsService) runWorkflowTriggersTx(ctx context.Context, tx *sql.Tx, workflowID, event string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,required_permission_code,due_at,deduplication_key,source_trigger_type,is_blocking) SELECT wi.id,wi.order_id,wi.customer_user_id,t.title_fa,t.description_fa,'OPEN',t.priority,t.assigned_role_id,t.required_permission_code,CASE WHEN t.due_offset_hours IS NULL THEN NULL ELSE NOW()+MAKE_INTERVAL(hours=>t.due_offset_hours) END,'workflow-task:'||t.id||':'||$2,t.trigger_type,t.blocks_workflow_progress FROM workflow_instance_task_templates t JOIN workflow_instances wi ON wi.id=t.workflow_instance_id WHERE t.workflow_instance_id=$1 AND t.trigger_type=$2 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, workflowID, event)
	return err
}

func (s *OperationsService) GetWorkflowRuntime(ctx context.Context, actor, workflowID string) (RuntimeWorkflow, error) {
	var w RuntimeWorkflow
	var est sql.NullTime
	var current sql.NullString
	var parent sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT wi.id,wi.order_id,o.order_number,wi.customer_user_id,COALESCE(NULLIF(TRIM(CONCAT_WS(' ',u.first_name,u.last_name)),''),u.phone_normalized),COALESCE(wi.template_group_code,wt.template_group_code),COALESCE(wi.template_version_number,wt.version_number),wt.name_fa,wi.status,wi.started_at,wi.estimated_end_at,wi.current_step_instance_id,wi.scope_type,wi.scope_id,wi.parent_workflow_instance_id FROM workflow_instances wi JOIN orders o ON o.id=wi.order_id JOIN users u ON u.id=wi.customer_user_id JOIN workflow_templates wt ON wt.id=wi.workflow_template_id WHERE wi.id=$1`, workflowID).Scan(&w.ID, &w.OrderID, &w.OrderNumber, &w.CustomerUserID, &w.CustomerName, &w.TemplateGroupCode, &w.TemplateVersionNumber, &w.WorkflowName, &w.Status, &w.StartedAt, &est, &current, &w.ScopeType, &w.ScopeID, &parent)
	if err != nil {
		return w, err
	}
	if est.Valid {
		v := est.Time
		w.EstimatedEndAt = &v
	}
	if current.Valid {
		v := current.String
		w.CurrentStepInstanceID = &v
	}
	if parent.Valid {
		v := parent.String
		w.ParentWorkflowID = &v
	}
	isCustomer := actor == w.CustomerUserID && s.HasPermission(ctx, actor, "customer_portal.workflow.view_own")
	canAll := s.HasPermission(ctx, actor, "workflow_instances.view_all")
	canAssigned := false
	if s.HasPermission(ctx, actor, "workflow_instances.view_assigned") {
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_step_instances si WHERE si.workflow_instance_id=$1 AND (si.assigned_user_id=$2 OR (si.assigned_user_id IS NULL AND EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$2 AND ur.role_id=si.responsible_role_id))))`, workflowID, actor).Scan(&canAssigned)
	}
	if !isCustomer && !canAll && !canAssigned {
		return w, ErrForbidden
	}
	if isCustomer {
		w.ViewMode = "CUSTOMER"
	} else if s.HasPermission(ctx, actor, "workflow_fields.view_sales") && !s.HasPermission(ctx, actor, "workflow_fields.view_internal") {
		w.ViewMode = "SALES"
	} else {
		w.ViewMode = "INTERNAL"
	}
	rows, err := s.db.QueryContext(ctx, `SELECT si.id,COALESCE(si.step_code,ts.step_code),COALESCE(si.internal_title_fa,ts.internal_title_fa),COALESCE(si.internal_description_fa,''),COALESCE(si.customer_title_fa,ts.customer_title_fa),COALESCE(si.customer_description_fa,''),COALESCE(si.sequence_number,ts.sequence_number),si.status,si.responsible_role_id,COALESCE(r.name_fa,''),si.assigned_user_id,COALESCE(si.required_permission_code,''),si.requires_approval,si.approval_role_id,si.is_optional,si.is_skippable,si.customer_visible,si.estimated_start_at,si.estimated_end_at,si.actual_start_at,si.actual_end_at,si.rejection_reason,(SELECT COUNT(*) FROM action_items a WHERE a.workflow_step_instance_id=si.id AND a.status NOT IN ('COMPLETED','CANCELLED')),(SELECT EXISTS(SELECT 1 FROM workflow_discrepancies d WHERE (d.source_step_instance_id=si.id OR d.target_step_instance_id=si.id) AND d.status NOT IN ('RESOLVED','CANCELLED'))),si.iteration_number,si.path_state FROM workflow_step_instances si LEFT JOIN workflow_template_steps ts ON ts.id=si.workflow_template_step_id LEFT JOIN roles r ON r.id=si.responsible_role_id WHERE si.workflow_instance_id=$1 ORDER BY COALESCE(si.sequence_number,ts.sequence_number),si.iteration_number`, workflowID)
	if err != nil {
		return w, err
	}
	defer rows.Close()
	now := time.Now()
	for rows.Next() {
		var st RuntimeStep
		var role, approval sql.NullInt64
		var assigned, rejection sql.NullString
		var es, ee, as, ae sql.NullTime
		if err := rows.Scan(&st.ID, &st.StepCode, &st.InternalTitleFA, &st.InternalDescriptionFA, &st.CustomerTitleFA, &st.CustomerDescriptionFA, &st.SequenceNumber, &st.Status, &role, &st.ResponsibleRoleName, &assigned, &st.RequiredPermissionCode, &st.RequiresApproval, &approval, &st.IsOptional, &st.IsSkippable, &st.CustomerVisible, &es, &ee, &as, &ae, &rejection, &st.OpenActionCount, &st.HasDiscrepancy, &st.IterationNumber, &st.PathState); err != nil {
			return w, err
		}
		if isCustomer && (!st.CustomerVisible || st.Status == "SKIPPED" || st.PathState != "INCLUDED") {
			continue
		}
		if role.Valid {
			v := role.Int64
			st.ResponsibleRoleID = &v
		}
		if approval.Valid {
			v := approval.Int64
			st.ApprovalRoleID = &v
		}
		if assigned.Valid {
			v := assigned.String
			st.AssignedUserID = &v
		}
		if rejection.Valid && !isCustomer {
			v := rejection.String
			st.RejectionReason = &v
		}
		if es.Valid {
			v := es.Time
			st.EstimatedStartAt = &v
		}
		if ee.Valid {
			v := ee.Time
			st.EstimatedEndAt = &v
			if now.After(v) && st.Status != "COMPLETED" && st.Status != "SKIPPED" && st.Status != "CANCELLED" {
				st.IsOverdue = true
				st.DelayHours = int(now.Sub(v).Hours())
			}
		}
		if as.Valid {
			v := as.Time
			st.ActualStartAt = &v
		}
		if ae.Valid {
			v := ae.Time
			st.ActualEndAt = &v
		}
		st.CustomerStatus = customerStatus(st.Status)
		st.Fields, _ = s.runtimeFields(ctx, actor, st.ID, w.ViewMode)
		if isCustomer {
			st.InternalTitleFA = ""
			st.InternalDescriptionFA = ""
			st.ResponsibleRoleID = nil
			st.ResponsibleRoleName = ""
			st.AssignedUserID = nil
			st.RequiredPermissionCode = ""
			st.RejectionReason = nil
			st.HasDiscrepancy = false
			st.OpenActionCount = 0
		}
		w.Steps = append(w.Steps, st)
	}
	if !isCustomer {
		w.ActionItems, _ = s.runtimeActions(ctx, actor, workflowID)
		if s.HasPermission(ctx, actor, "workflow_discrepancies.view_all") || s.HasPermission(ctx, actor, "workflow_discrepancies.view_assigned") {
			w.Discrepancies, _ = s.ListWorkflowDiscrepancies(ctx, actor, workflowID)
		}
		if s.HasPermission(ctx, actor, "proformas.view_all") || s.HasPermission(ctx, actor, "proformas.view_own") {
			rows, _ := s.db.QueryContext(ctx, `SELECT id,proforma_number,order_id,status,currency,subtotal::text,discount_amount::text,total_amount::text,notes,issued_at,created_at FROM proformas WHERE order_id=$1 ORDER BY created_at DESC`, w.OrderID)
			if rows != nil {
				for rows.Next() {
					var p Proforma
					if rows.Scan(&p.ID, &p.ProformaNumber, &p.OrderID, &p.Status, &p.Currency, &p.Subtotal, &p.DiscountAmount, &p.TotalAmount, &p.Notes, &p.IssuedAt, &p.CreatedAt) == nil {
						w.Proformas = append(w.Proformas, p)
					}
				}
				rows.Close()
			}
		}
		if s.HasPermission(ctx, actor, "audit.view") {
			rows, _ := s.db.QueryContext(ctx, `SELECT COALESCE(NULLIF(TRIM(CONCAT_WS(' ',u.first_name,u.last_name)),''),u.phone_normalized,'سیستم'),a.action_code,a.entity_type,a.created_at FROM audit_logs a LEFT JOIN users u ON u.id=a.actor_user_id WHERE (a.entity_type='workflow_instance' AND a.entity_id=$1) OR (a.entity_type='workflow_step_instance' AND a.entity_id IN (SELECT id::text FROM workflow_step_instances WHERE workflow_instance_id=$1)) ORDER BY a.created_at DESC LIMIT 30`, workflowID)
			if rows != nil {
				for rows.Next() {
					var item RuntimeAudit
					if rows.Scan(&item.Actor, &item.ActionCode, &item.EntityType, &item.CreatedAt) == nil {
						w.AuditSummary = append(w.AuditSummary, item)
					}
				}
				rows.Close()
			}
		}
	}
	return w, rows.Err()
}

func (s *OperationsService) runtimeFields(ctx context.Context, actor, stepID, mode string) ([]RuntimeField, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT f.id,f.field_key,f.label_fa,f.description_fa,f.field_type,f.is_required,f.is_customer_visible,f.is_sales_visible,f.is_internal_cost,f.unit_code,f.currency_code,f.placeholder_fa,f.default_value,f.options_json,f.validation_json,v.value_json,f.sort_order FROM workflow_instance_field_definitions f LEFT JOIN workflow_step_field_values v ON v.field_definition_id=f.id AND v.workflow_step_instance_id=f.workflow_step_instance_id WHERE f.workflow_step_instance_id=$1 ORDER BY f.sort_order,f.id`, stepID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RuntimeField{}
	canCost := s.HasPermission(ctx, actor, "finance.internal_costs.view")
	for rows.Next() {
		var f RuntimeField
		var unit, currency, placeholder sql.NullString
		var def, opts, val, value []byte
		if err := rows.Scan(&f.ID, &f.FieldKey, &f.LabelFA, &f.DescriptionFA, &f.FieldType, &f.IsRequired, &f.IsCustomerVisible, &f.IsSalesVisible, &f.IsInternalCost, &unit, &currency, &placeholder, &def, &opts, &val, &value, &f.SortOrder); err != nil {
			return nil, err
		}
		if mode == "CUSTOMER" && !f.IsCustomerVisible {
			continue
		}
		if mode == "SALES" && (!f.IsSalesVisible || (f.IsInternalCost && !canCost)) {
			continue
		}
		if mode == "INTERNAL" && f.IsInternalCost && !canCost {
			continue
		}
		if unit.Valid {
			f.UnitCode = &unit.String
		}
		if currency.Valid {
			f.CurrencyCode = &currency.String
		}
		if placeholder.Valid {
			f.PlaceholderFA = &placeholder.String
		}
		f.DefaultValue = def
		f.OptionsJSON = opts
		f.ValidationJSON = val
		f.Value = value
		if mode != "INTERNAL" {
			f.IsInternalCost = false
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
func (s *OperationsService) runtimeActions(ctx context.Context, actor, workflowID string) ([]RuntimeAction, error) {
	viewAll := s.HasPermission(ctx, actor, "action_items.view_all")
	rows, err := s.db.QueryContext(ctx, `SELECT id,workflow_step_instance_id,title_fa,status,priority,due_at,is_blocking,COALESCE(source_trigger_type,'') FROM action_items WHERE workflow_instance_id=$1 AND status NOT IN ('COMPLETED','CANCELLED') AND COALESCE(source_trigger_type,'') NOT IN ('MAIN_STEP','APPROVAL','CORRECTION') AND ($3 OR assigned_user_id=$2 OR (assigned_user_id IS NULL AND EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$2 AND ur.role_id=action_items.assigned_role_id))) ORDER BY is_blocking DESC,due_at NULLS LAST`, workflowID, actor, viewAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RuntimeAction{}
	for rows.Next() {
		var a RuntimeAction
		var step sql.NullString
		var due sql.NullTime
		if err := rows.Scan(&a.ID, &step, &a.TitleFA, &a.Status, &a.Priority, &due, &a.IsBlocking, &a.SourceTriggerType); err != nil {
			return nil, err
		}
		if step.Valid {
			v := step.String
			a.StepInstanceID = &v
		}
		if due.Valid {
			v := due.Time
			a.DueAt = &v
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *OperationsService) lockStepTx(ctx context.Context, tx *sql.Tx, stepID string) (workflowID, status, perm string, roleID sql.NullInt64, assigned sql.NullString, requiresApproval, skippable bool, err error) {
	err = tx.QueryRowContext(ctx, `SELECT workflow_instance_id,status,COALESCE(required_permission_code,''),responsible_role_id,assigned_user_id,requires_approval,is_skippable FROM workflow_step_instances WHERE id=$1 FOR UPDATE`, stepID).Scan(&workflowID, &status, &perm, &roleID, &assigned, &requiresApproval, &skippable)
	return
}
func (s *OperationsService) canOperateStep(ctx context.Context, actor, perm string, roleID sql.NullInt64, assigned sql.NullString) bool {
	if assigned.Valid && assigned.String == actor {
		return perm == "" || s.HasPermission(ctx, actor, perm)
	}
	if roleID.Valid && s.userHasRole(ctx, actor, s.roleCode(ctx, roleID.Int64)) {
		return perm == "" || s.HasPermission(ctx, actor, perm)
	}
	return false
}
func (s *OperationsService) roleCode(ctx context.Context, id int64) string {
	var code string
	_ = s.db.QueryRowContext(ctx, `SELECT code FROM roles WHERE id=$1`, id).Scan(&code)
	return code
}
func (s *OperationsService) requireStepActor(ctx context.Context, actor, perm string, role sql.NullInt64, assigned sql.NullString, reason string) (bool, error) {
	if s.canOperateStep(ctx, actor, perm, role, assigned) {
		return false, nil
	}
	if s.HasPermission(ctx, actor, "workflow_steps.override") {
		if strings.TrimSpace(reason) == "" {
			return false, errors.New("override reason is required")
		}
		return true, nil
	}
	return false, ErrForbidden
}

func (s *OperationsService) StartWorkflowStep(ctx context.Context, actor, stepID, reason string) error {
	if !s.HasPermission(ctx, actor, "workflow_steps.submit") && !s.HasPermission(ctx, actor, "workflow_steps.override") {
		return ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	workflowID, status, perm, role, assigned, _, _, err := s.lockStepTx(ctx, tx, stepID)
	if err != nil {
		return err
	}
	if !validWorkflowStepTransition(status, "IN_PROGRESS") {
		return ErrInvalidTransition
	}
	override, err := s.requireStepActor(ctx, actor, perm, role, assigned, reason)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='IN_PROGRESS',assigned_user_id=COALESCE(assigned_user_id,$2),actual_start_at=COALESCE(actual_start_at,NOW()),customer_status_text='در حال انجام',updated_at=NOW() WHERE id=$1`, stepID, actor)
	if err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='IN_PROGRESS',updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type IN ('MAIN_STEP','CORRECTION') AND status='OPEN'`, stepID)
	if err = s.runStepTriggersTx(ctx, tx, workflowID, stepID, "ON_STEP_START"); err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, chooseAudit("workflow_steps.start", override), "workflow_step_instance", stepID, nil, map[string]any{"reason": reason})
	return tx.Commit()
}
func chooseAudit(base string, override bool) string {
	if override {
		return base + ".override"
	}
	return base
}

func (s *OperationsService) SaveWorkflowStepDraft(ctx context.Context, actor, stepID string, p StepValuesPayload) error {
	if !s.HasPermission(ctx, actor, "workflow_steps.save_draft") {
		return ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, status, perm, role, assigned, _, _, err := s.lockStepTx(ctx, tx, stepID)
	if err != nil {
		return err
	}
	if status != "IN_PROGRESS" {
		return ErrInvalidTransition
	}
	override, err := s.requireStepActor(ctx, actor, perm, role, assigned, p.Reason)
	if err != nil {
		return err
	}
	if err = s.saveValuesTx(ctx, tx, actor, stepID, p.Values, false); err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, chooseAudit("workflow_steps.draft", override), "workflow_step_instance", stepID, nil, map[string]any{"reason": p.Reason})
	return tx.Commit()
}
func (s *OperationsService) saveValuesTx(ctx context.Context, tx *sql.Tx, actor, stepID string, values map[string]json.RawMessage, requireAll bool) error {
	rows, err := tx.QueryContext(ctx, `SELECT id,field_key,field_type,is_required,options_json,validation_json,unit_code,currency_code FROM workflow_instance_field_definitions WHERE workflow_step_instance_id=$1 ORDER BY sort_order`, stepID)
	if err != nil {
		return err
	}
	type fieldDefinition struct {
		id             int64
		key, kind      string
		required       bool
		opts, val      []byte
		unit, currency sql.NullString
	}
	definitions := []fieldDefinition{}
	for rows.Next() {
		var definition fieldDefinition
		if err := rows.Scan(&definition.id, &definition.key, &definition.kind, &definition.required, &definition.opts, &definition.val, &definition.unit, &definition.currency); err != nil {
			rows.Close()
			return err
		}
		definitions = append(definitions, definition)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, definition := range definitions {
		id, key, kind, required, opts, val, unit, currency := definition.id, definition.key, definition.kind, definition.required, definition.opts, definition.val, definition.unit, definition.currency
		raw, ok := values[key]
		if !ok {
			if requireAll && required {
				var exists bool
				_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_step_field_values WHERE workflow_step_instance_id=$1 AND field_key=$2 AND value_json IS NOT NULL)`, stepID, key).Scan(&exists)
				if !exists {
					return fmt.Errorf("%w: %s is required", ErrValidation, key)
				}
			}
			continue
		}
		if err = validateRuntimeValue(kind, raw, opts, val, required, unit.String, currency.String); err != nil {
			return fmt.Errorf("%w: %s: %v", ErrValidation, key, err)
		}
		if kind == "FILE" || kind == "IMAGE" || kind == "SIGNATURE" {
			var reference map[string]any
			if err = json.Unmarshal(raw, &reference); err != nil {
				return fmt.Errorf("%w: %s: invalid file reference", ErrValidation, key)
			}
			if err = validateFileReferenceTx(ctx, tx, stepID, id, reference); err != nil {
				return fmt.Errorf("%w: %s: %v", ErrValidation, key, err)
			}
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO workflow_step_field_values(workflow_step_instance_id,field_definition_id,field_key,field_type,value_json,entered_by_user_id,updated_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$6) ON CONFLICT(workflow_step_instance_id,field_key) DO UPDATE SET value_json=EXCLUDED.value_json,updated_by_user_id=EXCLUDED.updated_by_user_id,updated_at=NOW()`, stepID, id, key, kind, []byte(raw), actor)
		if err != nil {
			return err
		}
		seen[key] = true
	}
	for key := range values {
		if !seen[key] {
			return fmt.Errorf("%w: unknown field %s", ErrValidation, key)
		}
	}
	return nil
}
func validateRuntimeValue(kind string, raw, options, validation []byte, required bool, expectedUnit, expectedCurrency string) error {
	if len(raw) == 0 || string(raw) == "null" {
		if required {
			return errors.New("value is required")
		}
		return nil
	}
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return errors.New("invalid json")
	}
	empty := false
	if x, ok := v.(string); ok {
		empty = strings.TrimSpace(x) == ""
	}
	if required && empty {
		return errors.New("value is required")
	}
	rules := map[string]any{}
	_ = json.Unmarshal(validation, &rules)
	number := func() (float64, bool) {
		switch x := v.(type) {
		case float64:
			return x, true
		case map[string]any:
			for _, k := range []string{"value", "amount"} {
				if n, ok := x[k].(float64); ok {
					return n, true
				}
			}
		}
		return 0, false
	}
	switch kind {
	case "SHORT_TEXT", "LONG_TEXT", "ADDRESS":
		if _, ok := v.(string); !ok {
			return errors.New("text required")
		}
	case "INTEGER":
		n, ok := v.(float64)
		if !ok || math.Trunc(n) != n {
			return errors.New("integer required")
		}
	case "DECIMAL":
		if _, ok := v.(float64); !ok {
			return errors.New("numeric value required")
		}
	case "WEIGHT", "AREA", "VOLUME", "QUANTITY":
		object, ok := v.(map[string]any)
		if !ok {
			return errors.New("measurement object required")
		}
		if _, ok = object["value"].(float64); !ok {
			return errors.New("numeric value required")
		}
		unit, ok := object["unit"].(string)
		if !ok || unit == "" || (expectedUnit != "" && unit != expectedUnit) {
			return errors.New("invalid unit")
		}
	case "MONEY":
		object, ok := v.(map[string]any)
		if !ok {
			return errors.New("money object required")
		}
		if _, ok = object["amount"].(float64); !ok {
			return errors.New("numeric value required")
		}
		currency, ok := object["currency"].(string)
		if !ok || len(currency) != 3 || (expectedCurrency != "" && currency != expectedCurrency) {
			return errors.New("invalid currency")
		}
	case "BOOLEAN":
		if _, ok := v.(bool); !ok {
			return errors.New("boolean required")
		}
	case "DATE", "DATETIME", "TIME":
		text, ok := v.(string)
		if !ok {
			return errors.New("date/time string required")
		}
		var e error
		if kind == "DATE" {
			_, e = time.Parse("2006-01-02", text)
		} else if kind == "TIME" {
			_, e = time.Parse("15:04", text)
		} else {
			_, e = time.Parse(time.RFC3339, text)
		}
		if e != nil {
			return errors.New("invalid date/time")
		}
	case "SELECT":
		if !optionAllowed(v, options) {
			return errors.New("invalid option")
		}
	case "MULTI_SELECT":
		arr, ok := v.([]any)
		if !ok {
			return errors.New("array required")
		}
		for _, x := range arr {
			if !optionAllowed(x, options) {
				return errors.New("invalid option")
			}
		}
	case "PHONE":
		text, ok := v.(string)
		if !ok || NormalizePhone(text) == "" {
			return errors.New("invalid phone")
		}
	case "FILE", "IMAGE", "SIGNATURE":
		obj, ok := v.(map[string]any)
		if !ok || obj["fileId"] == nil {
			return errors.New("fileId required")
		}
	}
	if n, ok := number(); ok {
		if min, ok := rules["min"].(float64); ok && n < min {
			return errors.New("below minimum")
		}
		if max, ok := rules["max"].(float64); ok && n > max {
			return errors.New("above maximum")
		}
	}
	if text, ok := v.(string); ok {
		if max, ok := rules["maxLength"].(float64); ok && len([]rune(text)) > int(max) {
			return errors.New("too long")
		}
	}
	return nil
}
func optionAllowed(value any, raw []byte) bool {
	var opts []any
	if json.Unmarshal(raw, &opts) != nil {
		return false
	}
	candidate := fmt.Sprint(value)
	for _, o := range opts {
		switch x := o.(type) {
		case string:
			if x == candidate {
				return true
			}
		case map[string]any:
			if fmt.Sprint(x["value"]) == candidate {
				return true
			}
		}
	}
	return false
}

func (s *OperationsService) SubmitWorkflowStep(ctx context.Context, actor, stepID string, p StepValuesPayload) error {
	if !s.HasPermission(ctx, actor, "workflow_steps.submit") && !s.HasPermission(ctx, actor, "workflow_steps.override") {
		return ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	workflowID, status, perm, role, assigned, approval, _, err := s.lockStepTx(ctx, tx, stepID)
	if err != nil {
		return err
	}
	if status != "IN_PROGRESS" {
		return ErrInvalidTransition
	}
	override, err := s.requireStepActor(ctx, actor, perm, role, assigned, p.Reason)
	if err != nil {
		return err
	}
	if err = s.saveValuesTx(ctx, tx, actor, stepID, p.Values, true); err != nil {
		return err
	}
	var domainEvent sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT domain_event_code FROM workflow_step_instances WHERE id=$1`, stepID).Scan(&domainEvent); err != nil {
		return err
	}
	if domainEvent.Valid {
		var done bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_domain_operation_completions WHERE workflow_step_instance_id=$1 AND domain_event_code=$2)`, stepID, domainEvent.String).Scan(&done); err != nil {
			return err
		}
		if !done {
			return conflict("DOMAIN_OPERATION_REQUIRED", "required domain operation has not been recorded")
		}
	}
	if p.ResultCode != "" {
		p.ResultCode = normalizeCode(p.ResultCode)
		if !allowedWorkflowResult(p.ResultCode) {
			return conflict("INVALID_TRANSITION", "unsupported workflow result")
		}
		_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET result_code=$2 WHERE id=$1`, stepID, p.ResultCode)
		if err != nil {
			return err
		}
	}
	if err = s.runStepTriggersTx(ctx, tx, workflowID, stepID, "ON_STEP_SUBMIT"); err != nil {
		return err
	}
	blocking, err := s.evaluateHandoffsTx(ctx, tx, actor, workflowID, stepID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET submitted_at=NOW(),submitted_by_user_id=$2,updated_at=NOW() WHERE id=$1`, stepID, actor)
	if err != nil {
		return err
	}
	var blockingTasks int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM action_items WHERE workflow_step_instance_id=$1 AND is_blocking AND status NOT IN ('COMPLETED','CANCELLED')`, stepID).Scan(&blockingTasks); err != nil {
		return err
	}
	if blocking {
		_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='HAS_MISMATCH',customer_status_text='در حال تکمیل' WHERE id=$1`, stepID)
		if err != nil {
			return err
		}
		_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='BLOCKED' WHERE workflow_step_instance_id=$1 AND source_trigger_type='MAIN_STEP'`, stepID)
	} else if blockingTasks > 0 {
		if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='BLOCKED',customer_status_text='در حال تکمیل',updated_at=NOW() WHERE id=$1`, stepID); err != nil {
			return err
		}
		_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='BLOCKED',updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type='MAIN_STEP'`, stepID)
	} else if approval {
		_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='WAITING_FOR_APPROVAL',customer_status_text='در حال بررسی' WHERE id=$1`, stepID)
		if err != nil {
			return err
		}
		_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type='MAIN_STEP' AND status NOT IN ('COMPLETED','CANCELLED')`, stepID, actor)
		_, err = tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT wi.id,si.id,wi.order_id,wi.customer_user_id,'تأیید '||si.internal_title_fa,'اطلاعات ثبت‌شده را بررسی کنید','WAITING_FOR_APPROVAL','HIGH',si.approval_role_id,'workflow_steps.approve',si.estimated_end_at,'approval:'||si.id,'APPROVAL' FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.id=$1 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, stepID)
		if err != nil {
			return err
		}
	} else {
		if err = s.completeStepTx(ctx, tx, actor, workflowID, stepID, "COMPLETED"); err != nil {
			return err
		}
	}
	s.auditTx(ctx, tx, actor, chooseAudit("workflow_steps.submit", override), "workflow_step_instance", stepID, nil, map[string]any{"reason": p.Reason})
	return tx.Commit()
}

func (s *OperationsService) evaluateHandoffsTx(ctx context.Context, tx *sql.Tx, actor, workflowID, stepID string) (bool, error) {
	rows, err := tx.QueryContext(ctx, `SELECT f.handoff_metric_key,m.unit_code,m.absolute_tolerance,m.percentage_tolerance,m.blocking_on_mismatch,v.value_json,si.sequence_number FROM workflow_instance_field_definitions f JOIN workflow_step_field_values v ON v.field_definition_id=f.id JOIN workflow_step_instances si ON si.id=f.workflow_step_instance_id JOIN workflow_instance_handoff_metrics m ON m.workflow_instance_id=f.workflow_instance_id AND m.metric_key=f.handoff_metric_key WHERE f.workflow_step_instance_id=$1 AND f.handoff_direction='IN'`, stepID)
	if err != nil {
		return false, err
	}
	type handoffInput struct {
		metric, unit   string
		absTol, pctTol sql.NullFloat64
		blocking       bool
		actualRaw      []byte
		seq            int
	}
	inputs := []handoffInput{}
	for rows.Next() {
		var input handoffInput
		if err := rows.Scan(&input.metric, &input.unit, &input.absTol, &input.pctTol, &input.blocking, &input.actualRaw, &input.seq); err != nil {
			rows.Close()
			return false, err
		}
		inputs = append(inputs, input)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return false, err
	}
	if err = rows.Close(); err != nil {
		return false, err
	}
	blockingFound := false
	for _, input := range inputs {
		metric, unit, absTol, pctTol, blocking, actualRaw, seq := input.metric, input.unit, input.absTol, input.pctTol, input.blocking, input.actualRaw, input.seq
		actual, ok := extractNumber(actualRaw)
		if !ok {
			return false, fmt.Errorf("%w: handoff metric %s is not numeric", ErrValidation, metric)
		}
		var sourceID string
		var expectedRaw []byte
		err = tx.QueryRowContext(ctx, `SELECT f.workflow_step_instance_id,v.value_json FROM workflow_instance_field_definitions f JOIN workflow_step_field_values v ON v.field_definition_id=f.id JOIN workflow_step_instances si ON si.id=f.workflow_step_instance_id WHERE f.workflow_instance_id=$1 AND f.handoff_metric_key=$2 AND f.handoff_direction='OUT' AND si.sequence_number<$3 ORDER BY si.sequence_number DESC LIMIT 1`, workflowID, metric, seq).Scan(&sourceID, &expectedRaw)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return false, err
		}
		expected, ok := extractNumber(expectedRaw)
		if !ok {
			return false, fmt.Errorf("%w: previous handoff metric %s is not numeric", ErrValidation, metric)
		}
		_, err = tx.ExecContext(ctx, `UPDATE workflow_discrepancies SET status='CANCELLED',resolution_note='superseded by corrected submission',resolved_by_user_id=$3,resolved_at=NOW(),updated_at=NOW() WHERE target_step_instance_id=$1 AND metric_key=$2 AND status IN ('OPEN','UNDER_REVIEW','CORRECTION_REQUIRED')`, stepID, metric, actor)
		if err != nil {
			return false, err
		}
		difference := actual - expected
		var absoluteTolerance, percentageTolerance *float64
		if absTol.Valid {
			v := absTol.Float64
			absoluteTolerance = &v
		}
		if pctTol.Valid {
			v := pctTol.Float64
			percentageTolerance = &v
		}
		allowed, pct := withinHandoffTolerance(expected, actual, absoluteTolerance, percentageTolerance)
		if allowed {
			continue
		}
		severity := "WARNING"
		if blocking {
			severity = "CRITICAL"
			blockingFound = true
		}
		var pctValue any = nil
		if pct != nil {
			pctValue = *pct
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO workflow_discrepancies(workflow_instance_id,source_step_instance_id,target_step_instance_id,metric_key,expected_value,actual_value,difference_value,difference_percentage,unit_code,severity,is_blocking,status,reported_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'OPEN',$12)`, workflowID, sourceID, stepID, metric, expected, actual, difference, pctValue, unit, severity, blocking, actor)
		if err != nil {
			return false, err
		}
	}
	return blockingFound, nil
}

func withinHandoffTolerance(expected, actual float64, absoluteTolerance, percentageTolerance *float64) (bool, *float64) {
	absDiff := math.Abs(actual - expected)
	if absoluteTolerance != nil && absDiff <= *absoluteTolerance {
		return true, percentageDifference(expected, absDiff)
	}
	pct := percentageDifference(expected, absDiff)
	if expected != 0 && percentageTolerance != nil && pct != nil && *pct <= *percentageTolerance {
		return true, pct
	}
	return absoluteTolerance == nil && percentageTolerance == nil && absDiff == 0, pct
}

func percentageDifference(expected, absoluteDifference float64) *float64 {
	if expected == 0 {
		return nil
	}
	value := absoluteDifference / math.Abs(expected) * 100
	return &value
}

func (s *OperationsService) completeStepTx(ctx context.Context, tx *sql.Tx, actor, workflowID, stepID, terminal string) error {
	var blockers int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM action_items WHERE workflow_step_instance_id=$1 AND is_blocking AND status NOT IN ('COMPLETED','CANCELLED')`, stepID).Scan(&blockers); err != nil {
		return err
	}
	if blockers > 0 {
		_, err := tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='BLOCKED',customer_status_text='در حال تکمیل',updated_at=NOW() WHERE id=$1`, stepID)
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status=$2,actual_end_at=NOW(),customer_status_text=$3,updated_at=NOW() WHERE id=$1`, stepID, terminal, customerStatus(terminal)); err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type IN ('MAIN_STEP','APPROVAL','CORRECTION') AND status NOT IN ('COMPLETED','CANCELLED')`, stepID, actor)
	if err := s.runStepTriggersTx(ctx, tx, workflowID, stepID, "ON_STEP_COMPLETE"); err != nil {
		return err
	}
	var customerVisible bool
	var title string
	_ = tx.QueryRowContext(ctx, `SELECT customer_visible,customer_title_fa FROM workflow_step_instances WHERE id=$1`, stepID).Scan(&customerVisible, &title)
	if customerVisible {
		_, _ = tx.ExecContext(ctx, `INSERT INTO order_timeline(order_id,title_fa,status_code,occurred_at) SELECT order_id,$2,$3,NOW() FROM workflow_instances WHERE id=$1`, workflowID, title, terminal)
	}
	if handled, err := s.routeCompletedStepTx(ctx, tx, actor, workflowID, stepID); err != nil {
		return err
	} else if handled {
		return nil
	}
	var nextID string
	var startsAutomatically bool
	err := tx.QueryRowContext(ctx, `SELECT id,starts_automatically FROM workflow_step_instances WHERE workflow_instance_id=$1 AND sequence_number>(SELECT sequence_number FROM workflow_step_instances WHERE id=$2) AND status='NOT_STARTED' AND path_state='INCLUDED' ORDER BY sequence_number LIMIT 1 FOR UPDATE`, workflowID, stepID).Scan(&nextID, &startsAutomatically)
	if errors.Is(err, sql.ErrNoRows) {
		var workflowBlockers, openDiscrepancies int
		_ = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM action_items WHERE workflow_instance_id=$1 AND workflow_step_instance_id IS NULL AND is_blocking AND status NOT IN ('COMPLETED','CANCELLED')`, workflowID).Scan(&workflowBlockers)
		_ = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_discrepancies WHERE workflow_instance_id=$1 AND is_blocking AND status NOT IN ('ACCEPTED','RESOLVED','CANCELLED')`, workflowID).Scan(&openDiscrepancies)
		if workflowBlockers+openDiscrepancies == 0 {
			_, err = tx.ExecContext(ctx, `UPDATE workflow_instances SET status='COMPLETED',completed_at=NOW(),current_step_instance_id=NULL,updated_at=NOW() WHERE id=$1`, workflowID)
			if err != nil {
				return err
			}
			_, _ = tx.ExecContext(ctx, `UPDATE orders SET status='COMPLETED',updated_at=NOW() WHERE id=(SELECT order_id FROM workflow_instances WHERE id=$1 AND scope_type='ORDER')`, workflowID)
			if err = s.completeChildDependencyTx(ctx, tx, actor, workflowID); err != nil {
				return err
			}
		}
		return nil
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='WAITING_FOR_ASSIGNEE',customer_status_text='در انتظار شروع',updated_at=NOW() WHERE id=$1`, nextID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_instances SET current_step_instance_id=$2,updated_at=NOW() WHERE id=$1`, workflowID, nextID)
	if err != nil {
		return err
	}
	if err = s.createMainStepActionTx(ctx, tx, workflowID, nextID); err != nil {
		return err
	}
	if err = s.runStepTriggersTx(ctx, tx, workflowID, nextID, "ON_STEP_OPEN"); err != nil {
		return err
	}
	if startsAutomatically {
		if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='IN_PROGRESS',actual_start_at=NOW(),customer_status_text='در حال انجام',updated_at=NOW() WHERE id=$1`, nextID); err != nil {
			return err
		}
		_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='IN_PROGRESS',updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type='MAIN_STEP'`, nextID)
		return s.runStepTriggersTx(ctx, tx, workflowID, nextID, "ON_STEP_START")
	}
	return nil
}

func (s *OperationsService) ApproveWorkflowStep(ctx context.Context, actor, stepID, reason string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	workflowID, status, _, _, _, _, _, err := s.lockStepTx(ctx, tx, stepID)
	if err != nil {
		return err
	}
	if !validWorkflowStepTransition(status, "COMPLETED") || status != "WAITING_FOR_APPROVAL" {
		return ErrInvalidTransition
	}
	var approvalRole sql.NullInt64
	if err = tx.QueryRowContext(ctx, `SELECT approval_role_id FROM workflow_step_instances WHERE id=$1`, stepID).Scan(&approvalRole); err != nil {
		return err
	}
	authorized := approvalRole.Valid && s.userHasRole(ctx, actor, s.roleCode(ctx, approvalRole.Int64)) && s.HasPermission(ctx, actor, "workflow_steps.approve")
	override := false
	if !authorized {
		if !s.HasPermission(ctx, actor, "workflow_steps.override") {
			return ErrForbidden
		}
		if strings.TrimSpace(reason) == "" {
			return errors.New("override reason is required")
		}
		override = true
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET approved_at=NOW(),approved_by_user_id=$2,result_code='APPROVED' WHERE id=$1`, stepID, actor)
	if err != nil {
		return err
	}
	if err = s.runStepTriggersTx(ctx, tx, workflowID, stepID, "ON_STEP_APPROVE"); err != nil {
		return err
	}
	if err = s.completeStepTx(ctx, tx, actor, workflowID, stepID, "COMPLETED"); err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, chooseAudit("workflow_steps.approve", override), "workflow_step_instance", stepID, nil, map[string]any{"reason": reason})
	return tx.Commit()
}
func (s *OperationsService) RejectWorkflowStep(ctx context.Context, actor, stepID, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("rejection reason is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	workflowID, status, _, _, _, _, _, err := s.lockStepTx(ctx, tx, stepID)
	if err != nil {
		return err
	}
	if !validWorkflowStepTransition(status, "NEEDS_CORRECTION") {
		return ErrInvalidTransition
	}
	var approvalRole sql.NullInt64
	if err = tx.QueryRowContext(ctx, `SELECT approval_role_id FROM workflow_step_instances WHERE id=$1`, stepID).Scan(&approvalRole); err != nil {
		return err
	}
	authorized := approvalRole.Valid && s.userHasRole(ctx, actor, s.roleCode(ctx, approvalRole.Int64)) && s.HasPermission(ctx, actor, "workflow_steps.reject")
	override := !authorized
	if override && !s.HasPermission(ctx, actor, "workflow_steps.override") {
		return ErrForbidden
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='NEEDS_CORRECTION',rejected_at=NOW(),rejected_by_user_id=$2,rejection_reason=$3,customer_status_text='در حال تکمیل',updated_at=NOW() WHERE id=$1`, stepID, actor, reason)
	if err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='CANCELLED',updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type='APPROVAL' AND status<>'COMPLETED'`, stepID)
	_, err = tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,assigned_user_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT wi.id,si.id,wi.order_id,wi.customer_user_id,'اصلاح '||si.internal_title_fa,$2,'OPEN','HIGH',si.responsible_role_id,si.assigned_user_id,si.required_permission_code,si.estimated_end_at,'correction:'||si.id||':'||EXTRACT(EPOCH FROM NOW())::bigint,'CORRECTION' FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.id=$1`, stepID, reason)
	if err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, chooseAudit("workflow_steps.reject", override), "workflow_step_instance", stepID, nil, map[string]any{"reason": reason, "workflow_id": workflowID})
	return tx.Commit()
}
func (s *OperationsService) SkipWorkflowStep(ctx context.Context, actor, stepID, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("skip reason is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	workflowID, status, permission, role, assigned, _, skippable, err := s.lockStepTx(ctx, tx, stepID)
	if err != nil {
		return err
	}
	if !validWorkflowStepTransition(status, "SKIPPED") {
		return ErrInvalidTransition
	}
	authorized := skippable && s.HasPermission(ctx, actor, "workflow_steps.skip") && s.canOperateStep(ctx, actor, permission, role, assigned)
	override := false
	if !authorized {
		if !s.HasPermission(ctx, actor, "workflow_steps.override") {
			return ErrForbidden
		}
		override = true
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET skipped_at=NOW(),skipped_by_user_id=$2,skip_reason=$3 WHERE id=$1`, stepID, actor, reason)
	if err != nil {
		return err
	}
	if err = s.completeStepTx(ctx, tx, actor, workflowID, stepID, "SKIPPED"); err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, chooseAudit("workflow_steps.skip", override), "workflow_step_instance", stepID, nil, map[string]any{"reason": reason})
	return tx.Commit()
}
func (s *OperationsService) ReopenWorkflowStep(ctx context.Context, actor, stepID, reason string) error {
	if strings.TrimSpace(reason) == "" || !s.HasPermission(ctx, actor, "workflow_steps.reopen") {
		return ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	workflowID, status, _, _, _, _, _, err := s.lockStepTx(ctx, tx, stepID)
	if err != nil {
		return err
	}
	if !validWorkflowStepTransition(status, "IN_PROGRESS") || (status != "COMPLETED" && status != "SKIPPED") {
		return ErrInvalidTransition
	}
	var later int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_step_instances WHERE workflow_instance_id=$1 AND sequence_number>(SELECT sequence_number FROM workflow_step_instances WHERE id=$2) AND status NOT IN ('NOT_STARTED','CANCELLED')`, workflowID, stepID).Scan(&later); err != nil {
		return err
	}
	if later > 0 {
		return errors.New("only the latest terminal step can be reopened")
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='IN_PROGRESS',actual_end_at=NULL,skipped_at=NULL,skipped_by_user_id=NULL,skip_reason=NULL,customer_status_text='در حال انجام',updated_at=NOW() WHERE id=$1`, stepID)
	if err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE workflow_instances SET status='IN_PROGRESS',completed_at=NULL,current_step_instance_id=$2 WHERE id=$1`, workflowID, stepID)
	if err = s.createMainStepActionTx(ctx, tx, workflowID, stepID); err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, "workflow_steps.reopen.override", "workflow_step_instance", stepID, nil, map[string]any{"reason": reason})
	return tx.Commit()
}
func (s *OperationsService) ReassignWorkflowStep(ctx context.Context, actor, stepID, userID, reason string) error {
	if !s.HasPermission(ctx, actor, "workflow_steps.reassign") {
		return ErrForbidden
	}
	if strings.TrimSpace(reason) == "" {
		return errors.New("reassignment reason is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, status, _, role, _, _, _, err := s.lockStepTx(ctx, tx, stepID)
	if err != nil {
		return err
	}
	if status == "COMPLETED" || status == "SKIPPED" || status == "CANCELLED" {
		return ErrInvalidTransition
	}
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT status='ACTIVE' FROM users WHERE id=$1`, userID).Scan(&active); err != nil || !active {
		return errors.New("assigned user must be active")
	}
	hasRole := role.Valid && s.userHasRole(ctx, userID, s.roleCode(ctx, role.Int64))
	override := false
	if !hasRole {
		if !s.HasPermission(ctx, actor, "workflow_steps.override") {
			return errors.New("override reason is required when assigning outside responsible role")
		}
		override = true
	}
	var before sql.NullString
	_ = tx.QueryRowContext(ctx, `SELECT assigned_user_id FROM workflow_step_instances WHERE id=$1`, stepID).Scan(&before)
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET assigned_user_id=$2,updated_at=NOW() WHERE id=$1`, stepID, userID)
	if err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE action_items SET assigned_user_id=$2,assigned_role_id=NULL,updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type IN ('MAIN_STEP','CORRECTION') AND status NOT IN ('COMPLETED','CANCELLED')`, stepID, userID)
	s.auditTx(ctx, tx, actor, chooseAudit("workflow_steps.reassign", override), "workflow_step_instance", stepID, map[string]any{"assigned_user_id": before.String}, map[string]any{"assigned_user_id": userID, "reason": reason})
	return tx.Commit()
}

func (s *OperationsService) ListWorkflowDiscrepancies(ctx context.Context, actor, workflowID string) ([]RuntimeDiscrepancy, error) {
	viewAll := s.HasPermission(ctx, actor, "workflow_discrepancies.view_all")
	viewAssigned := s.HasPermission(ctx, actor, "workflow_discrepancies.view_assigned")
	if !viewAll && !viewAssigned {
		return nil, ErrForbidden
	}
	if !viewAll {
		var assigned bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_step_instances si WHERE si.workflow_instance_id=$1 AND (si.assigned_user_id=$2 OR (si.assigned_user_id IS NULL AND EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$2 AND ur.role_id=si.responsible_role_id))))`, workflowID, actor).Scan(&assigned); err != nil {
			return nil, err
		}
		if !assigned {
			return nil, ErrForbidden
		}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT d.id,d.source_step_instance_id,d.target_step_instance_id,d.metric_key,d.expected_value,d.actual_value,d.difference_value,d.difference_percentage,COALESCE(d.unit_code,''),d.severity,d.is_blocking,d.status,d.resolution_note,d.reported_at FROM workflow_discrepancies d WHERE d.workflow_instance_id=$1 AND ($3 OR EXISTS(SELECT 1 FROM workflow_step_instances si WHERE si.id IN (d.source_step_instance_id,d.target_step_instance_id) AND (si.assigned_user_id=$2 OR (si.assigned_user_id IS NULL AND EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$2 AND ur.role_id=si.responsible_role_id))))) ORDER BY d.reported_at DESC`, workflowID, actor, viewAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RuntimeDiscrepancy{}
	for rows.Next() {
		var d RuntimeDiscrepancy
		var source, target, resolution sql.NullString
		var expected, actual, diff, pct sql.NullFloat64
		if err := rows.Scan(&d.ID, &source, &target, &d.MetricKey, &expected, &actual, &diff, &pct, &d.UnitCode, &d.Severity, &d.IsBlocking, &d.Status, &resolution, &d.ReportedAt); err != nil {
			return nil, err
		}
		if source.Valid {
			v := source.String
			d.SourceStepInstanceID = &v
		}
		if target.Valid {
			v := target.String
			d.TargetStepInstanceID = &v
		}
		if resolution.Valid {
			v := resolution.String
			d.ResolutionNote = &v
		}
		if expected.Valid {
			v := expected.Float64
			d.ExpectedValue = &v
		}
		if actual.Valid {
			v := actual.Float64
			d.ActualValue = &v
		}
		if diff.Valid {
			v := diff.Float64
			d.DifferenceValue = &v
		}
		if pct.Valid {
			v := pct.Float64
			d.DifferencePercentage = &v
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *OperationsService) GetWorkflowDiscrepancy(ctx context.Context, actor, id string) (RuntimeDiscrepancy, error) {
	var workflowID string
	if err := s.db.QueryRowContext(ctx, `SELECT workflow_instance_id FROM workflow_discrepancies WHERE id=$1`, id).Scan(&workflowID); err != nil {
		return RuntimeDiscrepancy{}, err
	}
	items, err := s.ListWorkflowDiscrepancies(ctx, actor, workflowID)
	if err != nil {
		return RuntimeDiscrepancy{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return RuntimeDiscrepancy{}, ErrForbidden
}

func (s *OperationsService) CreateWorkflowDiscrepancy(ctx context.Context, actor, workflowID string, p CreateDiscrepancyPayload) (RuntimeDiscrepancy, error) {
	if !s.HasPermission(ctx, actor, "workflow_discrepancies.review") && !s.HasPermission(ctx, actor, "workflow_discrepancies.override") {
		return RuntimeDiscrepancy{}, ErrForbidden
	}
	if _, err := s.ListWorkflowDiscrepancies(ctx, actor, workflowID); err != nil {
		return RuntimeDiscrepancy{}, err
	}
	if strings.TrimSpace(p.MetricKey) == "" || strings.TrimSpace(p.Explanation) == "" {
		return RuntimeDiscrepancy{}, errors.New("metric and explanation are required")
	}
	if p.IsBlocking {
		return RuntimeDiscrepancy{}, errors.New("manual discrepancies must be non-blocking")
	}
	p.Severity = "INFO"
	var target any = nil
	if p.TargetStepInstanceID != "" {
		var valid bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_step_instances WHERE id=$1 AND workflow_instance_id=$2)`, p.TargetStepInstanceID, workflowID).Scan(&valid); err != nil {
			return RuntimeDiscrepancy{}, err
		}
		if !valid {
			return RuntimeDiscrepancy{}, ErrValidation
		}
		target = p.TargetStepInstanceID
	}
	var id string
	err := s.db.QueryRowContext(ctx, `INSERT INTO workflow_discrepancies(workflow_instance_id,target_step_instance_id,metric_key,expected_value,actual_value,difference_value,unit_code,severity,is_blocking,status,reported_by_user_id,source_explanation) VALUES($1,$2,$3,$4,$5,CASE WHEN $4::numeric IS NULL OR $5::numeric IS NULL THEN NULL ELSE $5::numeric-$4::numeric END,NULLIF($6,''),$7,$8,'OPEN',$9,$10) RETURNING id`, workflowID, target, p.MetricKey, p.ExpectedValue, p.ActualValue, p.UnitCode, p.Severity, p.IsBlocking, actor, p.Explanation).Scan(&id)
	if err != nil {
		return RuntimeDiscrepancy{}, err
	}
	s.audit(ctx, actor, "workflow_discrepancies.create", "workflow_discrepancy", id, map[string]any{"workflow_instance_id": workflowID, "is_blocking": p.IsBlocking})
	return s.GetWorkflowDiscrepancy(ctx, actor, id)
}
func (s *OperationsService) ResolveWorkflowDiscrepancy(ctx context.Context, actor, id, status, note string) error {
	requiredPermission := "workflow_discrepancies.resolve"
	if status == "UNDER_REVIEW" {
		requiredPermission = "workflow_discrepancies.review"
	}
	if status == "CANCELLED" {
		requiredPermission = "workflow_discrepancies.override"
	}
	if !s.HasPermission(ctx, actor, requiredPermission) {
		return ErrForbidden
	}
	if strings.TrimSpace(note) == "" {
		return errors.New("resolution note is required")
	}
	if status != "ACCEPTED" && status != "RESOLVED" && status != "CORRECTION_REQUIRED" && status != "UNDER_REVIEW" && status != "CANCELLED" {
		return errors.New("invalid discrepancy status")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var workflowID string
	var target sql.NullString
	var oldStatus string
	if err = tx.QueryRowContext(ctx, `SELECT workflow_instance_id,target_step_instance_id,status FROM workflow_discrepancies WHERE id=$1 FOR UPDATE`, id).Scan(&workflowID, &target, &oldStatus); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_discrepancies SET status=$2,resolution_note=$3,resolved_by_user_id=CASE WHEN $2 IN ('ACCEPTED','RESOLVED','CANCELLED') THEN $4 ELSE resolved_by_user_id END,resolved_at=CASE WHEN $2 IN ('ACCEPTED','RESOLVED','CANCELLED') THEN NOW() ELSE resolved_at END,updated_at=NOW() WHERE id=$1`, id, status, note, actor)
	if err != nil {
		return err
	}
	if target.Valid && (status == "ACCEPTED" || status == "RESOLVED" || status == "CANCELLED") {
		var remaining int
		_ = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_discrepancies WHERE target_step_instance_id=$1 AND id<>$2 AND is_blocking AND status NOT IN ('ACCEPTED','RESOLVED','CANCELLED')`, target.String, id).Scan(&remaining)
		if remaining == 0 {
			var blockingTasks int
			if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM action_items WHERE workflow_step_instance_id=$1 AND is_blocking AND status NOT IN ('COMPLETED','CANCELLED')`, target.String).Scan(&blockingTasks); err != nil {
				return err
			}
			var requiresApproval bool
			if err = tx.QueryRowContext(ctx, `SELECT requires_approval FROM workflow_step_instances WHERE id=$1`, target.String).Scan(&requiresApproval); err != nil {
				return err
			}
			if blockingTasks > 0 {
				_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='BLOCKED',customer_status_text='در حال تکمیل',updated_at=NOW() WHERE id=$1 AND status='HAS_MISMATCH'`, target.String)
			} else if requiresApproval {
				_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='WAITING_FOR_APPROVAL',customer_status_text='در حال بررسی' WHERE id=$1 AND status='HAS_MISMATCH'`, target.String)
				if err == nil {
					_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type='MAIN_STEP' AND status NOT IN ('COMPLETED','CANCELLED')`, target.String, actor)
					err = s.createApprovalActionTx(ctx, tx, target.String)
				}
			} else {
				err = s.completeStepTx(ctx, tx, actor, workflowID, target.String, "COMPLETED")
			}
			if err != nil {
				return err
			}
		}
	}
	if target.Valid && status == "CORRECTION_REQUIRED" {
		if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='NEEDS_CORRECTION',customer_status_text='در حال تکمیل',updated_at=NOW() WHERE id=$1 AND status='HAS_MISMATCH'`, target.String); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,assigned_user_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT wi.id,si.id,wi.order_id,wi.customer_user_id,'اصلاح مغایرت '||si.internal_title_fa,$2,'OPEN','URGENT',si.responsible_role_id,si.assigned_user_id,si.required_permission_code,si.estimated_end_at,'mismatch-correction:'||$1,'CORRECTION' FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.id=$3 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, id, note, target.String)
		if err != nil {
			return err
		}
	}
	s.auditTx(ctx, tx, actor, "workflow_discrepancies.resolve", "workflow_discrepancy", id, map[string]any{"status": oldStatus}, map[string]any{"status": status, "note": note})
	return tx.Commit()
}

func (s *OperationsService) CompleteActionItem(ctx context.Context, actor, id string) error {
	if !s.HasPermission(ctx, actor, "action_items.complete") {
		return ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var assignedUser, stepID, workflowID sql.NullString
	var role sql.NullInt64
	var status, sourceTrigger string
	if err = tx.QueryRowContext(ctx, `SELECT assigned_user_id,assigned_role_id,status,workflow_step_instance_id,workflow_instance_id,COALESCE(source_trigger_type,'') FROM action_items WHERE id=$1 FOR UPDATE`, id).Scan(&assignedUser, &role, &status, &stepID, &workflowID, &sourceTrigger); err != nil {
		return err
	}
	if status == "COMPLETED" {
		return nil
	}
	if sourceTrigger == "MAIN_STEP" || sourceTrigger == "APPROVAL" || sourceTrigger == "CORRECTION" {
		return ErrInvalidTransition
	}
	allowed := assignedUser.Valid && assignedUser.String == actor
	if !allowed && role.Valid {
		allowed = s.userHasRole(ctx, actor, s.roleCode(ctx, role.Int64))
	}
	if !allowed && !s.HasPermission(ctx, actor, "action_items.view_all") {
		return ErrForbidden
	}
	_, err = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE id=$1`, id, actor)
	if err != nil {
		return err
	}
	if stepID.Valid && workflowID.Valid {
		var stepStatus string
		if err = tx.QueryRowContext(ctx, `SELECT status FROM workflow_step_instances WHERE id=$1 FOR UPDATE`, stepID.String).Scan(&stepStatus); err != nil {
			return err
		}
		if stepStatus == "BLOCKED" {
			var remainingTasks, remainingDiscrepancies int
			if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM action_items WHERE workflow_step_instance_id=$1 AND is_blocking AND status NOT IN ('COMPLETED','CANCELLED')`, stepID.String).Scan(&remainingTasks); err != nil {
				return err
			}
			if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_discrepancies WHERE target_step_instance_id=$1 AND is_blocking AND status NOT IN ('ACCEPTED','RESOLVED','CANCELLED')`, stepID.String).Scan(&remainingDiscrepancies); err != nil {
				return err
			}
			if remainingTasks == 0 && remainingDiscrepancies == 0 {
				var requiresApproval bool
				if err = tx.QueryRowContext(ctx, `SELECT requires_approval FROM workflow_step_instances WHERE id=$1`, stepID.String).Scan(&requiresApproval); err != nil {
					return err
				}
				if requiresApproval {
					if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='WAITING_FOR_APPROVAL',customer_status_text='در حال بررسی',updated_at=NOW() WHERE id=$1`, stepID.String); err != nil {
						return err
					}
					_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type='MAIN_STEP' AND status NOT IN ('COMPLETED','CANCELLED')`, stepID.String, actor)
					if err = s.createApprovalActionTx(ctx, tx, stepID.String); err != nil {
						return err
					}
				} else if err = s.completeStepTx(ctx, tx, actor, workflowID.String, stepID.String, "COMPLETED"); err != nil {
					return err
				}
			}
		}
	}
	s.auditTx(ctx, tx, actor, "action_items.complete", "action_item", id, nil, nil)
	return tx.Commit()
}

func (s *OperationsService) createApprovalActionTx(ctx context.Context, tx *sql.Tx, stepID string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT wi.id,si.id,wi.order_id,wi.customer_user_id,'تأیید '||si.internal_title_fa,'اطلاعات ثبت‌شده را بررسی کنید','WAITING_FOR_APPROVAL','HIGH',si.approval_role_id,'workflow_steps.approve',si.estimated_end_at,'approval:'||si.id,'APPROVAL' FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.id=$1 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, stepID)
	return err
}

func extractNumber(raw []byte) (float64, bool) {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return x, true
	case map[string]any:
		for _, k := range []string{"value", "amount"} {
			if n, ok := x[k].(float64); ok {
				return n, true
			}
		}
	case string:
		n, e := strconv.ParseFloat(x, 64)
		return n, e == nil
	}
	return 0, false
}
