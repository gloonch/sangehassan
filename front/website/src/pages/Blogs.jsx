import { useEffect, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";

export default function Blogs() {
  const { t } = useTranslation();
  const [blogs, setBlogs] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      try {
        const response = await fetchJSON("/api/blogs");
        if (!mounted) return;
        setBlogs(response.data || []);
      } catch (error) {
        if (!mounted) return;
        setBlogs([]);
      } finally {
        if (mounted) setLoading(false);
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, []);

  return (
    <section className="section-shell py-16">
      <div className="mb-8">
        <p className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("blogs.title")}</p>
        <h1 className="mt-2 font-display text-3xl md:text-4xl">{t("blogs.subtitle")}</h1>
      </div>

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : blogs.length === 0 ? (
        <p className="text-sm text-primary/70">{t("blogs.empty")}</p>
      ) : (
        <div className="grid gap-6 md:grid-cols-2">
          {blogs.map((blog) => (
            <article key={blog.id} className="glass-panel flex h-full flex-col rounded-2xl p-6">
              <div
                className="mb-4 h-40 rounded-xl bg-primary/10 bg-cover bg-center"
                style={blog.cover_image_url ? { backgroundImage: `url(${blog.cover_image_url})` } : undefined}
              />
              <h2 className="font-display text-2xl">{blog.title}</h2>
              <p className="mt-3 text-sm text-primary/70">{blog.excerpt}</p>
              <button type="button" className="mt-auto pt-6 text-left text-sm font-semibold text-accent">
                {t("blogs.readMore")}
              </button>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}
