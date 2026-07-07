-- Stone ads / deal requests schema
-- Idempotent: all objects created with IF NOT EXISTS and seeds guarded by NOT EXISTS

-- Listings: stone sale offers tied to a real catalog product.
CREATE TABLE IF NOT EXISTS listings (
  id BIGSERIAL PRIMARY KEY,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  product_id INT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
  title TEXT,
  stone_type TEXT,
  form TEXT,
  tonnage NUMERIC(12, 2),
  province TEXT,
  city TEXT,
  price_amount NUMERIC(14, 2),
  price_unit TEXT,
  description TEXT,
  extra_props JSONB NOT NULL DEFAULT '{}'::jsonb,
  status TEXT NOT NULL DEFAULT 'ACTIVE',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listings_status_created ON listings (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_listings_product ON listings (product_id);
CREATE INDEX IF NOT EXISTS idx_listings_stone_type ON listings (stone_type);
CREATE INDEX IF NOT EXISTS idx_listings_form ON listings (form);
CREATE INDEX IF NOT EXISTS idx_listings_tonnage ON listings (tonnage);
CREATE INDEX IF NOT EXISTS idx_listings_location ON listings (province, city);
CREATE INDEX IF NOT EXISTS idx_listings_price ON listings (price_amount);

-- Listing media (gallery)
-- Reserved for future listing-specific/admin media; user-submitted offers
-- currently render the selected product cover instead of uploaded images.
CREATE TABLE IF NOT EXISTS listing_images (
  id BIGSERIAL PRIMARY KEY,
  listing_id BIGINT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
  image_url TEXT NOT NULL,
  position INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(listing_id, image_url)
);

CREATE INDEX IF NOT EXISTS idx_listing_images_listing_position ON listing_images (listing_id, position);

-- Deal requests submitted by buyers
CREATE TABLE IF NOT EXISTS deal_requests (
  id BIGSERIAL PRIMARY KEY,
  listing_id BIGINT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
  buyer_id UUID REFERENCES users(id) ON DELETE SET NULL,
  seller_id UUID REFERENCES users(id) ON DELETE SET NULL,
  request_type TEXT NOT NULL, -- INSPECTION | PURCHASE | BOTH
  buyer_note TEXT,
  status TEXT NOT NULL DEFAULT 'PENDING_REVIEW',
  meeting_at TIMESTAMPTZ,
  admin_note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deal_requests_buyer ON deal_requests (buyer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_deal_requests_status ON deal_requests (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_deal_requests_listing ON deal_requests (listing_id);

-- Status history for requests
CREATE TABLE IF NOT EXISTS deal_status_history (
  id BIGSERIAL PRIMARY KEY,
  request_id BIGINT NOT NULL REFERENCES deal_requests(id) ON DELETE CASCADE,
  from_status TEXT,
  to_status TEXT NOT NULL,
  created_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deal_status_history_request ON deal_status_history (request_id, created_at);

-- Simple notifications table for dashboard alerts
CREATE TABLE IF NOT EXISTS notifications (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  read_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications (user_id, read_at);

CREATE TABLE IF NOT EXISTS listing_product_requests (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  query TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'NEW',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listing_product_requests_created ON listing_product_requests (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_listing_product_requests_status ON listing_product_requests (status, created_at DESC);
