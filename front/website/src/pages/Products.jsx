import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const getLocalized = (item, lang) => {
  if (!item) return "";
  if (lang === "fa") return item.title_fa;
  if (lang === "ar") return item.title_ar;
  return item.title_en;
};

const WIDTH_TAB_KEY = "width-40";

const stoneTypeTargets = [
  {
    key: "travertine",
    slugs: ["travertine"],
    fa: ["تراورتن"],
    en: ["travertine"],
    ar: ["ترافرتين"]
  },
  {
    key: "marble-stone",
    slugs: ["marble-stone", "marble"],
    fa: ["مرمریت"],
    en: ["marble"],
    ar: ["رخام"]
  },
  {
    key: "onyx",
    slugs: ["onyx"],
    fa: ["مرمر"],
    en: ["onyx"],
    ar: ["أونيكس", "اونكس"]
  },
  {
    key: "granite",
    slugs: ["granite"],
    fa: ["گرانیت"],
    en: ["granite"],
    ar: ["جرانيت"]
  },
  {
    key: "crystal",
    slugs: ["crystal"],
    fa: ["چینی", "کریستال"],
    en: ["crystal"],
    ar: ["كريستال", "كريستال"]
  }
];

const matchesTitle = (text, terms) => {
  if (!text) return false;
  return terms.some((term) => text.includes(term));
};

const categoryMatchesTarget = (category, target) => {
  const slug = category.slug?.toLowerCase() || "";
  if (target.slugs.some((value) => slug.includes(value))) return true;
  return (
    matchesTitle(category.title_fa, target.fa) ||
    matchesTitle(category.title_en?.toLowerCase(), target.en) ||
    matchesTitle(category.title_ar, target.ar)
  );
};

const findCategoryByTarget = (categories, target) => {
  return categories.find((category) => categoryMatchesTarget(category, target));
};

const isStoneTypeCategory = (category) => {
  return stoneTypeTargets.some((target) => categoryMatchesTarget(category, target));
};

const findWidthCategory = (categories) => {
  return categories.find((category) => {
    const slug = category.slug?.toLowerCase() || "";
    const faTitle = category.title_fa || "";
    return (
      slug.includes("40") ||
      faTitle.includes("عرض 40") ||
      faTitle.includes("عرض ۴۰") ||
      faTitle.includes("۴۰") ||
      faTitle.includes("40")
    );
  });
};

export default function Products() {
  const { t, lang } = useTranslation();
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [activeCategory, setActiveCategory] = useState("all");
  const [activeSubCategory, setActiveSubCategory] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      try {
        const [productRes, categoryRes] = await Promise.all([
          fetchJSON("/api/products"),
          fetchJSON("/api/categories")
        ]);
        if (!mounted) return;
        setProducts(productRes.data || []);
        setCategories(categoryRes.data || []);
      } catch (error) {
        if (!mounted) return;
        setProducts([]);
        setCategories([]);
      } finally {
        if (mounted) setLoading(false);
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, []);

  const filtered = useMemo(() => {
    if (activeCategory === "all") return products;
    if (activeCategory === WIDTH_TAB_KEY && activeSubCategory) {
      return products.filter((product) => product.category?.slug === activeSubCategory);
    }
    return products.filter((product) => product.category?.slug === activeCategory);
  }, [activeCategory, activeSubCategory, products]);

  const widthCategory = useMemo(() => findWidthCategory(categories), [categories]);
  const stoneTypeCategories = useMemo(
    () => stoneTypeTargets.map((target) => findCategoryByTarget(categories, target)).filter(Boolean),
    [categories]
  );

  const widthSubTabs = useMemo(() => {
    if (widthCategory) {
      const children = categories.filter((category) => category.parent_id === widthCategory.id);
      if (children.length) return children;
    }
    return stoneTypeCategories;
  }, [categories, stoneTypeCategories, widthCategory]);

  const defaultSubTab = useMemo(() => {
    const travertineTarget = stoneTypeTargets[0];
    return (
      findCategoryByTarget(widthSubTabs, travertineTarget)?.slug ||
      widthSubTabs[0]?.slug ||
      ""
    );
  }, [widthSubTabs]);

  useEffect(() => {
    if (activeCategory === WIDTH_TAB_KEY) {
      setActiveSubCategory(defaultSubTab);
    }
  }, [activeCategory, defaultSubTab]);

  const mainCategories = useMemo(() => {
    return categories.filter((category) => {
      if (widthCategory?.id && category.id === widthCategory.id) return false;
      if (isStoneTypeCategory(category)) return false;
      return true;
    });
  }, [categories, widthCategory]);

  return (
    <section className="section-shell py-16">
      <div className="mb-8 flex flex-col gap-4">
        <p className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("products.title")}</p>
        <h1 className="font-display text-3xl md:text-4xl">{t("products.subtitle")}</h1>
      </div>

      <div className="mb-8 flex flex-wrap items-center gap-3">
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
        {(widthCategory || widthSubTabs.length > 0) && (
          <button
            type="button"
            onClick={() => setActiveCategory(WIDTH_TAB_KEY)}
            className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${
              activeCategory === WIDTH_TAB_KEY
                ? "border-primary bg-primary text-sand"
                : "border-primary/20 text-primary/70 hover:border-primary/50"
            }`}
          >
            {t("products.width40")}
          </button>
        )}
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

      {activeCategory === WIDTH_TAB_KEY && widthSubTabs.length > 0 && (
        <div className="mb-10 flex flex-wrap items-center gap-3 border-t border-primary/10 pt-4">
          {widthSubTabs.map((category) => (
            <button
              key={category.id}
              type="button"
              onClick={() => setActiveSubCategory(category.slug)}
              className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${
                activeSubCategory === category.slug
                  ? "border-accent bg-accent/10 text-primary"
                  : "border-primary/20 text-primary/70 hover:border-primary/50"
              }`}
            >
              {getLocalized(category, lang)}
            </button>
          ))}
        </div>
      )}

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : filtered.length === 0 ? (
        <p className="text-sm text-primary/70">{t("products.empty")}</p>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filtered.map((product) => (
            <Link
              key={product.id}
              to={`/products/${product.slug}`}
              className="glass-panel flex h-full flex-col rounded-2xl p-5 transition hover:-translate-y-1 hover:shadow-2xl"
            >
              <div
                className="mb-4 h-40 rounded-xl bg-primary/10 bg-cover bg-center"
                style={
                  product.image_url
                    ? { backgroundImage: `url(${resolveImageUrl(product.image_url)})` }
                    : undefined
                }
              />
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">
                {product.category ? getLocalized(product.category, lang) : t("products.categoryLabel")}
              </p>
              <h3 className="mt-2 font-display text-xl">
                {getLocalized(product, lang) || product.title_en}
              </h3>
              <p className="mt-2 text-sm text-primary/70">{product.description}</p>
              <div className="mt-auto pt-4 text-sm font-semibold text-accent">
                {t("products.priceLabel")}: {product.price ? product.price : t("messages.empty")}
              </div>
            </Link>
          ))}
        </div>
      )}
    </section>
  );
}
