-- Production-readiness settings, diagnostics, search and query indexes.
-- This migration is intentionally idempotent and does not alter migrations 014-018.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version INT PRIMARY KEY,
  migration_name TEXT NOT NULL UNIQUE,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO schema_migrations(version,migration_name) VALUES
  (14,'operations_phase1'),
  (15,'operations_phase2'),
  (16,'operations_phase3'),
  (17,'finance_notifications_documents_reporting'),
  (18,'supplier_purchase_quality_installation'),
  (19,'application_settings_diagnostics_indexes')
ON CONFLICT(version) DO UPDATE SET migration_name=EXCLUDED.migration_name;

CREATE TABLE IF NOT EXISTS application_settings (
  id BIGSERIAL PRIMARY KEY,
  setting_key TEXT NOT NULL UNIQUE,
  setting_value_json JSONB NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  updated_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO application_settings(setting_key,setting_value_json,description) VALUES
  ('default_currency','"IRR"','ارز پیش‌فرض عملیات'),
  ('default_country_code','"IR"','کشور پیش‌فرض'),
  ('default_phone_country','"+98"','پیش‌شماره تلفن پیش‌فرض'),
  ('default_timezone','"Asia/Tehran"','منطقه زمانی رابط کاربری'),
  ('customer_portal_enabled','true','فعال بودن حساب عملیاتی مشتری'),
  ('sms_enabled','false','فعال بودن کانال پیامک'),
  ('installation_module_enabled','true','فعال بودن ایجاد عملیات نصب'),
  ('inventory_module_enabled','true','فعال بودن mutationهای موجودی'),
  ('supplier_module_enabled','true','فعال بودن ایجاد تأمین‌کننده و خرید'),
  ('allow_manager_force_close','true','اجازه بستن سفارش با هشدار'),
  ('allow_manager_workflow_override','true','اجازه override مدیریتی Workflow'),
  ('default_payment_due_days','7','تعداد روز پیش‌فرض سررسید'),
  ('default_workflow_warning_hours','24','ساعت هشدار تأخیر Workflow'),
  ('max_upload_size_mb','15','حداکثر اندازه فایل خصوصی')
ON CONFLICT(setting_key) DO NOTHING;

CREATE TABLE IF NOT EXISTS saved_views (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  view_key TEXT NOT NULL,
  name TEXT NOT NULL,
  filters_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id,view_key,name),
  CHECK(jsonb_typeof(filters_json)='object')
);
CREATE INDEX IF NOT EXISTS idx_saved_views_user ON saved_views(user_id,view_key,updated_at DESC);

CREATE TABLE IF NOT EXISTS reconciliation_findings (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  finding_key TEXT NOT NULL UNIQUE,
  check_code TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  severity TEXT NOT NULL DEFAULT 'WARNING',
  status TEXT NOT NULL DEFAULT 'OPEN',
  summary TEXT NOT NULL,
  details_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  safe_repair_code TEXT,
  first_detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMPTZ,
  resolved_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  resolution_reason TEXT,
  CHECK(severity IN ('INFO','WARNING','CRITICAL')),
  CHECK(status IN ('OPEN','REPAIRED','RESOLVED','IGNORED'))
);
CREATE INDEX IF NOT EXISTS idx_reconciliation_open ON reconciliation_findings(status,severity,last_detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_reconciliation_entity ON reconciliation_findings(entity_type,entity_id,status);

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS request_id TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_invalid_before TIMESTAMPTZ;

-- Query-pattern indexes used by dashboards, operational lists and diagnostics.
CREATE INDEX IF NOT EXISTS idx_orders_status_created ON orders(status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_customer_status ON orders(customer_user_id,status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_workflow_instances_order_status ON workflow_instances(order_id,status);
CREATE INDEX IF NOT EXISTS idx_workflow_instances_status_updated ON workflow_instances(status,updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_instance_status ON workflow_step_instances(workflow_instance_id,status,sequence_number);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_user_open ON workflow_step_instances(assigned_user_id,status,estimated_end_at) WHERE status NOT IN ('COMPLETED','SKIPPED','CANCELLED');
CREATE INDEX IF NOT EXISTS idx_action_items_user_open_due ON action_items(assigned_user_id,status,due_at) WHERE status NOT IN ('COMPLETED','CANCELLED');
CREATE INDEX IF NOT EXISTS idx_action_items_role_open_due ON action_items(assigned_role_id,status,due_at) WHERE status NOT IN ('COMPLETED','CANCELLED');
CREATE INDEX IF NOT EXISTS idx_batches_order_status_created ON fulfillment_batches(order_id,status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_inventory_lots_location_status ON inventory_lots(current_location_id,status,updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_shipments_order_status_created ON shipments(order_id,status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shipments_driver_open_eta ON shipments(driver_user_id,status,estimated_arrival_at) WHERE status NOT IN ('DELIVERED','CANCELLED');
CREATE INDEX IF NOT EXISTS idx_customer_payments_order_status_date ON customer_payments(order_id,status,paid_at DESC);
CREATE INDEX IF NOT EXISTS idx_costs_order_status_created ON operational_cost_entries(order_id,status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_status_created ON notifications(user_id,status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_entity_created ON audit_logs(entity_type,entity_id,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_actor_created ON audit_logs(actor_user_id,created_at DESC);

CREATE INDEX IF NOT EXISTS idx_users_operational_search_trgm
  ON users USING gin ((COALESCE(first_name,'')||' '||COALESCE(last_name,'')||' '||COALESCE(phone_normalized,'')) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_orders_number_trgm ON orders USING gin (order_number gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_proformas_number_trgm ON proformas USING gin (proforma_number gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_shipments_number_trgm ON shipments USING gin (shipment_number gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_batches_number_trgm ON fulfillment_batches USING gin (batch_number gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_payments_number_trgm ON customer_payments USING gin (payment_number gin_trgm_ops);

INSERT INTO permissions(code,name_fa,description_fa,group_code) VALUES
  ('settings.view','مشاهده تنظیمات','مشاهده تنظیمات و اطلاعات سیستم','SYSTEM'),
  ('settings.manage','مدیریت تنظیمات','ویرایش تنظیمات شناخته‌شده سیستم','SYSTEM'),
  ('admin_tools.view','مشاهده ابزارهای اصلاح','مشاهده ابزارهای امن اصلاح داده','SYSTEM'),
  ('admin_tools.workflow_repair','اصلاح Workflow','اجرای اصلاحات امن Workflow','SYSTEM'),
  ('admin_tools.order_repair','اصلاح سفارش','بازمحاسبه داده‌های سفارش','SYSTEM'),
  ('admin_tools.sessions.revoke','ابطال Session','خروج اجباری Sessionهای کاربر','USERS'),
  ('diagnostics.view','مشاهده عیب‌یابی','مشاهده یافته‌های سازگاری داده','SYSTEM'),
  ('diagnostics.repair','اصلاح یافته امن','اجرای Repairهای محدود و امن','SYSTEM'),
  ('exports.orders','خروجی سفارش‌ها','دریافت CSV سفارش‌ها','EXPORTS'),
  ('exports.customers','خروجی مشتریان','دریافت CSV مشتریان','EXPORTS'),
  ('exports.payments','خروجی پرداخت‌ها','دریافت CSV پرداخت‌ها','EXPORTS'),
  ('exports.costs','خروجی هزینه‌ها','دریافت CSV هزینه‌های داخلی','EXPORTS'),
  ('exports.shipments','خروجی محموله‌ها','دریافت CSV محموله‌ها','EXPORTS')
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa,description_fa=EXCLUDED.description_fa,group_code=EXCLUDED.group_code,is_active=TRUE;

INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.code IN ('SUPER_ADMIN','ADMIN') AND p.code IN (
  'settings.view','settings.manage','admin_tools.view','admin_tools.workflow_repair','admin_tools.order_repair',
  'admin_tools.sessions.revoke','diagnostics.view','diagnostics.repair','exports.orders','exports.customers',
  'exports.payments','exports.costs','exports.shipments')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p ON
  (r.code='ACCOUNTANT' AND p.code IN ('exports.payments','exports.costs')) OR
  (r.code='SALES' AND p.code IN ('exports.orders','exports.customers','exports.shipments'))
ON CONFLICT DO NOTHING;

-- Keep only actionable operational notifications enabled by default.
UPDATE notification_templates SET is_active=FALSE
WHERE event_type IN ('WORKFLOW_STEP_STARTED','WORKFLOW_STEP_COMPLETED')
  AND audience_type='ASSIGNED_USER';
