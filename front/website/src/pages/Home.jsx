import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { gsap } from "gsap";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { getAbsoluteUrl, getCanonicalUrl, getSiteOrigin } from "../lib/seo";
import blocksOverlayImage from "@shared/assets/landing_page/landingpage_blocks_overlay.webp";
import blockImage01 from "@shared/assets/landing_page/blocks/block-slide-01.webp";
import blockImage02 from "@shared/assets/landing_page/blocks/block-slide-02.webp";
import blockImage03 from "@shared/assets/landing_page/blocks/block-slide-03.webp";
import blockImage04 from "@shared/assets/landing_page/blocks/block-slide-04.webp";
import blockImage05 from "@shared/assets/landing_page/blocks/block-slide-05.webp";
import blockImage06 from "@shared/assets/landing_page/blocks/block-slide-06.webp";
import blockImage07 from "@shared/assets/landing_page/blocks/block-slide-07.webp";
import blockImage08 from "@shared/assets/landing_page/blocks/block-slide-08.webp";
import finishesImage01 from "@shared/assets/landing_page/products/finish-slide-01.webp";
import finishesImage02 from "@shared/assets/landing_page/products/finish-slide-02.webp";
import finishesImage03 from "@shared/assets/landing_page/products/finish-slide-03.webp";
import finishesImage04 from "@shared/assets/landing_page/products/finish-slide-04.webp";
import productImage01 from "@shared/assets/landing_page/products/product-slide-01.webp";
import productImage02 from "@shared/assets/landing_page/products/product-slide-02.webp";
import productImage03 from "@shared/assets/landing_page/products/product-slide-03.webp";

const productSlides = [
  { src: finishesImage01, width: 736, height: 981 },
  { src: finishesImage02, width: 736, height: 1508 },
  { src: finishesImage03, width: 736, height: 1308 },
  { src: finishesImage04, width: 735, height: 825 },
  { src: productImage01, width: 900, height: 862 },
  { src: productImage02, width: 900, height: 864 },
  { src: productImage03, width: 900, height: 850 }
];
const blockSlides = [
  { src: blockImage01, width: 900, height: 895 },
  { src: blockImage02, width: 900, height: 894 },
  { src: blockImage03, width: 725, height: 906 },
  { src: blockImage04, width: 900, height: 1123 },
  { src: blockImage05, width: 900, height: 1123 },
  { src: blockImage06, width: 900, height: 1104 },
  { src: blockImage07, width: 900, height: 1106 },
  { src: blockImage08, width: 900, height: 1104 }
];
const fallbackSlide = { src: blocksOverlayImage, width: 1068, height: 845 };

const shuffleSlides = (slides) => {
  const shuffled = [...slides];
  for (let index = shuffled.length - 1; index > 0; index -= 1) {
    const swapIndex = Math.floor(Math.random() * (index + 1));
    [shuffled[index], shuffled[swapIndex]] = [shuffled[swapIndex], shuffled[index]];
  }
  return shuffled;
};

const initialSlideDecks = () => ({
  products: productSlides.length ? productSlides : [fallbackSlide],
  blocks: blockSlides.length ? blockSlides : [fallbackSlide]
});

const getLocalStorageValue = (key) => {
  if (typeof window === "undefined") return null;
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
};

const setLocalStorageValue = (key, value) => {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(key, value);
  } catch {
    /* Storage can be unavailable in private or restricted browser modes. */
  }
};

const shuffleForNewVisit = (slides, storageKey) => {
  const shuffled = shuffleSlides(slides);
  if (typeof window === "undefined" || shuffled.length < 2) return shuffled;

  const previousFirstSlide = getLocalStorageValue(storageKey);
  if (previousFirstSlide && shuffled[0]?.src === previousFirstSlide) {
    const swapIndex = 1 + Math.floor(Math.random() * (shuffled.length - 1));
    [shuffled[0], shuffled[swapIndex]] = [shuffled[swapIndex], shuffled[0]];
  }

  setLocalStorageValue(storageKey, shuffled[0]?.src || "");
  return shuffled;
};

const createSlideDecks = () => ({
  products: shuffleForNewVisit(productSlides.length ? productSlides : [fallbackSlide], "sh-home-products-first-slide"),
  blocks: shuffleForNewVisit(blockSlides.length ? blockSlides : [fallbackSlide], "sh-home-blocks-first-slide")
});

const slideTransitionMs = 1800;

const getNextSlideIndex = (activeSlide, slides) => {
  if (!slides.length || slides.length < 2) return null;
  return (activeSlide + 1) % slides.length;
};

const isSlideDebugEnabled = () => {
  if (typeof window === "undefined") return false;
  return window.location.search.includes("debugSlides=1") || getLocalStorageValue("sh-debug-slides") === "1";
};

const getSlideAssetName = (src = "") => src.split("/").pop() || src;

const logSlideDebug = (event, payload = {}) => {
  if (!isSlideDebugEnabled()) return;
  console.info("[home-slides]", event, {
    at: Math.round(performance.now()),
    ...payload
  });
};

const homeSeoContent = {
  fa: {
    title: "سنگ حسن | تامین و تولید سنگ طبیعی",
    description:
      "شبکه تامین و تولید سنگ طبیعی سنگ حسن؛ از کوپ و بلوک تا محصولات فرآوری‌شده برای پروژه‌های حرفه‌ای، B2B و صادرات.",
    locale: "fa_IR"
  },
  en: {
    title: "SangeHassan | Natural Stone Supply & Production",
    description:
      "Integrated natural stone supply and production network, from quarry blocks to finished stone products for professional projects, B2B, and export.",
    locale: "en_US"
  },
  ar: {
    title: "سانج حسن | توريد وإنتاج الحجر الطبيعي",
    description:
      "شبكة متكاملة لتوريد وإنتاج الحجر الطبيعي من البلوك الخام حتى المنتجات المعالجة للمشاريع المهنية والتعاون B2B والتصدير.",
    locale: "ar_SA"
  }
};

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
  const [slideDecks, setSlideDecks] = useState(initialSlideDecks);
  const [activeSlides, setActiveSlides] = useState({ products: 0, blocks: 0 });
  const [previousSlides, setPreviousSlides] = useState({ products: null, blocks: null });
  const previousClearTimerRef = useRef(null);
  const rootRef = useRef(null);

  useEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return;

    const seo = homeSeoContent[lang] || homeSeoContent.fa;
    const pageUrl = getCanonicalUrl("/");
    const ogImage = getAbsoluteUrl(blocksOverlayImage);
    const previousTitle = document.title;
    const cleanups = [];

    const upsertMeta = (selector, createAttrs, value) => {
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
      const scriptId = "home-jsonld";
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
    upsertMeta('meta[name="robots"]', { name: "robots" }, "index,follow");
    upsertMeta('meta[property="og:type"]', { property: "og:type" }, "website");
    upsertMeta('meta[property="og:title"]', { property: "og:title" }, seo.title);
    upsertMeta('meta[property="og:description"]', { property: "og:description" }, seo.description);
    upsertMeta('meta[property="og:url"]', { property: "og:url" }, pageUrl);
    upsertMeta('meta[property="og:image"]', { property: "og:image" }, ogImage);
    upsertMeta('meta[property="og:locale"]', { property: "og:locale" }, seo.locale);
    upsertMeta('meta[name="twitter:card"]', { name: "twitter:card" }, "summary_large_image");
    upsertMeta('meta[name="twitter:title"]', { name: "twitter:title" }, seo.title);
    upsertMeta('meta[name="twitter:description"]', { name: "twitter:description" }, seo.description);
    upsertMeta('meta[name="twitter:image"]', { name: "twitter:image" }, ogImage);

    upsertJsonLd({
      "@context": "https://schema.org",
      "@type": "WebSite",
      inLanguage: lang,
      name: "SangeHassan",
      url: pageUrl,
      description: seo.description,
      publisher: {
        "@type": "Organization",
        name: "SangeHassan",
        url: getSiteOrigin()
      }
    });

    return () => {
      document.title = previousTitle;
      cleanups.reverse().forEach((fn) => fn());
    };
  }, [lang]);

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

  useEffect(() => {
    const decks = createSlideDecks();
    logSlideDebug("deck-init", {
      products: decks.products.map((slide) => getSlideAssetName(slide.src)),
      blocks: decks.blocks.map((slide) => getSlideAssetName(slide.src))
    });
    setSlideDecks(decks);
    setActiveSlides({ products: 0, blocks: 0 });
    setPreviousSlides({ products: null, blocks: null });
  }, []);

  useEffect(() => {
    const timer = window.setInterval(() => {
      setActiveSlides((current) => {
        const next = {
          products: (current.products + 1) % slideDecks.products.length,
          blocks: (current.blocks + 1) % slideDecks.blocks.length
        };

        logSlideDebug("tick", {
          previous: current,
          next,
          products: {
            previousAsset: getSlideAssetName(slideDecks.products[current.products]?.src),
            nextAsset: getSlideAssetName(slideDecks.products[next.products]?.src)
          },
          blocks: {
            previousAsset: getSlideAssetName(slideDecks.blocks[current.blocks]?.src),
            nextAsset: getSlideAssetName(slideDecks.blocks[next.blocks]?.src)
          }
        });

        setPreviousSlides(current);

        if (previousClearTimerRef.current) {
          window.clearTimeout(previousClearTimerRef.current);
        }
        previousClearTimerRef.current = window.setTimeout(() => {
          setPreviousSlides({ products: null, blocks: null });
        }, slideTransitionMs + 120);

        return next;
      });
    }, 2000);

    return () => {
      window.clearInterval(timer);
      if (previousClearTimerRef.current) {
        window.clearTimeout(previousClearTimerRef.current);
      }
    };
  }, [slideDecks]);

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
        cta_href: "/products",
        images: []
      }
    }),
    [t]
  );

  const blocksSection = sectionsByKey.blocks || fallbackSections.blocks;
  const finishedSection = sectionsByKey.finished || fallbackSections.finished;

  /*
  const renderDealNotification = (deal) => (
    <div
      dir="ltr"
      className="relative h-full overflow-hidden rounded-[1.15rem] border border-white/35 bg-white/[0.12] px-4 py-3 text-left text-sand shadow-[0_28px_75px_rgba(10,8,5,0.45)] backdrop-blur-[20px] sm:px-5 sm:py-3.5"
    >
      <span className="pointer-events-none absolute inset-0 bg-[linear-gradient(125deg,rgba(255,255,255,0.2)_0%,rgba(255,255,255,0.06)_38%,rgba(255,255,255,0)_100%)]" />
      <span className="pointer-events-none absolute -left-10 top-0 h-full w-16 rotate-[14deg] bg-white/15 blur-2xl" />
      <div className="relative">
        <div className="flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-2">
            <img src={miniHeroMessagesIcon} alt="" className="h-5 w-5 rounded-[6px]" />
            <p className="truncate text-[9px] font-semibold uppercase tracking-[0.18em] text-sand/80 sm:text-[10px]">
              {liveDealsMessagesLabel}
            </p>
          </div>
          <p className="shrink-0 text-[10px] text-sand/70 sm:text-[11px]">{deal.time}</p>
        </div>
        <p className="mt-2 text-[12px] font-semibold text-sand sm:text-[13px]">{liveDealsSenderName}</p>
        <p className="mt-1 text-[11px] leading-snug text-sand/92 sm:text-[12px]">{renderDealMessage(deal, liveDealsConfig)}</p>
        <p className="mt-1.5 text-[10px] font-medium text-sand/68 sm:text-[11px]">{liveDealsMoreMessagesText}</p>
      </div>
    </div>
  );
  */

  useEffect(() => {
    const root = rootRef.current;
    if (!root || typeof window === "undefined" || !window.matchMedia) return;
    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    const items = root.querySelectorAll("[data-home-anim='item']");
    if (!items.length) return;

    const ctx = gsap.context(() => {
      gsap.fromTo(
        items,
        { autoAlpha: 0, y: 20 },
        {
          autoAlpha: 1,
          y: 0,
          duration: reduceMotion ? 0.45 : 1.2,
          delay: reduceMotion ? 0.1 : 0.5,
          stagger: reduceMotion ? 0.03 : 0.09,
          ease: "power3.out",
          overwrite: "auto"
        }
      );
    }, root);

    return () => ctx.revert();
  }, [lang, blocksSection, finishedSection]);

  return (
    <div ref={rootRef} className="relative h-[100dvh] w-full overflow-hidden">
      <section className="relative flex h-full w-full flex-col overflow-hidden lg:flex-row">
        {[finishedSection, blocksSection].map((section, index) => {
          const isBlocks = section.key === "blocks";
          const title = pickField(section, "title", lang) || (isBlocks ? t("blocks.title") : t("products.title"));
          const subtitle = pickField(section, "subtitle", lang);
          const description = pickField(section, "description", lang);
          const ctaLabel = pickField(section, "cta_label", lang);
          const fallbackCTA = isBlocks ? t("blocks.cta") : t("hero.ctaPrimary");
          const rawCtaHref = section.cta_href || (isBlocks ? "/blocks" : "/products");
          const ctaHref =
            rawCtaHref === "/products/overview" ? "/products" : rawCtaHref === "/blocks/catalog" ? "/blocks" : rawCtaHref;
          const lines = splitLines(description);
          const slides = isBlocks ? slideDecks.blocks : slideDecks.products;
          const activeSlide = isBlocks ? activeSlides.blocks : activeSlides.products;
          const previousSlide = isBlocks ? previousSlides.blocks : previousSlides.products;
          const nextSlide = getNextSlideIndex(activeSlide, slides);
          const visibleSlideIndexes = [previousSlide, activeSlide, nextSlide].filter(
            (slideIndex, slideIndexPosition, slideIndexes) =>
              slideIndex !== null && slideIndexes.indexOf(slideIndex) === slideIndexPosition
          );

          return (
            <Link
              key={section.key || index}
              to={ctaHref}
              className="group relative block h-1/2 w-full overflow-hidden lg:h-full lg:w-1/2"
              aria-label={title}
            >
              <div className="absolute inset-0 z-0">
                {visibleSlideIndexes.map((slideIndex) => {
                  const image = slides[slideIndex];
                  const isActiveSlide = slideIndex === activeSlide;
                  const isPreviousSlide = slideIndex === previousSlide;
                  const role = isActiveSlide ? "active" : isPreviousSlide ? "previous" : "preload";
                  const panel = isBlocks ? "blocks" : "products";

                  return (
                    <img
                      key={image.src}
                      src={image.src}
                      alt=""
                      width={image.width}
                      height={image.height}
                      className={`landing-slide-layer absolute inset-0 h-full w-full object-cover object-center ${isActiveSlide
                        ? "landing-slide-active"
                        : isPreviousSlide
                          ? "landing-slide-previous"
                          : "landing-slide-preload"
                        }`}
                      data-slide-panel={panel}
                      data-slide-role={role}
                      data-slide-index={slideIndex}
                      loading={isPreviousSlide ? "lazy" : "eager"}
                      decoding="async"
                      fetchpriority={isActiveSlide ? "high" : "low"}
                      onLoad={(event) => {
                        const styles = window.getComputedStyle(event.currentTarget);
                        logSlideDebug("image-load", {
                          panel,
                          role,
                          slideIndex,
                          asset: getSlideAssetName(image.src),
                          opacity: styles.opacity,
                          transitionDuration: styles.transitionDuration
                        });
                      }}
                      onTransitionStart={(event) => {
                        if (event.propertyName !== "opacity") return;
                        const styles = window.getComputedStyle(event.currentTarget);
                        logSlideDebug("transition-start", {
                          panel,
                          role,
                          slideIndex,
                          asset: getSlideAssetName(image.src),
                          opacity: styles.opacity,
                          filter: styles.filter,
                          transitionDuration: styles.transitionDuration
                        });
                      }}
                      onTransitionEnd={(event) => {
                        if (event.propertyName !== "opacity") return;
                        const styles = window.getComputedStyle(event.currentTarget);
                        logSlideDebug("transition-end", {
                          panel,
                          role,
                          slideIndex,
                          asset: getSlideAssetName(image.src),
                          opacity: styles.opacity,
                          filter: styles.filter,
                          transitionDuration: styles.transitionDuration
                        });
                      }}
                    />
                  );
                })}
              </div>

              <div className="absolute inset-0 z-[1] bg-primary/25" />
              <div
                className={`absolute inset-0 z-[2] ${isBlocks
                  ? "bg-gradient-to-br from-primary/35 via-primary/15 to-primary/48"
                  : "bg-gradient-to-br from-primary/40 via-primary/20 to-primary/52"
                  }`}
              />
              <div className="absolute inset-0 z-[3] bg-[radial-gradient(circle_at_25%_20%,rgba(165,141,102,0.16),transparent_48%)]" />

              <div className="relative z-10 flex h-full items-center justify-center px-6 py-10 text-sand sm:px-10 lg:px-12">
                <div className="mx-auto w-full max-w-[38rem] text-center">
                  <p data-home-anim="item" className="text-[10px] uppercase tracking-[0.34em] text-sand/72 sm:text-xs">
                    {isBlocks ? t("blocks.title") : t("products.title")}
                  </p>
                  <h1 data-home-anim="item" className="mt-2 font-display text-xl leading-tight sm:text-2xl md:text-3xl lg:text-4xl">
                    {title}
                  </h1>
                  {subtitle ? <p data-home-anim="item" className="mx-auto mt-3 max-w-[33rem] text-[11px] text-sand/88 sm:text-xs md:text-sm lg:text-base">{subtitle}</p> : null}

                  {lines.length > 0 && (
                    <ul className="mx-auto mt-4 max-w-[32rem] space-y-1.5 text-left text-[10px] text-sand/88 sm:text-[11px] md:text-xs lg:text-sm">
                      {lines.slice(0, 3).map((line) => (
                        <li key={line} data-home-anim="item" className="flex items-start gap-2">
                          <span className="mt-1.5 h-1.5 w-1.5 rounded-full bg-accent/85" />
                          <span>{line}</span>
                        </li>
                      ))}
                    </ul>
                  )}

                  <div data-home-anim="item" className="mt-7 flex items-center justify-center gap-3">
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
      </section>
    </div>
  );
}
