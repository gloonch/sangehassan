-- Additional metadata for stone products
-- Idempotent: uses IF NOT EXISTS + non-destructive indexes

ALTER TABLE products
ADD COLUMN IF NOT EXISTS aliases TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS variants TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS mines TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS finishes TEXT[] NOT NULL DEFAULT '{}'::text[];

-- Prevent duplicate base products within the same category (case-insensitive on Persian name)
-- Exclude legacy category_id = 5 slab dataset to avoid blocking existing rows.
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_category_title_fa
  ON products (main_category_id, lower(title_fa))
  WHERE main_category_id IS NOT NULL AND main_category_id <> 5;

CREATE INDEX IF NOT EXISTS idx_products_title_fa_lower ON products (lower(title_fa));
