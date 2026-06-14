package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"sangehassan/back/internal/config"
)

func NewDB(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := ensureMediaColumns(db); err != nil {
		return nil, err
	}
	if err := ensureProductTermMetadata(db); err != nil {
		return nil, err
	}
	if err := ensureCatalogMetadata(db); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureCatalogMetadata(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	statements := []string{
		`ALTER TABLE IF EXISTS categories
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
		ADD COLUMN IF NOT EXISTS is_indexable BOOLEAN NOT NULL DEFAULT TRUE`,
		`ALTER TABLE IF EXISTS products
		ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
		ADD COLUMN IF NOT EXISTS is_indexable BOOLEAN NOT NULL DEFAULT TRUE`,
		`ALTER TABLE IF EXISTS product_terms
		ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
		ADD COLUMN IF NOT EXISTS is_indexable BOOLEAN NOT NULL DEFAULT TRUE`,
		`CREATE INDEX IF NOT EXISTS idx_categories_catalog ON categories (is_active, is_indexable, slug)`,
		`CREATE INDEX IF NOT EXISTS idx_products_catalog_category ON products (main_category_id, is_active, is_indexable)`,
		`CREATE INDEX IF NOT EXISTS idx_product_categories_category_product ON product_categories (category_id, product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_product_term_links_term_product ON product_term_links (term_id, product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_product_terms_catalog ON product_terms (taxonomy, term_key, is_active, is_indexable)`,
		`CREATE TABLE IF NOT EXISTS catalog_facet_pages (
			id SERIAL PRIMARY KEY,
			category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			term_id INT NOT NULL REFERENCES product_terms(id) ON DELETE CASCADE,
			title_en VARCHAR(255), title_fa VARCHAR(255), title_ar VARCHAR(255),
			description_en TEXT, description_fa TEXT, description_ar TEXT,
			h1_en VARCHAR(255), h1_fa VARCHAR(255), h1_ar VARCHAR(255),
			intro_en TEXT, intro_fa TEXT, intro_ar TEXT,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			is_indexable BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ,
			UNIQUE (category_id, term_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_catalog_facet_pages_lookup ON catalog_facet_pages (category_id, term_id, is_active, is_indexable)`,
		`INSERT INTO product_terms (taxonomy, term_key, label_en, label_fa, label_ar)
		VALUES
			('availability', 'in-stock', 'In stock', 'موجود', 'متوفر'),
			('availability', 'available-on-order', 'Available on order', 'قابل سفارش', 'متوفر حسب الطلب'),
			('availability', 'limited', 'Limited availability', 'موجودی محدود', 'توفر محدود'),
			('availability', 'unavailable', 'Unavailable', 'ناموجود', 'غير متوفر')
		ON CONFLICT (taxonomy, term_key) DO NOTHING`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func ensureMediaColumns(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	statements := []string{
		`ALTER TABLE IF EXISTS products ADD COLUMN IF NOT EXISTS video_url TEXT`,
		`ALTER TABLE IF EXISTS projects ADD COLUMN IF NOT EXISTS video_url TEXT`,
		`CREATE TABLE IF NOT EXISTS project_products (
			project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (project_id, product_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_project_products_project ON project_products(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_project_products_product ON project_products(product_id)`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func ensureProductTermMetadata(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	statements := []string{
		`ALTER TABLE IF EXISTS products
		ADD COLUMN IF NOT EXISTS aliases TEXT[] NOT NULL DEFAULT '{}'::text[],
		ADD COLUMN IF NOT EXISTS variants TEXT[] NOT NULL DEFAULT '{}'::text[],
		ADD COLUMN IF NOT EXISTS mines TEXT[] NOT NULL DEFAULT '{}'::text[],
		ADD COLUMN IF NOT EXISTS finishes TEXT[] NOT NULL DEFAULT '{}'::text[]`,
		`CREATE TABLE IF NOT EXISTS product_terms (
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
		)`,
		`ALTER TABLE IF EXISTS product_terms ADD COLUMN IF NOT EXISTS link_url TEXT`,
		`CREATE TABLE IF NOT EXISTS product_term_links (
			product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			term_id INT NOT NULL REFERENCES product_terms(id) ON DELETE CASCADE,
			PRIMARY KEY (product_id, term_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_product_terms_taxonomy ON product_terms(taxonomy)`,
		`CREATE INDEX IF NOT EXISTS idx_product_term_links_product ON product_term_links(product_id)`,
		`INSERT INTO product_terms (taxonomy, term_key, label_en, label_fa, label_ar)
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
			updated_at = NOW()`,
		`WITH finish_products AS (
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
		  AND COALESCE(t.link_url, '') IS DISTINCT FROM matched_finish_links.link_url`,
		`WITH source_finish_links AS (
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
		ON CONFLICT DO NOTHING`,
		`WITH source_terms AS (
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
		ON CONFLICT (taxonomy, term_key) DO NOTHING`,
		`UPDATE product_terms
		SET
			label_en = CASE WHEN label_fa ~ '[؀-ۿ]' THEN '' ELSE label_en END,
			label_ar = '',
			label_fa = CASE WHEN label_fa ~ '[؀-ۿ]' THEN label_fa ELSE '' END,
			updated_at = NOW()
		WHERE taxonomy IN ('variants', 'mines')
		  AND term_key LIKE 'legacy-%'
		  AND label_en = label_fa
		  AND label_fa = label_ar`,
		`WITH source_links AS (
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
		ON CONFLICT DO NOTHING`,
		`INSERT INTO product_terms (taxonomy, term_key, label_en, label_fa, label_ar)
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
			updated_at = NOW()`,
		`WITH rule_terms AS (
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
		WHERE p.id = additions.product_id`,
		`WITH canonical_terms AS (
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
		  AND p.finishes IS DISTINCT FROM rebuilt_finishes.finishes`,
		`WITH rule_terms AS (
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
		ON CONFLICT DO NOTHING`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
