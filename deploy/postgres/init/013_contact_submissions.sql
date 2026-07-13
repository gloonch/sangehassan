-- Footer contact form submissions

CREATE TABLE IF NOT EXISTS contact_submissions (
  id BIGSERIAL PRIMARY KEY,
  full_name TEXT NOT NULL,
  email TEXT,
  country_iso TEXT NOT NULL,
  country_code TEXT NOT NULL,
  phone_number TEXT NOT NULL,
  phone_e164 TEXT NOT NULL,
  message TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT 'footer',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contact_submissions_created
  ON contact_submissions (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_contact_submissions_phone
  ON contact_submissions (phone_e164);
