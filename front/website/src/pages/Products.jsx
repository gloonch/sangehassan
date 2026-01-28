import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const getLocalized = (item, lang) => {
  if (!item) return "";
  if (lang === "fa") return item.title_fa;
  if (lang === "ar") return item.title_ar;
  return item.title_en;
};

export default function Products() {
  const { t, lang } = useTranslation();
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [activeCategory, setActiveCategory] = useState("all");
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
    return products.filter((product) => product.category?.slug === activeCategory);
  }, [activeCategory, products]);

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
        {categories.map((category) => (
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

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : filtered.length === 0 ? (
        <p className="text-sm text-primary/70">{t("products.empty")}</p>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filtered.map((product) => (
            <div key={product.id} className="glass-panel flex h-full flex-col rounded-2xl p-5">
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
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
