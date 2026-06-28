import { useEffect, useMemo, useRef, useState } from "react";
import { Link, useLocation, useParams } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { getCanonicalUrl, usePageSeo } from "../lib/seo";
import { usePrerenderData } from "../lib/prerenderData";
import { useTranslation } from "../lib/i18n";
import {
  clearCatalogProductReturnState,
  currentCatalogReturnPath,
  currentHistoryKey,
  readCatalogProductReturnState,
  writeCatalogProductReturnState
} from "../lib/productReturnState";
import {
  catalogAlternates,
  catalogBasePath,
  catalogCopy,
  catalogLocaleConfig,
  localizedField
} from "../lib/catalogLocale";
import { formatOfferPrice, getProductOfferPrice } from "../lib/productOffers";
import NotFound from "./NotFound";

const PAGE_SIZE = 24;
const facetOrder = ["color", "application", "finish", "form", "origin", "pattern", "availability"];

const localizedTermLabel = (term, lang) => localizedField(term, "label", lang) || term?.key || "";

function catalogApiPath(categorySlug, facet, value, search, lang, limit = PAGE_SIZE) {
  const route = facet && value
    ? `/api/catalog/categories/${categorySlug}/${facet}/${value}`
    : `/api/catalog/categories/${categorySlug}`;
  const params = new URLSearchParams(search || "");
  params.set("locale", lang);
  params.set("limit", String(limit));
  params.set("offset", "0");
  return `${route}?${params.toString()}`;
}

function selectedCount(selected) {
  return Object.values(selected || {}).reduce((total, values) => total + values.length, 0);
}

function buildFilterHref(basePath, categorySlug, selected, facetKey, valueKey, query) {
  const next = Object.fromEntries(Object.entries(selected || {}).map(([key, values]) => [key, [...values]]));
  const values = new Set(next[facetKey] || []);
  if (values.has(valueKey)) values.delete(valueKey);
  else values.add(valueKey);
  if (values.size) next[facetKey] = [...values];
  else delete next[facetKey];

  const total = selectedCount(next);
  if (total === 0 && !query) return `${basePath}/${categorySlug}`;
  if (total === 1 && !query) {
    const [singleFacet, singleValues] = Object.entries(next)[0];
    return `${basePath}/${categorySlug}/${singleFacet}/${singleValues[0]}`;
  }

  const params = new URLSearchParams();
  for (const key of facetOrder) {
    for (const selectedValue of next[key] || []) params.append(key, selectedValue);
  }
  if (query) params.set("q", query);
  return `${basePath}/${categorySlug}?${params.toString()}`;
}

function ProductCard({ product, lang, copy, returnPath, onRememberReturnState }) {
  const title = localizedField(product, "title", lang);
  const isRTL = lang === "fa" || lang === "ar";
  const gradientDirection = isRTL ? "bg-gradient-to-tl" : "bg-gradient-to-tr";
  const offerPrice = getProductOfferPrice(product);
  const legacyDetails = [product.variants, product.mines];
  const fallbackTaxonomies = ["use_case_form", "mines"];
  const detailGroups = legacyDetails
    .map((values, index) => values?.length
      ? values
      : (product.terms || [])
        .filter((term) => term.taxonomy === fallbackTaxonomies[index])
        .map((term) => localizedTermLabel(term, lang)))
    .filter((values) => values.length > 0)
    .slice(0, 2);
  return (
    <Link
      to={`/${lang}/products/${product.slug}`}
      state={{ catalogRouteKind: "product", productReturnTo: returnPath }}
      onClick={() => onRememberReturnState(product.slug)}
      className="group flex h-full flex-col overflow-hidden transition hover:-translate-y-1 hover:shadow-xl"
      style={{ contentVisibility: "auto", containIntrinsicSize: "360px" }}
    >
      <div className="relative aspect-square w-full overflow-hidden bg-primary/10">
        {product.is_popular ? (
          <span className="absolute left-3 top-3 z-10 rounded-full border border-white/90 bg-white/40 px-3 py-1 text-[11px] font-semibold text-white shadow-sm backdrop-blur">
            {copy.popularBadge}
          </span>
        ) : null}
        {product.image_url ? (
          <img src={resolveImageUrl(product.image_url)} alt={title} loading="lazy" decoding="async" className="h-full w-full object-cover transition duration-500 group-hover:scale-105" />
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-primary/50">{copy.noImage}</div>
        )}
        <div className={`pointer-events-none absolute inset-x-0 bottom-0 ${gradientDirection} from-black/70 via-black/25 to-transparent p-4 pt-16`}>
          {offerPrice > 0 ? (
            <p className="mb-2 inline-flex max-w-full min-w-0 items-center gap-1 rounded-full bg-white/15 px-3 py-1 text-[11px] font-bold text-white/95 backdrop-blur">
              <span>{copy.offerLabel}</span>
              <span aria-hidden="true">/</span>
              <span className="min-w-0 truncate">{formatOfferPrice(offerPrice, lang)}</span>
            </p>
          ) : null}
          <h2 className="font-display text-xl leading-tight text-white drop-shadow-[0_10px_24px_rgba(0,0,0,0.55)]">{title}</h2>
          {detailGroups.length > 0 ? (
            <div className="mt-2 space-y-1 text-[11px] font-semibold text-white/85 drop-shadow-[0_8px_18px_rgba(0,0,0,0.45)]">
              {detailGroups.map((values) => <p key={values.join("-")} className="truncate">{values.join(" • ")}</p>)}
            </div>
          ) : null}
        </div>
      </div>
    </Link>
  );
}

function FacetGroups({ page, query, lang, basePath, numberLocale }) {
  return (
    <div className="space-y-4">
      {page.facets.map((facet) => (
        <section key={facet.key} className="flex min-w-0 items-center gap-3" aria-labelledby={`facet-${facet.key}`}>
          <h2 id={`facet-${facet.key}`} className="w-20 shrink-0 text-xs font-semibold uppercase tracking-[0.18em] text-primary/60">{localizedField(facet, "label", lang)}</h2>
          <div className="flex min-w-0 flex-1 flex-nowrap gap-2 overflow-x-auto pb-1">
            {facet.values.map((facetValue) => {
              const selected = page.selected_filters?.[facet.key]?.includes(facetValue.key);
              return (
                <Link
                  key={facetValue.key}
                  to={buildFilterHref(basePath, page.category.slug, page.selected_filters, facet.key, facetValue.key, query)}
                  className={`inline-flex min-h-9 shrink-0 items-center gap-2 whitespace-nowrap rounded-full border px-4 py-2 text-xs font-semibold transition ${selected ? "border-primary bg-primary text-sand" : "border-primary/20 text-primary/70 hover:border-primary/50"}`}
                  aria-current={selected ? "page" : undefined}
                >
                  <span>{localizedTermLabel(facetValue, lang)}</span>
                  <span className={selected ? "text-white/65" : "text-primary/45"}>{facetValue.count.toLocaleString(numberLocale)}</span>
                </Link>
              );
            })}
          </div>
        </section>
      ))}
    </div>
  );
}

export default function ProductCatalog({ categorySlugOverride = "", initialPageOverride = null }) {
  const { lang } = useTranslation();
  const config = catalogLocaleConfig[lang] || catalogLocaleConfig.en;
  const copy = catalogCopy[lang] || catalogCopy.en;
  const basePath = catalogBasePath(lang);
  const { categorySlug: routeCategorySlug, facet, value } = useParams();
  const categorySlug = categorySlugOverride || routeCategorySlug;
  const location = useLocation();
  const returnPath = currentCatalogReturnPath(location);
  const restoreReturnState = location.state?.restoreProductReturn === true;
  const prerenderedCatalog = usePrerenderData("catalogPage");
  const prerendered = initialPageOverride || prerenderedCatalog;
  const pendingReturnStateRef = useRef(null);
  const matchesPrerender = prerendered?.locale === lang
    && prerendered?.category?.slug === categorySlug
    && (prerendered?.selected_facet_key || "") === (facet || "")
    && (prerendered?.selected_facet?.key || "") === (value || "");
  const [page, setPage] = useState(matchesPrerender ? prerendered : null);
  const [loading, setLoading] = useState(!matchesPrerender);
  const [loadingMore, setLoadingMore] = useState(false);
  const [notFound, setNotFound] = useState(false);
  const query = new URLSearchParams(location.search).get("q") || "";
  const suffix = `/${categorySlug}${facet && value ? `/${facet}/${value}` : ""}`;

  useEffect(() => {
    const storedReturnState = readCatalogProductReturnState(returnPath);
    const historyRestore = Boolean(storedReturnState?.historyKey && storedReturnState.historyKey === currentHistoryKey());
    const returnState = restoreReturnState || historyRestore ? storedReturnState : null;
    pendingReturnStateRef.current = returnState;
    const restoreLimit = returnState ? Math.max(PAGE_SIZE, returnState.productCount || PAGE_SIZE) : PAGE_SIZE;
    const initialMatches = matchesPrerender && !location.search && !returnState;
    if (initialMatches) {
      setPage(prerendered);
      setLoading(false);
      setNotFound(false);
      return undefined;
    }
    let active = true;
    setLoading(true);
    setNotFound(false);
    fetchJSON(catalogApiPath(categorySlug, facet, value, location.search, lang, restoreLimit))
      .then((response) => {
        if (active) setPage(response.data || null);
      })
      .catch((error) => {
        if (!active) return;
        if (error.status === 404) setNotFound(true);
        setPage(null);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [categorySlug, facet, lang, location.search, matchesPrerender, prerendered, restoreReturnState, returnPath, value]);

  useEffect(() => {
    const returnState = pendingReturnStateRef.current;
    if (loading || !page || !returnState) return undefined;
    const productCount = page.products?.length || 0;
    const wantedCount = Math.min(
      Math.max(PAGE_SIZE, returnState.productCount || PAGE_SIZE),
      page.pagination?.total || productCount
    );
    if (productCount < wantedCount) return undefined;

    const frameId = window.requestAnimationFrame(() => {
      window.scrollTo({ top: returnState.scrollY, behavior: "auto" });
      window.setTimeout(() => {
        window.scrollTo({ top: returnState.scrollY, behavior: "auto" });
      }, 120);
      pendingReturnStateRef.current = null;
      clearCatalogProductReturnState();
    });

    return () => window.cancelAnimationFrame(frameId);
  }, [loading, page]);

  const seo = page?.seo || {
    title: copy.products,
    description: copy.products,
    h1: copy.products,
    intro: "",
    canonical: `${basePath}/${categorySlug}${facet && value ? `/${facet}/${value}` : ""}`,
    robots: notFound ? "noindex,follow" : "index,follow"
  };
  const alternates = location.search ? [] : catalogAlternates(suffix);

  const jsonLd = useMemo(() => {
    if (!page) return null;
    const canonical = getCanonicalUrl(seo.canonical);
    const categoryTitle = localizedField(page.category, "title", lang);
    const breadcrumbs = [
      { "@type": "ListItem", position: 1, name: copy.home, item: getCanonicalUrl("/") },
      { "@type": "ListItem", position: 2, name: copy.products, item: getCanonicalUrl(basePath) },
      { "@type": "ListItem", position: 3, name: categoryTitle, item: getCanonicalUrl(`${basePath}/${page.category.slug}`) }
    ];
    if (page.selected_facet) {
      breadcrumbs.push({ "@type": "ListItem", position: 4, name: localizedTermLabel(page.selected_facet, lang), item: canonical });
    }
    return {
      "@context": "https://schema.org",
      "@graph": [
        { "@type": "CollectionPage", "@id": `${canonical}#webpage`, url: canonical, name: seo.title, description: seo.description, inLanguage: lang },
        { "@type": "BreadcrumbList", itemListElement: breadcrumbs },
        {
          "@type": "ItemList",
          itemListElement: page.products.map((product, index) => ({
            "@type": "ListItem", position: page.pagination.offset + index + 1,
            name: localizedField(product, "title", lang), url: getCanonicalUrl(`/${lang}/products/${product.slug}`)
          }))
        }
      ]
    };
  }, [basePath, copy.home, copy.products, lang, page, seo.canonical, seo.description, seo.title]);

  usePageSeo({
    title: seo.title,
    description: seo.description,
    path: seo.canonical,
    lang,
    locale: config.locale,
    image: page?.category?.preview_image || "",
    robots: location.search ? "noindex,follow" : seo.robots,
    alternates,
    jsonLd,
    jsonLdId: "catalog-page-jsonld"
  });

  const loadMore = async () => {
    if (!page || loadingMore) return;
    setLoadingMore(true);
    try {
      const url = new URL(catalogApiPath(categorySlug, facet, value, location.search, lang), "http://catalog.local");
      url.searchParams.set("offset", String(page.products.length));
      const response = await fetchJSON(`${url.pathname}?${url.searchParams.toString()}`);
      const next = response.data;
      setPage((current) => ({ ...next, products: [...(current?.products || []), ...(next?.products || [])] }));
    } finally {
      setLoadingMore(false);
    }
  };

  const rememberCatalogReturnState = (productSlug) => {
    if (typeof window === "undefined" || !page) return;
    writeCatalogProductReturnState({
      path: returnPath,
      scrollY: window.scrollY,
      productCount: page.products?.length || PAGE_SIZE,
      productSlug,
      lang
    });
  };

  if (loading && !page) return <div className="section-shell min-h-[55vh] py-20 text-center text-sm text-primary/60" dir={config.dir}>{copy.loading}</div>;
  if (notFound || !page) return <NotFound />;

  const hasMore = page.products.length < page.pagination.total;
  const selectedEntries = Object.entries(page.selected_filters || {});
  const categoryTitle = localizedField(page.category, "title", lang);

  return (
    <section className="section-shell pb-16 pt-10" dir={config.dir}>
      <nav className="text-sm text-primary/60" aria-label={copy.breadcrumb}>
        <Link to="/" className="hover:text-primary">{copy.home}</Link><span className="px-2">/</span>
        <Link to={basePath} className="hover:text-primary">{copy.products}</Link><span className="px-2">/</span>
        <span>{categoryTitle}</span>
      </nav>

      <header className="mt-8 max-w-4xl">
        <h1 className="font-display text-3xl leading-tight md:text-5xl">{seo.h1}</h1>
        <p className="mt-5 max-w-3xl text-sm leading-8 text-primary/70 md:text-base">{seo.intro}</p>
      </header>

      <div className="mt-8 border-y border-primary/10 py-6">
        <FacetGroups page={page} query={query} lang={lang} basePath={basePath} numberLocale={config.numberLocale} />
      </div>

      <form className="mt-5" action={`${basePath}/${categorySlug}`} method="get" role="search">
        <label className="sr-only" htmlFor="catalog-search">{copy.searchLabel}</label>
        <input id="catalog-search" name="q" defaultValue={query} className="w-full rounded-full border border-primary/20 bg-white/70 px-4 py-2.5 text-sm font-semibold text-primary outline-none transition focus:border-primary/60" placeholder={copy.searchPlaceholder} />
        <button type="submit" className="sr-only">{copy.search}</button>
      </form>

      <p className="mt-5 text-sm font-semibold text-primary/70">{page.pagination.total.toLocaleString(config.numberLocale)} {copy.productCount}</p>

      {selectedEntries.length > 0 || query ? (
        <div className="mt-5 flex flex-wrap items-center gap-2">
          {selectedEntries.flatMap(([facetKey, values]) => values.map((selectedValue) => {
            const facetGroup = page.facets.find((item) => item.key === facetKey);
            const facetValue = facetGroup?.values.find((item) => item.key === selectedValue);
            return <Link key={`${facetKey}-${selectedValue}`} to={buildFilterHref(basePath, categorySlug, page.selected_filters, facetKey, selectedValue, query)} className="rounded-full bg-primary px-4 py-2 text-xs font-semibold text-sand">{localizedTermLabel(facetValue, lang) || selectedValue} ×</Link>;
          }))}
          {query ? <Link to={`${basePath}/${categorySlug}`} className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold">{copy.removeSearch}</Link> : null}
          <Link to={`${basePath}/${categorySlug}`} className="text-xs font-semibold text-accent">{copy.clearAll}</Link>
        </div>
      ) : null}

      <div className="mt-8">
        {page.products.length === 0 ? <div className="py-20 text-center text-sm text-primary/60">{copy.emptyProducts}</div> : <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">{page.products.map((product) => <ProductCard key={product.id} product={product} lang={lang} copy={copy} returnPath={returnPath} onRememberReturnState={rememberCatalogReturnState} />)}</div>}
        {hasMore ? <div className="mt-9 text-center"><button type="button" disabled={loadingMore} onClick={loadMore} className="rounded-full border border-primary/25 px-6 py-3 text-sm font-semibold disabled:opacity-50">{loadingMore ? copy.loadingMore : copy.loadMore}</button></div> : null}
      </div>

      {page.related_projects?.length ? (
        <section className="mt-12 border-t border-primary/10 pt-10">
          <h2 className="font-display text-2xl">{copy.relatedProjects}</h2>
          <div className="mt-6 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {page.related_projects.map((project) => (
              <Link key={project.id} to={`/projects/${project.id}`} className="group block overflow-hidden rounded-lg bg-white/60 shadow-sm transition hover:-translate-y-1 hover:shadow-lg">
                <div className="aspect-[4/3] overflow-hidden bg-primary/10">
                  {project.cover_image_url ? <img src={project.cover_image_url} alt={`${copy.project} ${project.id}`} loading="lazy" className="h-full w-full object-cover transition duration-500 group-hover:scale-105" /> : null}
                </div>
                <p className="px-4 py-3 text-sm font-semibold">{copy.project} {project.id}</p>
              </Link>
            ))}
          </div>
        </section>
      ) : null}

    </section>
  );
}
