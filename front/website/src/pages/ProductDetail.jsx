import { useEffect, useMemo, useState } from "react";
import { Link, useLocation, useParams } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl, resolveProtectedImageUrl, resolveProtectedThumbnailUrl } from "../lib/assets";
import { getAbsoluteUrl, getCanonicalUrl, usePageSeo } from "../lib/seo";
import { usePrerenderData } from "../lib/prerenderData";
import ProtectedImage from "../components/ProtectedImage";
import { catalogAlternates } from "../lib/catalogLocale";
import { hasLegacyProductsReturnState, readCatalogProductReturnState } from "../lib/productReturnState";
import { formatOfferPrice, getProductOfferPrice } from "../lib/productOffers";

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

const rtlScriptRegex = /[\u0600-\u06FF\u0750-\u077F\u08A0-\u08FF]/;
const persianSpecificRegex = /[پچژگکیی]/;

const isLabelUsableForLang = (value, lang) => {
  const label = String(value || "").trim();
  if (!label) return false;
  const hasRTL = rtlScriptRegex.test(label);
  if (lang === "en") return !hasRTL;
  if (lang === "ar") return hasRTL && !persianSpecificRegex.test(label);
  return hasRTL;
};

const getLocalizedTerm = (term, lang) => {
  if (!term) return "";
  const label = lang === "fa" ? term.label_fa : lang === "ar" ? term.label_ar : term.label_en;
  return isLabelUsableForLang(label, lang) ? String(label).trim() : "";
};

const clickableTermTaxonomies = new Set(["variants", "finishes"]);

const getTermLink = (sectionKey, term) => {
  const link = String(term?.link_url || "").trim();
  if (!clickableTermTaxonomies.has(sectionKey) || !link) return "";
  return link;
};

const getLocalizedProjectDescription = (project, lang) => {
  if (!project) return "";
  if (lang === "fa") return project.description_fa || project.description_en || project.description_ar || project.description || "";
  if (lang === "ar") return project.description_ar || project.description_en || project.description_fa || project.description || "";
  return project.description_en || project.description_fa || project.description_ar || project.description || "";
};

const getLocalizedProjectTitle = (project, lang, t) => {
  if (!project) return "";

  const explicitTitle =
    lang === "fa"
      ? project.title_fa || project.title_en || project.title_ar
      : lang === "ar"
        ? project.title_ar || project.title_en || project.title_fa
        : project.title_en || project.title_fa || project.title_ar;

  if (explicitTitle) return explicitTitle;

  const description = getLocalizedProjectDescription(project, lang).trim();
  if (description) {
    const firstLine = description.split(/\r?\n/)[0].trim();
    if (firstLine) return firstLine.length > 56 ? `${firstLine.slice(0, 56)}...` : firstLine;
  }

  return `${t("projects.itemTitle")} ${project.id}`;
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
  const location = useLocation();
  const { t, lang } = useTranslation();
  const isRTL = lang === "fa" || lang === "ar";
  const prerenderedProduct = usePrerenderData("product");
  const initialProduct = prerenderedProduct?.slug === slug ? prerenderedProduct : null;
  const [product, setProduct] = useState(initialProduct);
  const [loading, setLoading] = useState(!initialProduct);
  const [activeIndex, setActiveIndex] = useState(0);
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [activeHotspotId, setActiveHotspotId] = useState(null);

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      const hasInitialProduct = initialProduct?.slug === slug;
      if (hasInitialProduct) {
        setLoading(false);
        return;
      }
      if (!hasInitialProduct) {
        setLoading(true);
      }
      try {
        const res = await fetchJSON(`/api/products/${slug}`);
        if (!mounted) return;
        setProduct(res.data || null);
      } catch (error) {
        if (!mounted) return;
        if (!hasInitialProduct) {
          setProduct(null);
        }
      } finally {
        if (mounted) setLoading(false);
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, [initialProduct, slug]);

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

  useEffect(() => {
    if (!lightboxOpen) return undefined;
    const originalOverflow = document.body.style.overflow;
    const handleKeyDown = (event) => {
      if (event.key === "Escape") {
        setLightboxOpen(false);
      }
    };

    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      document.body.style.overflow = originalOverflow;
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [lightboxOpen]);

  const activeImage = images[activeIndex] || images[0] || "";
  const openLightbox = () => {
    if (activeImage) {
      setLightboxOpen(true);
    }
  };
  const closeLightbox = () => setLightboxOpen(false);
  const localizedTitle = getLocalized(product, lang) || product?.title_en || "";
  const localizedDescriptionHTML =
    (lang === "fa" ? product?.description_html_fa : lang === "ar" ? product?.description_html_ar : product?.description_html_en) ||
    product?.description_html ||
    product?.description ||
    "";
  const relatedProducts = Array.isArray(product?.related_products) ? product.related_products : [];
  const relatedProjects = Array.isArray(product?.related_projects) ? product.related_projects : [];
  const seoTitle = localizedTitle ? `${localizedTitle} | SangeHassan` : "Product Detail | SangeHassan";
  const seoDescription =
    stripHTML(localizedDescriptionHTML).slice(0, 160) ||
    "Detailed natural stone product page from SangeHassan with images, sourcing information, and project references.";
  const localizedProductPath = `/${lang}/products/${slug}`;
  const isLegacyProductPath = /^\/products\/[^/]+\/?$/.test(location.pathname);
  const productOfferPrice = getProductOfferPrice(product);
  const productJsonLd = useMemo(() => {
    if (!product || !localizedTitle) return null;

    const pageUrl = getCanonicalUrl(localizedProductPath);
    const imageUrl = activeImage ? getAbsoluteUrl(resolveProtectedImageUrl(activeImage)) : undefined;
    const offer = productOfferPrice > 0
      ? {
        "@type": "Offer",
        url: pageUrl,
        priceCurrency: "IRR",
        price: String(productOfferPrice),
        availability: "https://schema.org/InStock",
        itemCondition: "https://schema.org/NewCondition",
        seller: {
          "@type": "Organization",
          name: "SangeHassan"
        },
        priceSpecification: {
          "@type": "UnitPriceSpecification",
          name: t("productDetail.offerLabel"),
          price: String(productOfferPrice),
          priceCurrency: "IRR",
          description: t("productDetail.offerSchemaNote")
        }
      }
      : undefined;

    return {
      "@context": "https://schema.org",
      "@type": "Product",
      "@id": `${pageUrl}#product`,
      name: localizedTitle,
      description: seoDescription,
      image: imageUrl,
      url: pageUrl,
      brand: {
        "@type": "Brand",
        name: "SangeHassan"
      },
      offers: offer
    };
  }, [activeImage, lang, localizedProductPath, localizedTitle, product, productOfferPrice, seoDescription, t]);

  usePageSeo({
    title: seoTitle,
    description: seoDescription,
    path: localizedProductPath,
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    image: activeImage ? resolveProtectedImageUrl(activeImage) : "",
    type: "product",
    robots: isLegacyProductPath || (!loading && !product) ? "noindex,follow" : "index,follow",
    alternates: catalogAlternates(`/${slug}`),
    jsonLd: productJsonLd,
    jsonLdId: "product-detail-jsonld"
  });

  const terms = product?.terms || [];
  const categories = product?.categories?.length ? product.categories : product?.category ? [product.category] : [];
  const categoriesAdjusted = categories.map((cat) =>
    cat.title_fa === "چینی" || cat.title_fa === "چینی کریستال"
      ? { ...cat, title_fa: "چینی کریستال", slug: "chinese-crystal" }
      : cat
  );
  const categoryLine = categoriesAdjusted.map((cat) => getLocalized(cat, lang)).filter(Boolean).join(" • ");
  const productsPath = `/${lang}/products`;
  const catalogReturnState = readCatalogProductReturnState();
  const stateReturnPath = typeof location.state?.productReturnTo === "string" ? location.state.productReturnTo : "";
  const storedReturnPath = catalogReturnState?.productSlug === slug ? catalogReturnState.path : "";
  const fallbackCategoryPath = categoriesAdjusted[0]?.slug ? `/${lang}/products/${categoriesAdjusted[0].slug}` : productsPath;
  const productBackPath = stateReturnPath || storedReturnPath || (hasLegacyProductsReturnState() ? "/products" : fallbackCategoryPath);
  const restoreProductBack = Boolean(stateReturnPath || storedReturnPath);
  const productBackState = new RegExp(`^/${lang}/products/[^/?#]+`).test(productBackPath)
    ? { catalogRouteKind: "category", restoreProductReturn: restoreProductBack }
    : undefined;

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
    { key: "variants", label: t("productDetail.variants") },
    { key: "mines", label: t("productDetail.mines") },
    { key: "finishes", label: t("productDetail.finishes") },
    { key: "tone", label: t("productDetail.tone") },
    { key: "pattern", label: t("productDetail.pattern") },
    { key: "visual_impact", label: t("productDetail.visualImpact") },
    { key: "use_case_space", label: t("productDetail.useCaseSpaces") },
    { key: "use_case_form", label: t("productDetail.useCaseForms") },
    { key: "use_case_application", label: t("productDetail.useCaseApplications") },
    { key: "use_case_project_type", label: t("productDetail.useCaseProjects") },
    { key: "use_case_special", label: t("productDetail.useCaseSpecial") }
  ];

  const localizedInfoSections = infoSections
    .map((section) => ({
      ...section,
      items: (termsByTaxonomy[section.key] || [])
        .map((term) => ({ term, label: getLocalizedTerm(term, lang) }))
        .filter((item) => item.label)
    }))
    .filter((section) => section.items.length > 0);
  const hasMoreInfo = localizedInfoSections.length > 0;
  const phoneValue = t("footer.phoneValue");
  const phoneItems = [phoneValue, "09121193835", "09121193935"]
    .map((value) => ({ value, normalized: normalizePhone(value) }))
    .filter((item) => item.normalized);
  const siteUrl = String(import.meta.env.VITE_SITE_URL || "").toLowerCase();
  const showDomesticMessengerLinks = siteUrl.includes("sangehassan.ir");
  const hasDetailBlocks = hasMoreInfo;

  const fallbackHotspots = useMemo(() => {
    const options = [];
    const stoneType = terms.find((term) => term?.taxonomy === "stone_type");
    if (stoneType) options.push(getLocalizedTerm(stoneType, lang));

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
  }, [lang, terms]);

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
            <div className={`flex flex-wrap items-center gap-3 ${isRTL ? "justify-end" : "justify-start"}`}>
              <nav aria-label="Breadcrumb" className="flex flex-wrap items-center gap-2 text-xs font-semibold text-primary/55">
                <Link to="/" className="transition hover:text-primary">
                  SangeHassan
                </Link>
                <span>/</span>
                <Link to={productsPath} className="transition hover:text-primary">
                  {t("nav.products")}
                </Link>
                {categoriesAdjusted[0]?.slug ? (
                  <>
                    <span>/</span>
                    <Link to={`/${lang}/products/${categoriesAdjusted[0].slug}`} state={{ catalogRouteKind: "category" }} className="transition hover:text-primary">
                      {getLocalized(categoriesAdjusted[0], lang)}
                    </Link>
                  </>
                ) : null}
                {localizedTitle && (
                  <>
                    <span>/</span>
                    <span className="text-primary/75">{localizedTitle}</span>
                  </>
                )}
              </nav>
              <Link
                to={productBackPath}
                state={productBackState}
                className={`${isRTL ? "mr-auto" : "ml-auto"} rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70 transition hover:border-primary/50`}
              >
                {t("productDetail.backToProducts")}
              </Link>
            </div>

            <div className="relative overflow-hidden bg-primary/10">
              {activeImage ? (
                <button
                  type="button"
                  className="group relative block h-[42vh] min-h-[320px] w-full cursor-zoom-in appearance-none overflow-hidden border-0 bg-transparent p-0 text-start focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/55 md:h-[52vh]"
                  onClick={openLightbox}
                  aria-label={t("productDetail.openGallery")}
                  aria-haspopup="dialog"
                  aria-expanded={lightboxOpen}
                >
                  <ProtectedImage
                    src={resolveProtectedImageUrl(activeImage)}
                    alt={localizedTitle}
                    wrapperClassName="h-full w-full"
                    fit="cover"
                    className="transition duration-500 group-hover:scale-[1.02]"
                    loading="eager"
                    fetchPriority="high"
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
                          src={resolveProtectedThumbnailUrl(image)}
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
              {(categoryLine || product?.is_popular) && (
                <div className={`flex flex-wrap items-center gap-2 ${isRTL ? "justify-end" : "justify-start"}`}>
                  {categoryLine ? (
                    <div className="flex flex-wrap gap-2 text-sm font-semibold uppercase tracking-[0.24em] text-primary/65">
                      {categoriesAdjusted.map((category, index) => (
                        <span key={category.id || category.slug}>
                          {index > 0 ? <span className="px-1">•</span> : null}
                          {category.slug ? (
                            <Link to={`/${lang}/products/${category.slug}`} state={{ catalogRouteKind: "category" }} className="transition hover:text-primary">
                              {getLocalized(category, lang)}
                            </Link>
                          ) : getLocalized(category, lang)}
                        </span>
                      ))}
                    </div>
                  ) : null}
                  {product?.is_popular && (
                    <span className="rounded-full border border-accent/25 bg-accent/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em] text-accent">
                      {t("productDetail.popularBadge")}
                    </span>
                  )}
                </div>
              )}
              <div className="flex flex-wrap items-start gap-3 md:gap-4">
                <h1 className="font-display text-3xl leading-tight text-primary md:text-5xl">{localizedTitle}</h1>
                {phoneItems.length > 0 && (
                  <div className={`flex w-full max-w-[18rem] flex-col items-stretch gap-1.5 ${isRTL ? "mr-auto" : "ml-auto"}`}>
                    {phoneItems.map((item) => (
                      <a
                        key={item.normalized}
                        href={`tel:${item.normalized}`}
                        className="call-cta-shimmer grid h-12 grid-cols-[2.5rem_1fr] items-center gap-3 rounded-full border border-primary/25 px-3.5 text-primary/85 transition hover:border-primary/55 hover:text-primary"
                      >
                        <span className="inline-flex h-10 w-10 items-center justify-center rounded-full">
                          <PhoneIcon className="h-7 w-7" />
                        </span>
                        <span className="min-w-0 truncate text-center text-xs font-semibold tabular-nums md:text-sm" dir="ltr">
                          {item.value}
                        </span>
                      </a>
                    ))}

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
              {productOfferPrice > 0 && (
                <div className={`mt-5 max-w-xl space-y-2 ${isRTL ? "text-right" : "text-left"}`}>
                  <p className="text-xs font-semibold uppercase tracking-[0.26em] text-accent">{t("productDetail.offerLabel")}</p>
                  <div className="space-y-1">
                    <p className="text-xs font-semibold uppercase tracking-[0.16em] text-primary/55">{t("productDetail.offerFloorLabel")}</p>
                    <p className="font-display text-3xl leading-tight text-primary md:text-4xl">
                      {formatOfferPrice(productOfferPrice, lang)}
                    </p>
                  </div>
                  <p className="max-w-lg text-xs leading-6 text-primary/60">{t("productDetail.offerDisclaimer")}</p>
                </div>
              )}
            </div>
          </div>

          <div className="space-y-8 border-t border-primary/20 pt-6">
            {hasDetailBlocks && (
              <section className="space-y-4">
                <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.moreInfo")}</p>
                <div className="space-y-5 text-sm text-primary">
                  {localizedInfoSections.map((section) => (
                    <div key={section.key} className="space-y-2">
                      <p className="font-semibold">{section.label}</p>
                      <div className="flex flex-wrap gap-2 text-[12px] font-semibold text-primary/80">
                        {section.items.map(({ term, label }) => {
                          const termLink = getTermLink(section.key, term);
                          if (!termLink) {
                            return (
                              <span key={`${section.key}-${term.id}`} className="px-2 py-1">
                                {label}
                              </span>
                            );
                          }
                          const linkClass = "px-2 py-1 underline underline-offset-4 transition hover:text-primary";
                          if (/^https?:\/\//i.test(termLink)) {
                            return (
                              <a key={`${section.key}-${term.id}`} href={termLink} className={linkClass}>
                                {label}
                              </a>
                            );
                          }
                          return (
                            <Link key={`${section.key}-${term.id}`} to={termLink} className={linkClass}>
                              {label}
                            </Link>
                          );
                        })}
                      </div>
                    </div>
                  ))}
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

            {relatedProducts.length > 0 && (
              <section className="space-y-4 border-t border-primary/20 pt-6">
                <div className="space-y-1">
                  <p className="text-xs uppercase tracking-[0.3em] text-primary/60">
                    {t("productDetail.relatedProductsTitle")}
                  </p>
                  <p className="text-sm text-primary/60">{t("productDetail.relatedProductsSubtitle")}</p>
                </div>
                <div className="grid gap-3 md:grid-cols-3">
                  {relatedProducts.map((relatedProduct) => {
                    const title = getLocalized(relatedProduct, lang) || relatedProduct.title_en || "";
                    const categoryTitle = relatedProduct.category ? getLocalized(relatedProduct.category, lang) : "";
                    return (
                      <Link
                        key={relatedProduct.id}
                        to={`/${lang}/products/${relatedProduct.slug}`}
                        state={{ catalogRouteKind: "product" }}
                        className="group overflow-hidden border border-primary/15 bg-white/55 transition hover:border-primary/40"
                      >
                        <div className="aspect-[4/3] bg-primary/10">
                          {relatedProduct.image_url ? (
                            <img
                              src={resolveImageUrl(relatedProduct.image_url)}
                              alt={title}
                              className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.03]"
                              loading="lazy"
                            />
                          ) : null}
                        </div>
                        <div className="space-y-1 px-3 py-3">
                          <p className="line-clamp-2 text-sm font-semibold text-primary">{title}</p>
                          {categoryTitle ? (
                            <p className="truncate text-xs text-primary/50">{categoryTitle}</p>
                          ) : null}
                        </div>
                      </Link>
                    );
                  })}
                </div>
              </section>
            )}

            {relatedProjects.length > 0 && (
              <section className="space-y-4 border-t border-primary/20 pt-6">
                <div className="space-y-1">
                  <p className="text-xs uppercase tracking-[0.3em] text-primary/60">
                    {t("productDetail.relatedProjectsTitle")}
                  </p>
                  <p className="text-sm text-primary/60">{t("productDetail.relatedProjectsSubtitle")}</p>
                </div>
                <div className="grid gap-3 md:grid-cols-3">
                  {relatedProjects.map((project) => {
                    const title = getLocalizedProjectTitle(project, lang, t);
                    const description = getLocalizedProjectDescription(project, lang);
                    return (
                      <Link
                        key={project.id}
                        to={`/projects/${project.id}`}
                        className="group overflow-hidden border border-primary/15 bg-white/55 transition hover:border-primary/40"
                      >
                        <div className="aspect-[4/3] bg-primary/10">
                          {project.cover_image_url ? (
                            <img
                              src={resolveImageUrl(project.cover_image_url)}
                              alt={title}
                              className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.03]"
                              loading="lazy"
                            />
                          ) : null}
                        </div>
                        <div className="space-y-1 px-3 py-3">
                          <p className="line-clamp-2 text-sm font-semibold text-primary">{title}</p>
                          {description ? (
                            <p className="line-clamp-2 text-xs text-primary/55">{description}</p>
                          ) : null}
                        </div>
                      </Link>
                    );
                  })}
                </div>
              </section>
            )}

          </div>
        </div>
      )}

      {lightboxOpen && (
        <div className="fixed inset-0 z-[9999] bg-[#E5E1DD] text-primary" role="dialog" aria-modal="true" aria-label={localizedTitle}>
          <div className="section-shell flex h-20 items-center justify-between gap-4">
            <p className="truncate text-xs font-semibold uppercase tracking-[0.22em] text-primary/60">
              {localizedTitle}
            </p>
            <button
              type="button"
              onClick={closeLightbox}
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
              <ProtectedImage
                src={resolveProtectedImageUrl(activeImage)}
                alt={localizedTitle}
                wrapperClassName="h-full w-full"
                fit="contain"
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
