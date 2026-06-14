import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { getCanonicalUrl, usePageSeo } from "../lib/seo";
import { usePrerenderData } from "../lib/prerenderData";
import { useTranslation } from "../lib/i18n";
import {
  catalogAlternates,
  catalogBasePath,
  catalogCopy,
  catalogLocaleConfig,
  localizedField
} from "../lib/catalogLocale";

const fallbackSEO = {
  en: {
    title: "Natural Stone Categories | SangeHassan",
    description: "Browse natural stone categories and find the right stone for architectural and building projects.",
    h1: "Natural stone categories",
    intro: "Choose a stone category to view available products, colors, finishes and applications."
  },
  fa: {
    title: "انواع سنگ ساختمانی | تراورتن، گرانیت، مرمریت و سنگ چینی",
    description: "مشاهده دسته‌بندی انواع سنگ ساختمانی و طبیعی و انتخاب سنگ مناسب پروژه.",
    h1: "دسته‌بندی انواع سنگ ساختمانی",
    intro: "برای مشاهده محصولات و مشخصات هر نوع سنگ، دسته‌بندی موردنظر را انتخاب کنید."
  },
  ar: {
    title: "أنواع الحجر الطبيعي | سانج حسن",
    description: "تصفح فئات الحجر الطبيعي واختيار الحجر المناسب للمشاريع المعمارية والإنشائية.",
    h1: "فئات الحجر الطبيعي",
    intro: "اختر فئة الحجر لعرض المنتجات والألوان والتشطيبات والاستخدامات المتاحة."
  }
};

const categoryBackgrounds = {
  accessories: "/images/category_images/accessories_category.jpg",
  basalt: "/images/category_images/basalt_category.jpg",
  crystal: "/images/category_images/crystal_category.jpg",
  finishings: "/images/category_images/finishings_category.jpg",
  granite: "/images/category_images/granite_category.jpg",
  "imported-stones": "/images/category_images/imported_stones_category.png",
  onyx: "/images/category_images/onyx_category.jpg",
  sandstone: "/images/category_images/sandstone_category.jpg",
  travertine: "/images/category_images/travertine_category.jpg"
};

const getCategoryBackground = (category) =>
  categoryBackgrounds[category.slug] || category.preview_image || category.image_url || "";

export default function ProductHub() {
  const { lang } = useTranslation();
  const config = catalogLocaleConfig[lang] || catalogLocaleConfig.en;
  const copy = catalogCopy[lang] || catalogCopy.en;
  const basePath = catalogBasePath(lang);
  const prerendered = usePrerenderData("catalogHub");
  const initialHub = prerendered?.locale === lang ? prerendered : null;
  const [hub, setHub] = useState(initialHub);
  const [loading, setLoading] = useState(!initialHub);

  useEffect(() => {
    if (initialHub) return undefined;
    let active = true;
    setLoading(true);
    fetchJSON(`/api/catalog/categories?locale=${lang}`)
      .then((response) => {
        if (active) setHub(response.data || null);
      })
      .catch(() => {
        if (active) setHub(null);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [initialHub, lang]);

  const seo = { ...(fallbackSEO[lang] || fallbackSEO.en), ...(hub?.seo || {}), canonical: basePath, robots: "index,follow" };
  const categories = hub?.categories || [];
  const alternates = catalogAlternates();
  const jsonLd = useMemo(() => ({
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "CollectionPage",
        "@id": `${getCanonicalUrl(basePath)}#webpage`,
        url: getCanonicalUrl(basePath),
        name: seo.title,
        description: seo.description,
        inLanguage: lang
      },
      {
        "@type": "BreadcrumbList",
        itemListElement: [
          { "@type": "ListItem", position: 1, name: copy.home, item: getCanonicalUrl("/") },
          { "@type": "ListItem", position: 2, name: copy.products, item: getCanonicalUrl(basePath) }
        ]
      }
    ]
  }), [basePath, copy, lang, seo.description, seo.title]);

  usePageSeo({
    title: seo.title,
    description: seo.description,
    path: basePath,
    lang,
    locale: config.locale,
    robots: seo.robots,
    alternates,
    jsonLd,
    jsonLdId: "catalog-hub-jsonld"
  });

  return (
    <section className="section-shell pb-16 pt-12" dir={config.dir}>
      <nav className="mb-8 text-sm text-primary/60" aria-label={copy.breadcrumb}>
        <Link to="/" className="hover:text-primary">{copy.home}</Link>
        <span className="px-2" aria-hidden="true">/</span>
        <span>{copy.products}</span>
      </nav>

      <header className="mx-auto max-w-3xl text-center">
        <h1 className="font-display text-3xl leading-tight md:text-5xl">{seo.h1}</h1>
        <p className="mx-auto mt-5 max-w-2xl text-sm leading-8 text-primary/70 md:text-base">{seo.intro}</p>
      </header>

      {loading ? (
        <div className="mx-auto mt-12 grid max-w-6xl sm:grid-cols-3" aria-hidden="true">
          {Array.from({ length: 6 }).map((_, index) => <div key={index} className="h-20 animate-pulse bg-primary/10 md:h-24" />)}
        </div>
      ) : categories.length === 0 ? (
        <p className="mt-16 text-center text-sm text-primary/60">{copy.emptyCategories}</p>
      ) : (
        <div className="mx-auto mt-12 grid max-w-6xl sm:grid-cols-3">
          {categories.map((category) => {
            const image = getCategoryBackground(category);
            const title = localizedField(category, "title", lang);
            return (
              <Link
                key={category.id}
                to={`${basePath}/${category.slug}`}
                state={{ catalogRouteKind: "category" }}
                className="group relative flex h-20 items-center overflow-hidden bg-primary text-white transition focus:outline-none focus-visible:z-10 focus-visible:ring-2 focus-visible:ring-white md:h-24"
              >
                {image ? <img src={resolveImageUrl(image)} alt="" loading="lazy" decoding="async" className="absolute inset-0 h-full w-full object-cover transition duration-500 group-hover:scale-105" /> : null}
                <div className="absolute inset-0 bg-gradient-to-r from-primary/60 via-primary/35 to-primary/15 rtl:bg-gradient-to-l" />
                <div className="relative z-10 flex w-full items-center justify-between gap-5 px-6 md:px-8">
                  <h2 className="min-w-0 truncate text-sm font-bold text-white lg:text-base">{title}</h2>
                  <p className="shrink-0 text-xs font-bold text-white lg:text-sm">{category.product_count.toLocaleString(config.numberLocale)} {copy.productCount}</p>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </section>
  );
}
