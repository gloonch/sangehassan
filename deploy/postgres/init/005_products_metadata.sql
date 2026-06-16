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
  WHERE NOT (
    rt.term_key = 'honed'
    AND EXISTS (
      SELECT 1
      FROM product_category_slugs pcs
      WHERE pcs.product_id = p.id
        AND pcs.slug = 'accessories'
    )
  )
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
  WHERE NOT (
    rt.term_key = 'honed'
    AND EXISTS (
      SELECT 1
      FROM product_category_slugs pcs
      WHERE pcs.product_id = p.id
        AND pcs.slug = 'accessories'
    )
  )
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

-- Prevent duplicate base products within the same category (case-insensitive on Persian name)
-- Exclude legacy category_id = 5 slab dataset to avoid blocking existing rows.
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_category_title_fa
  ON products (main_category_id, lower(title_fa))
  WHERE main_category_id IS NOT NULL AND main_category_id <> 5;

CREATE INDEX IF NOT EXISTS idx_products_title_fa_lower ON products (lower(title_fa));
