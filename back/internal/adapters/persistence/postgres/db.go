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

	return db, nil
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
		ON CONFLICT (taxonomy, term_key) DO NOTHING`,
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
		ON CONFLICT DO NOTHING`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
