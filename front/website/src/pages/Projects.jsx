import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { withIranAccessSeoNotice } from "../lib/seo";

const projectsSeoContent = {
  fa: {
    title: "پروژه‌ها | سنگ حسن",
    description:
      "نمونه پروژه‌های اجراشده سنگ حسن در نما و طراحی داخلی؛ شامل خروجی واقعی، کیفیت اجرا و جزئیات کاربرد سنگ.",
    locale: "fa_IR"
  },
  en: {
    title: "Projects | SangeHassan",
    description:
      "Explore SangeHassan's completed stone projects across facade and interior applications with real-world execution results.",
    locale: "en_US"
  },
  ar: {
    title: "المشاريع | سانج حسن",
    description:
      "اطّلع على مشاريع سانج حسن المنجزة في الواجهات والتصميم الداخلي مع نتائج تنفيذ واقعية.",
    locale: "ar_SA"
  }
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

export default function Projects() {
  const { t, lang } = useTranslation();
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;

    const loadDetails = async (cards) => {
      const details = await Promise.allSettled(
        cards.map((project) => fetchJSON(`/api/projects/${project.id}`))
      );
      if (!active) return;

      const enriched = cards.map((project, index) => {
        const detail = details[index];
        if (detail.status !== "fulfilled" || !detail.value?.data) return project;
        return {
          ...project,
          ...detail.value.data,
          cover_image_url: detail.value.data.cover_image_url || project.cover_image_url
        };
      });

      setProjects(enriched);
    };

    const load = async () => {
      try {
        const response = await fetchJSON("/api/projects");
        if (!active) return;
        const cards = Array.isArray(response.data) ? response.data : [];
        setProjects(cards);
        setLoading(false);
        if (cards.length > 0) {
          loadDetails(cards).catch(() => { });
        }
      } catch (_) {
        if (!active) return;
        setProjects([]);
        setLoading(false);
      }
    };

    load();
    return () => {
      active = false;
    };
  }, []);

  const directionClass = useMemo(() => {
    const isRTL = lang === "fa" || lang === "ar";
    return isRTL ? "bg-gradient-to-tl" : "bg-gradient-to-tr";
  }, [lang]);

  useEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return;

    const seo = withIranAccessSeoNotice(projectsSeoContent[lang] || projectsSeoContent.fa);
    const pageUrl = `${window.location.origin}/projects`;
    const previousTitle = document.title;
    const cleanups = [];

    const firstCover = projects.find((project) => project?.cover_image_url)?.cover_image_url;
    const ogImage = firstCover ? resolveImageUrl(firstCover) : "";

    const upsertMeta = (selector, createAttrs, value) => {
      if (!value) return;

      let el = document.head.querySelector(selector);
      const created = !el;

      if (!el) {
        el = document.createElement("meta");
        Object.entries(createAttrs).forEach(([key, attrValue]) => {
          el.setAttribute(key, attrValue);
        });
        document.head.appendChild(el);
      }

      const prevContent = el.getAttribute("content");
      el.setAttribute("content", value);

      cleanups.push(() => {
        if (created) {
          el.remove();
          return;
        }
        if (prevContent === null) {
          el.removeAttribute("content");
          return;
        }
        el.setAttribute("content", prevContent);
      });
    };

    const upsertCanonical = (href) => {
      let el = document.head.querySelector('link[rel="canonical"]');
      const created = !el;

      if (!el) {
        el = document.createElement("link");
        el.setAttribute("rel", "canonical");
        document.head.appendChild(el);
      }

      const prevHref = el.getAttribute("href");
      el.setAttribute("href", href);

      cleanups.push(() => {
        if (created) {
          el.remove();
          return;
        }
        if (prevHref === null) {
          el.removeAttribute("href");
          return;
        }
        el.setAttribute("href", prevHref);
      });
    };

    const upsertJsonLd = (payload) => {
      const scriptId = "projects-jsonld";
      let script = document.getElementById(scriptId);
      const created = !script;

      if (!script) {
        script = document.createElement("script");
        script.setAttribute("id", scriptId);
        script.setAttribute("type", "application/ld+json");
        document.head.appendChild(script);
      }

      const prevContent = script.textContent;
      script.textContent = JSON.stringify(payload);

      cleanups.push(() => {
        if (created) {
          script.remove();
          return;
        }
        script.textContent = prevContent;
      });
    };

    document.title = seo.title;
    upsertCanonical(pageUrl);
    upsertMeta('meta[name="description"]', { name: "description" }, seo.description);
    upsertMeta('meta[name="robots"]', { name: "robots" }, "index,follow,max-image-preview:large");
    upsertMeta('meta[property="og:type"]', { property: "og:type" }, "website");
    upsertMeta('meta[property="og:title"]', { property: "og:title" }, seo.title);
    upsertMeta('meta[property="og:description"]', { property: "og:description" }, seo.description);
    upsertMeta('meta[property="og:url"]', { property: "og:url" }, pageUrl);
    upsertMeta('meta[property="og:locale"]', { property: "og:locale" }, seo.locale);
    upsertMeta('meta[property="og:image"]', { property: "og:image" }, ogImage);
    upsertMeta('meta[name="twitter:card"]', { name: "twitter:card" }, "summary_large_image");
    upsertMeta('meta[name="twitter:title"]', { name: "twitter:title" }, seo.title);
    upsertMeta('meta[name="twitter:description"]', { name: "twitter:description" }, seo.description);
    upsertMeta('meta[name="twitter:image"]', { name: "twitter:image" }, ogImage);

    upsertJsonLd({
      "@context": "https://schema.org",
      "@type": "CollectionPage",
      inLanguage: lang,
      name: seo.title,
      description: seo.description,
      url: pageUrl,
      mainEntity: {
        "@type": "ItemList",
        itemListElement: projects.slice(0, 12).map((project, index) => ({
          "@type": "ListItem",
          position: index + 1,
          url: `${window.location.origin}/projects/${project.id}`,
          name: getLocalizedProjectTitle(project, lang, t),
          image: project.cover_image_url ? resolveImageUrl(project.cover_image_url) : undefined
        }))
      }
    });

    return () => {
      document.title = previousTitle;
      cleanups.reverse().forEach((fn) => fn());
    };
  }, [lang, projects, t]);

  return (
    <section className="section-shell pt-16 pb-12">
      <div className="mb-8 flex flex-col gap-4">
        <p className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("projects.title")}</p>
        <h1 className="font-display text-3xl md:text-4xl">{t("projects.subtitle")}</h1>
      </div>

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : projects.length === 0 ? (
        <p className="text-sm text-primary/70">{t("projects.empty")}</p>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => {
            const title = getLocalizedProjectTitle(project, lang, t);
            const description = getLocalizedProjectDescription(project, lang);

            return (
              <Link
                key={project.id}
                to={`/projects/${project.id}`}
                className="group flex h-full flex-col overflow-hidden transition hover:-translate-y-1 hover:shadow-xl"
                style={{ contentVisibility: "auto", containIntrinsicSize: "360px" }}
              >
                <div className="relative aspect-square w-full overflow-hidden bg-primary/10">
                  {project.cover_image_url ? (
                    <img
                      src={resolveImageUrl(project.cover_image_url)}
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

                  <div className={`pointer-events-none absolute inset-x-0 bottom-0 ${directionClass} from-black/70 via-black/25 to-transparent p-4`}>
                    <h3 className="font-display text-xl leading-tight text-white drop-shadow-[0_10px_24px_rgba(0,0,0,0.55)]">{title}</h3>

                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </section>
  );
}
