ALTER TABLE blogs
  ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'draft',
  ADD COLUMN IF NOT EXISTS author_name VARCHAR(160) NOT NULL DEFAULT 'SangeHassan',
  ADD COLUMN IF NOT EXISTS og_image_url TEXT,
  ADD COLUMN IF NOT EXISTS category_slug VARCHAR(120),
  ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}'::TEXT[],
  ADD COLUMN IF NOT EXISTS is_featured BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS blog_translations (
  id BIGSERIAL PRIMARY KEY,
  blog_id INT NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
  locale VARCHAR(2) NOT NULL CHECK (locale IN ('fa', 'en', 'ar')),
  title VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL,
  excerpt TEXT,
  content_json JSONB NOT NULL DEFAULT '{"type":"doc","content":[]}'::JSONB,
  content_html TEXT NOT NULL DEFAULT '',
  seo_title VARCHAR(255),
  seo_description TEXT,
  canonical_url TEXT,
  robots VARCHAR(40) NOT NULL DEFAULT 'index,follow',
  translation_status VARCHAR(20) NOT NULL DEFAULT 'draft',
  featured_image_alt TEXT,
  og_image_alt TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ,
  UNIQUE (blog_id, locale),
  UNIQUE (locale, slug),
  CHECK (translation_status IN ('draft', 'published'))
);

CREATE TABLE IF NOT EXISTS blog_slug_redirects (
  id BIGSERIAL PRIMARY KEY,
  blog_id INT NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
  locale VARCHAR(2) NOT NULL CHECK (locale IN ('fa', 'en', 'ar')),
  old_slug VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (locale, old_slug)
);

CREATE INDEX IF NOT EXISTS idx_blogs_publication
  ON blogs (status, published_at, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_blog_translations_public
  ON blog_translations (locale, translation_status, slug);
CREATE INDEX IF NOT EXISTS idx_blog_slug_redirects_lookup
  ON blog_slug_redirects (locale, old_slug);

INSERT INTO blog_translations (
  blog_id, locale, title, slug, excerpt, content_html, translation_status
)
SELECT
  b.id,
  'en',
  b.title,
  'legacy-blog-' || b.id,
  COALESCE(b.excerpt, ''),
  COALESCE(b.content, ''),
  CASE WHEN b.status = 'published' THEN 'published' ELSE 'draft' END
FROM blogs b
WHERE NOT EXISTS (
  SELECT 1 FROM blog_translations bt WHERE bt.blog_id = b.id
)
ON CONFLICT DO NOTHING;
