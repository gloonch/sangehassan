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
  video_url TEXT,
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

ALTER TABLE products
ADD COLUMN IF NOT EXISTS aliases TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS variants TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS mines TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS finishes TEXT[] NOT NULL DEFAULT '{}'::text[];

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
    '/products', 2, TRUE
  ),
  (
    'blocks', 'hero',
    'Quarry Blocks', 'کوپ سنگ', 'كتل المحاجر',
    'Natural stone blocks prepared for custom cutting', 'بلوک‌های استخراج‌شده از معدن، آماده برش سفارشی', 'كتل حجر طبيعي جاهزة للقص المخصص',
    E'Block selection is driven by stone type, dimensions, and weight.\nBest fit for large projects and export orders.\nSend requirements to get availability and pricing.',
    E'انتخاب کوپ بر اساس نوع سنگ، ابعاد و وزن انجام می‌شود.\nبهترین گزینه برای پروژه‌های حجیم و سفارش‌های صادراتی.\nجزئیات خود را ارسال کنید تا موجودی و قیمت اعلام شود.',
    E'يعتمد الاختيار على نوع الحجر والأبعاد والوزن.\nأفضل للمشاريع الكبيرة وطلبات التصدير.\nأرسل متطلباتك للحصول على التوفر والسعر.',
    'Browse blocks', 'مشاهده کوپ‌ها', 'استعرض الكتل',
    '/blocks', 1, TRUE
  ),
  (
    'products', 'hero',
    'Finished Stone', 'سنگ‌های فرآوری‌شده', 'الحجر المُعالج',
    'Slabs and ready products with multiple finishes', 'اسلب و محصولات آماده با فینیش‌های متنوع', 'ألواح ومنتجات جاهزة بتشطيبات متنوعة',
    E'Choose by stone type and surface finish.\nPerfect for facades, floors, walls, and counters.\nExplore the full catalog and technical details.',
    E'انتخاب براساس نوع سنگ و فینیش سطح.\nمناسب برای نما، کف، دیوار و کانتر.\nدسترسی به کاتالوگ و جزئیات فنی.',
    E'اختر حسب نوع الحجر وتشطيب السطح.\nمثالي للواجهات والأرضيات والجدران والكونتر.\nاستعرض الكتالوج والتفاصيل الفنية.',
    'Explore products', 'مشاهده محصولات', 'استعرض المنتجات',
    '/products', 1, TRUE
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
  link_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ,
  UNIQUE (taxonomy, term_key)
);

ALTER TABLE product_terms
ADD COLUMN IF NOT EXISTS link_url TEXT;

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

INSERT INTO product_terms (taxonomy, term_key, label_en, label_fa, label_ar)
VALUES
  ('finishes', 'cloudy-finish', 'Cloudy Finish', 'ابری', 'تشطيب سحابي'),
  ('finishes', 'scratch', 'Scratch', 'اسکرچ', 'تشطيب مخدوش'),
  ('finishes', 'unfilled', 'Unfilled', 'آنفیلد', 'غير مملوء'),
  ('finishes', 'bush-hammer', 'Bush Hammer', 'بوش همر', 'بوش هامر'),
  ('finishes', 'leathered-bush-hammer', 'Leathered Bush Hammer', 'بوش همر چرمی', 'بوش هامر جلدي'),
  ('finishes', 'pre-cut', 'Pre-Cut', 'پری‌کات', 'مقطّع مسبقاً'),
  ('finishes', 'polished', 'Polished', 'پولیش (صیقلی)', 'مصقول'),
  ('finishes', 'mesh-backed', 'Mesh-Backed', 'توری', 'مدعّم بشبكة'),
  ('finishes', 'mesh-backed-resin-epoxy', 'Mesh-Backed, Resin and Epoxy Treated', 'توری، رزین و اپوکسی', 'مدعّم بشبكة ومعالج بالراتنج والإيبوكسي'),
  ('finishes', 'chiseled-machine-hand', 'Chiseled (Machine and Hand)', 'تیشه‌ای (ماشینی و دستی)', 'منقور بالإزميل (آلي ويدوي)'),
  ('finishes', 'leathered', 'Leathered', 'چرمی', 'تشطيب جلدي'),
  ('finishes', 'rock', 'Rock', 'راک (بادبر)', 'تشطيب صخري'),
  ('finishes', 'resin-epoxy', 'Resin and Epoxy Treated', 'رزین و اپوکسی', 'معالج بالراتنج والإيبوكسي'),
  ('finishes', 'sandblasted', 'Sandblasted', 'سندبلاست', 'مسفوع بالرمل'),
  ('finishes', 'shot-blast', 'Shot Blast', 'شات بلاست', 'تشطيب بالقذف الحبيبي'),
  ('finishes', 'grooved', 'Grooved', 'شیار', 'محزّز'),
  ('finishes', 'flamed', 'Flamed', 'فلیم', 'معالج باللهب'),
  ('finishes', 'leather-flamed', 'Leather Flamed', 'فلیم چرمی', 'معالج باللهب وبتشطيب جلدي'),
  ('finishes', 'flamed-scratch', 'Flamed Scratch', 'فلیم اسکرچ', 'معالج باللهب ومخدوش'),
  ('finishes', 'filled', 'Filled', 'فیلد', 'مملوء'),
  ('finishes', 'cut-broken', 'Cut Broken', 'کات بروکن', 'قطع مكسّر'),
  ('finishes', 'cut-hammered', 'Cut Hammered', 'کات هَمِرد', 'قطع مطرّق'),
  ('finishes', 'lava', 'Lava', 'لاوا', 'لافا'),
  ('finishes', 'honed', 'Honed', 'هوند (نسابیده)', 'مصقول مطفي'),
  ('finishes', 'oblique', 'Oblique', 'اوبلیک (مورب)', 'مائل'),
  ('finishes', 'super-rock', 'Super Rock', 'سوپر راک', 'سوبر روك'),
  ('finishes', 'chevron', 'Chevron', 'شورون (هفت‌وهشتی)', 'شيفرون'),
  ('finishes', 'lineal', 'Lineal', 'لینئال', 'لينيل'),
  ('finishes', 'lineal-plus', 'Lineal Plus', 'لینئال پلاس', 'لينيل بلس'),
  ('finishes', 'lineal-extreme', 'Lineal Extreme', 'لینئال اکستریم', 'لينيل إكستريم'),
  ('finishes', 'vintage', 'Vintage', 'وینتیج', 'فينتج'),
  ('finishes', 'cotton', 'Cotton', 'کاتن', 'كوتون')
ON CONFLICT (taxonomy, term_key) DO UPDATE SET
  label_en = EXCLUDED.label_en,
  label_fa = EXCLUDED.label_fa,
  label_ar = EXCLUDED.label_ar,
  updated_at = NOW();

WITH finish_products AS (
  SELECT
    p.slug,
    LOWER(REGEXP_REPLACE(REGEXP_REPLACE(TRIM(p.title_en), '[^a-zA-Z0-9]+', '-', 'g'), '(^-|-$)', '', 'g')) AS title_key
  FROM products p
  JOIN categories c ON c.id = p.main_category_id
  WHERE c.slug = 'finishings'
),
matched_finish_links AS (
  SELECT DISTINCT ON (t.id)
    t.id AS term_id,
    '/products/' || fp.slug AS link_url
  FROM product_terms t
  JOIN finish_products fp
    ON t.term_key = fp.slug
    OR t.term_key = fp.title_key
    OR (t.term_key = 'sandblasted' AND fp.slug = 'sandblast')
  WHERE t.taxonomy = 'finishes'
  ORDER BY t.id, fp.slug
)
UPDATE product_terms t
SET
  link_url = matched_finish_links.link_url,
  updated_at = NOW()
FROM matched_finish_links
WHERE t.id = matched_finish_links.term_id
  AND COALESCE(t.link_url, '') IS DISTINCT FROM matched_finish_links.link_url;

WITH source_finish_links AS (
  SELECT p.id AS product_id, LOWER(REGEXP_REPLACE(LEFT(TRIM(value), 255), '[[:space:]]+', '', 'g')) AS normalized_label
  FROM products p
  CROSS JOIN LATERAL UNNEST(COALESCE(p.finishes, '{}'::text[])) AS value
  WHERE TRIM(value) <> ''
),
matched_terms AS (
  SELECT DISTINCT sl.product_id, t.id AS term_id
  FROM source_finish_links sl
  JOIN product_terms t
    ON t.taxonomy = 'finishes'
   AND COALESCE(t.link_url, '') <> ''
   AND (
      LOWER(REGEXP_REPLACE(TRIM(t.label_en), '[[:space:]]+', '', 'g')) = sl.normalized_label
      OR LOWER(REGEXP_REPLACE(TRIM(t.label_fa), '[[:space:]]+', '', 'g')) = sl.normalized_label
      OR LOWER(REGEXP_REPLACE(TRIM(t.label_ar), '[[:space:]]+', '', 'g')) = sl.normalized_label
      OR LOWER(REGEXP_REPLACE(TRIM(t.term_key), '[[:space:]]+', '', 'g')) = sl.normalized_label
   )
)
INSERT INTO product_term_links (product_id, term_id)
SELECT product_id, term_id
FROM matched_terms
ON CONFLICT DO NOTHING;

WITH source_terms AS (
  SELECT 'variants' AS taxonomy, LEFT(TRIM(value), 255) AS label
  FROM products p
  CROSS JOIN LATERAL UNNEST(COALESCE(p.variants, '{}'::text[])) AS value
  UNION ALL
  SELECT 'mines' AS taxonomy, LEFT(TRIM(value), 255) AS label
  FROM products p
  CROSS JOIN LATERAL UNNEST(COALESCE(p.mines, '{}'::text[])) AS value
),
normalized_terms AS (
  SELECT
    taxonomy,
    label,
    LOWER(REGEXP_REPLACE(label, '[[:space:]]+', ' ', 'g')) AS normalized_label
  FROM source_terms
  WHERE label <> ''
),
unique_terms AS (
  SELECT DISTINCT ON (taxonomy, normalized_label)
    taxonomy,
    label,
    normalized_label
  FROM normalized_terms
  ORDER BY taxonomy, normalized_label, label
)
INSERT INTO product_terms (taxonomy, term_key, label_en, label_fa, label_ar)
SELECT
  taxonomy,
  'legacy-' || taxonomy || '-' || SUBSTR(MD5(normalized_label), 1, 16),
  CASE WHEN label ~ '[؀-ۿ]' THEN '' ELSE label END,
  CASE WHEN label ~ '[؀-ۿ]' THEN label ELSE '' END,
  ''
FROM unique_terms u
WHERE NOT EXISTS (
  SELECT 1
  FROM product_terms t
  WHERE t.taxonomy = u.taxonomy
    AND (
      LOWER(REGEXP_REPLACE(TRIM(t.label_en), '[[:space:]]+', ' ', 'g')) = u.normalized_label
      OR LOWER(REGEXP_REPLACE(TRIM(t.label_fa), '[[:space:]]+', ' ', 'g')) = u.normalized_label
      OR LOWER(REGEXP_REPLACE(TRIM(t.label_ar), '[[:space:]]+', ' ', 'g')) = u.normalized_label
    )
)
ON CONFLICT (taxonomy, term_key) DO NOTHING;

UPDATE product_terms
SET
  label_en = CASE WHEN label_fa ~ '[؀-ۿ]' THEN '' ELSE label_en END,
  label_ar = '',
  label_fa = CASE WHEN label_fa ~ '[؀-ۿ]' THEN label_fa ELSE '' END,
  updated_at = NOW()
WHERE taxonomy IN ('variants', 'mines')
  AND term_key LIKE 'legacy-%'
  AND label_en = label_fa
  AND label_fa = label_ar;

WITH source_links AS (
  SELECT p.id AS product_id, 'variants' AS taxonomy, LOWER(REGEXP_REPLACE(LEFT(TRIM(value), 255), '[[:space:]]+', ' ', 'g')) AS normalized_label
  FROM products p
  CROSS JOIN LATERAL UNNEST(COALESCE(p.variants, '{}'::text[])) AS value
  UNION ALL
  SELECT p.id AS product_id, 'mines' AS taxonomy, LOWER(REGEXP_REPLACE(LEFT(TRIM(value), 255), '[[:space:]]+', ' ', 'g')) AS normalized_label
  FROM products p
  CROSS JOIN LATERAL UNNEST(COALESCE(p.mines, '{}'::text[])) AS value
),
matched_terms AS (
  SELECT DISTINCT sl.product_id, t.id AS term_id
  FROM source_links sl
  JOIN product_terms t
    ON t.taxonomy = sl.taxonomy
   AND (
      LOWER(REGEXP_REPLACE(TRIM(t.label_en), '[[:space:]]+', ' ', 'g')) = sl.normalized_label
      OR LOWER(REGEXP_REPLACE(TRIM(t.label_fa), '[[:space:]]+', ' ', 'g')) = sl.normalized_label
      OR LOWER(REGEXP_REPLACE(TRIM(t.label_ar), '[[:space:]]+', ' ', 'g')) = sl.normalized_label
   )
  WHERE sl.normalized_label <> ''
)
INSERT INTO product_term_links (product_id, term_id)
SELECT product_id, term_id
FROM matched_terms
ON CONFLICT DO NOTHING;

INSERT INTO product_terms (taxonomy, term_key, label_en, label_fa, label_ar)
VALUES
  ('finishes', 'pre-cut', 'Pre-Cut', 'پری‌کات', 'مقطّع مسبقاً'),
  ('finishes', 'chiseled-machine-hand', 'Chiseled (Machine and Hand)', 'تیشه‌ای (ماشینی و دستی)', 'منقور بالإزميل (آلي ويدوي)'),
  ('finishes', 'grooved', 'Grooved', 'شیار', 'محزّز'),
  ('finishes', 'flamed', 'Flamed', 'فلیم', 'معالج باللهب'),
  ('finishes', 'honed', 'Honed', 'هوند (نسابیده)', 'مصقول مطفي')
ON CONFLICT (taxonomy, term_key) DO UPDATE SET
  label_en = EXCLUDED.label_en,
  label_fa = EXCLUDED.label_fa,
  label_ar = EXCLUDED.label_ar,
  updated_at = NOW();

WITH rule_terms AS (
  SELECT * FROM (VALUES
    ('pre-cut', 'Pre-Cut', 'پری‌کات', 'مقطّع مسبقاً', ARRAY['granite', 'marble', 'crystal', 'basalt']::text[], FALSE, ARRAY[]::text[]),
    ('chiseled-machine-hand', 'Chiseled (Machine and Hand)', 'تیشه‌ای (ماشینی و دستی)', 'منقور بالإزميل (آلي ويدوي)', ARRAY['marble', 'travertine', 'granite']::text[], FALSE, ARRAY[]::text[]),
    ('leathered-bush-hammer', 'Leathered Bush Hammer', 'بوش همر چرمی', 'بوش هامر جلدي', ARRAY['granite', 'travertine', 'basalt']::text[], FALSE, ARRAY[]::text[]),
    ('flamed', 'Flamed', 'فلیم', 'معالج باللهب', ARRAY['granite']::text[], FALSE, ARRAY[]::text[]),
    ('honed', 'Honed', 'هوند (نسابیده)', 'مصقول مطفي', ARRAY[]::text[], TRUE, ARRAY[]::text[]),
    ('grooved', 'Grooved', 'شیار', 'محزّز', ARRAY[]::text[], FALSE, ARRAY['lineal', 'scratch', 'لینئال', 'اسکرچ', 'لينيل', 'تشطيب مخدوش']::text[])
  ) AS v(term_key, label_en, label_fa, label_ar, category_slugs, all_products, equivalent_labels)
),
product_category_slugs AS (
  SELECT DISTINCT p.id AS product_id, c.slug
  FROM products p
  JOIN categories c ON c.id = p.main_category_id
  UNION
  SELECT DISTINCT pc.product_id, c.slug
  FROM product_categories pc
  JOIN categories c ON c.id = pc.category_id
),
category_candidates AS (
  SELECT DISTINCT pcs.product_id, rt.term_key, rt.label_en, rt.label_fa, rt.label_ar
  FROM product_category_slugs pcs
  JOIN rule_terms rt ON pcs.slug = ANY(rt.category_slugs)
  WHERE NOT rt.all_products
    AND CARDINALITY(rt.equivalent_labels) = 0
),
all_product_candidates AS (
  SELECT p.id AS product_id, rt.term_key, rt.label_en, rt.label_fa, rt.label_ar
  FROM products p
  JOIN rule_terms rt ON rt.all_products
),
grooved_candidates AS (
  SELECT DISTINCT p.id AS product_id, rt.term_key, rt.label_en, rt.label_fa, rt.label_ar
  FROM products p
  JOIN rule_terms rt ON rt.term_key = 'grooved'
  WHERE EXISTS (
    SELECT 1
    FROM UNNEST(COALESCE(p.finishes, '{}'::text[])) AS existing_finish
    WHERE LOWER(REGEXP_REPLACE(TRIM(existing_finish), '[[:space:]]+', ' ', 'g')) = ANY(rt.equivalent_labels)
  )
),
candidate_pairs AS (
  SELECT * FROM category_candidates
  UNION
  SELECT * FROM all_product_candidates
  UNION
  SELECT * FROM grooved_candidates
),
missing_pairs AS (
  SELECT cp.*
  FROM candidate_pairs cp
  JOIN products p ON p.id = cp.product_id
  WHERE NOT EXISTS (
    SELECT 1
    FROM UNNEST(COALESCE(p.finishes, '{}'::text[])) AS existing_finish
    WHERE LOWER(REGEXP_REPLACE(TRIM(existing_finish), '[[:space:]]+', '', 'g')) IN (
      LOWER(REGEXP_REPLACE(TRIM(cp.label_en), '[[:space:]]+', '', 'g')),
      LOWER(REGEXP_REPLACE(TRIM(cp.label_fa), '[[:space:]]+', '', 'g')),
      LOWER(REGEXP_REPLACE(TRIM(cp.label_ar), '[[:space:]]+', '', 'g'))
    )
  )
),
additions AS (
  SELECT product_id, ARRAY_AGG(label_fa ORDER BY term_key) AS labels
  FROM missing_pairs
  GROUP BY product_id
)
UPDATE products p
SET
  finishes = COALESCE(p.finishes, '{}'::text[]) || additions.labels,
  updated_at = NOW()
FROM additions
WHERE p.id = additions.product_id;

WITH canonical_terms AS (
  SELECT * FROM (VALUES
    ('pre-cut', 'Pre-Cut', 'پری‌کات', 'مقطّع مسبقاً'),
    ('chiseled-machine-hand', 'Chiseled (Machine and Hand)', 'تیشه‌ای (ماشینی و دستی)', 'منقور بالإزميل (آلي ويدوي)'),
    ('grooved', 'Grooved', 'شیار', 'محزّز'),
    ('flamed', 'Flamed', 'فلیم', 'معالج باللهب'),
    ('leathered-bush-hammer', 'Leathered Bush Hammer', 'بوش همر چرمی', 'بوش هامر جلدي'),
    ('honed', 'Honed', 'هوند (نسابیده)', 'مصقول مطفي')
  ) AS v(term_key, label_en, label_fa, label_ar)
),
mapped_finishes AS (
  SELECT
    p.id AS product_id,
    existing_finish.ordinality,
    COALESCE((
      SELECT ct.label_fa
      FROM canonical_terms ct
      WHERE LOWER(REGEXP_REPLACE(TRIM(existing_finish.label), '[[:space:]]+', '', 'g')) IN (
        LOWER(REGEXP_REPLACE(TRIM(ct.label_en), '[[:space:]]+', '', 'g')),
        LOWER(REGEXP_REPLACE(TRIM(ct.label_fa), '[[:space:]]+', '', 'g')),
        LOWER(REGEXP_REPLACE(TRIM(ct.label_ar), '[[:space:]]+', '', 'g'))
      )
      LIMIT 1
    ), existing_finish.label) AS label
  FROM products p
  CROSS JOIN LATERAL UNNEST(COALESCE(p.finishes, '{}'::text[])) WITH ORDINALITY AS existing_finish(label, ordinality)
),
deduped_finishes AS (
  SELECT DISTINCT ON (product_id, LOWER(REGEXP_REPLACE(TRIM(label), '[[:space:]]+', '', 'g')))
    product_id,
    label,
    ordinality
  FROM mapped_finishes
  WHERE TRIM(label) <> ''
  ORDER BY product_id, LOWER(REGEXP_REPLACE(TRIM(label), '[[:space:]]+', '', 'g')), ordinality
),
rebuilt_finishes AS (
  SELECT product_id, ARRAY_AGG(label ORDER BY ordinality) AS finishes
  FROM deduped_finishes
  GROUP BY product_id
)
UPDATE products p
SET
  finishes = rebuilt_finishes.finishes,
  updated_at = NOW()
FROM rebuilt_finishes
WHERE p.id = rebuilt_finishes.product_id
  AND p.finishes IS DISTINCT FROM rebuilt_finishes.finishes;

WITH rule_terms AS (
  SELECT * FROM (VALUES
    ('pre-cut', ARRAY['granite', 'marble', 'crystal', 'basalt']::text[], FALSE, ARRAY[]::text[]),
    ('chiseled-machine-hand', ARRAY['marble', 'travertine', 'granite']::text[], FALSE, ARRAY[]::text[]),
    ('leathered-bush-hammer', ARRAY['granite', 'travertine', 'basalt']::text[], FALSE, ARRAY[]::text[]),
    ('flamed', ARRAY['granite']::text[], FALSE, ARRAY[]::text[]),
    ('honed', ARRAY[]::text[], TRUE, ARRAY[]::text[]),
    ('grooved', ARRAY[]::text[], FALSE, ARRAY['lineal', 'scratch', 'لینئال', 'اسکرچ', 'لينيل', 'تشطيب مخدوش']::text[])
  ) AS v(term_key, category_slugs, all_products, equivalent_labels)
),
product_category_slugs AS (
  SELECT DISTINCT p.id AS product_id, c.slug
  FROM products p
  JOIN categories c ON c.id = p.main_category_id
  UNION
  SELECT DISTINCT pc.product_id, c.slug
  FROM product_categories pc
  JOIN categories c ON c.id = pc.category_id
),
category_candidates AS (
  SELECT DISTINCT pcs.product_id, rt.term_key
  FROM product_category_slugs pcs
  JOIN rule_terms rt ON pcs.slug = ANY(rt.category_slugs)
  WHERE NOT rt.all_products
    AND CARDINALITY(rt.equivalent_labels) = 0
),
all_product_candidates AS (
  SELECT p.id AS product_id, rt.term_key
  FROM products p
  JOIN rule_terms rt ON rt.all_products
),
grooved_candidates AS (
  SELECT DISTINCT p.id AS product_id, rt.term_key
  FROM products p
  JOIN rule_terms rt ON rt.term_key = 'grooved'
  WHERE EXISTS (
    SELECT 1
    FROM UNNEST(COALESCE(p.finishes, '{}'::text[])) AS existing_finish
    WHERE LOWER(REGEXP_REPLACE(TRIM(existing_finish), '[[:space:]]+', ' ', 'g')) = ANY(rt.equivalent_labels)
  )
),
candidate_pairs AS (
  SELECT * FROM category_candidates
  UNION
  SELECT * FROM all_product_candidates
  UNION
  SELECT * FROM grooved_candidates
)
INSERT INTO product_term_links (product_id, term_id)
SELECT cp.product_id, t.id
FROM candidate_pairs cp
JOIN product_terms t ON t.taxonomy = 'finishes' AND t.term_key = cp.term_key
ON CONFLICT DO NOTHING;

-- Remove limestone stone_type from existing DBs (and its product links) so it does not appear in UI filters.
DELETE FROM product_term_links
WHERE term_id IN (
  SELECT id FROM product_terms WHERE taxonomy = 'stone_type' AND term_key = 'limestone'
);

DELETE FROM product_terms
WHERE taxonomy = 'stone_type' AND term_key = 'limestone';
