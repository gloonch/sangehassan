package usecase

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"math/big"
	"sort"
	"strings"
	"time"
)

func clampPercent(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return math.Round(v*100) / 100
}
func ratioPercent(value, ordered *big.Rat) float64 {
	if ordered == nil || ordered.Sign() <= 0 {
		return 0
	}
	f, _ := new(big.Rat).Mul(new(big.Rat).Quo(value, ordered), big.NewRat(100, 1)).Float64()
	return clampPercent(f)
}
func maxRat(values ...*big.Rat) *big.Rat {
	out := new(big.Rat)
	for _, v := range values {
		if v != nil && v.Cmp(out) > 0 {
			out.Set(v)
		}
	}
	return out
}

func (s *OperationsService) sumOrderItemQuantityTx(ctx context.Context, tx *sql.Tx, itemID, orderUnit, query string) (*big.Rat, error) {
	rows, err := tx.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	total := new(big.Rat)
	for rows.Next() {
		var q, u string
		if err = rows.Scan(&q, &u); err != nil {
			return nil, err
		}
		converted, e := convertQuantityTx(ctx, tx, itemID, q, u, orderUnit)
		if e != nil {
			return nil, e
		}
		total.Add(total, converted)
	}
	return total, rows.Err()
}

func (s *OperationsService) OrderProgress(ctx context.Context, actor, orderID string, customer bool) (OrderProgress, error) {
	var out OrderProgress
	if customer {
		var owner string
		if err := s.db.QueryRowContext(ctx, `SELECT customer_user_id FROM orders WHERE id=$1`, orderID).Scan(&owner); err != nil {
			return out, err
		}
		if owner != actor {
			return out, ErrForbidden
		}
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	if err = tx.QueryRowContext(ctx, `SELECT order_number FROM orders WHERE id=$1`, orderID).Scan(&out.OrderNumber); err != nil {
		return out, err
	}
	out.OrderID = orderID
	rows, err := tx.QueryContext(ctx, `SELECT id,stone_name,ordered_quantity::text,quantity_unit,progress_weight::text,requires_production,requires_packaging FROM order_items WHERE order_id=$1 ORDER BY created_at`, orderID)
	if err != nil {
		return out, err
	}
	type base struct {
		id, name, ordered, unit, weight string
		production, packaging           bool
	}
	bases := []base{}
	for rows.Next() {
		var b base
		if err = rows.Scan(&b.id, &b.name, &b.ordered, &b.unit, &b.weight, &b.production, &b.packaging); err != nil {
			return out, err
		}
		bases = append(bases, b)
	}
	rows.Close()
	out.Items = []OrderItemProgress{}
	weighted, totalWeight := 0.0, 0.0
	for _, b := range bases {
		ordered, _ := new(big.Rat).SetString(b.ordered)
		planned, e := s.sumOrderItemQuantityTx(ctx, tx, b.id, b.unit, `SELECT planned_quantity::text,quantity_unit FROM fulfillment_batches WHERE order_item_id=$1 AND status NOT IN ('SPLIT','MERGED','CANCELLED')`)
		if e != nil {
			return out, e
		}
		reserved, e := s.sumOrderItemQuantityTx(ctx, tx, b.id, b.unit, `SELECT (r.reserved_quantity-r.consumed_quantity)::text,r.quantity_unit FROM inventory_reservations r WHERE r.order_item_id=$1 AND r.status IN ('ACTIVE','PARTIALLY_CONSUMED')`)
		if e != nil {
			return out, e
		}
		inProduction, e := s.sumOrderItemQuantityTx(ctx, tx, b.id, b.unit, `SELECT planned_quantity::text,quantity_unit FROM fulfillment_batches WHERE order_item_id=$1 AND status IN ('IN_PRODUCTION','READY_FOR_QC')`)
		if e != nil {
			return out, e
		}
		produced, e := s.sumOrderItemQuantityTx(ctx, tx, b.id, b.unit, `SELECT actual_quantity::text,quantity_unit FROM fulfillment_batches WHERE order_item_id=$1 AND status IN ('READY_FOR_QC','QC_APPROVED','READY_FOR_PACKAGING','PACKAGED','READY_FOR_SHIPMENT','PARTIALLY_SHIPPED','SHIPPED','PARTIALLY_DELIVERED','DELIVERED')`)
		if e != nil {
			return out, e
		}
		packaged, e := s.sumOrderItemQuantityTx(ctx, tx, b.id, b.unit, `SELECT p.quantity::text,p.quantity_unit FROM packaging_units p JOIN fulfillment_batches b ON b.id=p.batch_id WHERE b.order_item_id=$1 AND p.status NOT IN ('DRAFT','CANCELLED','DAMAGED')`)
		if e != nil {
			return out, e
		}
		shipped, e := s.sumOrderItemQuantityTx(ctx, tx, b.id, b.unit, `SELECT si.loaded_quantity::text,si.quantity_unit FROM shipment_items si JOIN fulfillment_batches b ON b.id=si.batch_id WHERE b.order_item_id=$1 AND si.loaded_quantity>0`)
		if e != nil {
			return out, e
		}
		delivered, e := s.sumOrderItemQuantityTx(ctx, tx, b.id, b.unit, `SELECT si.delivered_quantity::text,si.quantity_unit FROM shipment_items si JOIN fulfillment_batches b ON b.id=si.batch_id WHERE b.order_item_id=$1 AND si.delivered_quantity>0`)
		if e != nil {
			return out, e
		}
		procured := maxRat(new(big.Rat).Add(new(big.Rat).Set(reserved), inProduction), produced, packaged, shipped, delivered)
		remaining := new(big.Rat).Sub(new(big.Rat).Set(ordered), delivered)
		if remaining.Sign() < 0 {
			remaining.SetInt64(0)
		}
		procP := ratioPercent(procured, ordered)
		prodP := ratioPercent(produced, ordered)
		packP := ratioPercent(packaged, ordered)
		shipP := ratioPercent(shipped, ordered)
		delP := ratioPercent(delivered, ordered)
		stageSum, stageWeight := procP*20.0, 20.0
		if b.production {
			stageSum += prodP * 30
			stageWeight += 30
		}
		if b.packaging {
			stageSum += packP * 15
			stageWeight += 15
		}
		stageSum += shipP*15 + delP*20
		stageWeight += 35
		overall := clampPercent(stageSum / stageWeight)
		weightRat, _ := new(big.Rat).SetString(b.weight)
		weight, _ := weightRat.Float64()
		weighted += overall * weight
		totalWeight += weight
		item := OrderItemProgress{OrderItemID: b.id, StoneName: b.name, OrderedQuantity: b.ordered, QuantityUnit: b.unit, Planned: ProgressStage{ratString(planned), ratioPercent(planned, ordered)}, Reserved: ProgressStage{ratString(reserved), ratioPercent(reserved, ordered)}, InProduction: ProgressStage{ratString(inProduction), ratioPercent(inProduction, ordered)}, Produced: ProgressStage{ratString(produced), prodP}, Packaged: ProgressStage{ratString(packaged), packP}, Shipped: ProgressStage{ratString(shipped), shipP}, Delivered: ProgressStage{ratString(delivered), delP}, RemainingQuantity: ratString(remaining), ProcurementProgress: procP, ProductionProgress: prodP, PackagingProgress: packP, ShippingProgress: shipP, DeliveryProgress: delP, OverallProgress: overall, ProgressWeight: b.weight}
		out.Items = append(out.Items, item)
	}
	if totalWeight > 0 {
		out.OverallProgress = clampPercent(weighted / totalWeight)
	}
	if err = tx.Commit(); err != nil {
		return out, err
	}
	return out, nil
}

func (s *OperationsService) validateCostEntity(ctx context.Context, entityType, entityID string) error {
	table := map[string]string{"ORDER": "orders", "BATCH": "fulfillment_batches", "SHIPMENT": "shipments", "INVENTORY_MOVEMENT": "inventory_movements", "INSTALLATION": "installation_jobs"}[entityType]
	if table == "" {
		return ErrValidation
	}
	var ok bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM `+table+` WHERE id=$1)`, entityID).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return sql.ErrNoRows
	}
	return nil
}
func (s *OperationsService) CreateOperationalCost(ctx context.Context, actor string, p CostPayload) (map[string]any, error) {
	return s.CreateCost(ctx, actor, "legacy-"+randomUUIDText(), CostEntryPayload{CostPayload: p, Status: "REPORTED"})
}
func (s *OperationsService) ListOperationalCosts(ctx context.Context, actor, entityType, entityID string) ([]map[string]any, error) {
	if !s.HasPermission(ctx, actor, "finance.operational_costs.view") && !s.HasPermission(ctx, actor, "finance.transport_costs.view") {
		return nil, ErrForbidden
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,entity_type,entity_id,cost_type,amount::text,currency,status,COALESCE(notes,''),incurred_at,created_by_user_id,approved_by_user_id,approved_at FROM operational_cost_entries WHERE ($1='' OR entity_type=$1) AND ($2='' OR entity_id=$2::uuid) ORDER BY incurred_at DESC`, normalizeCode(entityType), entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, et, eid, ct, amount, currency, status, notes, created string
		var incurred time.Time
		var approved sql.NullString
		var approvedAt sql.NullTime
		if err = rows.Scan(&id, &et, &eid, &ct, &amount, &currency, &status, &notes, &incurred, &created, &approved, &approvedAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "entity_type": et, "entity_id": eid, "cost_type": ct, "amount": amount, "currency": currency, "status": status, "notes": notes, "incurred_at": incurred, "created_by_user_id": created, "approved_by_user_id": scanNullableString(approved), "approved_at": nullableTime(approvedAt)})
	}
	return out, rows.Err()
}
func nullableTime(v sql.NullTime) any {
	if !v.Valid {
		return nil
	}
	return v.Time
}
func (s *OperationsService) ApproveOperationalCost(ctx context.Context, actor, id string) error {
	r, err := s.db.ExecContext(ctx, `UPDATE operational_cost_entries SET status='APPROVED',approved_by_user_id=$2,approved_at=NOW(),submitted_by_user_id=COALESCE(submitted_by_user_id,created_by_user_id),submitted_at=COALESCE(submitted_at,NOW()),updated_at=NOW() WHERE id=$1 AND status IN ('REPORTED','PENDING_APPROVAL')`, id, actor)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict("INVALID_COST_STATE", "cost is not approvable")
	}
	s.audit(ctx, actor, "operational_costs.approve", "operational_cost_entry", id, nil)
	return nil
}
func (s *OperationsService) VoidOperationalCost(ctx context.Context, actor, id, reason string) error {
	if requireReason(reason) != nil {
		return ErrValidation
	}
	r, err := s.db.ExecContext(ctx, `UPDATE operational_cost_entries SET status='CANCELLED',cancelled_at=NOW(),cancellation_reason=$2,updated_at=NOW() WHERE id=$1 AND status NOT IN ('CANCELLED','PAID')`, id, reason)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict("INVALID_COST_STATE", "cost is already void")
	}
	s.audit(ctx, actor, "operational_costs.void", "operational_cost_entry", id, map[string]string{"reason": reason})
	return nil
}

func (s *OperationsService) OperationsDashboardSummary(ctx context.Context, actor string) (map[string]any, error) {
	out := map[string]any{}
	queries := map[string]string{}
	_, permissions, err := s.authorization(ctx, actor)
	if err != nil {
		return nil, err
	}
	permissionSet := make(map[string]bool, len(permissions))
	for _, permission := range permissions {
		permissionSet[permission] = true
	}
	has := func(code string) bool { return permissionSet[code] }
	if has("batches.view_all") {
		queries["active_batches"] = `SELECT COUNT(*) FROM fulfillment_batches WHERE status NOT IN ('DELIVERED','CANCELLED','SPLIT','MERGED')`
		queries["batches_without_source"] = `SELECT COUNT(*) FROM fulfillment_batches WHERE status IN ('DRAFT','PLANNED') AND source_location_id IS NULL`
		queries["batches_without_reservation"] = `SELECT COUNT(*) FROM fulfillment_batches b WHERE b.status NOT IN ('SPLIT','MERGED','CANCELLED','DELIVERED') AND NOT EXISTS(SELECT 1 FROM inventory_reservations r WHERE r.batch_id=b.id AND r.status IN ('ACTIVE','PARTIALLY_CONSUMED'))`
		queries["late_production"] = `SELECT COUNT(*) FROM fulfillment_batches WHERE estimated_ready_at<NOW() AND status NOT IN ('READY_FOR_SHIPMENT','PARTIALLY_SHIPPED','SHIPPED','PARTIALLY_DELIVERED','DELIVERED','CANCELLED','SPLIT','MERGED')`
		queries["ready_for_shipment"] = `SELECT COUNT(*) FROM fulfillment_batches WHERE status='READY_FOR_SHIPMENT'`
	} else if has("batches.view_assigned") {
		var n int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM fulfillment_batches b WHERE b.status NOT IN ('DELIVERED','CANCELLED','SPLIT','MERGED') AND EXISTS(SELECT 1 FROM workflow_step_instances si JOIN user_roles ur ON ur.role_id=si.responsible_role_id WHERE si.workflow_instance_id=b.workflow_instance_id AND ur.user_id=$1)`, actor).Scan(&n); err != nil {
			return nil, err
		}
		out["assigned_batches"] = n
	}
	if has("shipments.view_all") {
		queries["active_shipments"] = `SELECT COUNT(*) FROM shipments WHERE status NOT IN ('DELIVERED','CANCELLED')`
		queries["shipments_without_driver"] = `SELECT COUNT(*) FROM shipments WHERE driver_user_id IS NULL AND COALESCE(external_driver_name,'')='' AND status NOT IN ('DELIVERED','CANCELLED')`
		queries["shipments_today"] = `SELECT COUNT(*) FROM shipments WHERE planned_departure_at::date=CURRENT_DATE`
		queries["late_shipments"] = `SELECT COUNT(*) FROM shipments WHERE estimated_arrival_at<NOW() AND status NOT IN ('DELIVERED','CANCELLED')`
		queries["shipment_discrepancies"] = `SELECT COUNT(*) FROM shipments WHERE status='HAS_DISCREPANCY'`
		queries["partial_deliveries"] = `SELECT COUNT(*) FROM shipments WHERE status='PARTIALLY_DELIVERED'`
		queries["customs_waiting"] = `SELECT COUNT(*) FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE wi.scope_type='SHIPMENT' AND si.step_code='PORT_CUSTOMS' AND si.status IN ('WAITING_FOR_ASSIGNEE','IN_PROGRESS','BLOCKED')`
	} else if has("shipments.view_assigned") {
		var active, partial int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(*) FILTER(WHERE status='PARTIALLY_DELIVERED') FROM shipments WHERE driver_user_id=$1 AND status NOT IN ('DELIVERED','CANCELLED')`, actor).Scan(&active, &partial); err != nil {
			return nil, err
		}
		out["assigned_shipments"], out["partial_deliveries"] = active, partial
	}
	if has("orders.view_all") {
		queries["remaining_orders"] = `SELECT COUNT(DISTINCT o.id) FROM orders o JOIN order_items i ON i.order_id=o.id WHERE i.ordered_quantity>COALESCE((SELECT SUM(si.delivered_quantity) FROM fulfillment_batches b JOIN shipment_items si ON si.batch_id=b.id WHERE b.order_item_id=i.id AND si.quantity_unit=i.quantity_unit),0)`
	}
	if has("containers.view") {
		queries["containers_ready"] = `SELECT COUNT(*) FROM shipment_containers WHERE loaded_at IS NOT NULL AND verified_by_user_id IS NOT NULL`
	}
	if has("purchases.view_all") {
		queries["open_purchases"] = `SELECT COUNT(*) FROM purchase_records WHERE status IN ('DRAFT','CONFIRMED','PARTIALLY_RECEIVED')`
		queries["overdue_purchases"] = `SELECT COUNT(*) FROM purchase_records WHERE expected_at<NOW() AND status IN ('CONFIRMED','PARTIALLY_RECEIVED')`
	} else if has("purchases.view_assigned") {
		var n int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM purchase_records p WHERE p.status IN ('DRAFT','CONFIRMED','PARTIALLY_RECEIVED') AND (p.assigned_user_id=$1 OR EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$1 AND ur.role_id=p.assigned_role_id))`, actor).Scan(&n); err != nil {
			return nil, err
		}
		out["assigned_purchases"] = n
	}
	if has("quality.view_all") {
		queries["pending_quality_inspections"] = `SELECT COUNT(*) FROM quality_inspections WHERE status='PENDING'`
		queries["failed_quality_inspections"] = `SELECT COUNT(*) FROM quality_inspections WHERE status='FAILED'`
		queries["quality_rework"] = `SELECT COUNT(*) FROM quality_inspections WHERE status='REWORK_REQUIRED'`
	} else if has("quality.view_assigned") {
		var n int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quality_inspections q WHERE q.status IN ('PENDING','FAILED','REWORK_REQUIRED') AND (q.assigned_user_id=$1 OR EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$1 AND ur.role_id=q.assigned_role_id))`, actor).Scan(&n); err != nil {
			return nil, err
		}
		out["assigned_quality_inspections"] = n
	}
	if has("installation.view_all") {
		queries["active_installations"] = `SELECT COUNT(*) FROM installation_jobs WHERE status NOT IN ('COMPLETED','CANCELLED')`
		queries["installations_today"] = `SELECT COUNT(*) FROM installation_jobs WHERE planned_start_at::date=CURRENT_DATE AND status NOT IN ('COMPLETED','CANCELLED')`
		queries["overdue_installations"] = `SELECT COUNT(*) FROM installation_jobs WHERE estimated_end_at<NOW() AND status NOT IN ('COMPLETED','CANCELLED')`
		queries["open_installation_issues"] = `SELECT COUNT(*) FROM installation_issues WHERE status='OPEN'`
	} else if has("installation.view_assigned") {
		var jobs, issues int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT j.id),COUNT(DISTINCT i.id) FILTER(WHERE i.status='OPEN') FROM installation_jobs j LEFT JOIN installation_job_members m ON m.installation_job_id=j.id LEFT JOIN installation_issues i ON i.installation_job_id=j.id WHERE j.status NOT IN ('COMPLETED','CANCELLED') AND (j.installation_lead_user_id=$1 OR m.user_id=$1)`, actor).Scan(&jobs, &issues); err != nil {
			return nil, err
		}
		out["assigned_installations"], out["open_installation_issues"] = jobs, issues
	}
	if has("orders.close") || has("orders.close_with_warnings") {
		queries["orders_ready_to_close"] = `SELECT COUNT(*) FROM orders o WHERE o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED') AND NOT EXISTS(SELECT 1 FROM action_items a WHERE a.order_id=o.id AND a.status IN ('OPEN','IN_PROGRESS')) AND NOT EXISTS(SELECT 1 FROM workflow_instances w WHERE w.order_id=o.id AND w.status NOT IN ('COMPLETED','CANCELLED')) AND NOT EXISTS(SELECT 1 FROM shipments s WHERE s.order_id=o.id AND s.status NOT IN ('DELIVERED','CANCELLED')) AND NOT EXISTS(SELECT 1 FROM installation_jobs j WHERE j.order_id=o.id AND j.status NOT IN ('COMPLETED','CANCELLED'))`
	}
	if len(queries) > 0 {
		keys := make([]string, 0, len(queries))
		for key := range queries {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		values := make([]int64, len(keys))
		targets := make([]any, len(keys))
		for index, key := range keys {
			parts = append(parts, "("+queries[key]+")")
			targets[index] = &values[index]
		}
		if err := s.db.QueryRowContext(ctx, "SELECT "+strings.Join(parts, ",")).Scan(targets...); err != nil {
			return nil, err
		}
		for index, key := range keys {
			out[key] = values[index]
		}
	}
	if has("inventory.lots.view") {
		inventory, err := s.InventoryDashboard(ctx, actor)
		if err != nil {
			return nil, err
		}
		for k, v := range inventory {
			out[k] = v
		}
	}
	return out, nil
}

var _ = errors.Is
var _ = strings.TrimSpace
