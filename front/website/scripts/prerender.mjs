import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { createServer } from "vite";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, "..");
const distDir = path.join(rootDir, "dist");
const templatePath = path.join(distDir, "index.html");
const manifestPath = path.join(distDir, ".vite", "manifest.json");
const siteUrl = (process.env.VITE_SITE_URL || "https://sangehassan.com").replace(/\/+$/, "");
const prerenderApiBase = (process.env.VITE_PRERENDER_API_BASE_URL || "").replace(/\/+$/, "");
const configuredSitemapLastmod = process.env.VITE_SITEMAP_LASTMOD || "";
const sitemapLastmod = /^\d{4}-\d{2}-\d{2}$/.test(configuredSitemapLastmod)
  ? configuredSitemapLastmod
  : new Date().toISOString().slice(0, 10);
const organizationId = `${siteUrl}/#organization`;
const websiteId = `${siteUrl}/#website`;
const defaultRobots = "index,follow";
const protectedImageRev = process.env.VITE_PROTECTED_IMAGE_REV || "protected-2026-06-09-r3";
const sharedAssetUrl = (sourcePath) => `/@fs${path.resolve(rootDir, sourcePath).replace(/\\/g, "/")}`;
const defaultShareImage = sharedAssetUrl("../shared/assets/landing_page/landingpage_blocks_overlay.webp");
const fetchTimeoutMs = Number(process.env.VITE_PRERENDER_FETCH_TIMEOUT_MS || 10000);

const catalogLocales = ["en", "fa", "ar"];
const blogLocaleMeta = {
  en: { locale: "en_US", home: "Home", articles: "Articles", title: "Natural stone articles | SangeHassan", description: "Practical guidance for choosing, processing, installing and maintaining natural stone." },
  fa: { locale: "fa_IR", home: "خانه", articles: "مقالات", title: "مقالات سنگ طبیعی | سنگ حسن", description: "راهنماهای کاربردی برای انتخاب، فرآوری، اجرا و نگهداری سنگ طبیعی." },
  ar: { locale: "ar_SA", home: "الرئيسية", articles: "المقالات", title: "مقالات الحجر الطبيعي | سانج حسن", description: "أدلة عملية لاختيار الحجر الطبيعي ومعالجته وتركيبه وصيانته." }
};

const requiredFaBlogSlugs = [
  "everything-about-granite-stone",
  "everything-about-travertine-stone"
];

const faArticleSeoOverrides = {
  "everything-about-granite-stone": {
    title: "سنگ گرانیت چیست؟ ویژگی‌ها، کاربردها و انواع گرانیت | سنگ حسن",
    description: "راهنمای کامل سنگ گرانیت؛ بررسی ویژگی‌ها، مزایا، معایب، کاربردها، انواع رنگ و نکات مهم خرید گرانیت برای پروژه‌های ساختمانی.",
    h1: "سنگ گرانیت چیست؟ راهنمای کامل ویژگی‌ها، کاربردها و انواع گرانیت",
    canonical: "/fa/blogs/everything-about-granite-stone"
  },
  "everything-about-travertine-stone": {
    title: "سنگ تراورتن چیست؟ انواع، معادن، کاربردها و ویژگی‌ها | سنگ حسن",
    description: "راهنمای کامل سنگ تراورتن؛ بررسی انواع تراورتن، ویژگی‌ها، مزایا، معایب، کاربرد در نما و ساختمان و نکات مهم خرید سنگ تراورتن.",
    h1: "سنگ تراورتن چیست؟ راهنمای کامل انواع، معادن و کاربردها",
    canonical: "/fa/blogs/everything-about-travertine-stone"
  }
};

function articleSeoOverride(locale, slug) {
  return locale === "fa" ? faArticleSeoOverrides[slug] || null : null;
}
const catalogLocaleMeta = {
  en: {
    locale: "en_US",
    home: "Home",
    products: "Products",
    hubTitle: "Natural Stone Categories | SangeHassan",
    hubDescription: "Browse natural stone categories and find the right stone for architectural and building projects.",
    hubH1: "Natural stone categories",
    hubIntro: "Choose a stone category to view available products, colors, finishes and applications."
  },
  fa: {
    locale: "fa_IR",
    home: "خانه",
    products: "محصولات",
    hubTitle: "انواع سنگ ساختمانی | تراورتن، گرانیت، مرمریت و سنگ چینی",
    hubDescription: "مشاهده و بررسی دسته‌بندی انواع سنگ ساختمانی و طبیعی و انتخاب سنگ مناسب پروژه.",
    hubH1: "دسته‌بندی انواع سنگ ساختمانی",
    hubIntro: "برای مشاهده محصولات، رنگ‌ها، فرآوری‌ها و کاربردهای هر سنگ، دسته‌بندی موردنظر را انتخاب کنید."
  },
  ar: {
    locale: "ar_SA",
    home: "الرئيسية",
    products: "المنتجات",
    hubTitle: "أنواع الحجر الطبيعي | سانج حسن",
    hubDescription: "تصفح فئات الحجر الطبيعي واختيار الحجر المناسب للمشاريع المعمارية والإنشائية.",
    hubH1: "فئات الحجر الطبيعي",
    hubIntro: "اختر فئة الحجر لعرض المنتجات والألوان والتشطيبات والاستخدامات المتاحة."
  }
};

function catalogPath(locale, suffix = "") {
  return `/${locale}/products${suffix}`;
}

function catalogAlternates(suffix = "") {
  return [
    ...catalogLocales.map((lang) => ({ lang, path: catalogPath(lang, suffix) })),
    { lang: "x-default", path: catalogPath("en", suffix) }
  ];
}

function localizedField(item, field, locale) {
  return item?.[`${field}_${locale}`] || item?.[`${field}_en`] || item?.[`${field}_fa`] || item?.[`${field}_ar`] || item?.[field] || "";
}

const staticRoutes = [
  {
    path: "/",
    title: "SangeHassan | Natural Stone Supply & Production",
    description:
      "Integrated natural stone supply and production network, from quarry blocks to finished stone products for professional projects, B2B, and export.",
    schemaType: "WebPage",
    image: defaultShareImage,
    changefreq: "weekly",
    priority: 1
  },
  {
    path: "/products",
    title: "Natural Stone Products | SangeHassan",
    description:
      "Browse SangeHassan natural stone products, including slabs, tiles, and finished stones for building projects, B2B supply, and export.",
    schemaType: "CollectionPage",
    breadcrumbName: "Products",
    image: defaultShareImage,
    changefreq: "weekly",
    priority: 0.9
  },
  ...catalogLocales.map((lang) => ({
    path: catalogPath(lang),
    title: catalogLocaleMeta[lang].hubTitle,
    description: catalogLocaleMeta[lang].hubDescription,
    schemaType: "CollectionPage",
    breadcrumbName: catalogLocaleMeta[lang].products,
    image: defaultShareImage,
    lang,
    locale: catalogLocaleMeta[lang].locale,
    alternates: catalogAlternates(),
    changefreq: "weekly",
    priority: 0.9
  })),
  {
    path: "/blocks",
    title: "Stone Blocks | SangeHassan",
    description:
      "Browse SangeHassan natural stone blocks for project supply, slab production, and wholesale B2B sourcing.",
    schemaType: "CollectionPage",
    breadcrumbName: "Stone Blocks",
    image: defaultShareImage,
    changefreq: "weekly",
    priority: 0.85
  },
  {
    path: "/projects",
    title: "Projects | SangeHassan",
    description:
      "Explore SangeHassan's completed stone projects across facade and interior applications with real-world execution results.",
    schemaType: "CollectionPage",
    breadcrumbName: "Projects",
    image: defaultShareImage,
    changefreq: "weekly",
    priority: 0.8
  },
  ...catalogLocales.map((lang) => ({
    path: `/${lang}/blogs`,
    title: blogLocaleMeta[lang].title,
    description: blogLocaleMeta[lang].description,
    schemaType: "Blog",
    breadcrumbName: blogLocaleMeta[lang].articles,
    image: defaultShareImage,
    lang,
    locale: blogLocaleMeta[lang].locale,
    alternates: [...catalogLocales.map((code) => ({ lang: code, path: `/${code}/blogs` })), { lang: "x-default", path: "/en/blogs" }],
    changefreq: "weekly",
    priority: 0.75
  })),
  {
    path: "/blogs",
    canonical: "/en/blogs",
    title: blogLocaleMeta.en.title,
    description: blogLocaleMeta.en.description,
    schemaType: "Blog",
    robots: "noindex,follow",
    sitemap: false
  },
  {
    path: "/about",
    title: "About Us | SangeHassan",
    description:
      "SangeHassan is an integrated natural stone supply and production network, from quarry blocks to finished products with B2B reliability and export focus.",
    schemaType: "AboutPage",
    breadcrumbName: "About Us",
    image: defaultShareImage,
    changefreq: "monthly",
    priority: 0.7
  },
  {
    path: "/ads",
    title: "Trade Board | SangeHassan",
    description: "Private members-only space for posting and browsing stone sale offers.",
    schemaType: "CollectionPage",
    breadcrumbName: "Trade Board",
    image: defaultShareImage,
    changefreq: "daily",
    priority: 0.65
  }
];

function absoluteUrl(value = "") {
  if (!value) return "";
  if (/^https?:\/\//i.test(value)) return value;
  return `${siteUrl}${value.startsWith("/") ? "" : "/"}${value}`;
}

function protectedProductImagePath(value = "") {
  if (!value || /^https?:\/\//i.test(value)) return value;

  const protectedPath = value.startsWith("/images/products/")
    ? value.replace(/^\/images\//, "/protected-images/")
    : value.startsWith("/protected-images/products/")
      ? value
      : "";
  if (!protectedPath) return value;

  const [urlWithoutHash, hash = ""] = protectedPath.split("#");
  const [pathname, query = ""] = urlWithoutHash.split("?");
  const params = new URLSearchParams(query);
  params.set("pv", protectedImageRev);
  const nextQuery = params.toString();
  return `${pathname}${nextQuery ? `?${nextQuery}` : ""}${hash ? `#${hash}` : ""}`;
}

function escapeAttr(value = "") {
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

function escapeXml(value = "") {
  return escapeAttr(value).replace(/'/g, "&apos;");
}

function safeJsonLd(payload) {
  return JSON.stringify(payload).replace(/</g, "\\u003c");
}

function safeScriptJson(payload) {
  return JSON.stringify(payload)
    .replace(/</g, "\\u003c")
    .replace(/\u2028/g, "\\u2028")
    .replace(/\u2029/g, "\\u2029");
}

function stripHTML(value = "") {
  return String(value).replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim();
}

function truncate(value = "", limit = 160) {
  const text = stripHTML(value);
  if (text.length <= limit) return text;
  return `${text.slice(0, limit - 1).trim()}…`;
}

function normalizeDate(value) {
  if (!value) return sitemapLastmod;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return sitemapLastmod;
  return date.toISOString().slice(0, 10);
}

function organizationJsonLd() {
  return {
    "@type": "Organization",
    "@id": organizationId,
    name: "SangeHassan",
    alternateName: "سنگ حسن",
    url: siteUrl,
    logo: absoluteUrl("/favicon.png"),
    sameAs: [
      "https://linkedin.com/company/sangehassan",
      "https://instagram.com/sangehassan_com",
      "https://t.me/sangehassan_com"
    ]
  };
}

function websiteJsonLd(lang = "en") {
  return {
    "@type": "WebSite",
    "@id": websiteId,
    name: "SangeHassan",
    url: absoluteUrl("/"),
    inLanguage: lang,
    publisher: {
      "@id": organizationId
    }
  };
}

function routeBreadcrumbs(route) {
  const homeName = route.lang === "fa" ? "خانه" : route.lang === "ar" ? "الرئيسية" : "Home";
  if (route.path === "/") {
    return [{ name: homeName, path: "/" }];
  }

  if (route.breadcrumbs?.length) {
    return [{ name: homeName, path: "/" }, ...route.breadcrumbs];
  }

  return [
    { name: homeName, path: "/" },
    { name: route.breadcrumbName || route.title.replace(/\s*\|\s*SangeHassan.*$/i, ""), path: route.path }
  ];
}

function breadcrumbJsonLd(route) {
  const items = routeBreadcrumbs(route);
  const canonicalPath = route.canonical || route.path;

  return {
    "@type": "BreadcrumbList",
    "@id": `${absoluteUrl(canonicalPath)}#breadcrumb`,
    itemListElement: items.map((item, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: item.name,
      item: absoluteUrl(item.path)
    }))
  };
}

function pageJsonLd(route) {
  const canonical = absoluteUrl(route.canonical || route.path);
  const image = route.image ? absoluteUrl(route.image) : "";
  const page = {
    "@type": route.schemaType || "WebPage",
    "@id": `${canonical}#webpage`,
    url: canonical,
    name: route.title,
    description: route.description,
    inLanguage: route.lang || "en",
    isPartOf: {
      "@id": websiteId
    },
    publisher: {
      "@id": organizationId
    },
    breadcrumb: {
      "@id": `${canonical}#breadcrumb`
    }
  };

  if (image) {
    page.primaryImageOfPage = {
      "@type": "ImageObject",
      url: image
    };
  }

  if (route.schemaType === "Article" || route.schemaType === "BlogPosting") {
    page.headline = route.headline || route.title;
    page.author = {
      "@id": organizationId
    };
    page.publisher = {
      "@id": organizationId
    };
    if (route.datePublished) {
      page.datePublished = route.datePublished;
    }
    page.dateModified = normalizeDate(route.lastmod);
    if (route.schemaLanguage) {
      page.inLanguage = route.schemaLanguage;
    }
    page.mainEntityOfPage = canonical;
    if (image) {
      page.image = image;
    }
  }

  if (route.schemaType === "AboutPage") {
    page.mainEntity = {
      "@id": organizationId
    };
  }

  if (route.mainEntity) {
    page.mainEntity = route.mainEntity;
  }

  return page;
}

function routeJsonLd(route) {
  return {
    "@context": "https://schema.org",
    "@graph": [
      organizationJsonLd(),
      websiteJsonLd(route.lang || "en"),
      pageJsonLd(route),
      breadcrumbJsonLd(route),
      ...(route.structuredData || [])
    ]
  };
}

function buildHead(route) {
  const canonical = absoluteUrl(route.canonical || route.path);
  const image = route.image ? absoluteUrl(route.image) : "";
  const robots = route.robots || defaultRobots;
  const locale = route.locale || "en_US";
  const type = route.type || "website";
  const tags = [
    `<title>${escapeAttr(route.title)}</title>`,
    `<meta name="description" content="${escapeAttr(route.description)}" />`,
    `<meta name="robots" content="${escapeAttr(robots)}" />`,
    `<link rel="canonical" href="${escapeAttr(canonical)}" />`,
    `<meta property="og:type" content="${escapeAttr(type)}" />`,
    `<meta property="og:title" content="${escapeAttr(route.title)}" />`,
    `<meta property="og:description" content="${escapeAttr(route.description)}" />`,
    `<meta property="og:url" content="${escapeAttr(canonical)}" />`,
    `<meta property="og:locale" content="${escapeAttr(locale)}" />`,
    `<meta name="twitter:card" content="${image ? "summary_large_image" : "summary"}" />`,
    `<meta name="twitter:title" content="${escapeAttr(route.title)}" />`,
    `<meta name="twitter:description" content="${escapeAttr(route.description)}" />`
  ];

  for (const alternate of route.alternates || []) {
    tags.push(`<link rel="alternate" hreflang="${escapeAttr(alternate.lang)}" href="${escapeAttr(absoluteUrl(alternate.path))}" />`);
  }

  if (image) {
    tags.push(`<meta property="og:image" content="${escapeAttr(image)}" />`);
    tags.push(`<meta name="twitter:image" content="${escapeAttr(image)}" />`);
  }

  tags.push(`<script type="application/ld+json">${safeJsonLd(routeJsonLd(route))}</script>`);

  return tags.join("\n  ");
}

function injectRouteHtml(template, route, appHtml) {
  const prerenderDataScript = route.prerenderData
    ? `  <script>window.__SH_PRERENDER_DATA__=${safeScriptJson(route.prerenderData)};</script>\n`
    : "";

  return template
    .replace(/<html[^>]*>/i, `<html lang="${escapeAttr(route.lang || "en")}" dir="${route.lang === "fa" || route.lang === "ar" ? "rtl" : "ltr"}">`)
    .replace(/<title>[\s\S]*?<\/title>\s*/i, "")
    .replace("</head>", `  ${buildHead(route)}\n</head>`)
    .replace('<div id="root"></div>', `<div id="root">${appHtml}</div>`)
    .replace("</body>", `${prerenderDataScript}</body>`);
}

async function loadAssetReplacements() {
  const rawManifest = await fs.readFile(manifestPath, "utf8");
  const manifest = JSON.parse(rawManifest);
  const replacements = [];

  for (const entry of Object.values(manifest)) {
    if (!entry?.src || !entry?.file) continue;

    const absoluteSource = path.resolve(rootDir, entry.src).replace(/\\/g, "/");
    replacements.push([`/@fs${absoluteSource}`, `/${entry.file}`]);
  }

  return replacements;
}

function replaceAssetUrls(html, replacements) {
  return replacements.reduce((output, [from, to]) => output.split(from).join(to), html);
}

function routeOutputPaths(routePath) {
  if (routePath === "/") {
    return [path.join(distDir, "index.html")];
  }

  const cleanPath = routePath.replace(/^\/+|\/+$/g, "");
  return [
    path.join(distDir, `${cleanPath}.html`),
    path.join(distDir, cleanPath, "index.html")
  ];
}

async function writeRoute(routePath, html) {
  const outputs = routeOutputPaths(routePath);
  await Promise.all(
    outputs.map(async (outputPath) => {
      await fs.mkdir(path.dirname(outputPath), { recursive: true });
      await fs.writeFile(outputPath, html);
    })
  );
}

function isIndexable(route) {
  const robots = route.robots || defaultRobots;
  return route.sitemap !== false && !/\bnoindex\b/i.test(robots);
}

function sitemapRoutes(routes) {
  const uniqueRoutes = new Map();

  for (const route of routes) {
    if (!isIndexable(route)) continue;
    uniqueRoutes.set(route.path, route);
  }

  return [...uniqueRoutes.values()];
}

function sitemapXml(routes) {
  const entries = sitemapRoutes(routes)
    .map((route) => {
      const loc = absoluteUrl(route.path);
      const lastmod = normalizeDate(route.lastmod);
      const changefreq = route.changefreq || "weekly";
      const priority = route.priority ?? (route.path === "/" ? 1 : 0.7);

      return [
        "  <url>",
        `    <loc>${escapeXml(loc)}</loc>`,
        ...(route.alternates || []).map((alternate) => `    <xhtml:link rel="alternate" hreflang="${escapeXml(alternate.lang)}" href="${escapeXml(absoluteUrl(alternate.path))}" />`),
        `    <lastmod>${escapeXml(lastmod)}</lastmod>`,
        `    <changefreq>${escapeXml(changefreq)}</changefreq>`,
        `    <priority>${escapeXml(priority.toFixed(2))}</priority>`,
        "  </url>"
      ].join("\n");
    })
    .join("\n");

  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">',
    entries,
    "</urlset>",
    ""
  ].join("\n");
}

function robotsTxt() {
  return [
    "User-agent: *",
    "Allow: /",
    "Disallow: /api/",
    "Disallow: /panel/",
    "Disallow: /login",
    "Disallow: /signup",
    "Disallow: /profile",
    "Disallow: /gallery",
    "Disallow: /ads/new",
    "",
    `Sitemap: ${absoluteUrl("/sitemap.xml")}`,
    ""
  ].join("\n");
}

async function writeSearchFiles(routes) {
  await Promise.all([
    fs.writeFile(path.join(distDir, "sitemap.xml"), sitemapXml(routes)),
    fs.writeFile(path.join(distDir, "robots.txt"), robotsTxt())
  ]);
}

async function fetchApi(pathname) {
  if (!prerenderApiBase) return null;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), fetchTimeoutMs);
  const response = await fetch(`${prerenderApiBase}${pathname}`, { signal: controller.signal }).finally(() => {
    clearTimeout(timeout);
  });
  if (!response.ok) {
    throw new Error(`${pathname} returned ${response.status}`);
  }

  const payload = await response.json();
  return payload?.data ?? payload;
}

async function detailRoutes(items, routeFactory, detailPath, dataKey, label) {
  if (!Array.isArray(items)) return [];

  const routes = await Promise.all(
    items.map(async (item) => {
      let data = item;
      try {
        data = (await fetchApi(detailPath(item))) || item;
      } catch (error) {
        console.warn(`prerender: ${label} detail skipped for ${detailPath(item)}: ${error.message}`);
      }

      const createdRoutes = routeFactory(data);
      if (!createdRoutes) return [];
      return (Array.isArray(createdRoutes) ? createdRoutes : [createdRoutes]).map((route) => ({
        ...route,
        prerenderData: { [dataKey]: dataKey === "product" ? protectProductPrerenderData(data) : data }
      }));
    })
  );

  return routes.flat().filter(Boolean);
}

function firstImage(item) {
  if (Array.isArray(item?.images) && item.images.length) {
    const first = item.images[0];
    if (typeof first === "string") return first;
    return first?.image_url || "";
  }
  return item?.image_url || item?.cover_image_url || "";
}

function shareImage(item) {
  return firstImage(item) || defaultShareImage;
}

function productShareImage(product) {
  return protectedProductImagePath(firstImage(product)) || defaultShareImage;
}

function productOfferPrice(product) {
  if (!product?.is_popular) return 0;
  const value = typeof product.price === "number" ? product.price : Number(product.price);
  return Number.isFinite(value) && value > 0 ? Math.round(value) : 0;
}

function productOfferSchemaNote(lang = "en") {
  if (lang === "fa") {
    return "این قیمت کف قیمت پیشنهادی است و برای استعلام معتبر باید با سنگ حسن تماس گرفته شود.";
  }
  if (lang === "ar") {
    return "هذا سعر ابتدائي، ويجب التواصل مع سانج حسن للحصول على السعر الحالي الموثوق.";
  }
  return "This is a starting floor price; contact SangeHassan for the current valid quote.";
}

function productOfferJsonLd(product, routePath, lang = "en") {
  const price = productOfferPrice(product);
  if (price <= 0) return undefined;
  const url = absoluteUrl(routePath);
  return {
    "@type": "Offer",
    url,
    priceCurrency: "IRR",
    price: String(price),
    availability: "https://schema.org/InStock",
    itemCondition: "https://schema.org/NewCondition",
    seller: {
      "@type": "Organization",
      name: "SangeHassan"
    },
    priceSpecification: {
      "@type": "UnitPriceSpecification",
      name: "Offer",
      price: String(price),
      priceCurrency: "IRR",
      description: productOfferSchemaNote(lang)
    }
  };
}

function protectProductPrerenderData(product) {
  if (!product || typeof product !== "object") return product;

  return {
    ...product,
    image_url: protectedProductImagePath(product.image_url || ""),
    images: Array.isArray(product.images)
      ? product.images.map((image) => {
          if (typeof image === "string") return protectedProductImagePath(image);
          if (image && typeof image === "object") {
            return {
              ...image,
              image_url: protectedProductImagePath(image.image_url || "")
            };
          }
          return image;
        })
      : product.images
  };
}

function productRoute(product) {
  const slug = product?.slug;
  if (!slug) return null;

  const localizedRoutes = catalogLocales.map((locale) => {
    const meta = catalogLocaleMeta[locale];
    const routePath = catalogPath(locale, `/${slug}`);
    const title = localizedField(product, "title", locale) || slug;
    const localizedDescription = locale === "fa"
      ? product.description_html_fa || product.short_description_html_fa
      : locale === "ar"
        ? product.description_html_ar || product.short_description_html_ar
        : product.description_html_en || product.short_description_html_en;
    const description = truncate(localizedDescription || product.description_html || product.short_description_html || product.description || "") ||
      (locale === "fa"
        ? "صفحه معرفی محصول سنگ طبیعی شامل تصاویر، مشخصات و اطلاعات کاربردی پروژه."
        : locale === "ar"
          ? "صفحة منتج الحجر الطبيعي مع الصور والمواصفات ومعلومات الاستخدام في المشاريع."
          : "Detailed natural stone product page with images, specifications and project references.");
    const category = product.category || product.categories?.[0];
    const categorySlug = category?.slug;
    const categoryTitle = localizedField(category, "title", locale);
    return {
      path: routePath,
      title: `${title} | SangeHassan`,
      description,
      image: productShareImage(product),
      type: "product",
      schemaType: "ItemPage",
      routeKind: "product-detail",
      lang: locale,
      locale: meta.locale,
      alternates: catalogAlternates(`/${slug}`),
      breadcrumbs: [
        { name: meta.products, path: catalogPath(locale) },
        ...(categorySlug && categoryTitle ? [{ name: categoryTitle, path: catalogPath(locale, `/${categorySlug}`) }] : []),
        { name: title, path: routePath }
      ],
      changefreq: "weekly",
      priority: 0.75,
      lastmod: product.updated_at || product.updatedAt || product.created_at,
      mainEntity: {
        "@type": "Product",
        "@id": `${absoluteUrl(routePath)}#product`,
        name: title,
        description,
        image: firstImage(product) ? absoluteUrl(productShareImage(product)) : undefined,
        url: absoluteUrl(routePath),
        brand: {
          "@type": "Brand",
          name: "SangeHassan"
        },
        offers: productOfferJsonLd(product, routePath, locale)
      }
    };
  });

  const englishRoute = localizedRoutes[0];
  return [
    ...localizedRoutes,
    {
      ...englishRoute,
      path: `/products/${slug}`,
      canonical: englishRoute.path,
      robots: "noindex,follow",
      sitemap: false
    }
  ];
}

function catalogHubRoute(hub, locale) {
  const seo = hub?.seo;
  if (!seo) return null;
  const meta = catalogLocaleMeta[locale];
  const categories = Array.isArray(hub.categories) ? hub.categories : [];
  const localizedHub = { ...hub, locale };
  return {
    path: catalogPath(locale),
    title: seo.title,
    description: seo.description,
    image: categories.find((category) => category.preview_image)?.preview_image || defaultShareImage,
      schemaType: "CollectionPage",
      routeKind: "catalog-hub",
    breadcrumbName: meta.products,
    lang: locale,
    locale: meta.locale,
    alternates: catalogAlternates(),
    changefreq: "weekly",
    priority: 0.9,
    prerenderData: { catalogHub: localizedHub }
  };
}

function catalogRouteSuffix(routeMeta) {
  const category = `/${routeMeta.category_slug}`;
  return routeMeta.type === "facet" ? `${category}/${routeMeta.facet}/${routeMeta.value}` : category;
}

function catalogPageRoute(page, routeMeta, locale) {
  if (!page?.category?.slug || !page?.seo) return null;
  const meta = catalogLocaleMeta[locale];
  const suffix = catalogRouteSuffix(routeMeta);
  const routePath = catalogPath(locale, suffix);
  const products = Array.isArray(page.products) ? page.products : [];
  const structuredData = products.length ? [{
    "@type": "ItemList",
    "@id": `${absoluteUrl(routePath)}#products`,
    itemListElement: products.map((product, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: localizedField(product, "title", locale) || product.slug,
      url: absoluteUrl(catalogPath(locale, `/${product.slug}`))
    }))
  }] : [];
  const localizedPage = {
    ...page,
    locale,
    seo: { ...page.seo, canonical: routePath }
  };
  return {
    path: routePath,
    title: page.seo.title,
    description: page.seo.description,
    image: page.category.preview_image || page.category.image_url || defaultShareImage,
    schemaType: "CollectionPage",
    routeKind: routeMeta.type === "facet" ? "catalog-facet" : "catalog-category",
    breadcrumbs: [
      { name: meta.products, path: catalogPath(locale) },
      { name: localizedField(page.category, "title", locale), path: catalogPath(locale, `/${page.category.slug}`) },
      ...(routeMeta.type === "facet" && page.selected_facet
        ? [{ name: localizedField(page.selected_facet, "label", locale), path: routePath }]
        : [])
    ],
    lang: locale,
    locale: meta.locale,
    alternates: catalogAlternates(suffix),
    robots: routeMeta.indexable && page.indexable ? "index,follow" : "noindex,follow",
    sitemap: Boolean(routeMeta.indexable && page.indexable),
    changefreq: "weekly",
    priority: routeMeta.type === "category" ? 0.82 : 0.7,
    lastmod: page.category.updated_at || page.category.created_at,
    structuredData,
    prerenderData: { catalogPage: localizedPage }
  };
}

const legacyCatalogFacets = [
  { route: "color", taxonomy: "tone", label_en: "Color", label_fa: "رنگ", label_ar: "اللون" },
  { route: "application", taxonomy: "use_case_application", label_en: "Application", label_fa: "کاربرد", label_ar: "الاستخدام" },
  { route: "finish", taxonomy: "finishes", label_en: "Finish", label_fa: "نوع فرآوری", label_ar: "التشطيب" },
  { route: "form", taxonomy: "use_case_form", label_en: "Form", label_fa: "فرم عرضه", label_ar: "الشكل" },
  { route: "origin", taxonomy: "mines", label_en: "Origin", label_fa: "مبدأ یا معدن", label_ar: "المنشأ" },
  { route: "pattern", taxonomy: "pattern", label_en: "Pattern", label_fa: "نوع موج و طرح", label_ar: "النمط" },
  { route: "availability", taxonomy: "availability", label_en: "Availability", label_fa: "وضعیت تأمین", label_ar: "التوفر" }
];

function legacyCategoryProducts(products, category) {
  return products.filter((product) => {
    if (product.is_active === false || product.is_indexable === false) return false;
    if (product.category?.slug === category.slug) return true;
    return Array.isArray(product.categories) && product.categories.some((item) => item?.slug === category.slug);
  });
}

function legacyFacetHeading(categoryTitle, facet, valueLabel, locale) {
  const category = locale === "fa" && !categoryTitle.startsWith("سنگ ")
    ? `سنگ ${categoryTitle}`
    : locale === "ar" && !categoryTitle.startsWith("حجر ") ? `حجر ${categoryTitle}` : categoryTitle;
  if (locale === "fa") {
    if (facet === "application") return `${category} مناسب ${valueLabel}`;
    if (facet === "finish") return `${category} با فرآوری ${valueLabel}`;
    if (facet === "origin") return `${category} معدن ${valueLabel}`;
    if (facet === "pattern") return `${category} با طرح ${valueLabel}`;
  }
  if (locale === "ar") {
    if (facet === "application") return `${category} مناسب لـ ${valueLabel}`;
    if (facet === "finish") return `${category} بتشطيب ${valueLabel}`;
    if (facet === "origin") return `${category} من ${valueLabel}`;
    if (facet === "pattern") return `${category} بنمط ${valueLabel}`;
  }
  return `${category} ${valueLabel}`.trim();
}

function legacyCatalogPage(category, products, selectedFacet, locale) {
  const facetGroups = legacyCatalogFacets.map((definition) => {
    const values = new Map();
    for (const product of products) {
      for (const term of product.terms || []) {
        if (term.taxonomy !== definition.taxonomy || term.is_active === false) continue;
        const current = values.get(term.key) || { ...term, count: 0 };
        current.count += 1;
        values.set(term.key, current);
      }
    }
    return {
      key: definition.route,
      taxonomy: definition.taxonomy,
      label_en: definition.label_en,
      label_fa: definition.label_fa,
      label_ar: definition.label_ar,
      values: [...values.values()].map((term) => ({
        key: term.key,
        label_en: term.label_en,
        label_fa: term.label_fa || term.label_en || term.key,
        label_ar: term.label_ar,
        count: term.count,
        is_indexable: term.is_indexable !== false
      }))
    };
  }).filter((facet) => facet.values.length > 0);

  const categoryTitle = localizedField(category, "title", locale) || category.slug;
  const heading = selectedFacet
    ? legacyFacetHeading(categoryTitle, selectedFacet.facet, localizedField(selectedFacet.value, "label", locale), locale)
    : legacyFacetHeading(categoryTitle, "", "", locale);
  const canonical = selectedFacet
    ? catalogPath(locale, `/${category.slug}/${selectedFacet.facet}/${selectedFacet.value.key}`)
    : catalogPath(locale, `/${category.slug}`);
  const visibleProducts = selectedFacet
    ? products.filter((product) => (product.terms || []).some((term) => term.taxonomy === selectedFacet.taxonomy && term.key === selectedFacet.value.key))
    : products;
  const indexable = category.is_indexable !== false && (!selectedFacet || selectedFacet.value.is_indexable !== false && visibleProducts.length >= 2);
  const categorySeoTitle = localizedField(category, "seo_title", locale);
  const categorySeoDescription = localizedField(category, "seo_description", locale);
  const categoryIntro = localizedField(category, "intro", locale);
  const generatedDescription = locale === "fa"
    ? `مشاهده انواع ${heading}، رنگ‌ها، فرآوری‌ها، کاربردها و محصولات موجود برای پروژه‌های ساختمانی.`
    : locale === "ar"
      ? `تصفح أنواع ${heading} والألوان والتشطيبات والاستخدامات المتاحة للمشاريع الإنشائية.`
      : `Browse ${heading} products, colors, finishes and applications for building projects.`;
  return {
    locale,
    category: {
      ...category,
      product_count: products.length,
      preview_image: category.image_url || products.find((product) => product.image_url)?.image_url || ""
    },
    seo: {
      title: selectedFacet
        ? `${heading} | ${locale === "fa" ? "سنگ حسن" : locale === "ar" ? "سانج حسن" : "SangeHassan"}`
        : categorySeoTitle || `${heading} | ${locale === "fa" ? "انواع، کاربرد و خرید" : locale === "ar" ? "الأنواع والاستخدامات" : "SangeHassan"}`,
      description: categorySeoDescription || generatedDescription,
      h1: heading,
      intro: categoryIntro || generatedDescription,
      canonical,
      robots: indexable ? "index,follow" : "noindex,follow"
    },
    facets: facetGroups,
    selected_filters: selectedFacet ? { [selectedFacet.facet]: [selectedFacet.value.key] } : {},
    products: visibleProducts.slice(0, 24),
    pagination: { limit: 24, offset: 0, total: visibleProducts.length },
    indexable,
    selected_facet: selectedFacet?.value,
    selected_facet_key: selectedFacet?.facet || "",
    related_categories: []
  };
}

async function loadLegacyCatalogRoutes(missingHubLocales = catalogLocales) {
  const [categories, products] = await Promise.all([
    fetchApi("/api/categories"),
    fetchApi("/api/products?limit=500&offset=0")
  ]);
  const activeCategories = (Array.isArray(categories) ? categories : []).filter((category) => category.is_active !== false && !category.parent_id);
  const allProducts = Array.isArray(products) ? products : [];
  const routes = [];
  const hubCategories = activeCategories.map((category) => {
    const categoryProducts = legacyCategoryProducts(allProducts, category);
    return {
      ...category,
      product_count: categoryProducts.length,
      preview_image: category.image_url || categoryProducts.find((product) => product.image_url)?.image_url || ""
    };
  }).filter((category) => category.product_count > 0);

  for (const locale of missingHubLocales) {
    const meta = catalogLocaleMeta[locale];
    routes.push(catalogHubRoute({
      locale,
      categories: hubCategories,
      seo: {
        title: meta.hubTitle,
        description: meta.hubDescription,
        h1: meta.hubH1,
        intro: meta.hubIntro,
        canonical: catalogPath(locale),
        robots: "index,follow"
      }
    }, locale));
  }

  for (const locale of catalogLocales) {
    for (const category of hubCategories) {
      const categoryProducts = legacyCategoryProducts(allProducts, category);
      const categoryPage = legacyCatalogPage(category, categoryProducts, null, locale);
      routes.push(catalogPageRoute(categoryPage, {
        category_slug: category.slug,
        type: "category",
        indexable: category.is_indexable !== false
      }, locale));
      for (const facet of categoryPage.facets) {
        for (const value of facet.values) {
          const page = legacyCatalogPage(category, categoryProducts, { facet: facet.key, taxonomy: facet.taxonomy, value }, locale);
          routes.push(catalogPageRoute(page, {
            category_slug: category.slug,
            facet: facet.key,
            value: value.key,
            type: "facet",
            indexable: page.indexable
          }, locale));
        }
      }
    }
  }
  return routes.filter(Boolean);
}

async function loadCatalogRoutes() {
  const routes = [];
  const missingHubLocales = [];
  for (const locale of catalogLocales) {
    try {
      const hub = await fetchApi(`/api/catalog/categories?locale=${locale}`);
      const hubRoute = catalogHubRoute(hub, locale);
      if (hubRoute) routes.push(hubRoute);
      else missingHubLocales.push(locale);
    } catch (error) {
      missingHubLocales.push(locale);
      console.warn(`prerender: ${locale} product hub fallback used: ${error.message}`);
    }
  }

  let routeList = [];
  try {
    routeList = (await fetchApi("/api/catalog/routes")) || [];
  } catch (error) {
    console.warn(`prerender: catalog API unavailable, using existing product APIs: ${error.message}`);
    try {
      return [...routes, ...(await loadLegacyCatalogRoutes(missingHubLocales))];
    } catch (fallbackError) {
      console.warn(`prerender: catalog fallback skipped: ${fallbackError.message}`);
      return routes;
    }
  }

  for (const locale of catalogLocales) {
    for (const routeMeta of routeList) {
      try {
        const endpoint = routeMeta.type === "facet"
          ? `/api/catalog/categories/${routeMeta.category_slug}/${routeMeta.facet}/${routeMeta.value}?locale=${locale}&limit=24&offset=0`
          : `/api/catalog/categories/${routeMeta.category_slug}?locale=${locale}&limit=24&offset=0`;
        const page = await fetchApi(endpoint);
        const route = catalogPageRoute(page, routeMeta, locale);
        if (route) routes.push(route);
      } catch (error) {
        console.warn(`prerender: ${locale} catalog page skipped for ${routeMeta.category_slug}: ${error.message}`);
      }
    }
  }
  return routes;
}

function blockRoute(block) {
  const slug = block?.slug;
  if (!slug) return null;

  const title = block.title_en || block.title_fa || block.title_ar || slug;
  const details = [block.stone_type, block.quarry, block.dimensions, block.description].filter(Boolean).join(" | ");
  const description =
    truncate(details) ||
    "Natural stone block detail from SangeHassan with quarry, dimensions, weight, and availability information.";

  return {
    path: `/blocks/${slug}`,
    title: `${title} | Stone Blocks | SangeHassan`,
    description,
    image: shareImage(block),
    type: "product",
    schemaType: "ItemPage",
    breadcrumbs: [
      { name: "Stone Blocks", path: "/blocks" },
      { name: title, path: `/blocks/${slug}` }
    ],
    changefreq: "weekly",
    priority: 0.7,
    lastmod: block.updated_at || block.updatedAt || block.created_at,
    mainEntity: {
      "@type": "Product",
      "@id": `${absoluteUrl(`/blocks/${slug}`)}#product`,
      name: title,
      description,
      image: firstImage(block) ? absoluteUrl(firstImage(block)) : undefined,
      url: absoluteUrl(`/blocks/${slug}`),
      brand: {
        "@type": "Brand",
        name: "SangeHassan"
      }
    }
  };
}

function projectRoute(project) {
  const id = project?.id;
  if (!id) return null;

  const title = project.title_en || project.title_fa || project.title_ar || `Project ${id}`;
  const description =
    truncate(project.description_en || project.description || project.description_fa || project.description_ar || "") ||
    "Detailed view of a completed SangeHassan project including gallery and project description.";

  return {
    path: `/projects/${id}`,
    title: `${title} | Projects | SangeHassan`,
    description,
    image: project.cover_image_url || defaultShareImage,
    type: "article",
    schemaType: "Article",
    breadcrumbs: [
      { name: "Projects", path: "/projects" },
      { name: title, path: `/projects/${id}` }
    ],
    changefreq: "monthly",
    priority: 0.65,
    lastmod: project.updated_at || project.updatedAt || project.created_at
  };
}

function adRoute(ad) {
  const id = ad?.id;
  if (!id) return null;

  const title = ad.title || ad.stone_type || `Stone offer ${id}`;
  const description = truncate(ad.description || "Stone sale offer on the SangeHassan trade board.");

  return {
    path: `/ads/${id}`,
    title: `${title} | Trade Board | SangeHassan`,
    description,
    image: shareImage(ad),
    type: "article",
    schemaType: "ItemPage",
    breadcrumbs: [
      { name: "Trade Board", path: "/ads" },
      { name: title, path: `/ads/${id}` }
    ],
    changefreq: "daily",
    priority: 0.55,
    lastmod: ad.updated_at || ad.updatedAt || ad.created_at,
    mainEntity: {
      "@type": "Offer",
      "@id": `${absoluteUrl(`/ads/${id}`)}#offer`,
      name: title,
      description,
      url: absoluteUrl(`/ads/${id}`),
      image: firstImage(ad) ? absoluteUrl(firstImage(ad)) : undefined,
      offeredBy: {
        "@id": organizationId
      }
    }
  };
}

function blogRoute(blog, locale) {
  if (!blog?.slug) return null;
  const meta = blogLocaleMeta[locale];
  const routePath = `/${locale}/blogs/${blog.slug}`;
  const seoOverride = articleSeoOverride(locale, blog.slug);
  const description = seoOverride?.description || blog.seo_description || blog.excerpt || truncate(blog.content_html) || blog.title;
  const translations = Array.isArray(blog.translations) ? blog.translations : [];
  const alternateItems = translations.map((item) => ({ lang: item.locale, path: `/${item.locale}/blogs/${item.slug}` }));
  const defaultAlternate = alternateItems.find((item) => item.lang === "en") || alternateItems[0];
  const canonical = seoOverride?.canonical || blog.canonical_url || routePath;
  const headline = seoOverride?.h1 || blog.title;
  return {
    path: routePath,
    canonical,
    title: seoOverride?.title || blog.seo_title || blog.title,
    description,
    image: blog.og_image_url || blog.cover_image_url || defaultShareImage,
    type: "article",
    schemaType: "BlogPosting",
    routeKind: "blog-detail",
    lang: locale,
    locale: meta.locale,
    robots: blog.robots || "index,follow",
    alternates: defaultAlternate ? [...alternateItems, { lang: "x-default", path: defaultAlternate.path }] : alternateItems,
    breadcrumbs: [
      { name: meta.articles, path: `/${locale}/blogs` },
      { name: blog.title, path: routePath }
    ],
    changefreq: "monthly",
    priority: 0.7,
    lastmod: blog.updated_at || blog.published_at || blog.created_at,
    datePublished: blog.published_at || blog.created_at,
    headline,
    schemaLanguage: locale === "fa" ? "fa-IR" : locale,
    prerenderData: { blog }
  };
}

async function loadBlogRoutes() {
  const routes = [];
  const seenDetailKeys = new Set();

  const addBlogRoute = (blog, locale) => {
    const route = blogRoute(blog, locale);
    if (!route) return;
    const key = `${locale}/${blog.slug}`;
    if (seenDetailKeys.has(key)) return;
    seenDetailKeys.add(key);
    routes.push(route);
  };

  for (const locale of catalogLocales) {
    try {
      const blogs = (await fetchApi(`/api/blogs?locale=${locale}`)) || [];
      const meta = blogLocaleMeta[locale];
      routes.push({
        path: `/${locale}/blogs`,
        title: meta.title,
        description: meta.description,
        schemaType: "Blog",
        routeKind: "blog-list",
        image: blogs.find((blog) => blog.cover_image_url)?.cover_image_url || defaultShareImage,
        lang: locale,
        locale: meta.locale,
        alternates: [...catalogLocales.map((code) => ({ lang: code, path: `/${code}/blogs` })), { lang: "x-default", path: "/en/blogs" }],
        changefreq: "weekly",
        priority: 0.75,
        lastmod: blogs[0]?.updated_at || blogs[0]?.published_at,
        prerenderData: { blogs }
      });
      for (const summary of blogs) {
        try {
          const blog = (await fetchApi(`/api/blogs/${locale}/${summary.slug}`)) || summary;
          addBlogRoute(blog, locale);
        } catch (error) {
          console.warn(`prerender: blog skipped for ${locale}/${summary.slug}: ${error.message}`);
        }
      }
    } catch (error) {
      console.warn(`prerender: ${locale} blogs skipped: ${error.message}`);
    }
  }

  for (const slug of requiredFaBlogSlugs) {
    try {
      const blog = await fetchApi(`/api/blogs/fa/${slug}`);
      addBlogRoute(blog, "fa");
    } catch (error) {
      console.warn(`prerender: required fa blog skipped for ${slug}: ${error.message}`);
    }
  }

  return routes;
}

async function loadDynamicRoutes() {
  if (!prerenderApiBase) {
    console.log("prerender: VITE_PRERENDER_API_BASE_URL is not set; dynamic detail pages were skipped.");
    return [];
  }

  try {
    const [products, blocks, projects, ads] = await Promise.all([
      fetchApi("/api/products?limit=500&offset=0").catch((error) => {
        console.warn(`prerender: products skipped: ${error.message}`);
        return [];
      }),
      fetchApi("/api/blocks").catch((error) => {
        console.warn(`prerender: blocks skipped: ${error.message}`);
        return [];
      }),
      fetchApi("/api/projects").catch((error) => {
        console.warn(`prerender: projects skipped: ${error.message}`);
        return [];
      }),
      fetchApi("/api/ads").catch((error) => {
        console.warn(`prerender: ads skipped: ${error.message}`);
        return [];
      })
    ]);

    const [productRoutes, blockRoutes, projectRoutes, adRoutes, catalogRoutes, blogRoutes] = await Promise.all([
      detailRoutes(products, productRoute, (product) => `/api/products/${product.slug}`, "product", "product"),
      detailRoutes(blocks, blockRoute, (block) => `/api/blocks/${block.slug}`, "block", "block"),
      detailRoutes(projects, projectRoute, (project) => `/api/projects/${project.id}`, "project", "project"),
      detailRoutes(ads, adRoute, (ad) => `/api/ads/${ad.id}`, "ad", "ad"),
      loadCatalogRoutes(),
      loadBlogRoutes()
    ]);

    return [...productRoutes, ...blockRoutes, ...projectRoutes, ...adRoutes, ...catalogRoutes, ...blogRoutes];
  } catch (error) {
    console.warn(`prerender: dynamic detail pages skipped: ${error.message}`);
    return [];
  }
}

async function main() {
  const template = await fs.readFile(templatePath, "utf8");
  const assetReplacements = await loadAssetReplacements();
  const vite = await createServer({
    root: rootDir,
    appType: "custom",
    logLevel: "error",
    server: { middlewareMode: true }
  });

  try {
    const { render } = await vite.ssrLoadModule("/src/entry-server.jsx");
    const dynamicRoutes = await loadDynamicRoutes();
    const routeMap = new Map();
    for (const route of [...staticRoutes, ...dynamicRoutes]) {
      routeMap.set(route.path, route);
    }
    const routes = [...routeMap.values()];

    for (const route of routes) {
      const appHtml = render(route.path, route.prerenderData || null);
      const html = replaceAssetUrls(injectRouteHtml(template, route, appHtml), assetReplacements);
      await writeRoute(route.path, html);
    }

    await writeSearchFiles(routes);

    const counts = routes.reduce((out, route) => {
      const type = route.routeKind || "other";
      out[type] = (out[type] || 0) + 1;
      return out;
    }, {});
    console.log(`prerender: wrote ${routes.length} route(s) with canonical origin ${siteUrl}`, counts);
  } finally {
    await vite.close();
  }
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
