package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"
)

func TestLogisticsDecimalAndIdentifierContracts(t *testing.T) {
	for _, value := range []string{"0.0001", "18", "99999999999999.9999"} {
		if !validPositiveDecimal(value) {
			t.Fatalf("expected valid positive decimal: %s", value)
		}
	}
	for _, value := range []string{"", "-1", "NaN", "1,000"} {
		if validPositiveDecimal(value) {
			t.Fatalf("expected invalid positive decimal: %s", value)
		}
	}
	if got := normalizePlate("۱۲ ب ـ ۳۴۵ ایران ۶۷"); got != "12ب345ایران67" {
		t.Fatalf("unexpected normalized plate %q", got)
	}
	if !containerPattern.MatchString(normalizeCode("mscu1234567")) {
		t.Fatal("valid ISO-style container number rejected")
	}
	if containerPattern.MatchString("MSCU123") {
		t.Fatal("invalid container number accepted")
	}
	if cmp, ok := decimalCmp("0.1000", "0.1"); !ok || cmp != 0 {
		t.Fatal("decimal comparison must be exact")
	}
	if clampPercent(121) != 100 || clampPercent(-2) != 0 {
		t.Fatal("progress must be clamped")
	}
}

func TestLogisticsStableConflictCode(t *testing.T) {
	err := conflict("INSUFFICIENT_STOCK", "not enough")
	var typed *OperationConflict
	if !errors.As(err, &typed) || typed.Code != "INSUFFICIENT_STOCK" {
		t.Fatalf("unexpected conflict: %#v", err)
	}
}

func TestConcurrentReservationDoesNotOverdrawIntegration(t *testing.T) {
	dsn := os.Getenv("WORKFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("WORKFLOW_TEST_DATABASE_URL is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	suffix := randomDigits(8)
	actor := randomUUIDText()
	customer := randomUUIDText()
	for _, row := range []struct{ id, phone, kind string }{{actor, "+98910" + suffix, "INTERNAL"}, {customer, "+98911" + suffix, "CUSTOMER"}} {
		if _, err = db.ExecContext(ctx, `INSERT INTO users(id,email,password_hash,phone,phone_normalized,user_type,status,is_active) VALUES($1,$2,'x',$3,$3,$4,'ACTIVE',TRUE)`, row.id, row.id+"@example.test", row.phone, row.kind); err != nil {
			t.Fatal(err)
		}
	}
	var templateID int64
	if err = db.QueryRowContext(ctx, `SELECT id FROM workflow_templates WHERE scope_type='ORDER' AND status='PUBLISHED' ORDER BY id LIMIT 1`).Scan(&templateID); err != nil {
		t.Fatal(err)
	}
	var orderID, itemID, batchID, locationID, lotID string
	if err = db.QueryRowContext(ctx, `INSERT INTO orders(order_number,customer_user_id,created_by_user_id,workflow_template_id) VALUES($1,$2,$3,$4) RETURNING id`, `RACE-`+suffix, customer, actor, templateID).Scan(&orderID); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, `INSERT INTO order_items(order_id,stone_name,ordered_quantity,quantity_unit,created_by_user_id) VALUES($1,'تست',20,'TON',$2) RETURNING id`, orderID, actor).Scan(&itemID); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, `INSERT INTO fulfillment_batches(batch_number,order_id,order_item_id,source_type,stone_name,planned_quantity,quantity_unit,status,created_by_user_id) VALUES($1,$2,$3,'MINE','تست',20,'TON','PLANNED',$4) RETURNING id`, `RACE-`+suffix+"-B01", orderID, itemID, actor).Scan(&batchID); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, `INSERT INTO inventory_locations(code,name_fa,location_type,created_by_user_id) VALUES($1,'تست','MINE',$2) RETURNING id`, `RACE-`+suffix, actor).Scan(&locationID); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, `INSERT INTO inventory_lots(lot_number,origin_type,current_location_id,stone_name,initial_quantity,available_quantity,quantity_unit,status,created_by_user_id) VALUES($1,'MINE',$2,'تست',10,10,'TON','AVAILABLE',$3) RETURNING id`, `LOT-`+suffix, locationID, actor).Scan(&lotID); err != nil {
		t.Fatal(err)
	}
	payload := ReservationPayload{OrderID: orderID, OrderItemID: itemID, BatchID: batchID, Quantity: "7", QuantityUnit: "TON"}
	service := NewOperationsService(db)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := service.CreateReservation(ctx, actor, lotID, "race-"+randomDigits(8), payload)
			errs <- e
		}()
	}
	wg.Wait()
	close(errs)
	success := 0
	for e := range errs {
		if e == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("expected exactly one successful reservation, got %d", success)
	}
	var available, reserved string
	if err = db.QueryRowContext(ctx, `SELECT available_quantity::text,reserved_quantity::text FROM inventory_lots WHERE id=$1`, lotID).Scan(&available, &reserved); err != nil {
		t.Fatal(err)
	}
	if cmp, _ := decimalCmp(available, "3"); cmp != 0 {
		t.Fatalf("available=%s", available)
	}
	if cmp, _ := decimalCmp(reserved, "7"); cmp != 0 {
		t.Fatalf("reserved=%s", reserved)
	}
}

func TestBatchScopeAndManualBranchIntegration(t *testing.T) {
	dsn := os.Getenv("WORKFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("WORKFLOW_TEST_DATABASE_URL is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	suffix := randomDigits(8)
	actor := randomUUIDText()
	customer := randomUUIDText()
	for _, row := range []struct{ id, phone, kind string }{{actor, "+98930" + suffix, "INTERNAL"}, {customer, "+98931" + suffix, "CUSTOMER"}} {
		if _, err = db.ExecContext(ctx, `INSERT INTO users(id,email,password_hash,phone,phone_normalized,user_type,status,is_active) VALUES($1,$2,'x',$3,$3,$4,'ACTIVE',TRUE)`, row.id, row.id+"@example.test", row.phone, row.kind); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id) SELECT $1,id FROM roles WHERE code='SUPPLY'`, actor); err != nil {
		t.Fatal(err)
	}
	var orderTemplate int64
	if err = db.QueryRowContext(ctx, `SELECT id FROM workflow_templates WHERE scope_type='ORDER' AND status='PUBLISHED' ORDER BY id LIMIT 1`).Scan(&orderTemplate); err != nil {
		t.Fatal(err)
	}
	var orderID string
	if err = db.QueryRowContext(ctx, `INSERT INTO orders(order_number,customer_user_id,created_by_user_id,workflow_template_id) VALUES($1,$2,$3,$4) RETURNING id`, `BRANCH-`+suffix, customer, actor, orderTemplate).Scan(&orderID); err != nil {
		t.Fatal(err)
	}
	service := NewOperationsService(db)
	item, err := service.CreateOrderItem(ctx, actor, orderID, OrderItemPayload{StoneName: "مرمریت", OrderedQuantity: "10", QuantityUnit: "TON", ProgressWeight: "1"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := service.CreateBatch(ctx, actor, orderID, "batch-create-"+suffix, BatchPayload{OrderItemID: item.ID, SourceType: "MINE", StoneName: "مرمریت", PlannedQuantity: "10", QuantityUnit: "TON"})
	if err != nil {
		t.Fatal(err)
	}
	retry, err := service.CreateBatch(ctx, actor, orderID, "batch-create-"+suffix, BatchPayload{OrderItemID: item.ID, SourceType: "MINE", StoneName: "مرمریت", PlannedQuantity: "10", QuantityUnit: "TON"})
	if err != nil {
		t.Fatal(err)
	}
	if retry.ID != batch.ID {
		t.Fatalf("idempotent batch retry changed id: %s != %s", retry.ID, batch.ID)
	}
	var scope, scopeID, current string
	if err = db.QueryRowContext(ctx, `SELECT scope_type,scope_id,current_step_instance_id FROM workflow_instances WHERE id=$1`, *batch.WorkflowInstanceID).Scan(&scope, &scopeID, &current); err != nil {
		t.Fatal(err)
	}
	if scope != "BATCH" || scopeID != batch.ID {
		t.Fatalf("unexpected scope %s/%s", scope, scopeID)
	}
	if err = service.StartWorkflowStep(ctx, actor, current, ""); err != nil {
		t.Fatal(err)
	}
	if err = service.SubmitWorkflowStep(ctx, actor, current, StepValuesPayload{Values: map[string]json.RawMessage{}}); err != nil {
		t.Fatal(err)
	}
	transitions, err := service.GetRuntimeTransitions(ctx, actor, current)
	if err != nil {
		t.Fatal(err)
	}
	if len(transitions) != 2 {
		t.Fatalf("expected two manual source branches, got %d", len(transitions))
	}
	selected, err := service.SelectWorkflowTransition(ctx, actor, current, "branch-select-"+suffix, SelectTransitionPayload{TransitionCode: "USE_EXISTING_STOCK"})
	if err != nil {
		t.Fatal(err)
	}
	if selected["target_step_instance_id"] == "" {
		t.Fatal("branch did not activate target")
	}
	var waiting, notSelected int
	if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FILTER(WHERE status='WAITING_FOR_ASSIGNEE' AND path_state='INCLUDED'),COUNT(*) FILTER(WHERE path_state='NOT_SELECTED') FROM workflow_step_instances WHERE workflow_instance_id=$1`, *batch.WorkflowInstanceID).Scan(&waiting, &notSelected); err != nil {
		t.Fatal(err)
	}
	if waiting != 1 || notSelected == 0 {
		t.Fatalf("branch states waiting=%d not_selected=%d", waiting, notSelected)
	}
}

func TestInventoryShipmentTraceabilityIntegration(t *testing.T) {
	dsn := os.Getenv("WORKFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("WORKFLOW_TEST_DATABASE_URL is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	suffix := randomDigits(8)
	actor := randomUUIDText()
	customer := randomUUIDText()
	for _, row := range []struct{ id, phone, kind string }{{actor, "+98940" + suffix, "INTERNAL"}, {customer, "+98941" + suffix, "CUSTOMER"}} {
		if _, err = db.ExecContext(ctx, `INSERT INTO users(id,email,password_hash,phone,phone_normalized,user_type,status,is_active) VALUES($1,$2,'x',$3,$3,$4,'ACTIVE',TRUE)`, row.id, row.id+"@example.test", row.phone, row.kind); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id) SELECT $1,id FROM roles WHERE code='SUPER_ADMIN'`, actor); err != nil {
		t.Fatal(err)
	}
	var orderTemplate int64
	if err = db.QueryRowContext(ctx, `SELECT id FROM workflow_templates WHERE scope_type='ORDER' AND status='PUBLISHED' ORDER BY id LIMIT 1`).Scan(&orderTemplate); err != nil {
		t.Fatal(err)
	}
	var orderID string
	if err = db.QueryRowContext(ctx, `INSERT INTO orders(order_number,customer_user_id,created_by_user_id,workflow_template_id) VALUES($1,$2,$3,$4) RETURNING id`, `TRACE-`+suffix, customer, actor, orderTemplate).Scan(&orderID); err != nil {
		t.Fatal(err)
	}
	service := NewOperationsService(db)
	item, err := service.CreateOrderItem(ctx, actor, orderID, OrderItemPayload{StoneName: "تراورتن", OrderedQuantity: "10", QuantityUnit: "TON", ProgressWeight: "1"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := service.CreateBatch(ctx, actor, orderID, "trace-batch-"+suffix, BatchPayload{OrderItemID: item.ID, SourceType: "MINE", StoneName: "تراورتن", PlannedQuantity: "10", QuantityUnit: "TON"})
	if err != nil {
		t.Fatal(err)
	}
	origin, err := service.CreateLocation(ctx, actor, LocationPayload{Code: "ORIGIN-" + suffix, NameFA: "مبدأ", LocationType: "FACTORY"})
	if err != nil {
		t.Fatal(err)
	}
	destination, err := service.CreateLocation(ctx, actor, LocationPayload{Code: "DEST-" + suffix, NameFA: "مقصد", LocationType: "CUSTOMER_LOCATION"})
	if err != nil {
		t.Fatal(err)
	}
	lot, err := service.ReceiveInventory(ctx, actor, "trace-receipt-"+suffix, ReceiptPayload{Lot: LotPayload{OriginType: "MINE", LocationID: origin.ID, StoneName: "تراورتن", Quantity: "10", QuantityUnit: "TON"}, OrderID: &orderID, BatchID: &batch.ID, Reason: "test receipt"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.CreateReservation(ctx, actor, lot.ID, "trace-reserve-"+suffix, ReservationPayload{OrderID: orderID, OrderItemID: item.ID, BatchID: batch.ID, Quantity: "10", QuantityUnit: "TON"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.CreatePackaging(ctx, actor, batch.ID, "trace-package-"+suffix, PackagingPayload{InventoryLotID: lot.ID, PackageType: "CRATE", Quantity: "10", QuantityUnit: "TON", CustomerVisible: true}); err != nil {
		t.Fatal(err)
	}
	shipment, err := service.CreateShipment(ctx, actor, orderID, "trace-shipment-"+suffix, ShipmentPayload{ShipmentType: "DOMESTIC_TRUCK", OriginLocationID: origin.ID, DestinationLocationID: &destination.ID})
	if err != nil {
		t.Fatal(err)
	}
	shipmentItem, err := service.AddShipmentItem(ctx, actor, shipment.ID, ShipmentItemPayload{BatchID: batch.ID, InventoryLotID: lot.ID, PlannedQuantity: "10", QuantityUnit: "TON", PackageCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	quantities := []ShipmentQuantity{{ShipmentItemID: shipmentItem.ID, Quantity: "10"}}
	if _, err = service.LoadShipment(ctx, actor, shipment.ID, "trace-load-"+suffix, ShipmentOperationPayload{Items: quantities, FinalizeLoading: true, Reason: "loaded"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.DispatchShipment(ctx, actor, shipment.ID, "trace-dispatch-"+suffix, ShipmentOperationPayload{Reason: "dispatch"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.ArriveShipment(ctx, actor, shipment.ID, "trace-arrive-"+suffix, ShipmentOperationPayload{Reason: "arrival"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.DeliverShipment(ctx, actor, shipment.ID, "trace-deliver-"+suffix, ShipmentOperationPayload{Items: quantities, FinalizeDelivery: true, ReceiverName: "گیرنده", Reason: "delivery"}, false); err != nil {
		t.Fatal(err)
	}
	final, err := service.GetShipment(ctx, actor, shipment.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if final.Status != "DELIVERED" || !final.CustomerVisible {
		t.Fatalf("unexpected shipment state: %#v", final)
	}
	var deliveredLot string
	if err = db.QueryRowContext(ctx, `SELECT sei.inventory_lot_id FROM shipment_event_items sei JOIN shipment_events se ON se.id=sei.shipment_event_id WHERE se.shipment_id=$1 AND se.event_type='DELIVERY' LIMIT 1`, shipment.ID).Scan(&deliveredLot); err != nil {
		t.Fatal(err)
	}
	trace, err := service.LotTraceability(ctx, deliveredLot)
	if err != nil {
		t.Fatal(err)
	}
	if len(trace) < 2 {
		t.Fatalf("expected delivered lot lineage, got %#v", trace)
	}
	var movementCount int
	if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_movements WHERE shipment_id=$1`, shipment.ID).Scan(&movementCount); err != nil {
		t.Fatal(err)
	}
	if movementCount < 3 {
		t.Fatalf("expected immutable shipment movements, got %d", movementCount)
	}
}
