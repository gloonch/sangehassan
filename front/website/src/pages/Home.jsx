import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { gsap } from "gsap";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import blocksOverlayImage from "@shared/assets/landing_page/landingpage_blocks_overlay.png";
import finishesOverlayVideo from "@shared/assets/landing_page/landingpage_finishes_overlay.mp4";
import miniHeroMessagesIcon from "@shared/assets/landing_page/mini-hero-messages.png";

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

const liveDealsByLang = {
  fa: {
    messagesLabel: "MESSAGES",
    senderName: "SangeHassan",
    messagePrefix: "معامله انجام شد",
    moreMessagesText: "28 more messages",
    deals: [
      { product: "اسلب گرانیت", amount: "۳۲ متر", time: "4 hours ago" },
      { product: "کوپ تراورتن", amount: "۱ کوپ", time: "9 hours ago" },
      { product: "اسلب چینی کریستال", amount: "۲۴ متر", time: "11 hours ago" },
      { product: "اکسسوری مرمریت", amount: "۶ عدد", time: "7 hours ago" },
      { product: "اسلب مرمر", amount: "۱۸ متر", time: "1 day ago" },
      { product: "کوپ گرانیت", amount: "۲ کوپ", time: "2 days ago" },
      { product: "اسلب تراورتن", amount: "۱۴ متر", time: "5 hours ago" },
      { product: "اکسسوری چینی کریستال", amount: "۴ عدد", time: "45 minutes ago" }
    ]
  },
  en: {
    messagesLabel: "MESSAGES",
    senderName: "SangeHassan",
    messagePrefix: "Deal completed",
    moreMessagesText: "28 more messages",
    deals: [
      { product: "Granite Slab", amount: "32 sqm", time: "4 hours ago" },
      { product: "Travertine Block", amount: "1 block", time: "9 hours ago" },
      { product: "Chinese Crystal Slab", amount: "24 sqm", time: "11 hours ago" },
      { product: "Marmarite Accessory", amount: "6 units", time: "7 hours ago" },
      { product: "Marmar Slab", amount: "18 sqm", time: "1 day ago" },
      { product: "Granite Block", amount: "2 blocks", time: "2 days ago" },
      { product: "Travertine Slab", amount: "14 sqm", time: "5 hours ago" },
      { product: "Chinese Crystal Accessory", amount: "4 units", time: "45 minutes ago" }
    ]
  },
  ar: {
    messagesLabel: "MESSAGES",
    senderName: "SangeHassan",
    messagePrefix: "تمت الصفقة",
    moreMessagesText: "28 more messages",
    deals: [
      { product: "ألواح جرانيت", amount: "٣٢ متر", time: "4 hours ago" },
      { product: "بلوك ترافرتين", amount: "١ بلوك", time: "9 hours ago" },
      { product: "ألواح كريستال صيني", amount: "٢٤ متر", time: "11 hours ago" },
      { product: "إكسسوار مرمريت", amount: "٦ قطع", time: "7 hours ago" },
      { product: "ألواح مرمر", amount: "١٨ متر", time: "1 day ago" },
      { product: "بلوك جرانيت", amount: "٢ بلوك", time: "2 days ago" },
      { product: "ألواح ترافرتين", amount: "١٤ متر", time: "5 hours ago" },
      { product: "إكسسوار كريستال صيني", amount: "٤ قطع", time: "45 minutes ago" }
    ]
  }
};

export default function Home() {
  const { t, lang } = useTranslation();
  const [sections, setSections] = useState([]);
  const [activeDealIndex, setActiveDealIndex] = useState(0);
  const [incomingDealIndex, setIncomingDealIndex] = useState(null);
  const [isDealTransitioning, setIsDealTransitioning] = useState(false);
  const rootRef = useRef(null);
  const dealSwitchTimeoutRef = useRef(null);

  useEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return;

    const seo = homeSeoContent[lang] || homeSeoContent.fa;
    const pageUrl = `${window.location.origin}/`;
    const ogImage = blocksOverlayImage.startsWith("http")
      ? blocksOverlayImage
      : `${window.location.origin}${blocksOverlayImage.startsWith("/") ? "" : "/"}${blocksOverlayImage}`;
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
    upsertMeta('meta[name="robots"]', { name: "robots" }, "index,follow,max-image-preview:large");
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
        url: window.location.origin
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
  const liveDealsConfig = liveDealsByLang[lang] || liveDealsByLang.fa;
  const liveDeals = liveDealsConfig.deals;
  const liveDealsMessagesLabel = liveDealsConfig.messagesLabel;
  const liveDealsSenderName = liveDealsConfig.senderName;
  const liveDealsMessagePrefix = liveDealsConfig.messagePrefix;
  const liveDealsMoreMessagesText = liveDealsConfig.moreMessagesText;
  const activeDeal = liveDeals[activeDealIndex] || liveDeals[0];
  const incomingDeal = incomingDealIndex !== null ? liveDeals[incomingDealIndex] : null;
  const previewDealOne = liveDeals.length > 1 ? liveDeals[(activeDealIndex + 1) % liveDeals.length] : null;
  const previewDealTwo = liveDeals.length > 2 ? liveDeals[(activeDealIndex + 2) % liveDeals.length] : null;

  useEffect(() => {
    setActiveDealIndex(0);
    setIncomingDealIndex(null);
    setIsDealTransitioning(false);
    if (dealSwitchTimeoutRef.current) {
      window.clearTimeout(dealSwitchTimeoutRef.current);
    }
  }, [lang]);

  useEffect(() => {
    if (liveDeals.length <= 1 || isDealTransitioning) return undefined;

    const intervalId = window.setInterval(() => {
      const nextIndex = (activeDealIndex + 1) % liveDeals.length;
      setIncomingDealIndex(nextIndex);
      setIsDealTransitioning(true);
      if (dealSwitchTimeoutRef.current) {
        window.clearTimeout(dealSwitchTimeoutRef.current);
      }
      dealSwitchTimeoutRef.current = window.setTimeout(() => {
        setActiveDealIndex(nextIndex);
        setIncomingDealIndex(null);
        setIsDealTransitioning(false);
      }, 680);
    }, 5600);

    return () => {
      window.clearInterval(intervalId);
      if (dealSwitchTimeoutRef.current) {
        window.clearTimeout(dealSwitchTimeoutRef.current);
      }
    };
  }, [activeDealIndex, isDealTransitioning, liveDeals.length]);

  const renderDealMessage = (deal) => `${liveDealsMessagePrefix}: ${deal.amount} | ${deal.product}`;

  const renderDealDepthLayer = (level, deal) => (
    <div
      dir="ltr"
      className={`relative h-full overflow-hidden rounded-[1.08rem] border ${level === 1 ? "border-white/26 bg-white/[0.11]" : "border-white/20 bg-white/[0.08]"} backdrop-blur-[14px]`}
    >
      <span className="pointer-events-none absolute inset-0 bg-[linear-gradient(125deg,rgba(255,255,255,0.14)_0%,rgba(255,255,255,0.03)_55%,rgba(255,255,255,0)_100%)]" />
      <div className="relative px-4 py-3">
        <span className={`block h-1.5 rounded-full ${level === 1 ? "w-20 bg-sand/28" : "w-16 bg-sand/20"}`} />
        <p className={`mt-2 truncate text-[10px] ${level === 1 ? "text-sand/36" : "text-sand/28"}`}>{deal.product}</p>
      </div>
    </div>
  );

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
        <p className="mt-1 text-[11px] leading-snug text-sand/92 sm:text-[12px]">{renderDealMessage(deal)}</p>
        <p className="mt-1.5 text-[10px] font-medium text-sand/68 sm:text-[11px]">{liveDealsMoreMessagesText}</p>
      </div>
    </div>
  );

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
        <span className="pointer-events-none absolute inset-y-0 left-1/2 z-20 hidden w-28 -translate-x-1/2 bg-gradient-to-r from-transparent via-primary/10 to-transparent backdrop-blur-sm lg:block" />
      </section>
      <div
        dir="ltr"
        className="pointer-events-none absolute left-1/2 top-1/2 z-30 h-[10.4rem] w-[min(94vw,25rem)] -translate-x-1/2 -translate-y-1/2 [perspective:1000px] sm:h-[10.8rem]"
      >
        {previewDealTwo ? (
          <div
            className="absolute inset-x-[7%] top-[2.2rem] h-[7.8rem] opacity-40"
            style={{ transform: "scale(0.92) rotateX(11deg)" }}
          >
            {renderDealDepthLayer(2, previewDealTwo)}
          </div>
        ) : null}
        {previewDealOne ? (
          <div
            className="absolute inset-x-[3.5%] top-[1.25rem] h-[7.8rem] opacity-55"
            style={{ transform: "scale(0.965) rotateX(7deg)" }}
          >
            {renderDealDepthLayer(1, previewDealOne)}
          </div>
        ) : null}
        {activeDeal ? (
          <div
            className={`absolute inset-x-0 top-0 h-[7.8rem] transition-all duration-700 ease-[cubic-bezier(0.22,1,0.36,1)] ${isDealTransitioning ? "translate-y-[76px] opacity-0" : "translate-y-0 opacity-100"}`}
          >
            {renderDealNotification(activeDeal)}
          </div>
        ) : null}
        {incomingDeal ? (
          <div
            className={`absolute inset-x-0 top-0 h-[7.8rem] transition-all duration-700 ease-[cubic-bezier(0.22,1,0.36,1)] ${isDealTransitioning ? "translate-y-0 opacity-100" : "-translate-y-[76px] opacity-0"}`}
          >
            {renderDealNotification(incomingDeal)}
          </div>
        ) : null}
      </div>
    </div>
  );
}
