import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

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
  const [loading, setLoading] = useState(true);

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
      })
      .finally(() => {
        if (!mounted) return;
        setLoading(false);
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
    <div className="bg-quiet-grid">
      <section className="relative">
        <div className="section-shell flex h-[calc(100vh-96px)] min-h-[520px] flex-col justify-between overflow-y-hidden py-8">
          <header className="max-w-2xl">
            <p className="text-xs uppercase tracking-[0.4em] text-primary/50">{t("landing.title")}</p>
            <h1 className="mt-3 font-display text-3xl text-primary md:text-4xl">
              {t("landing.subtitle")}
            </h1>
          </header>

          <div className="mt-6 flex-1">
            <div className="flex h-full gap-6 overflow-x-auto pb-4 md:grid md:grid-cols-2 md:overflow-visible md:pb-0">
              {[blocksSection, finishedSection].map((section, index) => {
                const title = pickField(section, "title", lang);
                const subtitle = pickField(section, "subtitle", lang);
                const description = pickField(section, "description", lang);
                const ctaLabel = pickField(section, "cta_label", lang);
                const fallbackCTA = section.key === "blocks" ? t("blocks.cta") : t("hero.ctaPrimary");
                const ctaHref = section.cta_href || "#";
                const lines = splitLines(description);
                const image = section.images?.[0];

                return (
                  <article
                    key={section.key || index}
                    className="relative flex min-w-full flex-col justify-between overflow-hidden rounded-3xl border border-primary/10 bg-white/70 p-6 shadow-[0_28px_60px_rgba(8,58,79,0.12)] backdrop-blur md:min-w-0"
                  >
                    <div className="absolute inset-0 opacity-70">
                      <div className="absolute -right-20 -top-24 h-52 w-52 rounded-full bg-accent/20 blur-3xl" />
                      <div className="absolute -bottom-16 -left-20 h-60 w-60 rounded-full bg-primary/10 blur-3xl" />
                    </div>

                    <div className="relative z-10">
                      <p className="text-xs uppercase tracking-[0.3em] text-primary/60">
                        {section.key === "blocks" ? t("blocks.title") : t("products.title")}
                      </p>
                      <h2 className="mt-2 font-display text-2xl text-primary md:text-3xl">{title}</h2>
                      <p className="mt-3 text-sm text-primary/70">{subtitle}</p>

                      {lines.length > 0 && (
                        <ul className="mt-4 space-y-2 text-sm text-primary/80">
                          {lines.map((line) => (
                            <li key={line} className="flex items-start gap-2">
                              <span className="mt-2 h-1.5 w-1.5 rounded-full bg-accent" />
                              <span>{line}</span>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>

                    <div className="relative z-10 mt-6 flex flex-wrap items-center justify-between gap-4">
                      <Link
                        to={ctaHref}
                        className="rounded-full bg-primary px-6 py-2 text-xs font-semibold uppercase tracking-[0.3em] text-sand"
                      >
                        {ctaLabel || fallbackCTA}
                      </Link>
                      <div className="h-24 w-32 overflow-hidden rounded-2xl border border-primary/10 bg-primary/5">
                        {image ? (
                          <img
                            src={resolveImageUrl(image)}
                            alt=""
                            className="h-full w-full object-cover"
                          />
                        ) : (
                          <div className="flex h-full w-full items-center justify-center text-[10px] uppercase tracking-[0.3em] text-primary/40">
                            {loading ? "..." : "image"}
                          </div>
                        )}
                      </div>
                    </div>
                  </article>
                );
              })}
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
