package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

func orderClosureWarningsQuery(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, orderID string) ([]map[string]any, error) {
	warnings := []map[string]any{}
	checks := []struct {
		code, message, query string
	}{
		{"OPEN_ACTION_ITEMS", "action items are still open", `SELECT COUNT(*) FROM action_items WHERE order_id=$1 AND status NOT IN ('COMPLETED','CANCELLED')`},
		{"OUTSTANDING_BALANCE", "customer balance is still outstanding", `SELECT COUNT(*) FROM order_financial_summaries WHERE order_id=$1 AND outstanding_amount>0`},
		{"ACTIVE_WORKFLOW", "workflow instances are still active", `SELECT COUNT(*) FROM workflow_instances WHERE order_id=$1 AND status NOT IN ('COMPLETED','CANCELLED')`},
		{"ACTIVE_SHIPMENT", "shipments are not fully delivered", `SELECT COUNT(*) FROM shipments WHERE order_id=$1 AND status NOT IN ('DELIVERED','CANCELLED')`},
		{"PENDING_QUALITY", "quality inspections require attention", `SELECT COUNT(*) FROM quality_inspections WHERE order_id=$1 AND status IN ('PENDING','FAILED','REWORK_REQUIRED')`},
		{"ACTIVE_INSTALLATION", "installation is not completed", `SELECT COUNT(*) FROM installation_jobs WHERE order_id=$1 AND status NOT IN ('COMPLETED','CANCELLED')`},
		{"MISSING_ACCEPTANCE", "final customer acceptance is not recorded", `SELECT COUNT(*) FROM orders o WHERE o.id=$1 AND (o.installation_required OR EXISTS(SELECT 1 FROM shipments s WHERE s.order_id=o.id AND s.status='DELIVERED')) AND NOT EXISTS(SELECT 1 FROM customer_order_acceptances a WHERE a.order_id=o.id AND a.accepted)`},
		{"MISSING_DOCUMENT", "required documents are still missing", `SELECT COUNT(*) FROM workflow_instance_document_requirements r JOIN workflow_instances wi ON wi.id=r.workflow_instance_id WHERE wi.order_id=$1 AND r.is_required AND r.status='PENDING'`},
	}
	for _, check := range checks {
		var count int
		if err := q.QueryRowContext(ctx, check.query, orderID).Scan(&count); err != nil {
			return nil, err
		}
		if count > 0 {
			warnings = append(warnings, map[string]any{"code": check.code, "message": check.message, "count": count})
		}
	}
	return warnings, nil
}

func (s *OperationsService) OrderClosureReadiness(ctx context.Context, orderID string) (map[string]any, error) {
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM orders WHERE id=$1`, orderID).Scan(&status); err != nil {
		return nil, err
	}
	warnings, err := orderClosureWarningsQuery(ctx, s.db, orderID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"order_id": orderID, "status": status, "can_close": status != "CANCELLED", "requires_force": len(warnings) > 0, "warnings": warnings}, nil
}

func (s *OperationsService) CloseOrder(ctx context.Context, actor, orderID, key string, p OrderClosurePayload) (map[string]any, error) {
	if p.Force {
		if !s.HasPermission(ctx, actor, "orders.close_with_warnings") {
			return nil, ErrForbidden
		}
		if requireReason(p.Reason) != nil {
			return nil, ErrValidation
		}
	} else if !s.HasPermission(ctx, actor, "orders.close") {
		return nil, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "ORDER_CLOSE", key, map[string]any{"order_id": orderID, "payload": p})
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
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&status); err != nil {
		return nil, err
	}
	if status == "CANCELLED" {
		return nil, conflict("INVALID_ORDER_TRANSITION", "cancelled order cannot be closed")
	}
	if status == "CLOSED" {
		out := map[string]any{"order_id": orderID, "status": "CLOSED", "already_closed": true, "warnings": []any{}}
		if err = finishOperationTx(ctx, tx, actor, "ORDER_CLOSE", key, out); err != nil {
			return nil, err
		}
		return out, tx.Commit()
	}
	warnings, err := orderClosureWarningsQuery(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 && !p.Force {
		codes := make([]string, 0, len(warnings))
		for _, warning := range warnings {
			codes = append(codes, fmt.Sprint(warning["code"]))
		}
		return nil, conflict("ORDER_CLOSE_WARNINGS", strings.Join(codes, ","))
	}
	if _, err = tx.ExecContext(ctx, `UPDATE orders SET status='CLOSED',closed_at=NOW(),closed_by_user_id=$2,closure_forced=$3,closure_reason=NULLIF($4,''),updated_at=NOW() WHERE id=$1`, orderID, actor, p.Force, p.Reason); err != nil {
		return nil, err
	}
	if err = refreshFinancialSummaryTx(ctx, tx, orderID); err != nil {
		return nil, err
	}
	action := "ORDER_CLOSED"
	if p.Force {
		action = "ORDER_FORCE_CLOSED"
	}
	s.auditTx(ctx, tx, actor, action, "order", orderID, map[string]any{"status": status}, map[string]any{"status": "CLOSED", "forced": p.Force, "reason": p.Reason, "warnings": warnings})
	out := map[string]any{"order_id": orderID, "status": "CLOSED", "forced": p.Force, "warnings": warnings}
	if err = finishOperationTx(ctx, tx, actor, "ORDER_CLOSE", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}
