const origin = new URL(process.argv[2] || "https://sangehassan.com").origin;
const concurrency = Math.max(1, Number(process.env.INTERNAL_LINK_VERIFY_CONCURRENCY || 10));
const placeholderPatterns = [
  /URL را من وارد می‌کنم/i,
  /\[لینک داخلی/i,
  /\b(?:TODO|FIXME)\b/i,
  /https?:\/\/example\.com\b/i
];

async function fetchText(url, options = {}) {
  const response = await fetch(url, {
    redirect: options.redirect || "follow",
    headers: { "User-Agent": "SangeHassan-SEO-Verification/1.0" }
  });
  return { response, text: await response.text() };
}

function publicBlogPage(payload) {
  const body = payload?.data ?? payload;
  return {
    items: Array.isArray(body?.items) ? body.items : [],
    total: Number(body?.total || 0),
    limit: Number(body?.limit || 100)
  };
}

async function loadBlogs(locale) {
  const items = [];
  let offset = 0;
  do {
    const { response, text } = await fetchText(`${origin}/api/blogs?locale=${locale}&limit=100&offset=${offset}`);
    if (!response.ok) throw new Error(`${locale} blog API returned ${response.status}`);
    const page = publicBlogPage(JSON.parse(text));
    items.push(...page.items);
    offset += page.limit;
    if (offset >= page.total) break;
  } while (true);
  return items;
}

function articleHTML(html) {
  return html.match(/<article\b[\s\S]*?<\/article>/i)?.[0] || "";
}

function internalLinks(html, baseUrl) {
  const links = [];
  for (const match of html.matchAll(/<a\b[^>]*href=["']([^"']+)["'][^>]*>/gi)) {
    const href = match[1].trim();
    if (!href || href.startsWith("#") || href.startsWith("tel:") || href.startsWith("mailto:")) continue;
    const url = new URL(href, baseUrl);
    if (url.origin === origin) links.push(url.href);
  }
  return [...new Set(links)];
}

async function mapLimit(items, limit, mapper) {
  const results = [];
  let index = 0;
  async function worker() {
    while (index < items.length) {
      const current = index++;
      results[current] = await mapper(items[current]);
    }
  }
  await Promise.all(Array.from({ length: Math.min(limit, items.length) }, worker));
  return results;
}

const locales = ["fa", "en", "ar"];
const summaries = (await Promise.all(locales.map(loadBlogs))).flat();
const pages = await mapLimit(summaries, concurrency, async (blog) => {
  const url = `${origin}/${blog.locale}/blogs/${blog.slug}`;
  const { response, text } = await fetchText(url);
  return { blog, url, status: response.status, html: articleHTML(text) };
});

const issues = [];
const references = new Map();
for (const page of pages) {
  if (page.status !== 200) issues.push(`${page.url}: article returned ${page.status}`);
  if (!page.html) issues.push(`${page.url}: missing <article> content`);
  for (const pattern of placeholderPatterns) {
    if (pattern.test(page.html)) issues.push(`${page.url}: contains placeholder matching ${pattern}`);
  }
  for (const target of internalLinks(page.html, page.url)) {
    references.set(target, [...(references.get(target) || []), page.url]);
  }
}

const targets = await mapLimit([...references.keys()], concurrency, async (url) => {
  try {
    const { response } = await fetchText(url, { redirect: "manual" });
    return { url, status: response.status, location: response.headers.get("location") || "" };
  } catch (error) {
    return { url, status: 0, error: error.message };
  }
});

for (const target of targets) {
  if (target.status === 200) continue;
  const sources = references.get(target.url) || [];
  issues.push(`${target.url}: returned ${target.status}${target.location ? ` -> ${target.location}` : ""}; linked from ${sources.join(", ")}`);
}

console.log(`Checked ${pages.length} public blog(s) and ${targets.length} unique internal target(s).`);
if (issues.length) {
  console.error(`\nInternal-link issues: ${issues.length}`);
  for (const issue of issues) console.error(issue);
  process.exitCode = 1;
} else {
  console.log("No placeholders, redirects, or broken internal article links found.");
}
