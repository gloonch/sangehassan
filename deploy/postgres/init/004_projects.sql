CREATE TABLE IF NOT EXISTS projects (
  id BIGSERIAL PRIMARY KEY,
  description TEXT,
  description_en TEXT,
  description_fa TEXT,
  description_ar TEXT,
  cover_image_url TEXT NOT NULL,
  video_url TEXT,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

ALTER TABLE projects
ADD COLUMN IF NOT EXISTS description_en TEXT,
ADD COLUMN IF NOT EXISTS description_fa TEXT,
ADD COLUMN IF NOT EXISTS description_ar TEXT,
ADD COLUMN IF NOT EXISTS video_url TEXT;

UPDATE projects
SET description_en = COALESCE(description_en, description)
WHERE description_en IS NULL AND description IS NOT NULL;

CREATE TABLE IF NOT EXISTS project_images (
  id BIGSERIAL PRIMARY KEY,
  project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  image_url TEXT NOT NULL,
  position INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(project_id, image_url)
);

CREATE INDEX IF NOT EXISTS idx_projects_sort_order ON projects (sort_order, id DESC);
CREATE INDEX IF NOT EXISTS idx_project_images_project_position ON project_images (project_id, position);
