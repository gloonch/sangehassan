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
  const [activeIndex, setActiveIndex] = useState(0);

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

  const heroSection = useMemo(() => sections.find((s) => s.key === "hero") || {}, [sections]);
  const heroImages = heroSection.images || [];

  useEffect(() => {
    if (heroImages.length <= 1) return;
    const interval = setInterval(() => {
      setActiveIndex((prev) => (prev + 1) % heroImages.length);
    }, 4500);
    return () => clearInterval(interval);
  }, [heroImages.length]);

  const title = pickField(heroSection, "title", lang) || t("blocks.title");
  const subtitle = pickField(heroSection, "subtitle", lang) || t("blocks.subtitle");
  const description = pickField(heroSection, "description", lang);
  const ctaLabel = pickField(heroSection, "cta_label", lang) || t("blocks.cta");
  const ctaHref = heroSection.cta_href || "/blocks/catalog";
  const lines = splitLines(description || "");

  return (
    <div className="bg-quiet-grid">
      <section className="section-shell py-12">
        <div className="grid gap-8 lg:grid-cols-[1.05fr_0.95fr]">
          <div className="space-y-5">
            <p className="text-xs uppercase tracking-[0.4em] text-primary/50">{t("blocks.title")}</p>
            <h1 className="font-display text-3xl text-primary md:text-4xl">{title}</h1>
            <p className="text-sm text-primary/70">{subtitle}</p>

            {lines.length > 0 && (
              <ul className="space-y-2 text-sm text-primary/80">
                {lines.map((line) => (
                  <li key={line} className="flex items-start gap-2">
                    <span className="mt-2 h-1.5 w-1.5 rounded-full bg-accent" />
                    <span>{line}</span>
                  </li>
                ))}
              </ul>
            )}

            <div className="flex flex-wrap gap-3 pt-2">
              <Link
                to={ctaHref}
                className="rounded-full bg-primary px-6 py-3 text-xs font-semibold uppercase tracking-[0.3em] text-sand"
              >
                {ctaLabel}
              </Link>
              <Link
                to="/contact"
                className="rounded-full border border-primary/20 px-6 py-3 text-xs font-semibold uppercase tracking-[0.3em] text-primary/80"
              >
                {t("contact.title")}
              </Link>
            </div>
          </div>

          <div className="relative overflow-hidden rounded-3xl border border-primary/10 bg-white/60 shadow-[0_20px_50px_rgba(8,58,79,0.12)]">
            <div className="absolute inset-0">
              <div className="absolute -right-16 -top-20 h-48 w-48 rounded-full bg-accent/20 blur-3xl" />
              <div className="absolute -bottom-16 -left-20 h-56 w-56 rounded-full bg-primary/10 blur-3xl" />
            </div>
            <div className="relative h-[360px] sm:h-[420px]">
              {heroImages.length ? (
                <img
                  key={heroImages[activeIndex]}
                  src={resolveImageUrl(heroImages[activeIndex])}
                  alt=""
                  className="h-full w-full object-cover transition duration-700"
                />
              ) : (
                <div className="flex h-full w-full items-center justify-center text-xs uppercase tracking-[0.3em] text-primary/50">
                  {t("messages.loading")}
                </div>
              )}
            </div>
            {heroImages.length > 1 && (
              <div className="absolute bottom-4 left-1/2 flex -translate-x-1/2 items-center gap-2 rounded-full bg-white/80 px-3 py-2 shadow">
                {heroImages.map((_, index) => (
                  <button
                    key={index}
                    type="button"
                    onClick={() => setActiveIndex(index)}
                    className={`h-2.5 w-2.5 rounded-full ${
                      activeIndex === index ? "bg-primary" : "bg-primary/20"
                    }`}
                  />
                ))}
              </div>
            )}
          </div>
        </div>
      </section>

      <section className="section-shell pb-16">
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
