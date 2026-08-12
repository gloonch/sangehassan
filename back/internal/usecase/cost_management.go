package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

type CostEntryPayload struct {
	CostPayload
	Status          string  `json:"status"`
	VendorName      string  `json:"vendor_name"`
	VendorReference string  `json:"vendor_reference"`
	InvoiceNumber   string  `json:"invoice_number"`
	InvoiceFileID   *string `json:"invoice_file_id"`
	SupplierID      *string `json:"supplier_id"`
}

func costOrderIDTx(ctx context.Context, tx *sql.Tx, entityType, entityID string) (*string, error) {
	var id sql.NullString
	switch entityType {
	case "ORDER":
		id = sql.NullString{String: entityID, Valid: true}
	case "BATCH":
		if err := tx.QueryRowContext(ctx, `SELECT order_id FROM fulfillment_batches WHERE id=$1`, entityID).Scan(&id); err != nil {
			return nil, err
		}
	case "SHIPMENT":
		if err := tx.QueryRowContext(ctx, `SELECT order_id FROM shipments WHERE id=$1`, entityID).Scan(&id); err != nil {
			return nil, err
		}
	case "INVENTORY_MOVEMENT":
		if err := tx.QueryRowContext(ctx, `SELECT order_id FROM inventory_movements WHERE id=$1`, entityID).Scan(&id); err != nil {
			return nil, err
		}
	case "INSTALLATION":
		if err := tx.QueryRowContext(ctx, `SELECT order_id FROM installation_jobs WHERE id=$1`, entityID).Scan(&id); err != nil {
			return nil, err
		}
	}
	return scanNullableString(id), nil
}

func (s *OperationsService) canUseCostScope(ctx context.Context, actor, entityType, entityID string) bool {
	if s.HasPermission(ctx, actor, "finance.costs.view") || s.HasPermission(ctx, actor, "finance.costs.view_all") {
		return true
	}
	var ok bool
	switch entityType {
	case "SHIPMENT":
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM shipments WHERE id=$1 AND driver_user_id=$2) OR EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$2 AND r.code='SUPPLY')`, entityID, actor).Scan(&ok)
	case "BATCH", "INVENTORY_MOVEMENT":
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$1 AND r.code='SUPPLY')`, actor).Scan(&ok)
	case "INSTALLATION":
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$1 AND r.code='INSTALLATION_LEAD')`, actor).Scan(&ok)
	}
	return ok
}

func (s *OperationsService) CreateCost(ctx context.Context, actor, key string, p CostEntryPayload) (map[string]any, error) {
	p.EntityType = normalizeCode(p.EntityType)
	p.CostType = normalizeCode(p.CostType)
	p.Currency = normalizeCode(p.Currency)
	p.Status = normalizeCode(p.Status)
	if p.Status == "" {
		p.Status = "REPORTED"
	}
	validCostTypes := map[string]bool{"PURCHASE": true, "STONE_PURCHASE": true, "EXTRACTION": true, "MINE_LOADING": true, "LOADING": true, "TRANSPORT": true, "FACTORY_RECEIVING": true, "CUTTING": true, "PROCESSING": true, "FINISHING": true, "QC": true, "QUALITY_CONTROL": true, "PACKAGING": true, "WAREHOUSE": true, "CUSTOMS": true, "PORT": true, "CONTAINER": true, "INSURANCE": true, "INSTALLATION": true, "LABOR": true, "DAMAGE": true, "REWORK": true, "COMMISSION": true, "OTHER": true}
	validCurrencies := map[string]bool{"IRR": true, "USD": true, "EUR": true, "AED": true, "OMR": true}
	if p.Status != "ESTIMATED" && p.Status != "REPORTED" || !validNonNegativeDecimal(p.Amount) || !validCurrencies[p.Currency] || !validCostTypes[p.CostType] {
		return nil, ErrValidation
	}
	if err := s.validateCostEntity(ctx, p.EntityType, p.EntityID); err != nil {
		return nil, err
	}
	if !s.canUseCostScope(ctx, actor, p.EntityType, p.EntityID) {
		return nil, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "COST_CREATE", key, p)
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
	orderID, err := costOrderIDTx(ctx, tx, p.EntityType, p.EntityID)
	if err != nil {
		return nil, err
	}
	if err = ensureActiveSupplierTx(ctx, tx, p.SupplierID); err != nil {
		return nil, err
	}
	incurred := time.Now()
	if p.IncurredAt != nil {
		incurred = *p.IncurredAt
	}
	var id string
	err = tx.QueryRowContext(ctx, `INSERT INTO operational_cost_entries(entity_type,entity_id,cost_type,amount,currency,status,notes,incurred_at,created_by_user_id,order_id,batch_id,shipment_id,installation_id,vendor_name,vendor_reference,invoice_number,invoice_file_id,supplier_id) VALUES($1,$2,$3,$4::numeric,$5,$6,NULLIF($7,''),$8,$9,$10,CASE WHEN $1='BATCH' THEN $2::uuid END,CASE WHEN $1='SHIPMENT' THEN $2::uuid END,CASE WHEN $1='INSTALLATION' THEN $2::uuid END,NULLIF($11,''),NULLIF($12,''),NULLIF($13,''),$14,$15) RETURNING id`, p.EntityType, p.EntityID, p.CostType, p.Amount, p.Currency, p.Status, p.Notes, incurred, actor, orderID, p.VendorName, p.VendorReference, p.InvoiceNumber, p.InvoiceFileID, p.SupplierID).Scan(&id)
	if err != nil {
		return nil, err
	}
	if orderID != nil {
		if err = refreshFinancialSummaryTx(ctx, tx, *orderID); err != nil {
			return nil, err
		}
	}
	out := map[string]any{"id": id, "status": p.Status, "amount": p.Amount, "currency": p.Currency}
	if err = auditTx(ctx, tx, actor, "finance.costs.create", "operational_cost_entry", id, nil, p); err != nil {
		return nil, err
	}
	if err = finishOperationTx(ctx, tx, actor, "COST_CREATE", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) transitionCost(ctx context.Context, actor, id, key, operation, to string, p CostFlowPayload) (map[string]any, error) {
	if (to == "REJECTED" || to == "CANCELLED") && requireReason(p.Reason) != nil {
		return nil, ErrValidation
	}
	if operation == "COST_SUBMIT" {
		var entityType, entityID string
		if err := s.db.QueryRowContext(ctx, `SELECT entity_type,entity_id FROM operational_cost_entries WHERE id=$1`, id).Scan(&entityType, &entityID); err != nil {
			return nil, err
		}
		if !s.canUseCostScope(ctx, actor, entityType, entityID) {
			return nil, ErrForbidden
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
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
	var from string
	var orderID sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT status,order_id FROM operational_cost_entries WHERE id=$1 FOR UPDATE`, id).Scan(&from, &orderID); err != nil {
		return nil, err
	}
	valid := false
	switch to {
	case "PENDING_APPROVAL":
		valid = from == "ESTIMATED" || from == "REPORTED"
	case "APPROVED":
		valid = from == "PENDING_APPROVAL"
	case "REJECTED":
		valid = from == "PENDING_APPROVAL"
	case "PAID":
		valid = from == "APPROVED"
	case "CANCELLED":
		valid = from != "PAID" && from != "CANCELLED"
	}
	if !valid {
		return nil, conflict(ErrInvalidFinancialTransition, "invalid cost transition")
	}
	query := `UPDATE operational_cost_entries SET status=$2,updated_at=NOW()`
	args := []any{id, to}
	switch to {
	case "PENDING_APPROVAL":
		query += `,submitted_by_user_id=$3,submitted_at=NOW()`
		args = append(args, actor)
	case "APPROVED":
		query += `,approved_by_user_id=$3,approved_at=NOW()`
		args = append(args, actor)
	case "REJECTED":
		query += `,rejected_by_user_id=$3,rejected_at=NOW(),rejection_reason=$4`
		args = append(args, actor, p.Reason)
	case "PAID":
		query += `,paid_by_user_id=$3,paid_at=NOW(),payment_reference=NULLIF($4,'')`
		args = append(args, actor, p.PaymentReference)
	case "CANCELLED":
		query += `,cancelled_by_user_id=$3,cancelled_at=NOW(),cancellation_reason=$4`
		args = append(args, actor, p.Reason)
	}
	query += ` WHERE id=$1`
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}
	if to == "PENDING_APPROVAL" {
		_, err = tx.ExecContext(ctx, `INSERT INTO action_items(order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT c.order_id,o.customer_user_id,'بررسی هزینه عملیاتی',c.cost_type,'OPEN','HIGH',r.id,'finance.costs.approve',NOW(),'cost:approve:'||c.id,'COST_APPROVAL_REQUIRED' FROM operational_cost_entries c LEFT JOIN orders o ON o.id=c.order_id JOIN roles r ON r.code='ACCOUNTANT' WHERE c.id=$1 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, id)
		if err != nil {
			return nil, err
		}
		if err = emitNotificationToRoleTx(ctx, tx, "ACCOUNTANT", "COST_APPROVAL_REQUIRED", "cost-approval:"+id, "COST", id, "/panel/dashboard/finance", map[string]string{}); err != nil {
			return nil, err
		}
	}
	if to == "APPROVED" || to == "REJECTED" || to == "CANCELLED" {
		if _, err = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE deduplication_key='cost:approve:'||$1 AND status='OPEN'`, id, actor); err != nil {
			return nil, err
		}
	}
	if orderID.Valid {
		if err = refreshFinancialSummaryTx(ctx, tx, orderID.String); err != nil {
			return nil, err
		}
	}
	out := map[string]any{"id": id, "status": to}
	if err = auditTx(ctx, tx, actor, "finance.costs."+strings.ToLower(to), "operational_cost_entry", id, map[string]string{"status": from}, map[string]any{"status": to, "reason": p.Reason}); err != nil {
		return nil, err
	}
	if err = finishOperationTx(ctx, tx, actor, operation, key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) SubmitCost(ctx context.Context, a, id, key string) (map[string]any, error) {
	return s.transitionCost(ctx, a, id, key, "COST_SUBMIT", "PENDING_APPROVAL", CostFlowPayload{})
}
func (s *OperationsService) ApproveCost(ctx context.Context, a, id, key string) (map[string]any, error) {
	return s.transitionCost(ctx, a, id, key, "COST_APPROVE", "APPROVED", CostFlowPayload{})
}
func (s *OperationsService) RejectCost(ctx context.Context, a, id, key string, p CostFlowPayload) (map[string]any, error) {
	return s.transitionCost(ctx, a, id, key, "COST_REJECT", "REJECTED", p)
}
func (s *OperationsService) MarkCostPaid(ctx context.Context, a, id, key string, p CostFlowPayload) (map[string]any, error) {
	return s.transitionCost(ctx, a, id, key, "COST_PAID", "PAID", p)
}
func (s *OperationsService) CancelCost(ctx context.Context, a, id, key string, p CostFlowPayload) (map[string]any, error) {
	return s.transitionCost(ctx, a, id, key, "COST_CANCEL", "CANCELLED", p)
}

func (s *OperationsService) ListOrderCosts(ctx context.Context, actor, orderID string) ([]map[string]any, error) {
	viewAll := s.HasPermission(ctx, actor, "finance.costs.view") || s.HasPermission(ctx, actor, "finance.costs.view_all") || s.HasPermission(ctx, actor, "finance.operational_costs.view")
	if !viewAll && !s.HasPermission(ctx, actor, "finance.costs.view_assigned") {
		return nil, ErrForbidden
	}
	rows, err := s.db.QueryContext(ctx, `SELECT c.id,c.entity_type,c.entity_id,c.cost_type,c.amount::text,c.currency,c.status,COALESCE(c.vendor_name,''),COALESCE(c.invoice_number,''),COALESCE(c.notes,''),c.incurred_at FROM operational_cost_entries c WHERE c.order_id=$1 AND ($3 OR (c.entity_type='SHIPMENT' AND EXISTS(SELECT 1 FROM shipments s WHERE s.id=c.entity_id AND s.driver_user_id=$2)) OR (c.entity_type IN ('BATCH','INVENTORY_MOVEMENT') AND EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$2 AND r.code='SUPPLY')) OR (c.entity_type='INSTALLATION' AND EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$2 AND r.code='INSTALLATION_LEAD'))) ORDER BY c.incurred_at DESC`, orderID, actor, viewAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, et, eid, ct, amount, currency, status, vendor, invoice, notes string
		var incurred time.Time
		if err = rows.Scan(&id, &et, &eid, &ct, &amount, &currency, &status, &vendor, &invoice, &notes, &incurred); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "entity_type": et, "entity_id": eid, "cost_type": ct, "amount": amount, "currency": currency, "status": status, "vendor_name": vendor, "invoice_number": invoice, "notes": notes, "incurred_at": incurred})
	}
	return out, rows.Err()
}
