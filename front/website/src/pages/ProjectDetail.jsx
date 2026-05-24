import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { getAbsoluteUrl, getCanonicalUrl, getSiteOrigin } from "../lib/seo";

const projectDetailSeoContent = {
  fa: {
    suffix: "پروژه‌ها | سنگ حسن",
    descriptionFallback:
      "جزئیات پروژه اجراشده سنگ حسن شامل تصاویر، توضیحات فنی و خروجی واقعی اجرا."
  },
  en: {
    suffix: "Projects | SangeHassan",
    descriptionFallback:
      "Detailed view of a completed SangeHassan project including gallery and project description."
  },
  ar: {
    suffix: "المشاريع | سانج حسن",
    descriptionFallback:
      "صفحة تفاصيل مشروع منفذ من سانج حسن مع الصور ووصف المشروع."
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

const getVideoMimeType = (videoPath) => {
  const normalized = String(videoPath || "").toLowerCase();
  if (normalized.endsWith(".webm")) return "video/webm";
  if (normalized.endsWith(".mov")) return "video/quicktime";
  if (normalized.endsWith(".m4v")) return "video/x-m4v";
  return "video/mp4";
};

export default function ProjectDetail() {
  const { id } = useParams();
  const { t, lang } = useTranslation();
  const isRTL = lang === "fa" || lang === "ar";
  const [project, setProject] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeIndex, setActiveIndex] = useState(0);
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [videoEnabled, setVideoEnabled] = useState(false);

  useEffect(() => {
    let mounted = true;

    const load = async () => {
      setLoading(true);
      try {
        const response = await fetchJSON(`/api/projects/${id}`);
        if (!mounted) return;
        setProject(response.data || null);
      } catch (_) {
        if (!mounted) return;
        setProject(null);
      } finally {
        if (mounted) setLoading(false);
      }
    };

    load();
    return () => {
      mounted = false;
    };
  }, [id]);

  const images = useMemo(() => {
    if (!project) return [];
    const rawImages = [project.cover_image_url, ...(Array.isArray(project.gallery_images) ? project.gallery_images : [])].filter(Boolean);
    return [...new Set(rawImages)];
  }, [project]);

  useEffect(() => {
    if (activeIndex >= images.length) {
      setActiveIndex(0);
    }
  }, [activeIndex, images.length]);

  useEffect(() => {
    setActiveIndex(0);
  }, [id]);

  useEffect(() => {
    setVideoEnabled(false);
  }, [id, activeIndex, project?.video_url]);

  useEffect(() => {
    if (!lightboxOpen) return undefined;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [lightboxOpen]);

  const activeImage = images[activeIndex] || images[0] || "";
  const videoPath = project?.video_url ? resolveImageUrl(project.video_url) : "";
  const hasVideo = Boolean(videoPath);
  const localizedTitle = getLocalizedProjectTitle(project, lang, t);
  const localizedDescription = getLocalizedProjectDescription(project, lang);

  useEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return;

    const localizedSeo = projectDetailSeoContent[lang] || projectDetailSeoContent.fa;
    const pageUrl = getCanonicalUrl(`/projects/${id}`);
    const seoTitle = project
      ? `${localizedTitle} | ${localizedSeo.suffix}`
      : localizedSeo.suffix;
    const seoDescription = project
      ? (localizedDescription || localizedSeo.descriptionFallback)
      : localizedSeo.descriptionFallback;
    const ogImage = project?.cover_image_url
      ? getAbsoluteUrl(resolveImageUrl(project.cover_image_url))
      : activeImage
        ? getAbsoluteUrl(resolveImageUrl(activeImage))
        : "";
    const previousTitle = document.title;
    const cleanups = [];

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
      const scriptId = "project-detail-jsonld";
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

    document.title = seoTitle;
    upsertCanonical(pageUrl);
    upsertMeta('meta[name="description"]', { name: "description" }, seoDescription);
    upsertMeta('meta[name="robots"]', { name: "robots" }, project ? "index,follow,max-image-preview:large" : "noindex,follow");
    upsertMeta('meta[property="og:type"]', { property: "og:type" }, "article");
    upsertMeta('meta[property="og:title"]', { property: "og:title" }, seoTitle);
    upsertMeta('meta[property="og:description"]', { property: "og:description" }, seoDescription);
    upsertMeta('meta[property="og:url"]', { property: "og:url" }, pageUrl);
    upsertMeta('meta[property="og:image"]', { property: "og:image" }, ogImage);
    upsertMeta('meta[name="twitter:card"]', { name: "twitter:card" }, "summary_large_image");
    upsertMeta('meta[name="twitter:title"]', { name: "twitter:title" }, seoTitle);
    upsertMeta('meta[name="twitter:description"]', { name: "twitter:description" }, seoDescription);
    upsertMeta('meta[name="twitter:image"]', { name: "twitter:image" }, ogImage);

    if (project) {
      upsertJsonLd({
        "@context": "https://schema.org",
        "@type": "Article",
        headline: localizedTitle,
        description: seoDescription,
        inLanguage: lang,
        url: pageUrl,
        image: ogImage || undefined,
        author: {
          "@type": "Organization",
          name: "SangeHassan"
        },
        publisher: {
          "@type": "Organization",
          name: "SangeHassan",
          url: getSiteOrigin()
        }
      });
    }

    return () => {
      document.title = previousTitle;
      cleanups.reverse().forEach((fn) => fn());
    };
  }, [lang, id, project, activeImage, localizedTitle, localizedDescription]);

  const goPrev = () => {
    if (images.length <= 1) return;
    setActiveIndex((index) => (index - 1 + images.length) % images.length);
  };

  const goNext = () => {
    if (images.length <= 1) return;
    setActiveIndex((index) => (index + 1) % images.length);
  };

  return (
    <section className="section-shell pt-4 pb-12">
      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : !project ? (
        <p className="text-sm text-primary/70">{t("projects.notFound")}</p>
      ) : (
        <div className="space-y-8">
          <div className="space-y-4">
            <div className={`flex items-center ${isRTL ? "justify-end" : "justify-start"}`}>
              <Link
                to="/projects"
                className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70 transition hover:border-primary/50"
              >
                {t("projects.backToProjects")}
              </Link>
            </div>

            <div className="relative overflow-hidden bg-primary/10">
              {videoEnabled && hasVideo ? (
                <video
                  className="h-[42vh] min-h-[320px] w-full object-cover md:h-[52vh]"
                  controls
                  playsInline
                  autoPlay
                  preload="none"
                  poster={activeImage ? resolveImageUrl(activeImage) : undefined}
                >
                  <source src={videoPath} type={getVideoMimeType(project?.video_url)} />
                </video>
              ) : activeImage ? (
                <div className="group relative h-[42vh] min-h-[320px] w-full md:h-[52vh]">
                  <button
                    type="button"
                    className="h-full w-full"
                    onClick={() => setLightboxOpen(true)}
                  >
                    <img
                      src={resolveImageUrl(activeImage)}
                      alt={localizedTitle}
                      className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.02]"
                    />
                  </button>
                  {hasVideo ? (
                    <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
                      <button
                        type="button"
                        onClick={(event) => {
                          event.stopPropagation();
                          setVideoEnabled(true);
                        }}
                        className="pointer-events-auto inline-flex items-center gap-2 rounded-full border border-white/45 bg-black/45 px-4 py-2 text-xs font-semibold text-white transition hover:bg-black/55"
                        aria-label={t("actions.play")}
                      >
                        <span aria-hidden="true">▶</span>
                        <span>{t("actions.play")}</span>
                      </button>
                    </div>
                  ) : null}
                </div>
              ) : hasVideo ? (
                <button
                  type="button"
                  onClick={() => setVideoEnabled(true)}
                  className="flex h-[42vh] min-h-[320px] w-full items-center justify-center bg-primary/15 md:h-[52vh]"
                  aria-label={t("actions.play")}
                >
                  <span className="inline-flex items-center gap-2 rounded-full border border-primary/35 bg-white/75 px-4 py-2 text-xs font-semibold text-primary">
                    <span aria-hidden="true">▶</span>
                    <span>{t("actions.play")}</span>
                  </span>
                </button>
              ) : (
                <div className="flex h-[42vh] min-h-[320px] w-full items-center justify-center text-sm text-primary/60">
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
                        className={`relative h-16 w-16 flex-shrink-0 overflow-hidden border transition ${isActive ? "border-primary" : "border-primary/20 hover:border-primary/45"
                          }`}
                        aria-label={`project-image-${index + 1}`}
                      >
                        <img src={resolveImageUrl(image)} alt="" loading="lazy" className="h-full w-full object-cover" />
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            <div className="space-y-2">
              <h1 className="font-display text-3xl leading-tight text-primary md:text-5xl">{localizedTitle}</h1>
            </div>
          </div>

          <section className="space-y-4 border-t border-primary/20 pt-6">
            <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("projects.descriptionTitle")}</p>
            {localizedDescription ? (
              <p className="whitespace-pre-line text-sm text-primary/75">{localizedDescription}</p>
            ) : (
              <p className="text-sm text-primary/70">{t("projects.noDescription")}</p>
            )}
          </section>
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
              <img src={resolveImageUrl(activeImage)} alt={localizedTitle} className="h-full w-full object-contain" />
            </div>
            {images.length > 1 && (
              <div className="mt-4 flex items-center justify-between gap-3">
                <button
                  type="button"
                  onClick={goPrev}
                  className="rounded-full border border-primary/25 px-4 py-2 text-xs font-semibold text-primary transition hover:border-primary/50"
                >
                  {t("projects.prev")}
                </button>
                <span className="text-xs font-semibold text-primary/65">
                  {activeIndex + 1}/{images.length}
                </span>
                <button
                  type="button"
                  onClick={goNext}
                  className="rounded-full border border-primary/25 px-4 py-2 text-xs font-semibold text-primary transition hover:border-primary/50"
                >
                  {t("projects.next")}
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </section>
  );
}
