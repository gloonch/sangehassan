import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import blocksOverlayImage from "@shared/assets/landing_page/landingpage_blocks_overlay.png";
import finishesOverlayVideo from "@shared/assets/landing_page/landingpage_finishes_overlay.mp4";

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

export default function Home() {
  const { t, lang } = useTranslation();
  const [sections, setSections] = useState([]);

  useEffect(() => {
    let mounted = true;
    fetchJSON("/api/content-sections?page=home")
      .then((res) => {
        if (!mounted) return;
        setSections(res.data || []);
      })
      .catch(() => {
        if (!mounted) return;
        setSections([]);
      });
    return () => {
      mounted = false;
    };
  }, []);

  const sectionsByKey = useMemo(() => {
    const map = {};
    sections.forEach((section) => {
      map[section.key] = section;
    });
    return map;
  }, [sections]);

  const fallbackSections = useMemo(
    () => ({
      blocks: {
        key: "blocks",
        title_en: t("blocks.title"),
        title_fa: t("blocks.title"),
        title_ar: t("blocks.title"),
        subtitle_en: t("blocks.subtitle"),
        subtitle_fa: t("blocks.subtitle"),
        subtitle_ar: t("blocks.subtitle"),
        description_en: t("blocks.subtitle"),
        description_fa: t("blocks.subtitle"),
        description_ar: t("blocks.subtitle"),
        cta_label_en: t("blocks.cta"),
        cta_label_fa: t("blocks.cta"),
        cta_label_ar: t("blocks.cta"),
        cta_href: "/blocks",
        images: []
      },
      finished: {
        key: "finished",
        title_en: t("products.title"),
        title_fa: t("products.title"),
        title_ar: t("products.title"),
        subtitle_en: t("products.subtitle"),
        subtitle_fa: t("products.subtitle"),
        subtitle_ar: t("products.subtitle"),
        description_en: t("products.subtitle"),
        description_fa: t("products.subtitle"),
        description_ar: t("products.subtitle"),
        cta_label_en: t("hero.ctaPrimary"),
        cta_label_fa: t("hero.ctaPrimary"),
        cta_label_ar: t("hero.ctaPrimary"),
        cta_href: "/products/overview",
        images: []
      }
    }),
    [t]
  );

  const blocksSection = sectionsByKey.blocks || fallbackSections.blocks;
  const finishedSection = sectionsByKey.finished || fallbackSections.finished;

  return (
    <div className="relative h-[100dvh] w-full overflow-hidden">
      <section className="relative flex h-full w-full flex-col overflow-hidden lg:flex-row">
        {[finishedSection, blocksSection].map((section, index) => {
          const isBlocks = section.key === "blocks";
          const title = pickField(section, "title", lang) || (isBlocks ? t("blocks.title") : t("products.title"));
          const subtitle = pickField(section, "subtitle", lang);
          const description = pickField(section, "description", lang);
          const ctaLabel = pickField(section, "cta_label", lang);
          const fallbackCTA = isBlocks ? t("blocks.cta") : t("hero.ctaPrimary");
          const ctaHref = section.cta_href || (isBlocks ? "/blocks" : "/products/overview");
          const lines = splitLines(description);

          return (
            <Link
              key={section.key || index}
              to={ctaHref}
              className="group relative block h-1/2 w-full overflow-hidden lg:h-full lg:w-1/2"
              aria-label={title}
            >
              {isBlocks ? (
                <img
                  src={blocksOverlayImage}
                  alt=""
                  className="absolute inset-0 h-full w-full object-cover object-center"
                />
              ) : (
                <video
                  className="absolute inset-0 h-full w-full object-cover object-center"
                  autoPlay
                  muted
                  loop
                  playsInline
                  preload="auto"
                >
                  <source src={finishesOverlayVideo} type="video/mp4" />
                </video>
              )}

              <div className="absolute inset-0 bg-primary/25" />
              <div
                className={`absolute inset-0 ${isBlocks
                  ? "bg-gradient-to-br from-primary/35 via-primary/15 to-primary/48"
                  : "bg-gradient-to-br from-primary/40 via-primary/20 to-primary/52"
                  }`}
              />
              <div className="absolute inset-0 bg-[radial-gradient(circle_at_25%_20%,rgba(165,141,102,0.16),transparent_48%)]" />

              <div className="relative z-10 flex h-full items-center justify-center px-6 py-10 text-sand sm:px-10 lg:px-12">
                <div className="mx-auto w-full max-w-[38rem] text-center">
                  <p className="text-[10px] uppercase tracking-[0.34em] text-sand/72 sm:text-xs">
                    {isBlocks ? t("blocks.title") : t("products.title")}
                  </p>
                  <h1 className="mt-2 font-display text-xl leading-tight sm:text-2xl md:text-3xl lg:text-4xl">
                    {title}
                  </h1>
                  {subtitle ? <p className="mx-auto mt-3 max-w-[33rem] text-[11px] text-sand/88 sm:text-xs md:text-sm lg:text-base">{subtitle}</p> : null}

                  {lines.length > 0 && (
                    <ul className="mx-auto mt-4 max-w-[32rem] space-y-1.5 text-left text-[10px] text-sand/88 sm:text-[11px] md:text-xs lg:text-sm">
                      {lines.slice(0, 3).map((line) => (
                        <li key={line} className="flex items-start gap-2">
                          <span className="mt-1.5 h-1.5 w-1.5 rounded-full bg-accent/85" />
                          <span>{line}</span>
                        </li>
                      ))}
                    </ul>
                  )}

                  <div className="mt-7 flex items-center justify-center gap-3">
                    <span className="rounded-full border border-sand/40 bg-sand/10 px-4 py-2 text-[10px] font-semibold uppercase tracking-[0.24em] text-sand sm:px-5 sm:py-2.5 sm:text-xs">
                      {ctaLabel || fallbackCTA}
                    </span>
                    <span className="text-xl text-sand/75 transition-transform duration-500 group-hover:translate-x-1">
                      →
                    </span>
                  </div>
                </div>
              </div>

              <span className="pointer-events-none absolute inset-0 ring-0 ring-inset transition group-hover:ring-1 group-hover:ring-white/25" />
            </Link>
          );
        })}
        <span className="pointer-events-none absolute inset-x-0 top-1/2 z-20 h-20 -translate-y-1/2 bg-gradient-to-b from-transparent via-primary/10 to-transparent backdrop-blur-sm lg:hidden" />
        <span className="pointer-events-none absolute inset-y-0 left-1/2 z-20 hidden w-28 -translate-x-1/2 bg-gradient-to-r from-transparent via-primary/10 to-transparent backdrop-blur-sm lg:block" />
      </section>
    </div>
  );
}
