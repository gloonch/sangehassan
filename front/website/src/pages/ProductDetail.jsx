import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const getLocalized = (item, lang) => {
  if (!item) return "";
  if (lang === "fa") return item.title_fa;
  if (lang === "ar") return item.title_ar;
  return item.title_en;
};

export default function ProductDetail() {
  const { slug } = useParams();
  const { t, lang } = useTranslation();
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeIndex, setActiveIndex] = useState(0);
  const [lightboxOpen, setLightboxOpen] = useState(false);

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      try {
        const res = await fetchJSON(`/api/products/${slug}`);
        if (!mounted) return;
        setProduct(res.data || null);
      } catch (error) {
        if (!mounted) return;
        setProduct(null);
      } finally {
        if (mounted) setLoading(false);
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, [slug]);

  const images = useMemo(() => {
    if (!product) return [];
    if (product.images?.length) return product.images;
    if (product.image_url) return [product.image_url];
    return [];
  }, [product]);

  const activeImage = images[activeIndex] || images[0] || "";
  const localizedTitle = getLocalized(product, lang) || product?.title_en || "";

  const attributes = product?.attributes ? Object.entries(product.attributes) : [];

  return (
    <section className="section-shell py-16">
      <div className="mb-8 flex flex-wrap items-center justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("products.title")}</p>
          <h1 className="mt-3 font-display text-3xl md:text-4xl">{localizedTitle}</h1>
        </div>
        <Link
          to="/products"
          className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70 transition hover:border-primary/50"
        >
          {t("productDetail.backToProducts")}
        </Link>
      </div>

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : !product ? (
        <p className="text-sm text-primary/70">{t("productDetail.notFound")}</p>
      ) : (
        <div className="grid gap-8 lg:grid-cols-[1.2fr_1fr]">
          <div className="space-y-5">
            <div className="glass-panel relative overflow-hidden rounded-3xl border border-white/60 bg-quiet-grid p-4">
              <div className="aspect-[4/3] overflow-hidden rounded-2xl bg-primary/10">
                {activeImage ? (
                  <button
                    type="button"
                    className="group relative h-full w-full"
                    onClick={() => setLightboxOpen(true)}
                  >
                    <img
                      src={resolveImageUrl(activeImage)}
                      alt={localizedTitle}
                      className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
                    />
                    <span className="absolute bottom-4 right-4 rounded-full bg-primary/80 px-4 py-2 text-xs font-semibold text-sand shadow-lg">
                      {t("productDetail.openGallery")}
                    </span>
                  </button>
                ) : (
                  <div className="flex h-full items-center justify-center text-sm text-primary/60">
                    {t("productDetail.noImages")}
                  </div>
                )}
              </div>
            </div>

            {images.length > 0 && (
              <div className="flex gap-3 overflow-x-auto pb-2">
                {images.map((img, index) => (
                  <button
                    key={`${img}-${index}`}
                    type="button"
                    onClick={() => setActiveIndex(index)}
                    className={`h-20 w-24 flex-shrink-0 overflow-hidden rounded-2xl border transition ${
                      activeIndex === index ? "border-accent" : "border-primary/10"
                    }`}
                  >
                    <img
                      src={resolveImageUrl(img)}
                      alt={`${localizedTitle}-${index}`}
                      className="h-full w-full object-cover"
                    />
                  </button>
                ))}
              </div>
            )}
          </div>

          <div className="space-y-6">
            <div className="glass-panel rounded-3xl p-6">
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.category")}</p>
              <div className="mt-3 flex flex-wrap gap-2">
                {(product.categories?.length ? product.categories : product.category ? [product.category] : []).map((cat) => (
                  <span
                    key={cat.id}
                    className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold text-primary/70"
                  >
                    {getLocalized(cat, lang)}
                  </span>
                ))}
              </div>
              <div className="mt-6 flex items-baseline gap-3">
                <span className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.price")}</span>
                <span className="text-lg font-semibold text-accent">
                  {product.price ? product.price : t("messages.empty")}
                </span>
              </div>
            </div>

            <div className="glass-panel rounded-3xl p-6">
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.description")}</p>
              {product.description_html ? (
                <div
                  className="mt-3 space-y-3 text-sm text-primary/70"
                  dangerouslySetInnerHTML={{ __html: product.description_html }}
                />
              ) : (
                <p className="mt-3 text-sm text-primary/70">{t("messages.empty")}</p>
              )}
            </div>

            {attributes.length > 0 && (
              <div className="glass-panel rounded-3xl p-6">
                <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.attributes")}</p>
                <div className="mt-4 space-y-4">
                  {attributes.map(([name, terms]) => (
                    <div key={name}>
                      <p className="text-sm font-semibold text-primary">{name}</p>
                      <div className="mt-2 flex flex-wrap gap-2">
                        {terms.map((term) => (
                          <span
                            key={`${name}-${term}`}
                            className="rounded-full bg-primary/10 px-3 py-1 text-xs font-semibold text-primary/70"
                          >
                            {term}
                          </span>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {lightboxOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-primary/80 px-4 py-10">
          <div className="relative w-full max-w-5xl">
            <button
              type="button"
              onClick={() => setLightboxOpen(false)}
              className="absolute right-4 top-4 rounded-full bg-white/80 px-3 py-1 text-xs font-semibold text-primary"
            >
              {t("actions.close")}
            </button>
            <div className="overflow-hidden rounded-3xl border border-white/30 bg-white/10 p-4">
              <img
                src={resolveImageUrl(activeImage)}
                alt={localizedTitle}
                className="h-[70vh] w-full object-contain"
              />
            </div>
            {images.length > 1 && (
              <div className="mt-4 flex items-center justify-between">
                <button
                  type="button"
                  onClick={() => setActiveIndex((idx) => (idx - 1 + images.length) % images.length)}
                  className="rounded-full border border-white/40 px-4 py-2 text-xs font-semibold text-white"
                >
                  {t("productDetail.prev")}
                </button>
                <div className="text-xs text-white/70">
                  {activeIndex + 1} / {images.length}
                </div>
                <button
                  type="button"
                  onClick={() => setActiveIndex((idx) => (idx + 1) % images.length)}
                  className="rounded-full border border-white/40 px-4 py-2 text-xs font-semibold text-white"
                >
                  {t("productDetail.next")}
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </section>
  );
}
