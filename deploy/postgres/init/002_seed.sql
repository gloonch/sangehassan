INSERT INTO categories (title_en, title_fa, title_ar, slug)
VALUES
  ('Marble', 'سنگ مرمر', 'رخام', 'marble'),
  ('Granite', 'سنگ گرانیت', 'جرانيت', 'granite'),
  ('Travertine', 'سنگ تراورتن', 'ترافرتين', 'travertine'),
  ('Quartz', 'سنگ کوارتز', 'كوارتز', 'quartz'),
  ('Chinese Marble', 'سنگ چینی', 'حجر صيني', 'chinese-marble'),
  ('Antique Stone', 'سنگ آنتیک', 'حجر عتيق', 'antique-stone')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO admin_users (username, password_hash)
VALUES ('admin', '$2y$05$.4WCZ0rj7SGvUpfGQ1n.Qen3qCeQpQC5He0APLBBHERraPO3F7h.u')
ON CONFLICT (username) DO NOTHING;

INSERT INTO products (title_en, title_fa, title_ar, description, price, image_url, category_id, is_popular)
SELECT 'Abbas Abad Travertine', 'تراورتن عباس آباد', 'ترافرتين عباس آباد', 'Cream travertine with soft motion.', 150000, '/images/travertine.jpg', c.id, TRUE
FROM categories c
WHERE c.slug = 'travertine'
  AND NOT EXISTS (SELECT 1 FROM products WHERE title_en = 'Abbas Abad Travertine');

INSERT INTO products (title_en, title_fa, title_ar, description, price, image_url, category_id, is_popular)
SELECT 'Natanzi Black Granite', 'گرانیت مشکی نطنز', 'جرانيت نطنز الأسود', 'Deep black granite with high durability.', 210000, '/images/granite.jpg', c.id, TRUE
FROM categories c
WHERE c.slug = 'granite'
  AND NOT EXISTS (SELECT 1 FROM products WHERE title_en = 'Natanzi Black Granite');

INSERT INTO products (title_en, title_fa, title_ar, description, price, image_url, category_id, is_popular)
SELECT 'White Marble', 'مرمر سفید', 'رخام أبيض', 'Classic white marble for bright interiors.', 260000, '/images/marble.jpg', c.id, TRUE
FROM categories c
WHERE c.slug = 'marble'
  AND NOT EXISTS (SELECT 1 FROM products WHERE title_en = 'White Marble');

INSERT INTO templates (name, image_url, is_active)
VALUES
  ('Triangle', '/images/templates/triangle.png', TRUE),
  ('Rectangle', '/images/templates/rectangle.png', TRUE),
  ('Circle', '/images/templates/circle.png', TRUE)
ON CONFLICT DO NOTHING;
