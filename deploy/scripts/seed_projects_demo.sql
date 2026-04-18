CREATE TABLE IF NOT EXISTS projects (
  id BIGSERIAL PRIMARY KEY,
  description TEXT,
  description_en TEXT,
  description_fa TEXT,
  description_ar TEXT,
  cover_image_url TEXT NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

ALTER TABLE projects
ADD COLUMN IF NOT EXISTS description_en TEXT,
ADD COLUMN IF NOT EXISTS description_fa TEXT,
ADD COLUMN IF NOT EXISTS description_ar TEXT;

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

TRUNCATE TABLE project_images, projects RESTART IDENTITY CASCADE;

INSERT INTO projects (description, description_en, description_fa, description_ar, cover_image_url, sort_order)
SELECT
  format(
    'پروژه نمونه شماره %s - اجرای سنگ در نمای ساختمان و فضای داخلی با تمرکز بر هماهنگی رنگ و بافت متریال.',
    gs
  ),
  format(
    'Sample project %s - facade and interior stone execution with focus on coordinated tones and textures.',
    gs
  ),
  format(
    'پروژه نمونه شماره %s - اجرای سنگ در نما و فضای داخلی با تمرکز بر هماهنگی رنگ و بافت.',
    gs
  ),
  format(
    'مشروع تجريبي رقم %s - تنفيذ الحجر للواجهات والداخل مع تركيز على تناسق اللون والملمس.',
    gs
  ),
  CASE (gs % 3)
    WHEN 1 THEN '/images/projects/cover-1.jpg'
    WHEN 2 THEN '/images/projects/cover-2.jpg'
    ELSE '/images/projects/cover-3.jpg'
  END,
  gs
FROM generate_series(1, 20) AS gs;

INSERT INTO project_images (project_id, image_url, position)
SELECT
  p.id,
  img.image_url,
  img.position
FROM projects p
JOIN LATERAL (
  VALUES
    ('/images/projects/gallery-1.jpg', 0),
    ('/images/projects/gallery-2.jpg', 1),
    ('/images/projects/gallery-3.jpg', 2),
    ('/images/projects/gallery-4.png', 3),
    ('/images/projects/gallery-5.png', 4)
) AS img(image_url, position)
ON TRUE;
