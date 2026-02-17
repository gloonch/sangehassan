CREATE TABLE IF NOT EXISTS categories (
  id SERIAL PRIMARY KEY,
  title_en VARCHAR(255) NOT NULL,
  title_fa VARCHAR(255) NOT NULL,
  title_ar VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL UNIQUE,
  parent_id INT REFERENCES categories(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS products (
  id SERIAL PRIMARY KEY,
  title_en VARCHAR(255) NOT NULL,
  title_fa VARCHAR(255) NOT NULL,
  title_ar VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL UNIQUE,
  description_html TEXT,
  short_description_html TEXT,
  price NUMERIC(12, 2),
  price_html TEXT,
  main_category_id INT REFERENCES categories(id) ON DELETE SET NULL,
  image_url TEXT,
  is_popular BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

-- Products can belong to multiple categories (WooCommerce style)
CREATE TABLE IF NOT EXISTS product_categories (
  product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, category_id)
);

-- Product gallery images (multiple per product)
CREATE TABLE IF NOT EXISTS product_images (
  id SERIAL PRIMARY KEY,
  product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  image_url TEXT NOT NULL,
  position INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(product_id, image_url)
);

-- Attribute definitions (e.g., کاربرد, فضای کاربری, رنگ)
CREATE TABLE IF NOT EXISTS attributes (
  id SERIAL PRIMARY KEY,
  source_id INT NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  taxonomy VARCHAR(255) NOT NULL,
  type VARCHAR(50)
);

-- Attribute terms per attribute (slug kept as-is from source)
CREATE TABLE IF NOT EXISTS attribute_terms (
  id SERIAL PRIMARY KEY,
  attribute_id INT NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
  source_id INT NOT NULL,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL,
  UNIQUE(attribute_id, slug),
  UNIQUE(attribute_id, source_id)
);

-- Product ↔ attribute term join table (many-to-many)
CREATE TABLE IF NOT EXISTS product_attribute_terms (
  product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  attribute_term_id INT NOT NULL REFERENCES attribute_terms(id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, attribute_term_id)
);

CREATE TABLE IF NOT EXISTS templates (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  image_url TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS blogs (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  excerpt TEXT,
  content TEXT,
  cover_image_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS content_sections (
  id SERIAL PRIMARY KEY,
  page VARCHAR(50) NOT NULL,
  section_key VARCHAR(50) NOT NULL,
  title_en VARCHAR(255) NOT NULL,
  title_fa VARCHAR(255) NOT NULL,
  title_ar VARCHAR(255) NOT NULL,
  subtitle_en TEXT,
  subtitle_fa TEXT,
  subtitle_ar TEXT,
  description_en TEXT,
  description_fa TEXT,
  description_ar TEXT,
  cta_label_en VARCHAR(255),
  cta_label_fa VARCHAR(255),
  cta_label_ar VARCHAR(255),
  cta_href TEXT,
  order_index INT NOT NULL DEFAULT 0,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ,
  UNIQUE(page, section_key)
);

CREATE TABLE IF NOT EXISTS content_section_images (
  id SERIAL PRIMARY KEY,
  section_id INT NOT NULL REFERENCES content_sections(id) ON DELETE CASCADE,
  image_url TEXT NOT NULL,
  position INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(section_id, image_url)
);

CREATE TABLE IF NOT EXISTS blocks (
  id SERIAL PRIMARY KEY,
  title_en VARCHAR(255) NOT NULL,
  title_fa VARCHAR(255) NOT NULL,
  title_ar VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL UNIQUE,
  stone_type VARCHAR(255),
  quarry VARCHAR(255),
  dimensions VARCHAR(255),
  weight_ton NUMERIC(10, 2),
  status VARCHAR(50) NOT NULL DEFAULT 'available',
  description TEXT,
  image_url TEXT,
  is_featured BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS block_images (
  id SERIAL PRIMARY KEY,
  block_id INT NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
  image_url TEXT NOT NULL,
  position INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(block_id, image_url)
);

CREATE TABLE IF NOT EXISTS admin_users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(100) UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed admin user and templates (kept minimal to avoid conflicts with imports)
INSERT INTO admin_users (username, password_hash)
VALUES ('admin', '$2y$05$.4WCZ0rj7SGvUpfGQ1n.Qen3qCeQpQC5He0APLBBHERraPO3F7h.u')
ON CONFLICT (username) DO NOTHING;

INSERT INTO templates (name, image_url, is_active)
VALUES
  ('Triangle', '/images/templates/triangle.png', TRUE),
  ('Rectangle', '/images/templates/rectangle.png', TRUE),
  ('Circle', '/images/templates/circle.png', TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO content_sections (
  page, section_key,
  title_en, title_fa, title_ar,
  subtitle_en, subtitle_fa, subtitle_ar,
  description_en, description_fa, description_ar,
  cta_label_en, cta_label_fa, cta_label_ar,
  cta_href, order_index, is_active
)
VALUES
  (
    'home', 'blocks',
    'Stone Blocks (Quarry Blocks)', 'کوپ سنگ (Stone Block)', 'كتل الحجر (Stone Block)',
    'Raw stone extracted from the quarry for large-scale orders', 'سنگ خام معدن برای سفارش‌های حجیم و برش سفارشی', 'حجر خام من المحجر للطلبات الكبيرة والقصّ المخصص',
    E'Focus on stone type, dimensions, and tonnage.\nIdeal for large projects, export, and custom fabrication.\nReserve blocks and request tailored cutting.',
    E'تمرکز روی جنس سنگ، ابعاد و وزن بلوک است.\nمناسب برای پروژه‌های بزرگ، سفارش عمده و صادرات.\nرزرو کوپ و ثبت درخواست برش اختصاصی.',
    E'الاختيار يعتمد على نوع الحجر والأبعاد والوزن.\nمناسب للمشاريع الكبيرة والتصدير والتصنيع المخصص.\nاحجز الكتل واطلب قصاً خاصاً.',
    'Enter blocks', 'ورود به بخش کوپ', 'الدخول إلى الكتل',
    '/blocks', 1, TRUE
  ),
  (
    'home', 'finished',
    'Finished Stone', 'سنگ‌های فرآوری‌شده', 'الحجر المُعالج',
    'Slabs and ready products with multiple finishes', 'اسلب و محصولات آماده با فینیش‌های متنوع', 'ألواح ومنتجات جاهزة بتشطيبات متنوعة',
    E'Choose by stone type and surface finish.\nPerfect for facades, floors, walls, and counters.\nExplore the full catalog and technical details.',
    E'انتخاب براساس نوع سنگ و فینیش سطح.\nمناسب برای نما، کف، دیوار و کانتر.\nدسترسی به کاتالوگ و جزئیات فنی.',
    E'اختر حسب نوع الحجر وتشطيب السطح.\nمثالي للواجهات والأرضيات والجدران والكونتر.\nاستعرض الكتالوج والتفاصيل الفنية.',
    'Enter finished products', 'ورود به محصولات فرآوری‌شده', 'الدخول إلى المنتجات',
    '/products/overview', 2, TRUE
  ),
  (
    'blocks', 'hero',
    'Quarry Blocks', 'کوپ سنگ', 'كتل المحاجر',
    'Natural stone blocks prepared for custom cutting', 'بلوک‌های استخراج‌شده از معدن، آماده برش سفارشی', 'كتل حجر طبيعي جاهزة للقص المخصص',
    E'Block selection is driven by stone type, dimensions, and weight.\nBest fit for large projects and export orders.\nSend requirements to get availability and pricing.',
    E'انتخاب کوپ بر اساس نوع سنگ، ابعاد و وزن انجام می‌شود.\nبهترین گزینه برای پروژه‌های حجیم و سفارش‌های صادراتی.\nجزئیات خود را ارسال کنید تا موجودی و قیمت اعلام شود.',
    E'يعتمد الاختيار على نوع الحجر والأبعاد والوزن.\nأفضل للمشاريع الكبيرة وطلبات التصدير.\nأرسل متطلباتك للحصول على التوفر والسعر.',
    'Browse blocks', 'مشاهده کوپ‌ها', 'استعرض الكتل',
    '/blocks/catalog', 1, TRUE
  )
ON CONFLICT (page, section_key) DO NOTHING;

-- Localized product descriptions (HTML). Keep legacy description_html/short_description_html as EN fallback.
ALTER TABLE products
ADD COLUMN IF NOT EXISTS description_html_en TEXT,
ADD COLUMN IF NOT EXISTS description_html_fa TEXT,
ADD COLUMN IF NOT EXISTS description_html_ar TEXT,
ADD COLUMN IF NOT EXISTS short_description_html_en TEXT,
ADD COLUMN IF NOT EXISTS short_description_html_fa TEXT,
ADD COLUMN IF NOT EXISTS short_description_html_ar TEXT;

-- Backfill EN columns from legacy columns (safe on existing volumes; no-op if already filled).
UPDATE products
SET
  description_html_en = COALESCE(description_html_en, description_html),
  short_description_html_en = COALESCE(short_description_html_en, short_description_html)
WHERE
  (description_html_en IS NULL OR short_description_html_en IS NULL)
  AND (description_html IS NOT NULL OR short_description_html IS NOT NULL);

CREATE TABLE IF NOT EXISTS product_terms (
  id SERIAL PRIMARY KEY,
  taxonomy VARCHAR(64) NOT NULL,
  term_key VARCHAR(128) NOT NULL,
  label_en VARCHAR(255) NOT NULL,
  label_fa VARCHAR(255) NOT NULL,
  label_ar VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ,
  UNIQUE (taxonomy, term_key)
);

CREATE TABLE IF NOT EXISTS product_term_links (
  product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  term_id INT NOT NULL REFERENCES product_terms(id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, term_id)
);

CREATE INDEX IF NOT EXISTS idx_product_terms_taxonomy ON product_terms(taxonomy);
CREATE INDEX IF NOT EXISTS idx_product_term_links_product ON product_term_links(product_id);

-- Seed common taxonomies (use-case tags + stone types) so the admin panel can manage products immediately.
INSERT INTO product_terms (taxonomy, term_key, label_en, label_fa, label_ar)
VALUES
  -- Stone types
  ('stone_type', 'granite', 'Granite', 'گرانیت', 'جرانيت'),
  ('stone_type', 'marble', 'Marble', 'مرمریت', 'رخام'),
  ('stone_type', 'onyx', 'Onyx', 'اونیکس', 'أونيكس'),
  ('stone_type', 'travertine', 'Travertine', 'تراورتن', 'ترافرتين'),
  ('stone_type', 'crystal', 'Crystal', 'کریستال', 'كريستال'),
  ('stone_type', 'quartzite', 'Quartzite', 'کوارتزیت', 'كوارتزيت'),
  ('stone_type', 'agate', 'Agate', 'عقیق', 'عقيق'),
  ('stone_type', 'other', 'Other', 'سایر', 'أخرى'),

  -- Visual impact
  ('visual_impact', 'low', 'Low', 'کم', 'منخفض'),
  ('visual_impact', 'medium', 'Medium', 'متوسط', 'متوسط'),
  ('visual_impact', 'high', 'High', 'زیاد', 'مرتفع'),
  ('visual_impact', 'dramatic', 'Dramatic', 'دراماتیک', 'دراماتيكي'),

  -- Use-case: spaces
  ('use_case_space', 'interior', 'Interior', 'فضای داخلی', 'داخلي'),
  ('use_case_space', 'exterior', 'Exterior', 'فضای خارجی', 'خارجي'),
  ('use_case_space', 'wet_areas', 'Wet areas', 'فضاهای مرطوب', 'مناطق رطبة'),

  -- Use-case: forms
  ('use_case_form', 'slab', 'Slab', 'اسلب', 'ألواح'),
  ('use_case_form', 'tile', 'Tile', 'تایل', 'بلاط'),
  ('use_case_form', 'finishing', 'Finishing', 'فینیش/پرداخت', 'تشطيبات'),
  ('use_case_form', 'accessory', 'Accessory', 'اکسسوری', 'إكسسوارات'),
  ('use_case_form', 'furniture', 'Furniture', 'مبلمان/آیتم دکور', 'أثاث'),

  -- Use-case: applications
  ('use_case_application', 'countertops', 'Countertops', 'کانتر/صفحه کابینت', 'أسطح/كاونتر'),
  ('use_case_application', 'vanity', 'Vanity', 'روشویی/ونیتی', 'مغسلة/فانيتي'),
  ('use_case_application', 'flooring', 'Flooring', 'کف', 'أرضيات'),
  ('use_case_application', 'wall_cladding', 'Wall cladding', 'دیوارپوش', 'كسوة الجدران'),
  ('use_case_application', 'facade', 'Facade', 'نما', 'واجهات'),
  ('use_case_application', 'stairs', 'Stairs', 'پله', 'سلالم'),
  ('use_case_application', 'fireplace', 'Fireplace', 'شومینه', 'مدفأة'),
  ('use_case_application', 'feature_wall', 'Feature wall', 'دیوار شاخص', 'جدار مميز'),
  ('use_case_application', 'backlit', 'Backlit', 'نورپردازی از پشت', 'إضاءة خلفية'),
  ('use_case_application', 'capstone', 'Capstone/Coping', 'درپوش/سرپوش', 'تغطية/تاج'),
  ('use_case_application', 'landscaping', 'Landscaping', 'محوطه', 'تنسيق حدائق'),
  ('use_case_application', 'tabletop', 'Tabletop', 'روی میز', 'سطح طاولة'),
  ('use_case_application', 'tableware', 'Tableware', 'ظروف/سرو', 'أواني تقديم'),
  ('use_case_application', 'decor', 'Decor', 'دکور', 'ديكور'),
  ('use_case_application', 'furniture', 'Furniture', 'مبلمان', 'أثاث'),
  ('use_case_application', 'seating', 'Seating', 'نشیمن', 'جلوس'),
  ('use_case_application', 'lighting', 'Lighting', 'روشنایی', 'إضاءة'),
  ('use_case_application', 'gift', 'Gift', 'هدیه', 'هدية'),

  -- Use-case: project types
  ('use_case_project_type', 'residential', 'Residential', 'مسکونی', 'سكني'),
  ('use_case_project_type', 'commercial', 'Commercial', 'تجاری', 'تجاري'),
  ('use_case_project_type', 'hospitality', 'Hospitality', 'هتل/هُسپیـتالیتی', 'ضيافة'),
  ('use_case_project_type', 'office', 'Office', 'اداری', 'مكاتب'),
  ('use_case_project_type', 'villa', 'Villa', 'ویلا', 'فيلا'),
  ('use_case_project_type', 'retail', 'Retail', 'فروشگاهی', 'تجزئة'),

  -- Use-case: special
  ('use_case_special', 'high_traffic', 'High traffic', 'پرتردد', 'كثيف الاستخدام'),
  ('use_case_special', 'scratch_resistant', 'Scratch resistant', 'مقاوم به خط و خش', 'مقاوم للخدش'),
  ('use_case_special', 'heat_resistant', 'Heat resistant', 'مقاوم به حرارت', 'مقاوم للحرارة'),
  ('use_case_special', 'low_absorption', 'Low absorption', 'جذب آب پایین', 'امتصاص منخفض'),
  ('use_case_special', 'sealing_recommended', 'Sealing recommended', 'نیاز به سیلر', 'يُنصح بالعزل'),
  ('use_case_special', 'acid_sensitive', 'Acid sensitive', 'حساس به اسید', 'حساس للأحماض'),
  ('use_case_special', 'delicate', 'Delicate', 'حساس/ظریف', 'حساس'),
  ('use_case_special', 'translucent', 'Translucent', 'نیمه‌شفاف', 'شبه شفاف'),
  ('use_case_special', 'handmade', 'Handmade', 'دست‌ساز', 'يدوي')
ON CONFLICT (taxonomy, term_key) DO NOTHING;

-- Remove limestone stone_type from existing DBs (and its product links) so it does not appear in UI filters.
DELETE FROM product_term_links
WHERE term_id IN (
  SELECT id FROM product_terms WHERE taxonomy = 'stone_type' AND term_key = 'limestone'
);

DELETE FROM product_terms
WHERE taxonomy = 'stone_type' AND term_key = 'limestone';
