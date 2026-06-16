import { useEffect, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { ArrowLeft, ArrowRight } from "lucide-react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveVersionedImageUrl } from "../lib/assets";
import { getCanonicalUrl, usePageSeo } from "../lib/seo";
import { usePrerenderData } from "../lib/prerenderData";

const localeContent = {
  fa: { title: "مقالات سنگ طبیعی", intro: "راهنماهای کاربردی برای انتخاب، فرآوری، اجرا و نگهداری سنگ طبیعی.", empty: "مقاله‌ای پیدا نشد.", read: "مطالعه مقاله", min: "دقیقه", locale: "fa_IR" },
  en: { title: "Natural stone articles", intro: "Practical guidance for choosing, processing, installing and maintaining natural stone.", empty: "No articles found.", read: "Read article", min: "min read", locale: "en_US" },
  ar: { title: "مقالات الحجر الطبيعي", intro: "أدلة عملية لاختيار الحجر الطبيعي ومعالجته وتركيبه وصيانته.", empty: "لم يتم العثور على مقالات.", read: "قراءة المقال", min: "دقيقة", locale: "ar_SA" }
};

const pageSize = 9;
const freshRequest = () => ({
  cache: "no-store",
  headers: {
    "Cache-Control": "no-cache",
    Pragma: "no-cache"
  }
});

export default function Blogs() {
  const { lang } = useTranslation();
  const { locale: routeLocale } = useParams();
  const locale = localeContent[routeLocale] ? routeLocale : lang;
  const copy = localeContent[locale] || localeContent.en;
  const initialBlogs = usePrerenderData("blogs");
  const [blogs, setBlogs] = useState(() => initialBlogs || []);
  const [loading, setLoading] = useState(!initialBlogs);
  const [searchParams] = useSearchParams();
  const page = Math.max(1, Number(searchParams.get("page")) || 1);
  const basePath = `/${locale}/blogs`;
  const isRTL = locale === "fa" || locale === "ar";

  useEffect(() => {
    let active = true;

    const loadFreshBlogs = (showLoading = false) => {
      if (showLoading) setLoading(true);
      fetchJSON(`/api/blogs?locale=${locale}&_=${Date.now()}`, freshRequest())
        .then((response) => { if (active) setBlogs(response.data || []); })
        .catch(() => { if (active) setBlogs([]); })
        .finally(() => { if (active) setLoading(false); });
    };

    const refreshWhenVisible = () => {
      if (typeof document === "undefined" || document.visibilityState === "visible") {
        loadFreshBlogs(false);
      }
    };

    loadFreshBlogs(!initialBlogs);
    window.addEventListener("focus", refreshWhenVisible);
    document.addEventListener("visibilitychange", refreshWhenVisible);
    return () => {
      active = false;
      window.removeEventListener("focus", refreshWhenVisible);
      document.removeEventListener("visibilitychange", refreshWhenVisible);
    };
  }, [initialBlogs, locale]);

  const pageCount = Math.max(1, Math.ceil(blogs.length / pageSize));
  const visible = blogs.slice((page - 1) * pageSize, page * pageSize);
  const coverBlog = blogs.find((blog) => blog.cover_image_url);
  const coverImage = coverBlog?.cover_image_url || "";
  const coverImageVersion = coverBlog?.updated_at || coverBlog?.published_at || "";
  const path = page > 1 ? `${basePath}?page=${page}` : basePath;
  const alternates = ["fa", "en", "ar"].map((code) => ({ lang: code, path: `/${code}/blogs` }));
  alternates.push({ lang: "x-default", path: "/en/blogs" });

  usePageSeo({
    title: `${copy.title} | SangeHassan`,
    description: copy.intro,
    path,
    lang: locale,
    locale: copy.locale,
    image: coverImage ? resolveVersionedImageUrl(coverImage, coverImageVersion) : "",
    robots: "index,follow",
    alternates,
    jsonLdId: "blogs-jsonld",
    jsonLd: { "@context": "https://schema.org", "@type": "Blog", inLanguage: locale, name: copy.title, description: copy.intro, url: getCanonicalUrl(basePath) }
  });

  return (
    <section className="section-shell pb-16 pt-10" dir={isRTL ? "rtl" : "ltr"}>
      <header className="border-b border-primary/15 pb-8">
        <h1 className="font-display text-4xl md:text-5xl">{copy.title}</h1>
        <p className="mt-4 max-w-2xl text-base leading-8 text-primary/65">{copy.intro}</p>
      </header>

      {loading ? <p className="py-16 text-sm text-primary/60">...</p> : visible.length === 0 ? <p className="py-16 text-sm text-primary/60">{copy.empty}</p> : (
        <div className="grid gap-x-7 gap-y-10 py-10 md:grid-cols-2 lg:grid-cols-3">
          {visible.map((blog) => (
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
          const target = number === 1 ? basePath : `${basePath}?page=${number}`;
          return <Link key={number} to={target} className={`inline-flex h-10 w-10 items-center justify-center border text-sm ${number === page ? "border-primary bg-primary text-white" : "border-primary/20"}`}>{number}</Link>;
        })}
      </nav>}
    </section>
  );
}
