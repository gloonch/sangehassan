-- Supplier, purchase, quality-control, installation and order closure capabilities.
-- Everything in this migration is optional for legacy orders and workflow instances.

CREATE TABLE IF NOT EXISTS suppliers (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  supplier_code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  supplier_type TEXT NOT NULL,
  phone TEXT,
  secondary_phone TEXT,
  contact_name TEXT,
  address TEXT,
  city TEXT,
  province TEXT,
  country_code CHAR(2) NOT NULL DEFAULT 'IR',
  notes TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(supplier_type IN ('MINE','FACTORY','WORKSHOP','WAREHOUSE','STONE_DEALER','TRANSPORT_COMPANY','INSTALLATION_CONTRACTOR','PACKAGING_CONTRACTOR','OTHER'))
);
CREATE INDEX IF NOT EXISTS idx_suppliers_active_type ON suppliers(is_active,supplier_type,name);

ALTER TABLE fulfillment_batches ADD COLUMN IF NOT EXISTS supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL;
ALTER TABLE inventory_lots ADD COLUMN IF NOT EXISTS supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL;
ALTER TABLE operational_cost_entries ADD COLUMN IF NOT EXISTS supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL;
ALTER TABLE inventory_lot_conversions ADD COLUMN IF NOT EXISTS supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL;
ALTER TABLE packaging_units ADD COLUMN IF NOT EXISTS supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS purchase_records (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  purchase_number TEXT NOT NULL UNIQUE,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
  batch_id UUID REFERENCES fulfillment_batches(id) ON DELETE SET NULL,
  supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE RESTRICT,
  assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  assigned_role_id BIGINT REFERENCES roles(id) ON DELETE SET NULL,
  stone_name TEXT NOT NULL,
  description TEXT,
  quantity NUMERIC(18,4) NOT NULL,
  received_quantity NUMERIC(18,4) NOT NULL DEFAULT 0,
  quantity_unit TEXT NOT NULL,
  unit_price NUMERIC(18,4) NOT NULL DEFAULT 0,
  total_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
  currency_code CHAR(3) NOT NULL DEFAULT 'IRR' REFERENCES currencies(code),
  status TEXT NOT NULL DEFAULT 'DRAFT',
  expected_at TIMESTAMPTZ,
  received_at TIMESTAMPTZ,
  notes TEXT,
  cost_entry_id UUID UNIQUE REFERENCES operational_cost_entries(id) ON DELETE SET NULL,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  cancelled_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  cancelled_at TIMESTAMPTZ,
  cancellation_reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(quantity>0),
  CHECK(received_quantity>=0 AND received_quantity<=quantity),
  CHECK(unit_price>=0 AND total_amount>=0),
  CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER')),
  CHECK(status IN ('DRAFT','CONFIRMED','PARTIALLY_RECEIVED','RECEIVED','CANCELLED'))
);
CREATE INDEX IF NOT EXISTS idx_purchase_records_order ON purchase_records(order_id,status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_purchase_records_supplier ON purchase_records(supplier_id,status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_purchase_records_assignee ON purchase_records(assigned_user_id,assigned_role_id,status);

CREATE TABLE IF NOT EXISTS purchase_receipts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  purchase_record_id UUID NOT NULL REFERENCES purchase_records(id) ON DELETE RESTRICT,
  quantity NUMERIC(18,4) NOT NULL,
  quantity_unit TEXT NOT NULL,
  inventory_lot_id UUID REFERENCES inventory_lots(id) ON DELETE SET NULL,
  inventory_movement_id UUID REFERENCES inventory_movements(id) ON DELETE SET NULL,
  notes TEXT,
  received_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(quantity>0),
  CHECK(quantity_unit IN ('TON','KILOGRAM','CUBIC_METER','SQUARE_METER','BLOCK','SLAB','TILE','PIECE','PACKAGE','BUNDLE','CONTAINER'))
);
CREATE INDEX IF NOT EXISTS idx_purchase_receipts_purchase ON purchase_receipts(purchase_record_id,received_at DESC);

CREATE TABLE IF NOT EXISTS quality_inspections (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  inspection_number TEXT NOT NULL UNIQUE,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
  batch_id UUID REFERENCES fulfillment_batches(id) ON DELETE SET NULL,
  inventory_lot_id UUID REFERENCES inventory_lots(id) ON DELETE SET NULL,
  workflow_step_instance_id UUID REFERENCES workflow_step_instances(id) ON DELETE SET NULL,
  inspection_type TEXT NOT NULL DEFAULT 'GENERAL',
  status TEXT NOT NULL DEFAULT 'PENDING',
  assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  assigned_role_id BIGINT REFERENCES roles(id) ON DELETE SET NULL,
  inspected_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  inspected_at TIMESTAMPTZ,
  notes TEXT,
  decision_reason TEXT,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(inspection_type IN ('GENERAL','INCOMING','IN_PROCESS','FINAL','PACKAGING','INSTALLATION','OTHER')),
  CHECK(status IN ('PENDING','PASSED','PASSED_WITH_NOTES','FAILED','REWORK_REQUIRED'))
);
CREATE INDEX IF NOT EXISTS idx_quality_inspections_order ON quality_inspections(order_id,status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_quality_inspections_batch ON quality_inspections(batch_id,status);
CREATE INDEX IF NOT EXISTS idx_quality_inspections_assignee ON quality_inspections(assigned_user_id,assigned_role_id,status);

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS installation_required BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS closed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS closed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS closure_forced BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS closure_reason TEXT;

CREATE TABLE IF NOT EXISTS installation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  installation_number TEXT NOT NULL UNIQUE,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
  customer_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  project_name TEXT,
  project_address TEXT,
  contact_name TEXT,
  contact_phone TEXT,
  status TEXT NOT NULL DEFAULT 'DRAFT',
  installation_lead_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,
  planned_start_at TIMESTAMPTZ,
  actual_start_at TIMESTAMPTZ,
  estimated_end_at TIMESTAMPTZ,
  actual_end_at TIMESTAMPTZ,
  planned_area NUMERIC(18,4),
  installed_area NUMERIC(18,4) NOT NULL DEFAULT 0,
  area_unit TEXT NOT NULL DEFAULT 'SQUARE_METER',
  notes TEXT,
  workflow_instance_id UUID UNIQUE REFERENCES workflow_instances(id) ON DELETE SET NULL,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  cancelled_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  cancelled_at TIMESTAMPTZ,
  cancellation_reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(status IN ('DRAFT','PLANNED','READY','IN_PROGRESS','PAUSED','COMPLETED','CANCELLED')),
  CHECK(planned_area IS NULL OR planned_area>=0),
  CHECK(installed_area>=0),
  CHECK(area_unit IN ('SQUARE_METER','PIECE','SLAB','TILE'))
);
CREATE INDEX IF NOT EXISTS idx_installation_jobs_order ON installation_jobs(order_id,status);
CREATE INDEX IF NOT EXISTS idx_installation_jobs_lead ON installation_jobs(installation_lead_user_id,status);

ALTER TABLE operational_cost_entries
  ADD COLUMN IF NOT EXISTS installation_id UUID REFERENCES installation_jobs(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_operational_cost_installation ON operational_cost_entries(installation_id,status);

CREATE TABLE IF NOT EXISTS installation_job_members (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  installation_job_id UUID NOT NULL REFERENCES installation_jobs(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  name_override TEXT,
  role_label TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(user_id IS NOT NULL OR NULLIF(TRIM(name_override),'') IS NOT NULL)
);
CREATE INDEX IF NOT EXISTS idx_installation_members_job ON installation_job_members(installation_job_id);

CREATE TABLE IF NOT EXISTS installation_updates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  installation_job_id UUID NOT NULL REFERENCES installation_jobs(id) ON DELETE CASCADE,
  update_date DATE NOT NULL DEFAULT CURRENT_DATE,
  installed_quantity NUMERIC(18,4) NOT NULL DEFAULT 0,
  quantity_unit TEXT NOT NULL DEFAULT 'SQUARE_METER',
  status TEXT NOT NULL DEFAULT 'PROGRESS',
  description TEXT,
  customer_visible BOOLEAN NOT NULL DEFAULT FALSE,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(installed_quantity>=0),
  CHECK(quantity_unit IN ('SQUARE_METER','PIECE','SLAB','TILE')),
  CHECK(status IN ('PROGRESS','DELAY','PAUSED','RESUMED','NOTE'))
);
CREATE INDEX IF NOT EXISTS idx_installation_updates_job ON installation_updates(installation_job_id,update_date DESC,created_at DESC);

CREATE TABLE IF NOT EXISTS installation_update_activities (
  installation_update_id UUID NOT NULL REFERENCES installation_updates(id) ON DELETE CASCADE,
  activity_type TEXT NOT NULL,
  PRIMARY KEY(installation_update_id,activity_type),
  CHECK(activity_type IN ('SUBSTRATE_PREPARATION','CUTTING','INSTALLATION','RESIN','POLISHING','GROUTING','ANCHORING','BASE_INSTALLATION','REPAIR','CLEANUP','OTHER'))
);

CREATE TABLE IF NOT EXISTS installation_material_usage (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  installation_job_id UUID NOT NULL REFERENCES installation_jobs(id) ON DELETE CASCADE,
  material_name TEXT NOT NULL,
  quantity NUMERIC(18,4) NOT NULL,
  unit TEXT NOT NULL,
  cost_entry_id UUID REFERENCES operational_cost_entries(id) ON DELETE SET NULL,
  notes TEXT,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK(quantity>0)
);

CREATE TABLE IF NOT EXISTS installation_issues (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  installation_job_id UUID NOT NULL REFERENCES installation_jobs(id) ON DELETE CASCADE,
  issue_type TEXT NOT NULL,
  severity TEXT NOT NULL DEFAULT 'WARNING',
  description TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'OPEN',
  customer_visible BOOLEAN NOT NULL DEFAULT FALSE,
  reported_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  resolved_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  resolution_note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMPTZ,
  CHECK(issue_type IN ('STONE_DAMAGE','WRONG_DIMENSION','MISSING_MATERIAL','SITE_NOT_READY','CUSTOMER_CHANGE','SUBSTRATE_PROBLEM','OTHER')),
  CHECK(severity IN ('INFO','WARNING','CRITICAL')),
  CHECK(status IN ('OPEN','RESOLVED','CANCELLED'))
);
CREATE INDEX IF NOT EXISTS idx_installation_issues_job ON installation_issues(installation_job_id,status,severity);

CREATE TABLE IF NOT EXISTS customer_order_acceptances (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
  installation_job_id UUID REFERENCES installation_jobs(id) ON DELETE SET NULL,
  shipment_id UUID REFERENCES shipments(id) ON DELETE SET NULL,
  workflow_step_instance_id UUID REFERENCES workflow_step_instances(id) ON DELETE SET NULL,
  customer_name TEXT NOT NULL,
  accepted BOOLEAN NOT NULL,
  comment TEXT,
  signature_file_id UUID REFERENCES workflow_files(id) ON DELETE SET NULL,
  photo_file_id UUID REFERENCES workflow_files(id) ON DELETE SET NULL,
  recorded_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_customer_acceptances_order ON customer_order_acceptances(order_id,accepted_at DESC);

ALTER TABLE workflow_files ALTER COLUMN workflow_instance_id DROP NOT NULL;

-- Closed orders remain commercially valid. Replacing the Phase 4 trigger keeps
-- both future updates and the existing read model consistent without editing 017.
CREATE OR REPLACE FUNCTION initialize_order_financial_summary() RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO order_financial_summaries(order_id,currency,revenue_amount,outstanding_amount)
  SELECT NEW.order_id,NEW.currency,
    CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED') THEN NEW.final_customer_amount ELSE 0 END,
    CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED') THEN NEW.final_customer_amount ELSE 0 END
  FROM orders o WHERE o.id=NEW.order_id
  ON CONFLICT(order_id) DO UPDATE SET
    currency=EXCLUDED.currency,
    revenue_amount=EXCLUDED.revenue_amount,
    outstanding_amount=GREATEST(0,EXCLUDED.revenue_amount-order_financial_summaries.confirmed_payment_amount+order_financial_summaries.refunded_amount),
    updated_at=NOW();
  RETURN NEW;
END $$ LANGUAGE plpgsql;

UPDATE order_financial_summaries s SET
  revenue_amount=t.final_customer_amount,
  outstanding_amount=GREATEST(0,t.final_customer_amount-s.confirmed_payment_amount+s.refunded_amount),
  updated_at=NOW()
FROM orders o JOIN order_commercial_terms t ON t.order_id=o.id
WHERE s.order_id=o.id AND o.status='CLOSED';

DO $$ BEGIN
  ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
  ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_status_phase5;
  ALTER TABLE orders ADD CONSTRAINT chk_orders_status_phase5 CHECK(status IN ('DRAFT','PROFORMA_ISSUED','CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED','CANCELLED'));

  ALTER TABLE fulfillment_batches DROP CONSTRAINT IF EXISTS chk_batch_status;
  ALTER TABLE fulfillment_batches ADD CONSTRAINT chk_batch_status CHECK(status IN ('DRAFT','PLANNED','RESERVING_STOCK','STOCK_RESERVED','IN_PRODUCTION','READY_FOR_QC','QC_APPROVED','QC_REJECTED','READY_FOR_PACKAGING','PACKAGED','READY_FOR_SHIPMENT','PARTIALLY_SHIPPED','SHIPPED','PARTIALLY_DELIVERED','DELIVERED','BLOCKED','NEEDS_CORRECTION','SPLIT','MERGED','CANCELLED'));

  ALTER TABLE workflow_step_field_definitions DROP CONSTRAINT IF EXISTS chk_workflow_field_definition_type;
  ALTER TABLE workflow_step_field_definitions ADD CONSTRAINT chk_workflow_field_definition_type CHECK(field_type IN ('SHORT_TEXT','LONG_TEXT','INTEGER','DECIMAL','BOOLEAN','DATE','TIME','DATETIME','MONEY','SELECT','MULTI_SELECT','PHONE','ADDRESS','WEIGHT','AREA','VOLUME','QUANTITY','IMAGE','FILE','SIGNATURE','QC_CHECK'));
END $$;

INSERT INTO permissions(code,name_fa,description_fa,group_code) VALUES
('suppliers.view','مشاهده تأمین‌کنندگان','مشاهده اطلاعات عمومی تأمین‌کنندگان','SUPPLIERS'),
('suppliers.create','ایجاد تأمین‌کننده','ثبت تأمین‌کننده جدید','SUPPLIERS'),
('suppliers.update','ویرایش تأمین‌کننده','ویرایش اطلاعات تأمین‌کننده','SUPPLIERS'),
('suppliers.disable','غیرفعال‌سازی تأمین‌کننده','جلوگیری از استفاده جدید از تأمین‌کننده','SUPPLIERS'),
('purchases.view_assigned','مشاهده خریدهای مرتبط','مشاهده خریدهای تخصیص‌یافته','PURCHASES'),
('purchases.view_all','مشاهده همه خریدها','مشاهده همه خریدهای عملیاتی','PURCHASES'),
('purchases.create','ایجاد خرید','ثبت Purchase Record','PURCHASES'),
('purchases.update','ویرایش خرید','ویرایش Purchase پیش از دریافت','PURCHASES'),
('purchases.confirm','تأیید خرید','تأیید Purchase Record','PURCHASES'),
('purchases.receive','دریافت خرید','ثبت دریافت جزئی یا کامل خرید','PURCHASES'),
('purchases.cancel','لغو خرید','لغو Purchase Record با دلیل','PURCHASES'),
('quality.view_assigned','مشاهده QC مرتبط','مشاهده Inspection تخصیص‌یافته','QUALITY'),
('quality.view_all','مشاهده همه QCها','مشاهده تمام Inspectionها','QUALITY'),
('quality.inspect','اجرای QC','ثبت Checklist و نتیجه اولیه','QUALITY'),
('quality.accept','پذیرش QC','پذیرش نتیجه دارای توضیح','QUALITY'),
('quality.reject','رد QC','رد Batch یا Lot پس از QC','QUALITY'),
('quality.override','بازنویسی QC','عبور مدیریتی از نتیجه QC با دلیل','QUALITY'),
('installation.view_assigned','مشاهده نصب مرتبط','مشاهده پروژه نصب تخصیص‌یافته','INSTALLATION'),
('installation.view_all','مشاهده همه نصب‌ها','مشاهده همه پروژه‌های نصب','INSTALLATION'),
('installation.create','ایجاد نصب','ایجاد Installation Job','INSTALLATION'),
('installation.update','ویرایش نصب','ویرایش برنامه و تیم نصب','INSTALLATION'),
('installation.start','شروع نصب','شروع یا ادامه عملیات نصب','INSTALLATION'),
('installation.progress','ثبت پیشرفت نصب','ثبت Update و عکس نصب','INSTALLATION'),
('installation.complete','تکمیل نصب','ثبت پایان عملیات نصب','INSTALLATION'),
('installation.cancel','لغو نصب','لغو Installation Job با دلیل','INSTALLATION'),
('installation.override','بازنویسی نصب','عبور مدیریتی از هشدارهای نصب','INSTALLATION'),
('customer_acceptance.record','ثبت تأیید مشتری','ثبت تأیید عملیاتی مشتری','ORDERS'),
('customer_portal.acceptance.confirm_own','تأیید سفارش خود','ثبت تأیید نهایی سفارش متعلق به مشتری','CUSTOMER_PORTAL'),
('customer_portal.installation.view_own','مشاهده نصب خود','مشاهده پیشرفت نصب سفارش متعلق به مشتری','CUSTOMER_PORTAL'),
('orders.close','بستن سفارش','بستن سفارش آماده','ORDERS'),
('orders.close_with_warnings','بستن سفار‌ش دارای هشدار','بستن مدیریتی سفارش با دلیل','ORDERS'),
('workflow_start.installation','شروع Workflow نصب','شروع گردش مستقل نصب','WORKFLOW_START')
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa,description_fa=EXCLUDED.description_fa,group_code=EXCLUDED.group_code;

INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE
 r.code='SUPER_ADMIN'
 OR r.code='ADMIN' AND (p.code LIKE ANY(ARRAY['suppliers.%','purchases.%','quality.%','installation.%','customer_acceptance.%','orders.close%','workflow_start.installation']))
 OR r.code='SUPPLY' AND p.code IN ('suppliers.view','purchases.view_assigned','purchases.create','purchases.update','purchases.receive','quality.view_assigned','quality.inspect')
 OR r.code='ACCOUNTANT' AND p.code IN ('suppliers.view','purchases.view_all')
 OR r.code='SALES' AND p.code IN ('suppliers.view','installation.view_assigned','customer_acceptance.record')
 OR r.code='INSTALLATION_LEAD' AND p.code IN ('installation.view_assigned','installation.start','installation.progress','installation.complete','quality.view_assigned')
 OR r.code='CUSTOMER' AND p.code IN ('customer_portal.acceptance.confirm_own','customer_portal.installation.view_own')
ON CONFLICT DO NOTHING;

INSERT INTO workflow_step_catalogue(step_code,title_fa,customer_title_fa,description_fa,default_role_code,default_permission_code,default_duration_hours) VALUES
('REWORK_BATCH','اصلاح بچ','اصلاح محصول','اصلاح موارد ردشده در کنترل کیفیت','SUPPLY','quality.inspect',48),
('QC_REJECTION_REVIEW','تعیین تکلیف بچ ردشده','بررسی محصول','بررسی مدیریتی بچ ردشده بدون لغو سفارش','ADMIN','quality.reject',8),
('PLAN_INSTALLATION','برنامه‌ریزی نصب','برنامه‌ریزی نصب','تعیین تیم، محل و زمان نصب','INSTALLATION_LEAD','installation.update',12),
('START_INSTALLATION','شروع نصب','شروع عملیات نصب','ثبت شروع حضور در پروژه','INSTALLATION_LEAD','installation.start',4),
('EXECUTE_INSTALLATION','اجرای نصب','عملیات نصب','ثبت پیشرفت و تصاویر نصب','INSTALLATION_LEAD','installation.progress',120),
('COMPLETE_INSTALLATION','تکمیل نصب','تکمیل نصب','ثبت پایان عملیات نصب','INSTALLATION_LEAD','installation.complete',8)
ON CONFLICT(step_code) DO UPDATE SET title_fa=EXCLUDED.title_fa,customer_title_fa=EXCLUDED.customer_title_fa,description_fa=EXCLUDED.description_fa,default_role_code=EXCLUDED.default_role_code,default_permission_code=EXCLUDED.default_permission_code,default_duration_hours=EXCLUDED.default_duration_hours;

-- A new published batch version is used only for future snapshots.
INSERT INTO workflow_templates(template_group_code,version_number,code,name_fa,description_fa,icon_key,start_permission_code,is_active,status,sort_order,scope_type,published_at,max_iterations,created_from_template_id)
SELECT template_group_code,2,'batch_fulfillment_v2',name_fa,'تأمین، تولید، کنترل کیفیت پویا و اصلاح کنترل‌شده',icon_key,start_permission_code,is_active,'PUBLISHED',sort_order,scope_type,NOW(),20,id
FROM workflow_templates WHERE code='batch_fulfillment_v1'
ON CONFLICT(template_group_code,version_number) DO NOTHING;

INSERT INTO workflow_template_steps(workflow_template_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,responsible_role_code,responsible_role_id,required_permission_code,customer_visible,is_first_step,is_entry,is_active,default_duration_hours,starts_automatically,domain_event_code,is_optional,is_skippable,requires_approval,approval_role_id)
SELECT n.id,s.step_code,s.internal_title_fa,s.internal_description_fa,s.customer_title_fa,s.customer_description_fa,
  CASE WHEN s.sequence_number>=7 THEN s.sequence_number+1 ELSE s.sequence_number END,
  s.responsible_role_code,s.responsible_role_id,CASE WHEN s.step_code='QC_BATCH' THEN 'quality.inspect' ELSE s.required_permission_code END,
  s.customer_visible,s.is_first_step,s.is_entry,s.is_active,s.default_duration_hours,s.starts_automatically,s.domain_event_code,s.is_optional,s.is_skippable,s.requires_approval,s.approval_role_id
FROM workflow_templates n JOIN workflow_templates o ON o.code='batch_fulfillment_v1'
JOIN workflow_template_steps s ON s.workflow_template_id=o.id
WHERE n.code='batch_fulfillment_v2'
ON CONFLICT(workflow_template_id,step_code) DO NOTHING;

INSERT INTO workflow_template_steps(workflow_template_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,responsible_role_code,responsible_role_id,required_permission_code,customer_visible,is_first_step,is_entry,is_active,default_duration_hours,starts_automatically,is_optional,is_skippable)
SELECT wt.id,x.code,x.title,x.description,x.customer_title,x.customer_description,x.seq,x.role,r.id,x.permission,x.visible,FALSE,FALSE,TRUE,x.hours,FALSE,FALSE,FALSE
FROM workflow_templates wt
JOIN (VALUES
 ('REWORK_BATCH','اصلاح بچ','اصلاح موارد اعلام‌شده در QC','اصلاح محصول','محصول در حال اصلاح است',7,'SUPPLY','quality.inspect',TRUE,48),
 ('QC_REJECTION_REVIEW','تعیین تکلیف بچ ردشده','بررسی مدیریتی نتیجه رد QC','بررسی نتیجه کنترل کیفیت','نتیجه کنترل کیفیت در حال بررسی است',10,'ADMIN','quality.reject',FALSE,8)
) AS x(code,title,description,customer_title,customer_description,seq,role,permission,visible,hours) ON TRUE
JOIN roles r ON r.code=x.role
WHERE wt.code='batch_fulfillment_v2'
ON CONFLICT(workflow_template_id,step_code) DO NOTHING;

INSERT INTO workflow_step_field_definitions(workflow_template_step_id,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,validation_json,sort_order)
SELECT s.id,x.key,x.label,'نتیجه، توضیح، مقدار اندازه‌گیری و تصویر اختیاری','QC_CHECK',FALSE,FALSE,FALSE,'{"allowedResults":["PASS","FAIL","NOT_APPLICABLE"]}'::jsonb,x.sort
FROM workflow_templates wt JOIN workflow_template_steps s ON s.workflow_template_id=wt.id AND s.step_code='QC_BATCH'
JOIN (VALUES
 ('dimensions','ابعاد',1),('thickness','ضخامت',2),('color','رنگ',3),('sorting','سورت',4),('surface_quality','کیفیت سطح',5),
 ('edge_damage','لب‌پریدگی',6),('cracks','ترک',7),('breakage','شکستگی',8),('uniformity','یکنواختی',9),('finish_type','نوع Finish',10),
 ('polish_quality','کیفیت ساب',11),('resin_quality','کیفیت رزین',12),('count','تعداد',13),('area','متراژ',14),('packaging','بسته‌بندی',15)
) AS x(key,label,sort) ON TRUE
WHERE wt.code='batch_fulfillment_v2'
ON CONFLICT(workflow_template_step_id,field_key) DO NOTHING;

INSERT INTO workflow_step_transitions(workflow_template_id,source_step_id,target_step_id,transition_code,label_fa,transition_type,result_code,is_default,requires_permission_code,requires_reason,sort_order)
SELECT n.id,ns.id,nt.id,t.transition_code,t.label_fa,t.transition_type,t.result_code,t.is_default,t.requires_permission_code,t.requires_reason,t.sort_order
FROM workflow_templates n JOIN workflow_templates o ON o.code='batch_fulfillment_v1'
JOIN workflow_step_transitions t ON t.workflow_template_id=o.id
JOIN workflow_template_steps os ON os.id=t.source_step_id
JOIN workflow_template_steps ot ON ot.id=t.target_step_id
JOIN workflow_template_steps ns ON ns.workflow_template_id=n.id AND ns.step_code=os.step_code
JOIN workflow_template_steps nt ON nt.workflow_template_id=n.id AND nt.step_code=ot.step_code
WHERE n.code='batch_fulfillment_v2' AND os.step_code<>'QC_BATCH'
ON CONFLICT(workflow_template_id,source_step_id,transition_code) DO NOTHING;

INSERT INTO workflow_step_transitions(workflow_template_id,source_step_id,target_step_id,transition_code,label_fa,transition_type,result_code,is_default,requires_permission_code,requires_reason,sort_order)
SELECT wt.id,s.id,t.id,x.code,x.label,'RESULT_BASED',x.result,x.is_default,x.permission,x.requires_reason,x.sort
FROM workflow_templates wt
JOIN (VALUES
 ('QC_BATCH','PACKAGE_BATCH','QC_APPROVED','تأیید و ارسال به بسته‌بندی','APPROVED',TRUE,'quality.inspect',FALSE,1),
 ('QC_BATCH','REWORK_BATCH','QC_REWORK','ارسال برای اصلاح','CORRECTION_REQUIRED',FALSE,'quality.inspect',FALSE,2),
 ('QC_BATCH','QC_REJECTION_REVIEW','QC_REJECTED','رد بچ','REJECTED',FALSE,'quality.reject',FALSE,3)
) AS x(source,target,code,label,result,is_default,permission,requires_reason,sort) ON TRUE
JOIN workflow_template_steps s ON s.workflow_template_id=wt.id AND s.step_code=x.source
JOIN workflow_template_steps t ON t.workflow_template_id=wt.id AND t.step_code=x.target
WHERE wt.code='batch_fulfillment_v2'
ON CONFLICT(workflow_template_id,source_step_id,transition_code) DO NOTHING;

INSERT INTO workflow_step_transitions(workflow_template_id,source_step_id,target_step_id,transition_code,label_fa,transition_type,is_default,requires_permission_code,requires_reason,sort_order)
SELECT wt.id,s.id,t.id,'REWORK_TO_QC','کنترل کیفیت مجدد','AUTOMATIC',TRUE,'quality.inspect',FALSE,1
FROM workflow_templates wt
JOIN workflow_template_steps s ON s.workflow_template_id=wt.id AND s.step_code='REWORK_BATCH'
JOIN workflow_template_steps t ON t.workflow_template_id=wt.id AND t.step_code='QC_BATCH'
WHERE wt.code='batch_fulfillment_v2'
ON CONFLICT(workflow_template_id,source_step_id,transition_code) DO NOTHING;

INSERT INTO workflow_templates(template_group_code,version_number,code,name_fa,description_fa,icon_key,start_permission_code,is_active,status,sort_order,scope_type,published_at,max_iterations)
VALUES('installation_execution',1,'installation_execution_v1','اجرای نصب','برنامه‌ریزی، اجرا، تکمیل و تأیید اختیاری نصب','project','workflow_start.installation',TRUE,'PUBLISHED',130,'INSTALLATION',NOW(),10)
ON CONFLICT(template_group_code,version_number) DO NOTHING;

INSERT INTO workflow_template_steps(workflow_template_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,responsible_role_code,responsible_role_id,required_permission_code,customer_visible,is_first_step,is_entry,is_active,default_duration_hours,starts_automatically,is_optional,is_skippable,domain_event_code)
SELECT wt.id,x.code,x.title,x.description,x.customer_title,x.customer_description,x.seq,'INSTALLATION_LEAD',r.id,x.permission,x.visible,x.seq=1,x.seq=1,TRUE,x.hours,FALSE,x.optional,x.optional,x.domain_event
FROM workflow_templates wt
JOIN (VALUES
 ('PLAN_INSTALLATION','برنامه‌ریزی نصب','تیم و زمان نصب را مشخص کنید','برنامه‌ریزی نصب','زمان‌بندی نصب مشخص می‌شود',1,'installation.update',TRUE,12,FALSE,NULL),
 ('START_INSTALLATION','شروع نصب','شروع حضور در پروژه را ثبت کنید','شروع نصب','عملیات نصب آغاز می‌شود',2,'installation.start',TRUE,4,FALSE,'INSTALLATION_STARTED'),
 ('EXECUTE_INSTALLATION','اجرای نصب','پیشرفت و تصاویر نصب را ثبت کنید','اجرای نصب','عملیات نصب در حال انجام است',3,'installation.progress',TRUE,120,FALSE,NULL),
 ('COMPLETE_INSTALLATION','تکمیل نصب','اتمام نصب را ثبت کنید','تکمیل نصب','نصب تکمیل شده است',4,'installation.complete',TRUE,8,FALSE,'INSTALLATION_COMPLETED'),
 ('CUSTOMER_FINAL_ACCEPTANCE','تأیید نهایی مشتری','تأیید عملیاتی مشتری را ثبت کنید','تأیید نهایی','در انتظار تأیید نهایی',5,'customer_acceptance.record',TRUE,24,TRUE,'CUSTOMER_ACCEPTED')
) AS x(code,title,description,customer_title,customer_description,seq,permission,visible,hours,optional,domain_event) ON TRUE
JOIN roles r ON r.code='INSTALLATION_LEAD'
WHERE wt.code='installation_execution_v1'
ON CONFLICT(workflow_template_id,step_code) DO NOTHING;
