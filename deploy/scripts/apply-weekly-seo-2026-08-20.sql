-- Week 1 SEO cleanup and content schedule for sangehassan.com.
-- Idempotent: safe to rerun against the same production database.

BEGIN;

-- Repair internal product links after canonical product slug changes.
UPDATE blog_translations
SET content_html = replace(
      replace(
        replace(content_html,
          '/fa/products/hassan-travertine',
          '/fa/products/hassan-beige-travertine'
        ),
        '/fa/products/takab-travertine',
        '/fa/products/silver-takab-travertine'
      ),
      '/fa/products/chiseled-machine-and-hand',
      '/fa/products/hammerd-machine-and-hand'
    ),
    content_json = '{"type":"doc","content":[]}'::jsonb,
    updated_at = NOW()
WHERE locale = 'fa'
  AND (
    content_html LIKE '%/fa/products/hassan-travertine%'
    OR content_html LIKE '%/fa/products/takab-travertine%'
    OR content_html LIKE '%/fa/products/chiseled-machine-and-hand%'
  );

-- Abegarm is not an active catalog product, so retain the editorial mention without a dead link.
UPDATE blog_translations
SET content_html = regexp_replace(
      content_html,
      '<a[^>]*href=["'']\/fa\/products\/abegarm-travertine["''][^>]*>([^<]+)<\/a>',
      '\1',
      'gi'
    ),
    content_json = '{"type":"doc","content":[]}'::jsonb,
    updated_at = NOW()
WHERE locale = 'fa'
  AND content_html LIKE '%/fa/products/abegarm-travertine%';

-- Do not expose links to scheduled articles before their public URL exists.
DO $pending_links$
DECLARE
  target_slug TEXT;
BEGIN
  FOREACH target_slug IN ARRAY ARRAY[
    'black-dehbid-marble-stone',
    'marble-stone-cleaning-maintenance',
    'marble-stone-for-lobby',
    'marble-travertine-granite-interior-flooring'
  ]
  LOOP
    UPDATE blog_translations
    SET content_html = regexp_replace(
          content_html,
          '<a[^>]*href=["'']\/fa\/blogs\/' || target_slug || '["''][^>]*>([^<]+)<\/a>',
          '<span data-pending-blog-link="' || target_slug || '">\1</span>',
          'gi'
        ),
        content_json = '{"type":"doc","content":[]}'::jsonb,
        updated_at = NOW()
    WHERE locale = 'fa'
      AND content_html LIKE '%/fa/blogs/' || target_slug || '%'
      AND NOT EXISTS (
        SELECT 1
        FROM blog_translations target_translation
        JOIN blogs target_blog ON target_blog.id = target_translation.blog_id
        WHERE target_translation.locale = 'fa'
          AND target_translation.slug = target_slug
          AND target_translation.translation_status = 'published'
          AND target_blog.status = 'published'
          AND COALESCE(target_blog.published_at, target_blog.created_at) <= NOW()
      );
  END LOOP;
END
$pending_links$;

-- Remove the production placeholder from the Crystal pillar.
UPDATE blog_translations
SET content_html = replace(
      content_html,
      'برای مطالعه کامل‌تر درباره مرمریت، پیشنهاد می‌شود مقاله pillar مرمریت سنگ حسن نیز به این بخش لینک شود: [لینک داخلی به مقاله مرمریت: URL را من وارد می‌کنم]',
      'برای مطالعه کامل‌تر درباره تفاوت این دو خانواده سنگ، <a href="/fa/blogs/everything-about-marble-stone">راهنمای کامل سنگ مرمریت</a> را بخوانید.'
    ),
    content_json = '{"type":"doc","content":[]}'::jsonb,
    updated_at = NOW()
WHERE locale = 'fa'
  AND slug = 'everything-about-crystal-stone'
  AND content_html LIKE '%URL را من وارد می‌کنم%';

-- Keep two distinct products distinct in English and Arabic search results.
UPDATE products
SET title_en = 'Tree Trunk Travertine',
    title_ar = 'ترافرتين جذع الشجرة',
    short_description_html_en = CASE
      WHEN COALESCE(short_description_html_en, '') = '' THEN '<p>Tree Trunk Travertine has a directional pattern inspired by natural tree trunks. Review current slabs or tiles, finishes, dimensions, and project supply options before selection.</p>'
      ELSE short_description_html_en
    END,
    short_description_html_ar = CASE
      WHEN COALESCE(short_description_html_ar, '') = '' THEN '<p>ترافرتين جذع الشجرة يتميز بخطوط اتجاهية مستوحاة من شكل جذوع الأشجار. راجع صور الدفعة الحالية والأبعاد والتشطيبات وخيارات التوريد قبل الاختيار.</p>'
      ELSE short_description_html_ar
    END,
    updated_at = NOW()
WHERE slug = 'tree-trunk-travertine';

UPDATE products
SET title_en = 'Wood-Look Travertine',
    title_ar = 'ترافرتين بنقشة الخشب',
    short_description_html_en = CASE
      WHEN COALESCE(short_description_html_en, '') = '' THEN '<p>Wood-Look Travertine is selected for its wood-like natural movement and warm visual character. Compare the available batch, cut direction, finishes, and supply dimensions for the project.</p>'
      ELSE short_description_html_en
    END,
    short_description_html_ar = CASE
      WHEN COALESCE(short_description_html_ar, '') = '' THEN '<p>ترافرتين بنقشة الخشب يتم اختياره لحركته الطبيعية القريبة من مظهر الخشب. قارن الدفعة المتاحة واتجاه القطع والتشطيبات والأبعاد المطلوبة للمشروع.</p>'
      ELSE short_description_html_ar
    END,
    updated_at = NOW()
WHERE slug = 'wooden-travertine';

-- Assign the generic "سنگ اپن" intent to Natanz Black Granite; use specific variants elsewhere.
WITH countertop_copy(slug, section_html) AS (
  VALUES
    ('natanz-black-granite', '<section data-seo-section="countertop-2026"><h2>گرانیت مشکی نطنز برای سنگ اپن و صفحه کابینت</h2><p>گرانیت مشکی نطنز یکی از گزینه‌های قابل بررسی برای سنگ اپن، صفحه کابینت و کانتر سنگی است. برای این کاربرد باید ابعاد قطعه، ضخامت، محل برش سینک و گاز، نوع لبه و کیفیت فرآوری از روی سنگ قابل‌تحویل مشخص شود. سطح صیقلی ظاهر تیره و دانه‌بندی سنگ را واضح‌تر نشان می‌دهد، اما انتخاب نهایی باید با توجه به نور فضا و برنامه نگهداری انجام شود.</p><p>برای استعلام سنگ اپن مشکی، طول و عرض هر قطعه، تعداد بازشوها، ضخامت، نوع لبه، فرآوری و شهر پروژه را ارسال کنید تا امکان برش و تأمین بررسی شود.</p></section>'),
    ('khorramdareh-granite', '<section data-seo-section="countertop-2026"><h2>گرانیت خرمدره برای صفحه کابینت و کانتر</h2><p>گرانیت خرمدره برای صفحه کابینت و کانتر پروژه‌هایی قابل بررسی است که ظاهر دانه‌دار طبیعی و سطح سنگی مقاوم مدنظر است. پیش از سفارش باید رنگ و دانه‌بندی بچ، ابعاد مفید، ضخامت، کیفیت صیقل و جزئیات برش سینک و گاز کنترل شود.</p><p>برای قیمت دقیق، نقشه یا ابعاد صفحه، تعداد قطعات، نوع لبه، فرآوری و شهر پروژه را همراه با مقدار موردنیاز ارسال کنید.</p></section>'),
    ('nehbandan-granite', '<section data-seo-section="countertop-2026"><h2>گرانیت نهبندان برای کانتر و صفحه کابینت</h2><p>گرانیت نهبندان در صورت هماهنگی سورت، ضخامت و فرآوری می‌تواند برای کانتر و صفحه کابینت سنگی بررسی شود. ظاهر نهایی باید از روی تصاویر همان بار ارزیابی شود و محل بازشوها، اتصال قطعات و پرداخت لبه پیش از برش مشخص باشد.</p><p>ابعاد صفحه، ضخامت، تعداد بازشوها، نوع لبه و شهر پروژه را برای بررسی امکان تأمین و برش اعلام کنید.</p></section>'),
    ('morvarid-mashhad-granite', '<section data-seo-section="countertop-2026"><h2>گرانیت مروارید مشهد برای صفحه کابینت</h2><p>گرانیت مروارید مشهد برای صفحه کابینت و کانتر با ظاهر دانه‌دار روشن قابل بررسی است. یکنواختی سورت، کیفیت صیقل، ضخامت و سلامت قطعه در اطراف برش‌های سینک و گاز باید قبل از سفارش کنترل شود.</p><p>برای استعلام، ابعاد دقیق، تعداد قطعات، ضخامت، نوع لبه، فرآوری و شهر پروژه را ارسال کنید.</p></section>'),
    ('nayin-red-granite', '<section data-seo-section="countertop-2026"><h2>گرانیت قرمز نائین برای کانتر و سطوح شاخص</h2><p>گرانیت قرمز نائین برای کانتر یا صفحه سنگی شاخص در پروژه‌هایی قابل بررسی است که رنگ قرمز طبیعی بخشی از طراحی باشد. قبل از انتخاب، دامنه تغییر رنگ بچ، دانه‌بندی، کیفیت صیقل، ابعاد مفید و جزئیات لبه و بازشوها را بررسی کنید.</p><p>برای ارزیابی پروژه، ابعاد، تعداد، ضخامت، نوع لبه، فرآوری و شهر اجرا را ارسال کنید.</p></section>')
)
UPDATE products product
SET description_html_fa = COALESCE(product.description_html_fa, '') || countertop_copy.section_html,
    updated_at = NOW()
FROM countertop_copy
WHERE product.slug = countertop_copy.slug
  AND POSITION('data-seo-section="countertop-2026"' IN COALESCE(product.description_html_fa, '')) = 0;

-- Prepare links from the Granite pillar without exposing scheduled URLs.
UPDATE blog_translations
SET content_html = content_html || '<section data-seo-links="granite-week-2026"><h2>راهنماهای تکمیلی انتخاب گرانیت</h2><p>برای بررسی محدودیت‌های کاربرد و ادعاهای مربوط به سلامت، <span data-pending-blog-link="granite-advantages-disadvantages-health">راهنمای مزایا و معایب سنگ گرانیت</span> را بخوانید. برای سفارش قطعات بزرگ نیز <span data-pending-blog-link="granite-slab-guide">راهنمای اسلب گرانیت</span> ابعاد، فرآوری، حمل و نکات خرید را توضیح می‌دهد.</p></section>',
    content_json = '{"type":"doc","content":[]}'::jsonb,
    updated_at = NOW()
WHERE locale = 'fa'
  AND slug = 'everything-about-granite-stone'
  AND POSITION('data-seo-links="granite-week-2026"' IN content_html) = 0;

-- Article 1: granite advantages, disadvantages, and health claims.
DO $granite_health$
DECLARE
  target_blog_id INTEGER;
BEGIN
  SELECT blog_id INTO target_blog_id
  FROM blog_translations
  WHERE locale = 'fa' AND slug = 'granite-advantages-disadvantages-health';

  IF target_blog_id IS NULL THEN
    INSERT INTO blogs (
      title, excerpt, content, status, author_name, category_slug, tags,
      is_featured, scheduled_at, published_at, created_at, updated_at
    ) VALUES (
      'مزایا و معایب سنگ گرانیت؛ آیا گرانیت برای ساختمان ضرر دارد؟',
      'بررسی مزایا، محدودیت‌ها و ادعاهای مربوط به مضرات سنگ گرانیت برای کف، پله، نما، کانتر و فضای داخلی با تکیه بر منابع رسمی.',
      '', 'scheduled', 'تیم محتوای سنگ حسن', 'guide',
      ARRAY['گرانیت','مضرات سنگ گرانیت','مزایا و معایب گرانیت','سنگ ساختمانی'],
      FALSE, TIMESTAMPTZ '2026-08-21 09:00:00+03:30', NULL, NOW(), NOW()
    ) RETURNING id INTO target_blog_id;
  ELSE
    UPDATE blogs
    SET title = 'مزایا و معایب سنگ گرانیت؛ آیا گرانیت برای ساختمان ضرر دارد؟',
        excerpt = 'بررسی مزایا، محدودیت‌ها و ادعاهای مربوط به مضرات سنگ گرانیت برای کف، پله، نما، کانتر و فضای داخلی با تکیه بر منابع رسمی.',
        status = CASE WHEN status = 'published' THEN status ELSE 'scheduled' END,
        scheduled_at = CASE WHEN status = 'published' THEN NULL ELSE TIMESTAMPTZ '2026-08-21 09:00:00+03:30' END,
        author_name = 'تیم محتوای سنگ حسن', category_slug = 'guide',
        tags = ARRAY['گرانیت','مضرات سنگ گرانیت','مزایا و معایب گرانیت','سنگ ساختمانی'],
        updated_at = NOW()
    WHERE id = target_blog_id;
  END IF;

  INSERT INTO blog_translations (
    blog_id, locale, title, slug, excerpt, content_json, content_html,
    seo_title, seo_description, canonical_url, robots, translation_status,
    featured_image_alt, og_image_alt, created_at, updated_at
  ) VALUES (
    target_blog_id,
    'fa',
    'مزایا و معایب سنگ گرانیت؛ آیا گرانیت برای ساختمان ضرر دارد؟',
    'granite-advantages-disadvantages-health',
    'مزایا و معایب سنگ گرانیت برای کف، پله، نما، کانتر و فضای داخلی؛ بررسی علمی ادعای تشعشع و رادون و راهنمای انتخاب متناسب با پروژه.',
    '{"type":"doc","content":[]}'::jsonb,
    $content$
      <p>سنگ گرانیت به دلیل ساختار متراکم، تنوع رنگ و امکان اجرای فرآوری‌های مختلف برای کف، پله، محوطه، نما و کانتر بررسی می‌شود. با این حال، وزن بالا، دشواری برش، احتمال لغزندگی سطح صیقلی و تفاوت طبیعی بین سورت‌ها از محدودیت‌های آن هستند. درباره تشعشع و رادون نیز نباید با یک حکم کلی درباره همه گرانیت‌ها قضاوت کرد.</p>
      <p>انتخاب درست زمانی انجام می‌شود که نوع گرانیت، محل اجرا، ضخامت، ابعاد، فرآوری و شرایط نگهداری هم‌زمان بررسی شوند. برای شناخت خانواده این سنگ ابتدا <a href="/fa/blogs/everything-about-granite-stone">راهنمای کامل سنگ گرانیت</a> و برای مشاهده گزینه‌های موجود <a href="/fa/products/granite">صفحه محصولات گرانیت</a> را ببینید.</p>

      <h2>خلاصه مزایا و معایب سنگ گرانیت</h2>
      <table><tbody>
        <tr><th><p>موضوع</p></th><th><p>مزیت</p></th><th><p>محدودیت یا نکته کنترل</p></th></tr>
        <tr><td><p>کف و پله</p></td><td><p>ساختار متراکم و قابلیت تأمین با ضخامت و ابعاد مختلف</p></td><td><p>سطح صیقلی در فضای خیس می‌تواند لغزنده باشد؛ فرآوری باید متناسب با محل انتخاب شود.</p></td></tr>
        <tr><td><p>فضای بیرونی</p></td><td><p>امکان استفاده از فرآوری‌های بافت‌دار مانند فلیم یا بوش‌همر</p></td><td><p>تغییر ظاهر فرآوری و شرایط اقلیمی باید روی نمونه واقعی بررسی شود.</p></td></tr>
        <tr><td><p>کانتر و صفحه کابینت</p></td><td><p>سطح سنگ طبیعی متراکم و قابل برش برای نقشه پروژه</p></td><td><p>ابعاد مفید، درزها، بازشوها، لبه و نصب حرفه‌ای تعیین‌کننده‌اند.</p></td></tr>
        <tr><td><p>حمل و نصب</p></td><td><p>امکان تولید قطعات پروژه‌ای</p></td><td><p>وزن بالا و سختی برش می‌تواند هزینه حمل، زیرسازی و اجرا را افزایش دهد.</p></td></tr>
      </tbody></table>

      <h2>مهم‌ترین مزایای سنگ گرانیت</h2>
      <h3>ساختار متراکم و مناسب برای سطوح پرتردد</h3>
      <p>بسیاری از گرانیت‌های ساختمانی برای کف، راهرو، پله و محوطه بررسی می‌شوند؛ زیرا ساختار متراکم آن‌ها امکان استفاده در قطعات پرتردد را فراهم می‌کند. این مزیت به معنی مناسب‌بودن خودکار هر گرانیت برای هر پروژه نیست و ضخامت، سلامت قطعه و کیفیت نصب همچنان اهمیت دارند.</p>
      <h3>تنوع فرآوری برای فضای داخلی و خارجی</h3>
      <p>گرانیت می‌تواند با سطح صیقلی، هوند، فلیم، چرمی یا بوش‌همر عرضه شود. سطح صیقلی رنگ و دانه‌بندی را واضح‌تر می‌کند، درحالی‌که فرآوری‌های بافت‌دار برای بعضی فضاهای بیرونی یا محل‌هایی که کنترل لغزندگی مهم است بررسی می‌شوند.</p>
      <h3>تنوع رنگ و دانه‌بندی طبیعی</h3>
      <p>گرانیت‌های روشن، طوسی، مشکی، سبز و قرمز امکان هماهنگی با سبک‌های مختلف پروژه را دارند. بااین‌حال، تصویر یک نمونه کوچک نماینده کل سفارش نیست؛ سورت و بچ قابل‌تحویل باید روی چند قطعه و در نور مشابه فضای پروژه دیده شود.</p>
      <h3>امکان استفاده به صورت اسلب و تایل</h3>
      <p>بسته به ابعاد خام و برنامه تولید، گرانیت به شکل تایل یا اسلب قابل بررسی است. تایل برای شبکه‌های منظم کف و دیوار و اسلب برای قطعات بزرگ، کانتر، دیوار شاخص یا برش‌های پروژه‌ای کاربرد دارد.</p>

      <h2>معایب و محدودیت‌های سنگ گرانیت</h2>
      <h3>وزن و هزینه حمل</h3>
      <p>وزن قطعات گرانیت با افزایش ابعاد و ضخامت بیشتر می‌شود. در سفارش اسلب، پله یا قطعات بزرگ باید ظرفیت حمل، بسته‌بندی، تخلیه، دسترسی کارگاه و زیرسازی پیش از خرید بررسی شود.</p>
      <h3>برش و اجرای تخصصی‌تر</h3>
      <p>ساختار متراکم گرانیت می‌تواند برش، ابزارزنی لبه و ایجاد بازشو را دشوارتر کند. کیفیت تجهیزات و تجربه مجری روی سلامت لبه‌ها، دقت اتصال‌ها و پرت پروژه اثر مستقیم دارد.</p>
      <h3>لغزندگی بعضی سطوح صیقلی</h3>
      <p>سطح بسیار صیقلی در محیط‌های مرطوب یا ورودی‌های خیس ممکن است انتخاب مناسبی نباشد. برای چنین فضاهایی باید نمونه فرآوری نهایی از نظر بافت، تمیزشوندگی و اصطکاک بررسی شود؛ نام سنگ به‌تنهایی برای این تصمیم کافی نیست.</p>
      <h3>تفاوت طبیعی سورت و ظاهر</h3>
      <p>دانه‌بندی، لکه‌های طبیعی، رگه‌ها و دامنه رنگ می‌توانند بین بچ‌ها متفاوت باشند. خرید بر اساس یک تصویر آرشیوی بدون مشاهده بار قابل‌تحویل، احتمال اختلاف با انتظار پروژه را افزایش می‌دهد.</p>

      <h2>آیا سنگ گرانیت برای ساختمان ضرر دارد؟</h2>
      <p>گرانیت مانند بسیاری از مصالح طبیعی ممکن است مقدار کمی عناصر پرتوزای طبیعی داشته باشد و مقدار آن بین نمونه‌ها متفاوت است. بااین‌حال، <a href="https://www.epa.gov/radiation/granite-countertops-and-radiation" target="_blank" rel="noopener noreferrer">سازمان حفاظت محیط زیست آمریکا</a> اعلام می‌کند بسیار بعید است تابش کانترهای گرانیتی، دوز سالانه را بالاتر از پس‌زمینه طبیعی ببرد و رادون خاک زیر ساختمان معمولاً منبع مهم‌تری است.</p>
      <p><a href="https://www.cdc.gov/radiation-health/data-research/facts-stats/building-materials.html" target="_blank" rel="noopener noreferrer">مرکز کنترل و پیشگیری بیماری‌های آمریکا</a> نیز می‌گوید مصالحی مانند سنگ طبیعی و گرانیت معمولاً سطحی از مواد پرتوزا ندارند که دوز دریافتی را بالاتر از مقادیر کم پس‌زمینه روزانه افزایش دهد. این نتیجه به معنی یکسان‌بودن همه نمونه‌ها نیست؛ اگر درباره رادون فضای داخلی نگرانی وجود دارد، سنجش هوای ساختمان از قضاوت بر اساس نام یا رنگ سنگ معتبرتر است.</p>
      <blockquote><p>نتیجه عملی: نمی‌توان همه گرانیت‌های ساختمانی را مضر دانست و همچنین نمی‌توان بدون اندازه‌گیری درباره یک نمونه خاص ادعای صفر مطلق داشت. برای نگرانی واقعی درباره رادون، اندازه‌گیری محیط و مشورت با متخصص معتبر راه درست‌تری است.</p></blockquote>

      <h2>گرانیت برای کدام کاربردها مناسب‌تر است؟</h2>
      <ul>
        <li><p><strong>کف و راهرو:</strong> با انتخاب ضخامت و فرآوری متناسب با میزان تردد.</p></li>
        <li><p><strong>پله:</strong> با کنترل ضخامت، لبه، ابعاد یکسان و لغزندگی سطح.</p></li>
        <li><p><strong>محوطه:</strong> با بررسی فرآوری بافت‌دار، اقلیم و روش نصب.</p></li>
        <li><p><strong>کانتر:</strong> با نقشه دقیق، کنترل بازشوها، اتصال قطعات و اجرای تخصصی.</p></li>
        <li><p><strong>نما:</strong> فقط پس از بررسی وزن، مهار، ضخامت، فرآوری و جزئیات اجرایی پروژه.</p></li>
      </ul>

      <h2>چند گرانیت موجود برای مقایسه پروژه</h2>
      <p>برای مقایسه رنگ، تصاویر، فرآوری و کف قیمت می‌توانید صفحات <a href="/fa/products/natanz-black-granite">گرانیت مشکی نطنز</a>، <a href="/fa/products/khorramdareh-granite">گرانیت خرمدره</a>، <a href="/fa/products/nehbandan-granite">گرانیت نهبندان</a>، <a href="/fa/products/morvarid-mashhad-granite">گرانیت مروارید مشهد</a> و <a href="/fa/products/nayin-red-granite">گرانیت قرمز نائین</a> را بررسی کنید. مناسب‌بودن نهایی هر محصول باید با توجه به سورت، فرآوری و جزئیات پروژه تأیید شود.</p>

      <h2>هنگام خرید گرانیت چه اطلاعاتی آماده کنیم؟</h2>
      <ol>
        <li><p>محل دقیق اجرا و داخلی یا خارجی بودن فضا</p></li>
        <li><p>ابعاد، تعداد یا متراژ و ضخامت موردنیاز</p></li>
        <li><p>نوع فرآوری سطح و جزئیات لبه</p></li>
        <li><p>میزان تردد، رطوبت و شرایط نظافت</p></li>
        <li><p>شهر پروژه، روش حمل و محدودیت تخلیه</p></li>
      </ol>
      <p>برای استعلام، این اطلاعات را همراه نام گرانیت و تصاویر یا نقشه پروژه برای تیم فروش و تأمین سنگ حسن ارسال کنید.</p>

      <h2>سوالات متداول</h2>
      <h3>آیا همه سنگ‌های گرانیت رادیواکتیو و خطرناک هستند؟</h3>
      <p>خیر. گرانیت ممکن است مانند دیگر مصالح طبیعی عناصر پرتوزای طبیعی داشته باشد، اما منابع رسمی سطح تابش مصالح رایج گرانیتی را عموماً پایین می‌دانند. درباره یک نمونه خاص فقط اندازه‌گیری معتبر امکان قضاوت دقیق می‌دهد.</p>
      <h3>آیا گرانیت صیقلی برای فضای خیس مناسب است؟</h3>
      <p>سطح صیقلی ممکن است هنگام خیس‌شدن لغزنده باشد. برای فضای مرطوب باید نمونه فرآوری نهایی از نظر اصطکاک و نظافت بررسی و با الزامات پروژه تطبیق داده شود.</p>
      <h3>مهم‌ترین عیب گرانیت در اجرا چیست؟</h3>
      <p>وزن، سختی برش و نیاز به اجرای دقیق از مهم‌ترین محدودیت‌ها هستند. این موارد روی هزینه حمل، زیرسازی، ابزارزنی و نصب اثر می‌گذارند.</p>
      <h3>برای انتخاب گرانیت فقط نام معدن کافی است؟</h3>
      <p>خیر. سورت، دانه‌بندی، سلامت قطعه، ضخامت، ابعاد، فرآوری و هماهنگی کل بچ باید در کنار نام محصول بررسی شوند.</p>
    $content$,
    'مضرات سنگ گرانیت چیست؟ بررسی مزایا و معایب گرانیت | سنگ حسن',
    'بررسی علمی مضرات، مزایا و معایب سنگ گرانیت برای کف، پله، نما و کانتر؛ پاسخ به ادعای تشعشع و رادون و نکات انتخاب برای ساختمان.',
    '/fa/blogs/granite-advantages-disadvantages-health',
    'index,follow', 'published', '', '', NOW(), NOW()
  )
  ON CONFLICT (blog_id, locale) DO UPDATE SET
    title = EXCLUDED.title,
    slug = EXCLUDED.slug,
    excerpt = EXCLUDED.excerpt,
    content_json = EXCLUDED.content_json,
    content_html = EXCLUDED.content_html,
    seo_title = EXCLUDED.seo_title,
    seo_description = EXCLUDED.seo_description,
    canonical_url = EXCLUDED.canonical_url,
    robots = EXCLUDED.robots,
    translation_status = EXCLUDED.translation_status,
    updated_at = NOW();
END
$granite_health$;

-- Article 2: Granite slab buying guide.
DO $granite_slab$
DECLARE
  target_blog_id INTEGER;
BEGIN
  SELECT blog_id INTO target_blog_id
  FROM blog_translations
  WHERE locale = 'fa' AND slug = 'granite-slab-guide';

  IF target_blog_id IS NULL THEN
    INSERT INTO blogs (
      title, excerpt, content, status, author_name, category_slug, tags,
      is_featured, scheduled_at, published_at, created_at, updated_at
    ) VALUES (
      'اسلب گرانیت چیست؟ ابعاد، کاربردها، فرآوری و راهنمای خرید',
      'راهنمای انتخاب و خرید اسلب گرانیت؛ بررسی ابعاد، ضخامت، فرآوری، کاربرد در کانتر، کف و دیوار، عوامل قیمت، بسته‌بندی و حمل.',
      '', 'scheduled', 'تیم محتوای سنگ حسن', 'guide',
      ARRAY['گرانیت','اسلب گرانیت','قیمت اسلب گرانیت','خرید اسلب'],
      FALSE, TIMESTAMPTZ '2026-08-25 09:00:00+03:30', NULL, NOW(), NOW()
    ) RETURNING id INTO target_blog_id;
  ELSE
    UPDATE blogs
    SET title = 'اسلب گرانیت چیست؟ ابعاد، کاربردها، فرآوری و راهنمای خرید',
        excerpt = 'راهنمای انتخاب و خرید اسلب گرانیت؛ بررسی ابعاد، ضخامت، فرآوری، کاربرد در کانتر، کف و دیوار، عوامل قیمت، بسته‌بندی و حمل.',
        status = CASE WHEN status = 'published' THEN status ELSE 'scheduled' END,
        scheduled_at = CASE WHEN status = 'published' THEN NULL ELSE TIMESTAMPTZ '2026-08-25 09:00:00+03:30' END,
        author_name = 'تیم محتوای سنگ حسن', category_slug = 'guide',
        tags = ARRAY['گرانیت','اسلب گرانیت','قیمت اسلب گرانیت','خرید اسلب'],
        updated_at = NOW()
    WHERE id = target_blog_id;
  END IF;

  INSERT INTO blog_translations (
    blog_id, locale, title, slug, excerpt, content_json, content_html,
    seo_title, seo_description, canonical_url, robots, translation_status,
    featured_image_alt, og_image_alt, created_at, updated_at
  ) VALUES (
    target_blog_id,
    'fa',
    'اسلب گرانیت چیست؟ ابعاد، کاربردها، فرآوری و راهنمای خرید',
    'granite-slab-guide',
    'راهنمای خرید اسلب گرانیت؛ بررسی ابعاد و ضخامت، کاربرد در کانتر، کف و دیوار، انواع فرآوری، عوامل قیمت، بسته‌بندی و حمل پروژه‌ای.',
    '{"type":"doc","content":[]}'::jsonb,
    $content$
      <p>اسلب گرانیت قطعه بزرگ سنگ گرانیت است که برای برش‌های پروژه‌ای، کانتر و صفحه کابینت، دیوار شاخص، کف با قطعات بزرگ و ساخت پله یا اجزای سفارشی بررسی می‌شود. اسلب برخلاف تایل، اندازه ثابت و واحدی ندارد و ابعاد مفید آن به بلوک خام، روش برش، سلامت قطعه و برنامه تولید وابسته است.</p>
      <p>برای خرید اسلب گرانیت نباید فقط نام سنگ یا تصویر یک قطعه را ملاک قرار داد. ابعاد مفید، ضخامت، سورت، دانه‌بندی، سلامت لبه، کیفیت فرآوری، نقشه برش، پرت، بسته‌بندی و مسیر حمل همگی روی انتخاب و قیمت نهایی اثر دارند. برای شناخت ویژگی‌های پایه ابتدا <a href="/fa/blogs/everything-about-granite-stone">راهنمای کامل سنگ گرانیت</a> را بخوانید.</p>

      <h2>تفاوت اسلب گرانیت با تایل چیست؟</h2>
      <table><tbody>
        <tr><th><p>معیار</p></th><th><p>اسلب گرانیت</p></th><th><p>تایل گرانیت</p></th></tr>
        <tr><td><p>ابعاد</p></td><td><p>قطعات بزرگ با ابعاد مفید وابسته به بلوک و برش</p></td><td><p>قطعات تکرارشونده با ابعاد سفارش یا تولید</p></td></tr>
        <tr><td><p>کاربرد</p></td><td><p>کانتر، دیوار شاخص، برش سفارشی، کف بزرگ و قطعات پروژه‌ای</p></td><td><p>کف، دیوار، پله و شبکه‌های منظم اجرایی</p></td></tr>
        <tr><td><p>انتخاب طرح</p></td><td><p>نیازمند مشاهده کامل هر قطعه و تعیین نقشه برش</p></td><td><p>تمرکز بیشتر بر هماهنگی سورت و تکرار قطعات</p></td></tr>
        <tr><td><p>حمل</p></td><td><p>نیازمند خرک، مهاربندی و تجهیزات تخلیه مناسب</p></td><td><p>معمولاً بسته‌بندی و جابه‌جایی ساده‌تر در ابعاد کوچک‌تر</p></td></tr>
      </tbody></table>

      <h2>ابعاد و ضخامت اسلب گرانیت چگونه تعیین می‌شود؟</h2>
      <p>هیچ عدد ثابتی را نمی‌توان برای تمام اسلب‌های گرانیت اعلام کرد. طول و عرض مفید به اندازه بلوک، ترک‌ها یا نقاط قابل‌برش، تجهیزات کارخانه و حاشیه لازم برای اصلاح لبه وابسته است. ضخامت نیز باید بر اساس کاربرد، ابعاد قطعه، محل بازشو، نوع زیرسازی و روش حمل مشخص شود.</p>
      <p>در استعلام پروژه، ابعاد نهایی سطح را با ابعاد خام اسلب اشتباه نگیرید. بخشی از سنگ ممکن است برای اصلاح لبه، هماهنگ‌کردن طرح یا برش بازشوها مصرف شود. نقشه برش قبل از خرید، امکان محاسبه تعداد اسلب و پرت را دقیق‌تر می‌کند.</p>

      <h2>کاربردهای اسلب گرانیت</h2>
      <h3>کانتر و صفحه کابینت</h3>
      <p>اسلب امکان برش صفحه‌های بزرگ‌تر و کاهش تعداد اتصال‌ها را فراهم می‌کند. محل سینک، گاز، شیرآلات، پریز، نوع لبه و جهت دانه‌بندی باید پیش از برش تعیین شود. برای گزینه مشکی می‌توانید <a href="/fa/products/natanz-black-granite">گرانیت مشکی نطنز</a> را بررسی کنید.</p>
      <h3>کف و لابی</h3>
      <p>قطعات بزرگ می‌توانند تعداد بندها را کمتر کنند، اما وزن، ضخامت، زیرسازی، امکان حمل داخل ساختمان و روش نصب باید از ابتدا در طراحی دیده شوند. انتخاب سطح صیقلی یا هوند نیز با میزان تردد و شرایط نظافت مرتبط است.</p>
      <h3>دیوار شاخص و فضای داخلی</h3>
      <p>برای دیوار شاخص، تصویر کامل اسلب، محل برش و نحوه اتصال قطعات اهمیت دارد. برخلاف بعضی سنگ‌های رگه‌دار، گرانیت بیشتر با دانه‌بندی، لکه‌های طبیعی و تغییرات رنگی شناخته می‌شود و چیدمان باید بر همان ویژگی واقعی طراحی شود.</p>
      <h3>پله و قطعات سفارشی</h3>
      <p>اسلب می‌تواند ماده اولیه برش پله، پاگرد، پیشانی یا قطعات سفارشی باشد. هماهنگی ضخامت، جزئیات لبه و تکرار رنگ در تمام قطعات باید پیش از تولید کنترل شود.</p>

      <h2>کدام فرآوری برای اسلب گرانیت مناسب است؟</h2>
      <table><tbody>
        <tr><th><p>فرآوری</p></th><th><p>اثر ظاهری و کاربردی</p></th><th><p>نکته انتخاب</p></th></tr>
        <tr><td><p>صیقلی</p></td><td><p>رنگ و دانه‌بندی را واضح‌تر و سطح را براق‌تر نشان می‌دهد.</p></td><td><p>برای محیط خیس، لغزندگی و روش نگهداری بررسی شود.</p></td></tr>
        <tr><td><p>هوند</p></td><td><p>بازتاب کمتر و ظاهر مات‌تر ایجاد می‌کند.</p></td><td><p>نمونه نهایی از نظر رنگ و نظافت دیده شود.</p></td></tr>
        <tr><td><p>چرمی</p></td><td><p>بافت لمسی و بازتاب کنترل‌شده‌تری ایجاد می‌کند.</p></td><td><p>یکنواختی بافت و جزئیات لبه کنترل شود.</p></td></tr>
        <tr><td><p>فلیم یا بوش‌همر</p></td><td><p>سطح زبرتر برای برخی کاربردهای بیرونی ایجاد می‌کند.</p></td><td><p>رنگ نهایی می‌تواند با نمونه صیقلی متفاوت باشد.</p></td></tr>
      </tbody></table>

      <h2>قیمت اسلب گرانیت به چه عواملی بستگی دارد؟</h2>
      <ul>
        <li><p>نوع گرانیت، رنگ و سورت محموله</p></li>
        <li><p>ابعاد مفید و سلامت کامل اسلب</p></li>
        <li><p>ضخامت و کیفیت کالیبراسیون</p></li>
        <li><p>نوع فرآوری سطح و ابزارزنی لبه</p></li>
        <li><p>تعداد اسلب و امکان تأمین از یک بچ هماهنگ</p></li>
        <li><p>نقشه برش، تعداد بازشوها و میزان پرت</p></li>
        <li><p>خرک، بسته‌بندی، بیمه، حمل و تجهیزات تخلیه</p></li>
      </ul>
      <p>به همین دلیل مقایسه دو قیمت بدون یکسان‌کردن ابعاد، ضخامت، فرآوری و شرایط تحویل نتیجه دقیقی نمی‌دهد. قیمت باید بر اساس تصاویر و مشخصات اسلب‌های قابل‌تحویل تأیید شود.</p>

      <h2>بسته‌بندی و حمل اسلب گرانیت</h2>
      <p>اسلب‌ها به خرک مناسب، مهاربندی، فاصله‌گذاری و برنامه تخلیه نیاز دارند. وزن کل بار، ظرفیت وسیله حمل، مسیر ورود به پروژه و وجود جرثقیل یا تجهیزات جابه‌جایی باید پیش از ارسال مشخص شود. توافق درباره مسئولیت بارگیری و تخلیه نیز باید در پیش‌فاکتور ثبت شود.</p>

      <h2>چند اسلب گرانیت برای بررسی</h2>
      <p>در محصولات موجود سنگ حسن می‌توانید <a href="/fa/products/natanz-black-granite">گرانیت مشکی نطنز</a>، <a href="/fa/products/khorramdareh-granite">گرانیت خرمدره</a>، <a href="/fa/products/nehbandan-granite">گرانیت نهبندان</a>، <a href="/fa/products/morvarid-mashhad-granite">گرانیت مروارید مشهد</a> و <a href="/fa/products/nayin-red-granite">گرانیت قرمز نائین</a> را از نظر تصویر، فرآوری و قیمت پایه مقایسه کنید. موجودی اسلب و ابعاد هر بار باید جداگانه تأیید شود.</p>

      <h2>چک‌لیست خرید اسلب گرانیت</h2>
      <ol>
        <li><p>کاربرد و ابعاد نهایی هر قطعه را مشخص کنید.</p></li>
        <li><p>نقشه برش و محل بازشوها را آماده کنید.</p></li>
        <li><p>تصویر کامل و شماره هر اسلب را دریافت کنید.</p></li>
        <li><p>ضخامت، فرآوری، سلامت لبه و تلرانس ابعاد را ثبت کنید.</p></li>
        <li><p>تعداد موردنیاز را با پرت منطقی محاسبه کنید.</p></li>
        <li><p>شرایط خرک، حمل، تخلیه و شهر پروژه را در استعلام بنویسید.</p></li>
      </ol>
      <p>برای دریافت قیمت دقیق، نام گرانیت، ابعاد، تعداد اسلب، ضخامت، فرآوری، نقشه برش و شهر پروژه را برای تیم فروش و تأمین سنگ حسن ارسال کنید.</p>

      <h2>سوالات متداول</h2>
      <h3>آیا اسلب گرانیت ابعاد استاندارد ثابتی دارد؟</h3>
      <p>خیر. ابعاد مفید هر اسلب به بلوک خام، روش برش و سلامت قطعه وابسته است. ابعاد دقیق باید از روی بار قابل‌تحویل اعلام شود.</p>
      <h3>برای کانتر، اسلب بهتر است یا تایل؟</h3>
      <p>اسلب معمولاً امکان تولید قطعات بزرگ‌تر و اتصال کمتر را می‌دهد، اما انتخاب نهایی به نقشه، ابعاد موجود، حمل، بودجه و توان اجرایی پروژه بستگی دارد.</p>
      <h3>چرا قیمت دو اسلب از یک گرانیت متفاوت است؟</h3>
      <p>سورت، ابعاد سالم، ضخامت، فرآوری، سلامت لبه، کیفیت سطح و شرایط حمل می‌توانند قیمت دو قطعه با نام یکسان را متفاوت کنند.</p>
      <h3>برای استعلام اسلب چه اطلاعاتی لازم است؟</h3>
      <p>نام سنگ، ابعاد یا نقشه برش، تعداد، ضخامت، فرآوری، نوع لبه، محل بازشوها و شهر پروژه اطلاعات اصلی استعلام هستند.</p>
    $content$,
    'اسلب گرانیت چیست؟ ابعاد، کاربرد، قیمت و راهنمای خرید | سنگ حسن',
    'راهنمای خرید اسلب گرانیت؛ بررسی ابعاد، ضخامت، فرآوری، کاربرد در کانتر و کف، عوامل قیمت، نقشه برش، بسته‌بندی و حمل.',
    '/fa/blogs/granite-slab-guide',
    'index,follow', 'published', '', '', NOW(), NOW()
  )
  ON CONFLICT (blog_id, locale) DO UPDATE SET
    title = EXCLUDED.title,
    slug = EXCLUDED.slug,
    excerpt = EXCLUDED.excerpt,
    content_json = EXCLUDED.content_json,
    content_html = EXCLUDED.content_html,
    seo_title = EXCLUDED.seo_title,
    seo_description = EXCLUDED.seo_description,
    canonical_url = EXCLUDED.canonical_url,
    robots = EXCLUDED.robots,
    translation_status = EXCLUDED.translation_status,
    updated_at = NOW();
END
$granite_slab$;

COMMIT;
