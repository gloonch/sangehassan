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

INSERT INTO content_section_images (section_id, image_url, position)
SELECT id, 'https://images.pexels.com/photos/1664011/pexels-photo-1664011.jpeg?auto=compress&cs=tinysrgb&dpr=1&w=1200', 0
FROM content_sections
WHERE page = 'home' AND section_key = 'blocks'
ON CONFLICT DO NOTHING;

INSERT INTO content_section_images (section_id, image_url, position)
SELECT id, 'https://images.pexels.com/photos/13552642/pexels-photo-13552642.jpeg?auto=compress&cs=tinysrgb&dpr=1&w=1200', 1
FROM content_sections
WHERE page = 'home' AND section_key = 'blocks'
ON CONFLICT DO NOTHING;

INSERT INTO content_section_images (section_id, image_url, position)
SELECT id, 'https://images.pexels.com/photos/9041501/pexels-photo-9041501.jpeg?auto=compress&cs=tinysrgb&dpr=1&w=1200', 0
FROM content_sections
WHERE page = 'home' AND section_key = 'finished'
ON CONFLICT DO NOTHING;

INSERT INTO content_section_images (section_id, image_url, position)
SELECT id, 'https://images.pexels.com/photos/842562/pexels-photo-842562.jpeg?auto=compress&cs=tinysrgb&dpr=1&w=1200', 1
FROM content_sections
WHERE page = 'home' AND section_key = 'finished'
ON CONFLICT DO NOTHING;

INSERT INTO content_section_images (section_id, image_url, position)
SELECT id, 'https://images.pexels.com/photos/7195950/pexels-photo-7195950.jpeg?auto=compress&cs=tinysrgb&dpr=1&w=1200', 2
FROM content_sections
WHERE page = 'home' AND section_key = 'finished'
ON CONFLICT DO NOTHING;

INSERT INTO content_section_images (section_id, image_url, position)
SELECT id, 'https://images.pexels.com/photos/1664011/pexels-photo-1664011.jpeg?auto=compress&cs=tinysrgb&dpr=1&w=1200', 0
FROM content_sections
WHERE page = 'blocks' AND section_key = 'hero'
ON CONFLICT DO NOTHING;

INSERT INTO content_section_images (section_id, image_url, position)
SELECT id, 'https://images.pexels.com/photos/13552642/pexels-photo-13552642.jpeg?auto=compress&cs=tinysrgb&dpr=1&w=1200', 1
FROM content_sections
WHERE page = 'blocks' AND section_key = 'hero'
ON CONFLICT DO NOTHING;
