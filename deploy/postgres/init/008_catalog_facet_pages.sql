CREATE TABLE IF NOT EXISTS catalog_facet_pages (
  id SERIAL PRIMARY KEY,
  category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  term_id INT NOT NULL REFERENCES product_terms(id) ON DELETE CASCADE,
  title_en VARCHAR(255),
  title_fa VARCHAR(255),
  title_ar VARCHAR(255),
  description_en TEXT,
  description_fa TEXT,
  description_ar TEXT,
  h1_en VARCHAR(255),
  h1_fa VARCHAR(255),
  h1_ar VARCHAR(255),
  intro_en TEXT,
  intro_fa TEXT,
  intro_ar TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  is_indexable BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ,
  UNIQUE (category_id, term_id)
);

CREATE INDEX IF NOT EXISTS idx_catalog_facet_pages_lookup
  ON catalog_facet_pages (category_id, term_id, is_active, is_indexable);
