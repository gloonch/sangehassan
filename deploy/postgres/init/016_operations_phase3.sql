-- SangeHassan Operations Dashboard Phase 3
-- Batch, inventory, shipment and controlled branching. Safe on top of 014/015.

CREATE SEQUENCE IF NOT EXISTS operations_number_seq;

ALTER TABLE workflow_templates
  ADD COLUMN IF NOT EXISTS scope_type TEXT NOT NULL DEFAULT 'ORDER',
  ADD COLUMN IF NOT EXISTS max_iterations INT NOT NULL DEFAULT 20;
ALTER TABLE workflow_template_steps
  ADD COLUMN IF NOT EXISTS is_entry BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS domain_event_code TEXT;

UPDATE workflow_templates SET scope_type='ORDER' WHERE scope_type IS NULL;
UPDATE workflow_template_steps s SET is_entry=TRUE
WHERE s.sequence_number=(SELECT MIN(x.sequence_number) FROM workflow_template_steps x WHERE x.workflow_template_id=s.workflow_template_id AND x.is_active)
  AND NOT EXISTS(SELECT 1 FROM workflow_template_steps x WHERE x.workflow_template_id=s.workflow_template_id AND x.is_entry);

ALTER TABLE workflow_instances DROP CONSTRAINT IF EXISTS workflow_instances_order_id_key;
ALTER TABLE workflow_instances DROP CONSTRAINT IF EXISTS workflow_instances_status_check;
ALTER TABLE workflow_instances
  ADD COLUMN IF NOT EXISTS scope_type TEXT NOT NULL DEFAULT 'ORDER',
  ADD COLUMN IF NOT EXISTS scope_id UUID,
  ADD COLUMN IF NOT EXISTS parent_workflow_instance_id UUID REFERENCES workflow_instances(id) ON DELETE SET NULL;
UPDATE workflow_instances SET scope_type='ORDER',scope_id=order_id WHERE scope_id IS NULL;
ALTER TABLE workflow_instances ALTER COLUMN scope_id SET NOT NULL;
ALTER TABLE workflow_instances ADD CONSTRAINT workflow_instances_status_check
  CHECK(status IN ('IN_PROGRESS','WAITING_FOR_TRANSITION','BLOCKED','COMPLETED','CANCELLED'));
CREATE UNIQUE INDEX IF NOT EXISTS uq_workflow_active_scope
  ON workflow_instances(scope_type,scope_id) WHERE status<>'CANCELLED';
CREATE INDEX IF NOT EXISTS idx_workflow_parent ON workflow_instances(parent_workflow_instance_id);

DROP INDEX IF EXISTS uq_workflow_instance_step_sequence;
ALTER TABLE workflow_step_instances DROP CONSTRAINT IF EXISTS workflow_step_instances_status_check;
ALTER TABLE workflow_step_instances
  ADD COLUMN IF NOT EXISTS iteration_number INT NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS path_state TEXT NOT NULL DEFAULT 'INCLUDED',
  ADD COLUMN IF NOT EXISTS predecessor_step_instance_id UUID REFERENCES workflow_step_instances(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS result_code TEXT,
  ADD COLUMN IF NOT EXISTS domain_event_code TEXT;
ALTER TABLE workflow_step_instances ADD CONSTRAINT workflow_step_instances_status_check
  CHECK(status IN ('NOT_STARTED','WAITING_FOR_ASSIGNEE','IN_PROGRESS','SUBMITTED','WAITING_FOR_APPROVAL','WAITING_FOR_TRANSITION','APPROVED','HAS_MISMATCH','NEEDS_CORRECTION','BLOCKED','COMPLETED','SKIPPED','CANCELLED'));
CREATE UNIQUE INDEX IF NOT EXISTS uq_workflow_instance_step_iteration
  ON workflow_step_instances(workflow_instance_id,sequence_number,iteration_number) WHERE sequence_number IS NOT NULL;

CREATE TABLE IF NOT EXISTS order_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id INT REFERENCES products(id) ON DELETE SET NULL,
  stone_category TEXT, stone_name TEXT NOT NULL, stone_variant TEXT, finish_type TEXT, cut_type TEXT,
  thickness_value NUMERIC(18,4), thickness_unit TEXT, width_value NUMERIC(18,4), length_value NUMERIC(18,4), dimension_unit TEXT,
  ordered_quantity NUMERIC(18,4) NOT NULL, quantity_unit TEXT NOT NULL, quality_grade TEXT, color TEXT, pattern TEXT,
  progress_weight NUMERIC(18,4) NOT NULL DEFAULT 1, requires_production BOOLEAN NOT NULL DEFAULT TRUE,
  requires_packaging BOOLEAN NOT NULL DEFAULT TRUE, notes TEXT,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(ordered_quantity>0), CHECK(progress_weight>0)
);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id,created_at);

CREATE TABLE IF NOT EXISTS order_item_quantity_conversions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), order_item_id UUID NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
  from_quantity NUMERIC(18,4) NOT NULL, from_unit TEXT NOT NULL, to_quantity NUMERIC(18,4) NOT NULL, to_unit TEXT NOT NULL,
  reason TEXT NOT NULL, created_by_user_id UUID NOT NULL REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(from_quantity>0), CHECK(to_quantity>0), CHECK(from_unit<>to_unit), UNIQUE(order_item_id,from_unit,to_unit)
);

CREATE TABLE IF NOT EXISTS inventory_locations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), code TEXT NOT NULL UNIQUE, name_fa TEXT NOT NULL, location_type TEXT NOT NULL,
  address TEXT, city TEXT, province TEXT, country_code CHAR(2) NOT NULL DEFAULT 'IR', latitude NUMERIC(10,7), longitude NUMERIC(10,7),
  is_active BOOLEAN NOT NULL DEFAULT TRUE, created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fulfillment_batches (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), batch_number TEXT NOT NULL UNIQUE,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT, order_item_id UUID NOT NULL REFERENCES order_items(id) ON DELETE RESTRICT,
  parent_batch_id UUID REFERENCES fulfillment_batches(id) ON DELETE RESTRICT, source_type TEXT NOT NULL,
  source_location_id UUID REFERENCES inventory_locations(id) ON DELETE RESTRICT, target_location_id UUID REFERENCES inventory_locations(id) ON DELETE RESTRICT,
  stone_category TEXT, stone_name TEXT NOT NULL, stone_variant TEXT, finish_type TEXT, cut_type TEXT,
  thickness_value NUMERIC(18,4), thickness_unit TEXT, planned_quantity NUMERIC(18,4) NOT NULL,
  actual_quantity NUMERIC(18,4) NOT NULL DEFAULT 0, quantity_unit TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'DRAFT',
  priority TEXT NOT NULL DEFAULT 'NORMAL', is_required BOOLEAN NOT NULL DEFAULT TRUE,
  workflow_instance_id UUID REFERENCES workflow_instances(id) ON DELETE SET NULL,
  estimated_ready_at TIMESTAMPTZ, actual_ready_at TIMESTAMPTZ,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), cancelled_at TIMESTAMPTZ,
  CHECK(planned_quantity>=0), CHECK(actual_quantity>=0)
);
CREATE INDEX IF NOT EXISTS idx_batches_order ON fulfillment_batches(order_id,status);
CREATE INDEX IF NOT EXISTS idx_batches_item ON fulfillment_batches(order_item_id,status);

CREATE TABLE IF NOT EXISTS batch_merge_members (
  merged_batch_id UUID NOT NULL REFERENCES fulfillment_batches(id) ON DELETE RESTRICT,
  source_batch_id UUID NOT NULL REFERENCES fulfillment_batches(id) ON DELETE RESTRICT,
  quantity NUMERIC(18,4) NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY(merged_batch_id,source_batch_id), CHECK(quantity>0), CHECK(merged_batch_id<>source_batch_id)
);

CREATE TABLE IF NOT EXISTS inventory_lots (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), lot_number TEXT NOT NULL UNIQUE,
  parent_lot_id UUID REFERENCES inventory_lots(id) ON DELETE RESTRICT, origin_type TEXT NOT NULL, origin_reference_id TEXT,
  current_location_id UUID NOT NULL REFERENCES inventory_locations(id) ON DELETE RESTRICT,
  stone_category TEXT, stone_name TEXT NOT NULL, stone_variant TEXT, quality_grade TEXT, finish_type TEXT, cut_type TEXT,
  thickness_value NUMERIC(18,4), thickness_unit TEXT, width_value NUMERIC(18,4), length_value NUMERIC(18,4), dimension_unit TEXT,
  initial_quantity NUMERIC(18,4) NOT NULL, available_quantity NUMERIC(18,4) NOT NULL, reserved_quantity NUMERIC(18,4) NOT NULL DEFAULT 0,
  quantity_unit TEXT NOT NULL, secondary_quantity NUMERIC(18,4), secondary_unit TEXT, status TEXT NOT NULL DEFAULT 'AVAILABLE',
  received_at TIMESTAMPTZ, produced_at TIMESTAMPTZ, created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(initial_quantity>=0), CHECK(available_quantity>=0), CHECK(reserved_quantity>=0),
  CHECK(available_quantity+reserved_quantity<=initial_quantity)
);
CREATE INDEX IF NOT EXISTS idx_inventory_lots_location ON inventory_lots(current_location_id,status);
CREATE INDEX IF NOT EXISTS idx_inventory_lots_stone ON inventory_lots(stone_name,stone_variant,quantity_unit);

CREATE TABLE IF NOT EXISTS inventory_stock_policies (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), location_id UUID NOT NULL REFERENCES inventory_locations(id) ON DELETE CASCADE,
  stone_category TEXT, stone_name TEXT NOT NULL, stone_variant TEXT NOT NULL DEFAULT '', quantity_unit TEXT NOT NULL,
  low_stock_threshold NUMERIC(18,4) NOT NULL, is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(location_id,stone_name,stone_variant,quantity_unit), CHECK(low_stock_threshold>=0)
);

CREATE TABLE IF NOT EXISTS inventory_reservations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), inventory_lot_id UUID NOT NULL REFERENCES inventory_lots(id) ON DELETE RESTRICT,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT, order_item_id UUID NOT NULL REFERENCES order_items(id) ON DELETE RESTRICT,
  batch_id UUID NOT NULL REFERENCES fulfillment_batches(id) ON DELETE RESTRICT,
  reserved_quantity NUMERIC(18,4) NOT NULL, consumed_quantity NUMERIC(18,4) NOT NULL DEFAULT 0, quantity_unit TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'ACTIVE', reserved_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  reserved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), released_at TIMESTAMPTZ, consumed_at TIMESTAMPTZ, expires_at TIMESTAMPTZ,
  release_reason TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(reserved_quantity>0), CHECK(consumed_quantity>=0 AND consumed_quantity<=reserved_quantity)
);
CREATE INDEX IF NOT EXISTS idx_reservations_lot_active ON inventory_reservations(inventory_lot_id,status);
CREATE INDEX IF NOT EXISTS idx_reservations_batch ON inventory_reservations(batch_id,status);

CREATE TABLE IF NOT EXISTS vehicles (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), vehicle_type TEXT NOT NULL, plate_number TEXT, plate_normalized TEXT,
  trailer_number TEXT, capacity_value NUMERIC(18,4), capacity_unit TEXT, owner_name TEXT, carrier_name TEXT,
  driver_user_id UUID REFERENCES users(id) ON DELETE SET NULL, is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_vehicle_plate_normalized ON vehicles(plate_normalized) WHERE plate_normalized IS NOT NULL;

CREATE TABLE IF NOT EXISTS shipments (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), shipment_number TEXT NOT NULL UNIQUE,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT, shipment_type TEXT NOT NULL,
  origin_location_id UUID NOT NULL REFERENCES inventory_locations(id) ON DELETE RESTRICT,
  destination_location_id UUID REFERENCES inventory_locations(id) ON DELETE RESTRICT, status TEXT NOT NULL DEFAULT 'DRAFT',
  driver_user_id UUID REFERENCES users(id) ON DELETE SET NULL, external_driver_name TEXT, external_driver_phone TEXT,
  carrier_name TEXT, vehicle_id UUID REFERENCES vehicles(id) ON DELETE SET NULL,
  planned_departure_at TIMESTAMPTZ, actual_departure_at TIMESTAMPTZ, estimated_arrival_at TIMESTAMPTZ, actual_arrival_at TIMESTAMPTZ,
  delivery_contact_name TEXT, delivery_contact_phone TEXT, delivery_address TEXT,
  workflow_instance_id UUID REFERENCES workflow_instances(id) ON DELETE SET NULL,
  customer_title_fa TEXT NOT NULL DEFAULT 'محموله سفارش', customer_visible BOOLEAN NOT NULL DEFAULT FALSE,
  notes TEXT, created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), cancelled_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_shipments_order ON shipments(order_id,status);
CREATE INDEX IF NOT EXISTS idx_shipments_driver ON shipments(driver_user_id,status);

CREATE TABLE IF NOT EXISTS shipment_orders (
  shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
  is_primary BOOLEAN NOT NULL DEFAULT FALSE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY(shipment_id,order_id)
);

CREATE TABLE IF NOT EXISTS shipment_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
  batch_id UUID NOT NULL REFERENCES fulfillment_batches(id) ON DELETE RESTRICT,
  inventory_lot_id UUID NOT NULL REFERENCES inventory_lots(id) ON DELETE RESTRICT,
  planned_quantity NUMERIC(18,4) NOT NULL, loaded_quantity NUMERIC(18,4) NOT NULL DEFAULT 0,
  delivered_quantity NUMERIC(18,4) NOT NULL DEFAULT 0, quantity_unit TEXT NOT NULL,
  package_count INT NOT NULL DEFAULT 0, bundle_count INT NOT NULL DEFAULT 0, notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(shipment_id,batch_id,inventory_lot_id), CHECK(planned_quantity>0),
  CHECK(loaded_quantity>=0 AND loaded_quantity<=planned_quantity),
  CHECK(delivered_quantity>=0 AND delivered_quantity<=loaded_quantity), CHECK(package_count>=0), CHECK(bundle_count>=0)
);

CREATE TABLE IF NOT EXISTS inventory_operation_requests (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  operation_type TEXT NOT NULL, idempotency_key TEXT NOT NULL, payload_hash TEXT NOT NULL,
  response_json JSONB, status TEXT NOT NULL DEFAULT 'STARTED', created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), completed_at TIMESTAMPTZ,
  UNIQUE(actor_user_id,operation_type,idempotency_key), CHECK(status IN ('STARTED','COMPLETED','FAILED'))
);

CREATE TABLE IF NOT EXISTS inventory_movements (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), movement_number TEXT NOT NULL UNIQUE, operation_group_id UUID NOT NULL,
  movement_type TEXT NOT NULL, inventory_lot_id UUID NOT NULL REFERENCES inventory_lots(id) ON DELETE RESTRICT,
  source_location_id UUID REFERENCES inventory_locations(id) ON DELETE RESTRICT,
  destination_location_id UUID REFERENCES inventory_locations(id) ON DELETE RESTRICT,
  order_id UUID REFERENCES orders(id) ON DELETE RESTRICT, batch_id UUID REFERENCES fulfillment_batches(id) ON DELETE RESTRICT,
  shipment_id UUID REFERENCES shipments(id) ON DELETE RESTRICT, reservation_id UUID REFERENCES inventory_reservations(id) ON DELETE RESTRICT,
  quantity NUMERIC(18,4) NOT NULL, quantity_unit TEXT NOT NULL,
  before_available_quantity NUMERIC(18,4) NOT NULL, after_available_quantity NUMERIC(18,4) NOT NULL,
  before_reserved_quantity NUMERIC(18,4) NOT NULL, after_reserved_quantity NUMERIC(18,4) NOT NULL,
  reference_type TEXT, reference_id TEXT, reason TEXT, reversal_of_movement_id UUID REFERENCES inventory_movements(id) ON DELETE RESTRICT,
  performed_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), CHECK(quantity>0)
);
CREATE INDEX IF NOT EXISTS idx_movements_lot ON inventory_movements(inventory_lot_id,occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_movements_operation ON inventory_movements(operation_group_id);

CREATE TABLE IF NOT EXISTS inventory_lot_conversions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), operation_group_id UUID NOT NULL UNIQUE,
  batch_id UUID NOT NULL REFERENCES fulfillment_batches(id) ON DELETE RESTRICT,
  workflow_step_instance_id UUID REFERENCES workflow_step_instances(id) ON DELETE SET NULL,
  input_lot_id UUID NOT NULL REFERENCES inventory_lots(id) ON DELETE RESTRICT,
  output_lot_id UUID NOT NULL REFERENCES inventory_lots(id) ON DELETE RESTRICT,
  order_item_conversion_id UUID REFERENCES order_item_quantity_conversions(id) ON DELETE RESTRICT,
  input_quantity NUMERIC(18,4) NOT NULL, input_unit TEXT NOT NULL,
  output_quantity NUMERIC(18,4) NOT NULL, output_unit TEXT NOT NULL,
  waste_quantity NUMERIC(18,4) NOT NULL DEFAULT 0, waste_unit TEXT NOT NULL,
  conversion_type TEXT NOT NULL, created_by_user_id UUID NOT NULL REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(input_quantity>0), CHECK(output_quantity>=0), CHECK(waste_quantity>=0), CHECK(input_lot_id<>output_lot_id)
);

CREATE TABLE IF NOT EXISTS packaging_units (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), package_number TEXT NOT NULL UNIQUE,
  batch_id UUID NOT NULL REFERENCES fulfillment_batches(id) ON DELETE RESTRICT,
  inventory_lot_id UUID NOT NULL REFERENCES inventory_lots(id) ON DELETE RESTRICT,
  package_type TEXT NOT NULL, quantity NUMERIC(18,4) NOT NULL, quantity_unit TEXT NOT NULL,
  gross_weight NUMERIC(18,4), net_weight NUMERIC(18,4), weight_unit TEXT,
  width_value NUMERIC(18,4), length_value NUMERIC(18,4), height_value NUMERIC(18,4), dimension_unit TEXT,
  status TEXT NOT NULL DEFAULT 'DRAFT', customer_visible BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), CHECK(quantity>0)
);

CREATE TABLE IF NOT EXISTS shipment_package_assignments (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), packaging_unit_id UUID NOT NULL REFERENCES packaging_units(id) ON DELETE RESTRICT,
  shipment_item_id UUID NOT NULL REFERENCES shipment_items(id) ON DELETE RESTRICT,
  assigned_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL, assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  released_at TIMESTAMPTZ, release_reason TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_package_active_assignment ON shipment_package_assignments(packaging_unit_id) WHERE released_at IS NULL;

CREATE TABLE IF NOT EXISTS shipment_containers (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
  container_number TEXT NOT NULL, container_type TEXT NOT NULL, seal_number TEXT,
  tare_weight NUMERIC(18,4), gross_weight NUMERIC(18,4), net_weight NUMERIC(18,4), weight_unit TEXT,
  package_count INT NOT NULL DEFAULT 0, loaded_at TIMESTAMPTZ, verified_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(shipment_id,container_number), CHECK(package_count>=0)
);

CREATE TABLE IF NOT EXISTS shipment_container_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), shipment_container_id UUID NOT NULL REFERENCES shipment_containers(id) ON DELETE CASCADE,
  shipment_item_id UUID NOT NULL REFERENCES shipment_items(id) ON DELETE RESTRICT,
  quantity NUMERIC(18,4) NOT NULL, quantity_unit TEXT NOT NULL, package_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE(shipment_container_id,shipment_item_id),
  CHECK(quantity>0), CHECK(package_count>=0)
);

CREATE TABLE IF NOT EXISTS shipment_events (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
  operation_group_id UUID NOT NULL UNIQUE, event_type TEXT NOT NULL, reason TEXT,
  receiver_name TEXT, receiver_phone TEXT, proof_file_id UUID REFERENCES workflow_files(id) ON DELETE SET NULL,
  performed_by_user_id UUID NOT NULL REFERENCES users(id), occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS shipment_event_items (
  shipment_event_id UUID NOT NULL REFERENCES shipment_events(id) ON DELETE CASCADE,
  shipment_item_id UUID NOT NULL REFERENCES shipment_items(id) ON DELETE RESTRICT,
  inventory_lot_id UUID REFERENCES inventory_lots(id) ON DELETE RESTRICT,
  quantity NUMERIC(18,4) NOT NULL, quantity_unit TEXT NOT NULL,
  PRIMARY KEY(shipment_event_id,shipment_item_id,inventory_lot_id), CHECK(quantity>0)
);

CREATE TABLE IF NOT EXISTS operational_cost_entries (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), entity_type TEXT NOT NULL, entity_id UUID NOT NULL,
  cost_type TEXT NOT NULL, amount NUMERIC(18,2) NOT NULL, currency CHAR(3) NOT NULL,
  status TEXT NOT NULL DEFAULT 'DRAFT', notes TEXT, incurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by_user_id UUID NOT NULL REFERENCES users(id), approved_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  approved_at TIMESTAMPTZ, voided_at TIMESTAMPTZ, void_reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), CHECK(amount>=0)
);
CREATE INDEX IF NOT EXISTS idx_operational_cost_entity ON operational_cost_entries(entity_type,entity_id,status);

CREATE TABLE IF NOT EXISTS workflow_step_transitions (
  id BIGSERIAL PRIMARY KEY, workflow_template_id BIGINT NOT NULL REFERENCES workflow_templates(id) ON DELETE CASCADE,
  source_step_id BIGINT NOT NULL REFERENCES workflow_template_steps(id) ON DELETE CASCADE,
  target_step_id BIGINT NOT NULL REFERENCES workflow_template_steps(id) ON DELETE RESTRICT,
  transition_code TEXT NOT NULL, label_fa TEXT NOT NULL, transition_type TEXT NOT NULL,
  result_code TEXT, is_default BOOLEAN NOT NULL DEFAULT FALSE,
  requires_permission_code TEXT REFERENCES permissions(code) ON UPDATE CASCADE,
  requires_reason BOOLEAN NOT NULL DEFAULT FALSE, sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workflow_template_id,source_step_id,transition_code), CHECK(source_step_id<>target_step_id)
);
CREATE INDEX IF NOT EXISTS idx_template_transitions_source ON workflow_step_transitions(source_step_id,sort_order);

CREATE TABLE IF NOT EXISTS workflow_instance_step_transitions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  source_template_step_id BIGINT, target_template_step_id BIGINT, source_step_code TEXT NOT NULL, target_step_code TEXT NOT NULL,
  transition_code TEXT NOT NULL, label_fa TEXT NOT NULL, transition_type TEXT NOT NULL, result_code TEXT,
  is_default BOOLEAN NOT NULL DEFAULT FALSE, requires_permission_code TEXT, requires_reason BOOLEAN NOT NULL DEFAULT FALSE,
  sort_order INT NOT NULL DEFAULT 0, UNIQUE(workflow_instance_id,source_step_code,transition_code)
);

CREATE TABLE IF NOT EXISTS workflow_transition_selections (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  source_step_instance_id UUID NOT NULL REFERENCES workflow_step_instances(id) ON DELETE RESTRICT,
  transition_snapshot_id UUID NOT NULL REFERENCES workflow_instance_step_transitions(id) ON DELETE RESTRICT,
  target_step_instance_id UUID NOT NULL REFERENCES workflow_step_instances(id) ON DELETE RESTRICT,
  selected_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL, result_code TEXT, reason TEXT,
  is_override BOOLEAN NOT NULL DEFAULT FALSE, selected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(source_step_instance_id)
);

CREATE TABLE IF NOT EXISTS workflow_child_dependencies (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), parent_workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  parent_step_instance_id UUID REFERENCES workflow_step_instances(id) ON DELETE SET NULL,
  child_workflow_instance_id UUID NOT NULL UNIQUE REFERENCES workflow_instances(id) ON DELETE CASCADE,
  is_blocking BOOLEAN NOT NULL DEFAULT TRUE, status TEXT NOT NULL DEFAULT 'OPEN', action_item_id UUID REFERENCES action_items(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), completed_at TIMESTAMPTZ,
  CHECK(parent_workflow_instance_id<>child_workflow_instance_id), CHECK(status IN ('OPEN','COMPLETED','CANCELLED'))
);

CREATE TABLE IF NOT EXISTS workflow_domain_operation_completions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), workflow_step_instance_id UUID NOT NULL REFERENCES workflow_step_instances(id) ON DELETE CASCADE,
  domain_event_code TEXT NOT NULL, entity_type TEXT NOT NULL, entity_id UUID NOT NULL,
  operation_group_id UUID, completed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE(workflow_step_instance_id,domain_event_code,entity_id)
);

ALTER TABLE workflow_files
  ALTER COLUMN workflow_step_instance_id DROP NOT NULL,
  ADD COLUMN IF NOT EXISTS entity_type TEXT,
  ADD COLUMN IF NOT EXISTS entity_id UUID;
UPDATE workflow_files SET entity_type='WORKFLOW',entity_id=workflow_instance_id WHERE entity_type IS NULL OR entity_id IS NULL;
ALTER TABLE workflow_files ALTER COLUMN entity_type SET NOT NULL;
ALTER TABLE workflow_files ALTER COLUMN entity_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_workflow_files_entity ON workflow_files(entity_type,entity_id,customer_visible);

ALTER TABLE workflow_discrepancies
  ALTER COLUMN workflow_instance_id DROP NOT NULL,
  ADD COLUMN IF NOT EXISTS batch_id UUID REFERENCES fulfillment_batches(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS shipment_id UUID REFERENCES shipments(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS inventory_movement_id UUID REFERENCES inventory_movements(id) ON DELETE SET NULL;
ALTER TABLE action_items
  ADD COLUMN IF NOT EXISTS batch_id UUID REFERENCES fulfillment_batches(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS shipment_id UUID REFERENCES shipments(id) ON DELETE CASCADE;

DO $$ BEGIN
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_template_scope') THEN
    ALTER TABLE workflow_templates ADD CONSTRAINT chk_workflow_template_scope CHECK(scope_type IN ('ORDER','BATCH','SHIPMENT','INSTALLATION'));
  END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_template_iterations') THEN
    ALTER TABLE workflow_templates ADD CONSTRAINT chk_workflow_template_iterations CHECK(max_iterations BETWEEN 1 AND 100);
  END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_instance_scope') THEN
    ALTER TABLE workflow_instances ADD CONSTRAINT chk_workflow_instance_scope CHECK(scope_type IN ('ORDER','BATCH','SHIPMENT','INSTALLATION'));
  END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_step_path_state') THEN
    ALTER TABLE workflow_step_instances ADD CONSTRAINT chk_workflow_step_path_state CHECK(path_state IN ('INCLUDED','NOT_SELECTED','EXCLUDED'));
  END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_result_code') THEN
    ALTER TABLE workflow_step_instances ADD CONSTRAINT chk_workflow_result_code CHECK(result_code IS NULL OR result_code IN ('APPROVED','REJECTED','HAS_DISCREPANCY','CORRECTION_REQUIRED','CUSTOMER_CANCELLED','PAYMENT_PENDING'));
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_order_item_unit') THEN ALTER TABLE order_items ADD CONSTRAINT chk_order_item_unit CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_order_item_conversion_units') THEN ALTER TABLE order_item_quantity_conversions ADD CONSTRAINT chk_order_item_conversion_units CHECK(from_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER') AND to_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_batch_unit') THEN ALTER TABLE fulfillment_batches ADD CONSTRAINT chk_batch_unit CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_lot_units') THEN ALTER TABLE inventory_lots ADD CONSTRAINT chk_lot_units CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER') AND (secondary_unit IS NULL OR secondary_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER'))); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_reservation_unit') THEN ALTER TABLE inventory_reservations ADD CONSTRAINT chk_reservation_unit CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_movement_unit') THEN ALTER TABLE inventory_movements ADD CONSTRAINT chk_movement_unit CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_conversion_units') THEN ALTER TABLE inventory_lot_conversions ADD CONSTRAINT chk_conversion_units CHECK(input_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER') AND output_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER') AND waste_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_shipment_item_unit') THEN ALTER TABLE shipment_items ADD CONSTRAINT chk_shipment_item_unit CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_package_units') THEN ALTER TABLE packaging_units ADD CONSTRAINT chk_package_units CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER') AND (weight_unit IS NULL OR weight_unit IN ('TON','KILOGRAM'))); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_vehicle_capacity_unit') THEN ALTER TABLE vehicles ADD CONSTRAINT chk_vehicle_capacity_unit CHECK(capacity_unit IS NULL OR capacity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_container_weight_unit') THEN ALTER TABLE shipment_containers ADD CONSTRAINT chk_container_weight_unit CHECK(weight_unit IS NULL OR weight_unit IN ('TON','KILOGRAM')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_container_item_unit') THEN ALTER TABLE shipment_container_items ADD CONSTRAINT chk_container_item_unit CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_event_item_unit') THEN ALTER TABLE shipment_event_items ADD CONSTRAINT chk_event_item_unit CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_batch_source') THEN ALTER TABLE fulfillment_batches ADD CONSTRAINT chk_batch_source CHECK(source_type IN ('MINE','BLOCK_WAREHOUSE','PRODUCT_WAREHOUSE','FACTORY','SHOWROOM','EXTERNAL_SUPPLIER','CUSTOMER_RETURN')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_batch_status') THEN ALTER TABLE fulfillment_batches ADD CONSTRAINT chk_batch_status CHECK(status IN ('DRAFT','PLANNED','RESERVING_STOCK','STOCK_RESERVED','IN_PRODUCTION','READY_FOR_QC','QC_APPROVED','READY_FOR_PACKAGING','PACKAGED','READY_FOR_SHIPMENT','PARTIALLY_SHIPPED','SHIPPED','PARTIALLY_DELIVERED','DELIVERED','BLOCKED','NEEDS_CORRECTION','SPLIT','MERGED','CANCELLED')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_location_type') THEN ALTER TABLE inventory_locations ADD CONSTRAINT chk_location_type CHECK(location_type IN ('MINE','BLOCK_WAREHOUSE','PRODUCT_WAREHOUSE','FACTORY','WORKSHOP','SHOWROOM','PORT','CUSTOMS_AREA','PROJECT_SITE','CUSTOMER_LOCATION','TEMPORARY_TRANSIT')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_lot_status') THEN ALTER TABLE inventory_lots ADD CONSTRAINT chk_lot_status CHECK(status IN ('AVAILABLE','PARTIALLY_RESERVED','RESERVED','IN_PROCESS','IN_TRANSIT','QUARANTINED','QC_REJECTED','DAMAGED','CONSUMED','SOLD','CANCELLED')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_reservation_status') THEN ALTER TABLE inventory_reservations ADD CONSTRAINT chk_reservation_status CHECK(status IN ('ACTIVE','PARTIALLY_CONSUMED','CONSUMED','RELEASED','EXPIRED','CANCELLED')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_movement_type') THEN ALTER TABLE inventory_movements ADD CONSTRAINT chk_movement_type CHECK(movement_type IN ('RECEIPT','RESERVATION','RESERVATION_RELEASE','ISSUE_TO_PRODUCTION','PRODUCTION_OUTPUT','TRANSFER_OUT','TRANSFER_IN','SHIPMENT_LOADING','SHIPMENT_UNLOADING','DELIVERY','RETURN','DAMAGE','WASTE','ADJUSTMENT','CANCELLATION')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_conversion_type') THEN ALTER TABLE inventory_lot_conversions ADD CONSTRAINT chk_conversion_type CHECK(conversion_type IN ('BLOCK_TO_SLAB','BLOCK_TO_TILE','SLAB_TO_TILE','CUT_TO_SIZE','FINISHING','REPACKAGING','CUSTOM')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_shipment_type') THEN ALTER TABLE shipments ADD CONSTRAINT chk_shipment_type CHECK(shipment_type IN ('DOMESTIC_TRUCK','FACTORY_TRANSFER','WAREHOUSE_TRANSFER','PORT_TRANSFER','EXPORT_CONTAINER','CUSTOMER_PICKUP','PROJECT_DELIVERY')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_shipment_status') THEN ALTER TABLE shipments ADD CONSTRAINT chk_shipment_status CHECK(status IN ('DRAFT','PLANNED','READY_FOR_LOADING','LOADING','LOADED','IN_TRANSIT','ARRIVED','UNLOADING','PARTIALLY_DELIVERED','DELIVERED','HAS_DISCREPANCY','BLOCKED','CANCELLED')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_vehicle_type') THEN ALTER TABLE vehicles ADD CONSTRAINT chk_vehicle_type CHECK(vehicle_type IN ('TRUCK','TRAILER','SEMI_TRAILER','PICKUP','VAN','OTHER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_package_type') THEN ALTER TABLE packaging_units ADD CONSTRAINT chk_package_type CHECK(package_type IN ('PALLET','BUNDLE','CRATE','WOODEN_BOX','METAL_FRAME','SLAB_RACK','CONTAINER_PACKAGE','OTHER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_package_status') THEN ALTER TABLE packaging_units ADD CONSTRAINT chk_package_status CHECK(status IN ('DRAFT','PACKED','QC_APPROVED','ASSIGNED_TO_SHIPMENT','LOADED','DELIVERED','DAMAGED','CANCELLED')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_container_type') THEN ALTER TABLE shipment_containers ADD CONSTRAINT chk_container_type CHECK(container_type IN ('20FT','40FT','40FT_HIGH_CUBE','OPEN_TOP','FLAT_RACK','OTHER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_shipment_event_type') THEN ALTER TABLE shipment_events ADD CONSTRAINT chk_shipment_event_type CHECK(event_type IN ('LOADING','DISPATCH','ARRIVAL','DELIVERY','CANCELLATION')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_transition_type') THEN ALTER TABLE workflow_step_transitions ADD CONSTRAINT chk_transition_type CHECK(transition_type IN ('AUTOMATIC','MANUAL_SELECTION','RESULT_BASED')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_transition_result') THEN ALTER TABLE workflow_step_transitions ADD CONSTRAINT chk_transition_result CHECK(result_code IS NULL OR result_code IN ('APPROVED','REJECTED','HAS_DISCREPANCY','CORRECTION_REQUIRED','CUSTOMER_CANCELLED','PAYMENT_PENDING')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_cost_entity') THEN ALTER TABLE operational_cost_entries ADD CONSTRAINT chk_cost_entity CHECK(entity_type IN ('ORDER','BATCH','SHIPMENT','INVENTORY_MOVEMENT','INSTALLATION')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_cost_type') THEN ALTER TABLE operational_cost_entries ADD CONSTRAINT chk_cost_type CHECK(cost_type IN ('PURCHASE','EXTRACTION','LOADING','TRANSPORT','PROCESSING','QC','PACKAGING','WAREHOUSE','CUSTOMS','PORT','CONTAINER','INSTALLATION','DAMAGE','OTHER')); END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_cost_status') THEN ALTER TABLE operational_cost_entries ADD CONSTRAINT chk_cost_status CHECK(status IN ('DRAFT','APPROVED','VOID')); END IF;
END $$;

INSERT INTO permissions(code,name_fa,description_fa,group_code) VALUES
('batches.view_assigned','مشاهده بچ‌های تخصیص‌یافته','مشاهده بچ‌های مرتبط با نقش یا کاربر','BATCHES'),
('batches.view_all','مشاهده همه بچ‌ها','مشاهده همه بچ‌های سفارش','BATCHES'),('batches.create','ایجاد بچ','تقسیم سفارش به بچ اجرایی','BATCHES'),
('batches.update','ویرایش بچ','ویرایش اطلاعات برنامه‌ریزی بچ','BATCHES'),('batches.split','تقسیم بچ','تقسیم کنترل‌شده بچ','BATCHES'),
('batches.merge','ادغام بچ','ادغام بچ‌های سازگار','BATCHES'),('batches.cancel','لغو بچ','لغو بچ و آزادسازی رزرو','BATCHES'),
('batches.override','بازنویسی بچ','عبور مدیریتی از محدودیت مقدار یا وضعیت','BATCHES'),
('inventory.locations.view','مشاهده مکان‌ها','مشاهده مکان‌های عملیاتی','INVENTORY'),('inventory.locations.manage','مدیریت مکان‌ها','ایجاد و غیرفعال‌سازی مکان','INVENTORY'),
('inventory.lots.view','مشاهده موجودی','مشاهده Lotهای موجودی','INVENTORY'),('inventory.lots.create','ایجاد Lot','ثبت Lot فیزیکی','INVENTORY'),
('inventory.lots.update_metadata','ویرایش مشخصات Lot','ویرایش مشخصات غیرمقداری Lot','INVENTORY'),
('inventory.reservations.view','مشاهده رزرو','مشاهده رزروهای موجودی','INVENTORY'),('inventory.reservations.create','ایجاد رزرو','رزرو موجودی برای بچ','INVENTORY'),
('inventory.reservations.release','آزادسازی رزرو','آزادسازی رزرو مصرف‌نشده','INVENTORY'),('inventory.reservations.override','بازنویسی رزرو','عملیات مدیریتی رزرو','INVENTORY'),
('inventory.movements.view','مشاهده گردش موجودی','مشاهده دفتر تغییرات موجودی','INVENTORY'),('inventory.movements.create','ثبت گردش موجودی','ثبت ورود و خروج موجودی','INVENTORY'),
('inventory.transfers.create','انتقال موجودی','انتقال Lot میان مکان‌ها','INVENTORY'),('inventory.adjustments.create','اصلاح موجودی','Adjustment با دلیل','INVENTORY'),
('inventory.conversions.create','ثبت تبدیل تولید','تبدیل Lot ورودی به خروجی','INVENTORY'),
('packaging.view','مشاهده بسته‌ها','مشاهده واحدهای بسته‌بندی','PACKAGING'),('packaging.create','ایجاد بسته','ثبت بسته‌بندی بچ','PACKAGING'),
('packaging.update','ویرایش بسته','ویرایش بسته‌بندی','PACKAGING'),('packaging.assign_to_shipment','تخصیص بسته به حمل','اتصال بسته به محموله','PACKAGING'),
('packaging.cancel','لغو بسته','لغو بسته‌بندی','PACKAGING'),
('shipments.view_assigned','مشاهده حمل تخصیص‌یافته','مشاهده حمل‌های مرتبط','SHIPMENTS'),('shipments.view_all','مشاهده همه حمل‌ها','مشاهده همه محموله‌ها','SHIPMENTS'),
('shipments.create','ایجاد محموله','ایجاد Shipment','SHIPMENTS'),('shipments.update','ویرایش محموله','ویرایش مشخصات حمل','SHIPMENTS'),
('shipments.plan','برنامه‌ریزی حمل','برنامه‌ریزی محموله','SHIPMENTS'),('shipments.load','بارگیری','ثبت بارگیری','SHIPMENTS'),
('shipments.dispatch','اعزام','ثبت خروج محموله','SHIPMENTS'),('shipments.confirm_arrival','تأیید رسیدن','ثبت رسیدن محموله','SHIPMENTS'),
('shipments.confirm_delivery','تأیید تحویل','ثبت تحویل جزئی یا کامل','SHIPMENTS'),('shipments.cancel','لغو محموله','لغو حمل','SHIPMENTS'),
('shipments.override','بازنویسی حمل','عبور مدیریتی از محدودیت حمل','SHIPMENTS'),
('vehicles.view','مشاهده خودرو','مشاهده ناوگان و خودرو خارجی','TRANSPORT'),('vehicles.manage','مدیریت خودرو','ایجاد و ویرایش خودرو','TRANSPORT'),
('containers.view','مشاهده کانتینر','مشاهده کانتینرها','TRANSPORT'),('containers.manage','مدیریت کانتینر','ثبت کانتینر و پلمپ','TRANSPORT'),
('workflow_transitions.view','مشاهده مسیرها','مشاهده Transitionهای Workflow','WORKFLOWS'),('workflow_transitions.manage','مدیریت مسیرها','طراحی Transition در Builder','WORKFLOWS'),
('workflow_transitions.select','انتخاب مسیر','انتخاب مسیر منتشرشده','WORKFLOWS'),('workflow_transitions.override','تغییر مسیر','بازنویسی مدیریتی مسیر','WORKFLOWS'),
('finance.operational_costs.view','مشاهده هزینه عملیاتی','مشاهده هزینه Batch و Shipment','FINANCE'),
('finance.operational_costs.record','ثبت هزینه عملیاتی','ثبت هزینه مرتبط با عملیات','FINANCE'),
('finance.operational_costs.approve','تأیید هزینه عملیاتی','تأیید هزینه ثبت‌شده','FINANCE'),
('customer_portal.shipments.view_own','مشاهده محموله خود','مشاهده محموله قابل نمایش','CUSTOMER_PORTAL'),
('customer_portal.shipments.confirm_delivery','تأیید تحویل مشتری','تأیید تحویل محموله متعلق به مشتری','CUSTOMER_PORTAL'),
('workflow_start.batch_fulfillment','شروع Workflow بچ','شروع گردش تأمین و تولید بچ','WORKFLOW_START'),
('workflow_start.domestic_shipment','شروع حمل داخلی','شروع گردش حمل داخلی','WORKFLOW_START'),
('workflow_start.export_shipment','شروع حمل صادراتی','شروع گردش حمل صادراتی','WORKFLOW_START')
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa,description_fa=EXCLUDED.description_fa,group_code=EXCLUDED.group_code;

INSERT INTO inventory_locations(code,name_fa,location_type,country_code,is_active)
VALUES('SYSTEM-TRANSIT','در مسیر','TEMPORARY_TRANSIT','IR',TRUE)
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa,location_type=EXCLUDED.location_type,is_active=TRUE;

INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE
 r.code='SUPER_ADMIN'
 OR r.code='ADMIN' AND p.code LIKE ANY(ARRAY['batches.%','inventory.%','packaging.%','shipments.%','vehicles.%','containers.%','workflow_transitions.%','finance.operational_costs.%','workflow_start.batch_fulfillment','workflow_start.domestic_shipment','workflow_start.export_shipment'])
 OR r.code='OPERATOR' AND p.code IN ('batches.view_assigned','batches.create','shipments.view_assigned','workflow_transitions.view','workflow_transitions.select','workflow_start.batch_fulfillment')
 OR r.code='SALES' AND p.code IN ('batches.view_assigned','shipments.view_assigned')
 OR r.code='ACCOUNTANT' AND p.code IN ('batches.view_all','shipments.view_all','inventory.movements.view','finance.operational_costs.view','finance.operational_costs.record','finance.operational_costs.approve')
 OR r.code='SUPPLY' AND (p.code='orders.view_all' OR p.code LIKE ANY(ARRAY['batches.%','inventory.locations.view','inventory.lots.%','inventory.reservations.%','inventory.movements.view','inventory.movements.create','inventory.transfers.create','inventory.conversions.create','packaging.%','shipments.view_assigned','shipments.create','shipments.update','shipments.plan','containers.view','containers.manage','workflow_transitions.view','workflow_transitions.select','workflow_start.batch_fulfillment','workflow_start.domestic_shipment','workflow_start.export_shipment','finance.operational_costs.record']))
 OR r.code='DRIVER' AND p.code IN ('shipments.view_assigned','shipments.load','shipments.dispatch','shipments.confirm_arrival','shipments.confirm_delivery','vehicles.view','containers.view','workflow_transitions.view','workflow_transitions.select','finance.operational_costs.record')
 OR r.code='INSTALLATION_LEAD' AND p.code IN ('batches.view_assigned','shipments.view_assigned','shipments.confirm_delivery','workflow_transitions.view','workflow_transitions.select')
 OR r.code='CUSTOMER' AND p.code IN ('customer_portal.shipments.view_own','customer_portal.shipments.confirm_delivery')
ON CONFLICT DO NOTHING;

INSERT INTO workflow_templates(template_group_code,version_number,code,name_fa,description_fa,icon_key,start_permission_code,is_active,status,sort_order,scope_type,published_at,max_iterations)
VALUES
('batch_fulfillment',1,'batch_fulfillment_v1','تأمین و تولید بچ','رزرو یا تأمین، تولید، کنترل کیفیت و بسته‌بندی','factory','workflow_start.batch_fulfillment',TRUE,'PUBLISHED',100,'BATCH',NOW(),20),
('domestic_shipment',1,'domestic_shipment_v1','حمل داخلی','بارگیری، اعزام، رسیدن و تحویل داخلی','project','workflow_start.domestic_shipment',TRUE,'PUBLISHED',110,'SHIPMENT',NOW(),10),
('export_shipment',1,'export_shipment_v1','حمل صادراتی','کانتینر، بارگیری، پلمپ، بندر، گمرک و تحویل','export','workflow_start.export_shipment',TRUE,'PUBLISHED',120,'SHIPMENT',NOW(),15)
ON CONFLICT(code) DO NOTHING;

INSERT INTO workflow_template_steps(workflow_template_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,responsible_role_code,responsible_role_id,required_permission_code,customer_visible,is_first_step,is_entry,is_active,default_duration_hours,starts_automatically,domain_event_code)
SELECT wt.id,x.step_code,x.title,x.description,x.customer_title,x.customer_description,x.seq,x.role_code,r.id,x.permission,x.visible,x.seq=1,x.seq=1,TRUE,x.hours,x.auto_start,x.domain_event
FROM workflow_templates wt
JOIN (VALUES
 ('batch_fulfillment_v1','PLAN_BATCH','برنامه‌ریزی بچ','منبع و مسیر اجرای بچ را تعیین کنید','برنامه‌ریزی سفارش','مسیر آماده‌سازی مشخص می‌شود',1,'SUPPLY','batches.update',TRUE,8,FALSE,NULL),
 ('batch_fulfillment_v1','RESERVE_INVENTORY','رزرو موجودی','Lotهای موجود را برای بچ رزرو کنید','تأمین از موجودی','سنگ موردنیاز رزرو می‌شود',2,'SUPPLY','inventory.reservations.create',TRUE,12,FALSE,'BATCH_STOCK_RESERVED'),
 ('batch_fulfillment_v1','RECEIVE_INVENTORY','تأمین و دریافت','سنگ تأمین‌شده را در موجودی دریافت کنید','تأمین سنگ','سنگ موردنیاز در حال تأمین است',3,'SUPPLY','inventory.movements.create',TRUE,72,FALSE,'BATCH_STOCK_RESERVED'),
 ('batch_fulfillment_v1','ISSUE_TO_PRODUCTION','تحویل به تولید','ورودی تولید را از رزرو مصرف کنید','شروع تولید','سنگ وارد فرایند تولید شد',4,'SUPPLY','inventory.conversions.create',TRUE,12,FALSE,NULL),
 ('batch_fulfillment_v1','REGISTER_PRODUCTION_OUTPUT','ثبت خروجی تولید','Lot خروجی و ضایعات را ثبت کنید','فراوری محصول','محصول در حال آماده‌سازی است',5,'SUPPLY','inventory.conversions.create',TRUE,120,FALSE,'BATCH_READY_FOR_QC'),
 ('batch_fulfillment_v1','QC_BATCH','کنترل کیفیت بچ','نتیجه کنترل کیفیت را ثبت کنید','کنترل کیفیت','محصول در حال کنترل کیفیت است',6,'SUPPLY','operations.quality_control.execute',TRUE,24,FALSE,NULL),
 ('batch_fulfillment_v1','PACKAGE_BATCH','بسته‌بندی بچ','واحدهای بسته‌بندی را ثبت کنید','بسته‌بندی','محصول در حال بسته‌بندی است',7,'SUPPLY','packaging.create',TRUE,36,FALSE,'BATCH_PACKAGED'),
 ('batch_fulfillment_v1','READY_FOR_SHIPMENT','آماده ارسال','آمادگی نهایی بچ را تأیید کنید','آماده ارسال','سفارش آماده ارسال است',8,'SUPPLY','shipments.create',TRUE,8,FALSE,NULL),
 ('domestic_shipment_v1','PLAN_SHIPMENT','برنامه‌ریزی محموله','خودرو، راننده و زمان‌ها را مشخص کنید','برنامه‌ریزی ارسال','ارسال سفارش برنامه‌ریزی می‌شود',1,'SUPPLY','shipments.plan',TRUE,8,FALSE,NULL),
 ('domestic_shipment_v1','LOAD_SHIPMENT','بارگیری محموله','اقلام و مقادیر بارگیری را ثبت کنید','بارگیری','محموله در حال بارگیری است',2,'SUPPLY','shipments.load',TRUE,12,FALSE,'SHIPMENT_LOADED'),
 ('domestic_shipment_v1','DISPATCH_SHIPMENT','اعزام محموله','خروج محموله را ثبت کنید','ارسال شد','محموله در مسیر است',3,'DRIVER','shipments.dispatch',TRUE,4,FALSE,'SHIPMENT_DISPATCHED'),
 ('domestic_shipment_v1','CONFIRM_ARRIVAL','تأیید رسیدن','رسیدن به مقصد را ثبت کنید','رسیدن به مقصد','محموله به مقصد رسیده است',4,'DRIVER','shipments.confirm_arrival',TRUE,48,FALSE,'SHIPMENT_ARRIVED'),
 ('domestic_shipment_v1','CONFIRM_DELIVERY','ثبت تحویل','مقدار و مدرک تحویل را ثبت کنید','تحویل سفارش','محموله تحویل شده است',5,'DRIVER','shipments.confirm_delivery',TRUE,12,FALSE,'SHIPMENT_DELIVERED'),
 ('export_shipment_v1','PLAN_EXPORT_SHIPMENT','برنامه‌ریزی حمل صادراتی','مسیر، زمان و شرکت حمل را مشخص کنید','برنامه‌ریزی صادرات','حمل صادراتی برنامه‌ریزی می‌شود',1,'SUPPLY','shipments.plan',TRUE,12,FALSE,NULL),
 ('export_shipment_v1','PREPARE_CONTAINERS','تخصیص کانتینر','کانتینرها و اقلام را ثبت کنید','آماده‌سازی کانتینر','کانتینر سفارش آماده می‌شود',2,'SUPPLY','containers.manage',TRUE,24,FALSE,NULL),
 ('export_shipment_v1','LOAD_EXPORT_SHIPMENT','بارگیری صادراتی','بارگیری کانتینرها را ثبت کنید','بارگیری','محموله صادراتی بارگیری می‌شود',3,'SUPPLY','shipments.load',TRUE,24,FALSE,'SHIPMENT_LOADED'),
 ('export_shipment_v1','VERIFY_SEAL','تأیید پلمپ','شماره Seal و وزن نهایی را تأیید کنید','تأیید کانتینر','کانتینر نهایی می‌شود',4,'SUPPLY','containers.manage',TRUE,8,FALSE,NULL),
 ('export_shipment_v1','TRANSFER_TO_PORT','انتقال به بندر','خروج به سمت بندر را ثبت کنید','ارسال به بندر','محموله به سمت بندر حرکت کرده است',5,'DRIVER','shipments.dispatch',TRUE,48,FALSE,'SHIPMENT_DISPATCHED'),
 ('export_shipment_v1','PORT_CUSTOMS','بندر و گمرک','عملیات بندری و گمرکی را ثبت کنید','امور صادرات','تشریفات صادرات در حال انجام است',6,'SUPPLY','operations.export.execute',TRUE,120,FALSE,NULL),
 ('export_shipment_v1','CONFIRM_EXPORT_DELIVERY','تأیید تحویل صادراتی','رسیدن و تحویل نهایی را ثبت کنید','تحویل محموله','محموله تحویل شده است',7,'DRIVER','shipments.confirm_delivery',TRUE,168,FALSE,'SHIPMENT_DELIVERED')
) AS x(template_code,step_code,title,description,customer_title,customer_description,seq,role_code,permission,visible,hours,auto_start,domain_event)
ON wt.code=x.template_code JOIN roles r ON r.code=x.role_code
ON CONFLICT(workflow_template_id,step_code) DO NOTHING;

INSERT INTO workflow_step_transitions(workflow_template_id,source_step_id,target_step_id,transition_code,label_fa,transition_type,is_default,requires_permission_code,requires_reason,sort_order)
SELECT wt.id,s.id,t.id,x.code,x.label,x.kind,x.is_default,x.permission,x.reason,x.sort_order
FROM workflow_templates wt
JOIN (VALUES
 ('batch_fulfillment_v1','PLAN_BATCH','RESERVE_INVENTORY','USE_EXISTING_STOCK','رزرو از موجودی','MANUAL_SELECTION',TRUE,'workflow_transitions.select',FALSE,1),
 ('batch_fulfillment_v1','PLAN_BATCH','RECEIVE_INVENTORY','PROCURE_NEW_STOCK','تأمین موجودی جدید','MANUAL_SELECTION',FALSE,'workflow_transitions.select',TRUE,2),
 ('batch_fulfillment_v1','RESERVE_INVENTORY','ISSUE_TO_PRODUCTION','RESERVED_TO_PRODUCTION','ارسال به تولید','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('batch_fulfillment_v1','RECEIVE_INVENTORY','ISSUE_TO_PRODUCTION','RECEIVED_TO_PRODUCTION','ارسال به تولید','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('batch_fulfillment_v1','ISSUE_TO_PRODUCTION','REGISTER_PRODUCTION_OUTPUT','START_PROCESSING','ثبت خروجی تولید','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('batch_fulfillment_v1','REGISTER_PRODUCTION_OUTPUT','QC_BATCH','OUTPUT_TO_QC','کنترل کیفیت','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('batch_fulfillment_v1','QC_BATCH','PACKAGE_BATCH','QC_TO_PACKAGING','بسته‌بندی','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('batch_fulfillment_v1','PACKAGE_BATCH','READY_FOR_SHIPMENT','PACKED_TO_READY','آماده ارسال','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('domestic_shipment_v1','PLAN_SHIPMENT','LOAD_SHIPMENT','PLAN_TO_LOAD','بارگیری','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('domestic_shipment_v1','LOAD_SHIPMENT','DISPATCH_SHIPMENT','LOAD_TO_DISPATCH','اعزام','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('domestic_shipment_v1','DISPATCH_SHIPMENT','CONFIRM_ARRIVAL','DISPATCH_TO_ARRIVAL','ثبت رسیدن','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('domestic_shipment_v1','CONFIRM_ARRIVAL','CONFIRM_DELIVERY','ARRIVAL_TO_DELIVERY','ثبت تحویل','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('export_shipment_v1','PLAN_EXPORT_SHIPMENT','PREPARE_CONTAINERS','PLAN_TO_CONTAINER','آماده‌سازی کانتینر','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('export_shipment_v1','PREPARE_CONTAINERS','LOAD_EXPORT_SHIPMENT','CONTAINER_TO_LOAD','بارگیری','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('export_shipment_v1','LOAD_EXPORT_SHIPMENT','VERIFY_SEAL','LOAD_TO_SEAL','تأیید پلمپ','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('export_shipment_v1','VERIFY_SEAL','TRANSFER_TO_PORT','SEAL_TO_PORT','ارسال به بندر','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('export_shipment_v1','TRANSFER_TO_PORT','PORT_CUSTOMS','PORT_TO_CUSTOMS','گمرک','AUTOMATIC',TRUE,NULL,FALSE,1),
 ('export_shipment_v1','PORT_CUSTOMS','CONFIRM_EXPORT_DELIVERY','CUSTOMS_TO_DELIVERY','تحویل','AUTOMATIC',TRUE,NULL,FALSE,1)
) AS x(template_code,source_code,target_code,code,label,kind,is_default,permission,reason,sort_order)
ON wt.code=x.template_code JOIN workflow_template_steps s ON s.workflow_template_id=wt.id AND s.step_code=x.source_code
JOIN workflow_template_steps t ON t.workflow_template_id=wt.id AND t.step_code=x.target_code
ON CONFLICT(workflow_template_id,source_step_id,transition_code) DO NOTHING;
