import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { getAbsoluteUrl, getCanonicalUrl, getSiteOrigin } from "../lib/seo";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "";
const IMAGE_REV = import.meta.env.VITE_IMAGE_REV || "about-2026-04-17-r4";
const imageUrl = (path) => `${API_BASE}/images/${path}?v=${IMAGE_REV}`;

const seoContent = {
  fa: {
    title: "درباره ما | سنگ حسن",
    description:
      "سنگ حسن، شبکه تامین و تولید سنگ طبیعی از معدن تا محصول نهایی؛ با تمرکز بر همکاری B2B، کنترل کیفیت و صادرات بین‌المللی.",
    locale: "fa_IR"
  },
  en: {
    title: "About Us | SangeHassan",
    description:
      "SangeHassan is an integrated natural stone supply and production network, from quarry blocks to finished products with B2B reliability and export focus.",
    locale: "en_US"
  },
  ar: {
    title: "من نحن | سانج حسن",
    description:
      "سانج حسن شبكة متكاملة لتوريد وإنتاج الحجر الطبيعي من المحجر حتى المنتج النهائي، مع تركيز على التعاون B2B والتصدير.",
    locale: "ar_SA"
  }
};

const baseImages = {
  hero: imageUrl("aboutus/workforce_at_travertine_quarry.png"),
  services: imageUrl("aboutus/services/services_six_points.png"),
  tracking: imageUrl("aboutus/tracking/tracking_four_points.png"),
  "1383": imageUrl("aboutus/1383/photo_2026-04-09%2011.52.40.jpeg"),
  "1388": imageUrl("aboutus/1388/Gemini_Generated_Image_e3c9bne3c9bne3c9.png"),
  "1394": imageUrl("aboutus/1394/Gemini_Generated_Image_uqqi9quqqi9quqqi.png"),
  "1400": imageUrl("aboutus/1400/Gemini_Generated_Image_1513k91513k91513.png")
};

const content = {
  fa: {
    hero: {
      title: "شبکه تامین و تولید سنگ حسن",
      body:
        "سنگ حسن یک شبکه تامین و تولید سنگ ساختمانی طبیعی است؛ از کوپ و سنگ خام معدن تا اسلب، تایل و محصولات فرآوری‌شده. تمرکز ما همکاری B2B و صادرات بین‌المللی با مسیر شفاف، کنترل کیفیت و تحویل قابل اتکاست.",
      chips: ["Block", "Slab", "B2B Supply", "Export"],
      image: baseImages.hero
    },
    chapters: [
      {
        id: "1383",
        year: "۱۳۸۳",
        rail: "۱۳۸۳",
        title: "شروع از بازار سنتی سنگ تهران",
        body:
          "شروع از بازار مرکزی سنگ تهران و یادگیری میدانی کیفیت واقعی، پایه شناخت دقیق خرید و فروش و تحلیل ماده را شکل داد.",
        image: baseImages["1383"],
        shape: "bg-[#C18B66]/35"
      },
      {
        id: "1388",
        year: "۱۳۸۸",
        rail: "۱۳۸۸",
        title: "ورود جدی به خام و کوپ",
        body:
          "ورود به دیلری سنگ خام و کوپ، مزدی‌بری و کنترل برش در کارخانه‌ها، مسیر تامین تا تبدیل را عملیاتی و قابل‌کنترل کرد.",
        image: baseImages["1388"],
        shape: "bg-[#D8C6A8]/46"
      },
      {
        id: "1394",
        year: "۱۳۹۴",
        rail: "۱۳۹۴",
        title: "کارخانه نطنز و استانداردسازی",
        body:
          "با راه‌اندازی کارخانه نطنز، ثبات کیفیت، استانداردسازی خروجی و پاسخ‌گویی دقیق به سفارش‌های چندمرحله‌ای به سطح حرفه‌ای رسید.",
        image: baseImages["1394"],
        shape: "bg-[#6D7F8B]/30"
      },
      {
        id: "1400",
        year: "۱۴۰۰",
        rail: "۱۴۰۰",
        title: "بهره‌برداری معدن تراورتن حسن",
        body:
          "بهره‌برداری از معدن تراورتن حسن حلقه زنجیره را کامل کرد: کنترل کیفیت از مبدا، ثبات تامین و برنامه‌ریزی دقیق‌تر تحویل.",
        image: baseImages["1400"],
        shape: "bg-[#C9A783]/38"
      }
    ],
    services: {
      rail: "۱۴۰۵",
      year: "SYSTEM",
      title: "خدمات",
      body:
        "با یکپارچه‌سازی فرآیندها، هر مرحله از تامین سنگ تا تحویل نهایی به‌صورت سیستماتیک تعریف و ثبت شد؛ از تامین از معدن و کنترل کیفیت تا تولید، فرآوری، بسته‌بندی، ارسال و رهگیری وضعیت سفارش. این ساختار، شفافیت، دقت و قابلیت اتکای بیشتری برای همکاری‌های حرفه‌ای و صادراتی ایجاد می‌کند.",
      image: baseImages.services,
      shape: "bg-[#9A8A76]/30",
      items: [
        { id: "quarry", title: "تامین از معدن", subtitle: "Quarry Block" },
        { id: "qc", title: "کنترل کیفیت", subtitle: "QC Review" },
        { id: "cut", title: "برش و تولید", subtitle: "Slab / Tile" },
        { id: "finish", title: "فراوری", subtitle: "Finishes" },
        { id: "pack", title: "بسته‌بندی", subtitle: "Packing" },
        { id: "ship", title: "ارسال", subtitle: "Shipping" }
      ]
    },
    tracking: {
      rail: "رهگیری",
      year: "TRACE",
      title: "شفافیت سفارش تا تحویل",
      body:
        "وضعیت سفارش به‌صورت مرحله‌ای ثبت می‌شود تا مشتری در هر لحظه بداند سفارش کجاست، مرحله بعد چیست و چه مدارکی ثبت شده است.",
      image: baseImages.tracking,
      shape: "bg-[#7F95A2]/28",
      points: ["وضعیت فعلی", "مرحله بعد", "ETA", "مدارک و گزارش QC"]
    },
    projects: {
      rail: "نمونه‌ها",
      year: "PROOF",
      title: "پروژه‌ها و نمونه‌کارها",
      body:
        "برای مشاهده خروجی واقعی از نمای ساختمان تا طراحی داخلی، نمونه‌پروژه‌ها را بررسی کنید.",
      image: baseImages["1400"],
      shape: "bg-[#B58B67]/34",
      cta: "مشاهده همه"
    }

  },
  en: {
    hero: {
      title: "SangeHassan Supply & Production Network",
      body:
        "SangeHassan is a natural building-stone supply and production network, from quarry raw blocks and rough stone to slabs, tiles, and finished products. Our focus is reliable B2B collaboration and international export with transparent workflow and quality control.",
      chips: ["Block", "Slab", "B2B Supply", "Export"],
      image: baseImages.hero
    },
    chapters: [
      {
        id: "1383",
        year: "2002 / 1383",
        rail: "2002",
        title: "Starting from Tehran's Traditional Stone Market",
        body:
          "The journey started in Tehran's central stone market. Practical, on-site learning of real quality became the base for accurate sourcing and trading decisions.",
        image: baseImages["1383"],
        shape: "bg-[#C18B66]/35"
      },
      {
        id: "1388",
        year: "2009 / 1388",
        rail: "2009",
        title: "Entering Raw Stone and Block Operations",
        body:
          "By moving into raw-stone dealership, toll cutting, and factory cutting supervision, the full path from sourcing to deliverable product became operational and controlled.",
        image: baseImages["1388"],
        shape: "bg-[#D8C6A8]/46"
      },
      {
        id: "1394",
        year: "2015 / 1394",
        rail: "2015",
        title: "Natanz Factory and Standardization",
        body:
          "Launching the Natanz factory enabled stronger process control, stable quality, output standardization, and more accurate response to multi-stage orders.",
        image: baseImages["1394"],
        shape: "bg-[#6D7F8B]/30"
      },
      {
        id: "1400",
        year: "2021 / 1400",
        rail: "2021",
        title: "Operating Hassan Travertine Quarry",
        body:
          "Operating the Hassan travertine quarry completed a key loop in the chain: better input quality control, steadier supply, and more predictable delivery planning.",
        image: baseImages["1400"],
        shape: "bg-[#C9A783]/38"
      }
    ],
    services: {
      rail: "2026",
      year: "SYSTEM",
      title: "Services",
      body:
        "With a unified process, each stage from stone sourcing to final delivery is systematically defined and documented: quarry supply, quality control, production, finishing, packing, shipping, and order tracking.",
      image: baseImages.services,
      shape: "bg-[#9A8A76]/30",
      items: [
        { id: "quarry", title: "Quarry Supply", subtitle: "Quarry Block" },
        { id: "qc", title: "Quality Control", subtitle: "QC Review" },
        { id: "cut", title: "Cutting & Production", subtitle: "Slab / Tile" },
        { id: "finish", title: "Processing", subtitle: "Finishes" },
        { id: "pack", title: "Packaging", subtitle: "Packing" },
        { id: "ship", title: "Shipping", subtitle: "Shipping" }
      ]
    },
    tracking: {
      rail: "Tracking",
      year: "TRACE",
      title: "Order-to-Delivery Transparency",
      body:
        "Order status is recorded step by step so clients always know where the order is, what comes next, and which documents are already registered.",
      image: baseImages.tracking,
      shape: "bg-[#7F95A2]/28",
      points: ["Current Status", "Next Stage", "ETA", "Documents & QC Report"]
    },
    projects: {
      rail: "Projects",
      year: "PROOF",
      title: "Projects & Portfolio",
      body:
        "To see real project outcomes in facade and interior applications, review our completed projects.",
      image: baseImages["1400"],
      shape: "bg-[#B58B67]/34",
      cta: "View All Projects"
    }
  },
  ar: {
    hero: {
      title: "شبكة سانج حسن للتوريد والإنتاج",
      body:
        "سانج حسن هي شبكة لتوريد وإنتاج الحجر الطبيعي للبناء، من البلوك الخام وحجر المحجر إلى الألواح والبلاط والمنتجات المعالجة. تركيزنا على التعاون المهني B2B والتصدير الدولي مع مسار واضح وضبط جودة.",
      chips: ["Block", "Slab", "B2B Supply", "Export"],
      image: baseImages.hero
    },
    chapters: [
      {
        id: "1383",
        year: "٢٠٠٢ / ١٣٨٣",
        rail: "٢٠٠٢",
        title: "البداية من سوق الحجر التقليدي في طهران",
        body:
          "بدأ المسار من السوق المركزي للحجر في طهران، حيث شكّل التعلم الميداني لجودة الحجر الحقيقية أساس الفهم الدقيق للشراء والبيع.",
        image: baseImages["1383"],
        shape: "bg-[#C18B66]/35"
      },
      {
        id: "1388",
        year: "٢٠٠٩ / ١٣٨٨",
        rail: "٢٠٠٩",
        title: "دخول جاد إلى الخام والبلوك",
        body:
          "الدخول إلى تجارة الحجر الخام والبلوك مع القصّ بالأجر ومتابعة القص في المصانع جعل مسار التوريد إلى المنتج النهائي أكثر انضباطًا.",
        image: baseImages["1388"],
        shape: "bg-[#D8C6A8]/46"
      },
      {
        id: "1394",
        year: "٢٠١٥ / ١٣٩٤",
        rail: "٢٠١٥",
        title: "مصنع نطنز وتوحيد المعايير",
        body:
          "إطلاق مصنع نطنز رفع مستوى التحكم في العملية، وثبات الجودة، وتوحيد المخرجات، والاستجابة الأدق للطلبات متعددة المراحل.",
        image: baseImages["1394"],
        shape: "bg-[#6D7F8B]/30"
      },
      {
        id: "1400",
        year: "٢٠٢١ / ١٤٠٠",
        rail: "٢٠٢١",
        title: "تشغيل منجم حسن للترافرتين",
        body:
          "تشغيل منجم الترافرتين الخاص بحسن أكمل حلقة مهمة في الشبكة: ضبط أفضل لجودة المدخلات، وثبات أكبر في التوريد، وتخطيط أدق للتسليم.",
        image: baseImages["1400"],
        shape: "bg-[#C9A783]/38"
      }
    ],
    services: {
      rail: "٢٠٢٦",
      year: "SYSTEM",
      title: "الخدمات",
      body:
        "بعد توحيد العمليات، أصبحت كل مرحلة من التوريد إلى التسليم النهائي موثقة ومنظمة: التوريد من المحجر، ضبط الجودة، الإنتاج، المعالجة، التغليف، الشحن، وتتبع الطلب.",
      image: baseImages.services,
      shape: "bg-[#9A8A76]/30",
      items: [
        { id: "quarry", title: "توريد من المحجر", subtitle: "Quarry Block" },
        { id: "qc", title: "ضبط الجودة", subtitle: "QC Review" },
        { id: "cut", title: "القص والإنتاج", subtitle: "Slab / Tile" },
        { id: "finish", title: "المعالجة", subtitle: "Finishes" },
        { id: "pack", title: "التغليف", subtitle: "Packing" },
        { id: "ship", title: "الشحن", subtitle: "Shipping" }
      ]
    },
    tracking: {
      rail: "التتبع",
      year: "TRACE",
      title: "شفافية الطلب حتى التسليم",
      body:
        "يتم تسجيل حالة الطلب خطوة بخطوة حتى يعرف العميل دائمًا موقع الطلب، المرحلة التالية، والوثائق المسجلة لكل مرحلة.",
      image: baseImages.tracking,
      shape: "bg-[#7F95A2]/28",
      points: ["الحالة الحالية", "المرحلة التالية", "ETA", "الوثائق وتقرير الجودة"]
    },
    projects: {
      rail: "النماذج",
      year: "PROOF",
      title: "المشاريع ونماذج الأعمال",
      body:
        "لمشاهدة النتائج الفعلية في الواجهات والتصميم الداخلي، راجع صفحة المشاريع المنجزة.",
      image: baseImages["1400"],
      shape: "bg-[#B58B67]/34",
      cta: "عرض جميع المشاريع"
    }
  }
};

function ServiceGlyph({ id }) {
  const iconProps = {
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.8,
    strokeLinecap: "round",
    strokeLinejoin: "round",
    className: "h-4 w-4"
  };

  if (id === "quarry") {
    return <svg {...iconProps}><path d="M3 18L8 10L13 18" /><path d="M10 18L14 12L19 18" /></svg>;
  }
  if (id === "qc") {
    return <svg {...iconProps}><path d="M5 12L10 17L19 7" /></svg>;
  }
  if (id === "cut") {
    return <svg {...iconProps}><path d="M4 8H20" /><path d="M4 16H20" /><path d="M12 5V19" /></svg>;
  }
  if (id === "finish") {
    return <svg {...iconProps}><path d="M12 3L14.3 9.7L21 12L14.3 14.3L12 21L9.7 14.3L3 12L9.7 9.7Z" /></svg>;
  }
  if (id === "pack") {
    return <svg {...iconProps}><path d="M4 8L12 4L20 8L12 12Z" /><path d="M4 8V16L12 20L20 16V8" /></svg>;
  }
  return <svg {...iconProps}><path d="M3 12H21" /><path d="M16 7L21 12L16 17" /></svg>;
}

function TrackingGlyph({ index }) {
  const iconProps = {
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.8,
    strokeLinecap: "round",
    strokeLinejoin: "round",
    className: "h-4 w-4"
  };

  if (index === 0) {
    return <svg {...iconProps}><circle cx="12" cy="12" r="8" /><circle cx="12" cy="12" r="2" fill="currentColor" /></svg>;
  }
  if (index === 1) {
    return <svg {...iconProps}><path d="M4 12H20" /><path d="M14 6L20 12L14 18" /></svg>;
  }
  if (index === 2) {
    return <svg {...iconProps}><circle cx="12" cy="12" r="8" /><path d="M12 8V12L15 14" /></svg>;
  }
  return <svg {...iconProps}><path d="M7 3H17L21 7V21H7Z" /><path d="M17 3V7H21" /><path d="M10 12H16" /><path d="M10 16H14" /></svg>;
}

function SectionExtraContent({ step }) {
  if (step.kind === "services") {
    return (
      <div className="grid grid-cols-3 gap-2 pt-2 sm:gap-3">
        {step.items.map((item) => (
          <div key={item.id} className="flex flex-col items-center justify-start gap-1 text-[11px] text-primary/80 sm:text-[12px]">
            <span className="inline-flex h-7 w-7 items-center justify-center bg-primary/10 text-primary">
              <ServiceGlyph id={item.id} />
            </span>
            <div className="min-w-0 text-center leading-5">
              <p className="font-semibold">{item.title}</p>
              <p className="text-[10px] text-primary/55">{item.subtitle}</p>
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (step.kind === "tracking") {
    return (
      <div className="mx-auto grid w-full max-w-[520px] grid-cols-2 gap-x-4 gap-y-3 pt-2">
        {step.points.map((point, index) => (
          <div key={point} className="flex h-12 items-center justify-start gap-2 text-[12px] text-primary/80">
            <span className="inline-flex h-7 w-7 items-center justify-center bg-primary/10 text-primary">
              <TrackingGlyph index={index} />
            </span>
            <span className="text-start leading-5">{point}</span>
          </div>
        ))}
      </div>
    );
  }

  return null;
}

function StoryChapterContent({ step, contentDir = "rtl" }) {
  return (
    <div className="relative isolate min-h-[220px] py-10 text-center sm:min-h-[260px] sm:py-12 mt-10" dir={contentDir}>
      <div className="relative z-10 space-y-3">
        <div className="relative">
          <span
            aria-hidden
            className={`pointer-events-none absolute left-1/2 bottom-10 z-0 -translate-x-1/2 select-none text-center font-display text-[clamp(6.2rem,24vw,12rem)] font-black leading-[0.72] tracking-tight text-accent/60 [text-shadow:0_10px_20px_rgba(165,141,102,0.14)] sm:text-[clamp(5.6rem,14vw,11.6rem)] ${step.kind === "projects" ? "mt-10 translate-y-8" : ""}`}
          >
            {step.rail}
          </span>
          <h3 className="relative z-10 font-display text-2xl leading-tight text-primary sm:text-3xl">{step.title}</h3>
        </div>
        <p className="text-[14px] leading-7 text-primary/82 sm:text-[15px] sm:leading-8">{step.body}</p>
        <SectionExtraContent step={step} />
      </div>
    </div>
  );
}

const getLocalizedProjectDescription = (project, lang) => {
  if (!project) return "";
  if (lang === "fa") return project.description_fa || project.description_en || project.description_ar || project.description || "";
  if (lang === "ar") return project.description_ar || project.description_en || project.description_fa || project.description || "";
  return project.description_en || project.description_fa || project.description_ar || project.description || "";
};

const getLocalizedProjectTitle = (project, lang) => {
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
    if (firstLine) return firstLine.length > 60 ? `${firstLine.slice(0, 60)}...` : firstLine;
  }

  const genericLabel = lang === "ar" ? "مشروع" : lang === "en" ? "Project" : "پروژه";
  return `${genericLabel} ${project.id || ""}`.trim();
};

const getProjectRecency = (project) => {
  if (!project) return 0;
  const dateValue = project.updated_at || project.updatedAt || project.created_at || project.createdAt || "";
  const timestamp = Date.parse(dateValue);
  if (Number.isFinite(timestamp) && timestamp > 0) return timestamp;
  const numericId = Number(project.id);
  return Number.isFinite(numericId) ? numericId : 0;
};

const projectsUiByLang = {
  fa: {
    loading: "در حال بارگذاری پروژه‌ها...",
    empty: "پروژه‌ای برای نمایش وجود ندارد.",
    noImage: "بدون تصویر",
    projectLabel: "پروژه"
  },
  en: {
    loading: "Loading projects...",
    empty: "No projects to display.",
    noImage: "No image",
    projectLabel: "Project"
  },
  ar: {
    loading: "جاري تحميل المشاريع...",
    empty: "لا توجد مشاريع للعرض.",
    noImage: "بدون صورة",
    projectLabel: "مشروع"
  }
};

function ProjectsShowcaseSection({
  step,
  contentDir,
  lang,
  projects,
  loading
}) {
  const ui = projectsUiByLang[lang] || projectsUiByLang.fa;
  const marqueeShiftPercent = (projects.length * 100) / 3;
  const marqueeDuration = `${Math.max(22, projects.length * 5)}s`;
  const loopProjects = projects.length > 0 ? [...projects, ...projects] : [];

  return (
    <article className="relative flex min-h-[62dvh] items-center mt-20">
      <div className="mx-auto w-full max-w-[1060px] space-y-1 px-2 sm:px-4" dir={contentDir}>
        <div className="-mb-10 sm:-mb-12">
          <StoryChapterContent step={step} contentDir={contentDir} />
        </div>

        {loading ? (
          <p className="text-center text-sm text-primary/65">{ui.loading}</p>
        ) : projects.length === 0 ? (
          <p className="text-center text-sm text-primary/65">{ui.empty}</p>
        ) : (
          <div className="space-y-2">
            <div className="overflow-hidden border border-primary/12 bg-primary/3">
              <div
                className="about-projects-marquee flex"
                style={{
                  "--about-marquee-shift": `${marqueeShiftPercent}%`,
                  "--about-marquee-duration": marqueeDuration
                }}
                dir="ltr"
              >
                {loopProjects.map((project, idx) => {
                  const title = getLocalizedProjectTitle(project, lang);

                  return (
                    <Link key={`${project.id}-${idx}`} to="/projects" className="block w-1/3 shrink-0 p-1.5">
                      <article className="group relative aspect-square overflow-hidden">
                        <div className="relative h-full w-full overflow-hidden">
                          {project.cover_image_url ? (
                            <img
                              src={resolveImageUrl(project.cover_image_url)}
                              alt={title}
                              loading="lazy"
                              decoding="async"
                              className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.02]"
                            />
                          ) : (
                            <div className="flex h-full items-center justify-center text-sm text-primary/55">
                              {ui.noImage}
                            </div>
                          )}
                          <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/66 via-black/20 to-transparent" />
                          <div className="absolute inset-x-0 bottom-0 p-2 text-center text-white sm:p-3">
                            {/* <h4 className="line-clamp-2 text-[11px] font-semibold leading-5 sm:text-sm">{title}</h4> */}
                          </div>
                        </div>
                      </article>
                    </Link>
                  );
                })}
              </div>
            </div>

            <div className="flex justify-center">
              <Link to="/projects" className="group inline-flex">
                <div data-home-anim="item" className="mt-7 flex items-center justify-center gap-3">
                  <span className="rounded-full border border-primary/25 bg-primary/10 px-4 py-2 text-[10px] font-semibold uppercase tracking-[0.24em] text-primary backdrop-blur-[3px] sm:px-5 sm:py-2.5 sm:text-xs">
                    {step.cta}
                  </span>
                  <span className="text-xl text-primary/75 transition-transform duration-500 group-hover:translate-x-1">
                    →
                  </span>
                </div>
              </Link>
            </div>
          </div>
        )}
      </div>
    </article>
  );
}

function StoryTimelineStage({ children }) {
  return (
    <section className="relative py-8" dir="ltr">
      <div className="relative mx-auto w-full max-w-[1280px] px-5 sm:px-8 lg:px-10">
        <div className="relative z-10">{children}</div>
      </div>
    </section>
  );
}

function StorySection({
  image,
  imageAlt,
  mediaSide = "right",
  accentBlock,
  imageRatioClass = "aspect-[5/4]",
  imageSizeClass = "max-w-[520px]",
  className = "",
  children,
}) {
  const imageOnRight = mediaSide === "right";
  const overlapClass = imageOnRight ? "lg:-translate-x-[10%]" : "lg:-translate-x-[90%]";
  const shapeOffsetClass = imageOnRight ? "-right-6" : "-left-6";
  const mediaPlacementClass = imageOnRight ? "lg:order-2" : "lg:order-1";
  const textPlacementClass = imageOnRight ? "lg:order-1 lg:justify-end" : "lg:order-2 lg:justify-start";
  const textOffsetClass = imageOnRight ? "lg:-translate-x-10" : "lg:translate-x-10";
  const desktopImagePin = `lg:absolute lg:left-1/2 lg:top-1/2 lg:z-10 lg:-translate-y-1/2 ${overlapClass}`;

  return (
    <article className={`relative flex min-h-[62dvh] items-center py-8 sm:py-10 ${className}`}>
      <span className={`absolute top-1/2 z-10 h-px w-20 bg-primary/25 ${imageOnRight ? "left-1/2" : "right-1/2"}`} />

      <div className="grid w-full grid-cols-1 items-center gap-10 lg:grid-cols-2 lg:gap-0">
        <div className={`flex items-center justify-center ${mediaPlacementClass}`}>
          <div className={`relative w-full ${imageSizeClass} ${desktopImagePin}`}>
            <div className={`absolute -top-6 ${shapeOffsetClass} h-[76%] w-[72%] ${accentBlock}`} />

            <div className="relative overflow-hidden shadow-[0_26px_64px_rgba(8,58,79,0.16)]">
              <img
                src={image}
                alt={imageAlt}
                className={`h-full w-full object-cover ${imageRatioClass}`}
                loading="lazy"
              />
              <div className={`pointer-events-none absolute inset-0 ${imageOnRight ? "bg-gradient-to-l" : "bg-gradient-to-r"} from-black/22 via-transparent to-transparent`} />
              <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/20 via-transparent to-transparent" />
            </div>
          </div>
        </div>

        <div className={`relative z-30 flex items-center justify-center ${textPlacementClass}`}>
          <div className={`w-full max-w-[490px] text-center transition-transform lg:px-5 ${textOffsetClass}`}>{children}</div>
        </div>
      </div>
    </article>
  );
}

export default function About() {
  const { lang } = useTranslation();
  const data = useMemo(() => content[lang] || content.fa, [lang]);
  const contentDir = lang === "fa" || lang === "ar" ? "rtl" : "ltr";
  const [recentProjects, setRecentProjects] = useState([]);
  const [projectsLoading, setProjectsLoading] = useState(true);
  const stepRefs = useRef([]);
  const scrollLockRef = useRef(false);
  const lockTimeoutRef = useRef(null);
  const scrollRafRef = useRef(null);

  const storySteps = useMemo(
    () => [
      ...data.chapters.map((chapter) => ({ ...chapter, kind: "chapter" })),
      { ...data.services, id: "services", kind: "services" },
      { ...data.tracking, id: "tracking", kind: "tracking" },
      { ...data.projects, id: "projects", kind: "projects" }
    ],
    [data]
  );

  useEffect(() => {
    if (typeof window === "undefined") return undefined;

    const navOffset = 96;
    const settlePadding = 12;
    const easeOutCubic = (t) => 1 - Math.pow(1 - t, 3);

    const getStepElements = () => stepRefs.current.filter(Boolean);

    const getStepTops = () =>
      getStepElements().map((el) => Math.max(0, Math.round(el.getBoundingClientRect().top + window.scrollY)));

    const findCurrentStepIndex = () => {
      const tops = getStepTops();
      if (!tops.length) return 0;

      const marker = window.scrollY + navOffset + settlePadding;
      let index = 0;
      for (let i = 0; i < tops.length; i += 1) {
        if (marker >= tops[i]) index = i;
        else break;
      }
      return index;
    };

    const lockFor = (ms) => {
      scrollLockRef.current = true;
      if (lockTimeoutRef.current) {
        window.clearTimeout(lockTimeoutRef.current);
      }
      lockTimeoutRef.current = window.setTimeout(() => {
        scrollLockRef.current = false;
      }, ms);
    };

    const animateTo = (targetY, durationMs) => {
      const startY = window.scrollY;
      const deltaY = targetY - startY;

      if (Math.abs(deltaY) < 2) {
        window.scrollTo(0, targetY);
        return;
      }

      if (scrollRafRef.current) {
        window.cancelAnimationFrame(scrollRafRef.current);
      }

      const startedAt = performance.now();

      if (Math.abs(deltaY) > 40) {
        const kickY = Math.round(startY + (deltaY * 0.035));
        window.scrollTo(0, kickY);
      }

      const step = (now) => {
        const elapsed = now - startedAt;
        const progress = Math.min(1, elapsed / durationMs);
        const eased = easeOutCubic(progress);
        const nextY = Math.round(startY + (deltaY * eased));
        window.scrollTo(0, nextY);

        if (progress < 1) {
          scrollRafRef.current = window.requestAnimationFrame(step);
          return;
        }
        scrollRafRef.current = null;
      };

      scrollRafRef.current = window.requestAnimationFrame(step);
    };

    const scrollToStep = (targetIndex) => {
      const tops = getStepTops();
      if (!tops.length) return false;
      if (targetIndex < 0 || targetIndex >= tops.length) return false;

      const targetY = Math.max(0, tops[targetIndex] - navOffset);
      const distance = Math.abs(targetY - window.scrollY);
      const durationMs = Math.min(1800, Math.max(950, Math.round(distance * 0.9)));
      const lockMs = durationMs + 180;

      animateTo(targetY, durationMs);
      lockFor(lockMs);
      return true;
    };

    const stepBy = (direction) => {
      const current = findCurrentStepIndex();
      const next = current + direction;
      return scrollToStep(next);
    };

    const onWheel = (event) => {
      if (scrollLockRef.current) {
        event.preventDefault();
        return;
      }

      if (Math.abs(event.deltaY) < 8) return;
      const moved = stepBy(event.deltaY > 0 ? 1 : -1);
      if (moved) event.preventDefault();
    };

    const onKeyDown = (event) => {
      const activeTag = event.target?.tagName;
      if (activeTag && ["INPUT", "TEXTAREA", "SELECT"].includes(activeTag)) return;

      if (scrollLockRef.current && (event.key === "ArrowDown" || event.key === "ArrowUp")) {
        event.preventDefault();
        return;
      }

      if (event.key === "ArrowDown") {
        const moved = stepBy(1);
        if (moved) event.preventDefault();
      } else if (event.key === "ArrowUp") {
        const moved = stepBy(-1);
        if (moved) event.preventDefault();
      }
    };

    window.addEventListener("wheel", onWheel, { passive: false });
    window.addEventListener("keydown", onKeyDown);

    return () => {
      if (scrollRafRef.current) {
        window.cancelAnimationFrame(scrollRafRef.current);
      }
      if (lockTimeoutRef.current) {
        window.clearTimeout(lockTimeoutRef.current);
      }
      scrollRafRef.current = null;
      scrollLockRef.current = false;
      window.removeEventListener("wheel", onWheel);
      window.removeEventListener("keydown", onKeyDown);
    };
  }, []);

  useEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return;

    const seo = seoContent[lang] || seoContent.fa;
    const pageUrl = getCanonicalUrl("/about");
    const heroImage = data.hero.image;
    const heroImageUrl = getAbsoluteUrl(heroImage);
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
      const scriptId = "about-jsonld";
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
    upsertMeta('meta[property="og:image"]', { property: "og:image" }, heroImageUrl);
    upsertMeta('meta[property="og:locale"]', { property: "og:locale" }, seo.locale);
    upsertMeta('meta[name="twitter:card"]', { name: "twitter:card" }, "summary_large_image");
    upsertMeta('meta[name="twitter:title"]', { name: "twitter:title" }, seo.title);
    upsertMeta('meta[name="twitter:description"]', { name: "twitter:description" }, seo.description);
    upsertMeta('meta[name="twitter:image"]', { name: "twitter:image" }, heroImageUrl);

    upsertJsonLd({
      "@context": "https://schema.org",
      "@type": "AboutPage",
      inLanguage: lang,
      name: seo.title,
      description: seo.description,
      url: pageUrl,
      mainEntity: {
        "@type": "Organization",
        name: "SangeHassan",
        url: getSiteOrigin()
      }
    });

    return () => {
      document.title = previousTitle;
      cleanups.reverse().forEach((fn) => fn());
    };
  }, [lang, data.hero.image]);

  useEffect(() => {
    let active = true;

    const loadProjects = async () => {
      try {
        const response = await fetchJSON("/api/projects");
        if (!active) return;

        const cards = Array.isArray(response.data) ? response.data : [];
        const recent = [...cards]
          .sort((a, b) => getProjectRecency(b) - getProjectRecency(a))
          .slice(0, 5);

        if (recent.length === 0) {
          setRecentProjects([]);
          setProjectsLoading(false);
          return;
        }

        const details = await Promise.allSettled(recent.map((project) => fetchJSON(`/api/projects/${project.id}`)));
        if (!active) return;

        const merged = recent.map((project, index) => {
          const detail = details[index];
          if (detail.status !== "fulfilled" || !detail.value?.data) return project;
          return {
            ...project,
            ...detail.value.data,
            cover_image_url: detail.value.data.cover_image_url || project.cover_image_url
          };
        });

        setRecentProjects(merged);
      } catch (_) {
        if (!active) return;
        setRecentProjects([]);
      } finally {
        if (active) setProjectsLoading(false);
      }
    };

    loadProjects();
    return () => {
      active = false;
    };
  }, []);

  return (
    <div className="bg-sand text-primary">
      <div ref={(el) => { stepRefs.current[0] = el; }} className="scroll-mt-24">
        <section className="section-shell">
          <div className="overflow-hidden">
            <div className="relative">
              <div className="relative h-[14rem] w-full sm:h-[14rem] md:h-[16rem] lg:h-[20rem]">
                <img
                  src={data.hero.image}
                  alt={data.hero.title}
                  className="h-full w-full object-cover"
                  loading="eager"
                  fetchpriority="high"
                  decoding="async"
                />
                <div className="pointer-events-none absolute inset-x-0 bottom-0 h-28 bg-gradient-to-b from-transparent via-sand/58 to-sand" />
              </div>

              <div className="px-6 sm:px-8 lg:px-12" dir={contentDir}>
                <div className="mx-auto max-w-4xl space-y-3 text-center">
                  <h1 className="font-display text-3xl text-primary sm:text-4xl">{data.hero.title}</h1>
                  <p className="mx-auto max-w-3xl text-xs leading-6 text-primary/78 sm:text-sm sm:leading-7">
                    {data.hero.body}
                  </p>
                  <div className="flex flex-wrap items-center justify-center gap-2.5 text-xs font-semibold">
                    {data.hero.chips.map((chip) => (
                      <span
                        key={chip}
                        className="rounded-full bg-accent/20 px-4 py-2 text-[10px] font-semibold uppercase tracking-[0.24em] text-accent backdrop-blur-[3px] sm:px-5 sm:py-2.5 sm:text-xs"
                      >
                        {chip}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <StoryTimelineStage>
        {storySteps.map((step, idx) => {
          const stepNodeIndex = idx + 1;

          return (
            <div key={step.id} ref={(el) => { stepRefs.current[stepNodeIndex] = el; }} className="scroll-mt-24">
              {step.kind === "projects" ? (
                <ProjectsShowcaseSection
                  step={step}
                  contentDir={contentDir}
                  lang={lang}
                  projects={recentProjects}
                  loading={projectsLoading}
                />
              ) : (
                <StorySection
                  image={step.image}
                  imageAlt={step.title}
                  mediaSide={idx % 2 === 0 ? "right" : "left"}
                  accentBlock={step.shape}
                  imageSizeClass={step.kind === "services" ? "max-w-[600px]" : undefined}
                  imageRatioClass={step.kind === "services" ? "aspect-[3/1]" : undefined}
                >
                  <StoryChapterContent step={step} contentDir={contentDir} />
                </StorySection>
              )}
            </div>
          );
        })}
      </StoryTimelineStage>
    </div>
  );
}
