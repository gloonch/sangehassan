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

    return {
      url,
      status: response.status,
      finalUrl: response.url,
      noindex: hasNoindex(text, response.headers),
      canonical,
      canonicalMismatch: canonical !== url,
      missingTitle: !title,
      missingDescription: !description
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
      missingDescription: true
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

if (noindex.length || canonicalMismatch.length || non200.length || missingMeta.length) {
  process.exitCode = 1;
}
