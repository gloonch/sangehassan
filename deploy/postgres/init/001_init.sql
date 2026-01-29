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
