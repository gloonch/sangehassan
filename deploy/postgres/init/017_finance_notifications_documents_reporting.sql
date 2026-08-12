-- Finance, notifications, documents and reporting for the operations dashboard.
-- This migration is intentionally idempotent and does not modify previous migrations.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS currencies (
  code CHAR(3) PRIMARY KEY,
  name_fa TEXT NOT NULL,
  symbol TEXT NOT NULL,
  decimal_places SMALLINT NOT NULL DEFAULT 2 CHECK (decimal_places BETWEEN 0 AND 4),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO currencies(code,name_fa,symbol,decimal_places) VALUES
('IRR','ریال ایران','ریال',0),('USD','دلار آمریکا','$',2),('EUR','یورو','€',2),
('AED','درهم امارات','د.إ',2),('OMR','ریال عمان','ر.ع.',3)
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa,symbol=EXCLUDED.symbol,decimal_places=EXCLUDED.decimal_places,is_active=TRUE;

CREATE TABLE IF NOT EXISTS exchange_rates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  base_currency CHAR(3) NOT NULL REFERENCES currencies(code),
  quote_currency CHAR(3) NOT NULL REFERENCES currencies(code),
  rate NUMERIC(18,8) NOT NULL CHECK(rate>0),
  effective_at TIMESTAMPTZ NOT NULL,
  source TEXT NOT NULL DEFAULT 'MANUAL' CHECK(source='MANUAL'),
  notes TEXT,
  created_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(base_currency,quote_currency,effective_at),
  CHECK(base_currency<>quote_currency)
);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_lookup ON exchange_rates(base_currency,quote_currency,effective_at DESC);

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS sales_owner_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS estimated_delivery_at TIMESTAMPTZ;

ALTER TABLE order_items
  ADD COLUMN IF NOT EXISTS unit_price NUMERIC(18,4) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS line_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS currency CHAR(3) NOT NULL DEFAULT 'IRR';

DO $$ BEGIN
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_order_item_commercial_amounts') THEN
    ALTER TABLE order_items ADD CONSTRAINT chk_order_item_commercial_amounts
      CHECK(unit_price>=0 AND discount_amount>=0 AND line_amount>=0);
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS order_commercial_terms (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  order_id UUID NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
  terms_type TEXT NOT NULL DEFAULT 'CUSTOM',
  currency CHAR(3) NOT NULL DEFAULT 'IRR' REFERENCES currencies(code),
  subtotal NUMERIC(18,4) NOT NULL DEFAULT 0,
  discount_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  tax_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  additional_charge_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  final_customer_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  deposit_percentage NUMERIC(7,4),
  deposit_amount NUMERIC(18,4),
  payment_terms_text TEXT,
  delivery_terms_text TEXT,
  version_number INT NOT NULL DEFAULT 1 CHECK(version_number>0),
  last_change_reason TEXT,
  updated_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(terms_type IN ('FULL_PREPAYMENT','DEPOSIT_AND_BALANCE','INSTALLMENTS','PAY_ON_DELIVERY','CREDIT','CUSTOM')),
  CHECK(subtotal>=0 AND discount_amount>=0 AND tax_amount>=0 AND additional_charge_amount>=0 AND final_customer_amount>=0),
  CHECK(deposit_percentage IS NULL OR (deposit_percentage>=0 AND deposit_percentage<=100)),
  CHECK(deposit_amount IS NULL OR deposit_amount>=0)
);

INSERT INTO order_commercial_terms(order_id,currency,subtotal,discount_amount,final_customer_amount,terms_type)
SELECT o.id,COALESCE(p.currency,'IRR'),COALESCE(p.subtotal,0),COALESCE(p.discount_amount,0),COALESCE(p.total_amount,0),'CUSTOM'
FROM orders o
LEFT JOIN LATERAL (
  SELECT currency,subtotal,discount_amount,total_amount FROM proformas
  WHERE order_id=o.id AND status<>'CANCELLED' ORDER BY COALESCE(issued_at,created_at) DESC LIMIT 1
) p ON TRUE
ON CONFLICT(order_id) DO NOTHING;

CREATE OR REPLACE FUNCTION ensure_order_commercial_terms() RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO order_commercial_terms(order_id,currency,subtotal,discount_amount,tax_amount,additional_charge_amount,final_customer_amount,terms_type)
  VALUES(NEW.id,'IRR',0,0,0,0,0,'CUSTOM') ON CONFLICT(order_id) DO NOTHING;
  RETURN NEW;
END $$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_ensure_order_commercial_terms ON orders;
CREATE TRIGGER trg_ensure_order_commercial_terms AFTER INSERT ON orders FOR EACH ROW EXECUTE FUNCTION ensure_order_commercial_terms();

ALTER TABLE proformas
  ALTER COLUMN subtotal TYPE NUMERIC(18,4),
  ALTER COLUMN discount_amount TYPE NUMERIC(18,4),
  ALTER COLUMN total_amount TYPE NUMERIC(18,4),
  ADD COLUMN IF NOT EXISTS tax_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS additional_charge_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS commercial_terms_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS version_number INT NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS supersedes_proforma_id UUID REFERENCES proformas(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS proforma_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  proforma_id UUID NOT NULL REFERENCES proformas(id) ON DELETE CASCADE,
  order_item_id UUID REFERENCES order_items(id) ON DELETE SET NULL,
  line_number INT NOT NULL,
  description_fa TEXT NOT NULL,
  stone_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
  quantity NUMERIC(18,4) NOT NULL,
  quantity_unit TEXT NOT NULL,
  unit_price NUMERIC(18,4) NOT NULL,
  discount_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  line_amount NUMERIC(18,4) NOT NULL,
  currency CHAR(3) NOT NULL REFERENCES currencies(code),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(proforma_id,line_number),
  CHECK(quantity>0 AND unit_price>=0 AND discount_amount>=0 AND line_amount>=0)
);

CREATE TABLE IF NOT EXISTS order_payment_schedule (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  sequence_number INT NOT NULL,
  title_fa TEXT NOT NULL,
  payment_type TEXT NOT NULL DEFAULT 'CUSTOM',
  due_at TIMESTAMPTZ,
  amount NUMERIC(18,4) NOT NULL CHECK(amount>0),
  percentage_of_order NUMERIC(7,4),
  currency CHAR(3) NOT NULL REFERENCES currencies(code),
  paid_amount NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK(paid_amount>=0),
  status TEXT NOT NULL DEFAULT 'UPCOMING',
  trigger_type TEXT NOT NULL DEFAULT 'DATE',
  trigger_step_code TEXT,
  customer_visible BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(order_id,sequence_number),
  CHECK(payment_type IN ('DEPOSIT','PROGRESS_PAYMENT','BEFORE_LOADING','BEFORE_SHIPMENT','ON_DELIVERY','POST_DELIVERY','INSTALLMENT','FINAL_PAYMENT','CUSTOM')),
  CHECK(percentage_of_order IS NULL OR (percentage_of_order>0 AND percentage_of_order<=100)),
  CHECK(status IN ('UPCOMING','DUE','PARTIALLY_PAID','PAID','OVERDUE','WAIVED','CANCELLED')),
  CHECK(trigger_type IN ('DATE','ORDER_CONFIRMATION','STEP_OPEN','STEP_COMPLETE','LOADING','DISPATCH','DELIVERY','MANUAL'))
);
CREATE INDEX IF NOT EXISTS idx_payment_schedule_due ON order_payment_schedule(status,due_at);

CREATE TABLE IF NOT EXISTS customer_payments (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  payment_number TEXT NOT NULL UNIQUE,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
  customer_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  amount NUMERIC(18,4) NOT NULL CHECK(amount>0),
  currency CHAR(3) NOT NULL REFERENCES currencies(code),
  payment_method TEXT NOT NULL,
  status TEXT NOT NULL,
  reference_number TEXT,
  bank_name TEXT,
  receipt_file_id UUID REFERENCES workflow_files(id) ON DELETE SET NULL,
  paid_at TIMESTAMPTZ NOT NULL,
  notes TEXT,
  reported_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  confirmed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  confirmed_at TIMESTAMPTZ,
  rejected_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  rejected_at TIMESTAMPTZ,
  rejection_reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(payment_method IN ('BANK_TRANSFER','CARD','CARD_TO_CARD','CASH','CHEQUE','SWIFT','WIRE_TRANSFER','POS','OTHER')),
  CHECK(status IN ('REPORTED','PENDING_CONFIRMATION','CONFIRMED','REJECTED','PARTIALLY_REFUNDED','REFUNDED','CANCELLED'))
);
CREATE INDEX IF NOT EXISTS idx_customer_payments_order ON customer_payments(order_id,status,paid_at DESC);

CREATE TABLE IF NOT EXISTS payment_allocations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  payment_id UUID NOT NULL REFERENCES customer_payments(id) ON DELETE RESTRICT,
  schedule_id UUID NOT NULL REFERENCES order_payment_schedule(id) ON DELETE RESTRICT,
  amount NUMERIC(18,4) NOT NULL CHECK(amount>0),
  allocated_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(payment_id,schedule_id)
);

CREATE TABLE IF NOT EXISTS payment_refunds (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  refund_number TEXT NOT NULL UNIQUE,
  payment_id UUID NOT NULL REFERENCES customer_payments(id) ON DELETE RESTRICT,
  amount NUMERIC(18,4) NOT NULL CHECK(amount>0),
  currency CHAR(3) NOT NULL REFERENCES currencies(code),
  reason TEXT NOT NULL,
  reference_number TEXT,
  refunded_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  refunded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_refund_allocations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  refund_id UUID NOT NULL REFERENCES payment_refunds(id) ON DELETE CASCADE,
  payment_allocation_id UUID NOT NULL REFERENCES payment_allocations(id) ON DELETE RESTRICT,
  amount NUMERIC(18,4) NOT NULL CHECK(amount>0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(refund_id,payment_allocation_id)
);

DO $$ BEGIN
  ALTER TABLE operational_cost_entries DROP CONSTRAINT IF EXISTS chk_cost_status;
END $$;

ALTER TABLE operational_cost_entries
  ALTER COLUMN amount TYPE NUMERIC(18,4),
  ALTER COLUMN status SET DEFAULT 'REPORTED',
  ADD COLUMN IF NOT EXISTS order_id UUID REFERENCES orders(id) ON DELETE RESTRICT,
  ADD COLUMN IF NOT EXISTS batch_id UUID REFERENCES fulfillment_batches(id) ON DELETE RESTRICT,
  ADD COLUMN IF NOT EXISTS shipment_id UUID REFERENCES shipments(id) ON DELETE RESTRICT,
  ADD COLUMN IF NOT EXISTS vendor_name TEXT,
  ADD COLUMN IF NOT EXISTS vendor_reference TEXT,
  ADD COLUMN IF NOT EXISTS invoice_number TEXT,
  ADD COLUMN IF NOT EXISTS invoice_file_id UUID REFERENCES workflow_files(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS submitted_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS rejected_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS rejection_reason TEXT,
  ADD COLUMN IF NOT EXISTS paid_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS paid_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS payment_reference TEXT,
  ADD COLUMN IF NOT EXISTS cancelled_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS cancellation_reason TEXT;
UPDATE operational_cost_entries SET status='REPORTED' WHERE status='DRAFT';
UPDATE operational_cost_entries SET status='CANCELLED',cancelled_at=COALESCE(cancelled_at,voided_at),cancellation_reason=COALESCE(cancellation_reason,void_reason) WHERE status='VOID';
UPDATE operational_cost_entries c SET order_id=CASE c.entity_type WHEN 'ORDER' THEN c.entity_id WHEN 'BATCH' THEN (SELECT b.order_id FROM fulfillment_batches b WHERE b.id=c.entity_id) WHEN 'SHIPMENT' THEN (SELECT s.order_id FROM shipments s WHERE s.id=c.entity_id) WHEN 'INVENTORY_MOVEMENT' THEN (SELECT m.order_id FROM inventory_movements m WHERE m.id=c.entity_id) ELSE NULL END WHERE c.order_id IS NULL;
DO $$ BEGIN
  ALTER TABLE operational_cost_entries DROP CONSTRAINT IF EXISTS chk_cost_type;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_operational_cost_status') THEN
    ALTER TABLE operational_cost_entries ADD CONSTRAINT chk_operational_cost_status CHECK(status IN ('ESTIMATED','REPORTED','PENDING_APPROVAL','APPROVED','REJECTED','PAID','CANCELLED'));
  END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_operational_cost_type') THEN
    ALTER TABLE operational_cost_entries ADD CONSTRAINT chk_operational_cost_type CHECK(cost_type IN ('PURCHASE','STONE_PURCHASE','EXTRACTION','MINE_LOADING','LOADING','TRANSPORT','FACTORY_RECEIVING','CUTTING','PROCESSING','FINISHING','QC','QUALITY_CONTROL','PACKAGING','WAREHOUSE','CUSTOMS','PORT','CONTAINER','INSURANCE','INSTALLATION','LABOR','DAMAGE','REWORK','COMMISSION','OTHER'));
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS workflow_payment_blocks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  workflow_step_instance_id UUID REFERENCES workflow_step_instances(id) ON DELETE CASCADE,
  schedule_id UUID NOT NULL REFERENCES order_payment_schedule(id) ON DELETE CASCADE,
  trigger_type TEXT NOT NULL,
  required_amount NUMERIC(18,4) NOT NULL,
  currency CHAR(3) NOT NULL REFERENCES currencies(code),
  previous_step_status TEXT,
  status TEXT NOT NULL DEFAULT 'OPEN',
  action_item_id UUID REFERENCES action_items(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMPTZ,
  UNIQUE(workflow_instance_id,workflow_step_instance_id,schedule_id,trigger_type),
  CHECK(status IN ('OPEN','RESOLVED','CANCELLED'))
);

CREATE TABLE IF NOT EXISTS notification_templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  event_type TEXT NOT NULL,
  channel TEXT NOT NULL,
  locale TEXT NOT NULL DEFAULT 'fa',
  audience_type TEXT NOT NULL DEFAULT 'ASSIGNED_USER',
  title_template TEXT NOT NULL,
  body_template TEXT NOT NULL,
  allowed_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  version_number INT NOT NULL DEFAULT 1 CHECK(version_number>0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(event_type,channel,locale),
  CHECK(channel IN ('IN_APP','SMS')),
  CHECK(audience_type IN ('CUSTOMER','SALES','ACCOUNTANT','ADMIN','ASSIGNED_USER','ASSIGNED_ROLE'))
);

CREATE TABLE IF NOT EXISTS notification_preferences (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  in_app_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  sms_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY(user_id,event_type)
);

ALTER TABLE notifications
  ADD COLUMN IF NOT EXISTS event_type TEXT,
  ADD COLUMN IF NOT EXISTS event_key TEXT,
  ADD COLUMN IF NOT EXISTS title TEXT,
  ADD COLUMN IF NOT EXISTS body TEXT,
  ADD COLUMN IF NOT EXISTS entity_type TEXT,
  ADD COLUMN IF NOT EXISTS entity_id UUID,
  ADD COLUMN IF NOT EXISTS deep_link TEXT,
  ADD COLUMN IF NOT EXISTS data_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS priority TEXT NOT NULL DEFAULT 'NORMAL',
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'UNREAD';
UPDATE notifications SET event_type=COALESCE(event_type,type),event_key=COALESCE(event_key,'legacy:'||id::text),title=COALESCE(title,type),body=COALESCE(body,payload::text),data_json=CASE WHEN data_json='{}'::jsonb THEN payload ELSE data_json END,status=CASE WHEN read_at IS NULL THEN 'UNREAD' ELSE 'READ' END;
ALTER TABLE notifications
  ALTER COLUMN event_type SET NOT NULL,
  ALTER COLUMN event_key SET NOT NULL,
  ALTER COLUMN title SET NOT NULL,
  ALTER COLUMN body SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_notifications_event_key ON notifications(event_key);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id,status,created_at DESC);
DO $$ BEGIN
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_notifications_status') THEN
    ALTER TABLE notifications ADD CONSTRAINT chk_notifications_status CHECK(status IN ('UNREAD','READ','ARCHIVED'));
  END IF;
  IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname='chk_notifications_priority') THEN
    ALTER TABLE notifications ADD CONSTRAINT chk_notifications_priority CHECK(priority IN ('NORMAL','HIGH','URGENT'));
  END IF;
END $$;
CREATE OR REPLACE FUNCTION normalize_notification_compatibility() RETURNS TRIGGER AS $$
BEGIN
  NEW.event_type := COALESCE(NEW.event_type,NEW.type,'LEGACY');
  NEW.event_key := COALESCE(NEW.event_key,'legacy:'||COALESCE(NEW.id::text,uuid_generate_v4()::text));
  NEW.title := COALESCE(NEW.title,NEW.type,'اعلان');
  NEW.body := COALESCE(NEW.body,NEW.payload::text,'');
  NEW.data_json := CASE WHEN NEW.data_json IS NULL OR NEW.data_json='{}'::jsonb THEN COALESCE(NEW.payload,'{}'::jsonb) ELSE NEW.data_json END;
  IF NEW.read_at IS NOT NULL AND (TG_OP='INSERT' OR OLD.read_at IS NULL) THEN NEW.status := 'READ'; END IF;
  RETURN NEW;
END $$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_normalize_notification_compatibility ON notifications;
CREATE TRIGGER trg_normalize_notification_compatibility BEFORE INSERT OR UPDATE ON notifications FOR EACH ROW EXECUTE FUNCTION normalize_notification_compatibility();

CREATE TABLE IF NOT EXISTS notification_outbox (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  notification_id BIGINT REFERENCES notifications(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  channel TEXT NOT NULL,
  event_key TEXT NOT NULL UNIQUE,
  recipient TEXT NOT NULL,
  message_body TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING',
  attempt_count INT NOT NULL DEFAULT 0,
  next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  provider_message_id TEXT,
  last_error TEXT,
  locked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  sent_at TIMESTAMPTZ,
  CHECK(channel='SMS'),
  CHECK(status IN ('PENDING','PROCESSING','SENT','RETRY','FAILED','CANCELLED'))
);
CREATE INDEX IF NOT EXISTS idx_notification_outbox_ready ON notification_outbox(status,next_attempt_at) WHERE status IN ('PENDING','RETRY');

ALTER TABLE workflow_files ALTER COLUMN workflow_instance_id DROP NOT NULL;

CREATE TABLE IF NOT EXISTS document_templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  code TEXT NOT NULL UNIQUE,
  name_fa TEXT NOT NULL,
  document_type TEXT NOT NULL,
  locale TEXT NOT NULL DEFAULT 'fa',
  template_json JSONB NOT NULL,
  allowed_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
  version_number INT NOT NULL DEFAULT 1,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS documents (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  document_number TEXT NOT NULL UNIQUE,
  document_type TEXT NOT NULL,
  scope_type TEXT NOT NULL,
  scope_id UUID NOT NULL,
  order_id UUID REFERENCES orders(id) ON DELETE RESTRICT,
  customer_user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
  document_template_id UUID REFERENCES document_templates(id) ON DELETE SET NULL,
  version_number INT NOT NULL DEFAULT 1,
  supersedes_document_id UUID REFERENCES documents(id) ON DELETE SET NULL,
  status TEXT NOT NULL DEFAULT 'DRAFT',
  template_snapshot_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  snapshot_json JSONB NOT NULL,
  workflow_file_id UUID NOT NULL UNIQUE REFERENCES workflow_files(id) ON DELETE RESTRICT,
  customer_visible BOOLEAN NOT NULL DEFAULT FALSE,
  generated_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  issued_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  issued_at TIMESTAMPTZ,
  cancelled_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  cancelled_at TIMESTAMPTZ,
  cancellation_reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(document_type,scope_type,scope_id,version_number),
  CHECK(document_type IN ('PROFORMA','SALES_INVOICE','PAYMENT_RECEIPT','ORDER_SUMMARY','PACKING_LIST','DELIVERY_NOTE','TRANSPORT_RECEIPT','INSTALLATION_REPORT','QUALITY_REPORT','CUSTOMER_ACCEPTANCE','COMMERCIAL_INVOICE','EXPORT_PACKING_LIST','CUSTOMS_DOCUMENT','CERTIFICATE_OF_ORIGIN','CUSTOMS_DECLARATION','BILL_OF_LADING','CERTIFICATE','OTHER')),
  CHECK(scope_type IN ('ORDER','PAYMENT','SHIPMENT','BATCH','WORKFLOW')),
  CHECK(status IN ('DRAFT','ISSUED','SUPERSEDED','CANCELLED'))
);
ALTER TABLE documents ADD COLUMN IF NOT EXISTS template_snapshot_json JSONB NOT NULL DEFAULT '{}'::jsonb;
CREATE INDEX IF NOT EXISTS idx_documents_order ON documents(order_id,status,created_at DESC);

CREATE TABLE IF NOT EXISTS document_export_details (
  document_id UUID PRIMARY KEY REFERENCES documents(id) ON DELETE CASCADE,
  exporter_name TEXT,
  consignee_name TEXT,
  destination_country TEXT,
  incoterm TEXT,
  hs_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
  customs_reference TEXT,
  container_numbers JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE IF NOT EXISTS workflow_template_document_requirements (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  workflow_template_id BIGINT NOT NULL REFERENCES workflow_templates(id) ON DELETE CASCADE,
  workflow_template_step_id BIGINT REFERENCES workflow_template_steps(id) ON DELETE CASCADE,
  document_type TEXT NOT NULL,
  title_fa TEXT NOT NULL,
  is_required BOOLEAN NOT NULL DEFAULT TRUE,
  is_blocking BOOLEAN NOT NULL DEFAULT TRUE,
  customer_visible BOOLEAN NOT NULL DEFAULT FALSE,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workflow_template_id,workflow_template_step_id,document_type)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_template_document_requirement_scope ON workflow_template_document_requirements(workflow_template_id,COALESCE(workflow_template_step_id,0),document_type);

CREATE TABLE IF NOT EXISTS workflow_instance_document_requirements (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  workflow_step_instance_id UUID REFERENCES workflow_step_instances(id) ON DELETE CASCADE,
  source_requirement_id UUID REFERENCES workflow_template_document_requirements(id) ON DELETE SET NULL,
  document_type TEXT NOT NULL,
  title_fa TEXT NOT NULL,
  is_required BOOLEAN NOT NULL,
  is_blocking BOOLEAN NOT NULL,
  customer_visible BOOLEAN NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING',
  satisfied_document_id UUID REFERENCES documents(id) ON DELETE SET NULL,
  satisfied_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(status IN ('PENDING','SATISFIED','WAIVED','CANCELLED'))
);
CREATE INDEX IF NOT EXISTS idx_document_requirements_pending ON workflow_instance_document_requirements(workflow_instance_id,status,is_blocking);

CREATE TABLE IF NOT EXISTS customer_contact_logs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  customer_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
  contact_type TEXT NOT NULL,
  direction TEXT NOT NULL,
  reason_code TEXT,
  result_code TEXT,
  subject TEXT,
  summary TEXT NOT NULL,
  follow_up_at TIMESTAMPTZ,
  contacted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(contact_type IN ('PHONE','SMS','EMAIL','WHATSAPP','IN_PERSON','OTHER')),
  CHECK(direction IN ('INBOUND','OUTBOUND')),
  CHECK(result_code IS NULL OR result_code IN ('ANSWERED','NO_ANSWER','CUSTOMER_CONFIRMED','PAYMENT_PROMISED','FOLLOW_UP_REQUIRED','ISSUE_REPORTED','OTHER'))
);
CREATE INDEX IF NOT EXISTS idx_contact_logs_customer ON customer_contact_logs(customer_user_id,contacted_at DESC);

CREATE TABLE IF NOT EXISTS scheduled_job_runs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_code TEXT NOT NULL,
  scheduled_bucket TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  affected_count INT NOT NULL DEFAULT 0,
  error_text TEXT,
  UNIQUE(job_code,scheduled_bucket),
  CHECK(status IN ('RUNNING','COMPLETED','FAILED','SKIPPED'))
);

CREATE TABLE IF NOT EXISTS order_financial_summaries (
  order_id UUID PRIMARY KEY REFERENCES orders(id) ON DELETE CASCADE,
  currency CHAR(3) NOT NULL REFERENCES currencies(code),
  revenue_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  confirmed_payment_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  refunded_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  approved_cost_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  outstanding_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE OR REPLACE FUNCTION initialize_order_financial_summary() RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO order_financial_summaries(order_id,currency,revenue_amount,outstanding_amount)
  SELECT NEW.order_id,NEW.currency,CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED') THEN NEW.final_customer_amount ELSE 0 END,CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED') THEN NEW.final_customer_amount ELSE 0 END FROM orders o WHERE o.id=NEW.order_id
  ON CONFLICT(order_id) DO UPDATE SET currency=EXCLUDED.currency,revenue_amount=EXCLUDED.revenue_amount,outstanding_amount=GREATEST(0,EXCLUDED.revenue_amount-order_financial_summaries.confirmed_payment_amount+order_financial_summaries.refunded_amount),updated_at=NOW();
  RETURN NEW;
END $$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_initialize_order_financial_summary ON order_commercial_terms;
CREATE TRIGGER trg_initialize_order_financial_summary AFTER INSERT OR UPDATE ON order_commercial_terms FOR EACH ROW EXECUTE FUNCTION initialize_order_financial_summary();

CREATE TABLE IF NOT EXISTS daily_operations_report (
  report_date DATE PRIMARY KEY,
  active_workflows INT NOT NULL DEFAULT 0,
  overdue_steps INT NOT NULL DEFAULT 0,
  open_discrepancies INT NOT NULL DEFAULT 0,
  shipments_dispatched INT NOT NULL DEFAULT 0,
  shipments_delivered INT NOT NULL DEFAULT 0,
  on_time_deliveries INT NOT NULL DEFAULT 0,
  average_step_duration_hours NUMERIC(18,4) NOT NULL DEFAULT 0,
  rework_iterations INT NOT NULL DEFAULT 0,
  refreshed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE daily_operations_report
  ADD COLUMN IF NOT EXISTS average_step_duration_hours NUMERIC(18,4) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS rework_iterations INT NOT NULL DEFAULT 0;

INSERT INTO document_templates(code,name_fa,document_type,template_json,allowed_variables) VALUES
('PROFORMA_FA_V1','پیش‌فاکتور فارسی','PROFORMA','{"title":"پیش‌فاکتور","sections":["customer","items","totals","terms"]}'::jsonb,'["document_number","order_number","customer_name","items","subtotal","discount","tax","charges","total","currency","issued_at"]'::jsonb),
('PAYMENT_RECEIPT_FA_V1','رسید پرداخت فارسی','PAYMENT_RECEIPT','{"title":"رسید پرداخت","sections":["customer","payment"]}'::jsonb,'["document_number","order_number","customer_name","payment_number","amount","currency","paid_at","reference"]'::jsonb),
('ORDER_SUMMARY_FA_V1','خلاصه سفارش فارسی','ORDER_SUMMARY','{"title":"خلاصه سفارش","sections":["customer","items","delivery"]}'::jsonb,'["document_number","order_number","customer_name","items","status","estimated_delivery_at"]'::jsonb),
('PACKING_LIST_FA_V1','فهرست بسته‌بندی فارسی','PACKING_LIST','{"title":"فهرست بسته‌بندی","sections":["shipment","packages"]}'::jsonb,'["document_number","order_number","shipment_number","packages","issued_at"]'::jsonb),
('DELIVERY_NOTE_FA_V1','رسید تحویل فارسی','DELIVERY_NOTE','{"title":"رسید تحویل","sections":["shipment","receiver"]}'::jsonb,'["document_number","order_number","shipment_number","receiver_name","delivered_at","items"]'::jsonb)
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa,template_json=EXCLUDED.template_json,allowed_variables=EXCLUDED.allowed_variables;

INSERT INTO notification_templates(event_type,channel,locale,title_template,body_template,allowed_variables) VALUES
('PAYMENT_DUE','IN_APP','fa','سررسید پرداخت سفارش {{order_number}}','مبلغ {{amount}} {{currency}} سررسید شده است.','["order_number","amount","currency"]'::jsonb),
('PAYMENT_CONFIRMED','IN_APP','fa','پرداخت تأیید شد','پرداخت {{amount}} {{currency}} برای سفارش {{order_number}} تأیید شد.','["order_number","amount","currency"]'::jsonb),
('WORKFLOW_DELAY','IN_APP','fa','تأخیر در فرایند {{order_number}}','مرحله {{step_name}} از زمان برنامه عقب افتاده است.','["order_number","step_name"]'::jsonb),
('SHIPMENT_ETA','IN_APP','fa','زمان تقریبی تحویل','محموله {{shipment_number}} در تاریخ {{eta}} تحویل می‌شود.','["shipment_number","eta"]'::jsonb),
('SALES_FOLLOW_UP','IN_APP','fa','پیگیری مشتری','برای سفارش {{order_number}} با مشتری تماس بگیرید.','["order_number"]'::jsonb),
('PAYMENT_CONFIRMED','SMS','fa','پرداخت تأیید شد','پرداخت {{amount}} {{currency}} سفارش {{order_number}} تأیید شد.','["order_number","amount","currency"]'::jsonb)
ON CONFLICT(event_type,channel,locale) DO UPDATE SET title_template=EXCLUDED.title_template,body_template=EXCLUDED.body_template,allowed_variables=EXCLUDED.allowed_variables,is_active=TRUE;

INSERT INTO notification_templates(event_type,channel,locale,audience_type,title_template,body_template,allowed_variables) VALUES
('CUSTOMER_ACCOUNT_CREATED','IN_APP','fa','CUSTOMER','حساب مشتری ایجاد شد','حساب کاربری سفارش شما ایجاد شده است.','[]'::jsonb),
('CUSTOMER_ACCOUNT_ACTIVATED','IN_APP','fa','CUSTOMER','حساب مشتری فعال شد','حساب کاربری شما با موفقیت فعال شد.','[]'::jsonb),
('PROFORMA_ISSUED','IN_APP','fa','CUSTOMER','پیش‌فاکتور صادر شد','پیش‌فاکتور جدید برای سفارش شما صادر شده است.','[]'::jsonb),
('ORDER_CONFIRMED','IN_APP','fa','CUSTOMER','سفارش تأیید شد','سفارش شما تأیید و وارد فرایند اجرا شد.','[]'::jsonb),
('PAYMENT_REQUIRED','IN_APP','fa','CUSTOMER','پرداخت موردنیاز است','برای ادامه سفارش، پرداخت برنامه‌ریزی‌شده را انجام دهید.','[]'::jsonb),
('PAYMENT_DUE_SOON','IN_APP','fa','CUSTOMER','سررسید پرداخت نزدیک است','سررسید یکی از پرداخت‌های سفارش شما نزدیک است.','[]'::jsonb),
('PAYMENT_OVERDUE','IN_APP','fa','CUSTOMER','پرداخت عقب‌افتاده','یکی از پرداخت‌های سفارش شما از سررسید گذشته است.','[]'::jsonb),
('PAYMENT_REJECTED','IN_APP','fa','CUSTOMER','پرداخت رد شد','پرداخت ثبت‌شده تأیید نشد؛ لطفاً جزئیات را بررسی کنید.','[]'::jsonb),
('WORKFLOW_STARTED','IN_APP','fa','ASSIGNED_USER','فرایند آغاز شد','یک فرایند عملیاتی جدید آغاز شده است.','[]'::jsonb),
('WORKFLOW_STEP_STARTED','IN_APP','fa','ASSIGNED_USER','مرحله آغاز شد','مرحله عملیاتی تخصیص‌یافته آغاز شده است.','[]'::jsonb),
('WORKFLOW_STEP_COMPLETED','IN_APP','fa','ASSIGNED_USER','مرحله تکمیل شد','مرحله عملیاتی با موفقیت تکمیل شد.','[]'::jsonb),
('WORKFLOW_DELAYED','IN_APP','fa','ADMIN','تأخیر در فرایند','یک مرحله عملیاتی از زمان برنامه عقب افتاده است.','[]'::jsonb),
('ORDER_IN_PRODUCTION','IN_APP','fa','CUSTOMER','سفارش در حال تولید است','سفارش شما وارد مرحله تولید شده است.','[]'::jsonb),
('ORDER_READY_FOR_PACKAGING','IN_APP','fa','CUSTOMER','سفارش آماده بسته‌بندی است','سفارش شما آماده بسته‌بندی شده است.','[]'::jsonb),
('ORDER_READY_FOR_SHIPMENT','IN_APP','fa','CUSTOMER','سفارش آماده ارسال است','سفارش شما آماده ارسال شده است.','[]'::jsonb),
('SHIPMENT_LOADING','IN_APP','fa','CUSTOMER','بارگیری آغاز شد','بارگیری محموله سفارش شما آغاز شده است.','[]'::jsonb),
('SHIPMENT_DISPATCHED','IN_APP','fa','CUSTOMER','محموله ارسال شد','محموله سفارش شما ارسال شده است.','[]'::jsonb),
('SHIPMENT_ARRIVED','IN_APP','fa','CUSTOMER','محموله رسید','محموله سفارش شما به مقصد رسیده است.','[]'::jsonb),
('SHIPMENT_DELIVERED','IN_APP','fa','CUSTOMER','محموله تحویل شد','تحویل محموله سفارش شما ثبت شد.','[]'::jsonb),
('INSTALLATION_SCHEDULED','IN_APP','fa','CUSTOMER','نصب زمان‌بندی شد','زمان اجرای نصب سفارش شما مشخص شده است.','[]'::jsonb),
('INSTALLATION_STARTED','IN_APP','fa','CUSTOMER','نصب آغاز شد','عملیات نصب سفارش شما آغاز شده است.','[]'::jsonb),
('INSTALLATION_COMPLETED','IN_APP','fa','CUSTOMER','نصب تکمیل شد','عملیات نصب سفارش شما تکمیل شده است.','[]'::jsonb),
('ORDER_COMPLETED','IN_APP','fa','CUSTOMER','سفارش تکمیل شد','سفارش شما با موفقیت تکمیل شده است.','[]'::jsonb),
('ACTION_ITEM_ASSIGNED','IN_APP','fa','ASSIGNED_USER','اقدام جدید','یک اقدام جدید به شما تخصیص داده شده است.','[]'::jsonb),
('ACTION_ITEM_OVERDUE','IN_APP','fa','ASSIGNED_USER','اقدام عقب‌افتاده','مهلت یک اقدام تخصیص‌یافته گذشته است.','[]'::jsonb),
('DISCREPANCY_CREATED','IN_APP','fa','ADMIN','مغایرت جدید','یک مغایرت عملیاتی جدید ثبت شده است.','[]'::jsonb),
('DISCREPANCY_RESOLVED','IN_APP','fa','ADMIN','مغایرت رفع شد','مغایرت عملیاتی رفع شده است.','[]'::jsonb),
('PAYMENT_CONFIRMATION_REQUIRED','IN_APP','fa','ACCOUNTANT','تأیید پرداخت موردنیاز است','یک پرداخت جدید در انتظار بررسی حسابدار است.','[]'::jsonb),
('COST_APPROVAL_REQUIRED','IN_APP','fa','ACCOUNTANT','تأیید هزینه موردنیاز است','یک هزینه عملیاتی در انتظار تأیید است.','[]'::jsonb),
('SHIPMENT_DELAYED','IN_APP','fa','ADMIN','تأخیر محموله','محموله از زمان تخمینی تحویل عقب افتاده است.','[]'::jsonb)
ON CONFLICT(event_type,channel,locale) DO NOTHING;

INSERT INTO notification_templates(event_type,channel,locale,audience_type,title_template,body_template,allowed_variables,is_active) VALUES
('CUSTOMER_ACCOUNT_CREATED','SMS','fa','CUSTOMER','حساب مشتری','حساب سفارش شما در سنگ حسن ایجاد شد.','[]'::jsonb,TRUE),
('CUSTOMER_ACCOUNT_ACTIVATED','SMS','fa','CUSTOMER','فعال‌سازی حساب','حساب سفارش شما در سنگ حسن فعال شد.','[]'::jsonb,TRUE),
('PROFORMA_ISSUED','SMS','fa','CUSTOMER','صدور پیش‌فاکتور','پیش‌فاکتور جدید شما در حساب سنگ حسن آماده است.','[]'::jsonb,TRUE),
('ORDER_CONFIRMED','SMS','fa','CUSTOMER','تأیید سفارش','سفارش شما در سنگ حسن تأیید شد.','[]'::jsonb,TRUE),
('PAYMENT_REQUIRED','SMS','fa','CUSTOMER','درخواست پرداخت','برای ادامه سفارش، وضعیت پرداخت حساب خود را بررسی کنید.','[]'::jsonb,TRUE),
('SHIPMENT_DISPATCHED','SMS','fa','CUSTOMER','ارسال سفارش','محموله سفارش شما ارسال شد.','[]'::jsonb,TRUE),
('SHIPMENT_ETA','SMS','fa','CUSTOMER','زمان تحویل','زمان تخمینی تحویل محموله شما نزدیک است.','[]'::jsonb,TRUE),
('SHIPMENT_DELIVERED','SMS','fa','CUSTOMER','تحویل سفارش','تحویل محموله سفارش شما ثبت شد.','[]'::jsonb,TRUE),
('INSTALLATION_STARTED','SMS','fa','CUSTOMER','شروع نصب','عملیات نصب سفارش شما آغاز شد.','[]'::jsonb,TRUE),
('ORDER_COMPLETED','SMS','fa','CUSTOMER','تکمیل سفارش','سفارش شما در سنگ حسن تکمیل شد.','[]'::jsonb,TRUE)
ON CONFLICT(event_type,channel,locale) DO NOTHING;

INSERT INTO permissions(code,name_fa,description_fa,group_code) VALUES
('orders.confirm','تأیید سفارش','تأیید تجاری سفارش','ORDERS'),
('finance.commercial_terms.view','مشاهده شرایط تجاری','مشاهده مبلغ و شرایط فروش','FINANCE'),
('finance.commercial_terms.manage','مدیریت شرایط تجاری','ویرایش مبلغ و شرایط فروش','FINANCE'),
('finance.payment_schedule.view','مشاهده برنامه پرداخت','مشاهده اقساط و سررسیدها','FINANCE'),
('finance.payment_schedule.manage','مدیریت برنامه پرداخت','ویرایش اقساط و سررسیدها','FINANCE'),
('finance.payments.refund','بازپرداخت وجه','ثبت بازپرداخت وجه مشتری','FINANCE'),
('finance.costs.view','مشاهده هزینه‌ها','مشاهده هزینه عملیاتی','FINANCE'),
('finance.costs.view_assigned','مشاهده هزینه‌های تخصیص‌یافته','مشاهده هزینه Scope عملیاتی خود','FINANCE'),
('finance.costs.record','ثبت هزینه','ثبت هزینه عملیاتی','FINANCE'),
('finance.costs.approve','تأیید هزینه','تأیید یا رد هزینه','FINANCE'),
('finance.costs.pay','پرداخت هزینه','ثبت پرداخت هزینه','FINANCE'),
('finance.exchange_rates.manage','مدیریت نرخ ارز','ثبت نرخ دستی ارز','FINANCE'),
('notifications.view_own','مشاهده اعلان‌های خود','مشاهده اعلان داخلی','NOTIFICATIONS'),
('notifications.preferences.manage','تنظیم اعلان‌ها','مدیریت ترجیحات اعلان','NOTIFICATIONS'),
('notifications.templates.manage','مدیریت قالب اعلان','ویرایش قالب‌های اعلان','NOTIFICATIONS'),
('notifications.deliveries.retry','تلاش مجدد اعلان','تلاش مجدد تحویل اعلان','NOTIFICATIONS'),
('documents.view','مشاهده اسناد','مشاهده اسناد سفارش','DOCUMENTS'),
('documents.generate','تولید سند','تولید PDF خصوصی','DOCUMENTS'),
('documents.upload','بارگذاری سند','بارگذاری سند خصوصی','DOCUMENTS'),
('documents.issue','صدور سند','صدور نسخه نهایی سند','DOCUMENTS'),
('documents.cancel','لغو سند','لغو سند صادرشده','DOCUMENTS'),
('documents.download','دریافت سند','دریافت فایل سند','DOCUMENTS'),
('document_templates.manage','مدیریت قالب سند','مدیریت قالب JSON سند','DOCUMENTS'),
('workflow_document_requirements.manage','مدیریت چک‌لیست اسناد','مدیریت الزام سند Workflow','DOCUMENTS'),
('reports.overview.view','گزارش مدیریتی','مشاهده نمای کلی گزارش‌ها','REPORTS'),
('reports.receivables.view','گزارش مطالبات','مشاهده سررسید و مطالبات','REPORTS'),
('reports.costs.view','گزارش هزینه‌ها','مشاهده گزارش هزینه','REPORTS'),
('reports.profitability.view','گزارش سودآوری','مشاهده سودآوری سفارش','REPORTS'),
('reports.operations.view','گزارش عملیات','مشاهده شاخص‌های عملیات','REPORTS'),
('reports.sales.view','گزارش فروش','مشاهده پیگیری‌های فروش','REPORTS'),
('customers.contacts.view','مشاهده ارتباطات مشتری','مشاهده تاریخچه تماس','CUSTOMERS'),
('customers.contacts.record','ثبت ارتباط مشتری','ثبت تماس و پیگیری','CUSTOMERS'),
('customer_portal.financial_summary.view_own','مشاهده مالی سفارش خود','مشاهده مانده حساب مشتری','CUSTOMER_PORTAL'),
('customer_portal.documents.view_own','مشاهده اسناد خود','مشاهده اسناد قابل نمایش مشتری','CUSTOMER_PORTAL'),
('finance.customer_payments.view','مشاهده پرداخت مشتری','مشاهده پرداخت‌های مشتری','FINANCE'),
('finance.customer_payments.record','ثبت پرداخت مشتری','ثبت اعلام پرداخت مشتری','FINANCE'),
('finance.customer_payments.confirm','تأیید پرداخت مشتری','تأیید و تخصیص پرداخت مشتری','FINANCE'),
('finance.customer_payments.reject','رد پرداخت مشتری','رد پرداخت گزارش‌شده','FINANCE'),
('finance.customer_payments.refund','بازپرداخت مشتری','ثبت بازپرداخت مشتری','FINANCE'),
('finance.costs.view_all','مشاهده همه هزینه‌ها','مشاهده همه هزینه‌های داخلی','FINANCE'),
('finance.costs.mark_paid','ثبت پرداخت هزینه','ثبت پرداخت هزینه تأییدشده','FINANCE'),
('finance.exchange_rates.view','مشاهده نرخ ارز','مشاهده نرخ‌های دستی ارز','FINANCE'),
('notifications.view_all','مشاهده همه اعلان‌ها','مشاهده اعلان‌های عملیاتی','NOTIFICATIONS'),
('notifications.templates.view','مشاهده قالب اعلان','مشاهده قالب‌های اعلان','NOTIFICATIONS'),
('notifications.delivery.view','مشاهده تحویل اعلان','مشاهده وضعیت Outbox اعلان','NOTIFICATIONS'),
('notifications.retry','تلاش مجدد اعلان','تلاش مجدد تحویل اعلان','NOTIFICATIONS'),
('documents.view_assigned','مشاهده اسناد تخصیص‌یافته','مشاهده اسناد Scope تخصیص‌یافته','DOCUMENTS'),
('documents.view_all','مشاهده همه اسناد','مشاهده همه اسناد داخلی','DOCUMENTS'),
('documents.create','ایجاد سند','بارگذاری یا ایجاد رکورد سند','DOCUMENTS'),
('documents.download_internal','دانلود سند داخلی','دانلود فایل سند خصوصی','DOCUMENTS'),
('documents.templates.manage','مدیریت قالب سند','مدیریت قالب‌های سند','DOCUMENTS'),
('reports.finance.view','گزارش مالی','مشاهده گزارش مالی مدیریتی','REPORTS')
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa,description_fa=EXCLUDED.description_fa,group_code=EXCLUDED.group_code;

INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE
 r.code='SUPER_ADMIN'
 OR r.code='ADMIN' AND (p.group_code IN ('FINANCE','NOTIFICATIONS','DOCUMENTS','REPORTS','CUSTOMERS') OR p.code='orders.confirm')
 OR r.code='SALES' AND p.code IN ('orders.confirm','finance.commercial_terms.view','finance.commercial_terms.manage','finance.payment_schedule.view','finance.payments.view','finance.payments.record','finance.customer_payments.view','finance.customer_payments.record','documents.view_assigned','documents.generate','documents.download','reports.sales.view','customers.contacts.view','customers.contacts.record','notifications.view_own','notifications.preferences.manage')
 OR r.code='ACCOUNTANT' AND p.code IN ('finance.commercial_terms.view','finance.payment_schedule.view','finance.payment_schedule.manage','finance.payments.view','finance.payments.record','finance.payments.confirm','finance.payments.refund','finance.customer_payments.view','finance.customer_payments.record','finance.customer_payments.confirm','finance.customer_payments.reject','finance.customer_payments.refund','finance.costs.view','finance.costs.view_all','finance.costs.record','finance.costs.approve','finance.costs.pay','finance.costs.mark_paid','finance.exchange_rates.view','finance.exchange_rates.manage','finance.profit.view','reports.overview.view','reports.finance.view','reports.receivables.view','reports.costs.view','reports.profitability.view','documents.view','documents.view_all','documents.generate','documents.issue','documents.download','documents.download_internal','notifications.view_own','notifications.preferences.manage')
 OR r.code IN ('SUPPLY','DRIVER','INSTALLATION_LEAD') AND p.code IN ('finance.costs.view_assigned','finance.costs.record','notifications.view_own','notifications.preferences.manage','documents.view_assigned','documents.upload','documents.create')
 OR r.code='OPERATOR' AND p.code IN ('notifications.view_own','notifications.preferences.manage','documents.view','documents.generate','documents.download','customers.contacts.view','customers.contacts.record')
 OR r.code='CUSTOMER' AND p.code IN ('notifications.view_own','notifications.preferences.manage','customer_portal.financial_summary.view_own','customer_portal.documents.view_own')
ON CONFLICT DO NOTHING;

INSERT INTO order_financial_summaries(order_id,currency,revenue_amount,outstanding_amount)
SELECT order_id,currency,CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED') THEN final_customer_amount ELSE 0 END,
       CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED') THEN final_customer_amount ELSE 0 END
FROM order_commercial_terms t JOIN orders o ON o.id=t.order_id
ON CONFLICT(order_id) DO NOTHING;
