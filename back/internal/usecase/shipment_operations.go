package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

func (s *OperationsService) ListVehicles(ctx context.Context, includeInactive bool) ([]Vehicle, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,vehicle_type,COALESCE(plate_number,''),COALESCE(plate_normalized,''),capacity_value::text,COALESCE(capacity_unit,''),driver_user_id,is_active FROM vehicles WHERE $1 OR is_active ORDER BY created_at DESC`, includeInactive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Vehicle{}
	for rows.Next() {
		var x Vehicle
		var cap, driver sql.NullString
		if err = rows.Scan(&x.ID, &x.VehicleType, &x.PlateNumber, &x.PlateNormalized, &cap, &x.CapacityUnit, &driver, &x.IsActive); err != nil {
			return nil, err
		}
		x.CapacityValue = scanNullableString(cap)
		x.DriverUserID = scanNullableString(driver)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *OperationsService) CreateVehicle(ctx context.Context, actor string, p VehiclePayload) (Vehicle, error) {
	var x Vehicle
	p.VehicleType = normalizeCode(p.VehicleType)
	plate := normalizePlate(p.PlateNumber)
	active := true
	if p.IsActive != nil {
		active = *p.IsActive
	}
	err := s.db.QueryRowContext(ctx, `INSERT INTO vehicles(vehicle_type,plate_number,plate_normalized,trailer_number,capacity_value,capacity_unit,owner_name,carrier_name,driver_user_id,is_active) VALUES($1,NULLIF($2,''),NULLIF($3,''),NULLIF($4,''),NULLIF($5,'')::numeric,NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),$9,$10) RETURNING id,vehicle_type,COALESCE(plate_number,''),COALESCE(plate_normalized,''),capacity_value::text,COALESCE(capacity_unit,''),driver_user_id,is_active`, p.VehicleType, p.PlateNumber, plate, p.TrailerNumber, valueOrEmpty(p.CapacityValue), normalizeCode(p.CapacityUnit), p.OwnerName, p.CarrierName, p.DriverUserID, active).Scan(&x.ID, &x.VehicleType, &x.PlateNumber, &x.PlateNormalized, &x.CapacityValue, &x.CapacityUnit, &x.DriverUserID, &x.IsActive)
	if err == nil {
		s.audit(ctx, actor, "vehicles.create", "vehicle", x.ID, p)
	}
	return x, err
}
func (s *OperationsService) UpdateVehicle(ctx context.Context, actor, id string, p VehiclePayload) error {
	active := true
	if p.IsActive != nil {
		active = *p.IsActive
	}
	r, err := s.db.ExecContext(ctx, `UPDATE vehicles SET vehicle_type=$2,plate_number=NULLIF($3,''),plate_normalized=NULLIF($4,''),trailer_number=NULLIF($5,''),capacity_value=NULLIF($6,'')::numeric,capacity_unit=NULLIF($7,''),owner_name=NULLIF($8,''),carrier_name=NULLIF($9,''),driver_user_id=$10,is_active=$11,updated_at=NOW() WHERE id=$1`, id, normalizeCode(p.VehicleType), p.PlateNumber, normalizePlate(p.PlateNumber), p.TrailerNumber, valueOrEmpty(p.CapacityValue), normalizeCode(p.CapacityUnit), p.OwnerName, p.CarrierName, p.DriverUserID, active)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	s.audit(ctx, actor, "vehicles.update", "vehicle", id, p)
	return nil
}

func (s *OperationsService) CreateShipment(ctx context.Context, actor, orderID, key string, p ShipmentPayload) (Shipment, error) {
	var out Shipment
	p.ShipmentType = normalizeCode(p.ShipmentType)
	if p.OriginLocationID == "" || p.ShipmentType == "" {
		return out, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	operation := "SHIPMENT_CREATE:" + orderID
	idem, err := claimOperationTx(ctx, tx, actor, operation, key, p)
	if err != nil {
		return out, err
	}
	if idem.Existing {
		var stored struct {
			ID string `json:"id"`
		}
		if err = json.Unmarshal(idem.Response, &stored); err != nil {
			return out, err
		}
		return s.GetShipment(ctx, actor, stored.ID, false)
	}
	var orderNumber string
	if err = tx.QueryRowContext(ctx, `SELECT order_number FROM orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&orderNumber); err != nil {
		return out, err
	}
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT is_active FROM inventory_locations WHERE id=$1 FOR SHARE`, p.OriginLocationID).Scan(&active); err != nil {
		return out, err
	}
	if !active {
		return out, conflict("INACTIVE_LOCATION", "origin is inactive")
	}
	if err = ensureActiveSupplierTx(ctx, tx, p.SupplierID); err != nil {
		return out, err
	}
	number, err := nextReadableNumberTx(ctx, tx, "SHP")
	if err != nil {
		return out, err
	}
	title := strings.TrimSpace(p.CustomerTitleFA)
	if title == "" {
		title = "محموله سفارش"
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO shipments(shipment_number,order_id,shipment_type,origin_location_id,destination_location_id,status,driver_user_id,external_driver_name,external_driver_phone,carrier_name,vehicle_id,planned_departure_at,estimated_arrival_at,delivery_contact_name,delivery_contact_phone,delivery_address,customer_title_fa,notes,created_by_user_id,supplier_id) VALUES($1,$2,$3,$4,$5,'DRAFT',$6,NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),$10,$11,$12,NULLIF($13,''),NULLIF($14,''),NULLIF($15,''),$16,NULLIF($17,''),$18,$19) RETURNING id`, number, orderID, p.ShipmentType, p.OriginLocationID, p.DestinationLocationID, p.DriverUserID, p.ExternalDriverName, NormalizePhone(p.ExternalDriverPhone), p.CarrierName, p.VehicleID, p.PlannedDepartureAt, p.EstimatedArrivalAt, p.DeliveryContactName, NormalizePhone(p.DeliveryContactPhone), p.DeliveryAddress, title, p.Notes, actor, p.SupplierID).Scan(&out.ID)
	if err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO shipment_orders(shipment_id,order_id,is_primary) VALUES($1,$2,TRUE)`, out.ID, orderID)
	if err != nil {
		return out, err
	}
	templateID := p.WorkflowTemplateID
	if templateID == nil {
		group := "domestic_shipment"
		if p.ShipmentType == "EXPORT_CONTAINER" {
			group = "export_shipment"
		}
		var selected int64
		if err = tx.QueryRowContext(ctx, `SELECT id FROM workflow_templates WHERE template_group_code=$1 AND scope_type='SHIPMENT' AND status='PUBLISHED' AND is_active ORDER BY version_number DESC LIMIT 1`, group).Scan(&selected); err != nil {
			return out, err
		}
		templateID = &selected
	}
	wid, err := s.startScopedWorkflowTx(ctx, tx, actor, orderID, "SHIPMENT", out.ID, templateID, p.ParentWorkflowID, p.ParentStepID)
	if err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE shipments SET workflow_instance_id=$2 WHERE id=$1`, out.ID, wid)
	if err != nil {
		return out, err
	}
	s.auditTx(ctx, tx, actor, "shipments.create", "shipment", out.ID, nil, p)
	if err = finishOperationTx(ctx, tx, actor, operation, key, map[string]string{"id": out.ID}); err != nil {
		return out, err
	}
	if err = tx.Commit(); err != nil {
		return out, err
	}
	return s.GetShipment(ctx, actor, out.ID, false)
}

func (s *OperationsService) canViewShipment(ctx context.Context, actor, id string, customer bool) bool {
	var ok bool
	if customer {
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM shipments sh JOIN orders o ON o.id=sh.order_id WHERE sh.id=$1 AND o.customer_user_id=$2 AND sh.customer_visible)`, id, actor).Scan(&ok)
		return ok
	}
	if s.HasPermission(ctx, actor, "shipments.view_all") {
		return true
	}
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM shipments sh WHERE sh.id=$1 AND (sh.driver_user_id=$2 OR EXISTS(SELECT 1 FROM workflow_step_instances si JOIN user_roles ur ON ur.role_id=si.responsible_role_id WHERE si.workflow_instance_id=sh.workflow_instance_id AND ur.user_id=$2)))`, id, actor).Scan(&ok)
	return ok
}
func (s *OperationsService) ListShipments(ctx context.Context, actor, status, orderID string, customer bool) ([]Shipment, error) {
	viewAll := !customer && s.HasPermission(ctx, actor, "shipments.view_all")
	rows, err := s.db.QueryContext(ctx, `SELECT sh.id,sh.shipment_number,sh.order_id,o.order_number,sh.shipment_type,sh.status,sh.origin_location_id,sh.destination_location_id,sh.driver_user_id,sh.vehicle_id,sh.workflow_instance_id,sh.customer_visible,sh.customer_title_fa,sh.planned_departure_at,sh.estimated_arrival_at,sh.actual_departure_at,sh.actual_arrival_at,sh.created_at FROM shipments sh JOIN orders o ON o.id=sh.order_id WHERE ($2 OR ($3 AND o.customer_user_id=$1 AND sh.customer_visible) OR (NOT $3 AND (sh.driver_user_id=$1 OR EXISTS(SELECT 1 FROM workflow_step_instances si JOIN user_roles ur ON ur.role_id=si.responsible_role_id WHERE si.workflow_instance_id=sh.workflow_instance_id AND ur.user_id=$1)))) AND ($4='' OR sh.status=$4) AND ($5='' OR sh.order_id=$5::uuid) ORDER BY sh.created_at DESC`, actor, viewAll, customer, status, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Shipment{}
	for rows.Next() {
		x, e := scanShipment(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

type rowScanner interface{ Scan(...any) error }

func scanShipment(row rowScanner) (Shipment, error) {
	var x Shipment
	var dest, driver, vehicle, wid sql.NullString
	var planned, eta, depart, arrive sql.NullTime
	err := row.Scan(&x.ID, &x.ShipmentNumber, &x.OrderID, &x.OrderNumber, &x.ShipmentType, &x.Status, &x.OriginLocationID, &dest, &driver, &vehicle, &wid, &x.CustomerVisible, &x.CustomerTitleFA, &planned, &eta, &depart, &arrive, &x.CreatedAt)
	x.DestinationLocationID = scanNullableString(dest)
	x.DriverUserID = scanNullableString(driver)
	x.VehicleID = scanNullableString(vehicle)
	x.WorkflowInstanceID = scanNullableString(wid)
	if planned.Valid {
		t := planned.Time
		x.PlannedDepartureAt = &t
	}
	if eta.Valid {
		t := eta.Time
		x.EstimatedArrivalAt = &t
	}
	if depart.Valid {
		t := depart.Time
		x.ActualDepartureAt = &t
	}
	if arrive.Valid {
		t := arrive.Time
		x.ActualArrivalAt = &t
	}
	return x, err
}
func (s *OperationsService) GetShipment(ctx context.Context, actor, id string, customer bool) (Shipment, error) {
	var out Shipment
	if !s.canViewShipment(ctx, actor, id, customer) {
		return out, ErrForbidden
	}
	out, err := scanShipment(s.db.QueryRowContext(ctx, `SELECT sh.id,sh.shipment_number,sh.order_id,o.order_number,sh.shipment_type,sh.status,sh.origin_location_id,sh.destination_location_id,sh.driver_user_id,sh.vehicle_id,sh.workflow_instance_id,sh.customer_visible,sh.customer_title_fa,sh.planned_departure_at,sh.estimated_arrival_at,sh.actual_departure_at,sh.actual_arrival_at,sh.created_at FROM shipments sh JOIN orders o ON o.id=sh.order_id WHERE sh.id=$1`, id))
	if err != nil {
		return out, err
	}
	if !customer {
		out.Items, err = s.ListShipmentItems(ctx, id)
	} else {
		items, e := s.ListShipmentItems(ctx, id)
		if e != nil {
			return out, e
		}
		for i := range items {
			items[i].InventoryLotID = ""
		}
		out.Items = items
	}
	return out, err
}
func (s *OperationsService) ListShipmentItems(ctx context.Context, id string) ([]ShipmentItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT si.id,si.shipment_id,si.batch_id,b.batch_number,si.inventory_lot_id,si.planned_quantity::text,si.loaded_quantity::text,si.delivered_quantity::text,si.quantity_unit,si.package_count,si.bundle_count FROM shipment_items si JOIN fulfillment_batches b ON b.id=si.batch_id WHERE si.shipment_id=$1 ORDER BY si.created_at`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ShipmentItem{}
	for rows.Next() {
		var x ShipmentItem
		if err = rows.Scan(&x.ID, &x.ShipmentID, &x.BatchID, &x.BatchNumber, &x.InventoryLotID, &x.PlannedQuantity, &x.LoadedQuantity, &x.DeliveredQuantity, &x.QuantityUnit, &x.PackageCount, &x.BundleCount); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) UpdateShipment(ctx context.Context, actor, id string, p ShipmentPayload) error {
	if p.CustomerVisible != nil && !s.HasPermission(ctx, actor, "shipments.override") {
		return ErrForbidden
	}
	r, err := s.db.ExecContext(ctx, `UPDATE shipments SET driver_user_id=$2,external_driver_name=NULLIF($3,''),external_driver_phone=NULLIF($4,''),carrier_name=NULLIF($5,''),vehicle_id=$6,planned_departure_at=$7,estimated_arrival_at=$8,delivery_contact_name=NULLIF($9,''),delivery_contact_phone=NULLIF($10,''),delivery_address=NULLIF($11,''),customer_title_fa=COALESCE(NULLIF($12,''),customer_title_fa),customer_visible=COALESCE($13,customer_visible),status=CASE WHEN status='DRAFT' THEN 'PLANNED' ELSE status END,notes=NULLIF($14,''),updated_at=NOW() WHERE id=$1 AND status NOT IN ('DELIVERED','CANCELLED')`, id, p.DriverUserID, p.ExternalDriverName, NormalizePhone(p.ExternalDriverPhone), p.CarrierName, p.VehicleID, p.PlannedDepartureAt, p.EstimatedArrivalAt, p.DeliveryContactName, NormalizePhone(p.DeliveryContactPhone), p.DeliveryAddress, p.CustomerTitleFA, p.CustomerVisible, p.Notes)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict("INVALID_SHIPMENT_STATE", "shipment is terminal or missing")
	}
	s.audit(ctx, actor, "shipments.update", "shipment", id, p)
	return nil
}

func (s *OperationsService) AddShipmentItem(ctx context.Context, actor, shipmentID string, p ShipmentItemPayload) (ShipmentItem, error) {
	var out ShipmentItem
	p.QuantityUnit = normalizeCode(p.QuantityUnit)
	if !validPositiveDecimal(p.PlannedQuantity) || validateUnit(p.QuantityUnit) != nil {
		return out, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	var orderID, status string
	if err = tx.QueryRowContext(ctx, `SELECT order_id,status FROM shipments WHERE id=$1 FOR UPDATE`, shipmentID).Scan(&orderID, &status); err != nil {
		return out, err
	}
	if status != "DRAFT" && status != "PLANNED" && status != "READY_FOR_LOADING" {
		return out, conflict("INVALID_SHIPMENT_STATE", "items cannot be changed after loading")
	}
	var batchOrder, batchUnit string
	if err = tx.QueryRowContext(ctx, `SELECT order_id,quantity_unit FROM fulfillment_batches WHERE id=$1 AND status NOT IN ('SPLIT','MERGED','CANCELLED') FOR SHARE`, p.BatchID).Scan(&batchOrder, &batchUnit); err != nil {
		return out, err
	}
	if batchOrder != orderID {
		return out, conflict("CROSS_ORDER_SHIPMENT", "Phase 3 shipments accept one order only")
	}
	if batchUnit != p.QuantityUnit {
		return out, conflict("INCOMPATIBLE_UNIT", "shipment item unit differs from batch")
	}
	var lotUnit string
	if err = tx.QueryRowContext(ctx, `SELECT quantity_unit FROM inventory_lots WHERE id=$1 FOR SHARE`, p.InventoryLotID).Scan(&lotUnit); err != nil {
		return out, err
	}
	if lotUnit != p.QuantityUnit {
		return out, conflict("INCOMPATIBLE_UNIT", "shipment item unit differs from lot")
	}
	var already string
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(planned_quantity),0)::text FROM shipment_items WHERE batch_id=$1`, p.BatchID).Scan(&already); err != nil {
		return out, err
	}
	var planned string
	if err = tx.QueryRowContext(ctx, `SELECT planned_quantity::text FROM fulfillment_batches WHERE id=$1`, p.BatchID).Scan(&planned); err != nil {
		return out, err
	}
	sum := addDecimal(already, p.PlannedQuantity)
	if cmp, _ := decimalCmp(sum, planned); cmp > 0 {
		return out, conflict("OVER_ALLOCATION", "shipment plans exceed batch quantity")
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO shipment_items(shipment_id,batch_id,inventory_lot_id,planned_quantity,quantity_unit,package_count,bundle_count,notes) VALUES($1,$2,$3,$4::numeric,$5,$6,$7,NULLIF($8,'')) RETURNING id`, shipmentID, p.BatchID, p.InventoryLotID, p.PlannedQuantity, p.QuantityUnit, p.PackageCount, p.BundleCount, p.Notes).Scan(&out.ID)
	if err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE shipments SET status='READY_FOR_LOADING',updated_at=NOW() WHERE id=$1`, shipmentID)
	if err != nil {
		return out, err
	}
	s.auditTx(ctx, tx, actor, "shipments.items.create", "shipment_item", out.ID, nil, p)
	if err = tx.Commit(); err != nil {
		return out, err
	}
	items, e := s.ListShipmentItems(ctx, shipmentID)
	if e != nil {
		return out, e
	}
	for _, x := range items {
		if x.ID == out.ID {
			return x, nil
		}
	}
	return out, sql.ErrNoRows
}

func (s *OperationsService) consumeReservationForLoadTx(ctx context.Context, tx *sql.Tx, actor, shipmentID, itemID, batchID, lotID, quantity, unit, group, eventID, transitLocation string) (string, error) {
	var source, available, reserved, initial, status, lotUnit string
	if err := tx.QueryRowContext(ctx, `SELECT current_location_id,available_quantity::text,reserved_quantity::text,initial_quantity::text,status,quantity_unit FROM inventory_lots WHERE id=$1 FOR UPDATE`, lotID).Scan(&source, &available, &reserved, &initial, &status, &lotUnit); err != nil {
		return "", err
	}
	if lotUnit != unit {
		return "", conflict("INCOMPATIBLE_UNIT", "load unit differs from lot")
	}
	need, _ := new(big.Rat).SetString(quantity)
	rows, err := tx.QueryContext(ctx, `SELECT id,(reserved_quantity-consumed_quantity)::text FROM inventory_reservations WHERE inventory_lot_id=$1 AND batch_id=$2 AND status IN ('ACTIVE','PARTIALLY_CONSUMED') ORDER BY reserved_at,id FOR UPDATE`, lotID, batchID)
	if err != nil {
		return "", err
	}
	type use struct{ id, q string }
	uses := []use{}
	for rows.Next() && need.Sign() > 0 {
		var id, q string
		if err = rows.Scan(&id, &q); err != nil {
			return "", err
		}
		free, _ := new(big.Rat).SetString(q)
		take := new(big.Rat).Set(free)
		if take.Cmp(need) > 0 {
			take.Set(need)
		}
		uses = append(uses, use{id, take.FloatString(quantityScale)})
		need.Sub(need, take)
	}
	rows.Close()
	if need.Sign() > 0 {
		return "", conflict("INSUFFICIENT_STOCK", "loading requires reserved inventory")
	}
	for _, u := range uses {
		_, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET consumed_quantity=consumed_quantity+$2::numeric,status=CASE WHEN consumed_quantity+$2::numeric=reserved_quantity THEN 'CONSUMED' ELSE 'PARTIALLY_CONSUMED' END,consumed_at=CASE WHEN consumed_quantity+$2::numeric=reserved_quantity THEN NOW() ELSE consumed_at END,updated_at=NOW() WHERE id=$1`, u.id, u.q)
		if err != nil {
			return "", err
		}
	}
	afterR := subDecimal(reserved, quantity)
	afterInitial := subDecimal(initial, quantity)
	if cmp, _ := decimalCmp(afterInitial, "0"); cmp < 0 {
		return "", conflict("INSUFFICIENT_STOCK", "loading exceeds physical lot")
	}
	_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET initial_quantity=$2::numeric,reserved_quantity=$3::numeric,status=CASE WHEN available_quantity=0 AND $3::numeric=0 THEN 'CONSUMED' WHEN $3::numeric>0 THEN 'PARTIALLY_RESERVED' ELSE 'AVAILABLE' END,updated_at=NOW() WHERE id=$1`, lotID, afterInitial, afterR)
	if err != nil {
		return "", err
	}
	number, err := nextReadableNumberTx(ctx, tx, "LOT")
	if err != nil {
		return "", err
	}
	var transitLot string
	err = tx.QueryRowContext(ctx, `INSERT INTO inventory_lots(lot_number,parent_lot_id,origin_type,origin_reference_id,current_location_id,stone_category,stone_name,stone_variant,quality_grade,finish_type,cut_type,initial_quantity,available_quantity,reserved_quantity,quantity_unit,status,created_by_user_id) SELECT $1,id,'SHIPMENT',$2,$3,stone_category,stone_name,stone_variant,quality_grade,finish_type,cut_type,$4::numeric,0,0,quantity_unit,'IN_TRANSIT',$5 FROM inventory_lots WHERE id=$6 RETURNING id`, number, shipmentID, transitLocation, quantity, actor, lotID).Scan(&transitLot)
	if err != nil {
		return "", err
	}
	sourcePtr, destPtr := source, transitLocation
	if err = s.insertMovementTx(ctx, tx, actor, group, "SHIPMENT_LOADING", lotID, &sourcePtr, &destPtr, nil, &batchID, &shipmentID, nil, quantity, unit, available, available, reserved, afterR, "SHIPMENT_EVENT", eventID, "shipment loading", nil); err != nil {
		return "", err
	}
	if err = s.insertMovementTx(ctx, tx, actor, group, "TRANSFER_IN", transitLot, &sourcePtr, &destPtr, nil, &batchID, &shipmentID, nil, quantity, unit, "0.0000", "0.0000", "0.0000", "0.0000", "SHIPMENT_EVENT", eventID, "loaded transit lot", nil); err != nil {
		return "", err
	}
	return transitLot, nil
}

func (s *OperationsService) LoadShipment(ctx context.Context, actor, id, key string, p ShipmentOperationPayload) (map[string]any, error) {
	if len(p.Items) == 0 {
		return nil, ErrValidation
	}
	if !s.canViewShipment(ctx, actor, id, false) {
		return nil, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "SHIPMENT_LOADING", key, p)
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
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM shipments WHERE id=$1 FOR UPDATE`, id).Scan(&status); err != nil {
		return nil, err
	}
	if status != "READY_FOR_LOADING" && status != "LOADING" {
		return nil, conflict("INVALID_SHIPMENT_STATE", "shipment is not loadable")
	}
	var transit string
	if err = tx.QueryRowContext(ctx, `SELECT id FROM inventory_locations WHERE code='SYSTEM-TRANSIT' FOR SHARE`).Scan(&transit); err != nil {
		return nil, err
	}
	group := randomUUIDText()
	var eventID string
	if err = tx.QueryRowContext(ctx, `INSERT INTO shipment_events(shipment_id,operation_group_id,event_type,reason,performed_by_user_id) VALUES($1,$2,'LOADING',NULLIF($3,''),$4) RETURNING id`, id, group, p.Reason, actor).Scan(&eventID); err != nil {
		return nil, err
	}
	for _, entry := range p.Items {
		if !validPositiveDecimal(entry.Quantity) {
			return nil, ErrValidation
		}
		var batchID, lotID, unit, planned, loaded string
		if err = tx.QueryRowContext(ctx, `SELECT batch_id,inventory_lot_id,quantity_unit,planned_quantity::text,loaded_quantity::text FROM shipment_items WHERE id=$1 AND shipment_id=$2 FOR UPDATE`, entry.ShipmentItemID, id).Scan(&batchID, &lotID, &unit, &planned, &loaded); err != nil {
			return nil, err
		}
		after := addDecimal(loaded, entry.Quantity)
		if cmp, _ := decimalCmp(after, planned); cmp > 0 {
			return nil, conflict("OVER_ALLOCATION", "loaded quantity exceeds shipment item")
		}
		transitLot, err := s.consumeReservationForLoadTx(ctx, tx, actor, id, entry.ShipmentItemID, batchID, lotID, entry.Quantity, unit, group, eventID, transit)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO shipment_event_items(shipment_event_id,shipment_item_id,inventory_lot_id,quantity,quantity_unit) VALUES($1,$2,$3,$4::numeric,$5)`, eventID, entry.ShipmentItemID, transitLot, entry.Quantity, unit)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `UPDATE shipment_items SET loaded_quantity=$2::numeric,updated_at=NOW() WHERE id=$1`, entry.ShipmentItemID, after)
		if err != nil {
			return nil, err
		}
		s.updateBatchShipmentStatusTx(ctx, tx, batchID)
	}
	newStatus := "LOADING"
	if p.FinalizeLoading {
		var incomplete bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM shipment_items WHERE shipment_id=$1 AND loaded_quantity<planned_quantity)`, id).Scan(&incomplete); err != nil {
			return nil, err
		}
		if incomplete {
			return nil, conflict("INCOMPLETE_LOADING", "all planned quantities must be loaded before finalization")
		}
		newStatus = "LOADED"
	}
	_, err = tx.ExecContext(ctx, `UPDATE shipments SET status=$2,updated_at=NOW() WHERE id=$1`, id, newStatus)
	if err != nil {
		return nil, err
	}
	if err = emitShipmentCustomerNotificationTx(ctx, tx, id, "SHIPMENT_LOADING"); err != nil {
		return nil, err
	}
	if p.FinalizeLoading {
		if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, "SHIPMENT_LOADED", "SHIPMENT", id, group); err != nil {
			return nil, err
		}
	}
	out := map[string]any{"event_id": eventID, "operation_group_id": group, "status": newStatus}
	s.auditTx(ctx, tx, actor, "shipments.load", "shipment", id, map[string]any{"status": status}, out)
	if err = finishOperationTx(ctx, tx, actor, "SHIPMENT_LOADING", key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OperationsService) DispatchShipment(ctx context.Context, actor, id, key string, p ShipmentOperationPayload) (map[string]any, error) {
	if !s.canViewShipment(ctx, actor, id, false) {
		return nil, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "SHIPMENT_DISPATCH", key, p)
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
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM shipments WHERE id=$1 FOR UPDATE`, id).Scan(&status); err != nil {
		return nil, err
	}
	if status != "LOADED" {
		return nil, conflict("INVALID_SHIPMENT_STATE", "shipment must be loaded before dispatch")
	}
	group := randomUUIDText()
	var eventID string
	if err = tx.QueryRowContext(ctx, `INSERT INTO shipment_events(shipment_id,operation_group_id,event_type,reason,performed_by_user_id) VALUES($1,$2,'DISPATCH',NULLIF($3,''),$4) RETURNING id`, id, group, p.Reason, actor).Scan(&eventID); err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE shipments SET status='IN_TRANSIT',actual_departure_at=NOW(),customer_visible=TRUE,updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, "SHIPMENT_DISPATCHED", "SHIPMENT", id, group); err != nil {
		return nil, err
	}
	if err = emitShipmentCustomerNotificationTx(ctx, tx, id, "SHIPMENT_DISPATCHED"); err != nil {
		return nil, err
	}
	out := map[string]any{"event_id": eventID, "status": "IN_TRANSIT"}
	s.auditTx(ctx, tx, actor, "shipments.dispatch", "shipment", id, map[string]any{"status": status}, out)
	if err = finishOperationTx(ctx, tx, actor, "SHIPMENT_DISPATCH", key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OperationsService) ArriveShipment(ctx context.Context, actor, id, key string, p ShipmentOperationPayload) (map[string]any, error) {
	if !s.canViewShipment(ctx, actor, id, false) {
		return nil, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "SHIPMENT_ARRIVAL", key, p)
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
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM shipments WHERE id=$1 FOR UPDATE`, id).Scan(&status); err != nil {
		return nil, err
	}
	if status != "IN_TRANSIT" {
		return nil, conflict("INVALID_SHIPMENT_STATE", "shipment is not in transit")
	}
	group := randomUUIDText()
	var eventID string
	if err = tx.QueryRowContext(ctx, `INSERT INTO shipment_events(shipment_id,operation_group_id,event_type,reason,performed_by_user_id) VALUES($1,$2,'ARRIVAL',NULLIF($3,''),$4) RETURNING id`, id, group, p.Reason, actor).Scan(&eventID); err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE shipments SET status='ARRIVED',actual_arrival_at=NOW(),updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, "SHIPMENT_ARRIVED", "SHIPMENT", id, group); err != nil {
		return nil, err
	}
	if err = emitShipmentCustomerNotificationTx(ctx, tx, id, "SHIPMENT_ARRIVED"); err != nil {
		return nil, err
	}
	out := map[string]any{"event_id": eventID, "status": "ARRIVED"}
	s.auditTx(ctx, tx, actor, "shipments.arrive", "shipment", id, map[string]any{"status": status}, out)
	if err = finishOperationTx(ctx, tx, actor, "SHIPMENT_ARRIVAL", key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OperationsService) DeliverShipment(ctx context.Context, actor, id, key string, p ShipmentOperationPayload, customer bool) (map[string]any, error) {
	if len(p.Items) == 0 && !customer {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	operation := "SHIPMENT_DELIVERY"
	idem, err := claimOperationTx(ctx, tx, actor, operation, key, p)
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
	var status, orderID, origin string
	var destination sql.NullString
	var owner, driver sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT sh.status,sh.order_id,sh.origin_location_id,sh.destination_location_id,o.customer_user_id,sh.driver_user_id FROM shipments sh JOIN orders o ON o.id=sh.order_id WHERE sh.id=$1 FOR UPDATE OF sh,o`, id).Scan(&status, &orderID, &origin, &destination, &owner, &driver); err != nil {
		return nil, err
	}
	if customer {
		if !owner.Valid || owner.String != actor {
			return nil, ErrForbidden
		}
	} else if !s.HasPermission(ctx, actor, "shipments.view_all") && (!driver.Valid || driver.String != actor) {
		return nil, ErrForbidden
	}
	if status != "ARRIVED" && status != "UNLOADING" && status != "PARTIALLY_DELIVERED" {
		return nil, conflict("INVALID_SHIPMENT_STATE", "shipment has not arrived")
	}
	if p.ProofFileID != nil {
		var validProof bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_files WHERE id=$1 AND entity_type IN ('SHIPMENT','DELIVERY') AND entity_id=$2)`, *p.ProofFileID, id).Scan(&validProof); err != nil {
			return nil, err
		}
		if !validProof {
			return nil, conflict("SCOPE_MISMATCH", "delivery proof belongs to another entity")
		}
	}
	if customer && len(p.Items) == 0 {
		rows, queryErr := tx.QueryContext(ctx, `SELECT id,(loaded_quantity-delivered_quantity)::text FROM shipment_items WHERE shipment_id=$1 AND loaded_quantity>delivered_quantity FOR UPDATE`, id)
		if queryErr != nil {
			return nil, queryErr
		}
		for rows.Next() {
			var item ShipmentQuantity
			if queryErr = rows.Scan(&item.ShipmentItemID, &item.Quantity); queryErr != nil {
				rows.Close()
				return nil, queryErr
			}
			p.Items = append(p.Items, item)
		}
		if queryErr = rows.Close(); queryErr != nil {
			return nil, queryErr
		}
		if len(p.Items) == 0 {
			return nil, conflict("INVALID_SHIPMENT_STATE", "shipment has no remaining delivery quantity")
		}
		p.FinalizeDelivery = true
	}
	group := randomUUIDText()
	var eventID string
	if err = tx.QueryRowContext(ctx, `INSERT INTO shipment_events(shipment_id,operation_group_id,event_type,reason,receiver_name,receiver_phone,proof_file_id,performed_by_user_id) VALUES($1,$2,'DELIVERY',NULLIF($3,''),NULLIF($4,''),NULLIF($5,''),$6,$7) RETURNING id`, id, group, p.Reason, p.ReceiverName, NormalizePhone(p.ReceiverPhone), p.ProofFileID, actor).Scan(&eventID); err != nil {
		return nil, err
	}
	for _, entry := range p.Items {
		if !validPositiveDecimal(entry.Quantity) {
			return nil, ErrValidation
		}
		var batchID, unit, loaded, delivered string
		if err = tx.QueryRowContext(ctx, `SELECT batch_id,quantity_unit,loaded_quantity::text,delivered_quantity::text FROM shipment_items WHERE id=$1 AND shipment_id=$2 FOR UPDATE`, entry.ShipmentItemID, id).Scan(&batchID, &unit, &loaded, &delivered); err != nil {
			return nil, err
		}
		after := addDecimal(delivered, entry.Quantity)
		if cmp, _ := decimalCmp(after, loaded); cmp > 0 {
			return nil, conflict("OVER_ALLOCATION", "delivery exceeds loaded quantity")
		}
		need, _ := new(big.Rat).SetString(entry.Quantity)
		lots, err := tx.QueryContext(ctx, `SELECT sei.inventory_lot_id,l.initial_quantity::text,l.current_location_id FROM shipment_event_items sei JOIN shipment_events se ON se.id=sei.shipment_event_id AND se.event_type='LOADING' JOIN inventory_lots l ON l.id=sei.inventory_lot_id WHERE sei.shipment_item_id=$1 AND l.status='IN_TRANSIT' ORDER BY se.occurred_at,l.id FOR UPDATE OF l`, entry.ShipmentItemID)
		if err != nil {
			return nil, err
		}
		type transitLot struct{ id, q, location string }
		availableLots := []transitLot{}
		for lots.Next() {
			var x transitLot
			if err = lots.Scan(&x.id, &x.q, &x.location); err != nil {
				return nil, err
			}
			availableLots = append(availableLots, x)
		}
		lots.Close()
		for _, lot := range availableLots {
			if need.Sign() <= 0 {
				break
			}
			q, _ := new(big.Rat).SetString(lot.q)
			take := new(big.Rat).Set(q)
			if take.Cmp(need) > 0 {
				take.Set(need)
			}
			takeText := take.FloatString(quantityScale)
			target := lot.id
			if take.Cmp(q) < 0 {
				_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET initial_quantity=initial_quantity-$2::numeric,updated_at=NOW() WHERE id=$1`, lot.id, takeText)
				if err != nil {
					return nil, err
				}
				number, e := nextReadableNumberTx(ctx, tx, "LOT")
				if e != nil {
					return nil, e
				}
				loc := lot.location
				if destination.Valid {
					loc = destination.String
				}
				err = tx.QueryRowContext(ctx, `INSERT INTO inventory_lots(lot_number,parent_lot_id,origin_type,origin_reference_id,current_location_id,stone_category,stone_name,stone_variant,quality_grade,finish_type,cut_type,initial_quantity,available_quantity,reserved_quantity,quantity_unit,status,created_by_user_id) SELECT $1,id,'DELIVERY',$2,$3,stone_category,stone_name,stone_variant,quality_grade,finish_type,cut_type,$4::numeric,0,0,quantity_unit,'SOLD',$5 FROM inventory_lots WHERE id=$6 RETURNING id`, number, id, loc, takeText, actor, lot.id).Scan(&target)
				if err != nil {
					return nil, err
				}
			} else {
				if destination.Valid {
					_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET current_location_id=$2,status='SOLD',updated_at=NOW() WHERE id=$1`, lot.id, destination.String)
				} else {
					_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET status='SOLD',updated_at=NOW() WHERE id=$1`, lot.id)
				}
				if err != nil {
					return nil, err
				}
			}
			sourcePtr := lot.location
			var destPtr *string
			if destination.Valid {
				d := destination.String
				destPtr = &d
			}
			if err = s.insertMovementTx(ctx, tx, actor, group, "DELIVERY", target, &sourcePtr, destPtr, &orderID, &batchID, &id, nil, takeText, unit, "0.0000", "0.0000", "0.0000", "0.0000", "SHIPMENT_EVENT", eventID, p.Reason, nil); err != nil {
				return nil, err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO shipment_event_items(shipment_event_id,shipment_item_id,inventory_lot_id,quantity,quantity_unit) VALUES($1,$2,$3,$4::numeric,$5)`, eventID, entry.ShipmentItemID, target, takeText, unit)
			if err != nil {
				return nil, err
			}
			need.Sub(need, take)
		}
		if need.Sign() > 0 {
			return nil, conflict("INSUFFICIENT_STOCK", "loaded transit lots are insufficient")
		}
		_, err = tx.ExecContext(ctx, `UPDATE shipment_items SET delivered_quantity=$2::numeric,updated_at=NOW() WHERE id=$1`, entry.ShipmentItemID, after)
		if err != nil {
			return nil, err
		}
		s.updateBatchDeliveryStatusTx(ctx, tx, batchID)
	}
	var incomplete bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM shipment_items WHERE shipment_id=$1 AND delivered_quantity<loaded_quantity)`, id).Scan(&incomplete); err != nil {
		return nil, err
	}
	newStatus := "PARTIALLY_DELIVERED"
	if !incomplete {
		newStatus = "DELIVERED"
	} else if p.FinalizeDelivery {
		newStatus = "HAS_DISCREPANCY"
		_, _ = tx.ExecContext(ctx, `INSERT INTO workflow_discrepancies(workflow_instance_id,shipment_id,metric_key,severity,is_blocking,status,reported_by_user_id,source_explanation) SELECT workflow_instance_id,id,'DELIVERED_QUANTITY','CRITICAL',TRUE,'OPEN',$2,$3 FROM shipments WHERE id=$1`, id, actor, "final delivery is lower than loaded quantity")
	}
	_, err = tx.ExecContext(ctx, `UPDATE shipments SET status=$2,updated_at=NOW() WHERE id=$1`, id, newStatus)
	if err != nil {
		return nil, err
	}
	if newStatus == "DELIVERED" {
		if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, "SHIPMENT_DELIVERED", "SHIPMENT", id, group); err != nil {
			return nil, err
		}
		if err = emitShipmentCustomerNotificationTx(ctx, tx, id, "SHIPMENT_DELIVERED"); err != nil {
			return nil, err
		}
	}
	out := map[string]any{"event_id": eventID, "status": newStatus}
	s.auditTx(ctx, tx, actor, "shipments.deliver", "shipment", id, map[string]any{"status": status}, out)
	if err = finishOperationTx(ctx, tx, actor, operation, key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OperationsService) updateBatchShipmentStatusTx(ctx context.Context, tx *sql.Tx, batchID string) {
	var planned, loaded string
	_ = tx.QueryRowContext(ctx, `SELECT b.planned_quantity::text,COALESCE((SELECT SUM(si.loaded_quantity) FROM shipment_items si WHERE si.batch_id=b.id),0)::text FROM fulfillment_batches b WHERE b.id=$1`, batchID).Scan(&planned, &loaded)
	status := "PARTIALLY_SHIPPED"
	if cmp, _ := decimalCmp(loaded, planned); cmp >= 0 {
		status = "SHIPPED"
	}
	_, _ = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET status=$2,updated_at=NOW() WHERE id=$1`, batchID, status)
}
func (s *OperationsService) updateBatchDeliveryStatusTx(ctx context.Context, tx *sql.Tx, batchID string) {
	var planned, delivered string
	_ = tx.QueryRowContext(ctx, `SELECT b.planned_quantity::text,COALESCE((SELECT SUM(si.delivered_quantity) FROM shipment_items si WHERE si.batch_id=b.id),0)::text FROM fulfillment_batches b WHERE b.id=$1`, batchID).Scan(&planned, &delivered)
	status := "PARTIALLY_DELIVERED"
	if cmp, _ := decimalCmp(delivered, planned); cmp >= 0 {
		status = "DELIVERED"
	}
	_, _ = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET status=$2,updated_at=NOW() WHERE id=$1`, batchID, status)
}

func (s *OperationsService) CancelShipment(ctx context.Context, actor, id, key, reason string) error {
	if requireReason(reason) != nil {
		return ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "SHIPMENT_CANCEL", key, map[string]string{"id": id, "reason": reason})
	if err != nil {
		return err
	}
	if idem.Existing {
		return tx.Commit()
	}
	var status, origin string
	if err = tx.QueryRowContext(ctx, `SELECT status,origin_location_id FROM shipments WHERE id=$1 FOR UPDATE`, id).Scan(&status, &origin); err != nil {
		return err
	}
	if status == "DELIVERED" || status == "PARTIALLY_DELIVERED" {
		return conflict("INVALID_SHIPMENT_STATE", "delivered shipment cannot be cancelled")
	}
	if (status == "IN_TRANSIT" || status == "ARRIVED") && !s.HasPermission(ctx, actor, "shipments.override") {
		return ErrForbidden
	}
	group := randomUUIDText()
	rows, err := tx.QueryContext(ctx, `SELECT l.id,l.current_location_id,l.initial_quantity::text,l.quantity_unit,(SELECT m.id FROM inventory_movements m WHERE m.shipment_id=$1 AND m.inventory_lot_id=l.id AND m.movement_type='TRANSFER_IN' ORDER BY m.occurred_at DESC LIMIT 1) FROM inventory_lots l WHERE l.id IN (SELECT sei.inventory_lot_id FROM shipment_event_items sei JOIN shipment_events se ON se.id=sei.shipment_event_id AND se.event_type='LOADING' WHERE se.shipment_id=$1) AND l.status='IN_TRANSIT' ORDER BY l.id FOR UPDATE OF l`, id)
	if err != nil {
		return err
	}
	type loadedLot struct{ id, location, q, unit, reversal string }
	lots := []loadedLot{}
	for rows.Next() {
		var x loadedLot
		if err = rows.Scan(&x.id, &x.location, &x.q, &x.unit, &x.reversal); err != nil {
			return err
		}
		lots = append(lots, x)
	}
	rows.Close()
	for _, lot := range lots {
		_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET current_location_id=$2,available_quantity=initial_quantity,status='AVAILABLE',updated_at=NOW() WHERE id=$1`, lot.id, origin)
		if err != nil {
			return err
		}
		source, dest := lot.location, origin
		if err = s.insertMovementTx(ctx, tx, actor, group, "CANCELLATION", lot.id, &source, &dest, nil, nil, &id, nil, lot.q, lot.unit, "0.0000", lot.q, "0.0000", "0.0000", "SHIPMENT", id, reason, &lot.reversal); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE shipments SET status='CANCELLED',cancelled_at=NOW(),updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE workflow_instances SET status='CANCELLED',cancelled_at=NOW() WHERE scope_type='SHIPMENT' AND scope_id=$1`, id)
	_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='CANCELLED',updated_at=NOW() WHERE workflow_instance_id IN (SELECT id FROM workflow_instances WHERE scope_type='SHIPMENT' AND scope_id=$1) AND status NOT IN ('COMPLETED','CANCELLED')`, id)
	var eventID string
	_ = tx.QueryRowContext(ctx, `INSERT INTO shipment_events(shipment_id,operation_group_id,event_type,reason,performed_by_user_id) VALUES($1,$2,'CANCELLATION',$3,$4) RETURNING id`, id, group, reason, actor).Scan(&eventID)
	s.auditTx(ctx, tx, actor, "shipments.cancel", "shipment", id, map[string]any{"status": status}, map[string]any{"status": "CANCELLED", "reason": reason})
	if err = finishOperationTx(ctx, tx, actor, "SHIPMENT_CANCEL", key, map[string]bool{"cancelled": true}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OperationsService) CreatePackaging(ctx context.Context, actor, batchID, key string, p PackagingPayload) (map[string]any, error) {
	if !validPositiveDecimal(p.Quantity) || validateUnit(normalizeCode(p.QuantityUnit)) != nil {
		return nil, ErrValidation
	}
	var id, number string
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	operation := "PACKAGING_CREATE:" + batchID
	idem, err := claimOperationTx(ctx, tx, actor, operation, key, p)
	if err != nil {
		return nil, err
	}
	if idem.Existing {
		var existing map[string]any
		if err = json.Unmarshal(idem.Response, &existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	var batchNumber, itemID, planned, batchUnit string
	if err = tx.QueryRowContext(ctx, `SELECT batch_number,order_item_id,planned_quantity::text,quantity_unit FROM fulfillment_batches WHERE id=$1 AND status NOT IN ('SPLIT','MERGED','CANCELLED') FOR UPDATE`, batchID).Scan(&batchNumber, &itemID, &planned, &batchUnit); err != nil {
		return nil, err
	}
	var lotUnit, lotQuantity string
	var scoped bool
	if err = tx.QueryRowContext(ctx, `SELECT l.quantity_unit,l.initial_quantity::text,EXISTS(SELECT 1 FROM inventory_reservations r WHERE r.batch_id=$2 AND r.inventory_lot_id=l.id UNION ALL SELECT 1 FROM inventory_lot_conversions c WHERE c.batch_id=$2 AND c.output_lot_id=l.id) FROM inventory_lots l WHERE l.id=$1 FOR SHARE`, p.InventoryLotID, batchID).Scan(&lotUnit, &lotQuantity, &scoped); err != nil {
		return nil, err
	}
	if !scoped {
		return nil, conflict("SCOPE_MISMATCH", "package lot is not traceable to batch")
	}
	if lotUnit != normalizeCode(p.QuantityUnit) {
		return nil, conflict("INCOMPATIBLE_UNIT", "package unit differs from lot")
	}
	packageRows, queryErr := tx.QueryContext(ctx, `SELECT quantity::text,quantity_unit FROM packaging_units WHERE batch_id=$1 AND status<>'CANCELLED'`, batchID)
	if queryErr != nil {
		return nil, queryErr
	}
	type packageQuantity struct{ quantity, unit string }
	existingPackages := []packageQuantity{}
	for packageRows.Next() {
		var value packageQuantity
		if queryErr = packageRows.Scan(&value.quantity, &value.unit); queryErr != nil {
			packageRows.Close()
			return nil, queryErr
		}
		existingPackages = append(existingPackages, value)
	}
	if queryErr = packageRows.Close(); queryErr != nil {
		return nil, queryErr
	}
	totalPackaged := new(big.Rat)
	for _, value := range existingPackages {
		converted, convertErr := convertQuantityTx(ctx, tx, itemID, value.quantity, value.unit, batchUnit)
		if convertErr != nil {
			return nil, convertErr
		}
		totalPackaged.Add(totalPackaged, converted)
	}
	convertedNew, convertErr := convertQuantityTx(ctx, tx, itemID, p.Quantity, normalizeCode(p.QuantityUnit), batchUnit)
	if convertErr != nil {
		return nil, convertErr
	}
	totalPackaged.Add(totalPackaged, convertedNew)
	plannedRat, _ := new(big.Rat).SetString(planned)
	if totalPackaged.Cmp(plannedRat) > 0 {
		return nil, conflict("OVER_ALLOCATION", "package quantity exceeds batch")
	}
	var lotPackaged string
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity),0)::text FROM packaging_units WHERE inventory_lot_id=$1 AND status<>'CANCELLED'`, p.InventoryLotID).Scan(&lotPackaged); err != nil {
		return nil, err
	}
	if cmp, _ := decimalCmp(addDecimal(lotPackaged, p.Quantity), lotQuantity); cmp > 0 {
		return nil, conflict("OVER_ALLOCATION", "package quantity exceeds lot")
	}
	if err = tx.QueryRowContext(ctx, `SELECT $1||'-P'||LPAD((COUNT(*)+1)::text,2,'0') FROM packaging_units WHERE batch_id=$2`, batchNumber, batchID).Scan(&number); err != nil {
		return nil, err
	}
	if err = ensureActiveSupplierTx(ctx, tx, p.SupplierID); err != nil {
		return nil, err
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO packaging_units(package_number,batch_id,inventory_lot_id,package_type,quantity,quantity_unit,gross_weight,net_weight,weight_unit,status,customer_visible,supplier_id) VALUES($1,$2,$3,$4,$5::numeric,$6,NULLIF($7,'')::numeric,NULLIF($8,'')::numeric,NULLIF($9,''),'PACKED',$10,$11) RETURNING id`, number, batchID, p.InventoryLotID, normalizeCode(p.PackageType), p.Quantity, normalizeCode(p.QuantityUnit), valueOrEmpty(p.GrossWeight), valueOrEmpty(p.NetWeight), normalizeCode(p.WeightUnit), p.CustomerVisible, p.SupplierID).Scan(&id)
	if err != nil {
		return nil, err
	}
	if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, "BATCH_PACKAGED", "BATCH", batchID, randomUUIDText()); err != nil {
		return nil, err
	}
	out := map[string]any{"id": id, "package_number": number, "status": "PACKED"}
	s.auditTx(ctx, tx, actor, "packaging.create", "packaging_unit", id, nil, p)
	if err = finishOperationTx(ctx, tx, actor, operation, key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}
func (s *OperationsService) UpdatePackagingStatus(ctx context.Context, actor, id, status, reason string) error {
	status = normalizeCode(status)
	if status == "CANCELLED" && requireReason(reason) != nil {
		return ErrValidation
	}
	r, err := s.db.ExecContext(ctx, `UPDATE packaging_units SET status=$2,updated_at=NOW() WHERE id=$1 AND status NOT IN ('LOADED','DELIVERED','CANCELLED')`, id, status)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict("INVALID_PACKAGE_STATE", "package cannot be updated")
	}
	s.audit(ctx, actor, "packaging.status", "packaging_unit", id, map[string]any{"status": status, "reason": reason})
	return nil
}
func (s *OperationsService) ListPackaging(ctx context.Context, batchID string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,package_number,batch_id,inventory_lot_id,package_type,quantity::text,quantity_unit,status,customer_visible FROM packaging_units WHERE ($1='' OR batch_id=$1::uuid) ORDER BY created_at DESC`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, n, b, l, t, q, u, st string
		var visible bool
		if err = rows.Scan(&id, &n, &b, &l, &t, &q, &u, &st, &visible); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "package_number": n, "batch_id": b, "inventory_lot_id": l, "package_type": t, "quantity": q, "quantity_unit": u, "status": st, "customer_visible": visible})
	}
	return out, rows.Err()
}

func (s *OperationsService) AssignPackage(ctx context.Context, actor, packageID, shipmentItemID, reason string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var packageBatch, itemBatch string
	if err = tx.QueryRowContext(ctx, `SELECT batch_id FROM packaging_units WHERE id=$1 AND status IN ('PACKED','QC_APPROVED') FOR UPDATE`, packageID).Scan(&packageBatch); err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx, `SELECT batch_id FROM shipment_items WHERE id=$1`, shipmentItemID).Scan(&itemBatch); err != nil {
		return err
	}
	if packageBatch != itemBatch {
		return conflict("SCOPE_MISMATCH", "package and shipment item have different batches")
	}
	var activeAssignment, activeItem sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id,shipment_item_id FROM shipment_package_assignments WHERE packaging_unit_id=$1 AND released_at IS NULL FOR UPDATE`, packageID).Scan(&activeAssignment, &activeItem)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if activeItem.Valid && activeItem.String == shipmentItemID {
		return tx.Commit()
	}
	if activeAssignment.Valid {
		if requireReason(reason) != nil {
			return errors.New("reason required for package reassignment")
		}
		if _, err = tx.ExecContext(ctx, `UPDATE shipment_package_assignments SET released_at=NOW(),release_reason=$2 WHERE id=$1`, activeAssignment.String, reason); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO shipment_package_assignments(packaging_unit_id,shipment_item_id,assigned_by_user_id) VALUES($1,$2,$3)`, packageID, shipmentItemID, actor)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE packaging_units SET status='ASSIGNED_TO_SHIPMENT',updated_at=NOW() WHERE id=$1`, packageID)
	if err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, "packaging.assign_to_shipment", "packaging_unit", packageID, map[string]any{"shipment_item_id": scanNullableString(activeItem)}, map[string]any{"shipment_item_id": shipmentItemID, "reason": reason})
	return tx.Commit()
}

func validContainerNumber(v string) bool { return containerPattern.MatchString(normalizePlate(v)) }
func (s *OperationsService) CreateContainer(ctx context.Context, actor, shipmentID string, p ContainerPayload) (map[string]any, error) {
	number := normalizePlate(p.ContainerNumber)
	if !validContainerNumber(number) {
		return nil, conflict("INVALID_CONTAINER_NUMBER", "container number must contain four letters and seven digits")
	}
	var id string
	err := s.db.QueryRowContext(ctx, `INSERT INTO shipment_containers(shipment_id,container_number,container_type,seal_number,tare_weight,gross_weight,net_weight,weight_unit,package_count) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,'')::numeric,NULLIF($6,'')::numeric,NULLIF($7,'')::numeric,NULLIF($8,''),$9) RETURNING id`, shipmentID, number, normalizeCode(p.ContainerType), p.SealNumber, valueOrEmpty(p.TareWeight), valueOrEmpty(p.GrossWeight), valueOrEmpty(p.NetWeight), normalizeCode(p.WeightUnit), p.PackageCount).Scan(&id)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, actor, "containers.create", "shipment_container", id, p)
	return map[string]any{"id": id, "container_number": number}, nil
}
func (s *OperationsService) UpdateContainer(ctx context.Context, actor, id string, p ContainerPayload) error {
	number := normalizePlate(p.ContainerNumber)
	if !validContainerNumber(number) {
		return conflict("INVALID_CONTAINER_NUMBER", "invalid container number")
	}
	r, err := s.db.ExecContext(ctx, `UPDATE shipment_containers SET container_number=$2,container_type=$3,seal_number=NULLIF($4,''),tare_weight=NULLIF($5,'')::numeric,gross_weight=NULLIF($6,'')::numeric,net_weight=NULLIF($7,'')::numeric,weight_unit=NULLIF($8,''),package_count=$9,updated_at=NOW() WHERE id=$1`, id, number, normalizeCode(p.ContainerType), p.SealNumber, valueOrEmpty(p.TareWeight), valueOrEmpty(p.GrossWeight), valueOrEmpty(p.NetWeight), normalizeCode(p.WeightUnit), p.PackageCount)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	s.audit(ctx, actor, "containers.update", "shipment_container", id, p)
	return nil
}
func (s *OperationsService) AddContainerItem(ctx context.Context, actor, containerID string, p ContainerItemPayload) error {
	if !validPositiveDecimal(p.Quantity) {
		return ErrValidation
	}
	r, err := s.db.ExecContext(ctx, `INSERT INTO shipment_container_items(shipment_container_id,shipment_item_id,quantity,quantity_unit,package_count) SELECT $1,$2,$3::numeric,$4,$5 WHERE EXISTS(SELECT 1 FROM shipment_containers c JOIN shipment_items si ON si.shipment_id=c.shipment_id WHERE c.id=$1 AND si.id=$2) ON CONFLICT(shipment_container_id,shipment_item_id) DO UPDATE SET quantity=EXCLUDED.quantity,quantity_unit=EXCLUDED.quantity_unit,package_count=EXCLUDED.package_count`, containerID, p.ShipmentItemID, p.Quantity, normalizeCode(p.QuantityUnit), p.PackageCount)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict("SCOPE_MISMATCH", "container and item have different shipments")
	}
	s.audit(ctx, actor, "containers.items.upsert", "shipment_container", containerID, p)
	return nil
}

var _ = regexp.MustCompile
var _ = time.Now
var _ = errors.Is
var _ = fmt.Sprint
