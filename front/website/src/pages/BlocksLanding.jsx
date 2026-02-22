import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { useTranslation } from "../lib/i18n";

const pickField = (section, field, lang) => {
  if (!section) return "";
  const key = `${field}_${lang}`;
  return section[key] || section[`${field}_en`] || "";
};

const splitLines = (text) =>
  text
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);

export default function BlocksLanding() {
  const { t, lang } = useTranslation();
  const [sections, setSections] = useState([]);
  const [blocks, setBlocks] = useState([]);
  const [activeSlide, setActiveSlide] = useState(0);

  useEffect(() => {
    fetchJSON("/api/content-sections?page=blocks")
      .then((res) => setSections(res.data || []))
      .catch(() => setSections([]));
  }, []);

  useEffect(() => {
    fetchJSON("/api/blocks?featured=true")
      .then((res) => setBlocks(res.data || []))
      .catch(() => setBlocks([]));
  }, []);

  const heroSection = useMemo(() => sections.find((section) => section.key === "hero") || {}, [sections]);
  const title = pickField(heroSection, "title", lang) || t("blocks.title");
  const subtitle = pickField(heroSection, "subtitle", lang) || t("blocks.subtitle");
  const description = pickField(heroSection, "description", lang);
  const ctaLabel = pickField(heroSection, "cta_label", lang) || t("blocks.cta");
  const ctaHref = heroSection.cta_href || "/blocks/catalog";
  const lines = splitLines(description || "");
  const heroImages = useMemo(
    () => (Array.isArray(heroSection.images) ? heroSection.images.filter(Boolean) : []),
    [heroSection]
  );

  useEffect(() => {
    setActiveSlide(0);
  }, [heroImages.length]);

  useEffect(() => {
    heroImages.forEach((image) => {
      const preloaded = new Image();
      preloaded.src = resolveImageUrl(image);
    });
  }, [heroImages]);

  useEffect(() => {
    if (heroImages.length <= 1) return undefined;
    const interval = setInterval(() => {
      setActiveSlide((prev) => (prev + 1) % heroImages.length);
    }, 7000);
    return () => clearInterval(interval);
  }, [heroImages.length]);

  return (
    <div className="bg-quiet-grid">
      <section className="relative h-[60vh] min-h-[480px] max-h-[720px] w-full overflow-hidden">
        <div className="absolute inset-0">
          {heroImages.length > 0 ? (
            heroImages.map((image, index) => (
              <img
                key={`${image}-${index}`}
                src={resolveImageUrl(image)}
                alt=""
                className={`absolute inset-0 h-full w-full object-cover transition-all duration-[2200ms] ease-in-out ${
                  index === activeSlide
                    ? "translate-y-0 scale-105 opacity-100"
                    : "-translate-y-2 scale-110 opacity-0"
                }`}
              />
            ))
          ) : (
            <div className="absolute inset-0 bg-primary/30" />
          )}
        </div>

        <div className="absolute inset-0 bg-primary/28" />
        <div className="absolute inset-0 bg-gradient-to-b from-primary/48 via-primary/34 to-primary/54" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_18%_15%,rgba(165,141,102,0.22),transparent_44%)]" />
        <div className="pointer-events-none absolute inset-x-0 bottom-0 h-[18%] bg-gradient-to-t from-sand via-sand/80 to-transparent" />

        <div className="relative z-10 h-full w-full">
          <div className="section-shell flex h-full items-center">
            <div className="max-w-4xl space-y-7 py-8 text-sand md:space-y-8">
              <p className="text-xs uppercase tracking-[0.34em] text-sand/75 md:text-sm">{t("blocks.title")}</p>
              <h1 className="font-display text-4xl leading-tight sm:text-5xl lg:text-6xl">
                {title}
              </h1>
              <p className="max-w-3xl text-lg text-sand/92 md:text-2xl">{subtitle}</p>
              {lines.length > 0 && (
                <ul className="max-w-3xl space-y-2.5 text-sm text-sand/90 md:text-base lg:text-lg">
                  {lines.map((line) => (
                    <li key={line} className="flex items-start gap-2.5">
                      <span className="mt-2 h-1.5 w-1.5 rounded-full bg-accent/95" />
                      <span>{line}</span>
                    </li>
                  ))}
                </ul>
              )}
              <div className="flex flex-wrap gap-3">
                <Link
                  to={ctaHref}
                  className="rounded-full bg-sand/95 px-7 py-3 text-sm font-semibold text-primary transition hover:bg-sand md:text-base"
                >
                  {ctaLabel}
                </Link>
              </div>
              <p className="text-[11px] uppercase tracking-[0.28em] text-sand/72 md:text-xs">{t("hero.scrollHint")}</p>
            </div>
          </div>
        </div>
      </section>

      <section className="section-shell pb-16 pt-12">
        <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.4em] text-primary/50">{t("blocks.catalogTitle")}</p>
            <h2 className="mt-2 font-display text-2xl text-primary">{t("blocks.catalogSubtitle")}</h2>
          </div>
          <Link to="/blocks/catalog" className="text-sm font-semibold text-accent">
            {t("blocks.cta")}
          </Link>
        </div>

        {blocks.length === 0 ? (
          <p className="text-sm text-primary/70">{t("blocks.empty")}</p>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {blocks.slice(0, 6).map((block) => (
              <Link
                to={`/blocks/${block.slug}`}
                key={block.id}
                className="group relative overflow-hidden rounded-3xl border border-primary/10 bg-white/70 shadow-[0_18px_40px_rgba(8,58,79,0.12)]"
              >
                <div className="h-44 overflow-hidden">
                  {block.image_url ? (
                    <img
                      src={resolveImageUrl(block.image_url)}
                      alt=""
                      className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
                    />
                  ) : (
                    <div className="flex h-full items-center justify-center text-xs uppercase tracking-[0.3em] text-primary/40">
                      {t("blocks.title")}
                    </div>
                  )}
                </div>
                <div className="p-5">
                  <p className="text-sm font-semibold text-primary">{block.title_fa || block.title_en}</p>
                  <p className="mt-1 text-xs text-primary/60">
                    {block.stone_type ? `${t("blocks.stoneType")}: ${block.stone_type}` : t("blocks.stoneType")}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
