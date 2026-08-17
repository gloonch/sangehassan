-- Persistent, administrator-managed product ordering shared by the panel and public catalog.
-- The initial ranking preserves the public ordering used before this migration.

ALTER TABLE products
  ADD COLUMN IF NOT EXISTS display_order INTEGER;

WITH ranked_products AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY is_popular DESC, id) - 1 AS initial_order
  FROM products
)
UPDATE products p
SET display_order = ranked_products.initial_order
FROM ranked_products
WHERE p.id = ranked_products.id
  AND p.display_order IS NULL;

ALTER TABLE products
  ALTER COLUMN display_order SET DEFAULT 1000000000,
  ALTER COLUMN display_order SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'products_display_order_nonnegative'
  ) THEN
    ALTER TABLE products
      ADD CONSTRAINT products_display_order_nonnegative CHECK (display_order >= 0);
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_products_display_order
  ON products(display_order, id);

INSERT INTO schema_migrations(version, migration_name)
VALUES (20, 'product_display_order')
ON CONFLICT(version) DO UPDATE SET migration_name = EXCLUDED.migration_name;
