CREATE TABLE IF NOT EXISTS team_members (
  id BIGSERIAL PRIMARY KEY,
  name_en VARCHAR(255) NOT NULL,
  name_fa VARCHAR(255) NOT NULL,
  name_ar VARCHAR(255) NOT NULL,
  role_en VARCHAR(255) NOT NULL,
  role_fa VARCHAR(255) NOT NULL,
  role_ar VARCHAR(255) NOT NULL,
  bio_en TEXT,
  bio_fa TEXT,
  bio_ar TEXT,
  photo_url TEXT,
  linkedin_url TEXT,
  order_index INT NOT NULL DEFAULT 0,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_team_members_public
  ON team_members (is_active, order_index, id);
