import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, ArrowRight } from "lucide-react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveVersionedImageUrl } from "../lib/assets";
import { getCanonicalUrl, usePageSeo } from "../lib/seo";
import { usePrerenderData } from "../lib/prerenderData";
import { blogHubAlternates, defaultBlogLocale, normalizeBlogLocales } from "../lib/blogLocales";

const localeContent = {
  fa: { title: "مقالات سنگ طبیعی", intro: "راهنماهای کاربردی برای انتخاب، فرآوری، اجرا و نگهداری سنگ طبیعی.", empty: "مقاله‌ای پیدا نشد.", read: "مطالعه مقاله", min: "دقیقه", locale: "fa_IR", persianArticles: "مقالات فارسی" },
  en: { title: "Natural stone articles", intro: "Practical guidance for choosing, processing, installing and maintaining natural stone.", empty: "No articles have been published in English yet.", read: "Read article", min: "min read", locale: "en_US", persianArticles: "View published Persian articles" },
  ar: { title: "مقالات الحجر الطبيعي", intro: "أدلة عملية لاختيار الحجر الطبيعي ومعالجته وتركيبه وصيانته.", empty: "لم تنشر مقالات باللغة العربية بعد.", read: "قراءة المقال", min: "دقيقة", locale: "ar_SA", persianArticles: "عرض المقالات المنشورة بالفارسية" }
};

const pageSize = 9;
const pageDescription = (copy, locale, page) => {
  if (page === 1) return copy.intro;
  if (locale === "fa") return `${copy.intro} صفحه ${page}.`;
  if (locale === "ar") return `${copy.intro} الصفحة ${page}.`;
  return `${copy.intro} Page ${page}.`;
};
const freshRequest = () => ({
  cache: "no-store",
  headers: {
    "Cache-Control": "no-cache",
    Pragma: "no-cache"
  }
});

export default function Blogs() {
  const { lang } = useTranslation();
  const { locale: routeLocale, pageNumber } = useParams();
  const locale = localeContent[routeLocale] ? routeLocale : lang;
  const copy = localeContent[locale] || localeContent.en;
  const initialPage = usePrerenderData("blogPage");
  const page = Math.max(1, Number(pageNumber) || 1);
  const offset = (page - 1) * pageSize;
  const [blogPage, setBlogPage] = useState(() => initialPage || { items: [], total: 0, limit: pageSize, offset, availableLocales: [] });
  const [loading, setLoading] = useState(!initialPage);
  const basePath = `/${locale}/blogs`;
  const isRTL = locale === "fa" || locale === "ar";

  useEffect(() => {
    let active = true;

    const loadFreshBlogs = (showLoading = false) => {
      if (showLoading) setLoading(true);
      fetchJSON(`/api/blogs?locale=${locale}&limit=${pageSize}&offset=${offset}&_=${Date.now()}`, freshRequest())
        .then((response) => {
          if (!active) return;
          setBlogPage((current) => ({
            ...(response.data || { items: [], total: 0, limit: pageSize, offset }),
            availableLocales: current.availableLocales?.length
              ? current.availableLocales
              : initialPage?.availableLocales || ((response.data?.total || 0) > 0 ? [locale] : [])
          }));
        })
        .catch(() => {
          if (active) setBlogPage((current) => ({ items: [], total: 0, limit: pageSize, offset, availableLocales: current.availableLocales || [] }));
        })
        .finally(() => { if (active) setLoading(false); });
    };

    const refreshWhenVisible = () => {
      if (typeof document === "undefined" || document.visibilityState === "visible") {
        loadFreshBlogs(false);
      }
    };

    loadFreshBlogs(!initialPage || initialPage.offset !== offset);
    window.addEventListener("focus", refreshWhenVisible);
    document.addEventListener("visibilitychange", refreshWhenVisible);
    return () => {
      active = false;
      window.removeEventListener("focus", refreshWhenVisible);
      document.removeEventListener("visibilitychange", refreshWhenVisible);
    };
  }, [initialPage, locale, offset]);

  useEffect(() => {
    if (initialPage?.availableLocales?.length) return undefined;
    let active = true;
    Promise.all(Object.keys(localeContent).map(async (code) => {
      try {
        const response = await fetchJSON(`/api/blogs?locale=${code}&limit=1&offset=0`, freshRequest());
        return (response.data?.total || 0) > 0 ? code : "";
      } catch (_) {
        return "";
      }
    })).then((locales) => {
      if (active) setBlogPage((current) => ({ ...current, availableLocales: locales.filter(Boolean) }));
    });
    return () => { active = false; };
  }, [initialPage?.availableLocales]);

  const blogs = Array.isArray(blogPage.items) ? blogPage.items : [];
  const pageCount = Math.max(1, Math.ceil((blogPage.total || 0) / pageSize));
  const coverBlog = blogs.find((blog) => blog.cover_image_url);
  const coverImage = coverBlog?.cover_image_url || "";
  const coverImageVersion = coverBlog?.updated_at || coverBlog?.published_at || "";
  const path = page > 1 ? `${basePath}/page/${page}` : basePath;
  const availableLocales = normalizeBlogLocales(blogPage.availableLocales || []);
  const isIndexable = availableLocales.includes(locale) || (blogPage.total || 0) > 0;
  const alternates = page === 1 ? blogHubAlternates(availableLocales) : [];
  const fallbackLocale = defaultBlogLocale(availableLocales);
  const pageTitle = page > 1 ? `${copy.title} - ${page} | SangeHassan` : `${copy.title} | SangeHassan`;
  const description = pageDescription(copy, locale, page);

  usePageSeo({
    title: pageTitle,
    description,
    path,
    lang: locale,
    locale: copy.locale,
    image: coverImage ? resolveVersionedImageUrl(coverImage, coverImageVersion) : "",
    robots: isIndexable ? "index,follow" : "noindex,follow",
    alternates,
    previousPath: page === 2 ? basePath : page > 2 ? `${basePath}/page/${page - 1}` : "",
    nextPath: page < pageCount ? `${basePath}/page/${page + 1}` : "",
    jsonLdId: "blogs-jsonld",
    jsonLd: { "@context": "https://schema.org", "@type": "Blog", "@id": `${getCanonicalUrl(path)}#webpage`, inLanguage: locale, name: copy.title, description, url: getCanonicalUrl(path) }
  });

  return (
    <section className="section-shell pb-16 pt-10" dir={isRTL ? "rtl" : "ltr"}>
      <header className="border-b border-primary/15 pb-8">
        <h1 className="font-display text-4xl md:text-5xl">{copy.title}</h1>
        <p className="mt-4 max-w-2xl text-base leading-8 text-primary/65">{copy.intro}</p>
      </header>

      {loading ? <p className="py-16 text-sm text-primary/60">...</p> : blogs.length === 0 ? (
        <div className="py-16 text-sm leading-7 text-primary/60">
          <p>{copy.empty}</p>
          {locale !== "fa" && availableLocales.includes("fa") ? (
            <Link className="mt-3 inline-flex font-semibold text-accent underline underline-offset-4" to="/fa/blogs">
              {copy.persianArticles}
            </Link>
          ) : fallbackLocale && fallbackLocale !== locale ? (
            <Link className="mt-3 inline-flex font-semibold text-accent underline underline-offset-4" to={`/${fallbackLocale}/blogs`}>
              {localeContent[fallbackLocale]?.title || copy.persianArticles}
            </Link>
          ) : null}
        </div>
      ) : (
        <div className="grid gap-x-7 gap-y-10 py-10 md:grid-cols-2 lg:grid-cols-3">
          {blogs.map((blog) => (
            <article key={blog.id} className="group flex min-w-0 flex-col border-b border-primary/15 pb-7">
              <Link to={`${basePath}/${blog.slug}`} className="block aspect-[4/3] overflow-hidden bg-primary/5">
                {blog.cover_image_url && <img src={resolveVersionedImageUrl(blog.cover_image_url, blog.updated_at || blog.published_at || "")} alt={blog.featured_image_alt || blog.title} className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.025]" loading="lazy" />}
              </Link>
              <div className="mt-5 flex items-center gap-3 text-xs text-primary/45">
                <time dateTime={blog.published_at || blog.created_at}>{new Date(blog.published_at || blog.created_at).toLocaleDateString(locale)}</time>
                {blog.reading_time_minutes > 0 && <><span>·</span><span>{blog.reading_time_minutes} {copy.min}</span></>}
              </div>
              <h2 className="mt-3 font-display text-2xl leading-snug"><Link to={`${basePath}/${blog.slug}`}>{blog.title}</Link></h2>
              <p className="mt-3 line-clamp-3 text-sm leading-7 text-primary/65">{blog.excerpt}</p>
              <Link to={`${basePath}/${blog.slug}`} className="mt-5 inline-flex items-center gap-2 text-sm font-semibold text-accent underline decoration-accent/35 underline-offset-4">
                {copy.read}{isRTL ? <ArrowLeft size={16} /> : <ArrowRight size={16} />}
              </Link>
            </article>
          ))}
        </div>
      )}

      {pageCount > 1 && <nav className="flex flex-wrap gap-2 border-t border-primary/15 pt-7" aria-label="Pagination">
        {Array.from({ length: pageCount }, (_, index) => index + 1).map((number) => {
          const target = number === 1 ? basePath : `${basePath}/page/${number}`;
          return <Link key={number} to={target} className={`inline-flex h-10 w-10 items-center justify-center border text-sm ${number === page ? "border-primary bg-primary text-white" : "border-primary/20"}`}>{number}</Link>;
        })}
      </nav>}
    </section>
  );
}
