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
const defaultRobots = "index,follow,max-image-preview:large";
const sharedAssetUrl = (sourcePath) => `/@fs${path.resolve(rootDir, sourcePath).replace(/\\/g, "/")}`;
const defaultShareImage = sharedAssetUrl("../shared/assets/landing_page/landingpage_blocks_overlay.webp");

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
  {
    path: "/blogs",
    title: "Natural Stone Articles | SangeHassan",
    description:
      "Read SangeHassan articles and guides about choosing, sourcing, and using natural stone in building projects.",
    schemaType: "Blog",
    breadcrumbName: "Articles",
    image: defaultShareImage,
    changefreq: "weekly",
    priority: 0.75
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
    url: siteUrl,
    logo: absoluteUrl("/favicon.png"),
    sameAs: [
      "https://linkedin.com/company/sangehassan",
      "https://instagram.com/sangehassan_com",
      "https://t.me/sangehassan_com"
    ]
  };
}

function websiteJsonLd() {
  return {
    "@type": "WebSite",
    "@id": websiteId,
    name: "SangeHassan",
    url: absoluteUrl("/"),
    inLanguage: "en",
    publisher: {
      "@id": organizationId
    }
  };
}

function routeBreadcrumbs(route) {
  if (route.path === "/") {
    return [{ name: "Home", path: "/" }];
  }

  if (route.breadcrumbs?.length) {
    return [{ name: "Home", path: "/" }, ...route.breadcrumbs];
  }

  return [
    { name: "Home", path: "/" },
    { name: route.breadcrumbName || route.title.replace(/\s*\|\s*SangeHassan.*$/i, ""), path: route.path }
  ];
}

function breadcrumbJsonLd(route) {
  const items = routeBreadcrumbs(route);

  return {
    "@type": "BreadcrumbList",
    "@id": `${absoluteUrl(route.path)}#breadcrumb`,
    itemListElement: items.map((item, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: item.name,
      item: absoluteUrl(item.path)
    }))
  };
}

function pageJsonLd(route) {
  const canonical = absoluteUrl(route.path);
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

  if (route.schemaType === "Article") {
    page.headline = route.title;
    page.author = {
      "@id": organizationId
    };
    page.dateModified = normalizeDate(route.lastmod);
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
      websiteJsonLd(),
      pageJsonLd(route),
      breadcrumbJsonLd(route),
      ...(route.structuredData || [])
    ]
  };
}

function buildHead(route) {
  const canonical = absoluteUrl(route.path);
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
    .replace(/<html[^>]*>/i, `<html lang="en" dir="ltr">`)
    .replace(/<title>[\s\S]*?<\/title>\s*/i, "")
    .replace("</head>", `  ${buildHead(route)}\n</head>`)
    .replace('<div id="root"></div>', `<div id="root">${appHtml}</div>`)
    .replace(/(\s*<script type="module")/, `\n${prerenderDataScript}$1`);
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
        `    <lastmod>${escapeXml(lastmod)}</lastmod>`,
        `    <changefreq>${escapeXml(changefreq)}</changefreq>`,
        `    <priority>${escapeXml(priority.toFixed(2))}</priority>`,
        "  </url>"
      ].join("\n");
    })
    .join("\n");

  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
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

  const response = await fetch(`${prerenderApiBase}${pathname}`);
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

      const route = routeFactory(data);
      if (!route) return null;
      route.prerenderData = { [dataKey]: data };
      return route;
    })
  );

  return routes.filter(Boolean);
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

function productRoute(product) {
  const slug = product?.slug;
  if (!slug) return null;

  const title = product.title_en || product.title_fa || product.title_ar || slug;
  const description =
    truncate(
      product.description_html_en ||
        product.description_html ||
        product.short_description_html_en ||
        product.short_description_html ||
        product.description ||
        ""
    ) ||
    "Detailed natural stone product page from SangeHassan with images, finishes, variants, and sourcing information.";

  return {
    path: `/products/${slug}`,
    title: `${title} | SangeHassan`,
    description,
    image: shareImage(product),
    type: "product",
    schemaType: "ItemPage",
    breadcrumbs: [
      { name: "Products", path: "/products" },
      { name: title, path: `/products/${slug}` }
    ],
    changefreq: "weekly",
    priority: 0.75,
    lastmod: product.updated_at || product.updatedAt || product.created_at,
    mainEntity: {
      "@type": "Product",
      "@id": `${absoluteUrl(`/products/${slug}`)}#product`,
      name: title,
      description,
      image: firstImage(product) ? absoluteUrl(firstImage(product)) : undefined,
      url: absoluteUrl(`/products/${slug}`),
      brand: {
        "@type": "Brand",
        name: "SangeHassan"
      }
    }
  };
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

    const [productRoutes, blockRoutes, projectRoutes, adRoutes] = await Promise.all([
      detailRoutes(products, productRoute, (product) => `/api/products/${product.slug}`, "product", "product"),
      detailRoutes(blocks, blockRoute, (block) => `/api/blocks/${block.slug}`, "block", "block"),
      detailRoutes(projects, projectRoute, (project) => `/api/projects/${project.id}`, "project", "project"),
      detailRoutes(ads, adRoute, (ad) => `/api/ads/${ad.id}`, "ad", "ad")
    ]);

    return [...productRoutes, ...blockRoutes, ...projectRoutes, ...adRoutes];
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
    const routes = [...staticRoutes, ...dynamicRoutes];

    for (const route of routes) {
      const appHtml = render(route.path, route.prerenderData || null);
      const html = replaceAssetUrls(injectRouteHtml(template, route, appHtml), assetReplacements);
      await writeRoute(route.path, html);
    }

    await writeSearchFiles(routes);

    console.log(`prerender: wrote ${routes.length} route(s) with canonical origin ${siteUrl}`);
  } finally {
    await vite.close();
  }
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
