import { useEffect, useMemo, useState } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, ArrowRight, Clock3 } from "lucide-react";
import { fetchJSON } from "../lib/api";
import { appendImageVersion, resolveVersionedImageUrl } from "../lib/assets";
import { getCanonicalUrl, usePageSeo } from "../lib/seo";
import { usePrerenderData } from "../lib/prerenderData";
import { getLanguageFromPath } from "../lib/i18n";

const localeMeta = {
  fa: { locale: "fa_IR", articles: "مقالات", home: "خانه", updated: "به‌روزرسانی", min: "دقیقه مطالعه", toc: "در این مقاله", back: "بازگشت به مقالات" },
  en: { locale: "en_US", articles: "Articles", home: "Home", updated: "Updated", min: "min read", toc: "In this article", back: "Back to articles" },
  ar: { locale: "ar_SA", articles: "المقالات", home: "الرئيسية", updated: "آخر تحديث", min: "دقيقة قراءة", toc: "في هذا المقال", back: "العودة إلى المقالات" }
};

const faArticleSeoOverrides = {
  "everything-about-granite-stone": {
    title: "سنگ گرانیت چیست؟ ویژگی‌ها، کاربردها و انواع گرانیت | سنگ حسن",
    description: "راهنمای کامل سنگ گرانیت؛ بررسی ویژگی‌ها، مزایا، معایب، کاربردها، انواع رنگ و نکات مهم خرید گرانیت برای پروژه‌های ساختمانی.",
    h1: "سنگ گرانیت چیست؟ راهنمای کامل ویژگی‌ها، کاربردها و انواع گرانیت",
    canonical: "/fa/blogs/everything-about-granite-stone"
  },
  "everything-about-travertine-stone": {
    title: "سنگ تراورتن چیست؟ انواع، معادن، کاربردها و ویژگی‌ها | سنگ حسن",
    description: "راهنمای کامل سنگ تراورتن؛ بررسی انواع تراورتن، ویژگی‌ها، مزایا، معایب، کاربرد در نما و ساختمان و نکات مهم خرید سنگ تراورتن.",
    h1: "سنگ تراورتن چیست؟ راهنمای کامل انواع، معادن و کاربردها",
    canonical: "/fa/blogs/everything-about-travertine-stone"
  }
};

function getArticleSeoOverride(locale, slug) {
  return locale === "fa" ? faArticleSeoOverrides[slug] || null : null;
}

const freshRequest = () => ({
  cache: "no-store",
  headers: {
    "Cache-Control": "no-cache",
    Pragma: "no-cache"
  }
});

function versionBlogImages(value = "", version = "") {
  if (!version) return value;
  return value.replace(/(<img\b[^>]*\bsrc=)(["'])([^"']+)\2/gi, (match, prefix, quote, url) => {
    if (!url.includes("/images/blogs/")) return match;
    return `${prefix}${quote}${appendImageVersion(url, version)}${quote}`;
  });
}

function articleHTML(value = "", imageVersion = "") {
  const headings = [];
  let index = 0;
  const html = versionBlogImages(value, imageVersion).replace(/<(h[234])([^>]*)>([\s\S]*?)<\/\1>/gi, (match, tag, attrs, inner) => {
    const text = inner.replace(/<[^>]+>/g, "").replace(/&nbsp;/g, " ").trim();
    const id = `section-${++index}`;
    headings.push({ id, level: Number(tag.slice(1)), text });
    const cleanAttrs = attrs.replace(/\sid=(['"]).*?\1/i, "");
    return `<${tag}${cleanAttrs} id="${id}">${inner}</${tag}>`;
  });
  return { html, headings };
}

export default function BlogDetail() {
  const { slug } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const locale = getLanguageFromPath(location.pathname);
  const prerenderedBlog = usePrerenderData("blog");
  const initialBlog = useMemo(() => {
    if (!prerenderedBlog) return null;
    return prerenderedBlog.locale === locale && prerenderedBlog.slug === slug ? prerenderedBlog : null;
  }, [locale, prerenderedBlog, slug]);
  const [blog, setBlog] = useState(initialBlog || null);
  const [loading, setLoading] = useState(!initialBlog);
  const [notFound, setNotFound] = useState(false);
  const meta = localeMeta[locale] || localeMeta.en;
  const isRTL = locale === "fa" || locale === "ar";
  const basePath = `/${locale}/blogs`;

  useEffect(() => {
    let active = true;

    const loadFreshBlog = (showLoading = false) => {
      setNotFound(false);
      if (showLoading) setLoading(true);
      fetchJSON(`/api/blogs/${locale}/${slug}?_=${Date.now()}`, freshRequest())
        .then((response) => {
          if (!active) return;
          const item = response.data;
          if (item.redirected_from && item.slug !== slug) {
            navigate(`/${locale}/blogs/${item.slug}`, { replace: true });
            return;
          }
          setBlog(item);
        })
        .catch((error) => {
          if (!active) return;
          if (error?.status !== 404) {
            setBlog((current) => current || initialBlog || null);
            return;
          }
          setBlog(null);
          setNotFound(true);
        })
        .finally(() => { if (active) setLoading(false); });
    };

    const refreshWhenVisible = () => {
      if (typeof document === "undefined" || document.visibilityState === "visible") {
        loadFreshBlog(false);
      }
    };

    loadFreshBlog(!initialBlog);
    window.addEventListener("focus", refreshWhenVisible);
    document.addEventListener("visibilitychange", refreshWhenVisible);
    return () => {
      active = false;
      window.removeEventListener("focus", refreshWhenVisible);
      document.removeEventListener("visibilitychange", refreshWhenVisible);
    };
  }, [initialBlog, locale, navigate, slug]);

  const imageVersion = blog?.updated_at || blog?.published_at || "";
  const prepared = useMemo(() => articleHTML(blog?.content_html || "", imageVersion), [blog?.content_html, imageVersion]);
  const alternates = useMemo(() => {
    const items = (blog?.translations || []).map((item) => ({ lang: item.locale, path: `/${item.locale}/blogs/${item.slug}` }));
    const english = items.find((item) => item.lang === "en") || items[0];
    return english ? [...items, { lang: "x-default", path: english.path }] : items;
  }, [blog?.translations]);
  const seoOverride = getArticleSeoOverride(locale, blog?.slug || slug);
  const path = seoOverride?.canonical || blog?.canonical_url || `${basePath}/${blog?.slug || slug}`;
  const title = seoOverride?.title || blog?.seo_title || blog?.title || "Article";
  const description = seoOverride?.description || blog?.seo_description || blog?.excerpt || title;
  const headingTitle = seoOverride?.h1 || blog?.title || title;
  const published = blog?.published_at || blog?.created_at;
  const image = blog?.og_image_url || blog?.cover_image_url || "";
  const resolvedImage = image ? resolveVersionedImageUrl(image, imageVersion) : "";
  const robots = notFound ? "noindex,follow" : blog?.robots || "index,follow";
  const canonical = getCanonicalUrl(path);
  const jsonLd = blog ? [
    { "@context": "https://schema.org", "@type": "BlogPosting", headline: headingTitle, description, image: resolvedImage || undefined, author: { "@type": "Organization", name: blog.author_name || "SangeHassan" }, publisher: { "@type": "Organization", name: locale === "fa" ? "سنگ حسن" : "SangeHassan", url: getCanonicalUrl("/") }, datePublished: published, dateModified: blog.updated_at, mainEntityOfPage: canonical, inLanguage: locale === "fa" ? "fa-IR" : locale },
    { "@context": "https://schema.org", "@type": "BreadcrumbList", itemListElement: [{ "@type": "ListItem", position: 1, name: meta.home, item: getCanonicalUrl("/") }, { "@type": "ListItem", position: 2, name: meta.articles, item: getCanonicalUrl(basePath) }, { "@type": "ListItem", position: 3, name: blog.title, item: canonical }] }
  ] : null;

  usePageSeo({ title, description, path, lang: locale, locale: meta.locale, image: resolvedImage, type: "article", robots, alternates, jsonLd, jsonLdId: "blog-article-jsonld" });

  useEffect(() => {
    if (!blog?.translations || typeof window === "undefined") return undefined;
    window.__SH_BLOG_ALTERNATES__ = Object.fromEntries(blog.translations.map((item) => [item.locale, `/${item.locale}/blogs/${item.slug}`]));
    return () => { delete window.__SH_BLOG_ALTERNATES__; };
  }, [blog?.translations]);

  if (loading) return <div className="section-shell min-h-[55vh] py-16 text-sm text-primary/55">...</div>;
  if (notFound || !blog) return <div className="section-shell min-h-[55vh] py-16" dir={isRTL ? "rtl" : "ltr"}><h1 className="font-display text-4xl">404</h1><Link to={basePath} className="mt-6 inline-block text-accent underline">{meta.back}</Link></div>;

  return (
    <article dir={isRTL ? "rtl" : "ltr"} itemScope itemType="https://schema.org/BlogPosting">
      <div className="section-shell pt-5">
        <nav className="flex flex-wrap items-center gap-2 text-xs text-primary/45"><Link to="/">{meta.home}</Link><span>/</span><Link to={basePath}>{meta.articles}</Link><span>/</span><span>{blog.title}</span></nav>
      </div>
      <header className="section-shell pb-9 pt-10">
        {blog.category_slug && <p className="text-xs font-semibold uppercase text-accent">{blog.category_slug}</p>}
        <h1 itemProp="headline" className="mt-3 max-w-4xl font-display text-4xl leading-tight md:text-6xl">{headingTitle}</h1>
        {blog.excerpt && <p itemProp="description" className="mt-6 max-w-3xl text-lg leading-9 text-primary/65">{blog.excerpt}</p>}
        <div className="mt-7 flex flex-wrap items-center gap-x-5 gap-y-2 text-sm text-primary/50">
          <span itemProp="author" itemScope itemType="https://schema.org/Organization"><span itemProp="name">{blog.author_name}</span></span><time itemProp="datePublished" dateTime={published}>{new Date(published).toLocaleDateString(locale)}</time>
          {blog.reading_time_minutes > 0 && <span className="inline-flex items-center gap-1.5"><Clock3 size={15} />{blog.reading_time_minutes} {meta.min}</span>}
          {blog.updated_at && blog.updated_at !== blog.created_at && <span>{meta.updated}: <time itemProp="dateModified" dateTime={blog.updated_at}>{new Date(blog.updated_at).toLocaleDateString(locale)}</time></span>}
        </div>
      </header>
      {blog.cover_image_url && <div className="section-shell"><img itemProp="image" src={resolveVersionedImageUrl(blog.cover_image_url, imageVersion)} alt={blog.featured_image_alt || blog.title} width="1200" height="675" className="aspect-[16/9] w-full object-cover" /></div>}

      <div className="section-shell grid gap-10 py-12 lg:grid-cols-[220px_minmax(0,720px)] lg:justify-center">
        {prepared.headings.length > 2 && <aside className="lg:sticky lg:top-28 lg:self-start"><p className="text-xs font-semibold uppercase text-primary/45">{meta.toc}</p><ol className="mt-4 space-y-3 border-s border-primary/15 ps-4 text-sm text-primary/60">{prepared.headings.map((heading) => <li key={heading.id} className={heading.level > 2 ? "ps-3" : ""}><a href={`#${heading.id}`} className="hover:text-primary">{heading.text}</a></li>)}</ol></aside>}
        <div className={`blog-article-content ${prepared.headings.length <= 2 ? "lg:col-start-2" : ""}`} dangerouslySetInnerHTML={{ __html: prepared.html }} />
      </div>

      <footer className="section-shell border-t border-primary/15 py-9">
        <div className="flex flex-wrap items-center justify-between gap-5">
          <Link to={basePath} className="inline-flex items-center gap-2 text-sm font-semibold text-accent underline decoration-accent/35 underline-offset-4">{isRTL ? <ArrowRight size={17} /> : <ArrowLeft size={17} />}{meta.back}</Link>
          <div className="flex gap-2">{(blog.translations || []).filter((item) => item.locale !== locale).map((item) => <Link key={item.locale} to={`/${item.locale}/blogs/${item.slug}`} className="border border-primary/20 px-3 py-2 text-xs font-semibold uppercase">{item.locale}</Link>)}</div>
        </div>
      </footer>
    </article>
  );
}
