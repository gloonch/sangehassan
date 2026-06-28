import { ArrowLeft, ArrowRight, Home, SearchX } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import { getLanguageFromPath } from "../lib/i18n";
import { usePageSeo } from "../lib/seo";

const content = {
  en: {
    eyebrow: "Page not found",
    title: "This stone path does not exist",
    body:
      "The address may have changed, or the page may no longer be available. Return to the homepage and continue from a verified route.",
    home: "Back to homepage",
    products: "View products",
    codeLabel: "Error 404",
    locale: "en_US",
    dir: "ltr",
    productsPath: "/en/products"
  },
  fa: {
    eyebrow: "صفحه پیدا نشد",
    title: "این مسیر در سنگ حسن وجود ندارد",
    body:
      "ممکن است آدرس تغییر کرده باشد یا این صفحه دیگر در دسترس نباشد. از صفحه اصلی، مسیر درست را دوباره انتخاب کنید.",
    home: "بازگشت به صفحه اصلی",
    products: "مشاهده محصولات",
    codeLabel: "خطای ۴۰۴",
    locale: "fa_IR",
    dir: "rtl",
    productsPath: "/fa/products"
  },
  ar: {
    eyebrow: "الصفحة غير موجودة",
    title: "هذا المسار غير موجود في SangeHassan",
    body:
      "ربما تغير العنوان أو لم تعد هذه الصفحة متاحة. ارجع إلى الصفحة الرئيسية واختر المسار الصحيح.",
    home: "العودة إلى الرئيسية",
    products: "عرض المنتجات",
    codeLabel: "خطأ 404",
    locale: "ar_SA",
    dir: "rtl",
    productsPath: "/ar/products"
  }
};

export default function NotFound() {
  const location = useLocation();
  const lang = getLanguageFromPath(location.pathname);
  const copy = content[lang] || content.en;
  const ArrowIcon = copy.dir === "rtl" ? ArrowLeft : ArrowRight;

  usePageSeo({
    title: `${copy.eyebrow} | SangeHassan`,
    description: copy.body,
    path: location.pathname || "/404",
    lang,
    locale: copy.locale,
    robots: "noindex,follow,noarchive"
  });

  return (
    <section
      className="relative isolate flex min-h-[calc(100svh-7rem)] items-center overflow-hidden bg-sand py-14 text-primary sm:py-20"
      dir={copy.dir}
      aria-labelledby="not-found-title"
    >
      <span className="bg-quiet-grid pointer-events-none absolute inset-0 opacity-[0.42]" aria-hidden="true" />
      <span
        className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(229,225,221,0.82),rgba(255,255,255,0.42)_48%,rgba(229,225,221,0.92))]"
        aria-hidden="true"
      />
      <span
        className="pointer-events-none absolute inset-x-0 top-0 h-px bg-primary/10"
        aria-hidden="true"
      />

      <div className="section-shell relative z-10">
        <div className="mx-auto grid max-w-5xl items-center gap-10 lg:grid-cols-[0.9fr_1.1fr] lg:gap-14">
          <div className="relative mx-auto flex aspect-[4/3] w-full max-w-[24rem] items-center justify-center overflow-hidden rounded-lg bg-white/28 shadow-[0_28px_80px_rgba(8,58,79,0.12)] backdrop-blur-xl">
            <span className="absolute inset-x-8 top-8 h-px bg-white/60" aria-hidden="true" />
            <span className="absolute inset-y-8 left-8 w-px bg-white/50" aria-hidden="true" />
            <span className="absolute inset-x-8 bottom-8 h-px bg-primary/10" aria-hidden="true" />
            <span className="absolute inset-y-8 right-8 w-px bg-primary/10" aria-hidden="true" />
            <SearchX className="relative z-10 h-16 w-16 text-accent sm:h-20 sm:w-20" strokeWidth={1.35} aria-hidden="true" />
            <span
              className="pointer-events-none absolute bottom-4 font-display text-[4.8rem] font-semibold leading-none text-primary/[0.08] sm:text-[6.6rem]"
              aria-hidden="true"
            >
              404
            </span>
          </div>

          <div className="text-center lg:text-start">
            <p className="text-xs font-semibold uppercase tracking-[0.32em] text-accent">{copy.eyebrow}</p>
            <h1 id="not-found-title" className="mt-5 font-display text-4xl leading-tight text-primary sm:text-5xl lg:text-6xl">
              {copy.title}
            </h1>
            <p className="mx-auto mt-5 max-w-2xl text-sm leading-8 text-primary/68 sm:text-base lg:mx-0">
              {copy.body}
            </p>

            <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row lg:justify-start">
              <Link
                to="/"
                className="inline-flex h-12 min-w-[13rem] items-center justify-center gap-2 rounded-full bg-primary px-6 text-sm font-semibold text-sand transition hover:bg-primary/90 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/40"
              >
                <Home className="h-4 w-4" aria-hidden="true" />
                <span>{copy.home}</span>
              </Link>
              <Link
                to={copy.productsPath}
                className="inline-flex h-12 min-w-[13rem] items-center justify-center gap-2 rounded-full border border-primary/20 bg-white/35 px-6 text-sm font-semibold text-primary backdrop-blur-xl transition hover:border-primary/45 hover:bg-white/55 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/25"
              >
                <span>{copy.products}</span>
                <ArrowIcon className="h-4 w-4" aria-hidden="true" />
              </Link>
            </div>

            <p className="mt-8 text-xs font-semibold uppercase tracking-[0.26em] text-primary/35">{copy.codeLabel}</p>
          </div>
        </div>
      </div>
    </section>
  );
}
