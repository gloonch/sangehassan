package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

func (s *OperationsService) ListLocations(ctx context.Context, includeInactive bool) ([]InventoryLocation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,code,name_fa,location_type,COALESCE(address,''),COALESCE(city,''),COALESCE(province,''),country_code,is_active FROM inventory_locations WHERE $1 OR is_active ORDER BY name_fa`, includeInactive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InventoryLocation{}
	for rows.Next() {
		var x InventoryLocation
		if err = rows.Scan(&x.ID, &x.Code, &x.NameFA, &x.LocationType, &x.Address, &x.City, &x.Province, &x.CountryCode, &x.IsActive); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *OperationsService) CreateLocation(ctx context.Context, actor string, p LocationPayload) (InventoryLocation, error) {
	var x InventoryLocation
	p.Code = normalizeCode(p.Code)
	p.LocationType = normalizeCode(p.LocationType)
	if p.Code == "" || strings.TrimSpace(p.NameFA) == "" {
		return x, ErrValidation
	}
	active := true
	if p.IsActive != nil {
		active = *p.IsActive
	}
	if p.CountryCode == "" {
		p.CountryCode = "IR"
	}
	err := s.db.QueryRowContext(ctx, `INSERT INTO inventory_locations(code,name_fa,location_type,address,city,province,country_code,latitude,longitude,is_active,created_by_user_id) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),UPPER($7),NULLIF($8,'')::numeric,NULLIF($9,'')::numeric,$10,$11) RETURNING id,code,name_fa,location_type,COALESCE(address,''),COALESCE(city,''),COALESCE(province,''),country_code,is_active`, p.Code, p.NameFA, p.LocationType, p.Address, p.City, p.Province, p.CountryCode, valueOrEmpty(p.Latitude), valueOrEmpty(p.Longitude), active, actor).Scan(&x.ID, &x.Code, &x.NameFA, &x.LocationType, &x.Address, &x.City, &x.Province, &x.CountryCode, &x.IsActive)
	if err == nil {
		s.audit(ctx, actor, "inventory.locations.create", "inventory_location", x.ID, p)
	}
	return x, err
}
func (s *OperationsService) UpdateLocation(ctx context.Context, actor, id string, p LocationPayload) error {
	active := true
	if p.IsActive != nil {
		active = *p.IsActive
	}
	r, err := s.db.ExecContext(ctx, `UPDATE inventory_locations SET code=UPPER($2),name_fa=$3,location_type=$4,address=NULLIF($5,''),city=NULLIF($6,''),province=NULLIF($7,''),country_code=UPPER($8),latitude=NULLIF($9,'')::numeric,longitude=NULLIF($10,'')::numeric,is_active=$11,updated_at=NOW() WHERE id=$1`, id, p.Code, p.NameFA, normalizeCode(p.LocationType), p.Address, p.City, p.Province, p.CountryCode, valueOrEmpty(p.Latitude), valueOrEmpty(p.Longitude), active)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	s.audit(ctx, actor, "inventory.locations.update", "inventory_location", id, p)
	return nil
}

func (s *OperationsService) ListLots(ctx context.Context, location, status, stone string) ([]InventoryLot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT l.id,l.lot_number,l.parent_lot_id,l.current_location_id,loc.name_fa,l.origin_type,l.stone_name,COALESCE(l.stone_variant,''),l.available_quantity::text,l.reserved_quantity::text,l.quantity_unit,l.status,l.created_at FROM inventory_lots l JOIN inventory_locations loc ON loc.id=l.current_location_id WHERE ($1='' OR l.current_location_id=$1::uuid) AND ($2='' OR l.status=$2) AND ($3='' OR l.stone_name ILIKE '%'||$3||'%' OR l.stone_variant ILIKE '%'||$3||'%') ORDER BY l.created_at DESC`, location, status, stone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InventoryLot{}
	for rows.Next() {
		var x InventoryLot
		var parent sql.NullString
		if err = rows.Scan(&x.ID, &x.LotNumber, &parent, &x.CurrentLocationID, &x.LocationName, &x.OriginType, &x.StoneName, &x.StoneVariant, &x.AvailableQuantity, &x.ReservedQuantity, &x.QuantityUnit, &x.Status, &x.CreatedAt); err != nil {
			return nil, err
		}
		x.ParentLotID = scanNullableString(parent)
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) GetLot(ctx context.Context, id string) (InventoryLot, error) {
	var x InventoryLot
	var parent sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT l.id,l.lot_number,l.parent_lot_id,l.current_location_id,loc.name_fa,l.origin_type,l.stone_name,COALESCE(l.stone_variant,''),l.available_quantity::text,l.reserved_quantity::text,l.quantity_unit,l.status,l.created_at FROM inventory_lots l JOIN inventory_locations loc ON loc.id=l.current_location_id WHERE l.id=$1`, id).Scan(&x.ID, &x.LotNumber, &parent, &x.CurrentLocationID, &x.LocationName, &x.OriginType, &x.StoneName, &x.StoneVariant, &x.AvailableQuantity, &x.ReservedQuantity, &x.QuantityUnit, &x.Status, &x.CreatedAt)
	x.ParentLotID = scanNullableString(parent)
	return x, err
}

func (s *OperationsService) insertLotTx(ctx context.Context, tx *sql.Tx, actor string, p LotPayload, quantity, unit string) (InventoryLot, error) {
	var x InventoryLot
	unit = normalizeCode(unit)
	if !validPositiveDecimal(quantity) || validateUnit(unit) != nil || strings.TrimSpace(p.LocationID) == "" || strings.TrimSpace(p.StoneName) == "" {
		return x, ErrValidation
	}
	var active bool
	if err := tx.QueryRowContext(ctx, `SELECT is_active FROM inventory_locations WHERE id=$1 FOR SHARE`, p.LocationID).Scan(&active); err != nil {
		return x, err
	}
	if !active {
		return x, conflict("INACTIVE_LOCATION", "location is inactive")
	}
	number, err := nextReadableNumberTx(ctx, tx, "LOT")
	if err != nil {
		return x, err
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO inventory_lots(lot_number,parent_lot_id,origin_type,origin_reference_id,current_location_id,stone_category,stone_name,stone_variant,quality_grade,finish_type,cut_type,initial_quantity,available_quantity,reserved_quantity,quantity_unit,secondary_quantity,secondary_unit,status,received_at,created_by_user_id) VALUES($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8,$9,$10,$11,$12::numeric,$12::numeric,0,$13,NULLIF($14,'')::numeric,NULLIF($15,''),'AVAILABLE',NOW(),$16) RETURNING id`, number, p.ParentLotID, normalizeCode(p.OriginType), p.OriginReferenceID, p.LocationID, p.StoneCategory, p.StoneName, p.StoneVariant, p.QualityGrade, p.FinishType, p.CutType, quantity, unit, valueOrEmpty(p.SecondaryQuantity), p.SecondaryUnit, actor).Scan(&x.ID)
	if err != nil {
		return x, err
	}
	x.LotNumber = number
	x.ParentLotID = p.ParentLotID
	x.CurrentLocationID = p.LocationID
	x.OriginType = normalizeCode(p.OriginType)
	x.StoneName = p.StoneName
	x.StoneVariant = p.StoneVariant
	x.AvailableQuantity = quantity
	x.ReservedQuantity = "0.0000"
	x.QuantityUnit = unit
	x.Status = "AVAILABLE"
	return x, nil
}

func (s *OperationsService) ReceiveInventory(ctx context.Context, actor, key string, p ReceiptPayload) (InventoryLot, error) {
	var out InventoryLot
	if p.Reason == "" {
		p.Reason = "inventory receipt"
	}
	quantity := p.Lot.Quantity
	unit := normalizeCode(p.Lot.QuantityUnit)
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "INVENTORY_RECEIPT", key, p)
	if err != nil {
		return out, err
	}
	if idem.Existing {
		if err = json.Unmarshal(idem.Response, &out); err != nil {
			return out, err
		}
		return out, nil
	}
	out, err = s.insertLotTx(ctx, tx, actor, p.Lot, quantity, unit)
	if err != nil {
		return out, err
	}
	group := randomUUIDText()
	if err = s.insertMovementTx(ctx, tx, actor, group, "RECEIPT", out.ID, nil, &out.CurrentLocationID, p.OrderID, p.BatchID, nil, nil, quantity, unit, "0.0000", quantity, "0.0000", "0.0000", "RECEIPT", out.ID, p.Reason, nil); err != nil {
		return out, err
	}
	if p.BatchID != nil {
		if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, "BATCH_STOCK_RESERVED", "BATCH", *p.BatchID, group); err != nil {
			return out, err
		}
	}
	s.auditTx(ctx, tx, actor, "inventory.receipt", "inventory_lot", out.ID, nil, p)
	if err = finishOperationTx(ctx, tx, actor, "INVENTORY_RECEIPT", key, out); err != nil {
		return out, err
	}
	if err = tx.Commit(); err != nil {
		return out, err
	}
	return out, nil
}

func (s *OperationsService) UpdateLotMetadata(ctx context.Context, actor, id string, p LotPayload) error {
	r, err := s.db.ExecContext(ctx, `UPDATE inventory_lots SET origin_type=COALESCE(NULLIF($2,''),origin_type),origin_reference_id=NULLIF($3,''),stone_category=$4,stone_name=COALESCE(NULLIF($5,''),stone_name),stone_variant=$6,quality_grade=$7,finish_type=$8,cut_type=$9,updated_at=NOW() WHERE id=$1`, id, normalizeCode(p.OriginType), p.OriginReferenceID, p.StoneCategory, p.StoneName, p.StoneVariant, p.QualityGrade, p.FinishType, p.CutType)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	s.audit(ctx, actor, "inventory.lots.update_metadata", "inventory_lot", id, p)
	return nil
}

func (s *OperationsService) LotTraceability(ctx context.Context, id string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `WITH RECURSIVE lineage AS (SELECT id,parent_lot_id,lot_number,stone_name,quantity_unit,initial_quantity,0 depth FROM inventory_lots WHERE id=$1 UNION ALL SELECT p.id,p.parent_lot_id,p.lot_number,p.stone_name,p.quantity_unit,p.initial_quantity,l.depth+1 FROM inventory_lots p JOIN lineage l ON l.parent_lot_id=p.id) SELECT id,parent_lot_id,lot_number,stone_name,quantity_unit,initial_quantity::text,depth FROM lineage ORDER BY depth`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var lid, number, name, unit, qty string
		var parent sql.NullString
		var depth int
		if err = rows.Scan(&lid, &parent, &number, &name, &unit, &qty, &depth); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": lid, "parent_lot_id": scanNullableString(parent), "lot_number": number, "stone_name": name, "quantity_unit": unit, "initial_quantity": qty, "depth": depth})
	}
	return out, rows.Err()
}

func (s *OperationsService) CreateReservation(ctx context.Context, actor, lotID, key string, p ReservationPayload) (InventoryReservation, error) {
	var out InventoryReservation
	p.QuantityUnit = normalizeCode(p.QuantityUnit)
	if !validPositiveDecimal(p.Quantity) || validateUnit(p.QuantityUnit) != nil {
		return out, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "INVENTORY_RESERVATION", key, p)
	if err != nil {
		return out, err
	}
	if idem.Existing {
		if err = json.Unmarshal(idem.Response, &out); err != nil {
			return out, err
		}
		return out, nil
	}
	var available, reserved, unit, status, orderID, itemID, planned, batchUnit string
	if err = tx.QueryRowContext(ctx, `SELECT available_quantity::text,reserved_quantity::text,quantity_unit,status FROM inventory_lots WHERE id=$1 FOR UPDATE`, lotID).Scan(&available, &reserved, &unit, &status); err != nil {
		return out, err
	}
	if unit != p.QuantityUnit {
		return out, conflict("INCOMPATIBLE_UNIT", "reservation unit differs from lot")
	}
	if cmp, _ := decimalCmp(available, p.Quantity); cmp < 0 {
		return out, conflict("INSUFFICIENT_STOCK", "not enough free inventory")
	}
	if err = tx.QueryRowContext(ctx, `SELECT order_id,order_item_id,planned_quantity::text,quantity_unit FROM fulfillment_batches WHERE id=$1 AND status NOT IN ('SPLIT','MERGED','CANCELLED') FOR UPDATE`, p.BatchID).Scan(&orderID, &itemID, &planned, &batchUnit); err != nil {
		return out, err
	}
	if orderID != p.OrderID || itemID != p.OrderItemID {
		return out, conflict("SCOPE_MISMATCH", "reservation does not match batch")
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO inventory_reservations(inventory_lot_id,order_id,order_item_id,batch_id,reserved_quantity,quantity_unit,status,reserved_by_user_id,expires_at) VALUES($1,$2,$3,$4,$5::numeric,$6,'ACTIVE',$7,$8) RETURNING id,reserved_at`, lotID, p.OrderID, p.OrderItemID, p.BatchID, p.Quantity, p.QuantityUnit, actor, p.ExpiresAt).Scan(&out.ID, &out.ReservedAt)
	if err != nil {
		return out, err
	}
	afterA := subDecimal(available, p.Quantity)
	afterR := addDecimal(reserved, p.Quantity)
	_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET available_quantity=$2::numeric,reserved_quantity=$3::numeric,status=CASE WHEN $2::numeric=0 THEN 'RESERVED' ELSE 'PARTIALLY_RESERVED' END,updated_at=NOW() WHERE id=$1`, lotID, afterA, afterR)
	if err != nil {
		return out, err
	}
	group := randomUUIDText()
	if err = s.insertMovementTx(ctx, tx, actor, group, "RESERVATION", lotID, nil, nil, &p.OrderID, &p.BatchID, nil, &out.ID, p.Quantity, p.QuantityUnit, available, afterA, reserved, afterR, "RESERVATION", out.ID, "inventory reservation", nil); err != nil {
		return out, err
	}
	if p.WorkflowStepInstanceID != nil {
		rows, queryErr := tx.QueryContext(ctx, `SELECT (reserved_quantity-consumed_quantity)::text,quantity_unit FROM inventory_reservations WHERE batch_id=$1 AND status IN ('ACTIVE','PARTIALLY_CONSUMED')`, p.BatchID)
		if queryErr != nil {
			return out, queryErr
		}
		total := new(big.Rat)
		for rows.Next() {
			var quantity, reservationUnit string
			if queryErr = rows.Scan(&quantity, &reservationUnit); queryErr != nil {
				rows.Close()
				return out, queryErr
			}
			converted, convertErr := convertQuantityTx(ctx, tx, itemID, quantity, reservationUnit, batchUnit)
			if convertErr != nil {
				rows.Close()
				return out, convertErr
			}
			total.Add(total, converted)
		}
		rows.Close()
		plannedRat, _ := new(big.Rat).SetString(planned)
		if total.Cmp(plannedRat) >= 0 {
			if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, "BATCH_STOCK_RESERVED", "BATCH", p.BatchID, group); err != nil {
				return out, err
			}
		}
	}
	out.InventoryLotID = lotID
	out.BatchID = p.BatchID
	out.ReservedQuantity = p.Quantity
	out.ConsumedQuantity = "0.0000"
	out.QuantityUnit = p.QuantityUnit
	out.Status = "ACTIVE"
	out.ExpiresAt = p.ExpiresAt
	s.auditTx(ctx, tx, actor, "inventory.reservations.create", "inventory_reservation", out.ID, nil, p)
	if err = finishOperationTx(ctx, tx, actor, "INVENTORY_RESERVATION", key, out); err != nil {
		return out, err
	}
	if err = tx.Commit(); err != nil {
		return out, err
	}
	return out, nil
}

func (s *OperationsService) ListBatchReservations(ctx context.Context, batchID string) ([]InventoryReservation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,inventory_lot_id,batch_id,reserved_quantity::text,consumed_quantity::text,quantity_unit,status,reserved_at,expires_at FROM inventory_reservations WHERE batch_id=$1 ORDER BY reserved_at`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InventoryReservation{}
	for rows.Next() {
		var x InventoryReservation
		var expires sql.NullTime
		if err = rows.Scan(&x.ID, &x.InventoryLotID, &x.BatchID, &x.ReservedQuantity, &x.ConsumedQuantity, &x.QuantityUnit, &x.Status, &x.ReservedAt, &expires); err != nil {
			return nil, err
		}
		if expires.Valid {
			t := expires.Time
			x.ExpiresAt = &t
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) ReleaseReservation(ctx context.Context, actor, id, key, reason string) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "RESERVATION_RELEASE", key, map[string]string{"id": id, "reason": reason})
	if err != nil {
		return err
	}
	if idem.Existing {
		return tx.Commit()
	}
	var lotID, batchID, unit, status, remaining string
	if err = tx.QueryRowContext(ctx, `SELECT inventory_lot_id,batch_id,quantity_unit,status,(reserved_quantity-consumed_quantity)::text FROM inventory_reservations WHERE id=$1 FOR UPDATE`, id).Scan(&lotID, &batchID, &unit, &status, &remaining); err != nil {
		return err
	}
	if status != "ACTIVE" && status != "PARTIALLY_CONSUMED" {
		return conflict("INVALID_RESERVATION_STATE", "reservation is not releasable")
	}
	var beforeA, beforeR string
	if err = tx.QueryRowContext(ctx, `SELECT available_quantity::text,reserved_quantity::text FROM inventory_lots WHERE id=$1 FOR UPDATE`, lotID).Scan(&beforeA, &beforeR); err != nil {
		return err
	}
	afterA := addDecimal(beforeA, remaining)
	afterR := subDecimal(beforeR, remaining)
	_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET available_quantity=$2::numeric,reserved_quantity=$3::numeric,status=CASE WHEN $3::numeric=0 THEN 'AVAILABLE' ELSE 'PARTIALLY_RESERVED' END,updated_at=NOW() WHERE id=$1`, lotID, afterA, afterR)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='RELEASED',released_at=NOW(),release_reason=$2,updated_at=NOW() WHERE id=$1`, id, reason)
	if err != nil {
		return err
	}
	group := randomUUIDText()
	if err = s.insertMovementTx(ctx, tx, actor, group, "RESERVATION_RELEASE", lotID, nil, nil, nil, &batchID, nil, &id, remaining, unit, beforeA, afterA, beforeR, afterR, "RESERVATION", id, reason, nil); err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, "inventory.reservations.release", "inventory_reservation", id, map[string]any{"status": status}, map[string]any{"status": "RELEASED", "reason": reason})
	if err = finishOperationTx(ctx, tx, actor, "RESERVATION_RELEASE", key, map[string]bool{"released": true}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OperationsService) insertMovementTx(ctx context.Context, tx *sql.Tx, actor, group, kind, lotID string, source, destination, orderID, batchID, shipmentID, reservationID *string, quantity, unit, beforeA, afterA, beforeR, afterR, refType, refID, reason string, reversal *string) error {
	number, err := nextReadableNumberTx(ctx, tx, "MOV")
	if err != nil {
		return err
	}
	var id string
	err = tx.QueryRowContext(ctx, `INSERT INTO inventory_movements(movement_number,operation_group_id,movement_type,inventory_lot_id,source_location_id,destination_location_id,order_id,batch_id,shipment_id,reservation_id,quantity,quantity_unit,before_available_quantity,after_available_quantity,before_reserved_quantity,after_reserved_quantity,reference_type,reference_id,reason,reversal_of_movement_id,performed_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::numeric,$12,$13::numeric,$14::numeric,$15::numeric,$16::numeric,$17,$18,$19,$20,$21) RETURNING id`, number, group, kind, lotID, source, destination, orderID, batchID, shipmentID, reservationID, quantity, unit, beforeA, afterA, beforeR, afterR, nullableString(refType), nullableString(refID), nullableString(reason), reversal, actor).Scan(&id)
	if err == nil {
		s.auditTx(ctx, tx, actor, "inventory.movements.create", "inventory_movement", id, nil, map[string]any{"movement_number": number, "type": kind, "lot_id": lotID, "quantity": quantity, "unit": unit, "reason": reason})
	}
	return err
}

func (s *OperationsService) ListMovements(ctx context.Context, lotID, batchID, shipmentID string) ([]InventoryMovement, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,movement_number,operation_group_id,movement_type,inventory_lot_id,quantity::text,quantity_unit,before_available_quantity::text,after_available_quantity::text,before_reserved_quantity::text,after_reserved_quantity::text,COALESCE(reason,''),occurred_at FROM inventory_movements WHERE ($1='' OR inventory_lot_id=$1::uuid) AND ($2='' OR batch_id=$2::uuid) AND ($3='' OR shipment_id=$3::uuid) ORDER BY occurred_at DESC`, lotID, batchID, shipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InventoryMovement{}
	for rows.Next() {
		var x InventoryMovement
		if err = rows.Scan(&x.ID, &x.MovementNumber, &x.OperationGroupID, &x.MovementType, &x.InventoryLotID, &x.Quantity, &x.QuantityUnit, &x.BeforeAvailable, &x.AfterAvailable, &x.BeforeReserved, &x.AfterReserved, &x.Reason, &x.OccurredAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) TransferInventory(ctx context.Context, actor, key string, p TransferPayload) (map[string]any, error) {
	if !validPositiveDecimal(p.Quantity) || requireReason(p.Reason) != nil {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "INVENTORY_TRANSFER", key, p)
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
	var source, available, reserved, initial, unit, status string
	if err = tx.QueryRowContext(ctx, `SELECT current_location_id,available_quantity::text,reserved_quantity::text,initial_quantity::text,quantity_unit,status FROM inventory_lots WHERE id=$1 FOR UPDATE`, p.LotID).Scan(&source, &available, &reserved, &initial, &unit, &status); err != nil {
		return nil, err
	}
	if source == p.DestinationLocationID {
		return nil, conflict("SAME_LOCATION", "source and destination are equal")
	}
	if cmp, _ := decimalCmp(available, p.Quantity); cmp < 0 {
		return nil, conflict("INSUFFICIENT_STOCK", "transfer exceeds free quantity")
	}
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT is_active FROM inventory_locations WHERE id=$1 FOR SHARE`, p.DestinationLocationID).Scan(&active); err != nil {
		return nil, err
	}
	if !active {
		return nil, conflict("INACTIVE_LOCATION", "destination is inactive")
	}
	group := randomUUIDText()
	targetLot := p.LotID
	afterSource := available
	cmp, _ := decimalCmp(available, p.Quantity)
	fullTransfer := cmp == 0 && (reserved == "0.0000" || reserved == "0")
	if fullTransfer {
		_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET current_location_id=$2,status='AVAILABLE',updated_at=NOW() WHERE id=$1`, p.LotID, p.DestinationLocationID)
		if err != nil {
			return nil, err
		}
	} else {
		afterSource = subDecimal(available, p.Quantity)
		_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET initial_quantity=initial_quantity-$2::numeric,available_quantity=$3::numeric,updated_at=NOW() WHERE id=$1`, p.LotID, p.Quantity, afterSource)
		if err != nil {
			return nil, err
		}
		var meta LotPayload
		if err = tx.QueryRowContext(ctx, `SELECT id,origin_type,COALESCE(origin_reference_id,''),stone_category,stone_name,COALESCE(stone_variant,''),COALESCE(quality_grade,''),COALESCE(finish_type,''),COALESCE(cut_type,'') FROM inventory_lots WHERE id=$1`, p.LotID).Scan(&meta.ParentLotID, &meta.OriginType, &meta.OriginReferenceID, &meta.StoneCategory, &meta.StoneName, &meta.StoneVariant, &meta.QualityGrade, &meta.FinishType, &meta.CutType); err != nil {
			return nil, err
		}
		meta.LocationID = p.DestinationLocationID
		child, e := s.insertLotTx(ctx, tx, actor, meta, p.Quantity, unit)
		if e != nil {
			return nil, e
		}
		targetLot = child.ID
	}
	sourcePtr, destPtr := source, p.DestinationLocationID
	movementSourceAfter := afterSource
	if fullTransfer {
		movementSourceAfter = "0.0000"
	}
	if err = s.insertMovementTx(ctx, tx, actor, group, "TRANSFER_OUT", p.LotID, &sourcePtr, &destPtr, nil, nil, nil, nil, p.Quantity, unit, available, movementSourceAfter, reserved, reserved, "TRANSFER", group, p.Reason, nil); err != nil {
		return nil, err
	}
	targetBefore := "0.0000"
	if err = s.insertMovementTx(ctx, tx, actor, group, "TRANSFER_IN", targetLot, &sourcePtr, &destPtr, nil, nil, nil, nil, p.Quantity, unit, targetBefore, p.Quantity, "0.0000", "0.0000", "TRANSFER", group, p.Reason, nil); err != nil {
		return nil, err
	}
	out := map[string]any{"operation_group_id": group, "source_lot_id": p.LotID, "target_lot_id": targetLot}
	if err = finishOperationTx(ctx, tx, actor, "INVENTORY_TRANSFER", key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OperationsService) AdjustInventory(ctx context.Context, actor, key string, p AdjustmentPayload) (map[string]any, error) {
	if requireReason(p.Reason) != nil {
		return nil, ErrValidation
	}
	delta, ok := new(big.Rat).SetString(p.QuantityDelta)
	if !ok || delta.Sign() == 0 {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "INVENTORY_ADJUSTMENT", key, p)
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
	var available, reserved, unit, location string
	if err = tx.QueryRowContext(ctx, `SELECT available_quantity::text,reserved_quantity::text,quantity_unit,current_location_id FROM inventory_lots WHERE id=$1 FOR UPDATE`, p.LotID).Scan(&available, &reserved, &unit, &location); err != nil {
		return nil, err
	}
	av, _ := new(big.Rat).SetString(available)
	after := new(big.Rat).Add(av, delta)
	if after.Sign() < 0 {
		return nil, conflict("INSUFFICIENT_STOCK", "adjustment would make inventory negative")
	}
	quantity := new(big.Rat).Abs(delta).FloatString(quantityScale)
	kind := "ADJUSTMENT"
	_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET initial_quantity=initial_quantity+$2::numeric,available_quantity=$3::numeric,status=CASE WHEN $3::numeric=0 AND reserved_quantity=0 THEN 'CONSUMED' WHEN reserved_quantity>0 THEN 'PARTIALLY_RESERVED' ELSE 'AVAILABLE' END,updated_at=NOW() WHERE id=$1`, p.LotID, p.QuantityDelta, after.FloatString(quantityScale))
	if err != nil {
		return nil, err
	}
	group := randomUUIDText()
	loc := location
	if err = s.insertMovementTx(ctx, tx, actor, group, kind, p.LotID, &loc, &loc, nil, nil, nil, nil, quantity, unit, available, after.FloatString(quantityScale), reserved, reserved, "ADJUSTMENT", group, p.Reason, p.ReversalOfMovementID); err != nil {
		return nil, err
	}
	out := map[string]any{"operation_group_id": group, "available_quantity": after.FloatString(quantityScale)}
	if err = finishOperationTx(ctx, tx, actor, "INVENTORY_ADJUSTMENT", key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OperationsService) ConvertInventory(ctx context.Context, actor, key string, p ConversionPayload) (map[string]any, error) {
	if !validPositiveDecimal(p.InputQuantity) || !validNonNegativeDecimal(p.OutputQuantity) || !validNonNegativeDecimal(p.WasteQuantity) {
		return nil, ErrValidation
	}
	p.InputUnit = normalizeCode(p.InputUnit)
	p.OutputUnit = normalizeCode(p.OutputUnit)
	p.WasteUnit = normalizeCode(p.WasteUnit)
	if p.WasteUnit == "" {
		p.WasteUnit = p.InputUnit
	}
	if validateUnit(p.InputUnit) != nil || validateUnit(p.OutputUnit) != nil || validateUnit(p.WasteUnit) != nil {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "PRODUCTION_CONVERSION", key, p)
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
	var orderID, itemID, batchUnit, status string
	if err = tx.QueryRowContext(ctx, `SELECT order_id,order_item_id,quantity_unit,status FROM fulfillment_batches WHERE id=$1 FOR UPDATE`, p.BatchID).Scan(&orderID, &itemID, &batchUnit, &status); err != nil {
		return nil, err
	}
	var lotUnit, beforeA, beforeR, location string
	if err = tx.QueryRowContext(ctx, `SELECT quantity_unit,available_quantity::text,reserved_quantity::text,current_location_id FROM inventory_lots WHERE id=$1 FOR UPDATE`, p.InputLotID).Scan(&lotUnit, &beforeA, &beforeR, &location); err != nil {
		return nil, err
	}
	if lotUnit != p.InputUnit {
		return nil, conflict("INCOMPATIBLE_UNIT", "input unit differs from lot")
	}
	remaining := p.InputQuantity
	rows, err := tx.QueryContext(ctx, `SELECT id,(reserved_quantity-consumed_quantity)::text FROM inventory_reservations WHERE inventory_lot_id=$1 AND batch_id=$2 AND status IN ('ACTIVE','PARTIALLY_CONSUMED') ORDER BY reserved_at,id FOR UPDATE`, p.InputLotID, p.BatchID)
	if err != nil {
		return nil, err
	}
	type consume struct{ id, q string }
	uses := []consume{}
	need, _ := new(big.Rat).SetString(remaining)
	for rows.Next() && need.Sign() > 0 {
		var id, q string
		if err = rows.Scan(&id, &q); err != nil {
			return nil, err
		}
		availableQ, _ := new(big.Rat).SetString(q)
		take := new(big.Rat).Set(availableQ)
		if take.Cmp(need) > 0 {
			take.Set(need)
		}
		uses = append(uses, consume{id, take.FloatString(quantityScale)})
		need.Sub(need, take)
	}
	rows.Close()
	if need.Sign() > 0 {
		return nil, conflict("INSUFFICIENT_STOCK", "conversion requires a batch reservation")
	}
	for _, u := range uses {
		_, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET consumed_quantity=consumed_quantity+$2::numeric,status=CASE WHEN consumed_quantity+$2::numeric=reserved_quantity THEN 'CONSUMED' ELSE 'PARTIALLY_CONSUMED' END,consumed_at=CASE WHEN consumed_quantity+$2::numeric=reserved_quantity THEN NOW() ELSE consumed_at END,updated_at=NOW() WHERE id=$1`, u.id, u.q)
		if err != nil {
			return nil, err
		}
	}
	afterR := subDecimal(beforeR, p.InputQuantity)
	_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET reserved_quantity=$2::numeric,initial_quantity=initial_quantity-$3::numeric,status=CASE WHEN available_quantity=0 AND $2::numeric=0 THEN 'CONSUMED' ELSE 'IN_PROCESS' END,updated_at=NOW() WHERE id=$1`, p.InputLotID, afterR, p.InputQuantity)
	if err != nil {
		return nil, err
	}
	group := randomUUIDText()
	loc := location
	if err = s.insertMovementTx(ctx, tx, actor, group, "ISSUE_TO_PRODUCTION", p.InputLotID, &loc, nil, &orderID, &p.BatchID, nil, nil, p.InputQuantity, p.InputUnit, beforeA, beforeA, beforeR, afterR, "CONVERSION", group, "production input", nil); err != nil {
		return nil, err
	}
	p.OutputLot.ParentLotID = &p.InputLotID
	if p.OutputLot.LocationID == "" {
		p.OutputLot.LocationID = location
	}
	p.OutputLot.Quantity = p.OutputQuantity
	p.OutputLot.QuantityUnit = p.OutputUnit
	p.OutputLot.OriginType = "PRODUCTION"
	output, err := s.insertLotTx(ctx, tx, actor, p.OutputLot, p.OutputQuantity, p.OutputUnit)
	if err != nil {
		return nil, err
	}
	dest := output.CurrentLocationID
	if err = s.insertMovementTx(ctx, tx, actor, group, "PRODUCTION_OUTPUT", output.ID, nil, &dest, &orderID, &p.BatchID, nil, nil, p.OutputQuantity, p.OutputUnit, "0.0000", p.OutputQuantity, "0.0000", "0.0000", "CONVERSION", group, "production output", nil); err != nil {
		return nil, err
	}
	if q, _ := new(big.Rat).SetString(p.WasteQuantity); q.Sign() > 0 {
		if p.WasteUnit != p.InputUnit {
			return nil, conflict("INCOMPATIBLE_UNIT", "waste unit must match input")
		}
		if err = s.insertMovementTx(ctx, tx, actor, group, "WASTE", p.InputLotID, &loc, nil, &orderID, &p.BatchID, nil, nil, p.WasteQuantity, p.WasteUnit, beforeA, beforeA, afterR, afterR, "CONVERSION", group, "production waste", nil); err != nil {
			return nil, err
		}
	}
	var conversionID string
	err = tx.QueryRowContext(ctx, `INSERT INTO inventory_lot_conversions(operation_group_id,batch_id,workflow_step_instance_id,input_lot_id,output_lot_id,order_item_conversion_id,input_quantity,input_unit,output_quantity,output_unit,waste_quantity,waste_unit,conversion_type,created_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7::numeric,$8,$9::numeric,$10,$11::numeric,$12,$13,$14) RETURNING id`, group, p.BatchID, p.WorkflowStepInstanceID, p.InputLotID, output.ID, p.OrderItemConversionID, p.InputQuantity, p.InputUnit, p.OutputQuantity, p.OutputUnit, p.WasteQuantity, p.WasteUnit, normalizeCode(p.ConversionType), actor).Scan(&conversionID)
	if err != nil {
		return nil, err
	}
	actual, err := convertQuantityTx(ctx, tx, itemID, p.OutputQuantity, p.OutputUnit, batchUnit)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET actual_quantity=actual_quantity+$2::numeric,status='READY_FOR_QC',updated_at=NOW() WHERE id=$1`, p.BatchID, actual.FloatString(quantityScale))
	if err != nil {
		return nil, err
	}
	out := map[string]any{"id": conversionID, "operation_group_id": group, "output_lot_id": output.ID}
	if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, "BATCH_READY_FOR_QC", "BATCH", p.BatchID, group); err != nil {
		return nil, err
	}
	s.auditTx(ctx, tx, actor, "inventory.conversions.create", "inventory_lot_conversion", conversionID, nil, p)
	if err = finishOperationTx(ctx, tx, actor, "PRODUCTION_CONVERSION", key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OperationsService) InventoryDashboard(ctx context.Context, actor string) (map[string]any, error) {
	out := map[string]any{}
	queries := map[string]string{"low_stock": `SELECT COUNT(*) FROM inventory_stock_policies p WHERE p.is_active AND COALESCE((SELECT SUM(l.available_quantity) FROM inventory_lots l WHERE l.current_location_id=p.location_id AND l.stone_name=p.stone_name AND COALESCE(l.stone_variant,'')=p.stone_variant AND l.quantity_unit=p.quantity_unit AND l.status NOT IN ('CANCELLED','SOLD','CONSUMED')),0)<=p.low_stock_threshold`, "stale_reservations": `SELECT COUNT(*) FROM inventory_reservations WHERE status IN ('ACTIVE','PARTIALLY_CONSUMED') AND (expires_at<NOW() OR reserved_at<NOW()-INTERVAL '72 hours')`, "quarantined": `SELECT COUNT(*) FROM inventory_lots WHERE status='QUARANTINED'`, "damaged": `SELECT COUNT(*) FROM inventory_lots WHERE status='DAMAGED'`}
	for key, q := range queries {
		var n int
		if err := s.db.QueryRowContext(ctx, q).Scan(&n); err != nil {
			return nil, err
		}
		out[key] = n
	}
	return out, nil
}

var _ = fmt.Sprint
var _ = time.Now
var _ = errors.Is
