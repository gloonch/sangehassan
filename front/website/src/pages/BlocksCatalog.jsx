import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { gsap } from "gsap";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { useTranslation } from "../lib/i18n";

const BLOCK_GRID_SKELETON_COUNT = 6;
const BLOCK_FILTER_SKELETON_COUNT = 4;

const getLocalized = (item, lang) => {
  if (!item) return "";
  if (lang === "fa") return item.title_fa;
  if (lang === "ar") return item.title_ar;
  return item.title_en;
};

const normalizeSearchText = (value) =>
  String(value || "")
    .normalize("NFKC")
    .replace(/\u200c/g, " ")
    .trim()
    .toLowerCase()
    .replace(/\s+/g, " ");

const BlockFilterSkeletons = () => (
  <>
    {Array.from({ length: BLOCK_FILTER_SKELETON_COUNT }).map((_, index) => (
      <span
        key={`block-filter-skeleton-${index}`}
        className="h-9 w-24 animate-pulse rounded-full border border-primary/10 bg-primary/5"
      />
    ))}
  </>
);

const BlockGridSkeleton = () => (
  <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3" aria-hidden="true">
    {Array.from({ length: BLOCK_GRID_SKELETON_COUNT }).map((_, index) => (
      <div key={`block-skeleton-${index}`} className="overflow-hidden bg-white/45">
        <div className="aspect-square w-full animate-pulse bg-primary/10" />
      </div>
    ))}
  </div>
);

const getStatusLabel = (status, t) => {
  if (status === "available") return t("blocks.statusAvailable");
  if (status === "reserved") return t("blocks.statusReserved");
  if (status === "sold") return t("blocks.statusSold");
  return status || t("messages.empty");
};

const getBlockMetaItems = (block, t) => {
  const items = [];
  if (block.stone_type) items.push(`${t("blocks.stoneType")}: ${block.stone_type}`);
  if (block.dimensions) items.push(`${t("blocks.dimensions")}: ${block.dimensions}`);
  if (block.weight_ton) items.push(`${t("blocks.weightTon")}: ${block.weight_ton}`);
  if (block.quarry) items.push(`${t("blocks.quarry")}: ${block.quarry}`);
  return items;
};

export default function BlocksCatalog({ embedded = false } = {}) {
  const { t, lang } = useTranslation();
  const pageRef = useRef(null);
  const hasEntranceAnimatedRef = useRef(false);
  const [blocks, setBlocks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filterType, setFilterType] = useState("all");
  const [filterStatus, setFilterStatus] = useState("all");
  const [searchInput, setSearchInput] = useState("");

  useEffect(() => {
    let mounted = true;

    const loadBlocks = async () => {
      setLoading(true);
      try {
        const res = await fetchJSON("/api/blocks");
        if (!mounted) return;
        setBlocks(Array.isArray(res.data) ? res.data : []);
      } catch (_) {
        if (!mounted) return;
        setBlocks([]);
      } finally {
        if (mounted) setLoading(false);
      }
    };

    loadBlocks();
    return () => {
      mounted = false;
    };
  }, []);

  const stoneTypes = useMemo(() => {
    const set = new Set();
    blocks.forEach((block) => {
      const value = String(block.stone_type || "").trim();
      if (value) set.add(value);
    });
    return Array.from(set).sort((a, b) =>
      a.localeCompare(b, lang === "fa" ? "fa" : lang === "ar" ? "ar" : "en", { sensitivity: "base" })
    );
  }, [blocks, lang]);

  const statuses = useMemo(() => {
    const preferred = ["available", "reserved", "sold"];
    const present = new Set(blocks.map((block) => block.status).filter(Boolean));
    return preferred.filter((status) => present.has(status));
  }, [blocks]);

  const filtered = useMemo(() => {
    const query = normalizeSearchText(searchInput);

    return blocks.filter((block) => {
      const typeMatch = filterType === "all" || block.stone_type === filterType;
      const statusMatch = filterStatus === "all" || block.status === filterStatus;
      if (!typeMatch || !statusMatch) return false;
      if (!query) return true;

      const haystack = normalizeSearchText(
        [
          getLocalized(block, lang),
          block.title_en,
          block.title_fa,
          block.title_ar,
          block.slug,
          block.stone_type,
          block.quarry,
          block.dimensions,
          block.description,
          getStatusLabel(block.status, t),
          block.weight_ton
        ].join(" ")
      );
      return haystack.includes(query);
    });
  }, [blocks, filterType, filterStatus, lang, searchInput, t]);

  useEffect(() => {
    if (filterType !== "all" && !stoneTypes.includes(filterType)) {
      setFilterType("all");
    }
    if (filterStatus !== "all" && !statuses.includes(filterStatus)) {
      setFilterStatus("all");
    }
  }, [filterType, filterStatus, stoneTypes, statuses]);

  useEffect(() => {
    if (loading) return;
    const page = pageRef.current;
    if (!page || typeof window === "undefined" || !window.matchMedia) return;
    if (hasEntranceAnimatedRef.current) return;
    hasEntranceAnimatedRef.current = true;

    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const leadItems = page.querySelectorAll("[data-blocks-anim='lead']");
    const cardItems = Array.from(page.querySelectorAll("[data-blocks-anim='card']")).slice(0, reduceMotion ? 4 : 8);

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

      if (cardItems.length) {
        gsap.fromTo(
          cardItems,
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

  const sectionClass = embedded ? "section-shell pt-16 pb-16" : "section-shell pt-16 pb-12";

  const content = (
    <section ref={pageRef} className={sectionClass} aria-busy={loading ? "true" : "false"}>
      <div className="mb-8 flex flex-col gap-4">
        <p data-blocks-anim="lead" className="text-sm uppercase tracking-[0.3em] text-primary/60">
          {t("blocks.title")}
        </p>
        <h1 data-blocks-anim="lead" className="font-display text-3xl md:text-4xl">
          {t("blocks.subtitle")}
        </h1>
      </div>

      <div className="mb-8 space-y-3">
        <div className="flex flex-wrap items-center gap-3">
          <button
            type="button"
            onClick={() => setFilterType("all")}
            data-blocks-anim="lead"
            className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${filterType === "all"
              ? "border-primary bg-primary text-sand"
              : "border-primary/20 text-primary/70 hover:border-primary/50"
              }`}
          >
            {t("blocks.filterAll")}
          </button>
          {loading ? <BlockFilterSkeletons /> : stoneTypes.map((stoneType) => (
            <button
              key={stoneType}
              type="button"
              onClick={() => setFilterType(stoneType)}
              data-blocks-anim="lead"
              className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${filterType === stoneType
                ? "border-primary bg-primary text-sand"
                : "border-primary/20 text-primary/70 hover:border-primary/50"
                }`}
            >
              {stoneType}
            </button>
          ))}
        </div>

        <label className="sr-only" htmlFor="block-search">
          {t("blocks.searchLabel")}
        </label>
        <div className="relative">
          <input
            id="block-search"
            type="search"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            data-blocks-anim="lead"
            placeholder={t("blocks.searchPlaceholder")}
            className="w-full rounded-full border border-primary/20 bg-white/70 px-4 py-2.5 pr-9 text-sm font-semibold text-primary outline-none transition focus:border-primary/60"
          />
          {loading && (
            <span
              aria-label={t("messages.loading")}
              className="absolute right-3 top-1/2 h-3 w-3 -translate-y-1/2 animate-spin rounded-full border border-primary/40 border-t-transparent"
            />
          )}
        </div>
      </div>

      {loading ? (
        <BlockGridSkeleton />
      ) : filtered.length === 0 ? (
        <div className="mx-auto mt-10 flex min-h-[180px] w-full items-center justify-center text-center">
          <p className="text-sm text-primary/70">{t("blocks.empty")}</p>
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filtered.map((block) => {
            const isRTL = lang === "fa" || lang === "ar";
            const gradientDir = isRTL ? "bg-gradient-to-tl" : "bg-gradient-to-tr";
            const title = getLocalized(block, lang) || block.title_en || "";
            const metaItems = getBlockMetaItems(block, t);

            return (
              <Link
                to={`/blocks/${block.slug}`}
                key={block.id}
                data-blocks-anim="card"
                className="group flex h-full flex-col overflow-hidden transition hover:-translate-y-1 hover:shadow-xl"
                style={{ contentVisibility: "auto", containIntrinsicSize: "360px" }}
              >
                <div className="relative aspect-square w-full overflow-hidden bg-primary/10">

                  {block.image_url ? (
                    <img
                      src={resolveImageUrl(block.image_url)}
                      alt={title}
                      loading="lazy"
                      decoding="async"
                      className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
                    />
                  ) : (
                    <div className="flex h-full items-center justify-center text-sm text-primary/60">
                      {t("productDetail.noImages")}
                    </div>
                  )}

                  <div className={`pointer-events-none absolute inset-x-0 bottom-0 ${gradientDir} from-black/70 via-black/25 to-transparent p-4`}>
                    <h3 className="font-display text-xl leading-tight text-white drop-shadow-[0_10px_24px_rgba(0,0,0,0.55)]">
                      {title}
                    </h3>
                    {metaItems.length > 0 && (
                      <div className="mt-2 space-y-1 text-[11px] font-semibold text-white/85 drop-shadow-[0_8px_18px_rgba(0,0,0,0.45)]">
                        {metaItems.slice(0, 3).map((item) => (
                          <p key={item} className="truncate">{item}</p>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </section>
  );

  if (embedded) {
    return content;
  }

  return (
    <div className="bg-quiet-grid">
      {content}
    </div>
  );
}
