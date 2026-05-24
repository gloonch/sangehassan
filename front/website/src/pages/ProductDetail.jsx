import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { usePageSeo } from "../lib/seo";

const stripHTML = (value) => (value || "").replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim();

const iconBaseProps = {
  xmlns: "http://www.w3.org/2000/svg",
  fill: "none",
  viewBox: "0 0 24 24",
  strokeWidth: 1.6,
  stroke: "currentColor",
  strokeLinecap: "round",
  strokeLinejoin: "round"
};

const PhoneIcon = ({ className }) => (
  <svg {...iconBaseProps} className={className} aria-hidden="true">
    <path d="M6.5 3h4l2 5l-2.5 2.5a16 16 0 0 0 3.5 3.5L16 11.5l5 2v4a2 2 0 0 1-2 2C10.7 19.5 4.5 13.3 4.5 5a2 2 0 0 1 2-2z" />
  </svg>
);

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

const toPercent = (value) => {
  if (typeof value === "number" && Number.isFinite(value)) {
    if (value >= 0 && value <= 1) return value * 100;
    if (value >= 0 && value <= 100) return value;
    return null;
  }

  if (typeof value === "string") {
    const cleaned = value.trim().replace("%", "");
    const parsed = Number(cleaned);
    if (!Number.isFinite(parsed)) return null;
    if (parsed >= 0 && parsed <= 1 && !cleaned.includes("%")) return parsed * 100;
    if (parsed >= 0 && parsed <= 100) return parsed;
  }

  return null;
};

const toLatinDigits = (value) =>
  value
    .replace(/[۰-۹]/g, (digit) => String("۰۱۲۳۴۵۶۷۸۹".indexOf(digit)))
    .replace(/[٠-٩]/g, (digit) => String("٠١٢٣٤٥٦٧٨٩".indexOf(digit)));

const normalizePhone = (value) => {
  if (!value || typeof value !== "string") return "";
  const cleaned = toLatinDigits(value).replace(/[^\d+]/g, "");
  if (!cleaned) return "";
  if (cleaned.startsWith("+")) {
    return `+${cleaned.slice(1).replace(/\+/g, "")}`;
  }
  return cleaned.replace(/\+/g, "");
};

export default function ProductDetail() {
  const { slug } = useParams();
  const { t, lang } = useTranslation();
  const isRTL = lang === "fa" || lang === "ar";
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeIndex, setActiveIndex] = useState(0);
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [activeHotspotId, setActiveHotspotId] = useState(null);

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

  useEffect(() => {
    if (activeIndex >= images.length) {
      setActiveIndex(0);
    }
  }, [activeIndex, images.length]);

  useEffect(() => {
    setActiveHotspotId(null);
  }, [activeIndex, slug]);

  const activeImage = images[activeIndex] || images[0] || "";
  const localizedTitle = getLocalized(product, lang) || product?.title_en || "";
  const localizedDescriptionHTML =
    (lang === "fa" ? product?.description_html_fa : lang === "ar" ? product?.description_html_ar : product?.description_html_en) ||
    product?.description_html ||
    product?.description ||
    "";
  const seoTitle = localizedTitle ? `${localizedTitle} | SangeHassan` : "Product Detail | SangeHassan";
  const seoDescription =
    stripHTML(localizedDescriptionHTML).slice(0, 160) ||
    "Detailed natural stone product page from SangeHassan with images, finishes, variants, and sourcing information.";

  usePageSeo({
    title: seoTitle,
    description: seoDescription,
    path: `/products/${slug}`,
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    image: activeImage ? resolveImageUrl(activeImage) : "",
    type: "product",
    robots: product ? "index,follow,max-image-preview:large" : "noindex,follow"
  });

  const terms = product?.terms || [];
  const categories = product?.categories?.length ? product.categories : product?.category ? [product.category] : [];
  const categoriesAdjusted = categories.map((cat) =>
    cat.title_fa === "چینی" || cat.title_fa === "چینی کریستال"
      ? { ...cat, title_fa: "چینی کریستال", slug: "chinese-crystal" }
      : cat
  );
  const categoryLine = categoriesAdjusted.map((cat) => getLocalized(cat, lang)).filter(Boolean).join(" • ");

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
  const metaLists = [
    { key: "variants", label: t("productDetail.variants"), items: product?.variants || [], linkable: false },
    { key: "mines", label: t("productDetail.mines"), items: product?.mines || [], linkable: true },
    { key: "finishes", label: t("productDetail.finishes"), items: product?.finishes || [], linkable: true }
  ];
  const hasMetaLists = metaLists.some((entry) => entry.items.length > 0);
  const hasDetailBlocks = hasMoreInfo || hasMetaLists;
  const phoneValue = t("footer.phoneValue");
  const normalizedPhone = normalizePhone(phoneValue);
  const phoneHref = normalizedPhone ? `tel:${normalizedPhone}` : "";
  const siteUrl = String(import.meta.env.VITE_SITE_URL || "").toLowerCase();
  const showDomesticMessengerLinks = import.meta.env.DEV || siteUrl.includes("sangehassan.ir");
  const primaryCategorySlug = categoriesAdjusted[0]?.slug || "";

  const buildFilterLink = (value) => {
    const params = new URLSearchParams();
    if (primaryCategorySlug) {
      params.set("category", primaryCategorySlug);
    }
    params.set("q", String(value || "").trim());
    return `/products?${params.toString()}`;
  };

  const fallbackHotspots = useMemo(() => {
    const options = [];
    const stoneType = terms.find((term) => term?.taxonomy === "stone_type");
    if (stoneType) options.push(getLocalizedTerm(stoneType, lang));
    if (product?.finishes?.[0]) options.push(product.finishes[0]);
    if (product?.variants?.[0]) options.push(product.variants[0]);
    if (product?.mines?.[0]) options.push(product.mines[0]);

    const seed = options.filter(Boolean).slice(0, 4);
    const points = [
      { x: 21, y: 28 },
      { x: 72, y: 40 },
      { x: 50, y: 66 },
      { x: 31, y: 56 }
    ];

    return seed.map((label, index) => ({
      id: `fallback-hotspot-${index}`,
      label,
      x: points[index].x,
      y: points[index].y
    }));
  }, [lang, product?.finishes, product?.mines, product?.variants, terms]);

  const hotspots = useMemo(() => {
    const rawHotspots = Array.isArray(product?.hotspots)
      ? product.hotspots
      : Array.isArray(product?.image_hotspots)
        ? product.image_hotspots
        : [];

    const fallbackLabelAt = (index) => {
      if (!fallbackHotspots.length) return "";
      return fallbackHotspots[index % fallbackHotspots.length]?.label || "";
    };

    const parsed = rawHotspots
      .map((point, index) => {
        const x = toPercent(point?.x ?? point?.left ?? point?.position_x ?? point?.x_pct ?? point?.x_percent);
        const y = toPercent(point?.y ?? point?.top ?? point?.position_y ?? point?.y_pct ?? point?.y_percent);
        if (x === null || y === null) return null;

        const localizedLabel =
          (lang === "fa" ? point?.label_fa : lang === "ar" ? point?.label_ar : point?.label_en) ||
          point?.label ||
          point?.title ||
          point?.name ||
          "";

        return {
          id: point?.id || `hotspot-${index}`,
          label: localizedLabel || fallbackLabelAt(index),
          x,
          y
        };
      })
      .filter(Boolean)
      .slice(0, 6);

    if (parsed.length > 0) return parsed;
    return fallbackHotspots;
  }, [fallbackHotspots, lang, product?.hotspots, product?.image_hotspots]);

  const goPrev = () => {
    if (images.length <= 1) return;
    setActiveIndex((idx) => (idx - 1 + images.length) % images.length);
  };

  const goNext = () => {
    if (images.length <= 1) return;
    setActiveIndex((idx) => (idx + 1) % images.length);
  };

  return (
    <section className="section-shell pt-4 pb-12">

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : !product ? (
        <p className="text-sm text-primary/70">{t("productDetail.notFound")}</p>
      ) : (
        <div className="space-y-8">
          <div className="space-y-4">
            <div className={`flex items-center ${isRTL ? "justify-end" : "justify-start"}`}>
              <Link
                to="/products"
                className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70 transition hover:border-primary/50"
              >
                {t("productDetail.backToProducts")}
              </Link>
            </div>

            <div className="relative overflow-hidden bg-primary/10">
              {activeImage ? (
                <button
                  type="button"
                  className="group relative h-[42vh] min-h-[320px] w-full md:h-[52vh]"
                  onClick={() => setLightboxOpen(true)}
                >
                  <img
                    src={resolveImageUrl(activeImage)}
                    alt={localizedTitle}
                    className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.02]"
                  />
                </button>
              ) : (
                <div className="flex h-[56vh] min-h-[420px] w-full items-center justify-center text-sm text-primary/60">
                  {t("productDetail.noImages")}
                </div>
              )}

            </div>

            {images.length > 0 && (
              <div className="overflow-x-auto no-scrollbar pb-1">
                <div className="flex flex-nowrap gap-2">
                  {images.map((image, index) => {
                    const isActive = index === activeIndex;
                    return (
                      <button
                        key={`${image}-${index}`}
                        type="button"
                        onClick={() => setActiveIndex(index)}
                        aria-label={`thumb-${index + 1}`}
                        className={`relative h-16 w-16 flex-shrink-0 overflow-hidden border transition ${isActive
                          ? "border-primary"
                          : "border-primary/20 hover:border-primary/45"
                          }`}
                      >
                        <img
                          src={resolveImageUrl(image)}
                          alt=""
                          className="h-full w-full object-cover"
                          loading="lazy"
                        />
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            <div className="space-y-2">
              {categoryLine ? (
                <p className="text-sm font-semibold uppercase tracking-[0.24em] text-primary/65">
                  {categoryLine}
                </p>
              ) : null}
              <div className="flex flex-wrap items-end gap-3 md:gap-4">
                <h1 className="font-display text-3xl leading-tight text-primary md:text-5xl">{localizedTitle}</h1>
                {phoneHref && (
                  <div className={`flex flex-col items-center gap-1.5 ${isRTL ? "mr-auto" : "ml-auto"}`}>
                    <a
                      href={phoneHref}
                      className="call-cta-shimmer inline-flex items-center gap-3 rounded-full border border-primary/25 px-3.5 py-2 text-primary/85 transition hover:border-primary/55 hover:text-primary"
                    >
                      <span className="inline-flex h-10 w-10 items-center justify-center rounded-full">
                        <PhoneIcon className="h-7 w-7" />
                      </span>
                      <span className="text-[11px] font-semibold md:text-xs">{t("productDetail.callForMoreInfo")}</span>
                      <span className="text-xs font-semibold tabular-nums md:text-sm" dir="ltr">{phoneValue}</span>
                    </a>

                    {showDomesticMessengerLinks && (
                      <div className="text-center">
                        <p className="text-[10px] text-primary/65">{t("productDetail.domesticMessengerHint")}</p>
                        <div className="mt-1 flex flex-wrap justify-center gap-1.5">
                          <a
                            href="https://bale.ir/sangehassan"
                            target="_blank"
                            rel="noreferrer"
                            className="rounded-full border border-primary/20 px-2 py-0.5 text-[10px] font-semibold text-primary/70 transition hover:border-primary/45 hover:text-primary"
                          >
                            bale
                          </a>
                          <a
                            href="https://rubika.ir/sangehassan"
                            target="_blank"
                            rel="noreferrer"
                            className="rounded-full border border-primary/20 px-2 py-0.5 text-[10px] font-semibold text-primary/70 transition hover:border-primary/45 hover:text-primary"
                          >
                            rubika
                          </a>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className="space-y-8 border-t border-primary/20 pt-6">
            {hasDetailBlocks && (
              <section className="space-y-4">
                <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.moreInfo")}</p>
                <div className="space-y-5 text-sm text-primary">
                  {infoSections.map((section) => {
                    const items = termsByTaxonomy[section.key] || [];
                    if (items.length === 0) return null;
                    return (
                      <div key={section.key} className="space-y-2">
                        <p className="font-semibold">{section.label}</p>
                        <div className="flex flex-wrap gap-2 text-[12px] font-semibold text-primary/80">
                          {items.map((term) => (
                            <span key={`${section.key}-${term.id}`} className="px-2 py-1">
                              {getLocalizedTerm(term, lang)}
                            </span>
                          ))}
                        </div>
                      </div>
                    );
                  })}
                  {metaLists.map((entry) => {
                    if (!entry.items.length) return null;
                    return (
                      <div key={entry.key} className="space-y-2">
                        <p className="font-semibold">{entry.label}</p>
                        <div className="flex flex-wrap gap-2">
                          {entry.items.map((value, index) => (
                            entry.linkable ? (
                              <Link
                                key={`${entry.key}-${index}-${value}`}
                                to={buildFilterLink(value)}
                                className="inline-flex items-center gap-1.5 rounded-full border border-primary/20 px-3 py-1.5 text-[12px] font-semibold text-primary/80 transition hover:border-primary/45 hover:text-primary"
                              >
                                <span className="h-1.5 w-1.5 rounded-full bg-primary/70" />
                                <span>{value}</span>
                              </Link>
                            ) : (
                              <span
                                key={`${entry.key}-${index}-${value}`}
                                className="inline-flex items-center gap-1.5 rounded-full border border-primary/20 px-3 py-1.5 text-[12px] font-semibold text-primary/80"
                              >
                                <span className="h-1.5 w-1.5 rounded-full bg-primary/70" />
                                <span>{value}</span>
                              </span>
                            )
                          ))}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </section>
            )}

            <section className="space-y-4 border-t border-primary/20 pt-6">
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.description")}</p>
              {localizedDescriptionHTML ? (
                <div className="space-y-3 text-sm text-primary/75" dangerouslySetInnerHTML={{ __html: localizedDescriptionHTML }} />
              ) : (
                <p className="text-sm text-primary/70">{t("messages.empty")}</p>
              )}
            </section>

          </div>
        </div>
      )}

      {lightboxOpen && (
        <div className="fixed inset-0 z-[9999] bg-[#E5E1DD] text-primary">
          <div className="section-shell flex h-20 items-center justify-between gap-4">
            <p className="truncate text-xs font-semibold uppercase tracking-[0.22em] text-primary/60">
              {localizedTitle}
            </p>
            <button
              type="button"
              onClick={() => setLightboxOpen(false)}
              className="inline-flex h-10 w-10 items-center justify-center rounded-full text-primary transition hover:bg-primary/10"
              aria-label={t("actions.close")}
            >
              <span className="relative block h-4 w-4">
                <span className="absolute left-0 top-[7px] h-0.5 w-full -rotate-45 rounded-full bg-primary" />
                <span className="absolute left-0 top-[7px] h-0.5 w-full rotate-45 rounded-full bg-primary" />
              </span>
            </button>
          </div>

          <div className="section-shell flex h-[calc(100dvh-5rem)] flex-col pb-6">
            <div className="flex-1 overflow-hidden">
              <img
                src={resolveImageUrl(activeImage)}
                alt={localizedTitle}
                className="h-full w-full object-contain"
              />
            </div>
            {images.length > 1 && (
              <div className="mt-4 flex items-center justify-between gap-3">
                <button
                  type="button"
                  onClick={goPrev}
                  className="rounded-full border border-primary/25 px-4 py-2 text-xs font-semibold text-primary transition hover:border-primary/50"
                >
                  {t("productDetail.prev")}
                </button>
                <span className="text-xs font-semibold text-primary/65">
                  {activeIndex + 1}/{images.length}
                </span>
                <button
                  type="button"
                  onClick={goNext}
                  className="rounded-full border border-primary/25 px-4 py-2 text-xs font-semibold text-primary transition hover:border-primary/50"
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
