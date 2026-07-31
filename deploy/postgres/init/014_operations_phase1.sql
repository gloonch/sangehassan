-- Operational dashboard phase 1. Safe to rerun on existing databases.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

ALTER TABLE users
  ALTER COLUMN email DROP NOT NULL,
  ADD COLUMN IF NOT EXISTS phone_normalized TEXT,
  ADD COLUMN IF NOT EXISTS first_name TEXT,
  ADD COLUMN IF NOT EXISTS last_name TEXT,
  ADD COLUMN IF NOT EXISTS user_type TEXT NOT NULL DEFAULT 'CUSTOMER' CHECK (user_type IN ('INTERNAL','CUSTOMER')),
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('INVITED','ACTIVE','DISABLED','LOCKED')),
  ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS disabled_at TIMESTAMPTZ;

UPDATE users SET phone_normalized = CASE
  WHEN REGEXP_REPLACE(phone, '[^0-9+]', '', 'g') ~ '^09[0-9]{9}$' THEN '+98' || SUBSTRING(REGEXP_REPLACE(phone, '[^0-9]', '', 'g') FROM 2)
  WHEN REGEXP_REPLACE(phone, '[^0-9+]', '', 'g') ~ '^0098[0-9]{10}$' THEN '+' || SUBSTRING(REGEXP_REPLACE(phone, '[^0-9]', '', 'g') FROM 3)
  WHEN REGEXP_REPLACE(phone, '[^0-9+]', '', 'g') ~ '^98[0-9]{10}$' THEN '+' || REGEXP_REPLACE(phone, '[^0-9]', '', 'g')
  ELSE REGEXP_REPLACE(phone, '[^0-9+]', '', 'g')
END WHERE phone_normalized IS NULL AND phone IS NOT NULL;
UPDATE users SET status = 'DISABLED' WHERE is_active = FALSE AND status = 'ACTIVE';
DROP INDEX IF EXISTS uq_users_phone_normalized;
CREATE UNIQUE INDEX uq_users_phone_normalized ON users(phone_normalized);
CREATE INDEX IF NOT EXISTS idx_users_status_type ON users(status, user_type);

CREATE TABLE IF NOT EXISTS roles (
  id BIGSERIAL PRIMARY KEY, code TEXT NOT NULL UNIQUE, name_fa TEXT NOT NULL,
  description_fa TEXT NOT NULL DEFAULT '', is_system BOOLEAN NOT NULL DEFAULT FALSE,
  is_protected BOOLEAN NOT NULL DEFAULT FALSE, is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS permissions (
  id BIGSERIAL PRIMARY KEY, code TEXT NOT NULL UNIQUE, name_fa TEXT NOT NULL,
  description_fa TEXT NOT NULL DEFAULT '', group_code TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS role_permissions (
  role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), PRIMARY KEY(role_id, permission_id)
);
CREATE TABLE IF NOT EXISTS user_roles (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
  assigned_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), PRIMARY KEY(user_id, role_id)
);
CREATE TABLE IF NOT EXISTS customer_profiles (
  id BIGSERIAL PRIMARY KEY, user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  customer_code TEXT NOT NULL UNIQUE, display_name TEXT, company_name TEXT, notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS customer_activation_tokens (
  id BIGSERIAL PRIMARY KEY, user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL, expires_at TIMESTAMPTZ NOT NULL, attempt_count INT NOT NULL DEFAULT 0,
  used_at TIMESTAMPTZ, created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_customer_activation_current ON customer_activation_tokens(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS workflow_templates (
  id BIGSERIAL PRIMARY KEY, code TEXT NOT NULL UNIQUE, name_fa TEXT NOT NULL,
  description_fa TEXT NOT NULL DEFAULT '', icon_key TEXT NOT NULL,
  start_permission_code TEXT NOT NULL REFERENCES permissions(code) ON UPDATE CASCADE,
  is_active BOOLEAN NOT NULL DEFAULT TRUE, sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS workflow_template_steps (
  id BIGSERIAL PRIMARY KEY, workflow_template_id BIGINT NOT NULL REFERENCES workflow_templates(id) ON DELETE CASCADE,
  step_code TEXT NOT NULL, internal_title_fa TEXT NOT NULL, customer_title_fa TEXT NOT NULL,
  sequence_number INT NOT NULL, responsible_role_code TEXT, required_permission_code TEXT,
  customer_visible BOOLEAN NOT NULL DEFAULT FALSE, is_first_step BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workflow_template_id, step_code)
);
CREATE TABLE IF NOT EXISTS orders (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), order_number TEXT NOT NULL UNIQUE,
  customer_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  workflow_template_id BIGINT NOT NULL REFERENCES workflow_templates(id) ON DELETE RESTRICT,
  status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','PROFORMA_ISSUED','CONFIRMED','IN_PROGRESS','COMPLETED','CANCELLED')), proforma_issued_at TIMESTAMPTZ,
  confirmed_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_user_id, created_at DESC);
CREATE TABLE IF NOT EXISTS workflow_instances (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), workflow_template_id BIGINT NOT NULL REFERENCES workflow_templates(id),
  order_id UUID NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
  customer_user_id UUID NOT NULL REFERENCES users(id), status TEXT NOT NULL DEFAULT 'IN_PROGRESS' CHECK (status IN ('IN_PROGRESS','COMPLETED','CANCELLED')),
  started_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL, idempotency_key TEXT,
  started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), completed_at TIMESTAMPTZ, cancelled_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(started_by_user_id, idempotency_key)
);
CREATE TABLE IF NOT EXISTS workflow_step_instances (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  workflow_template_step_id BIGINT NOT NULL REFERENCES workflow_template_steps(id), status TEXT NOT NULL DEFAULT 'NOT_STARTED' CHECK (status IN ('NOT_STARTED','WAITING_FOR_ASSIGNEE','IN_PROGRESS','SUBMITTED','WAITING_FOR_APPROVAL','APPROVED','HAS_MISMATCH','NEEDS_CORRECTION','BLOCKED','COMPLETED','SKIPPED','CANCELLED')),
  assigned_role_id BIGINT REFERENCES roles(id), assigned_user_id UUID REFERENCES users(id),
  estimated_start_at TIMESTAMPTZ, actual_start_at TIMESTAMPTZ, estimated_end_at TIMESTAMPTZ, actual_end_at TIMESTAMPTZ,
  customer_visible BOOLEAN NOT NULL DEFAULT FALSE, customer_status_text TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS action_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), workflow_instance_id UUID REFERENCES workflow_instances(id) ON DELETE CASCADE,
  workflow_step_instance_id UUID REFERENCES workflow_step_instances(id) ON DELETE CASCADE,
  order_id UUID REFERENCES orders(id) ON DELETE CASCADE, customer_user_id UUID REFERENCES users(id),
  title_fa TEXT NOT NULL, description_fa TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN','IN_PROGRESS','WAITING_FOR_APPROVAL','BLOCKED','COMPLETED','CANCELLED')),
  priority TEXT NOT NULL DEFAULT 'NORMAL' CHECK (priority IN ('LOW','NORMAL','HIGH','URGENT')), assigned_role_id BIGINT REFERENCES roles(id), assigned_user_id UUID REFERENCES users(id),
  required_permission_code TEXT, due_at TIMESTAMPTZ, completed_at TIMESTAMPTZ, completed_by_user_id UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_action_items_assignee ON action_items(assigned_user_id, assigned_role_id, status);
CREATE TABLE IF NOT EXISTS proformas (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), proforma_number TEXT NOT NULL UNIQUE,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE, customer_user_id UUID NOT NULL REFERENCES users(id),
  status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','ISSUED','ACCEPTED','REJECTED','CANCELLED')), currency CHAR(3) NOT NULL DEFAULT 'IRR', subtotal NUMERIC(18,2) NOT NULL DEFAULT 0,
  discount_amount NUMERIC(18,2) NOT NULL DEFAULT 0, total_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
  notes TEXT, issued_by_user_id UUID REFERENCES users(id), issued_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_proformas_customer ON proformas(customer_user_id, created_at DESC);
CREATE TABLE IF NOT EXISTS order_timeline (
  id BIGSERIAL PRIMARY KEY, order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  title_fa TEXT NOT NULL, status_code TEXT NOT NULL, occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGSERIAL PRIMARY KEY, actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  action_code TEXT NOT NULL, entity_type TEXT NOT NULL, entity_id TEXT,
  before_data JSONB, after_data JSONB, metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  ip_address INET, user_agent TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);

INSERT INTO permissions(code,name_fa,description_fa,group_code) VALUES
('dashboard.internal.view','مشاهده داشبورد داخلی','مشاهده پنل عملیاتی','SYSTEM'),('dashboard.customer.view','مشاهده داشبورد مشتری','مشاهده حساب مشتری','SYSTEM'),('audit.view','مشاهده گزارش تغییرات','مشاهده رویدادهای ممیزی','SYSTEM'),('content.manage','مدیریت محتوای وب‌سایت','مدیریت محصولات، نوشته‌ها و تنظیمات وب‌سایت','SYSTEM'),
('users.view','مشاهده کاربران','مشاهده فهرست کاربران','USERS'),('users.create','ایجاد کاربر','ایجاد کاربر داخلی','USERS'),('users.update','ویرایش کاربر','ویرایش اطلاعات کاربر','USERS'),('users.disable','غیرفعال‌سازی کاربر','تغییر وضعیت کاربر','USERS'),('users.reset_password','بازنشانی رمز','ایجاد رمز موقت','USERS'),('users.assign_roles','تخصیص نقش','مدیریت نقش‌های کاربر','USERS'),
('roles.view','مشاهده نقش‌ها','مشاهده نقش‌ها','ROLES'),('roles.create','ایجاد نقش','ایجاد نقش سفارشی','ROLES'),('roles.update','ویرایش نقش','ویرایش نقش','ROLES'),('roles.deactivate','غیرفعال‌سازی نقش','غیرفعال‌سازی نقش سفارشی','ROLES'),('roles.assign_permissions','تخصیص دسترسی','مدیریت دسترسی نقش','ROLES'),('roles.assign_users','تخصیص کاربران','تخصیص کاربران به نقش','ROLES'),('permissions.view','مشاهده دسترسی‌ها','مشاهده فهرست دسترسی‌ها','ROLES'),
('administrators.create','ایجاد مدیر','ایجاد مدیر جدید','ADMIN'),('administrators.update','ویرایش مدیر','ویرایش مدیر','ADMIN'),('administrators.disable','غیرفعال‌سازی مدیر','غیرفعال‌سازی مدیر','ADMIN'),
('customers.view','مشاهده مشتریان','مشاهده مشتریان','CUSTOMERS'),('customers.create','ایجاد مشتری','ایجاد مشتری','CUSTOMERS'),('customers.update','ویرایش مشتری','ویرایش مشتری','CUSTOMERS'),('customers.invite','دعوت مشتری','صدور کد فعال‌سازی','CUSTOMERS'),('customers.regenerate_activation','صدور مجدد کد','بازسازی کد فعال‌سازی','CUSTOMERS'),('customers.disable','غیرفعال‌سازی مشتری','غیرفعال‌سازی مشتری','CUSTOMERS'),
('workflow_templates.view','مشاهده الگوها','مشاهده الگوهای فرایند','WORKFLOWS'),('workflow_templates.manage','مدیریت الگوها','مدیریت الگوهای فرایند','WORKFLOWS'),('workflow_instances.start','شروع فرایند','ایجاد فرایند جدید','WORKFLOWS'),('workflow_instances.view_assigned','مشاهده فرایندهای محول‌شده','مشاهده فرایند محول‌شده','WORKFLOWS'),('workflow_instances.view_all','مشاهده همه فرایندها','مشاهده همه فرایندها','WORKFLOWS'),('workflow_instances.update','ویرایش فرایند','ویرایش فرایند','WORKFLOWS'),('workflow_instances.cancel','لغو فرایند','لغو فرایند','WORKFLOWS'),('workflow_instances.override','بازنویسی فرایند','عملیات مدیریتی فرایند','WORKFLOWS'),
('workflow_steps.view_assigned','مشاهده مراحل محول‌شده','مشاهده مراحل محول‌شده','STEPS'),('workflow_steps.submit','ثبت مرحله','ثبت نتیجه مرحله','STEPS'),('workflow_steps.approve','تأیید مرحله','تأیید مرحله','STEPS'),('workflow_steps.reject','رد مرحله','رد مرحله','STEPS'),('workflow_steps.skip','رد کردن مرحله','عبور از مرحله','STEPS'),('workflow_steps.reassign','تغییر مسئول مرحله','تخصیص مجدد مرحله','STEPS'),('workflow_steps.override','بازنویسی مرحله','عملیات مدیریتی مرحله','STEPS'),
('action_items.view_own','مشاهده اقدامات خود','مشاهده اقدامات محول‌شده','ACTIONS'),('action_items.view_all','مشاهده همه اقدامات','مشاهده همه اقدامات','ACTIONS'),('action_items.assign','تخصیص اقدام','تخصیص اقدام','ACTIONS'),('action_items.reassign','تغییر مسئول اقدام','تخصیص مجدد اقدام','ACTIONS'),('action_items.complete','تکمیل اقدام','تکمیل اقدام','ACTIONS'),
('orders.view_own','مشاهده سفارش‌های خود','مشاهده سفارش‌های مرتبط','ORDERS'),('orders.view_all','مشاهده همه سفارش‌ها','مشاهده همه سفارش‌ها','ORDERS'),('orders.create','ایجاد سفارش','ایجاد سفارش','ORDERS'),('orders.update','ویرایش سفارش','ویرایش سفارش','ORDERS'),('orders.cancel','لغو سفارش','لغو سفارش','ORDERS'),
('proformas.view_own','مشاهده پیش‌فاکتورهای خود','مشاهده پیش‌فاکتورهای مرتبط','PROFORMAS'),('proformas.view_all','مشاهده همه پیش‌فاکتورها','مشاهده همه پیش‌فاکتورها','PROFORMAS'),('proformas.create','ایجاد پیش‌فاکتور','ایجاد پیش‌فاکتور','PROFORMAS'),('proformas.update','ویرایش پیش‌فاکتور','ویرایش پیش‌فاکتور','PROFORMAS'),('proformas.issue','صدور پیش‌فاکتور','صدور پیش‌فاکتور برای مشتری','PROFORMAS'),('proformas.cancel','لغو پیش‌فاکتور','لغو پیش‌فاکتور','PROFORMAS'),('proformas.download','دریافت پیش‌فاکتور','مشاهده نسخه چاپی','PROFORMAS'),
('sales.sale_price.view','مشاهده قیمت فروش','مشاهده قیمت فروش','SALES'),('sales.sale_price.update','ویرایش قیمت فروش','ویرایش قیمت فروش','SALES'),('sales.discount.apply','اعمال تخفیف','اعمال تخفیف فروش','SALES'),('sales.customer_contact.view','مشاهده تماس مشتری','مشاهده سوابق تماس','SALES'),('sales.customer_contact.log','ثبت تماس مشتری','ثبت گزارش تماس','SALES'),
('finance.payments.view','مشاهده پرداخت‌ها','مشاهده پرداخت‌ها','FINANCE'),('finance.payments.record','ثبت پرداخت','ثبت پرداخت','FINANCE'),('finance.payments.confirm','تأیید پرداخت','تأیید پرداخت','FINANCE'),('finance.purchase_proformas.view','مشاهده پیش‌فاکتور خرید','مشاهده پیش‌فاکتور خرید','FINANCE'),('finance.sales_invoices.view','مشاهده فاکتور فروش','مشاهده فاکتور فروش','FINANCE'),('finance.transport_costs.view','مشاهده کرایه','مشاهده هزینه حمل','FINANCE'),('finance.side_costs.view','مشاهده هزینه جانبی','مشاهده هزینه جانبی','FINANCE'),('finance.internal_costs.view','مشاهده هزینه داخلی','مشاهده هزینه داخلی','FINANCE'),('finance.internal_costs.record','ثبت هزینه داخلی','ثبت هزینه داخلی','FINANCE'),('finance.profit.view','مشاهده سود','مشاهده سود سفارش','FINANCE'),
('operations.procurement.execute','اجرای تأمین','اجرای عملیات تأمین','OPERATIONS'),('operations.procurement.confirm_quantity','تأیید مقدار تأمین','تأیید مقدار تأمین‌شده','OPERATIONS'),('operations.transport.execute','اجرای حمل','اجرای حمل','OPERATIONS'),('operations.transport.confirm_pickup','تأیید بارگیری','تأیید دریافت بار','OPERATIONS'),('operations.transport.confirm_delivery','تأیید تحویل','تأیید تحویل بار','OPERATIONS'),('operations.processing.execute','اجرای فرآوری','اجرای فرآوری','OPERATIONS'),('operations.processing.confirm_quantity','تأیید مقدار فرآوری','تأیید مقدار فرآوری‌شده','OPERATIONS'),('operations.quality_control.execute','اجرای کنترل کیفیت','کنترل کیفیت','OPERATIONS'),('operations.packaging.execute','اجرای بسته‌بندی','بسته‌بندی','OPERATIONS'),('operations.export.execute','اجرای صادرات','امور صادرات','OPERATIONS'),('operations.installation.execute','اجرای نصب','عملیات نصب','OPERATIONS'),
('customer_portal.orders.view_own','مشاهده سفارش خود','مشاهده سفارش‌های مشتری','CUSTOMER_PORTAL'),('customer_portal.proformas.view_own','مشاهده پیش‌فاکتور خود','مشاهده پیش‌فاکتورهای مشتری','CUSTOMER_PORTAL'),('customer_portal.proformas.download_own','دریافت پیش‌فاکتور خود','نسخه چاپی پیش‌فاکتور مشتری','CUSTOMER_PORTAL'),('customer_portal.workflow.view_own','مشاهده روند سفارش','مشاهده مراحل عمومی سفارش','CUSTOMER_PORTAL'),('customer_portal.payments.view_own','مشاهده پرداخت خود','مشاهده پرداخت‌های مشتری','CUSTOMER_PORTAL'),
('workflow_start.mine_to_domestic_factory','شروع خرید معدن برای کارخانه','شروع فرایند','WORKFLOW_START'),('workflow_start.mine_to_domestic_dealer','شروع خرید معدن برای دلال','شروع فرایند','WORKFLOW_START'),('workflow_start.mine_to_export','شروع خرید معدن برای صادرات','شروع فرایند','WORKFLOW_START'),('workflow_start.block_warehouse_purchase','شروع خرید از انبار کوپ','شروع فرایند','WORKFLOW_START'),('workflow_start.factory_to_export','شروع خرید کارخانه برای صادرات','شروع فرایند','WORKFLOW_START'),('workflow_start.factory_to_dealer','شروع خرید کارخانه برای دلال','شروع فرایند','WORKFLOW_START'),('workflow_start.factory_to_project','شروع خرید کارخانه برای پروژه','شروع فرایند','WORKFLOW_START'),('workflow_start.showroom_purchase','شروع خرید از شوروم','شروع فرایند','WORKFLOW_START'),('workflow_start.warehouse_purchase','شروع خرید از انبار','شروع فرایند','WORKFLOW_START')
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa, description_fa=EXCLUDED.description_fa, group_code=EXCLUDED.group_code;

INSERT INTO roles(code,name_fa,description_fa,is_system,is_protected) VALUES
('SUPER_ADMIN','مدیر اصلی','دسترسی کامل و محافظت‌شده',TRUE,TRUE),('ADMIN','مدیر','مدیریت عملیات و کاربران',TRUE,FALSE),
('OPERATOR','اپراتور سایت','ثبت مشتری، سفارش و پیش‌فاکتور',TRUE,FALSE),('SALES','مسئول فروش','فروش و ارتباط با مشتری',TRUE,FALSE),
('ACCOUNTANT','حسابدار','عملیات مالی',TRUE,FALSE),('SUPPLY','مسئول تأمین','تأمین و عملیات',TRUE,FALSE),
('DRIVER','راننده','حمل و تحویل',TRUE,FALSE),('INSTALLATION_LEAD','سرگروه نصب','عملیات نصب',TRUE,FALSE),
('CUSTOMER','مشتری','دسترسی محدود حساب مشتری',TRUE,TRUE)
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa, description_fa=EXCLUDED.description_fa;

-- Super Admin is evaluated dynamically; these rows also make the editor explicit.
INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE r.code='SUPER_ADMIN' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r JOIN permissions p ON
  (r.code='ADMIN' AND p.code NOT LIKE 'administrators.%') OR
  (r.code='OPERATOR' AND (p.code IN ('dashboard.internal.view','customers.view','customers.create','customers.update','customers.invite','customers.regenerate_activation','workflow_templates.view','workflow_instances.start','workflow_instances.view_all','workflow_instances.update','action_items.view_own','action_items.complete','orders.view_all','orders.create','orders.update','proformas.view_all','proformas.create','proformas.update','proformas.issue','proformas.download','sales.sale_price.view','sales.sale_price.update','sales.customer_contact.view','sales.customer_contact.log') OR p.code LIKE 'workflow_start.%')) OR
  (r.code='SALES' AND (p.code IN ('dashboard.internal.view','customers.view','customers.create','customers.update','customers.invite','customers.regenerate_activation','workflow_templates.view','workflow_instances.start','workflow_instances.view_all','action_items.view_own','action_items.complete','orders.view_all','orders.create','orders.update','proformas.view_all','proformas.create','proformas.update','proformas.issue','proformas.download','sales.sale_price.view','sales.sale_price.update','sales.discount.apply','sales.customer_contact.view','sales.customer_contact.log','finance.payments.view') OR p.code LIKE 'workflow_start.%')) OR
  (r.code='ACCOUNTANT' AND p.code IN ('dashboard.internal.view','workflow_instances.view_all','action_items.view_own','action_items.complete','orders.view_all','proformas.view_all','proformas.download','finance.payments.view','finance.payments.record','finance.payments.confirm','finance.purchase_proformas.view','finance.sales_invoices.view','finance.transport_costs.view','finance.side_costs.view','finance.internal_costs.view','finance.internal_costs.record')) OR
  (r.code='SUPPLY' AND p.code IN ('dashboard.internal.view','workflow_instances.view_assigned','workflow_steps.view_assigned','workflow_steps.submit','workflow_steps.approve','workflow_steps.reject','action_items.view_own','action_items.complete','operations.procurement.execute','operations.procurement.confirm_quantity','operations.processing.execute','operations.processing.confirm_quantity','operations.quality_control.execute','operations.packaging.execute','operations.export.execute','finance.internal_costs.record')) OR
  (r.code='DRIVER' AND p.code IN ('dashboard.internal.view','workflow_instances.view_assigned','workflow_steps.view_assigned','workflow_steps.submit','action_items.view_own','action_items.complete','operations.transport.execute','operations.transport.confirm_pickup','operations.transport.confirm_delivery','finance.internal_costs.record')) OR
  (r.code='INSTALLATION_LEAD' AND p.code IN ('dashboard.internal.view','workflow_instances.view_assigned','workflow_steps.view_assigned','workflow_steps.submit','action_items.view_own','action_items.complete','operations.installation.execute','finance.internal_costs.record')) OR
  (r.code='CUSTOMER' AND p.code IN ('dashboard.customer.view','customer_portal.orders.view_own','customer_portal.proformas.view_own','customer_portal.proformas.download_own','customer_portal.workflow.view_own','customer_portal.payments.view_own'))
ON CONFLICT DO NOTHING;

INSERT INTO workflow_templates(code,name_fa,description_fa,icon_key,start_permission_code,sort_order) VALUES
('mine_to_domestic_factory','خرید از معدن برای کارخانه داخلی','تأمین معدن و تحویل کارخانه داخلی','mine','workflow_start.mine_to_domestic_factory',1),
('mine_to_domestic_dealer','خرید از معدن برای دلال داخلی','تأمین معدن برای فروش داخلی','mine','workflow_start.mine_to_domestic_dealer',2),
('mine_to_export','خرید از معدن برای صادرات','تأمین مستقیم برای صادرات','export','workflow_start.mine_to_export',3),
('block_warehouse_purchase','خرید از انبار کوپ','خرید بلوک از انبار','warehouse','workflow_start.block_warehouse_purchase',4),
('factory_to_export','خرید از کارخانه برای صادرات','خرید محصول کارخانه برای صادرات','factory','workflow_start.factory_to_export',5),
('factory_to_dealer','خرید از کارخانه برای دلال','خرید محصول کارخانه برای دلال','factory','workflow_start.factory_to_dealer',6),
('factory_to_project','خرید از کارخانه برای پروژه','خرید محصول کارخانه برای پروژه','project','workflow_start.factory_to_project',7),
('showroom_purchase','خرید از شوروم','خرید از شوروم','showroom','workflow_start.showroom_purchase',8),
('warehouse_purchase','خرید از انبار','خرید از موجودی انبار','warehouse','workflow_start.warehouse_purchase',9)
ON CONFLICT(code) DO UPDATE SET name_fa=EXCLUDED.name_fa,description_fa=EXCLUDED.description_fa,icon_key=EXCLUDED.icon_key,start_permission_code=EXCLUDED.start_permission_code,sort_order=EXCLUDED.sort_order;
INSERT INTO workflow_template_steps(workflow_template_id,step_code,internal_title_fa,customer_title_fa,sequence_number,responsible_role_code,required_permission_code,customer_visible,is_first_step)
SELECT id,'ISSUE_PROFORMA','صدور پیش‌فاکتور','پیش‌فاکتور صادر شد',1,'OPERATOR','proformas.issue',TRUE,TRUE FROM workflow_templates
ON CONFLICT(workflow_template_id,step_code) DO NOTHING;
