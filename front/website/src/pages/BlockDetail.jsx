import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { usePageSeo } from "../lib/seo";
import { usePrerenderData } from "../lib/prerenderData";
import NotFound from "./NotFound";

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

const getStatusLabel = (status, t) => {
  if (status === "available") return t("blocks.statusAvailable");
  if (status === "reserved") return t("blocks.statusReserved");
  if (status === "sold") return t("blocks.statusSold");
  return status || t("messages.empty");
};

export default function BlockDetail() {
  const { slug } = useParams();
  const { t, lang } = useTranslation();
  const isRTL = lang === "fa" || lang === "ar";
  const prerenderedBlock = usePrerenderData("block");
  const initialBlock = prerenderedBlock?.slug === slug ? prerenderedBlock : null;
  const [block, setBlock] = useState(initialBlock);
  const [loading, setLoading] = useState(!initialBlock);
  const [activeIndex, setActiveIndex] = useState(0);
  const [lightboxOpen, setLightboxOpen] = useState(false);

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      const hasInitialBlock = initialBlock?.slug === slug;
      if (!hasInitialBlock) {
        setLoading(true);
      }
      try {
        const res = await fetchJSON(`/api/blocks/${slug}`);
        if (!mounted) return;
        setBlock(res.data || null);
      } catch (_) {
        if (!mounted) return;
        if (!hasInitialBlock) {
          setBlock(null);
        }
      } finally {
        if (mounted) setLoading(false);
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, [initialBlock, slug]);

  const images = useMemo(() => {
    if (!block) return [];
    if (block.images?.length) return block.images;
    if (block.image_url) return [block.image_url];
    return [];
  }, [block]);

  useEffect(() => {
    if (activeIndex >= images.length) {
      setActiveIndex(0);
    }
  }, [activeIndex, images.length]);

  useEffect(() => {
    setActiveIndex(0);
  }, [slug]);

  const activeImage = images[activeIndex] || images[0] || "";
  const localizedTitle = getLocalized(block, lang) || block?.title_en || "";
  const headlineMeta = [block?.stone_type, block?.quarry].filter(Boolean).join(" • ") || t("blocks.title");
  const seoDescription =
    [
      block?.stone_type ? `${t("blocks.stoneType")}: ${block.stone_type}` : "",
      block?.quarry ? `${t("blocks.quarry")}: ${block.quarry}` : "",
      block?.dimensions ? `${t("blocks.dimensions")}: ${block.dimensions}` : "",
      block?.weight_ton ? `${t("blocks.weightTon")}: ${block.weight_ton}` : "",
      block?.description || ""
    ]
      .filter(Boolean)
      .join(" | ")
      .slice(0, 170) || "Natural stone block detail from SangeHassan with quarry, dimensions, weight, and availability information.";

  usePageSeo({
    title: localizedTitle ? `${localizedTitle} | ${t("blocks.title")} | SangeHassan` : `${t("blocks.title")} | SangeHassan`,
    description: seoDescription,
    path: `/blocks/${slug}`,
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    image: activeImage ? resolveImageUrl(activeImage) : "",
    type: "product",
    robots: !loading && !block ? "noindex,follow" : "index,follow"
  });

  const detailRows = useMemo(() => {
    if (!block) return [];
    return [
      { key: "status", label: t("blocks.filterStatus"), value: getStatusLabel(block.status, t) },
      { key: "stone_type", label: t("blocks.stoneType"), value: block.stone_type },
      { key: "quarry", label: t("blocks.quarry"), value: block.quarry },
      { key: "dimensions", label: t("blocks.dimensions"), value: block.dimensions },
      { key: "weight_ton", label: t("blocks.weightTon"), value: block.weight_ton ? String(block.weight_ton) : "" }
    ].filter((row) => row.value);
  }, [block, t]);

  const phoneItems = [t("footer.phoneValue"), "09121193835", "09121193935"]
    .map((value) => ({ value, normalized: normalizePhone(value) }))
    .filter((item) => item.normalized);
  const siteUrl = String(import.meta.env.VITE_SITE_URL || "").toLowerCase();
  const showDomesticMessengerLinks = siteUrl.includes("sangehassan.ir");

  const goPrev = () => {
    if (images.length <= 1) return;
    setActiveIndex((idx) => (idx - 1 + images.length) % images.length);
  };

  const goNext = () => {
    if (images.length <= 1) return;
    setActiveIndex((idx) => (idx + 1) % images.length);
  };

  if (!loading && !block) return <NotFound />;

  return (
    <section className="section-shell pt-4 pb-12">
      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : (
        <div className="space-y-8">
          <div className="space-y-4">
            <div className={`flex flex-wrap items-center gap-3 ${isRTL ? "justify-end" : "justify-start"}`}>
              <nav aria-label="Breadcrumb" className="flex flex-wrap items-center gap-2 text-xs font-semibold text-primary/55">
                <Link to="/" className="transition hover:text-primary">
                  SangeHassan
                </Link>
                <span>/</span>
                <Link to="/blocks" className="transition hover:text-primary">
                  {t("nav.blocks")}
                </Link>
                {localizedTitle && (
                  <>
                    <span>/</span>
                    <span className="text-primary/75">{localizedTitle}</span>
                  </>
                )}
              </nav>
              <Link
                to="/blocks"
                className={`${isRTL ? "mr-auto" : "ml-auto"} rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70 transition hover:border-primary/50`}
              >
                {t("blocks.backToBlocks")}
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
                        aria-label={`block-thumb-${index + 1}`}
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
              <div className={`flex flex-wrap items-center gap-2 ${isRTL ? "justify-end" : "justify-start"}`}>
                <p className="text-sm font-semibold uppercase tracking-[0.24em] text-primary/65">
                  {headlineMeta}
                </p>
                {block.status && (
                  <span className="rounded-full border border-accent/25 bg-accent/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em] text-accent">
                    {getStatusLabel(block.status, t)}
                  </span>
                )}
              </div>
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
            </div>
          </div>

          <div className="space-y-8 border-t border-primary/20 pt-6">
            {detailRows.length > 0 && (
              <section className="space-y-4">
                <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("blocks.detailsTitle")}</p>
                <div className="space-y-5 text-sm text-primary">
                  {detailRows.map((row) => (
                    <div key={row.key} className="space-y-2">
                      <p className="font-semibold">{row.label}</p>
                      <div className="flex flex-wrap gap-2">
                        <span className="inline-flex items-center gap-1.5 rounded-full border border-primary/20 px-3 py-1.5 text-[12px] font-semibold text-primary/80">
                          <span className="h-1.5 w-1.5 rounded-full bg-primary/70" />
                          <span>{row.value}</span>
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </section>
            )}

            <section className="space-y-4 border-t border-primary/20 pt-6">
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("productDetail.description")}</p>
              {block.description ? (
                <p className="whitespace-pre-line text-sm text-primary/75">{block.description}</p>
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
