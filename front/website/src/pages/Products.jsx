import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { gsap } from "gsap";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { getCanonicalUrl } from "../lib/seo";

const PRODUCTS_PAGE_SIZE = 20;
const PRODUCT_GRID_SKELETON_COUNT = 6;
const CATEGORY_SKELETON_COUNT = 6;
const PRODUCTS_RETURN_STATE_KEY = "sh_products_return_state";
const PRODUCTS_RETURN_STATE_MAX_AGE_MS = 30 * 60 * 1000;
const EMPTY_STATE_DELAY_MS = 2000;

const productsSeoContent = {
  fa: {
    title: "محصولات سنگ طبیعی | سنگ حسن",
    description:
      "کاتالوگ محصولات سنگ حسن شامل اسلب، تایل و سنگ‌های فرآوری‌شده برای پروژه‌های ساختمانی، همکاری B2B و صادرات.",
    locale: "fa_IR"
  },
  en: {
    title: "Natural Stone Products | SangeHassan",
    description:
      "Browse SangeHassan natural stone products, including slabs, tiles, and finished stones for building projects, B2B supply, and export.",
    locale: "en_US"
  },
  ar: {
    title: "منتجات الحجر الطبيعي | سانج حسن",
    description:
      "تصفح منتجات الحجر الطبيعي من سانج حسن، بما في ذلك الألواح والبلاط والأحجار المعالجة للمشاريع والتصدير.",
    locale: "ar_SA"
  }
};

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

const normalizeCategory = (category) => {
  if (!category) return category;
  const slug = normalizeChiniSlug(category.slug, category.title_fa);
  if (slug !== category.slug || category.title_fa === "چینی") {
    return { ...category, title_fa: "چینی کریستال", slug };
  }
  return category;
};

const normalizeChiniSlug = (slug, titleFa) => {
  // Normalize all older/short slugs to the proper spelling.
  if (slug === "chinese-crystal") return "chinese-crystal";
  if (slug === "chini-crystal" || slug === "chini" || slug === "chyny-krystal") return "chinese-crystal";
  if (titleFa === "چینی" || titleFa === "چینی کریستال") return "chinese-crystal";
  return slug;
};

const normalizeProductForList = (product) => ({
  ...product,
  category: normalizeCategory(product.category),
  categories: Array.isArray(product.categories) ? product.categories.map(normalizeCategory) : product.categories
});

const mergeProductsByID = (previousProducts, incomingProducts) => {
  if (!incomingProducts.length) return previousProducts;
  const seenIDs = new Set(previousProducts.map((product) => product.id));
  const merged = [...previousProducts];
  for (const product of incomingProducts) {
    if (seenIDs.has(product.id)) continue;
    seenIDs.add(product.id);
    merged.push(product);
  }
  return merged;
};

const hasProductTerm = (product, taxonomy, termValue) => {
  if (!product || !taxonomy || !termValue) return false;
  const terms = Array.isArray(product.terms) ? product.terms : [];
  return terms.some((term) => term?.taxonomy === taxonomy && getTermIdentifier(term) === termValue);
};

const normalizeSearchText = (value) =>
  String(value || "")
    .normalize("NFKC")
    .replace(/\u200c/g, " ")
    .trim()
    .toLowerCase()
    .replace(/\s+/g, " ");

const parseProductsQuery = (search) => {
  const params = new URLSearchParams(search || "");
  const category = normalizeChiniSlug(params.get("category") || "", "");
  const keyword = String(params.get("q") || "").trim();
  return { category, keyword };
};

const readReturnState = () => {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.sessionStorage.getItem(PRODUCTS_RETURN_STATE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    const ts = Number(parsed?.ts || 0);
    if (!ts || Date.now() - ts > PRODUCTS_RETURN_STATE_MAX_AGE_MS) {
      window.sessionStorage.removeItem(PRODUCTS_RETURN_STATE_KEY);
      return null;
    }
    const activeCategory = typeof parsed?.activeCategory === "string" ? parsed.activeCategory : "";
    const scrollY = Number.isFinite(parsed?.scrollY) ? Math.max(0, parsed.scrollY) : 0;
    const nextOffset = Number.isFinite(parsed?.nextOffset) ? Math.max(0, parsed.nextOffset) : PRODUCTS_PAGE_SIZE;
    return { activeCategory, scrollY, nextOffset };
  } catch (_) {
    return null;
  }
};

const CategorySkeletons = () => (
  <>
    {Array.from({ length: CATEGORY_SKELETON_COUNT }).map((_, index) => (
      <span
        key={`category-skeleton-${index}`}
        className="h-9 w-24 animate-pulse rounded-full border border-primary/10 bg-primary/5"
      />
    ))}
  </>
);

const ProductGridSkeleton = () => (
  <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3" aria-hidden="true">
    {Array.from({ length: PRODUCT_GRID_SKELETON_COUNT }).map((_, index) => (
      <div key={`product-skeleton-${index}`} className="overflow-hidden bg-white/45">
        <div className="aspect-square w-full animate-pulse bg-primary/10" />
      </div>
    ))}
  </div>
);

export default function Products() {
  const { t, lang } = useTranslation();
  const location = useLocation();
  const pageRef = useRef(null);
  const hasEntranceAnimatedRef = useRef(false);
  const hasLoadedInitialProductsRef = useRef(false);
  const pendingRestoreRef = useRef(null);
  const scrollRestoreRef = useRef(null);
  const viewLoadRequestRef = useRef(0);
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [categoriesLoading, setCategoriesLoading] = useState(true);
  const [activeCategory, setActiveCategory] = useState("");
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
  const [loading, setLoading] = useState(false);
  const [emptyStateVisible, setEmptyStateVisible] = useState(false);

  useEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return;

    const seo = productsSeoContent[lang] || productsSeoContent.fa;
    const pageUrl = getCanonicalUrl("/products");
    const previousTitle = document.title;
    const previousDescription = document.head.querySelector('meta[name="description"]')?.getAttribute("content") ?? null;
    const previousOgTitle = document.head.querySelector('meta[property="og:title"]')?.getAttribute("content") ?? null;
    const previousOgDescription =
      document.head.querySelector('meta[property="og:description"]')?.getAttribute("content") ?? null;

    const upsertMeta = (selector, attrs, value) => {
      let el = document.head.querySelector(selector);
      if (!el) {
        el = document.createElement("meta");
        Object.entries(attrs).forEach(([key, attrValue]) => el.setAttribute(key, attrValue));
        document.head.appendChild(el);
      }
      el.setAttribute("content", value);
    };

    let canonical = document.head.querySelector('link[rel="canonical"]');
    const createdCanonical = !canonical;
    const previousCanonical = canonical?.getAttribute("href") ?? null;
    if (!canonical) {
      canonical = document.createElement("link");
      canonical.setAttribute("rel", "canonical");
      document.head.appendChild(canonical);
    }

    document.title = seo.title;
    canonical.setAttribute("href", pageUrl);
    upsertMeta('meta[name="description"]', { name: "description" }, seo.description);
    upsertMeta('meta[property="og:title"]', { property: "og:title" }, seo.title);
    upsertMeta('meta[property="og:description"]', { property: "og:description" }, seo.description);
    upsertMeta('meta[property="og:url"]', { property: "og:url" }, pageUrl);
    upsertMeta('meta[property="og:locale"]', { property: "og:locale" }, seo.locale);
    upsertMeta('meta[name="twitter:title"]', { name: "twitter:title" }, seo.title);
    upsertMeta('meta[name="twitter:description"]', { name: "twitter:description" }, seo.description);

    return () => {
      document.title = previousTitle;
      const description = document.head.querySelector('meta[name="description"]');
      const ogTitle = document.head.querySelector('meta[property="og:title"]');
      const ogDescription = document.head.querySelector('meta[property="og:description"]');

      if (previousDescription === null) description?.removeAttribute("content");
      else description?.setAttribute("content", previousDescription);
      if (previousOgTitle === null) ogTitle?.removeAttribute("content");
      else ogTitle?.setAttribute("content", previousOgTitle);
      if (previousOgDescription === null) ogDescription?.removeAttribute("content");
      else ogDescription?.setAttribute("content", previousOgDescription);

      if (createdCanonical) canonical?.remove();
      else if (previousCanonical === null) canonical?.removeAttribute("href");
      else canonical?.setAttribute("href", previousCanonical);
    };
  }, [lang]);

  const normalizedProducts = useMemo(
    () => products.map(normalizeProductForList),
    [products]
  );

  const normalizedCategories = useMemo(() => {
    const map = new Map();
    for (const cat of categories) {
      const norm = normalizeCategory(cat);
      if (!norm?.slug) continue;
      if (!map.has(norm.slug)) map.set(norm.slug, norm);
    }
    return Array.from(map.values());
  }, [categories]);

  const normalizedActiveCategory =
    activeCategory === "chini" || activeCategory === "chini-crystal" || activeCategory === "chyny-krystal"
      ? "chinese-crystal"
      : activeCategory;
  const hasSelectedCategory = normalizedActiveCategory !== "";
  const queryPreset = useMemo(() => parseProductsQuery(location.search), [location.search]);
  const hasQueryPreset = Boolean(queryPreset.category || queryPreset.keyword);

  useEffect(() => {
    if (hasQueryPreset) {
      pendingRestoreRef.current = null;
      scrollRestoreRef.current = null;
      if (queryPreset.category) {
        setActiveCategory(queryPreset.category);
      } else if (queryPreset.keyword) {
        setActiveCategory("all");
      }
      if (queryPreset.keyword) {
        setSearchInput(queryPreset.keyword);
        setDebouncedSearch(normalizeSearchText(queryPreset.keyword));
        setSearchPending(false);
      }
      if (typeof window !== "undefined") {
        window.sessionStorage.removeItem(PRODUCTS_RETURN_STATE_KEY);
      }
      return;
    }

    const restore = readReturnState();
    if (!restore) return;
    pendingRestoreRef.current = restore;
    if (restore.activeCategory) {
      setActiveCategory(restore.activeCategory);
    }
  }, [hasQueryPreset, queryPreset.category, queryPreset.keyword]);

  const fetchProductPage = useCallback(async (offset) => {
    const response = await fetchJSON(`/api/products?limit=${PRODUCTS_PAGE_SIZE}&offset=${offset}`);
    return Array.isArray(response.data) ? response.data : [];
  }, []);

  const applyFetchedPage = useCallback((incomingProducts, offset, replace = false) => {
    setProducts((previousProducts) => {
      if (replace) return incomingProducts;
      if (incomingProducts.length === 0) return previousProducts;
      return mergeProductsByID(previousProducts, incomingProducts);
    });
    setNextOffset(offset + incomingProducts.length);
    setHasMoreProducts(incomingProducts.length === PRODUCTS_PAGE_SIZE);
  }, []);

  useEffect(() => {
    let mounted = true;

    const loadCategories = async () => {
      setCategoriesLoading(true);
      try {
        const categoryRes = await fetchJSON("/api/categories");
        if (!mounted) return;
        setCategories(categoryRes.data || []);
      } catch (_) {
        if (!mounted) return;
        setCategories([]);
      } finally {
        if (mounted) setCategoriesLoading(false);
      }
    };

    loadCategories();
    return () => {
      mounted = false;
    };
  }, []);

  useEffect(() => {
    if (loading || !hasSelectedCategory) return;
    if (typeof window === "undefined") return;
    const targetY = scrollRestoreRef.current;
    if (targetY === null || targetY === undefined) return;

    const raf = window.requestAnimationFrame(() => {
      window.scrollTo({ top: targetY, behavior: "auto" });
      window.setTimeout(() => {
        window.scrollTo({ top: targetY, behavior: "auto" });
      }, 120);
      scrollRestoreRef.current = null;
      window.sessionStorage.removeItem(PRODUCTS_RETURN_STATE_KEY);
    });
    return () => window.cancelAnimationFrame(raf);
  }, [loading, hasSelectedCategory, products.length]);

  const loadMoreProducts = useCallback(async () => {
    if (!hasSelectedCategory || loading || loadingMore || !hasMoreProducts) return;
    setLoadingMore(true);
    try {
      const incomingProducts = await fetchProductPage(nextOffset);
      applyFetchedPage(incomingProducts, nextOffset);
    } catch (_) {
      // keep current list on transient fetch errors
    } finally {
      setLoadingMore(false);
    }
  }, [hasSelectedCategory, loading, loadingMore, hasMoreProducts, fetchProductPage, nextOffset, applyFetchedPage]);

  useEffect(() => {
    const target = loadMoreTriggerRef.current;
    if (!target || !hasSelectedCategory || !hasMoreProducts) return undefined;
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
  }, [loadMoreProducts, hasSelectedCategory, hasMoreProducts]);

  useEffect(() => {
    const id = setTimeout(() => {
      setDebouncedSearch(normalizeSearchText(searchInput));
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

  const filterProductsForView = useCallback((items) => {
    if (!hasSelectedCategory) return [];

    let base = items;

    if (normalizedActiveCategory !== "all") {
      base = base.filter((product) => {
        if (normalizeChiniSlug(product.category?.slug, product.category?.title_fa) === normalizedActiveCategory) return true;
        if (Array.isArray(product.categories)) {
          return product.categories.some(
            (category) => normalizeChiniSlug(category?.slug, category?.title_fa) === normalizedActiveCategory
          );
        }
        return false;
      });
    }

    if (debouncedSearch) {
      base = base.filter((product) => {
        const title = normalizeSearchText(getLocalized(product, lang) || product.title_en || "");
        const slug = normalizeSearchText(product.slug || "");
        const descriptionHtmlRaw = getLocalizedDescriptionHTML(product);
        const description = normalizeSearchText((descriptionHtmlRaw || "").replace(/<[^>]+>/g, " "));
        const categoryText = normalizeSearchText(
          `${product.category?.title_en || ""} ${product.category?.title_fa || ""} ${product.category?.title_ar || ""}`
        );
        const termsText = normalizeSearchText(
          (Array.isArray(product.terms) ? product.terms : [])
            .map((term) => `${term?.label_en || ""} ${term?.label_fa || ""} ${term?.label_ar || ""} ${term?.key || ""}`)
            .join(" ")
        );
        const metaText = normalizeSearchText(
          [
            ...(Array.isArray(product.finishes) ? product.finishes : []),
            ...(Array.isArray(product.mines) ? product.mines : []),
            ...(Array.isArray(product.variants) ? product.variants : []),
            ...(Array.isArray(product.aliases) ? product.aliases : []),
            product.quarry || "",
            product.stone_type || "",
            product.stoneType || ""
          ].join(" ")
        );

        return (
          title.includes(debouncedSearch) ||
          slug.includes(debouncedSearch) ||
          description.includes(debouncedSearch) ||
          categoryText.includes(debouncedSearch) ||
          termsText.includes(debouncedSearch) ||
          metaText.includes(debouncedSearch)
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
    if (sortMode === "default") {
      sorted.sort((a, b) => Number(Boolean(b.is_popular)) - Number(Boolean(a.is_popular)));
    } else if (sortMode === "price_asc") {
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
    normalizedActiveCategory,
    hasSelectedCategory,
    debouncedSearch,
    lang,
    stoneTypeFilter,
    useCaseFilter,
    priceModeFilter,
    minPriceInput,
    maxPriceInput,
    sortMode
  ]);

  const filtered = useMemo(() => filterProductsForView(normalizedProducts), [filterProductsForView, normalizedProducts]);

  useEffect(() => {
    if (!hasSelectedCategory || searchPending) return undefined;

    const requestID = viewLoadRequestRef.current + 1;
    viewLoadRequestRef.current = requestID;
    let cancelled = false;

    const restoreState = pendingRestoreRef.current;
    const restoreTargetOffset = restoreState ? Math.max(PRODUCTS_PAGE_SIZE, restoreState.nextOffset) : 0;
    const needsInitialPage = !hasLoadedInitialProductsRef.current;
    const needsVisibleResult = filtered.length === 0 && hasMoreProducts;
    const needsRestorePages = Boolean(restoreState && nextOffset < restoreTargetOffset && hasMoreProducts);

    if (!needsInitialPage && !needsVisibleResult && !needsRestorePages) {
      setLoading(false);
      return () => {
        cancelled = true;
      };
    }

    const settleCurrentView = async () => {
      setLoading(true);
      setEmptyStateVisible(false);

      let localProducts = products;
      let localOffset = hasLoadedInitialProductsRef.current ? nextOffset : 0;
      let localHasMore = hasLoadedInitialProductsRef.current ? hasMoreProducts : true;

      if (!hasLoadedInitialProductsRef.current) {
        localProducts = [];
      }

      try {
        while (!cancelled && viewLoadRequestRef.current === requestID) {
          const normalizedLocalProducts = localProducts.map(normalizeProductForList);
          const visibleCount = filterProductsForView(normalizedLocalProducts).length;
          const shouldFetchFirstPage = !hasLoadedInitialProductsRef.current && localOffset < PRODUCTS_PAGE_SIZE;
          const shouldFetchRestorePages = Boolean(restoreState && localOffset < restoreTargetOffset && localHasMore);
          const shouldFetchUntilFirstMatch = visibleCount === 0 && localHasMore;

          if (!shouldFetchFirstPage && !shouldFetchRestorePages && !shouldFetchUntilFirstMatch) {
            break;
          }

          const page = await fetchProductPage(localOffset);
          if (cancelled || viewLoadRequestRef.current !== requestID) return;

          localProducts = mergeProductsByID(localProducts, page);
          localOffset += page.length;
          localHasMore = page.length === PRODUCTS_PAGE_SIZE;
          hasLoadedInitialProductsRef.current = true;

          setProducts(localProducts);
          setNextOffset(localOffset);
          setHasMoreProducts(localHasMore);

          if (page.length === 0) {
            break;
          }

          const nextVisibleCount = filterProductsForView(localProducts.map(normalizeProductForList)).length;
          if (!restoreState && nextVisibleCount > 0) {
            break;
          }
        }

        if (!cancelled && viewLoadRequestRef.current === requestID) {
          if (restoreState) {
            scrollRestoreRef.current = restoreState.scrollY ?? null;
            pendingRestoreRef.current = null;
          }
          setLoading(false);
        }
      } catch (_) {
        if (!cancelled && viewLoadRequestRef.current === requestID) {
          hasLoadedInitialProductsRef.current = true;
          setHasMoreProducts(false);
          setLoading(false);
        }
      }
    };

    settleCurrentView();
    return () => {
      cancelled = true;
    };
  }, [
    hasSelectedCategory,
    searchPending,
    filterProductsForView,
    fetchProductPage
  ]);

  useEffect(() => {
    if (!hasSelectedCategory || loading || hasMoreProducts || filtered.length > 0) {
      setEmptyStateVisible(false);
      return;
    }

    const timer = setTimeout(() => {
      setEmptyStateVisible(true);
    }, EMPTY_STATE_DELAY_MS);

    return () => clearTimeout(timer);
  }, [hasSelectedCategory, loading, hasMoreProducts, filtered.length]);

  const mainCategories = useMemo(() => {
    const topLevel = normalizedCategories.filter((category) => !category.parent_id);
    if (topLevel.length > 0) return topLevel;
    return normalizedCategories;
  }, [normalizedCategories]);

  const handleCategorySelect = (categorySlug) => {
    const nextCategory = normalizeChiniSlug(categorySlug, "");
    pendingRestoreRef.current = null;
    scrollRestoreRef.current = null;
    setEmptyStateVisible(false);
    if (normalizedActiveCategory !== nextCategory) {
      setLoading(true);
    }
    if (typeof window !== "undefined") {
      try {
        window.sessionStorage.removeItem(PRODUCTS_RETURN_STATE_KEY);
      } catch (_) {
        // ignore transient storage failures
      }
    }
    setActiveCategory(nextCategory);
  };

  const handleSearchChange = (event) => {
    const value = event.target.value;
    setSearchInput(value);
    setSearchPending(true);
    setEmptyStateVisible(false);
    if (hasSelectedCategory) {
      setLoading(true);
    }
  };
  const showSearchLoading = searchPending || (hasSelectedCategory && (loading || loadingMore));

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

  const rememberListState = useCallback(() => {
    if (typeof window === "undefined") return;
    try {
      window.sessionStorage.setItem(
        PRODUCTS_RETURN_STATE_KEY,
        JSON.stringify({
          activeCategory: normalizedActiveCategory,
          scrollY: window.scrollY,
          nextOffset,
          ts: Date.now()
        })
      );
    } catch (_) {
      // ignore transient storage write failures
    }
  }, [normalizedActiveCategory, nextOffset]);

  useEffect(() => {
    if (loading) return;
    const page = pageRef.current;
    if (!page || typeof window === "undefined" || !window.matchMedia) return;
    if (hasEntranceAnimatedRef.current) return;
    hasEntranceAnimatedRef.current = true;

    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const leadItems = page.querySelectorAll("[data-products-anim='lead']");
    const cardItems = Array.from(page.querySelectorAll("[data-products-anim='card']"));
    const cardsToAnimate = cardItems.slice(0, reduceMotion ? 4 : 8);
    if (!leadItems.length && !cardsToAnimate.length) return;

    const ctx = gsap.context(() => {
      if (leadItems.length) {
        gsap.fromTo(
          leadItems,
          { autoAlpha: 0, y: 10 },
          {
            autoAlpha: 1,
            y: 0,
            duration: reduceMotion ? 0.28 : 0.42,
            stagger: reduceMotion ? 0.015 : 0.03,
            ease: "power2.out",
            overwrite: "auto"
          }
        );
      }

      if (cardsToAnimate.length) {
        gsap.fromTo(
          cardsToAnimate,
          { autoAlpha: 0, y: 16 },
          {
            autoAlpha: 1,
            y: 0,
            duration: reduceMotion ? 0.32 : 0.5,
            delay: reduceMotion ? 0.02 : 0.06,
            stagger: reduceMotion ? 0.02 : 0.04,
            ease: "power2.out",
            overwrite: "auto"
          }
        );
      }
    }, page);

    return () => ctx.revert();
  }, [loading]);

  return (
    <section ref={pageRef} className="section-shell pt-16 pb-12" aria-busy={hasSelectedCategory && loading ? "true" : "false"}>
      <div className="mb-8 flex flex-col gap-4">
        <p data-products-anim="lead" className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("products.title")}</p>
        <h1 data-products-anim="lead" className="font-display text-3xl md:text-4xl">{t("products.subtitle")}</h1>
      </div>

      <div className="mb-8 space-y-3">
        <div className="flex flex-wrap items-center gap-3">
          <button
            type="button"
            onClick={() => handleCategorySelect("all")}
            data-products-anim="lead"
            className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${normalizedActiveCategory === "all"
              ? "border-primary bg-primary text-sand"
              : "border-primary/20 text-primary/70 hover:border-primary/50"
              }`}
          >
            {t("products.filterAll")}
          </button>
          {categoriesLoading ? <CategorySkeletons /> : mainCategories.map((category) => {
            const slug = normalizeChiniSlug(category.slug, category.title_fa);
            return (
              <button
                key={category.id}
                type="button"
                onClick={() => handleCategorySelect(slug)}
                data-products-anim="lead"
                className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${normalizedActiveCategory === slug
                  ? "border-primary bg-primary text-sand"
                  : "border-primary/20 text-primary/70 hover:border-primary/50"
                  }`}
              >
                {getLocalized(category, lang)}
              </button>
            );
          })}
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
            data-products-anim="lead"
            placeholder={t("products.searchPlaceholder")}
            className="w-full rounded-full border border-primary/20 bg-white/70 px-4 py-2.5 pr-9 text-sm font-semibold text-primary outline-none transition focus:border-primary/60"
          />
          {showSearchLoading && (
            <span
              aria-label={t("messages.loading")}
              className="absolute right-3 top-1/2 h-3 w-3 -translate-y-1/2 animate-spin rounded-full border border-primary/40 border-t-transparent"
            />
          )}
        </div>

      </div>

      {!hasSelectedCategory ? (
        <div className="mx-auto flex min-h-[260px] w-full max-w-3xl flex-col items-center justify-center rounded-3xl bg-primary/5 px-6 py-8 text-center md:w-1/2">
          <p className="font-display text-2xl text-primary md:text-3xl">{t("products.selectCategoryTitle")}</p>
          <p className="mt-3 text-sm text-primary/70">{t("products.selectCategoryHint")}</p>
          <div className="mt-6 w-full max-w-2xl rounded-2xl border border-primary/15 bg-white/50 px-5 py-4 text-center">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-primary/65">{t("products.supplyNoteTitle")}</p>
            <p className="mt-2 text-sm leading-7 text-primary/80">{t("products.supplyNoteBody")}</p>
          </div>
        </div>
      ) : loading && filtered.length === 0 ? (
        <ProductGridSkeleton />
      ) : filtered.length === 0 ? (
        emptyStateVisible && !hasMoreProducts ? (
          <div className="mx-auto mt-10 flex min-h-[180px] w-full items-center justify-center text-center">
            <p className="text-sm text-primary/70">{t("products.empty")}</p>
          </div>
        ) : (
          <ProductGridSkeleton />
        )
      ) : (
        <>
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {filtered.map((product) => {
              const isRTL = lang === "fa" || lang === "ar";
              const gradientDir = isRTL ? "bg-gradient-to-tl" : "bg-gradient-to-tr";
              return (
                <Link
                  key={product.id}
                  to={`/${lang}/products/${product.slug}`}
                  state={{ catalogRouteKind: "product" }}
                  onClick={rememberListState}
                  data-products-anim="card"
                  className="group flex h-full flex-col overflow-hidden transition hover:-translate-y-1 hover:shadow-xl"
                  style={{ contentVisibility: "auto", containIntrinsicSize: "360px" }}
                >
                  <div className="relative aspect-square w-full overflow-hidden bg-primary/10">
                    {product.is_popular && (
                      <span className="absolute left-3 top-3 z-10 rounded-full border border-white/90 bg-white/40 px-3 py-1 text-[11px] font-semibold text-white shadow-sm backdrop-blur">
                        {t("products.popularBadge")}
                      </span>
                    )}
                    {product.image_url ? (
                      <img
                        src={resolveImageUrl(product.image_url)}
                        alt={getLocalized(product, lang) || product.title_en}
                        loading="lazy"
                        decoding="async"
                        className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
                      />
                    ) : (
                      <div className="flex h-full items-center justify-center text-sm text-primary/60">{t("productDetail.noImages")}</div>
                    )}

                    <div className={`pointer-events-none absolute inset-x-0 bottom-0 ${gradientDir} from-black/70 via-black/25 to-transparent p-4`}>
                      <h3 className="font-display text-xl leading-tight text-white drop-shadow-[0_10px_24px_rgba(0,0,0,0.55)]">
                        {getLocalized(product, lang) || product.title_en}
                      </h3>
                      <div className="mt-2 space-y-1 text-[11px] font-semibold text-white/85 drop-shadow-[0_8px_18px_rgba(0,0,0,0.45)]">
                        {Array.isArray(product.finishes) && product.finishes.length > 0 && (
                          <p className="truncate">{product.finishes.join(" • ")}</p>
                        )}
                        {Array.isArray(product.variants) && product.variants.length > 0 && (
                          <p className="truncate">{product.variants.join(" • ")}</p>
                        )}
                        {Array.isArray(product.mines) && product.mines.length > 0 && (
                          <p className="truncate">{product.mines.join(" • ")}</p>
                        )}
                      </div>
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
          {loadingMore && (
            <div className="mt-6">
              <ProductGridSkeleton />
            </div>
          )}
        </>
      )}

      {!loading && hasSelectedCategory && (
        <div className="mt-8">
          {hasMoreProducts && <div ref={loadMoreTriggerRef} className="h-8 w-full" />}
        </div>
      )}
    </section>
  );
}
