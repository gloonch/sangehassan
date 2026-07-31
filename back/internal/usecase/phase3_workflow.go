package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func allowedWorkflowResult(v string) bool {
	switch v {
	case "APPROVED", "REJECTED", "HAS_DISCREPANCY", "CORRECTION_REQUIRED", "CUSTOMER_CANCELLED", "PAYMENT_PENDING":
		return true
	}
	return false
}

func (s *OperationsService) ListWorkflowTransitions(ctx context.Context, templateID int64) ([]WorkflowTransitionDefinition, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,workflow_template_id,source_step_id,target_step_id,transition_code,label_fa,transition_type,result_code,is_default,requires_permission_code,requires_reason,sort_order FROM workflow_step_transitions WHERE workflow_template_id=$1 ORDER BY source_step_id,sort_order,id`, templateID)
	if err != nil {
		return nil, err
	}
	out := []WorkflowTransitionDefinition{}
	for rows.Next() {
		var x WorkflowTransitionDefinition
		var result, permission sql.NullString
		if err = rows.Scan(&x.ID, &x.WorkflowTemplateID, &x.SourceStepID, &x.TargetStepID, &x.TransitionCode, &x.LabelFA, &x.TransitionType, &result, &x.IsDefault, &permission, &x.RequiresReason, &x.SortOrder); err != nil {
			return nil, err
		}
		x.ResultCode = scanNullableString(result)
		x.RequiresPermissionCode = scanNullableString(permission)
		out = append(out, x)
	}
	return out, rows.Err()
}
func validateTransitionPayload(p WorkflowTransitionPayload) error {
	p.TransitionCode = normalizeCode(p.TransitionCode)
	p.TransitionType = normalizeCode(p.TransitionType)
	if p.SourceStepID == p.TargetStepID || p.SourceStepID == 0 || p.TargetStepID == 0 || p.TransitionCode == "" || strings.TrimSpace(p.LabelFA) == "" || p.SortOrder < 0 {
		return ErrValidation
	}
	if p.TransitionType != "AUTOMATIC" && p.TransitionType != "MANUAL_SELECTION" && p.TransitionType != "RESULT_BASED" {
		return ErrValidation
	}
	if p.TransitionType == "RESULT_BASED" && (p.ResultCode == nil || !allowedWorkflowResult(normalizeCode(*p.ResultCode))) {
		return ErrValidation
	}
	if p.TransitionType != "RESULT_BASED" && p.ResultCode != nil {
		return ErrValidation
	}
	return nil
}
func (s *OperationsService) CreateWorkflowTransition(ctx context.Context, actor string, templateID int64, p WorkflowTransitionPayload) (WorkflowTransitionDefinition, error) {
	var out WorkflowTransitionDefinition
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return out, err
	}
	if err := validateTransitionPayload(p); err != nil {
		return out, err
	}
	var id int64
	err := s.db.QueryRowContext(ctx, `INSERT INTO workflow_step_transitions(workflow_template_id,source_step_id,target_step_id,transition_code,label_fa,transition_type,result_code,is_default,requires_permission_code,requires_reason,sort_order) SELECT $1,$2,$3,UPPER($4),$5,$6,NULLIF(UPPER($7),''),$8,$9,$10,$11 WHERE EXISTS(SELECT 1 FROM workflow_template_steps s JOIN workflow_template_steps t ON t.workflow_template_id=s.workflow_template_id WHERE s.id=$2 AND t.id=$3 AND s.workflow_template_id=$1) RETURNING id`, templateID, p.SourceStepID, p.TargetStepID, p.TransitionCode, p.LabelFA, normalizeCode(p.TransitionType), valueOrNil(p.ResultCode), p.IsDefault, p.RequiresPermissionCode, p.RequiresReason, p.SortOrder).Scan(&id)
	if err != nil {
		return out, err
	}
	s.audit(ctx, actor, "workflow_transitions.create", "workflow_step_transition", fmt.Sprint(id), p)
	items, err := s.ListWorkflowTransitions(ctx, templateID)
	if err != nil {
		return out, err
	}
	for _, x := range items {
		if x.ID == id {
			return x, nil
		}
	}
	return out, sql.ErrNoRows
}
func valueOrNil(v *string) any {
	if v == nil {
		return ""
	}
	return *v
}
func (s *OperationsService) UpdateWorkflowTransition(ctx context.Context, actor string, id int64, p WorkflowTransitionPayload) error {
	var templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT workflow_template_id FROM workflow_step_transitions WHERE id=$1`, id).Scan(&templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	if err := validateTransitionPayload(p); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `UPDATE workflow_step_transitions SET source_step_id=$2,target_step_id=$3,transition_code=UPPER($4),label_fa=$5,transition_type=$6,result_code=NULLIF(UPPER($7),''),is_default=$8,requires_permission_code=$9,requires_reason=$10,sort_order=$11,updated_at=NOW() WHERE id=$1`, id, p.SourceStepID, p.TargetStepID, p.TransitionCode, p.LabelFA, normalizeCode(p.TransitionType), valueOrNil(p.ResultCode), p.IsDefault, p.RequiresPermissionCode, p.RequiresReason, p.SortOrder)
	if err == nil {
		s.audit(ctx, actor, "workflow_transitions.update", "workflow_step_transition", fmt.Sprint(id), p)
	}
	return err
}
func (s *OperationsService) DeleteWorkflowTransition(ctx context.Context, actor string, id int64) error {
	var templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT workflow_template_id FROM workflow_step_transitions WHERE id=$1`, id).Scan(&templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM workflow_step_transitions WHERE id=$1`, id)
	if err == nil {
		s.audit(ctx, actor, "workflow_transitions.delete", "workflow_step_transition", fmt.Sprint(id), nil)
	}
	return err
}

func (s *OperationsService) validateWorkflowTransitions(ctx context.Context, templateID int64) error {
	var count, entries int
	var err error
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_step_transitions WHERE workflow_template_id=$1`, templateID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_template_steps WHERE workflow_template_id=$1 AND is_active AND is_entry`, templateID).Scan(&entries); err != nil {
		return err
	}
	if entries != 1 {
		return errors.New("branched template requires exactly one active entry step")
	}
	var invalid int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_step_transitions t JOIN workflow_template_steps s ON s.id=t.source_step_id JOIN workflow_template_steps d ON d.id=t.target_step_id LEFT JOIN permissions p ON p.code=t.requires_permission_code AND p.is_active WHERE t.workflow_template_id=$1 AND (s.workflow_template_id<>$1 OR d.workflow_template_id<>$1 OR NOT s.is_active OR NOT d.is_active OR (t.requires_permission_code IS NOT NULL AND p.id IS NULL) OR (t.transition_type='RESULT_BASED' AND t.result_code IS NULL) OR (t.transition_type<>'RESULT_BASED' AND t.result_code IS NOT NULL))`, templateID).Scan(&invalid); err != nil {
		return err
	}
	if invalid > 0 {
		return errors.New("template contains invalid transition references")
	}
	var ambiguous int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM (SELECT source_step_id FROM workflow_step_transitions WHERE workflow_template_id=$1 GROUP BY source_step_id HAVING COUNT(*) FILTER(WHERE is_default)>1 OR (COUNT(*) FILTER(WHERE transition_type='AUTOMATIC')>0 AND COUNT(*) FILTER(WHERE transition_type='AUTOMATIC' AND is_default)<>1) OR COUNT(*) FILTER(WHERE transition_type='RESULT_BASED')<>COUNT(DISTINCT result_code) FILTER(WHERE transition_type='RESULT_BASED')) x`, templateID).Scan(&ambiguous); err != nil {
		return err
	}
	if ambiguous > 0 {
		return errors.New("template contains ambiguous transitions")
	}
	var active, reachable int
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_template_steps WHERE workflow_template_id=$1 AND is_active`, templateID).Scan(&active); err != nil {
		return err
	}
	if err = s.db.QueryRowContext(ctx, `WITH RECURSIVE graph(id) AS (SELECT id FROM workflow_template_steps WHERE workflow_template_id=$1 AND is_active AND is_entry UNION SELECT t.target_step_id FROM workflow_step_transitions t JOIN graph g ON g.id=t.source_step_id WHERE t.workflow_template_id=$1) SELECT COUNT(DISTINCT id) FROM graph`, templateID).Scan(&reachable); err != nil {
		return err
	}
	if active != reachable {
		return errors.New("every active step must be reachable from the entry step")
	}
	return nil
}

func (s *OperationsService) GetRuntimeTransitions(ctx context.Context, actor, stepID string) ([]RuntimeTransition, error) {
	var workflowID, status, stepCode string
	if err := s.db.QueryRowContext(ctx, `SELECT workflow_instance_id,status,step_code FROM workflow_step_instances WHERE id=$1`, stepID).Scan(&workflowID, &status, &stepCode); err != nil {
		return nil, err
	}
	if status != "WAITING_FOR_TRANSITION" && status != "IN_PROGRESS" {
		return []RuntimeTransition{}, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,transition_code,label_fa,transition_type,result_code,requires_permission_code,requires_reason,target_step_code FROM workflow_instance_step_transitions WHERE workflow_instance_id=$1 AND source_step_code=$2 ORDER BY sort_order,id`, workflowID, stepCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RuntimeTransition{}
	for rows.Next() {
		var x RuntimeTransition
		var result, permission sql.NullString
		if err = rows.Scan(&x.ID, &x.TransitionCode, &x.LabelFA, &x.TransitionType, &result, &permission, &x.RequiresReason, &x.TargetStepCode); err != nil {
			return nil, err
		}
		if permission.Valid && !s.HasPermission(ctx, actor, permission.String) {
			continue
		}
		x.ResultCode = scanNullableString(result)
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) activateTransitionTx(ctx context.Context, tx *sql.Tx, actor, workflowID, sourceID, transitionID, reason string, override bool) (string, error) {
	var targetCode, code, label string
	var requiresReason bool
	var permission, result sql.NullString
	var err error
	if err := tx.QueryRowContext(ctx, `SELECT target_step_code,transition_code,label_fa,requires_reason,requires_permission_code,result_code FROM workflow_instance_step_transitions WHERE id=$1 AND workflow_instance_id=$2 FOR UPDATE`, transitionID, workflowID).Scan(&targetCode, &code, &label, &requiresReason, &permission, &result); err != nil {
		return "", err
	}
	if permission.Valid && !s.HasPermission(ctx, actor, permission.String) {
		return "", ErrForbidden
	}
	if (requiresReason || override) && requireReason(reason) != nil {
		return "", errors.New("transition reason is required")
	}
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_transition_selections WHERE source_step_instance_id=$1`, sourceID).Scan(&existing); err != nil {
		return "", err
	}
	if existing > 0 && !override {
		return "", conflict("INVALID_TRANSITION", "transition was already selected")
	}
	if override {
		if !s.HasPermission(ctx, actor, "workflow_transitions.override") {
			return "", ErrForbidden
		}
		var sideEffects bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_domain_operation_completions WHERE workflow_step_instance_id IN (SELECT id FROM workflow_step_instances WHERE workflow_instance_id=$1 AND actual_start_at IS NOT NULL AND id<>$2))`, workflowID, sourceID).Scan(&sideEffects); err != nil {
			return "", err
		}
		if sideEffects {
			return "", conflict("REVERSAL_REQUIRED", "downstream domain operations must be reversed before changing route")
		}
		_, _ = tx.ExecContext(ctx, `DELETE FROM workflow_transition_selections WHERE source_step_instance_id=$1`, sourceID)
	}
	var targetID string
	var targetStatus string
	err = tx.QueryRowContext(ctx, `SELECT id,status FROM workflow_step_instances WHERE workflow_instance_id=$1 AND step_code=$2 AND iteration_number=1 FOR UPDATE`, workflowID, targetCode).Scan(&targetID, &targetStatus)
	if err != nil {
		return "", err
	}
	if targetStatus != "NOT_STARTED" {
		var maxIteration, limit int
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(si.iteration_number),0),wt.max_iterations FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id JOIN workflow_templates wt ON wt.id=wi.workflow_template_id WHERE si.workflow_instance_id=$1 AND si.step_code=$2 GROUP BY wt.max_iterations`, workflowID, targetCode).Scan(&maxIteration, &limit); err != nil {
			return "", err
		}
		if maxIteration >= limit {
			return "", conflict("MAX_ITERATIONS", "workflow iteration limit reached")
		}
		var prior string
		if err = tx.QueryRowContext(ctx, `SELECT id FROM workflow_step_instances WHERE workflow_instance_id=$1 AND step_code=$2 ORDER BY iteration_number DESC LIMIT 1`, workflowID, targetCode).Scan(&prior); err != nil {
			return "", err
		}
		err = tx.QueryRowContext(ctx, `INSERT INTO workflow_step_instances(workflow_instance_id,workflow_template_step_id,template_step_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,status,assigned_role_id,responsible_role_id,required_permission_code,requires_approval,approval_role_id,is_optional,is_skippable,starts_automatically,customer_visible,estimated_start_at,estimated_end_at,customer_status_text,iteration_number,path_state,predecessor_step_instance_id,domain_event_code) SELECT workflow_instance_id,workflow_template_step_id,template_step_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,'WAITING_FOR_ASSIGNEE',assigned_role_id,responsible_role_id,required_permission_code,requires_approval,approval_role_id,is_optional,is_skippable,starts_automatically,customer_visible,NOW(),NOW()+COALESCE(estimated_end_at-estimated_start_at,INTERVAL '24 hours'),'در انتظار شروع',$3,'INCLUDED',$4,domain_event_code FROM workflow_step_instances WHERE id=$2 RETURNING id`, workflowID, prior, maxIteration+1, sourceID).Scan(&targetID)
		if err != nil {
			return "", err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO workflow_instance_field_definitions(workflow_instance_id,workflow_step_instance_id,source_field_definition_id,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,is_internal_cost,unit_code,currency_code,placeholder_fa,default_value,options_json,validation_json,sort_order,handoff_metric_key,handoff_direction) SELECT workflow_instance_id,$1,source_field_definition_id,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,is_internal_cost,unit_code,currency_code,placeholder_fa,default_value,options_json,validation_json,sort_order,handoff_metric_key,handoff_direction FROM workflow_instance_field_definitions WHERE workflow_step_instance_id=$2`, targetID, prior)
		if err != nil {
			return "", err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO workflow_instance_step_task_templates(workflow_instance_id,workflow_step_instance_id,source_task_template_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_step_completion) SELECT workflow_instance_id,$1,source_task_template_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_step_completion FROM workflow_instance_step_task_templates WHERE workflow_step_instance_id=$2`, targetID, prior)
		if err != nil {
			return "", err
		}
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='WAITING_FOR_ASSIGNEE',path_state='INCLUDED',predecessor_step_instance_id=$2,customer_status_text='در انتظار شروع',updated_at=NOW() WHERE id=$1`, targetID, sourceID)
		if err != nil {
			return "", err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='COMPLETED',actual_end_at=COALESCE(actual_end_at,NOW()),customer_status_text='تکمیل شده',updated_at=NOW() WHERE id=$1`, sourceID)
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_instances SET status='IN_PROGRESS',current_step_instance_id=$2,updated_at=NOW() WHERE id=$1`, workflowID, targetID)
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_transition_selections(workflow_instance_id,source_step_instance_id,transition_snapshot_id,target_step_instance_id,selected_by_user_id,result_code,reason,is_override) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8)`, workflowID, sourceID, transitionID, targetID, actor, result, reason, override)
	if err != nil {
		return "", err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE workflow_step_instance_id=$1 AND source_trigger_type='TRANSITION' AND status NOT IN ('COMPLETED','CANCELLED')`, sourceID, actor)
	if err = s.createMainStepActionTx(ctx, tx, workflowID, targetID); err != nil {
		return "", err
	}
	if err = s.runStepTriggersTx(ctx, tx, workflowID, targetID, "ON_STEP_OPEN"); err != nil {
		return "", err
	}
	s.auditTx(ctx, tx, actor, "workflow_transitions.select", "workflow_step_instance", sourceID, nil, map[string]any{"transition_code": code, "label": label, "target_step_instance_id": targetID, "reason": reason, "override": override})
	return targetID, nil
}

func (s *OperationsService) SelectWorkflowTransition(ctx context.Context, actor, stepID, key string, p SelectTransitionPayload) (map[string]any, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "WORKFLOW_TRANSITION", key, p)
	if err != nil {
		return nil, err
	}
	if idem.Existing {
		var out map[string]any
		if err = json.Unmarshal(idem.Response, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
	var workflowID, status, stepCode string
	if err = tx.QueryRowContext(ctx, `SELECT workflow_instance_id,status,step_code FROM workflow_step_instances WHERE id=$1 FOR UPDATE`, stepID).Scan(&workflowID, &status, &stepCode); err != nil {
		return nil, err
	}
	if status != "WAITING_FOR_TRANSITION" && !p.Override {
		return nil, conflict("INVALID_TRANSITION", "step is not waiting for route selection")
	}
	var transitionID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM workflow_instance_step_transitions WHERE workflow_instance_id=$1 AND source_step_code=$2 AND transition_code=UPPER($3) AND (transition_type='MANUAL_SELECTION' OR ($4<>'' AND result_code=UPPER($4)))`, workflowID, stepCode, p.TransitionCode, valueOrNil(p.ResultCode)).Scan(&transitionID)
	if err != nil {
		return nil, conflict("INVALID_TRANSITION", "transition is not available")
	}
	target, err := s.activateTransitionTx(ctx, tx, actor, workflowID, stepID, transitionID, p.Reason, p.Override)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"target_step_instance_id": target, "transition_code": normalizeCode(p.TransitionCode)}
	if err = finishOperationTx(ctx, tx, actor, "WORKFLOW_TRANSITION", key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OperationsService) routeCompletedStepTx(ctx context.Context, tx *sql.Tx, actor, workflowID, stepID string) (bool, error) {
	var stepCode string
	var result sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT step_code,result_code FROM workflow_step_instances WHERE id=$1`, stepID).Scan(&stepCode, &result); err != nil {
		return false, err
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_instance_step_transitions WHERE workflow_instance_id=$1 AND source_step_code=$2`, workflowID, stepCode).Scan(&count); err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	var transitionID string
	err := sql.ErrNoRows
	if result.Valid {
		err = tx.QueryRowContext(ctx, `SELECT id FROM workflow_instance_step_transitions WHERE workflow_instance_id=$1 AND source_step_code=$2 AND transition_type='RESULT_BASED' AND result_code=$3 ORDER BY is_default DESC,sort_order LIMIT 1`, workflowID, stepCode, result.String).Scan(&transitionID)
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `SELECT id FROM workflow_instance_step_transitions WHERE workflow_instance_id=$1 AND source_step_code=$2 AND transition_type='AUTOMATIC' AND is_default ORDER BY sort_order LIMIT 1`, workflowID, stepCode).Scan(&transitionID)
	}
	if err == nil {
		_, err = s.activateTransitionTx(ctx, tx, actor, workflowID, stepID, transitionID, "", false)
		return true, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return true, err
	}
	var manual int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_instance_step_transitions WHERE workflow_instance_id=$1 AND source_step_code=$2 AND transition_type='MANUAL_SELECTION'`, workflowID, stepCode).Scan(&manual); err != nil {
		return true, err
	}
	if manual == 0 {
		return true, conflict("INVALID_TRANSITION", "no transition matches the step result")
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='WAITING_FOR_TRANSITION',customer_status_text='در حال بررسی مسیر',updated_at=NOW() WHERE id=$1`, stepID)
	if err != nil {
		return true, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_instances SET status='WAITING_FOR_TRANSITION',current_step_instance_id=$2,updated_at=NOW() WHERE id=$1`, workflowID, stepID)
	if err != nil {
		return true, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,status,priority,assigned_role_id,required_permission_code,deduplication_key,source_trigger_type,is_blocking) SELECT wi.id,si.id,wi.order_id,wi.customer_user_id,'انتخاب مسیر '||si.internal_title_fa,'OPEN','HIGH',si.responsible_role_id,'workflow_transitions.select','transition:'||si.id,'TRANSITION',TRUE FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.id=$1 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, stepID)
	return true, err
}

func (s *OperationsService) completeChildDependencyTx(ctx context.Context, tx *sql.Tx, actor, workflowID string) error {
	rows, err := tx.QueryContext(ctx, `SELECT id,parent_workflow_instance_id,parent_step_instance_id,action_item_id FROM workflow_child_dependencies WHERE child_workflow_instance_id=$1 AND status='OPEN' FOR UPDATE`, workflowID)
	if err != nil {
		return err
	}
	defer rows.Close()
	type dep struct {
		id, parent   string
		step, action sql.NullString
	}
	deps := []dep{}
	for rows.Next() {
		var d dep
		if err = rows.Scan(&d.id, &d.parent, &d.step, &d.action); err != nil {
			return err
		}
		deps = append(deps, d)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for _, d := range deps {
		_, err = tx.ExecContext(ctx, `UPDATE workflow_child_dependencies SET status='COMPLETED',completed_at=NOW() WHERE id=$1`, d.id)
		if err != nil {
			return err
		}
		if d.action.Valid {
			_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),updated_at=NOW() WHERE id=$1`, d.action.String)
		}
		if d.step.Valid {
			var status string
			var blockers int
			if err = tx.QueryRowContext(ctx, `SELECT status,(SELECT COUNT(*) FROM action_items WHERE workflow_step_instance_id=$1 AND is_blocking AND status NOT IN ('COMPLETED','CANCELLED'))+(SELECT COUNT(*) FROM workflow_discrepancies WHERE workflow_instance_id=$2 AND is_blocking AND status NOT IN ('ACCEPTED','RESOLVED','CANCELLED')) FROM workflow_step_instances WHERE id=$1 FOR UPDATE`, d.step.String, d.parent).Scan(&status, &blockers); err != nil {
				return err
			}
			if status == "BLOCKED" && blockers == 0 {
				if err = s.completeStepTx(ctx, tx, actor, d.parent, d.step.String, "COMPLETED"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

var _ = time.Now
