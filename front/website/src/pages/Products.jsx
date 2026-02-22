import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const PRODUCTS_PAGE_SIZE = 20;

const getLocalized = (item, lang) => {
  if (!item) return "";
  if (lang === "fa") return item.title_fa;
  if (lang === "ar") return item.title_ar;
  return item.title_en;
};

const getTermIdentifier = (term) => {
  if (!term) return "";
  if (term.key) return String(term.key);
  if (term.id) return String(term.id);
  return "";
};

const getNumericPrice = (product) => {
  if (!product) return 0;
  if (typeof product.price === "number" && Number.isFinite(product.price)) {
    return product.price;
  }
  const parsed = Number(product.price);
  return Number.isFinite(parsed) ? parsed : 0;
};

const getFilterOptionLabel = (option, lang) => {
  if (!option) return "";
  if (lang === "fa") return option.label_fa || option.label_en || option.value;
  if (lang === "ar") return option.label_ar || option.label_en || option.value;
  return option.label_en || option.value;
};

const hasProductTerm = (product, taxonomy, termValue) => {
  if (!product || !taxonomy || !termValue) return false;
  const terms = Array.isArray(product.terms) ? product.terms : [];
  return terms.some((term) => term?.taxonomy === taxonomy && getTermIdentifier(term) === termValue);
};

export default function Products() {
  const { t, lang } = useTranslation();
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [activeCategory, setActiveCategory] = useState("all");
  const [searchInput, setSearchInput] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [searchPending, setSearchPending] = useState(false);
  const [stoneTypeFilter, setStoneTypeFilter] = useState("all");
  const [useCaseFilter, setUseCaseFilter] = useState("all");
  const [priceModeFilter, setPriceModeFilter] = useState("all");
  const [minPriceInput, setMinPriceInput] = useState("");
  const [maxPriceInput, setMaxPriceInput] = useState("");
  const [sortMode, setSortMode] = useState("default");
  const loadMoreTriggerRef = useRef(null);
  const [nextOffset, setNextOffset] = useState(0);
  const [hasMoreProducts, setHasMoreProducts] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [loading, setLoading] = useState(true);

  const fetchProductPage = useCallback(async (offset) => {
    const response = await fetchJSON(`/api/products?limit=${PRODUCTS_PAGE_SIZE}&offset=${offset}`);
    return Array.isArray(response.data) ? response.data : [];
  }, []);

  const applyFetchedPage = useCallback((incomingProducts, offset, replace = false) => {
    setProducts((previousProducts) => {
      if (replace) return incomingProducts;
      if (incomingProducts.length === 0) return previousProducts;
      const seenIDs = new Set(previousProducts.map((product) => product.id));
      const merged = [...previousProducts];
      for (const product of incomingProducts) {
        if (seenIDs.has(product.id)) continue;
        seenIDs.add(product.id);
        merged.push(product);
      }
      return merged;
    });
    setNextOffset(offset + incomingProducts.length);
    setHasMoreProducts(incomingProducts.length === PRODUCTS_PAGE_SIZE);
  }, []);

  useEffect(() => {
    let mounted = true;

    const loadInitial = async () => {
      setLoading(true);
      try {
        const [categoryRes, firstPageProducts] = await Promise.all([fetchJSON("/api/categories"), fetchProductPage(0)]);
        if (!mounted) return;
        setCategories(categoryRes.data || []);
        applyFetchedPage(firstPageProducts, 0, true);
      } catch (error) {
        if (!mounted) return;
        setProducts([]);
        setCategories([]);
        setNextOffset(0);
        setHasMoreProducts(false);
      } finally {
        if (mounted) setLoading(false);
      }
    };

    loadInitial();
    return () => {
      mounted = false;
    };
  }, [fetchProductPage, applyFetchedPage]);

  const loadMoreProducts = useCallback(async () => {
    if (loading || loadingMore || !hasMoreProducts) return;
    setLoadingMore(true);
    try {
      const incomingProducts = await fetchProductPage(nextOffset);
      applyFetchedPage(incomingProducts, nextOffset);
    } catch (_) {
      // keep current list on transient fetch errors
    } finally {
      setLoadingMore(false);
    }
  }, [loading, loadingMore, hasMoreProducts, fetchProductPage, nextOffset, applyFetchedPage]);

  useEffect(() => {
    const target = loadMoreTriggerRef.current;
    if (!target || !hasMoreProducts) return undefined;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          loadMoreProducts();
        }
      },
      { root: null, rootMargin: "320px 0px", threshold: 0 }
    );
    observer.observe(target);
    return () => observer.disconnect();
  }, [loadMoreProducts, hasMoreProducts]);

  useEffect(() => {
    const id = setTimeout(() => {
      setDebouncedSearch(searchInput.trim().toLowerCase());
      setSearchPending(false);
    }, 450);
    return () => clearTimeout(id);
  }, [searchInput]);

  const getLocalizedDescriptionHTML = (product) => {
    if (!product) return "";
    return (
      (lang === "fa"
        ? product.description_html_fa
        : lang === "ar"
          ? product.description_html_ar
          : product.description_html_en) ||
      product.description_html ||
      product.description ||
      ""
    );
  };

  const termFilterOptions = useMemo(() => {
    const stoneTypeMap = new Map();
    const useCaseMap = new Map();

    for (const product of products) {
      const terms = Array.isArray(product?.terms) ? product.terms : [];
      for (const term of terms) {
        const value = getTermIdentifier(term);
        if (!value) continue;
        const option = {
          value,
          label_en: term.label_en || term.key || value,
          label_fa: term.label_fa || term.label_en || term.key || value,
          label_ar: term.label_ar || term.label_en || term.key || value
        };
        if (term.taxonomy === "stone_type" && !stoneTypeMap.has(value)) {
          stoneTypeMap.set(value, option);
        }
        if (term.taxonomy === "use_case_application" && !useCaseMap.has(value)) {
          useCaseMap.set(value, option);
        }
      }
    }

    const sortByLocalizedLabel = (a, b) => {
      return getFilterOptionLabel(a, lang).localeCompare(getFilterOptionLabel(b, lang), lang === "fa" ? "fa" : lang === "ar" ? "ar" : "en", {
        sensitivity: "base"
      });
    };

    return {
      stoneTypes: [...stoneTypeMap.values()].sort(sortByLocalizedLabel),
      useCases: [...useCaseMap.values()].sort(sortByLocalizedLabel)
    };
  }, [products, lang]);

  useEffect(() => {
    if (stoneTypeFilter !== "all" && !termFilterOptions.stoneTypes.some((option) => option.value === stoneTypeFilter)) {
      setStoneTypeFilter("all");
    }
    if (useCaseFilter !== "all" && !termFilterOptions.useCases.some((option) => option.value === useCaseFilter)) {
      setUseCaseFilter("all");
    }
  }, [stoneTypeFilter, useCaseFilter, termFilterOptions]);

  const filtered = useMemo(() => {
    let base = products;

    if (activeCategory !== "all") {
      base = base.filter((product) => {
        if (product.category?.slug === activeCategory) return true;
        if (Array.isArray(product.categories)) {
          return product.categories.some((category) => category?.slug === activeCategory);
        }
        return false;
      });
    }

    if (debouncedSearch) {
      base = base.filter((product) => {
        const title = (getLocalized(product, lang) || product.title_en || "").toLowerCase();
        const slug = product.slug?.toLowerCase() || "";
        const descriptionHtmlRaw = getLocalizedDescriptionHTML(product);
        const description = (descriptionHtmlRaw || "").replace(/<[^>]+>/g, " ").toLowerCase();
        const categoryText = `${product.category?.title_en || ""} ${product.category?.title_fa || ""} ${product.category?.title_ar || ""}`.toLowerCase();
        const termsText = (Array.isArray(product.terms) ? product.terms : [])
          .map((term) => `${term?.label_en || ""} ${term?.label_fa || ""} ${term?.label_ar || ""} ${term?.key || ""}`)
          .join(" ")
          .toLowerCase();

        return (
          title.includes(debouncedSearch) ||
          slug.includes(debouncedSearch) ||
          description.includes(debouncedSearch) ||
          categoryText.includes(debouncedSearch) ||
          termsText.includes(debouncedSearch)
        );
      });
    }

    if (stoneTypeFilter !== "all") {
      base = base.filter((product) => hasProductTerm(product, "stone_type", stoneTypeFilter));
    }

    if (useCaseFilter !== "all") {
      base = base.filter((product) => hasProductTerm(product, "use_case_application", useCaseFilter));
    }

    if (priceModeFilter === "priced") {
      base = base.filter((product) => getNumericPrice(product) > 0);
    } else if (priceModeFilter === "unpriced") {
      base = base.filter((product) => getNumericPrice(product) <= 0);
    }

    const parsedMinPrice = Number(minPriceInput);
    const parsedMaxPrice = Number(maxPriceInput);
    const hasMinPrice = minPriceInput.trim() !== "" && Number.isFinite(parsedMinPrice) && parsedMinPrice >= 0;
    const hasMaxPrice = maxPriceInput.trim() !== "" && Number.isFinite(parsedMaxPrice) && parsedMaxPrice >= 0;

    if (hasMinPrice) {
      base = base.filter((product) => getNumericPrice(product) >= parsedMinPrice);
    }
    if (hasMaxPrice) {
      base = base.filter((product) => getNumericPrice(product) <= parsedMaxPrice);
    }

    const sorted = [...base];
    if (sortMode === "price_asc") {
      sorted.sort((a, b) => {
        const priceA = getNumericPrice(a);
        const priceB = getNumericPrice(b);
        const emptyA = priceA <= 0;
        const emptyB = priceB <= 0;
        if (emptyA && emptyB) return 0;
        if (emptyA) return 1;
        if (emptyB) return -1;
        return priceA - priceB;
      });
    } else if (sortMode === "price_desc") {
      sorted.sort((a, b) => {
        const priceA = getNumericPrice(a);
        const priceB = getNumericPrice(b);
        const emptyA = priceA <= 0;
        const emptyB = priceB <= 0;
        if (emptyA && emptyB) return 0;
        if (emptyA) return 1;
        if (emptyB) return -1;
        return priceB - priceA;
      });
    }

    return sorted;
  }, [
    activeCategory,
    products,
    debouncedSearch,
    lang,
    stoneTypeFilter,
    useCaseFilter,
    priceModeFilter,
    minPriceInput,
    maxPriceInput,
    sortMode
  ]);

  const mainCategories = useMemo(() => {
    const topLevel = categories.filter((category) => !category.parent_id);
    if (topLevel.length > 0) return topLevel;
    return categories;
  }, [categories]);

  const handleSearchChange = (event) => {
    const value = event.target.value;
    setSearchInput(value);
    setSearchPending(true);
  };

  const hasAdvancedFilters =
    stoneTypeFilter !== "all" ||
    useCaseFilter !== "all" ||
    priceModeFilter !== "all" ||
    minPriceInput.trim() !== "" ||
    maxPriceInput.trim() !== "" ||
    sortMode !== "default";

  const resetAdvancedFilters = () => {
    setStoneTypeFilter("all");
    setUseCaseFilter("all");
    setPriceModeFilter("all");
    setMinPriceInput("");
    setMaxPriceInput("");
    setSortMode("default");
  };

  return (
    <section className="section-shell py-16">
      <div className="mb-8 flex flex-col gap-4">
        <p className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("products.title")}</p>
        <h1 className="font-display text-3xl md:text-4xl">{t("products.subtitle")}</h1>
      </div>

      <div className="mb-8 space-y-3">
        <div className="flex flex-wrap items-center gap-3">
          <button
            type="button"
            onClick={() => setActiveCategory("all")}
            className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${
              activeCategory === "all"
                ? "border-primary bg-primary text-sand"
                : "border-primary/20 text-primary/70 hover:border-primary/50"
            }`}
          >
            {t("products.filterAll")}
          </button>
          {mainCategories.map((category) => (
            <button
              key={category.id}
              type="button"
              onClick={() => setActiveCategory(category.slug)}
              className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${
                activeCategory === category.slug
                  ? "border-primary bg-primary text-sand"
                  : "border-primary/20 text-primary/70 hover:border-primary/50"
              }`}
            >
              {getLocalized(category, lang)}
            </button>
          ))}
        </div>

        <label className="sr-only" htmlFor="product-search">
          {t("products.searchLabel")}
        </label>
        <div className="relative">
          <input
            id="product-search"
            type="search"
            value={searchInput}
            onChange={handleSearchChange}
            placeholder={t("products.searchPlaceholder")}
            className="w-full rounded-full border border-primary/20 bg-white/70 px-4 py-2.5 pr-9 text-sm font-semibold text-primary outline-none transition focus:border-primary/60"
          />
          {searchPending && (
            <span
              aria-label={t("messages.loading")}
              className="absolute right-3 top-1/2 h-3 w-3 -translate-y-1/2 animate-spin rounded-full border border-primary/40 border-t-transparent"
            />
          )}
        </div>

        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7">
          <div className="relative">
            <label className="sr-only" htmlFor="products-filter-stone-type">
              {t("products.filterStoneType")}
            </label>
            <select
              id="products-filter-stone-type"
              value={stoneTypeFilter}
              onChange={(event) => setStoneTypeFilter(event.target.value)}
              className="h-10 w-full appearance-none rounded-full border border-primary/20 bg-white/75 pl-4 pr-10 text-xs font-semibold text-primary/80 outline-none transition focus:border-primary/60"
            >
              <option value="all">{t("products.filterAny")}</option>
              {termFilterOptions.stoneTypes.map((option) => (
                <option key={`stone-type-filter-${option.value}`} value={option.value}>
                  {getFilterOptionLabel(option, lang)}
                </option>
              ))}
            </select>
            <span className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-primary/60">▾</span>
          </div>

          <div className="relative">
            <label className="sr-only" htmlFor="products-filter-use-case">
              {t("products.filterUseCase")}
            </label>
            <select
              id="products-filter-use-case"
              value={useCaseFilter}
              onChange={(event) => setUseCaseFilter(event.target.value)}
              className="h-10 w-full appearance-none rounded-full border border-primary/20 bg-white/75 pl-4 pr-10 text-xs font-semibold text-primary/80 outline-none transition focus:border-primary/60"
            >
              <option value="all">{t("products.filterAny")}</option>
              {termFilterOptions.useCases.map((option) => (
                <option key={`use-case-filter-${option.value}`} value={option.value}>
                  {getFilterOptionLabel(option, lang)}
                </option>
              ))}
            </select>
            <span className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-primary/60">▾</span>
          </div>

          <div className="relative">
            <label className="sr-only" htmlFor="products-filter-price-mode">
              {t("products.filterPriceMode")}
            </label>
            <select
              id="products-filter-price-mode"
              value={priceModeFilter}
              onChange={(event) => setPriceModeFilter(event.target.value)}
              className="h-10 w-full appearance-none rounded-full border border-primary/20 bg-white/75 pl-4 pr-10 text-xs font-semibold text-primary/80 outline-none transition focus:border-primary/60"
            >
              <option value="all">{t("products.priceModeAll")}</option>
              <option value="priced">{t("products.priceModeWith")}</option>
              <option value="unpriced">{t("products.priceModeWithout")}</option>
            </select>
            <span className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-primary/60">▾</span>
          </div>

          <div className="relative">
            <label className="sr-only" htmlFor="products-filter-sort">
              {t("products.filterSort")}
            </label>
            <select
              id="products-filter-sort"
              value={sortMode}
              onChange={(event) => setSortMode(event.target.value)}
              className="h-10 w-full appearance-none rounded-full border border-primary/20 bg-white/75 pl-4 pr-10 text-xs font-semibold text-primary/80 outline-none transition focus:border-primary/60"
            >
              <option value="default">{t("products.sortDefault")}</option>
              <option value="price_asc">{t("products.sortPriceLowToHigh")}</option>
              <option value="price_desc">{t("products.sortPriceHighToLow")}</option>
            </select>
            <span className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-primary/60">▾</span>
          </div>

          <label className="sr-only" htmlFor="products-filter-min-price">
            {t("products.filterMinPrice")}
          </label>
          <input
            id="products-filter-min-price"
            type="number"
            min="0"
            value={minPriceInput}
            onChange={(event) => setMinPriceInput(event.target.value)}
            placeholder={t("products.filterMinPrice")}
            className="h-10 w-full rounded-full border border-primary/20 bg-white/75 px-4 text-xs font-semibold text-primary/80 outline-none transition focus:border-primary/60"
          />

          <label className="sr-only" htmlFor="products-filter-max-price">
            {t("products.filterMaxPrice")}
          </label>
          <input
            id="products-filter-max-price"
            type="number"
            min="0"
            value={maxPriceInput}
            onChange={(event) => setMaxPriceInput(event.target.value)}
            placeholder={t("products.filterMaxPrice")}
            className="h-10 w-full rounded-full border border-primary/20 bg-white/75 px-4 text-xs font-semibold text-primary/80 outline-none transition focus:border-primary/60"
          />

          <button
            type="button"
            onClick={resetAdvancedFilters}
            disabled={!hasAdvancedFilters}
            className={`h-10 w-full rounded-full border px-4 text-xs font-semibold transition ${
              hasAdvancedFilters
                ? "border-primary/30 text-primary hover:border-primary/60"
                : "cursor-not-allowed border-primary/10 text-primary/40"
            }`}
          >
            {t("products.filtersReset")}
          </button>
        </div>
      </div>

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : filtered.length === 0 ? (
        <p className="text-sm text-primary/70">{t("products.empty")}</p>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filtered.map((product) => {
            return (
              <Link
                key={product.id}
                to={`/products/${product.slug}`}
                className="glass-panel group flex h-full flex-col overflow-hidden rounded-2xl transition hover:-translate-y-1 hover:shadow-2xl"
              >
                <div className="relative h-48 w-full overflow-hidden bg-primary/10">
                  {product.image_url ? (
                    <img
                      src={resolveImageUrl(product.image_url)}
                      alt={getLocalized(product, lang) || product.title_en}
                      className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
                    />
                  ) : (
                    <div className="flex h-full items-center justify-center text-sm text-primary/60">{t("productDetail.noImages")}</div>
                  )}

                  <div className="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-white/95 via-white/70 to-transparent p-4">
                    <p className="text-xs font-semibold uppercase tracking-[0.2em] text-primary/70">
                      {product.category ? getLocalized(product.category, lang) : t("products.categoryLabel")}
                    </p>
                    <h3 className="mt-1 font-display text-xl leading-tight text-primary">
                      {getLocalized(product, lang) || product.title_en}
                    </h3>
                  </div>
                </div>

                <div className="flex flex-1 flex-col p-5">
                  <div className="mt-auto pt-4 text-sm font-semibold text-accent">
                    {t("products.priceLabel")}: {product.price ? product.price : t("messages.empty")}
                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      )}

      {!loading && (
        <div className="mt-8">
          {loadingMore && (
            <p className="text-center text-sm text-primary/70">{t("messages.loading")}</p>
          )}
          {hasMoreProducts && <div ref={loadMoreTriggerRef} className="h-8 w-full" />}
        </div>
      )}
    </section>
  );
}
