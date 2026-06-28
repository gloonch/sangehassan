-- Floor offer prices for popular products by stone category.
-- Safe to rerun: it only updates products marked as popular in the mapped categories.

WITH offer_prices (category_slug, category_label_fa, offer_price, priority) AS (
  VALUES
    ('travertine', 'تراورتن', 450000::NUMERIC(12, 2), 1),
    ('granite', 'گرانیت', 320000::NUMERIC(12, 2), 2),
    ('onyx', 'مرمر', 490000::NUMERIC(12, 2), 3),
    ('marble', 'مرمریت', 450000::NUMERIC(12, 2), 4),
    ('basalt', 'بازالت', 350000::NUMERIC(12, 2), 5),
    ('crystal', 'چینی کریستال', 450000::NUMERIC(12, 2), 6)
),
matched_products AS (
  SELECT DISTINCT ON (p.id)
    p.id,
    op.category_label_fa,
    op.offer_price
  FROM products p
  JOIN LATERAL (
    SELECT c.id, c.slug
    FROM categories c
    WHERE c.id = p.main_category_id

    UNION

    SELECT c.id, c.slug
    FROM product_categories pc
    JOIN categories c ON c.id = pc.category_id
    WHERE pc.product_id = p.id
  ) product_category ON TRUE
  JOIN offer_prices op ON op.category_slug = product_category.slug
  WHERE p.is_popular = TRUE
  ORDER BY
    p.id,
    CASE WHEN product_category.id = p.main_category_id THEN 0 ELSE 1 END,
    op.priority
)
UPDATE products p
SET
  price = matched_products.offer_price,
  price_html = FORMAT(
    'Offer | کف قیمت %s: %s تومان. به‌دلیل نوسان لحظه‌ای بازار، برای قیمت معتبر و دقیق با سنگ حسن تماس بگیرید.',
    matched_products.category_label_fa,
    TRIM(TO_CHAR(matched_products.offer_price, 'FM999,999,999'))
  ),
  updated_at = NOW()
FROM matched_products
WHERE p.id = matched_products.id;
