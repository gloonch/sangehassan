-- Stone sample request schema
-- Idempotent: safe to run against existing volumes.

ALTER TABLE products
ADD COLUMN IF NOT EXISTS sample_available BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_products_sample_catalog
  ON products (sample_available, is_active, is_indexable, main_category_id);

CREATE TABLE IF NOT EXISTS user_addresses (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  label TEXT,
  address_text TEXT NOT NULL,
  city TEXT,
  province TEXT,
  postal_code TEXT,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, address_text)
);

CREATE INDEX IF NOT EXISTS idx_user_addresses_user
  ON user_addresses (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS user_contact_numbers (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  label TEXT,
  phone TEXT NOT NULL,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, phone)
);

CREATE INDEX IF NOT EXISTS idx_user_contact_numbers_user
  ON user_contact_numbers (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS stone_sample_requests (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'PENDING',
  box_count INT NOT NULL,
  stone_count INT NOT NULL,
  price_per_box_toman BIGINT NOT NULL,
  total_price_toman BIGINT NOT NULL,
  shipping_method TEXT NOT NULL,
  address_id BIGINT REFERENCES user_addresses(id) ON DELETE SET NULL,
  phone_id BIGINT REFERENCES user_contact_numbers(id) ON DELETE SET NULL,
  address_snapshot TEXT NOT NULL,
  phone_snapshot TEXT NOT NULL,
  admin_note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stone_sample_requests_user
  ON stone_sample_requests (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stone_sample_requests_status
  ON stone_sample_requests (status, created_at DESC);

CREATE TABLE IF NOT EXISTS stone_sample_request_items (
  id BIGSERIAL PRIMARY KEY,
  request_id BIGINT NOT NULL REFERENCES stone_sample_requests(id) ON DELETE CASCADE,
  product_id INT REFERENCES products(id) ON DELETE SET NULL,
  box_index INT NOT NULL,
  slot_index INT NOT NULL,
  product_title_en TEXT NOT NULL,
  product_title_fa TEXT NOT NULL,
  product_title_ar TEXT NOT NULL,
  product_slug TEXT NOT NULL,
  product_image_url TEXT,
  category_title_en TEXT,
  category_title_fa TEXT,
  category_title_ar TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(request_id, box_index, slot_index)
);

CREATE INDEX IF NOT EXISTS idx_stone_sample_request_items_request
  ON stone_sample_request_items (request_id, box_index, slot_index);

CREATE TABLE IF NOT EXISTS stone_sample_request_status_history (
  id BIGSERIAL PRIMARY KEY,
  request_id BIGINT NOT NULL REFERENCES stone_sample_requests(id) ON DELETE CASCADE,
  from_status TEXT,
  to_status TEXT NOT NULL,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  admin_note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stone_sample_status_history_request
  ON stone_sample_request_status_history (request_id, created_at);
