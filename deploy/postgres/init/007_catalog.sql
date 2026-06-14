-- Locale-aware product catalog and faceted navigation metadata.
-- Additive and safe to run against existing volumes.

ALTER TABLE categories
ADD COLUMN IF NOT EXISTS image_url TEXT,
ADD COLUMN IF NOT EXISTS intro_en TEXT,
ADD COLUMN IF NOT EXISTS intro_fa TEXT,
ADD COLUMN IF NOT EXISTS intro_ar TEXT,
ADD COLUMN IF NOT EXISTS seo_title_en VARCHAR(255),
ADD COLUMN IF NOT EXISTS seo_title_fa VARCHAR(255),
ADD COLUMN IF NOT EXISTS seo_title_ar VARCHAR(255),
ADD COLUMN IF NOT EXISTS seo_description_en TEXT,
ADD COLUMN IF NOT EXISTS seo_description_fa TEXT,
ADD COLUMN IF NOT EXISTS seo_description_ar TEXT,
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS is_indexable BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE products
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS is_indexable BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE product_terms
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS is_indexable BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_categories_catalog
  ON categories (is_active, is_indexable, slug);
CREATE INDEX IF NOT EXISTS idx_products_catalog_category
  ON products (main_category_id, is_active, is_indexable);
CREATE INDEX IF NOT EXISTS idx_product_categories_category_product
  ON product_categories (category_id, product_id);
CREATE INDEX IF NOT EXISTS idx_product_term_links_term_product
  ON product_term_links (term_id, product_id);
CREATE INDEX IF NOT EXISTS idx_product_terms_catalog
  ON product_terms (taxonomy, term_key, is_active, is_indexable);

INSERT INTO product_terms (taxonomy, term_key, label_en, label_fa, label_ar)
VALUES
  ('availability', 'in-stock', 'In stock', 'موجود', 'متوفر'),
  ('availability', 'available-on-order', 'Available on order', 'قابل سفارش', 'متوفر حسب الطلب'),
  ('availability', 'limited', 'Limited availability', 'موجودی محدود', 'توفر محدود'),
  ('availability', 'unavailable', 'Unavailable', 'ناموجود', 'غير متوفر')
ON CONFLICT (taxonomy, term_key) DO NOTHING;
