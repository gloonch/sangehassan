const sitemapUrl = process.argv[2] || "https://sangehassan.com/sitemap.xml";
const concurrency = Number(process.env.SEO_VERIFY_CONCURRENCY || 8);

function extractTag(html, pattern) {
  return html.match(pattern)?.[1]?.trim() || "";
}

function hasNoindex(html, headers) {
  const robotsTags = [...html.matchAll(/<meta\b[^>]*(?:name|property)=["']robots["'][^>]*>/gi)].map((match) => match[0]);
  const robotsNoindex = robotsTags.some((tag) => /\bnoindex\b/i.test(tag));
  const headerNoindex = /\bnoindex\b/i.test(headers.get("x-robots-tag") || "");
  return robotsNoindex || headerNoindex;
}

function parseSitemap(xml) {
  return [...xml.matchAll(/<loc>(.*?)<\/loc>/gi)].map((match) =>
    match[1].replace(/&amp;/g, "&").trim()
  );
}

async function fetchText(url) {
  const response = await fetch(url, { redirect: "follow" });
  const text = await response.text();
  return { response, text };
}

async function verifyUrl(url) {
  try {
    const { response, text } = await fetchText(url);
    const canonical = extractTag(text, /<link\b[^>]*rel=["']canonical["'][^>]*href=["']([^"']+)["'][^>]*>/i);
    const title = extractTag(text, /<title>([\s\S]*?)<\/title>/i);
    const description = extractTag(
      text,
      /<meta\b[^>]*name=["']description["'][^>]*content=["']([^"']*)["'][^>]*>/i
    );
    const h1Count = [...text.matchAll(/<h1\b/gi)].length;
    const jsonLdBlocks = [...text.matchAll(/<script\b[^>]*type=["']application\/ld\+json["'][^>]*>([\s\S]*?)<\/script>/gi)];
    const invalidJsonLd = jsonLdBlocks.some((match) => {
      try {
        JSON.parse(match[1]);
        return false;
      } catch (_) {
        return true;
      }
    });
    const alternates = [...text.matchAll(/<link\b[^>]*rel=["']alternate["'][^>]*hreflang=["']([^"']+)["'][^>]*href=["']([^"']+)["'][^>]*>/gi)]
      .map((match) => ({ lang: match[1], href: match[2] }));
    const localeMatch = new URL(url).pathname.match(/^\/(fa|en|ar)\//);
    const isPagination = /\/blogs\/page\/\d+$/.test(new URL(url).pathname);
    const missingSelfAlternate = Boolean(localeMatch && !isPagination && !alternates.some((item) => item.lang === localeMatch[1] && item.href === url));
    const missingDefaultAlternate = Boolean(localeMatch && !isPagination && !alternates.some((item) => item.lang === "x-default"));
    const missingPaginationPrev = isPagination && !/<link\b[^>]*rel=["']prev["']/i.test(text);
    const genericDescription = [
      "Detailed natural stone product page with images, specifications and project references.",
      "صفحه معرفی محصول سنگ طبیعی شامل تصاویر، مشخصات و اطلاعات کاربردی پروژه.",
      "صفحة منتج الحجر الطبيعي مع الصور والمواصفات ومعلومات الاستخدام في المشاريع.",
      "Detailed view of a completed SangeHassan project including gallery and project description."
    ].includes(description);

    return {
      url,
      status: response.status,
      finalUrl: response.url,
      noindex: hasNoindex(text, response.headers),
      canonical,
      canonicalMismatch: canonical !== url,
      missingTitle: !title,
      missingDescription: !description,
      title,
      description,
      h1Count,
      missingJsonLd: jsonLdBlocks.length === 0,
      invalidJsonLd,
      missingSelfAlternate,
      missingDefaultAlternate,
      missingPaginationPrev,
      genericDescription,
      containsNullByte: text.includes("\0")
    };
  } catch (error) {
    return {
      url,
      status: 0,
      error: error.message,
      noindex: false,
      canonical: "",
      canonicalMismatch: true,
      missingTitle: true,
      missingDescription: true,
      h1Count: 0,
      missingJsonLd: true,
      invalidJsonLd: true
    };
  }
}

async function mapLimit(items, limit, mapper) {
  const results = [];
  let index = 0;

  async function worker() {
    while (index < items.length) {
      const currentIndex = index;
      index += 1;
      results[currentIndex] = await mapper(items[currentIndex]);
    }
  }

  await Promise.all(Array.from({ length: Math.min(limit, items.length) }, worker));
  return results;
}

const { text: sitemapXml } = await fetchText(sitemapUrl);
const urls = parseSitemap(sitemapXml);
const results = await mapLimit(urls, concurrency, verifyUrl);

const non200 = results.filter((item) => item.status !== 200);
const noindex = results.filter((item) => item.noindex);
const canonicalMismatch = results.filter((item) => item.status === 200 && item.canonicalMismatch);
const missingMeta = results.filter((item) => item.status === 200 && (item.missingTitle || item.missingDescription));
const invalidH1 = results.filter((item) => item.status === 200 && item.h1Count !== 1);
const invalidStructuredData = results.filter((item) => item.status === 200 && (item.missingJsonLd || item.invalidJsonLd));
const invalidAlternates = results.filter((item) => item.status === 200 && (item.missingSelfAlternate || item.missingDefaultAlternate));
const invalidPagination = results.filter((item) => item.status === 200 && item.missingPaginationPrev);
const genericDescriptions = results.filter((item) => item.status === 200 && item.genericDescription);
const nullBytePages = results.filter((item) => item.status === 200 && item.containsNullByte);
const duplicateGroups = (field) => {
  const groups = new Map();
  for (const item of results.filter((result) => result.status === 200 && result[field])) {
    const key = item[field];
    groups.set(key, [...(groups.get(key) || []), item]);
  }
  return [...groups.entries()].filter(([, items]) => items.length > 1);
};
const duplicateTitles = duplicateGroups("title");
const duplicateDescriptions = duplicateGroups("description");

const printGroup = (title, items, formatter) => {
  console.log(`\n${title}: ${items.length}`);
  for (const item of items) {
    console.log(formatter(item));
  }
};

console.log(`Checked ${results.length} sitemap URL(s) from ${sitemapUrl}`);
printGroup("URLهای دارای noindex", noindex, (item) => item.url);
printGroup("URLهای canonical mismatch", canonicalMismatch, (item) => `${item.url} -> ${item.canonical || "(missing)"}`);
printGroup("URLهای status غیر 200", non200, (item) => `${item.status} ${item.url}${item.error ? ` (${item.error})` : ""}`);
printGroup(
  "URLهای بدون title/description",
  missingMeta,
  (item) => `${item.url} title=${!item.missingTitle} description=${!item.missingDescription}`
);
printGroup("URLهای با تعداد H1 نامعتبر", invalidH1, (item) => `${item.h1Count} ${item.url}`);
printGroup("URLهای با JSON-LD نامعتبر", invalidStructuredData, (item) => item.url);
printGroup("URLهای با hreflang ناقص", invalidAlternates, (item) => item.url);
printGroup("صفحات pagination بدون prev", invalidPagination, (item) => item.url);
printGroup("URLهای دارای description عمومی", genericDescriptions, (item) => item.url);
printGroup("URLهای دارای بایت NUL", nullBytePages, (item) => item.url);
console.log(`\nگروه‌های title تکراری: ${duplicateTitles.length}`);
console.log(`گروه‌های description تکراری: ${duplicateDescriptions.length}`);

const strictDuplicates = process.env.SEO_VERIFY_STRICT_DUPLICATES === "1";
if (noindex.length || canonicalMismatch.length || non200.length || missingMeta.length || invalidH1.length || invalidStructuredData.length || invalidAlternates.length || invalidPagination.length || genericDescriptions.length || nullBytePages.length || (strictDuplicates && (duplicateTitles.length || duplicateDescriptions.length))) {
  process.exitCode = 1;
}
