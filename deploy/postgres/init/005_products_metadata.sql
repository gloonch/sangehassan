-- Additional metadata for stone products
-- Idempotent: uses IF NOT EXISTS + non-destructive indexes

ALTER TABLE products
ADD COLUMN IF NOT EXISTS aliases TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS variants TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS mines TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS finishes TEXT[] NOT NULL DEFAULT '{}'::text[],
ADD COLUMN IF NOT EXISTS video_url TEXT;

ALTER TABLE product_terms
ADD COLUMN IF NOT EXISTS link_url TEXT;

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
  UNION ALL
  SELECT p.id AS product_id, 'finishes' AS taxonomy, LOWER(REGEXP_REPLACE(LEFT(TRIM(value), 255), '[[:space:]]+', ' ', 'g')) AS normalized_label
  FROM products p
  CROSS JOIN LATERAL UNNEST(COALESCE(p.finishes, '{}'::text[])) AS value
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

-- Prevent duplicate base products within the same category (case-insensitive on Persian name)
-- Exclude legacy category_id = 5 slab dataset to avoid blocking existing rows.
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_category_title_fa
  ON products (main_category_id, lower(title_fa))
  WHERE main_category_id IS NOT NULL AND main_category_id <> 5;

CREATE INDEX IF NOT EXISTS idx_products_title_fa_lower ON products (lower(title_fa));
