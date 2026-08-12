package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var supplierTypes = map[string]bool{
	"MINE": true, "FACTORY": true, "WORKSHOP": true, "WAREHOUSE": true,
	"STONE_DEALER": true, "TRANSPORT_COMPANY": true, "INSTALLATION_CONTRACTOR": true,
	"PACKAGING_CONTRACTOR": true, "OTHER": true,
}

func validateSupplierPayload(p SupplierPayload) (SupplierPayload, error) {
	p.SupplierType = normalizeCode(p.SupplierType)
	p.CountryCode = normalizeCode(p.CountryCode)
	if p.CountryCode == "" {
		p.CountryCode = "IR"
	}
	if strings.TrimSpace(p.Name) == "" || !supplierTypes[p.SupplierType] || len(p.CountryCode) != 2 {
		return p, ErrValidation
	}
	if p.Phone != "" {
		p.Phone = NormalizePhone(p.Phone)
		if p.Phone == "" {
			return p, ErrValidation
		}
	}
	if p.SecondaryPhone != "" {
		p.SecondaryPhone = NormalizePhone(p.SecondaryPhone)
		if p.SecondaryPhone == "" {
			return p, ErrValidation
		}
	}
	return p, nil
}

func scanSupplier(row interface{ Scan(...any) error }) (Supplier, error) {
	var x Supplier
	err := row.Scan(&x.ID, &x.SupplierCode, &x.Name, &x.SupplierType, &x.Phone, &x.SecondaryPhone, &x.ContactName, &x.Address, &x.City, &x.Province, &x.CountryCode, &x.Notes, &x.IsActive, &x.CreatedAt, &x.UpdatedAt)
	return x, err
}

const supplierSelect = `SELECT id,supplier_code,name,supplier_type,COALESCE(phone,''),COALESCE(secondary_phone,''),COALESCE(contact_name,''),COALESCE(address,''),COALESCE(city,''),COALESCE(province,''),country_code,COALESCE(notes,''),is_active,created_at,updated_at FROM suppliers`

func ensureActiveSupplierTx(ctx context.Context, tx *sql.Tx, supplierID *string) error {
	if supplierID == nil || strings.TrimSpace(*supplierID) == "" {
		return nil
	}
	var active bool
	if err := tx.QueryRowContext(ctx, `SELECT is_active FROM suppliers WHERE id=$1 FOR SHARE`, *supplierID).Scan(&active); err != nil {
		return err
	}
	if !active {
		return conflict("INACTIVE_SUPPLIER", "supplier is inactive")
	}
	return nil
}

func (s *OperationsService) ListSuppliers(ctx context.Context, search, kind string, includeInactive bool) ([]Supplier, error) {
	rows, err := s.db.QueryContext(ctx, supplierSelect+` WHERE ($1='' OR name ILIKE '%'||$1||'%' OR supplier_code ILIKE '%'||$1||'%' OR phone ILIKE '%'||$1||'%') AND ($2='' OR supplier_type=UPPER($2)) AND ($3 OR is_active) ORDER BY is_active DESC,name`, strings.TrimSpace(search), strings.TrimSpace(kind), includeInactive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Supplier{}
	for rows.Next() {
		x, scanErr := scanSupplier(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		x.Notes = ""
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) CreateSupplier(ctx context.Context, actor string, p SupplierPayload) (Supplier, error) {
	var out Supplier
	p, err := validateSupplierPayload(p)
	if err != nil {
		return out, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	number, err := nextReadableNumberTx(ctx, tx, "SUP")
	if err != nil {
		return out, err
	}
	out, err = scanSupplier(tx.QueryRowContext(ctx, `INSERT INTO suppliers(supplier_code,name,supplier_type,phone,secondary_phone,contact_name,address,city,province,country_code,notes,created_by_user_id) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),$10,NULLIF($11,''),$12) RETURNING id,supplier_code,name,supplier_type,COALESCE(phone,''),COALESCE(secondary_phone,''),COALESCE(contact_name,''),COALESCE(address,''),COALESCE(city,''),COALESCE(province,''),country_code,COALESCE(notes,''),is_active,created_at,updated_at`, number, strings.TrimSpace(p.Name), p.SupplierType, p.Phone, p.SecondaryPhone, p.ContactName, p.Address, p.City, p.Province, p.CountryCode, p.Notes, actor))
	if err != nil {
		return out, err
	}
	s.auditTx(ctx, tx, actor, "SUPPLIER_CREATED", "supplier", out.ID, nil, p)
	return out, tx.Commit()
}

func (s *OperationsService) UpdateSupplier(ctx context.Context, actor, id string, p SupplierPayload) (Supplier, error) {
	var out Supplier
	p, err := validateSupplierPayload(p)
	if err != nil {
		return out, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	var before Supplier
	before, err = scanSupplier(tx.QueryRowContext(ctx, supplierSelect+` WHERE id=$1 FOR UPDATE`, id))
	if err != nil {
		return out, err
	}
	out, err = scanSupplier(tx.QueryRowContext(ctx, `UPDATE suppliers SET name=$2,supplier_type=$3,phone=NULLIF($4,''),secondary_phone=NULLIF($5,''),contact_name=NULLIF($6,''),address=NULLIF($7,''),city=NULLIF($8,''),province=NULLIF($9,''),country_code=$10,notes=NULLIF($11,''),updated_at=NOW() WHERE id=$1 RETURNING id,supplier_code,name,supplier_type,COALESCE(phone,''),COALESCE(secondary_phone,''),COALESCE(contact_name,''),COALESCE(address,''),COALESCE(city,''),COALESCE(province,''),country_code,COALESCE(notes,''),is_active,created_at,updated_at`, id, strings.TrimSpace(p.Name), p.SupplierType, p.Phone, p.SecondaryPhone, p.ContactName, p.Address, p.City, p.Province, p.CountryCode, p.Notes))
	if err != nil {
		return out, err
	}
	s.auditTx(ctx, tx, actor, "SUPPLIER_UPDATED", "supplier", id, before, out)
	return out, tx.Commit()
}

func (s *OperationsService) DisableSupplier(ctx context.Context, actor, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var name string
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT name,is_active FROM suppliers WHERE id=$1 FOR UPDATE`, id).Scan(&name, &active); err != nil {
		return err
	}
	if active {
		if _, err = tx.ExecContext(ctx, `UPDATE suppliers SET is_active=FALSE,updated_at=NOW() WHERE id=$1`, id); err != nil {
			return err
		}
		s.auditTx(ctx, tx, actor, "SUPPLIER_UPDATED", "supplier", id, map[string]any{"name": name, "is_active": true}, map[string]any{"is_active": false})
	}
	return tx.Commit()
}

func (s *OperationsService) GetSupplier(ctx context.Context, id string) (Supplier, error) {
	out, err := scanSupplier(s.db.QueryRowContext(ctx, supplierSelect+` WHERE id=$1`, id))
	if err != nil {
		return out, err
	}
	var totalPurchases, completedOrders, qcFailures, totalBatches, totalCosts int
	var averageDelay sql.NullFloat64
	var lastCooperation sql.NullTime
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*),AVG(EXTRACT(EPOCH FROM (received_at-expected_at))/3600) FILTER(WHERE received_at IS NOT NULL AND expected_at IS NOT NULL),MAX(COALESCE(received_at,created_at)) FROM purchase_records WHERE supplier_id=$1`, id).Scan(&totalPurchases, &averageDelay, &lastCooperation)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT o.id) FROM orders o WHERE o.status IN ('COMPLETED','CLOSED') AND (EXISTS(SELECT 1 FROM purchase_records p WHERE p.order_id=o.id AND p.supplier_id=$1) OR EXISTS(SELECT 1 FROM fulfillment_batches b WHERE b.order_id=o.id AND b.supplier_id=$1) OR EXISTS(SELECT 1 FROM shipments sh WHERE sh.order_id=o.id AND sh.supplier_id=$1))`, id).Scan(&completedOrders)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM fulfillment_batches WHERE supplier_id=$1`, id).Scan(&totalBatches)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM operational_cost_entries WHERE supplier_id=$1 AND status<>'CANCELLED'`, id).Scan(&totalCosts)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quality_inspections q LEFT JOIN fulfillment_batches b ON b.id=q.batch_id LEFT JOIN inventory_lots l ON l.id=q.inventory_lot_id WHERE (b.supplier_id=$1 OR l.supplier_id=$1) AND q.status IN ('FAILED','REWORK_REQUIRED')`, id).Scan(&qcFailures)
	stats := map[string]any{"total_purchases": totalPurchases, "total_batches": totalBatches, "total_cost_entries": totalCosts, "total_purchase_amount_by_currency": []map[string]string{}, "average_delivery_delay_hours": nil, "qc_failure_count": qcFailures, "completed_orders_count": completedOrders, "last_cooperation_at": nil}
	if averageDelay.Valid {
		stats["average_delivery_delay_hours"] = averageDelay.Float64
	}
	if lastCooperation.Valid {
		stats["last_cooperation_at"] = lastCooperation.Time
	}
	rows, queryErr := s.db.QueryContext(ctx, `SELECT currency_code,SUM(total_amount)::text FROM purchase_records WHERE supplier_id=$1 AND status<>'CANCELLED' GROUP BY currency_code ORDER BY currency_code`, id)
	if queryErr != nil {
		return out, queryErr
	}
	amounts := []map[string]string{}
	for rows.Next() {
		var currency, amount string
		if err = rows.Scan(&currency, &amount); err != nil {
			rows.Close()
			return out, err
		}
		amounts = append(amounts, map[string]string{"currency": currency, "amount": amount})
	}
	if err = rows.Close(); err != nil {
		return out, err
	}
	stats["total_purchase_amount_by_currency"] = amounts
	out.Statistics = stats
	return out, nil
}

func purchaseAccessClause(viewAll bool) string {
	if viewAll {
		return "TRUE"
	}
	return `(p.assigned_user_id=$1 OR p.created_by_user_id=$1 OR EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$1 AND ur.role_id=p.assigned_role_id))`
}

func scanPurchase(row interface{ Scan(...any) error }) (Purchase, error) {
	var x Purchase
	var batch, assignedUser, cost sql.NullString
	var assignedRole sql.NullInt64
	var expected, received sql.NullTime
	err := row.Scan(&x.ID, &x.PurchaseNumber, &x.OrderID, &x.OrderNumber, &batch, &x.SupplierID, &x.SupplierName, &assignedUser, &assignedRole, &x.StoneName, &x.Description, &x.Quantity, &x.ReceivedQuantity, &x.QuantityUnit, &x.UnitPrice, &x.TotalAmount, &x.CurrencyCode, &x.Status, &expected, &received, &x.Notes, &cost, &x.CreatedAt, &x.UpdatedAt)
	x.BatchID = scanNullableString(batch)
	x.AssignedUserID = scanNullableString(assignedUser)
	if assignedRole.Valid {
		x.AssignedRoleID = &assignedRole.Int64
	}
	x.ExpectedAt = scanNullableTime(expected)
	x.ReceivedAt = scanNullableTime(received)
	x.CostEntryID = scanNullableString(cost)
	return x, err
}

const purchaseSelect = `SELECT p.id,p.purchase_number,p.order_id,o.order_number,p.batch_id,p.supplier_id,s.name,p.assigned_user_id,p.assigned_role_id,p.stone_name,COALESCE(p.description,''),p.quantity::text,p.received_quantity::text,p.quantity_unit,p.unit_price::text,p.total_amount::text,p.currency_code,p.status,p.expected_at,p.received_at,COALESCE(p.notes,''),p.cost_entry_id,p.created_at,p.updated_at FROM purchase_records p JOIN orders o ON o.id=p.order_id JOIN suppliers s ON s.id=p.supplier_id`

func (s *OperationsService) ListPurchases(ctx context.Context, actor, status, orderID, supplierID string) ([]Purchase, error) {
	viewAll := s.HasPermission(ctx, actor, "purchases.view_all")
	rows, err := s.db.QueryContext(ctx, purchaseSelect+` WHERE `+purchaseAccessClause(viewAll)+` AND ($2='' OR p.status=UPPER($2)) AND ($3='' OR p.order_id=$3::uuid) AND ($4='' OR p.supplier_id=$4::uuid) ORDER BY p.created_at DESC`, actor, status, orderID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Purchase{}
	for rows.Next() {
		x, scanErr := scanPurchase(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) GetPurchase(ctx context.Context, actor, id string) (Purchase, error) {
	viewAll := s.HasPermission(ctx, actor, "purchases.view_all")
	out, err := scanPurchase(s.db.QueryRowContext(ctx, purchaseSelect+` WHERE p.id=$2 AND `+purchaseAccessClause(viewAll), actor, id))
	if err != nil {
		return out, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,quantity::text,quantity_unit,inventory_lot_id,inventory_movement_id,COALESCE(notes,''),received_at FROM purchase_receipts WHERE purchase_record_id=$1 ORDER BY received_at DESC`, id)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	out.Receipts = []PurchaseReceipt{}
	for rows.Next() {
		var x PurchaseReceipt
		var lot, movement sql.NullString
		if err = rows.Scan(&x.ID, &x.Quantity, &x.QuantityUnit, &lot, &movement, &x.Notes, &x.ReceivedAt); err != nil {
			return out, err
		}
		x.InventoryLotID = scanNullableString(lot)
		x.InventoryMovementID = scanNullableString(movement)
		out.Receipts = append(out.Receipts, x)
	}
	return out, rows.Err()
}

func validatePurchasePayload(p PurchasePayload) (PurchasePayload, error) {
	p.QuantityUnit = normalizeCode(p.QuantityUnit)
	p.CurrencyCode = normalizeCode(p.CurrencyCode)
	if p.CurrencyCode == "" {
		p.CurrencyCode = "IRR"
	}
	if strings.TrimSpace(p.SupplierID) == "" || strings.TrimSpace(p.StoneName) == "" || !validPositiveDecimal(p.Quantity) || !validNonNegativeDecimal(p.UnitPrice) || validateUnit(p.QuantityUnit) != nil || len(p.CurrencyCode) != 3 {
		return p, ErrValidation
	}
	return p, nil
}

func (s *OperationsService) CreatePurchase(ctx context.Context, actor, orderID, key string, p PurchasePayload) (Purchase, error) {
	var out Purchase
	p, err := validatePurchasePayload(p)
	if err != nil {
		return out, err
	}
	if p.CreateCostEntry && !s.HasPermission(ctx, actor, "finance.costs.record") && !s.HasPermission(ctx, actor, "finance.operational_costs.record") {
		return out, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "PURCHASE_CREATE", key, map[string]any{"order_id": orderID, "payload": p})
	if err != nil {
		return out, err
	}
	if claim.Existing {
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return out, err
		}
		return out, tx.Commit()
	}
	var supplierName string
	var supplierActive bool
	if err = tx.QueryRowContext(ctx, `SELECT name,is_active FROM suppliers WHERE id=$1 FOR SHARE`, p.SupplierID).Scan(&supplierName, &supplierActive); err != nil {
		return out, err
	}
	if !supplierActive {
		return out, conflict("INACTIVE_SUPPLIER", "supplier is inactive")
	}
	var orderNumber string
	if err = tx.QueryRowContext(ctx, `SELECT order_number FROM orders WHERE id=$1 AND status NOT IN ('CANCELLED','CLOSED') FOR SHARE`, orderID).Scan(&orderNumber); err != nil {
		return out, err
	}
	if p.BatchID != nil {
		var belongs bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM fulfillment_batches WHERE id=$1 AND order_id=$2)`, *p.BatchID, orderID).Scan(&belongs); err != nil {
			return out, err
		}
		if !belongs {
			return out, conflict("SCOPE_MISMATCH", "batch belongs to another order")
		}
	}
	number, err := nextReadableNumberTx(ctx, tx, "PUR")
	if err != nil {
		return out, err
	}
	var id, total string
	err = tx.QueryRowContext(ctx, `INSERT INTO purchase_records(purchase_number,order_id,batch_id,supplier_id,assigned_user_id,assigned_role_id,stone_name,description,quantity,quantity_unit,unit_price,total_amount,currency_code,expected_at,notes,created_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($8,''),$9::numeric,$10,$11::numeric,ROUND($9::numeric*$11::numeric,4),$12,$13,NULLIF($14,''),$15) RETURNING id,total_amount::text`, number, orderID, p.BatchID, p.SupplierID, p.AssignedUserID, p.AssignedRoleID, strings.TrimSpace(p.StoneName), p.Description, p.Quantity, p.QuantityUnit, p.UnitPrice, p.CurrencyCode, p.ExpectedAt, p.Notes, actor).Scan(&id, &total)
	if err != nil {
		return out, err
	}
	var costID *string
	if p.CreateCostEntry {
		entityType, entityID := "ORDER", orderID
		if p.BatchID != nil {
			entityType, entityID = "BATCH", *p.BatchID
		}
		var createdCost string
		err = tx.QueryRowContext(ctx, `INSERT INTO operational_cost_entries(entity_type,entity_id,cost_type,amount,currency,status,notes,incurred_at,created_by_user_id,order_id,batch_id,vendor_name,vendor_reference,supplier_id) VALUES($1,$2,'STONE_PURCHASE',$3::numeric,$4,'REPORTED',$5,NOW(),$6,$7,$8,$9,$10,$11) RETURNING id`, entityType, entityID, total, p.CurrencyCode, "Purchase "+number, actor, orderID, p.BatchID, supplierName, number, p.SupplierID).Scan(&createdCost)
		if err != nil {
			return out, err
		}
		costID = &createdCost
		if _, err = tx.ExecContext(ctx, `UPDATE purchase_records SET cost_entry_id=$2 WHERE id=$1`, id, createdCost); err != nil {
			return out, err
		}
		if err = refreshFinancialSummaryTx(ctx, tx, orderID); err != nil {
			return out, err
		}
	}
	out = Purchase{ID: id, PurchaseNumber: number, OrderID: orderID, OrderNumber: orderNumber, BatchID: p.BatchID, SupplierID: p.SupplierID, SupplierName: supplierName, AssignedUserID: p.AssignedUserID, AssignedRoleID: p.AssignedRoleID, StoneName: p.StoneName, Description: p.Description, Quantity: p.Quantity, ReceivedQuantity: "0.0000", QuantityUnit: p.QuantityUnit, UnitPrice: p.UnitPrice, TotalAmount: total, CurrencyCode: p.CurrencyCode, Status: "DRAFT", ExpectedAt: p.ExpectedAt, Notes: p.Notes, CostEntryID: costID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.auditTx(ctx, tx, actor, "PURCHASE_CREATED", "purchase_record", id, nil, p)
	if err = finishOperationTx(ctx, tx, actor, "PURCHASE_CREATE", key, out); err != nil {
		return out, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) UpdatePurchase(ctx context.Context, actor, id string, p PurchasePayload) (Purchase, error) {
	var out Purchase
	p, err := validatePurchasePayload(p)
	if err != nil {
		return out, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	var orderID, status string
	var costID sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT order_id,status,cost_entry_id FROM purchase_records WHERE id=$1 FOR UPDATE`, id).Scan(&orderID, &status, &costID); err != nil {
		return out, err
	}
	if status != "DRAFT" {
		return out, conflict("INVALID_PURCHASE_TRANSITION", "only draft purchases can be edited")
	}
	var active bool
	var supplierName string
	if err = tx.QueryRowContext(ctx, `SELECT name,is_active FROM suppliers WHERE id=$1 FOR SHARE`, p.SupplierID).Scan(&supplierName, &active); err != nil {
		return out, err
	}
	if !active {
		return out, conflict("INACTIVE_SUPPLIER", "supplier is inactive")
	}
	if p.BatchID != nil {
		var belongs bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM fulfillment_batches WHERE id=$1 AND order_id=$2)`, *p.BatchID, orderID).Scan(&belongs); err != nil {
			return out, err
		}
		if !belongs {
			return out, conflict("SCOPE_MISMATCH", "batch belongs to another order")
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE purchase_records SET batch_id=$2,supplier_id=$3,assigned_user_id=$4,assigned_role_id=$5,stone_name=$6,description=NULLIF($7,''),quantity=$8::numeric,quantity_unit=$9,unit_price=$10::numeric,total_amount=ROUND($8::numeric*$10::numeric,4),currency_code=$11,expected_at=$12,notes=NULLIF($13,''),updated_at=NOW() WHERE id=$1`, id, p.BatchID, p.SupplierID, p.AssignedUserID, p.AssignedRoleID, p.StoneName, p.Description, p.Quantity, p.QuantityUnit, p.UnitPrice, p.CurrencyCode, p.ExpectedAt, p.Notes); err != nil {
		return out, err
	}
	if costID.Valid {
		entityType, entityID := "ORDER", orderID
		if p.BatchID != nil {
			entityType, entityID = "BATCH", *p.BatchID
		}
		if _, err = tx.ExecContext(ctx, `UPDATE operational_cost_entries SET entity_type=$2,entity_id=$3,order_id=$4,batch_id=$5,amount=ROUND($6::numeric*$7::numeric,4),currency=$8,vendor_name=$9,supplier_id=$10,updated_at=NOW() WHERE id=$1 AND status IN ('ESTIMATED','REPORTED')`, costID.String, entityType, entityID, orderID, p.BatchID, p.Quantity, p.UnitPrice, p.CurrencyCode, supplierName, p.SupplierID); err != nil {
			return out, err
		}
		if err = refreshFinancialSummaryTx(ctx, tx, orderID); err != nil {
			return out, err
		}
	}
	s.auditTx(ctx, tx, actor, "PURCHASE_UPDATED", "purchase_record", id, map[string]any{"status": status}, p)
	if err = tx.Commit(); err != nil {
		return out, err
	}
	return s.GetPurchase(ctx, actor, id)
}

func (s *OperationsService) transitionPurchase(ctx context.Context, actor, id, key, operation, target, reason string) (map[string]any, error) {
	if target == "CANCELLED" && requireReason(reason) != nil {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, operation, key, map[string]any{"id": id, "reason": reason})
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
	var current string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM purchase_records WHERE id=$1 FOR UPDATE`, id).Scan(&current); err != nil {
		return nil, err
	}
	valid := target == "CONFIRMED" && current == "DRAFT" || target == "CANCELLED" && (current == "DRAFT" || current == "CONFIRMED" || current == "PARTIALLY_RECEIVED")
	if !valid {
		return nil, conflict("INVALID_PURCHASE_TRANSITION", fmt.Sprintf("cannot move purchase from %s to %s", current, target))
	}
	if target == "CANCELLED" {
		_, err = tx.ExecContext(ctx, `UPDATE purchase_records SET status='CANCELLED',cancelled_by_user_id=$2,cancelled_at=NOW(),cancellation_reason=$3,updated_at=NOW() WHERE id=$1`, id, actor, strings.TrimSpace(reason))
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE purchase_records SET status='CONFIRMED',updated_at=NOW() WHERE id=$1`, id)
	}
	if err != nil {
		return nil, err
	}
	action := "PURCHASE_CONFIRMED"
	if target == "CANCELLED" {
		action = "PURCHASE_CANCELLED"
	}
	s.auditTx(ctx, tx, actor, action, "purchase_record", id, map[string]any{"status": current}, map[string]any{"status": target, "reason": reason})
	out := map[string]any{"id": id, "status": target}
	if err = finishOperationTx(ctx, tx, actor, operation, key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) ConfirmPurchase(ctx context.Context, actor, id, key string) (map[string]any, error) {
	return s.transitionPurchase(ctx, actor, id, key, "PURCHASE_CONFIRM", "CONFIRMED", "")
}

func (s *OperationsService) CancelPurchase(ctx context.Context, actor, id, key, reason string) (map[string]any, error) {
	return s.transitionPurchase(ctx, actor, id, key, "PURCHASE_CANCEL", "CANCELLED", reason)
}

func (s *OperationsService) ReceivePurchase(ctx context.Context, actor, id, key string, p PurchaseReceiptPayload) (map[string]any, error) {
	if !validPositiveDecimal(p.Quantity) {
		return nil, ErrValidation
	}
	if p.InventoryReceipt != nil && !s.FeatureEnabled(ctx, "inventory_module_enabled") {
		return nil, conflict("MODULE_DISABLED", "ماژول موجودی غیرفعال است")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "PURCHASE_RECEIVE", key, map[string]any{"id": id, "payload": p})
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
	var orderID, unit, status, total, received, supplierID, stone string
	var batchID sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT order_id,batch_id,quantity_unit,status,quantity::text,received_quantity::text,supplier_id,stone_name FROM purchase_records WHERE id=$1 FOR UPDATE`, id).Scan(&orderID, &batchID, &unit, &status, &total, &received, &supplierID, &stone); err != nil {
		return nil, err
	}
	if status != "CONFIRMED" && status != "PARTIALLY_RECEIVED" {
		return nil, conflict("INVALID_PURCHASE_TRANSITION", "purchase is not ready to receive")
	}
	newReceived := addDecimal(received, p.Quantity)
	totalRat, _ := new(big.Rat).SetString(total)
	receivedRat, ok := new(big.Rat).SetString(newReceived)
	if !ok || totalRat == nil || receivedRat.Cmp(totalRat) > 0 {
		return nil, conflict("OVER_ALLOCATION", "received quantity exceeds purchase quantity")
	}
	when := time.Now()
	if p.ReceivedAt != nil {
		when = *p.ReceivedAt
	}
	var lotID, movementID *string
	if p.InventoryReceipt != nil {
		if !s.HasPermission(ctx, actor, "inventory.lots.create") || !s.HasPermission(ctx, actor, "inventory.movements.create") {
			return nil, ErrForbidden
		}
		r := *p.InventoryReceipt
		r.Lot.Quantity = p.Quantity
		r.Lot.QuantityUnit = unit
		r.Lot.OriginType = "EXTERNAL_SUPPLIER"
		r.Lot.OriginReferenceID = id
		r.Lot.StoneName = firstNonEmpty(r.Lot.StoneName, stone)
		r.Lot.SupplierID = &supplierID
		r.OrderID = &orderID
		if batchID.Valid {
			r.BatchID = &batchID.String
		}
		lot, lotErr := s.insertLotTx(ctx, tx, actor, r.Lot, p.Quantity, unit)
		if lotErr != nil {
			return nil, lotErr
		}
		group := randomUUIDText()
		if err = s.insertMovementTx(ctx, tx, actor, group, "RECEIPT", lot.ID, nil, &lot.CurrentLocationID, &orderID, scanNullableString(batchID), nil, nil, p.Quantity, unit, "0.0000", p.Quantity, "0.0000", "0.0000", "PURCHASE", id, firstNonEmpty(p.Notes, "purchase receipt"), nil); err != nil {
			return nil, err
		}
		var movement string
		if err = tx.QueryRowContext(ctx, `SELECT id FROM inventory_movements WHERE operation_group_id=$1 LIMIT 1`, group).Scan(&movement); err != nil {
			return nil, err
		}
		lotID, movementID = &lot.ID, &movement
	}
	var receiptID string
	if err = tx.QueryRowContext(ctx, `INSERT INTO purchase_receipts(purchase_record_id,quantity,quantity_unit,inventory_lot_id,inventory_movement_id,notes,received_by_user_id,received_at) VALUES($1,$2::numeric,$3,$4,$5,NULLIF($6,''),$7,$8) RETURNING id`, id, p.Quantity, unit, lotID, movementID, p.Notes, actor, when).Scan(&receiptID); err != nil {
		return nil, err
	}
	newStatus := "PARTIALLY_RECEIVED"
	if receivedRat.Cmp(totalRat) == 0 {
		newStatus = "RECEIVED"
	}
	if _, err = tx.ExecContext(ctx, `UPDATE purchase_records SET received_quantity=$2::numeric,status=$3,received_at=CASE WHEN $3='RECEIVED' THEN $4 ELSE received_at END,updated_at=NOW() WHERE id=$1`, id, newReceived, newStatus, when); err != nil {
		return nil, err
	}
	s.auditTx(ctx, tx, actor, "PURCHASE_RECEIVED", "purchase_record", id, map[string]any{"status": status, "received_quantity": received}, map[string]any{"status": newStatus, "received_quantity": newReceived, "receipt_id": receiptID})
	out := map[string]any{"id": id, "receipt_id": receiptID, "status": newStatus, "received_quantity": newReceived, "quantity_unit": unit, "inventory_lot_id": lotID, "inventory_movement_id": movementID}
	if err = finishOperationTx(ctx, tx, actor, "PURCHASE_RECEIVE", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func scanNullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
