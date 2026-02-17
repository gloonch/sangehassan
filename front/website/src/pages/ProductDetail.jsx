import { useEffect, useMemo, useRef, useState } from "react";
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

const getLocalizedTerm = (term, lang) => {
  if (!term) return "";
  if (lang === "fa") return term.label_fa || term.label_en || "";
  if (lang === "ar") return term.label_ar || term.label_en || "";
  return term.label_en || "";
};

export default function ProductDetail() {
  const { slug } = useParams();
  const { t, lang } = useTranslation();
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeIndex, setActiveIndex] = useState(0);
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const backRowRef = useRef(null);
  const heroRef = useRef(null);
  const scrollAlignedRef = useRef(false);

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

  useEffect(() => {
    scrollAlignedRef.current = false;
  }, [slug]);

  const images = useMemo(() => {
    if (!product) return [];
    if (product.images?.length) return product.images;
    if (product.image_url) return [product.image_url];
    return [];
  }, [product]);

  useEffect(() => {
    if (activeIndex >= images.length) {
      setActiveIndex(0);
    }
  }, [activeIndex, images.length]);

  useEffect(() => {
    if (loading || !product || scrollAlignedRef.current) return;

    const alignScrollToDetailStart = () => {
      const backEl = backRowRef.current;
      const heroEl = heroRef.current;
      if (!backEl || !heroEl) return;

      const navEl = document.querySelector("header");
      const navHeight = navEl ? navEl.getBoundingClientRect().height : 0;

      const backRect = backEl.getBoundingClientRect();
      const heroRect = heroEl.getBoundingClientRect();
      const gap = Math.max(0, heroRect.top - backRect.bottom);

      // Place navbar bottom in the middle of the gap between back button row and hero image.
      const desiredHeroTop = navHeight + gap / 2;
      const delta = heroRect.top - desiredHeroTop;
      const targetY = Math.max(0, window.scrollY + delta);

      window.scrollTo({ top: targetY, behavior: "auto" });
      scrollAlignedRef.current = true;
    };

    const raf1 = window.requestAnimationFrame(() => {
      window.requestAnimationFrame(alignScrollToDetailStart);
    });

    return () => {
      window.cancelAnimationFrame(raf1);
    };
  }, [loading, product]);

  const activeImage = images[activeIndex] || images[0] || "";
  const localizedTitle = getLocalized(product, lang) || product?.title_en || "";
  const localizedDescriptionHTML =
    (lang === "fa" ? product?.description_html_fa : lang === "ar" ? product?.description_html_ar : product?.description_html_en) ||
    product?.description_html ||
    product?.description ||
    "";

  const terms = product?.terms || [];
  const categories = product?.categories?.length ? product.categories : product?.category ? [product.category] : [];
  const categoryLine = categories.map((cat) => getLocalized(cat, lang)).filter(Boolean).join(" • ");

  const termsByTaxonomy = useMemo(() => {
    const grouped = {};
    for (const term of terms) {
      if (!term?.taxonomy) continue;
      grouped[term.taxonomy] = grouped[term.taxonomy] || [];
      grouped[term.taxonomy].push(term);
    }
    return grouped;
  }, [terms]);

  const infoSections = [
    { key: "stone_type", label: t("productDetail.stoneType") },
    { key: "tone", label: t("productDetail.tone") },
    { key: "pattern", label: t("productDetail.pattern") },
    { key: "visual_impact", label: t("productDetail.visualImpact") },
    { key: "use_case_space", label: t("productDetail.useCaseSpaces") },
    { key: "use_case_form", label: t("productDetail.useCaseForms") },
    { key: "use_case_application", label: t("productDetail.useCaseApplications") },
    { key: "use_case_project_type", label: t("productDetail.useCaseProjects") },
    { key: "use_case_special", label: t("productDetail.useCaseSpecial") }
  ];

  const hasMoreInfo = infoSections.some((section) => (termsByTaxonomy[section.key] || []).length > 0);

  const goPrev = () => {
    if (images.length <= 1) return;
    setActiveIndex((idx) => (idx - 1 + images.length) % images.length);
  };

  const goNext = () => {
    if (images.length <= 1) return;
    setActiveIndex((idx) => (idx + 1) % images.length);
  };

  return (
    <section className="section-shell py-16">
      <div ref={backRowRef} className="mb-6 flex items-center justify-end">
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
        <div className="space-y-8">
          <div ref={heroRef} className="relative overflow-hidden rounded-3xl bg-primary/10">
            {activeImage ? (
              <button
                type="button"
                className="group relative h-[56vh] min-h-[420px] w-full md:h-[68vh]"
                onClick={() => setLightboxOpen(true)}
              >
                <img
                  src={resolveImageUrl(activeImage)}
                  alt={localizedTitle}
                  className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
                />
              </button>
            ) : (
              <div className="flex h-[56vh] min-h-[420px] w-full items-center justify-center text-sm text-primary/60">
                {t("productDetail.noImages")}
              </div>
            )}

            <div className="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-primary/85 via-primary/45 to-transparent p-6 md:p-8">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between md:gap-8">
                <div className="min-w-0 flex-1" style={{ textAlign: "start" }}>
                  <p className="text-sm font-semibold uppercase tracking-[0.24em] text-white/80">
                    {categoryLine || t("productDetail.category")}
                  </p>
                  <h1 className="mt-2 font-display text-4xl leading-tight text-white md:text-6xl">{localizedTitle}</h1>
                </div>

                <div className="flex-shrink-0" style={{ textAlign: "end" }}>
                  <p className="text-xs font-semibold uppercase tracking-[0.24em] text-white/80">{t("productDetail.price")}</p>
                  <p className="mt-1 text-3xl font-semibold leading-none text-white drop-shadow-[0_2px_10px_rgba(0,0,0,0.45)] md:text-5xl">
                    {product.price ? product.price : t("messages.empty")}
                  </p>
                </div>
              </div>
            </div>

            {images.length > 1 && (
              <>
                <button
                  type="button"
                  onClick={goPrev}
                  className="absolute left-4 top-1/2 -translate-y-1/2 rounded-full border border-white/40 bg-black/20 px-3 py-2 text-xs font-semibold text-white backdrop-blur-sm transition hover:bg-black/30"
                >
                  {t("productDetail.prev")}
                </button>
                <button
                  type="button"
                  onClick={goNext}
                  className="absolute right-4 top-1/2 -translate-y-1/2 rounded-full border border-white/40 bg-black/20 px-3 py-2 text-xs font-semibold text-white backdrop-blur-sm transition hover:bg-black/30"
                >
                  {t("productDetail.next")}
                </button>
              </>
            )}
          </div>

          {images.length > 1 && (
            <div className="flex items-center justify-center gap-2">
              {images.map((_, index) => (
                <button
                  key={`dot-${index}`}
                  type="button"
                  onClick={() => setActiveIndex(index)}
                  aria-label={`slide-${index + 1}`}
                  className={`h-2.5 w-2.5 rounded-full transition ${activeIndex === index ? "bg-accent" : "bg-primary/25"}`}
                />
              ))}
            </div>
          )}

          <div className={`grid gap-6 ${hasMoreInfo ? "lg:grid-cols-[2fr_1fr]" : ""}`}>
            <div className="glass-panel h-fit self-start rounded-3xl p-6 md:p-8">
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.description")}</p>
              {localizedDescriptionHTML ? (
                <div className="mt-4 space-y-3 text-sm text-primary/75" dangerouslySetInnerHTML={{ __html: localizedDescriptionHTML }} />
              ) : (
                <p className="mt-4 text-sm text-primary/70">{t("messages.empty")}</p>
              )}
            </div>

            {hasMoreInfo && (
              <div className="glass-panel rounded-3xl p-6 md:p-8">
                <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.moreInfo")}</p>
                <div className="mt-5 space-y-5">
                  {infoSections.map((section) => {
                    const items = termsByTaxonomy[section.key] || [];
                    if (items.length === 0) return null;
                    return (
                      <div key={section.key}>
                        <p className="text-sm font-semibold text-primary">{section.label}</p>
                        <div className="mt-2 flex flex-wrap gap-2">
                          {items.map((term) => (
                            <span
                              key={`${section.key}-${term.id}`}
                              className="rounded-full bg-primary/10 px-3 py-1 text-xs font-semibold text-primary/70"
                            >
                              {getLocalizedTerm(term, lang)}
                            </span>
                          ))}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {lightboxOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-primary/80 px-4 py-10">
          <div className="relative w-full max-w-6xl">
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
                className="h-[78vh] w-full object-contain"
              />
            </div>
            {images.length > 1 && (
              <div className="mt-4 flex items-center justify-between">
                <button
                  type="button"
                  onClick={goPrev}
                  className="rounded-full border border-white/40 px-4 py-2 text-xs font-semibold text-white"
                >
                  {t("productDetail.prev")}
                </button>
                <span className="text-xs font-semibold text-white/70">
                  {activeIndex + 1}/{images.length}
                </span>
                <button
                  type="button"
                  onClick={goNext}
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
