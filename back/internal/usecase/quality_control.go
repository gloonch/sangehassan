package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

const qualitySelect = `SELECT q.id,q.inspection_number,q.order_id,o.order_number,q.batch_id,b.batch_number,q.inventory_lot_id,l.lot_number,q.workflow_step_instance_id,q.inspection_type,q.status,q.assigned_user_id,q.assigned_role_id,q.inspected_by_user_id,q.inspected_at,COALESCE(q.notes,''),COALESCE(q.decision_reason,''),q.created_at,q.updated_at FROM quality_inspections q JOIN orders o ON o.id=q.order_id LEFT JOIN fulfillment_batches b ON b.id=q.batch_id LEFT JOIN inventory_lots l ON l.id=q.inventory_lot_id`

func scanQualityInspection(row interface{ Scan(...any) error }) (QualityInspection, error) {
	var out QualityInspection
	var batch, batchNumber, lot, lotNumber, step, assignedUser, inspector sql.NullString
	var assignedRole sql.NullInt64
	var inspected sql.NullTime
	err := row.Scan(&out.ID, &out.InspectionNumber, &out.OrderID, &out.OrderNumber, &batch, &batchNumber, &lot, &lotNumber, &step, &out.InspectionType, &out.Status, &assignedUser, &assignedRole, &inspector, &inspected, &out.Notes, &out.DecisionReason, &out.CreatedAt, &out.UpdatedAt)
	out.BatchID, out.BatchNumber = scanNullableString(batch), scanNullableString(batchNumber)
	out.InventoryLotID, out.LotNumber = scanNullableString(lot), scanNullableString(lotNumber)
	out.WorkflowStepInstanceID = scanNullableString(step)
	out.AssignedUserID = scanNullableString(assignedUser)
	if assignedRole.Valid {
		out.AssignedRoleID = &assignedRole.Int64
	}
	out.InspectedByUserID = scanNullableString(inspector)
	out.InspectedAt = scanNullableTime(inspected)
	return out, err
}

func qualityAccessClause(viewAll bool) string {
	if viewAll {
		return "TRUE"
	}
	return `(q.assigned_user_id=$1 OR q.created_by_user_id=$1 OR q.inspected_by_user_id=$1 OR EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$1 AND ur.role_id=q.assigned_role_id) OR EXISTS(SELECT 1 FROM workflow_step_instances si JOIN user_roles ur ON ur.role_id=si.responsible_role_id WHERE si.id=q.workflow_step_instance_id AND ur.user_id=$1))`
}

func (s *OperationsService) ListQualityInspections(ctx context.Context, actor, status, orderID, batchID string) ([]QualityInspection, error) {
	viewAll := s.HasPermission(ctx, actor, "quality.view_all")
	rows, err := s.db.QueryContext(ctx, qualitySelect+` WHERE `+qualityAccessClause(viewAll)+` AND ($2='' OR q.status=UPPER($2)) AND ($3='' OR q.order_id=$3::uuid) AND ($4='' OR q.batch_id=$4::uuid) ORDER BY q.created_at DESC`, actor, status, orderID, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []QualityInspection{}
	for rows.Next() {
		x, scanErr := scanQualityInspection(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) GetQualityInspection(ctx context.Context, actor, id string) (QualityInspection, error) {
	viewAll := s.HasPermission(ctx, actor, "quality.view_all")
	out, err := scanQualityInspection(s.db.QueryRowContext(ctx, qualitySelect+` WHERE q.id=$2 AND `+qualityAccessClause(viewAll), actor, id))
	if err != nil {
		return out, err
	}
	if out.WorkflowStepInstanceID != nil {
		rows, queryErr := s.db.QueryContext(ctx, `SELECT field_key,value_json FROM workflow_step_field_values WHERE workflow_step_instance_id=$1 ORDER BY field_key`, *out.WorkflowStepInstanceID)
		if queryErr != nil {
			return out, queryErr
		}
		defer rows.Close()
		out.Values = map[string]json.RawMessage{}
		for rows.Next() {
			var key string
			var raw []byte
			if err = rows.Scan(&key, &raw); err != nil {
				return out, err
			}
			out.Values[key] = json.RawMessage(raw)
		}
	}
	return out, nil
}

func (s *OperationsService) CreateQualityInspection(ctx context.Context, actor, key string, p QualityInspectionPayload) (QualityInspection, error) {
	var out QualityInspection
	p.InspectionType = normalizeCode(p.InspectionType)
	if p.InspectionType == "" {
		p.InspectionType = "GENERAL"
	}
	if !map[string]bool{"GENERAL": true, "INCOMING": true, "IN_PROCESS": true, "FINAL": true, "PACKAGING": true, "INSTALLATION": true, "OTHER": true}[p.InspectionType] {
		return out, ErrValidation
	}
	if strings.TrimSpace(p.OrderID) == "" && p.BatchID == nil && p.WorkflowStepInstanceID == nil {
		return out, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "QC_CREATE", key, p)
	if err != nil {
		return out, err
	}
	if claim.Existing {
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return out, err
		}
		return out, tx.Commit()
	}
	orderID := strings.TrimSpace(p.OrderID)
	if p.BatchID != nil {
		var batchOrder string
		if err = tx.QueryRowContext(ctx, `SELECT order_id FROM fulfillment_batches WHERE id=$1 FOR SHARE`, *p.BatchID).Scan(&batchOrder); err != nil {
			return out, err
		}
		if orderID != "" && orderID != batchOrder {
			return out, conflict("SCOPE_MISMATCH", "batch belongs to another order")
		}
		orderID = batchOrder
	}
	if p.WorkflowStepInstanceID != nil {
		var stepOrder, scopeType, scopeID string
		var isQC bool
		err = tx.QueryRowContext(ctx, `SELECT wi.order_id,wi.scope_type,wi.scope_id,EXISTS(SELECT 1 FROM workflow_instance_field_definitions f WHERE f.workflow_step_instance_id=si.id AND f.field_type='QC_CHECK') OR si.step_code ILIKE '%QC%' OR si.step_code ILIKE '%QUALITY%' FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.id=$1 FOR SHARE`, *p.WorkflowStepInstanceID).Scan(&stepOrder, &scopeType, &scopeID, &isQC)
		if err != nil {
			return out, err
		}
		if !isQC {
			return out, ErrValidation
		}
		if orderID != "" && orderID != stepOrder {
			return out, conflict("SCOPE_MISMATCH", "workflow step belongs to another order")
		}
		orderID = stepOrder
		if scopeType == "BATCH" && p.BatchID == nil {
			p.BatchID = &scopeID
		}
		if p.AssignedUserID == nil && p.AssignedRoleID == nil {
			var user sql.NullString
			var role sql.NullInt64
			_ = tx.QueryRowContext(ctx, `SELECT assigned_user_id,responsible_role_id FROM workflow_step_instances WHERE id=$1`, *p.WorkflowStepInstanceID).Scan(&user, &role)
			p.AssignedUserID = scanNullableString(user)
			if role.Valid {
				p.AssignedRoleID = &role.Int64
			}
		}
	}
	if orderID == "" {
		return out, ErrValidation
	}
	var orderNumber string
	if err = tx.QueryRowContext(ctx, `SELECT order_number FROM orders WHERE id=$1 AND status NOT IN ('CANCELLED','CLOSED') FOR SHARE`, orderID).Scan(&orderNumber); err != nil {
		return out, err
	}
	if p.InventoryLotID != nil {
		var exists bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM inventory_lots WHERE id=$1)`, *p.InventoryLotID).Scan(&exists); err != nil || !exists {
			if err != nil {
				return out, err
			}
			return out, sql.ErrNoRows
		}
	}
	number, err := nextReadableNumberTx(ctx, tx, "QC")
	if err != nil {
		return out, err
	}
	var id string
	if err = tx.QueryRowContext(ctx, `INSERT INTO quality_inspections(inspection_number,order_id,batch_id,inventory_lot_id,workflow_step_instance_id,inspection_type,assigned_user_id,assigned_role_id,notes,created_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,''),$10) RETURNING id`, number, orderID, p.BatchID, p.InventoryLotID, p.WorkflowStepInstanceID, p.InspectionType, p.AssignedUserID, p.AssignedRoleID, p.Notes, actor).Scan(&id); err != nil {
		return out, err
	}
	if p.WorkflowStepInstanceID != nil && len(p.Values) > 0 {
		if err = s.saveValuesTx(ctx, tx, actor, *p.WorkflowStepInstanceID, p.Values, false); err != nil {
			return out, err
		}
	}
	out = QualityInspection{ID: id, InspectionNumber: number, OrderID: orderID, OrderNumber: orderNumber, BatchID: p.BatchID, InventoryLotID: p.InventoryLotID, WorkflowStepInstanceID: p.WorkflowStepInstanceID, InspectionType: p.InspectionType, Status: "PENDING", AssignedUserID: p.AssignedUserID, AssignedRoleID: p.AssignedRoleID, Notes: p.Notes, Values: p.Values, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.auditTx(ctx, tx, actor, "QC_STARTED", "quality_inspection", id, nil, p)
	if err = finishOperationTx(ctx, tx, actor, "QC_CREATE", key, out); err != nil {
		return out, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) decideQualityInspection(ctx context.Context, actor, id, key, operation, decision string, p QualityDecisionPayload) (map[string]any, error) {
	requiresReason := decision == "REWORK_REQUIRED" || decision == "REJECTED" || decision == "OVERRIDDEN"
	if requiresReason && requireReason(p.Reason) != nil {
		return nil, ErrValidation
	}
	permission := "quality.inspect"
	if decision == "REJECTED" {
		permission = "quality.reject"
	}
	if decision == "OVERRIDDEN" {
		if s.HasPermission(ctx, actor, "quality.accept") {
			permission = "quality.accept"
		} else {
			permission = "quality.override"
		}
	}
	if !s.HasPermission(ctx, actor, permission) {
		return nil, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, operation, key, map[string]any{"id": id, "payload": p})
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out map[string]any
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return nil, err
		}
		return out, tx.Commit()
	}
	var current, orderID string
	var batchID, lotID, stepID sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT status,order_id,batch_id,inventory_lot_id,workflow_step_instance_id FROM quality_inspections WHERE id=$1 FOR UPDATE`, id).Scan(&current, &orderID, &batchID, &lotID, &stepID); err != nil {
		return nil, err
	}
	valid := decision == "FAILED" && current == "PENDING" || decision == "PASSED" && current == "PENDING" || (decision == "REWORK_REQUIRED" || decision == "REJECTED" || decision == "OVERRIDDEN") && current == "FAILED"
	if !valid {
		return nil, conflict("INVALID_QC_TRANSITION", "quality inspection is not in the expected state")
	}
	if stepID.Valid && decision == "FAILED" {
		var stepStatus string
		if err = tx.QueryRowContext(ctx, `SELECT status FROM workflow_step_instances WHERE id=$1 FOR UPDATE`, stepID.String).Scan(&stepStatus); err != nil {
			return nil, err
		}
		if stepStatus != "IN_PROGRESS" {
			return nil, conflict("INVALID_QC_TRANSITION", "quality workflow step is not in progress")
		}
		if err = s.saveValuesTx(ctx, tx, actor, stepID.String, p.Values, false); err != nil {
			return nil, err
		}
	}
	status, resultCode, auditAction := decision, "", "QC_FAILED"
	switch decision {
	case "PASSED":
		status, resultCode, auditAction = "PASSED", "APPROVED", "QC_PASSED"
		if strings.TrimSpace(p.Notes) != "" {
			status = "PASSED_WITH_NOTES"
		}
	case "REWORK_REQUIRED":
		resultCode, auditAction = "CORRECTION_REQUIRED", "QC_REWORK_REQUESTED"
	case "REJECTED":
		status, resultCode, auditAction = "FAILED", "REJECTED", "QC_FAILED"
	case "OVERRIDDEN":
		status, resultCode, auditAction = "PASSED_WITH_NOTES", "APPROVED", "QC_OVERRIDDEN"
	}
	if stepID.Valid && decision != "FAILED" {
		if err = s.submitWorkflowStepTx(ctx, tx, actor, stepID.String, StepValuesPayload{Values: p.Values, Reason: p.Reason, ResultCode: resultCode}); err != nil {
			return nil, err
		}
	}
	if decision == "REJECTED" {
		if batchID.Valid {
			if _, err = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET status='QC_REJECTED',updated_at=NOW() WHERE id=$1 AND status NOT IN ('CANCELLED','SPLIT','MERGED','DELIVERED')`, batchID.String); err != nil {
				return nil, err
			}
		}
		if lotID.Valid {
			if _, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET status='QC_REJECTED',updated_at=NOW() WHERE id=$1 AND status NOT IN ('CANCELLED','SOLD','CONSUMED')`, lotID.String); err != nil {
				return nil, err
			}
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE quality_inspections SET status=$2,inspected_by_user_id=$3,inspected_at=NOW(),notes=COALESCE(NULLIF($4,''),notes),decision_reason=NULLIF($5,''),updated_at=NOW() WHERE id=$1`, id, status, actor, p.Notes, p.Reason); err != nil {
		return nil, err
	}
	s.auditTx(ctx, tx, actor, auditAction, "quality_inspection", id, map[string]any{"status": current}, map[string]any{"status": status, "reason": p.Reason})
	out := map[string]any{"id": id, "order_id": orderID, "status": status, "result_code": resultCode}
	if err = finishOperationTx(ctx, tx, actor, operation, key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) PassQualityInspection(ctx context.Context, actor, id, key string, p QualityDecisionPayload) (map[string]any, error) {
	return s.decideQualityInspection(ctx, actor, id, key, "QC_PASS", "PASSED", p)
}

func (s *OperationsService) FailQualityInspection(ctx context.Context, actor, id, key string, p QualityDecisionPayload) (map[string]any, error) {
	return s.decideQualityInspection(ctx, actor, id, key, "QC_FAIL", "FAILED", p)
}

func (s *OperationsService) RequestQualityRework(ctx context.Context, actor, id, key string, p QualityDecisionPayload) (map[string]any, error) {
	return s.decideQualityInspection(ctx, actor, id, key, "QC_REWORK", "REWORK_REQUIRED", p)
}

func (s *OperationsService) RejectQualityInspection(ctx context.Context, actor, id, key string, p QualityDecisionPayload) (map[string]any, error) {
	return s.decideQualityInspection(ctx, actor, id, key, "QC_REJECT", "REJECTED", p)
}

func (s *OperationsService) OverrideQualityInspection(ctx context.Context, actor, id, key string, p QualityDecisionPayload) (map[string]any, error) {
	return s.decideQualityInspection(ctx, actor, id, key, "QC_OVERRIDE", "OVERRIDDEN", p)
}
