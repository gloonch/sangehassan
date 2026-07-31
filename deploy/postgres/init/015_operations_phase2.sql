-- SangeHassan Operations Dashboard Phase 2
-- Migration-safe on top of 014_operations_phase1.sql. Do not fold into Phase 1.

ALTER TABLE workflow_templates
  ADD COLUMN IF NOT EXISTS template_group_code TEXT,
  ADD COLUMN IF NOT EXISTS version_number INT,
  ADD COLUMN IF NOT EXISTS status TEXT,
  ADD COLUMN IF NOT EXISTS created_from_template_id BIGINT REFERENCES workflow_templates(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS published_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL;

UPDATE workflow_templates SET
  template_group_code = COALESCE(template_group_code, code),
  version_number = COALESCE(version_number, 1),
  status = COALESCE(status, 'PUBLISHED'),
  published_at = COALESCE(published_at, created_at)
WHERE template_group_code IS NULL OR version_number IS NULL OR status IS NULL OR published_at IS NULL;
ALTER TABLE workflow_templates ALTER COLUMN template_group_code SET NOT NULL;
ALTER TABLE workflow_templates ALTER COLUMN version_number SET NOT NULL;
ALTER TABLE workflow_templates ALTER COLUMN status SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_workflow_template_group_version ON workflow_templates(template_group_code, version_number);
CREATE INDEX IF NOT EXISTS idx_workflow_templates_latest ON workflow_templates(template_group_code, status, version_number DESC);

ALTER TABLE workflow_template_steps
  ADD COLUMN IF NOT EXISTS internal_description_fa TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS customer_description_fa TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS responsible_role_id BIGINT REFERENCES roles(id) ON DELETE RESTRICT,
  ADD COLUMN IF NOT EXISTS requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS approval_role_id BIGINT REFERENCES roles(id) ON DELETE RESTRICT,
  ADD COLUMN IF NOT EXISTS is_optional BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS is_skippable BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS default_duration_hours INT NOT NULL DEFAULT 24,
  ADD COLUMN IF NOT EXISTS starts_automatically BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE workflow_template_steps s SET responsible_role_id=r.id FROM roles r
WHERE s.responsible_role_id IS NULL AND r.code=s.responsible_role_code;
CREATE UNIQUE INDEX IF NOT EXISTS uq_workflow_template_step_sequence ON workflow_template_steps(workflow_template_id, sequence_number) WHERE is_active;

CREATE TABLE IF NOT EXISTS workflow_step_catalogue (
  id BIGSERIAL PRIMARY KEY, step_code TEXT NOT NULL UNIQUE, title_fa TEXT NOT NULL,
  customer_title_fa TEXT NOT NULL, description_fa TEXT NOT NULL DEFAULT '',
  default_role_code TEXT, default_permission_code TEXT, default_duration_hours INT NOT NULL DEFAULT 24,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_step_field_definitions (
  id BIGSERIAL PRIMARY KEY, workflow_template_step_id BIGINT NOT NULL REFERENCES workflow_template_steps(id) ON DELETE CASCADE,
  field_key TEXT NOT NULL, label_fa TEXT NOT NULL, description_fa TEXT NOT NULL DEFAULT '', field_type TEXT NOT NULL,
  is_required BOOLEAN NOT NULL DEFAULT FALSE, is_customer_visible BOOLEAN NOT NULL DEFAULT FALSE,
  is_sales_visible BOOLEAN NOT NULL DEFAULT FALSE, is_internal_cost BOOLEAN NOT NULL DEFAULT FALSE,
  unit_code TEXT, currency_code TEXT, placeholder_fa TEXT, default_value JSONB,
  options_json JSONB, validation_json JSONB NOT NULL DEFAULT '{}'::jsonb, sort_order INT NOT NULL DEFAULT 0,
  handoff_metric_key TEXT, handoff_direction TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workflow_template_step_id, field_key)
);
CREATE INDEX IF NOT EXISTS idx_workflow_field_definitions_step ON workflow_step_field_definitions(workflow_template_step_id, sort_order);

CREATE TABLE IF NOT EXISTS workflow_step_task_templates (
  id BIGSERIAL PRIMARY KEY, workflow_template_step_id BIGINT NOT NULL REFERENCES workflow_template_steps(id) ON DELETE CASCADE,
  trigger_type TEXT NOT NULL, title_fa TEXT NOT NULL, description_fa TEXT NOT NULL DEFAULT '',
  assigned_role_id BIGINT REFERENCES roles(id) ON DELETE RESTRICT, required_permission_code TEXT,
  priority TEXT NOT NULL DEFAULT 'NORMAL', due_offset_hours INT, blocks_step_completion BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS workflow_task_templates (
  id BIGSERIAL PRIMARY KEY, workflow_template_id BIGINT NOT NULL REFERENCES workflow_templates(id) ON DELETE CASCADE,
  trigger_type TEXT NOT NULL, title_fa TEXT NOT NULL, description_fa TEXT NOT NULL DEFAULT '',
  assigned_role_id BIGINT REFERENCES roles(id) ON DELETE RESTRICT, required_permission_code TEXT,
  priority TEXT NOT NULL DEFAULT 'NORMAL', due_offset_hours INT, blocks_workflow_progress BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS workflow_handoff_metric_definitions (
  id BIGSERIAL PRIMARY KEY, workflow_template_id BIGINT NOT NULL REFERENCES workflow_templates(id) ON DELETE CASCADE,
  metric_key TEXT NOT NULL, label_fa TEXT NOT NULL, unit_code TEXT NOT NULL,
  absolute_tolerance NUMERIC(18,4), percentage_tolerance NUMERIC(10,4), blocking_on_mismatch BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workflow_template_id, metric_key)
);

ALTER TABLE workflow_instances
  ADD COLUMN IF NOT EXISTS template_group_code TEXT,
  ADD COLUMN IF NOT EXISTS template_version_number INT,
  ADD COLUMN IF NOT EXISTS current_step_instance_id UUID,
  ADD COLUMN IF NOT EXISTS estimated_end_at TIMESTAMPTZ;
UPDATE workflow_instances wi SET template_group_code=wt.template_group_code,template_version_number=wt.version_number
FROM workflow_templates wt WHERE wt.id=wi.workflow_template_id AND wi.template_group_code IS NULL;

ALTER TABLE workflow_step_instances
  ADD COLUMN IF NOT EXISTS template_step_id BIGINT REFERENCES workflow_template_steps(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS step_code TEXT,
  ADD COLUMN IF NOT EXISTS internal_title_fa TEXT,
  ADD COLUMN IF NOT EXISTS internal_description_fa TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS customer_title_fa TEXT,
  ADD COLUMN IF NOT EXISTS customer_description_fa TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS sequence_number INT,
  ADD COLUMN IF NOT EXISTS responsible_role_id BIGINT REFERENCES roles(id) ON DELETE RESTRICT,
  ADD COLUMN IF NOT EXISTS required_permission_code TEXT,
  ADD COLUMN IF NOT EXISTS requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS approval_role_id BIGINT REFERENCES roles(id) ON DELETE RESTRICT,
  ADD COLUMN IF NOT EXISTS is_optional BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS is_skippable BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS starts_automatically BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS submitted_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS approved_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS rejected_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS rejection_reason TEXT,
  ADD COLUMN IF NOT EXISTS skipped_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS skipped_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS skip_reason TEXT;
UPDATE workflow_step_instances si SET
  template_step_id=COALESCE(si.template_step_id,si.workflow_template_step_id),step_code=COALESCE(si.step_code,ts.step_code),
  internal_title_fa=COALESCE(si.internal_title_fa,ts.internal_title_fa),customer_title_fa=COALESCE(si.customer_title_fa,ts.customer_title_fa),
  sequence_number=COALESCE(si.sequence_number,ts.sequence_number),responsible_role_id=COALESCE(si.responsible_role_id,si.assigned_role_id,ts.responsible_role_id),
  required_permission_code=COALESCE(si.required_permission_code,ts.required_permission_code),customer_visible=COALESCE(si.customer_visible,ts.customer_visible)
FROM workflow_template_steps ts WHERE ts.id=si.workflow_template_step_id AND si.step_code IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_workflow_instance_step_sequence ON workflow_step_instances(workflow_instance_id, sequence_number) WHERE sequence_number IS NOT NULL;

CREATE TABLE IF NOT EXISTS workflow_instance_field_definitions (
  id BIGSERIAL PRIMARY KEY, workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  workflow_step_instance_id UUID NOT NULL REFERENCES workflow_step_instances(id) ON DELETE CASCADE,
  source_field_definition_id BIGINT REFERENCES workflow_step_field_definitions(id) ON DELETE SET NULL,
  field_key TEXT NOT NULL, label_fa TEXT NOT NULL, description_fa TEXT NOT NULL DEFAULT '', field_type TEXT NOT NULL,
  is_required BOOLEAN NOT NULL DEFAULT FALSE, is_customer_visible BOOLEAN NOT NULL DEFAULT FALSE,
  is_sales_visible BOOLEAN NOT NULL DEFAULT FALSE, is_internal_cost BOOLEAN NOT NULL DEFAULT FALSE,
  unit_code TEXT, currency_code TEXT, placeholder_fa TEXT, default_value JSONB, options_json JSONB,
  validation_json JSONB NOT NULL DEFAULT '{}'::jsonb, sort_order INT NOT NULL DEFAULT 0,
  handoff_metric_key TEXT, handoff_direction TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workflow_step_instance_id, field_key)
);
CREATE TABLE IF NOT EXISTS workflow_instance_step_task_templates (
  id BIGSERIAL PRIMARY KEY, workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  workflow_step_instance_id UUID NOT NULL REFERENCES workflow_step_instances(id) ON DELETE CASCADE,
  source_task_template_id BIGINT REFERENCES workflow_step_task_templates(id) ON DELETE SET NULL,
  trigger_type TEXT NOT NULL, title_fa TEXT NOT NULL, description_fa TEXT NOT NULL DEFAULT '',
  assigned_role_id BIGINT REFERENCES roles(id), required_permission_code TEXT, priority TEXT NOT NULL,
  due_offset_hours INT, blocks_step_completion BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE TABLE IF NOT EXISTS workflow_instance_task_templates (
  id BIGSERIAL PRIMARY KEY, workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  source_task_template_id BIGINT REFERENCES workflow_task_templates(id) ON DELETE SET NULL,
  trigger_type TEXT NOT NULL, title_fa TEXT NOT NULL, description_fa TEXT NOT NULL DEFAULT '',
  assigned_role_id BIGINT REFERENCES roles(id), required_permission_code TEXT, priority TEXT NOT NULL,
  due_offset_hours INT, blocks_workflow_progress BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE TABLE IF NOT EXISTS workflow_instance_handoff_metrics (
  id BIGSERIAL PRIMARY KEY, workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  source_metric_definition_id BIGINT REFERENCES workflow_handoff_metric_definitions(id) ON DELETE SET NULL,
  metric_key TEXT NOT NULL, label_fa TEXT NOT NULL, unit_code TEXT NOT NULL,
  absolute_tolerance NUMERIC(18,4), percentage_tolerance NUMERIC(10,4), blocking_on_mismatch BOOLEAN NOT NULL DEFAULT FALSE,
  UNIQUE(workflow_instance_id, metric_key)
);
CREATE TABLE IF NOT EXISTS workflow_step_field_values (
  id BIGSERIAL PRIMARY KEY, workflow_step_instance_id UUID NOT NULL REFERENCES workflow_step_instances(id) ON DELETE CASCADE,
  field_definition_id BIGINT NOT NULL REFERENCES workflow_instance_field_definitions(id) ON DELETE CASCADE,
  field_key TEXT NOT NULL, field_type TEXT NOT NULL, value_json JSONB NOT NULL,
  entered_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL, entered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL, updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workflow_step_instance_id, field_key)
);

CREATE TABLE IF NOT EXISTS workflow_discrepancies (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  source_step_instance_id UUID REFERENCES workflow_step_instances(id), target_step_instance_id UUID REFERENCES workflow_step_instances(id),
  metric_key TEXT NOT NULL, expected_value NUMERIC(18,4), actual_value NUMERIC(18,4), difference_value NUMERIC(18,4),
  difference_percentage NUMERIC(18,4), unit_code TEXT, severity TEXT NOT NULL, is_blocking BOOLEAN NOT NULL DEFAULT FALSE,
  status TEXT NOT NULL DEFAULT 'OPEN', reported_by_user_id UUID REFERENCES users(id), reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  source_explanation TEXT, target_explanation TEXT, resolution_note TEXT, resolved_by_user_id UUID REFERENCES users(id),
  resolved_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_workflow_discrepancies_open ON workflow_discrepancies(workflow_instance_id,status,severity);

CREATE TABLE IF NOT EXISTS workflow_files (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  workflow_step_instance_id UUID NOT NULL REFERENCES workflow_step_instances(id) ON DELETE CASCADE,
  field_definition_id BIGINT REFERENCES workflow_instance_field_definitions(id) ON DELETE SET NULL,
  storage_key TEXT NOT NULL UNIQUE, original_file_name TEXT NOT NULL, mime_type TEXT NOT NULL, size_bytes BIGINT NOT NULL,
  customer_visible BOOLEAN NOT NULL DEFAULT FALSE, uploaded_by_user_id UUID REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE action_items
  ADD COLUMN IF NOT EXISTS deduplication_key TEXT,
  ADD COLUMN IF NOT EXISTS source_trigger_type TEXT,
  ADD COLUMN IF NOT EXISTS is_blocking BOOLEAN NOT NULL DEFAULT FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uq_action_items_deduplication_key ON action_items(deduplication_key) WHERE deduplication_key IS NOT NULL;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_template_version_status') THEN
    ALTER TABLE workflow_templates ADD CONSTRAINT chk_workflow_template_version_status CHECK (status IN ('DRAFT','PUBLISHED','ARCHIVED'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_field_definition_type') THEN
    ALTER TABLE workflow_step_field_definitions ADD CONSTRAINT chk_workflow_field_definition_type CHECK (field_type IN ('SHORT_TEXT','LONG_TEXT','INTEGER','DECIMAL','BOOLEAN','DATE','TIME','DATETIME','MONEY','SELECT','MULTI_SELECT','PHONE','ADDRESS','WEIGHT','AREA','VOLUME','QUANTITY','IMAGE','FILE','SIGNATURE'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_field_handoff_direction') THEN
    ALTER TABLE workflow_step_field_definitions ADD CONSTRAINT chk_workflow_field_handoff_direction CHECK (handoff_direction IS NULL OR handoff_direction IN ('IN','OUT'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_step_task_trigger') THEN
    ALTER TABLE workflow_step_task_templates ADD CONSTRAINT chk_workflow_step_task_trigger CHECK (trigger_type IN ('ON_STEP_OPEN','ON_STEP_START','ON_STEP_SUBMIT','ON_STEP_APPROVE','ON_STEP_COMPLETE'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_task_trigger') THEN
    ALTER TABLE workflow_task_templates ADD CONSTRAINT chk_workflow_task_trigger CHECK (trigger_type IN ('ON_WORKFLOW_START'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_discrepancy_status') THEN
    ALTER TABLE workflow_discrepancies ADD CONSTRAINT chk_workflow_discrepancy_status CHECK (status IN ('OPEN','UNDER_REVIEW','CORRECTION_REQUIRED','ACCEPTED','RESOLVED','CANCELLED'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='chk_workflow_discrepancy_severity') THEN
    ALTER TABLE workflow_discrepancies ADD CONSTRAINT chk_workflow_discrepancy_severity CHECK (severity IN ('INFO','WARNING','CRITICAL'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='fk_workflow_current_step') THEN
    ALTER TABLE workflow_instances ADD CONSTRAINT fk_workflow_current_step FOREIGN KEY(current_step_instance_id) REFERENCES workflow_step_instances(id) ON DELETE SET NULL;
  END IF;
END $$;

INSERT INTO permissions(code,name_fa,description_fa,group_code) VALUES
('workflow_fields.view_internal','مشاهده فیلدهای داخلی','مشاهده اطلاعات داخلی مراحل','WORKFLOW_FIELDS'),
('workflow_fields.view_sales','مشاهده فیلدهای فروش','مشاهده اطلاعات قابل اعلام توسط فروش','WORKFLOW_FIELDS'),
('workflow_fields.manage','مدیریت فیلدهای مراحل','طراحی فرم داینامیک مراحل','WORKFLOW_FIELDS'),
('workflow_templates.publish','انتشار گردش‌کار','انتشار نسخه Draft','WORKFLOWS'),
('workflow_templates.archive','بایگانی گردش‌کار','بایگانی نسخه منتشرشده','WORKFLOWS'),
('workflow_discrepancies.view_assigned','مشاهده مغایرت مرتبط','مشاهده مغایرت مراحل محول‌شده','DISCREPANCIES'),
('workflow_discrepancies.view_all','مشاهده همه مغایرت‌ها','مشاهده تمام مغایرت‌ها','DISCREPANCIES'),
('workflow_discrepancies.review','بررسی مغایرت','شروع بررسی مغایرت','DISCREPANCIES'),
('workflow_discrepancies.resolve','تعیین تکلیف مغایرت','پذیرش یا رفع مغایرت','DISCREPANCIES'),
('workflow_discrepancies.override','بازنویسی مغایرت','عبور مدیریتی از مغایرت','DISCREPANCIES'),
('workflow_steps.save_draft','ذخیره پیش‌نویس مرحله','ذخیره فرم بدون ارسال','STEPS'),
('workflow_steps.reopen','بازگشایی مرحله','بازگشایی مرحله تکمیل‌شده','STEPS'),
('workflow_steps.change_schedule','تغییر زمان‌بندی','ویرایش تاریخ تخمینی مرحله','STEPS'),
('workflow_files.upload','بارگذاری فایل مرحله','بارگذاری فایل و تصویر خصوصی','WORKFLOW_FILES'),
('workflow_files.view_internal','مشاهده فایل داخلی','دریافت فایل خصوصی عملیاتی','WORKFLOW_FILES'),
('workflow_files.view_customer','مشاهده فایل مشتری','دریافت فایل قابل نمایش مشتری','WORKFLOW_FILES')
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa,description_fa=EXCLUDED.description_fa,group_code=EXCLUDED.group_code;

INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.code='SUPER_ADMIN' OR (r.code='ADMIN' AND p.code IN (
  'workflow_fields.view_internal','workflow_fields.view_sales','workflow_fields.manage','workflow_templates.publish','workflow_templates.archive',
  'workflow_discrepancies.view_assigned','workflow_discrepancies.view_all','workflow_discrepancies.review','workflow_discrepancies.resolve','workflow_discrepancies.override',
  'workflow_steps.save_draft','workflow_steps.reopen','workflow_steps.change_schedule','workflow_files.upload','workflow_files.view_internal','workflow_files.view_customer'))
OR (r.code IN ('OPERATOR','SALES') AND p.code IN ('workflow_steps.view_assigned','workflow_steps.submit','workflow_steps.save_draft','workflow_fields.view_sales','workflow_files.upload','workflow_files.view_customer'))
OR (r.code='ACCOUNTANT' AND p.code IN ('workflow_steps.view_assigned','workflow_steps.submit','workflow_steps.save_draft','workflow_fields.view_internal','workflow_files.upload','workflow_files.view_internal'))
OR (r.code='SUPPLY' AND p.code IN ('workflow_steps.save_draft','workflow_fields.view_internal','workflow_discrepancies.view_assigned','workflow_files.upload','workflow_files.view_internal'))
OR (r.code IN ('DRIVER','INSTALLATION_LEAD') AND p.code IN ('workflow_steps.save_draft','workflow_fields.view_internal','workflow_files.upload','workflow_files.view_internal'))
OR (r.code='CUSTOMER' AND p.code='workflow_files.view_customer')
ON CONFLICT DO NOTHING;

INSERT INTO workflow_step_catalogue(step_code,title_fa,customer_title_fa,default_role_code,default_permission_code,default_duration_hours) VALUES
('ISSUE_PROFORMA','صدور پیش‌فاکتور','پیش‌فاکتور','OPERATOR','proformas.issue',8),
('CUSTOMER_CONFIRMATION','تأیید سفارش','تأیید سفارش','SALES','workflow_steps.submit',8),
('PAYMENT_CONFIRMATION','تأیید پرداخت','بررسی پرداخت','ACCOUNTANT','finance.payments.confirm',24),
('COLLECT_EXPORT_DEPOSIT','دریافت ۳۰٪ پیش‌پرداخت','دریافت پیش‌پرداخت','ACCOUNTANT','finance.payments.confirm',48),
('SOURCE_FROM_MINE','تأمین از معدن','تأمین سنگ','SUPPLY','operations.procurement.execute',120),
('SOURCE_FROM_BLOCK_WAREHOUSE','انتخاب کوپ از انبار','آماده‌سازی سنگ','SUPPLY','operations.procurement.execute',48),
('SOURCE_FROM_PRODUCT_WAREHOUSE','برداشت از انبار','آماده‌سازی سفارش','SUPPLY','operations.procurement.execute',24),
('SOURCE_FROM_FACTORY','تأمین از کارخانه','تأمین محصول','SUPPLY','operations.procurement.execute',96),
('SOURCE_FROM_SHOWROOM','رزرو یا برداشت از شوروم','آماده‌سازی سفارش','SUPPLY','operations.procurement.execute',24),
('LOAD_AT_SOURCE','بارگیری در مبدا','بارگیری سفارش','SUPPLY','operations.transport.execute',12),
('TRANSPORT_TO_FACTORY','حمل به کارخانه','حمل سفارش','DRIVER','operations.transport.execute',48),
('RECEIVE_AT_FACTORY','تحویل و تأیید کارخانه','تحویل در کارخانه','SUPPLY','operations.procurement.confirm_quantity',12),
('CUT_AND_PROCESS','برش و فراوری','فراوری سنگ','SUPPLY','operations.processing.execute',168),
('QUALITY_CONTROL','کنترل کیفیت','کنترل کیفیت','SUPPLY','operations.quality_control.execute',24),
('PACKAGING','بسته‌بندی','بسته‌بندی سفارش','SUPPLY','operations.packaging.execute',48),
('LOAD_FOR_DELIVERY','بارگیری برای ارسال','آماده‌سازی ارسال','SUPPLY','operations.transport.execute',12),
('DOMESTIC_TRANSPORT','حمل داخلی','ارسال سفارش','DRIVER','operations.transport.execute',72),
('DOMESTIC_DELIVERY','تحویل به مشتری','تحویل سفارش','DRIVER','operations.transport.confirm_delivery',12),
('TRANSPORT_TO_PORT','حمل به بندر','حمل به بندر','DRIVER','operations.transport.execute',96),
('CUSTOMS_PROCESS','فرایند گمرکی','فرایند صادرات','SUPPLY','operations.export.execute',120),
('PORT_LOADING','بارگیری بندری','بارگیری بندری','SUPPLY','operations.export.execute',48),
('PORT_HANDOFF','تحویل بندری','تحویل در بندر','SUPPLY','operations.export.execute',24),
('INSTALLATION','نصب','نصب سفارش','INSTALLATION_LEAD','operations.installation.execute',120),
('CUSTOMER_FINAL_ACCEPTANCE','تأیید نهایی مشتری','تأیید نهایی','SALES','workflow_steps.submit',24),
('ORDER_CLOSURE','بستن سفارش','تکمیل سفارش','ADMIN','workflow_steps.approve',8)
ON CONFLICT(step_code) DO UPDATE SET title_fa=EXCLUDED.title_fa,customer_title_fa=EXCLUDED.customer_title_fa,
default_role_code=EXCLUDED.default_role_code,default_permission_code=EXCLUDED.default_permission_code,default_duration_hours=EXCLUDED.default_duration_hours;

-- Keep Phase 1 rows as immutable legacy v1 and publish full v2 definitions.
INSERT INTO workflow_templates(template_group_code,version_number,code,name_fa,description_fa,icon_key,status,start_permission_code,is_active,created_from_template_id,published_at,sort_order)
SELECT template_group_code,2,template_group_code||'_v2',name_fa,description_fa,icon_key,'PUBLISHED',start_permission_code,is_active,id,NOW(),sort_order
FROM workflow_templates WHERE version_number=1
ON CONFLICT(template_group_code,version_number) DO NOTHING;

WITH sequences(group_code,seq,step_code,is_optional) AS (VALUES
('mine_to_domestic_factory',1,'ISSUE_PROFORMA',FALSE),('mine_to_domestic_factory',2,'CUSTOMER_CONFIRMATION',FALSE),('mine_to_domestic_factory',3,'PAYMENT_CONFIRMATION',FALSE),('mine_to_domestic_factory',4,'SOURCE_FROM_MINE',FALSE),('mine_to_domestic_factory',5,'LOAD_AT_SOURCE',FALSE),('mine_to_domestic_factory',6,'TRANSPORT_TO_FACTORY',FALSE),('mine_to_domestic_factory',7,'RECEIVE_AT_FACTORY',FALSE),('mine_to_domestic_factory',8,'ORDER_CLOSURE',FALSE),
('mine_to_domestic_dealer',1,'ISSUE_PROFORMA',FALSE),('mine_to_domestic_dealer',2,'CUSTOMER_CONFIRMATION',FALSE),('mine_to_domestic_dealer',3,'PAYMENT_CONFIRMATION',FALSE),('mine_to_domestic_dealer',4,'SOURCE_FROM_MINE',FALSE),('mine_to_domestic_dealer',5,'LOAD_AT_SOURCE',FALSE),('mine_to_domestic_dealer',6,'DOMESTIC_TRANSPORT',FALSE),('mine_to_domestic_dealer',7,'DOMESTIC_DELIVERY',FALSE),('mine_to_domestic_dealer',8,'ORDER_CLOSURE',FALSE),
('mine_to_export',1,'ISSUE_PROFORMA',FALSE),('mine_to_export',2,'CUSTOMER_CONFIRMATION',FALSE),('mine_to_export',3,'COLLECT_EXPORT_DEPOSIT',FALSE),('mine_to_export',4,'SOURCE_FROM_MINE',FALSE),('mine_to_export',5,'LOAD_AT_SOURCE',FALSE),('mine_to_export',6,'TRANSPORT_TO_PORT',FALSE),('mine_to_export',7,'CUSTOMS_PROCESS',FALSE),('mine_to_export',8,'PORT_HANDOFF',FALSE),('mine_to_export',9,'ORDER_CLOSURE',FALSE),
('block_warehouse_purchase',1,'ISSUE_PROFORMA',FALSE),('block_warehouse_purchase',2,'CUSTOMER_CONFIRMATION',FALSE),('block_warehouse_purchase',3,'PAYMENT_CONFIRMATION',FALSE),('block_warehouse_purchase',4,'SOURCE_FROM_BLOCK_WAREHOUSE',FALSE),('block_warehouse_purchase',5,'LOAD_AT_SOURCE',FALSE),('block_warehouse_purchase',6,'DOMESTIC_TRANSPORT',FALSE),('block_warehouse_purchase',7,'DOMESTIC_DELIVERY',FALSE),('block_warehouse_purchase',8,'ORDER_CLOSURE',FALSE),
('factory_to_export',1,'ISSUE_PROFORMA',FALSE),('factory_to_export',2,'CUSTOMER_CONFIRMATION',FALSE),('factory_to_export',3,'COLLECT_EXPORT_DEPOSIT',FALSE),('factory_to_export',4,'SOURCE_FROM_FACTORY',FALSE),('factory_to_export',5,'QUALITY_CONTROL',FALSE),('factory_to_export',6,'PACKAGING',FALSE),('factory_to_export',7,'LOAD_FOR_DELIVERY',FALSE),('factory_to_export',8,'TRANSPORT_TO_PORT',FALSE),('factory_to_export',9,'CUSTOMS_PROCESS',FALSE),('factory_to_export',10,'PORT_HANDOFF',FALSE),('factory_to_export',11,'ORDER_CLOSURE',FALSE),
('factory_to_dealer',1,'ISSUE_PROFORMA',FALSE),('factory_to_dealer',2,'CUSTOMER_CONFIRMATION',FALSE),('factory_to_dealer',3,'PAYMENT_CONFIRMATION',FALSE),('factory_to_dealer',4,'SOURCE_FROM_FACTORY',FALSE),('factory_to_dealer',5,'QUALITY_CONTROL',FALSE),('factory_to_dealer',6,'PACKAGING',FALSE),('factory_to_dealer',7,'LOAD_FOR_DELIVERY',FALSE),('factory_to_dealer',8,'DOMESTIC_TRANSPORT',FALSE),('factory_to_dealer',9,'DOMESTIC_DELIVERY',FALSE),('factory_to_dealer',10,'ORDER_CLOSURE',FALSE),
('factory_to_project',1,'ISSUE_PROFORMA',FALSE),('factory_to_project',2,'CUSTOMER_CONFIRMATION',FALSE),('factory_to_project',3,'PAYMENT_CONFIRMATION',FALSE),('factory_to_project',4,'SOURCE_FROM_FACTORY',FALSE),('factory_to_project',5,'CUT_AND_PROCESS',FALSE),('factory_to_project',6,'QUALITY_CONTROL',FALSE),('factory_to_project',7,'PACKAGING',FALSE),('factory_to_project',8,'DOMESTIC_TRANSPORT',FALSE),('factory_to_project',9,'DOMESTIC_DELIVERY',FALSE),('factory_to_project',10,'INSTALLATION',TRUE),('factory_to_project',11,'CUSTOMER_FINAL_ACCEPTANCE',FALSE),('factory_to_project',12,'ORDER_CLOSURE',FALSE),
('showroom_purchase',1,'ISSUE_PROFORMA',FALSE),('showroom_purchase',2,'CUSTOMER_CONFIRMATION',FALSE),('showroom_purchase',3,'PAYMENT_CONFIRMATION',FALSE),('showroom_purchase',4,'SOURCE_FROM_SHOWROOM',FALSE),('showroom_purchase',5,'PACKAGING',TRUE),('showroom_purchase',6,'DOMESTIC_TRANSPORT',TRUE),('showroom_purchase',7,'DOMESTIC_DELIVERY',FALSE),('showroom_purchase',8,'ORDER_CLOSURE',FALSE),
('warehouse_purchase',1,'ISSUE_PROFORMA',FALSE),('warehouse_purchase',2,'CUSTOMER_CONFIRMATION',FALSE),('warehouse_purchase',3,'PAYMENT_CONFIRMATION',FALSE),('warehouse_purchase',4,'SOURCE_FROM_PRODUCT_WAREHOUSE',FALSE),('warehouse_purchase',5,'QUALITY_CONTROL',FALSE),('warehouse_purchase',6,'PACKAGING',FALSE),('warehouse_purchase',7,'LOAD_FOR_DELIVERY',FALSE),('warehouse_purchase',8,'DOMESTIC_TRANSPORT',FALSE),('warehouse_purchase',9,'DOMESTIC_DELIVERY',FALSE),('warehouse_purchase',10,'ORDER_CLOSURE',FALSE)
)
INSERT INTO workflow_template_steps(workflow_template_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,responsible_role_code,responsible_role_id,required_permission_code,customer_visible,is_first_step,is_optional,is_skippable,is_active,default_duration_hours,starts_automatically)
SELECT wt.id,c.step_code,c.title_fa,c.description_fa,c.customer_title_fa,'',s.seq,c.default_role_code,r.id,c.default_permission_code,TRUE,(s.seq=1),s.is_optional,s.is_optional,TRUE,c.default_duration_hours,FALSE
FROM sequences s JOIN workflow_templates wt ON wt.template_group_code=s.group_code AND wt.version_number=2
JOIN workflow_step_catalogue c ON c.step_code=s.step_code LEFT JOIN roles r ON r.code=c.default_role_code
ON CONFLICT(workflow_template_id,step_code) DO NOTHING;

INSERT INTO workflow_handoff_metric_definitions(workflow_template_id,metric_key,label_fa,unit_code,absolute_tolerance,percentage_tolerance,blocking_on_mismatch)
SELECT id,'cargo_weight','وزن بار','TON',0.2,1,TRUE FROM workflow_templates WHERE version_number=2
ON CONFLICT(workflow_template_id,metric_key) DO NOTHING;

INSERT INTO workflow_step_field_definitions(workflow_template_step_id,field_key,label_fa,field_type,is_required,is_customer_visible,is_sales_visible,unit_code,sort_order,handoff_metric_key,handoff_direction,validation_json)
SELECT s.id,CASE WHEN s.step_code IN ('LOAD_AT_SOURCE','LOAD_FOR_DELIVERY') THEN 'cargo_weight_out' ELSE 'cargo_weight_in' END,
CASE WHEN s.step_code IN ('LOAD_AT_SOURCE','LOAD_FOR_DELIVERY') THEN 'وزن بار خروجی' ELSE 'وزن بار تحویل‌گرفته‌شده' END,
'WEIGHT',FALSE,FALSE,TRUE,'TON',10,'cargo_weight',CASE WHEN s.step_code IN ('LOAD_AT_SOURCE','LOAD_FOR_DELIVERY') THEN 'OUT' ELSE 'IN' END,'{"min":0,"max":1000}'::jsonb
FROM workflow_template_steps s JOIN workflow_templates wt ON wt.id=s.workflow_template_id
WHERE wt.version_number=2 AND s.step_code IN ('LOAD_AT_SOURCE','LOAD_FOR_DELIVERY','RECEIVE_AT_FACTORY','DOMESTIC_DELIVERY','PORT_HANDOFF')
ON CONFLICT(workflow_template_step_id,field_key) DO NOTHING;

INSERT INTO workflow_step_task_templates(workflow_template_step_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_step_completion)
SELECT s.id,'ON_STEP_START','تماس با مشتری و اعلام شروع بارگیری','وضعیت بارگیری و زمان تقریبی ارسال به مشتری اعلام شود',r.id,'sales.customer_contact.log','HIGH',8,FALSE
FROM workflow_template_steps s JOIN workflow_templates wt ON wt.id=s.workflow_template_id JOIN roles r ON r.code='SALES'
WHERE wt.version_number=2 AND s.step_code IN ('LOAD_AT_SOURCE','LOAD_FOR_DELIVERY')
AND NOT EXISTS(SELECT 1 FROM workflow_step_task_templates t WHERE t.workflow_template_step_id=s.id AND t.trigger_type='ON_STEP_START' AND t.title_fa='تماس با مشتری و اعلام شروع بارگیری');
