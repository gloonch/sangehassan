import { lazy, Suspense, useEffect, useMemo, useState } from "react";
import { ExternalLink, Eye, FilePlus2, Pencil, Search, Trash2, X } from "lucide-react";
import { API_BASE, fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const RichTextEditor = lazy(() => import("../components/RichTextEditor"));
const locales = ["fa", "en", "ar"];
const maxBlogUploadBytes = 25 * 1024 * 1024;
const blogImageTargets = {
  inline: { maxWidth: 1600, maxHeight: 1600, quality: 0.82 },
  cover: { maxWidth: 1600, maxHeight: 900, quality: 0.85 },
  og: { maxWidth: 1200, maxHeight: 630, quality: 0.85 }
};

const emptyTranslation = (locale) => ({
  locale,
  title: "",
  slug: "",
  excerpt: "",
  content_json: { type: "doc", content: [] },
  content_html: "",
  seo_title: "",
  seo_description: "",
  canonical_url: "",
  robots: "index,follow",
  translation_status: "draft",
  featured_image_alt: "",
  og_image_alt: ""
});

const createEmptyForm = () => ({
  status: "draft",
  author_name: "SangeHassan",
  cover_image_url: "",
  og_image_url: "",
  category_slug: "",
  tags: [],
  is_featured: false,
  scheduled_at: "",
  translations: locales.map(emptyTranslation)
});

function toLocalDateTime(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const offset = date.getTimezoneOffset() * 60000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

function normalizeBlog(blog) {
  const byLocale = new Map((blog.translations || []).map((item) => [item.locale, item]));
  return {
    ...createEmptyForm(),
    ...blog,
    scheduled_at: toLocalDateTime(blog.scheduled_at),
    translations: locales.map((locale) => ({ ...emptyTranslation(locale), ...(byLocale.get(locale) || {}) }))
  };
}

function contentStats(html = "") {
  if (typeof document === "undefined") return { words: 0, minutes: 0, links: 0, headings: [] };
  const doc = new DOMParser().parseFromString(html, "text/html");
  const words = (doc.body.textContent || "").trim().split(/\s+/).filter(Boolean).length;
  return {
    words,
    minutes: words ? Math.ceil(words / 200) : 0,
    links: doc.querySelectorAll("a[href]").length,
    headings: [...doc.querySelectorAll("h2,h3,h4")].map((node) => node.textContent?.trim()).filter(Boolean)
  };
}

function formatBytes(bytes = 0) {
  if (!bytes) return "0 KB";
  const units = ["B", "KB", "MB", "GB"];
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / (1024 ** index);
  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`;
}

function uploadPhaseLabel(uploadState) {
  if (!uploadState?.phase) return "";
  if (uploadState.phase === "optimizing") return "Optimizing image";
  if (uploadState.phase === "uploading") return `Uploading ${uploadState.percent || 0}%`;
  return "Preparing image";
}

function readImageSize(file) {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file);
    const image = new Image();
    image.onload = () => {
      URL.revokeObjectURL(url);
      resolve({ image, width: image.naturalWidth || image.width, height: image.naturalHeight || image.height });
    };
    image.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error("Could not read image"));
    };
    image.src = url;
  });
}

async function optimizeBlogImage(file, target) {
  if (!file || !file.type.startsWith("image/")) return file;
  const options = blogImageTargets[target] || blogImageTargets.inline;
  const { image, width, height } = await readImageSize(file);
  const scale = Math.min(1, options.maxWidth / width, options.maxHeight / height);
  const outputWidth = Math.max(1, Math.round(width * scale));
  const outputHeight = Math.max(1, Math.round(height * scale));

  if (scale === 1 && file.type === "image/webp" && file.size < 900 * 1024) {
    return file;
  }

  const canvas = document.createElement("canvas");
  canvas.width = outputWidth;
  canvas.height = outputHeight;
  const context = canvas.getContext("2d", { alpha: true });
  if (!context) return file;
  context.drawImage(image, 0, 0, outputWidth, outputHeight);

  const blob = await new Promise((resolve) => canvas.toBlob(resolve, "image/webp", options.quality));
  if (!blob) return file;

  const shouldKeepOriginal = scale === 1 && blob.size >= file.size;
  if (shouldKeepOriginal) return file;

  const basename = file.name.replace(/\.[^.]+$/, "") || "blog-image";
  return new File([blob], `${basename}.webp`, { type: "image/webp", lastModified: Date.now() });
}

function uploadBlogFile(file, onProgress) {
  return new Promise((resolve, reject) => {
    const data = new FormData();
    data.append("file", file);
    const request = new XMLHttpRequest();
    request.open("POST", `${API_BASE}/api/admin/upload/blog`);
    request.withCredentials = true;
    request.upload.onprogress = (event) => {
      if (!event.lengthComputable) return;
      onProgress(Math.round((event.loaded / event.total) * 100));
    };
    request.onload = () => {
      if (request.status < 200 || request.status >= 300) {
        reject(new Error("Upload failed"));
        return;
      }
      try {
        const payload = JSON.parse(request.responseText || "{}");
        resolve(payload?.data?.image_url || "");
      } catch (_) {
        reject(new Error("Upload failed"));
      }
    };
    request.onerror = () => reject(new Error("Upload failed"));
    request.send(data);
  });
}

export default function Blogs() {
  const [blogs, setBlogs] = useState([]);
  const [form, setForm] = useState(createEmptyForm);
  const [editingId, setEditingId] = useState(null);
  const [activeLocale, setActiveLocale] = useState("fa");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [previewOpen, setPreviewOpen] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [uploadState, setUploadState] = useState(null);

  const translation = form.translations.find((item) => item.locale === activeLocale) || emptyTranslation(activeLocale);
  const stats = useMemo(() => contentStats(translation.content_html), [translation.content_html]);

  const loadBlogs = async () => {
    try {
      const response = await fetchJSON("/api/admin/blogs");
      setBlogs(response.data || []);
      setError("");
    } catch (err) {
      setBlogs([]);
      setError(err?.message || "Could not load articles.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBlogs();
  }, []);

  useEffect(() => {
    if (!dirty) return undefined;
    const warn = (event) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  useEffect(() => {
    if (!dirty) return;
    const timer = window.setTimeout(() => {
      localStorage.setItem("blog-editor-draft", JSON.stringify({ editingId, form, saved_at: new Date().toISOString() }));
    }, 800);
    return () => window.clearTimeout(timer);
  }, [dirty, editingId, form]);

  const updateForm = (patch) => {
    setForm((current) => ({ ...current, ...patch }));
    setDirty(true);
  };

  const updateTranslation = (patch) => {
    setForm((current) => ({
      ...current,
      translations: current.translations.map((item) => item.locale === activeLocale ? { ...item, ...patch } : item)
    }));
    setDirty(true);
  };

  const uploadImage = async (file, target = "inline") => {
    if (!file) return "";
    if (file.size > maxBlogUploadBytes) {
      setError(`Image is too large. Please choose a file under ${formatBytes(maxBlogUploadBytes)}.`);
      return "";
    }
    setUploadState({ target, phase: "optimizing", percent: 0, originalSize: file.size, optimizedSize: 0 });
    setError("");
    try {
      const optimized = await optimizeBlogImage(file, target);
      setUploadState((current) => ({ ...current, phase: "uploading", percent: 0, optimizedSize: optimized.size }));
      const url = await uploadBlogFile(optimized, (percent) => {
        setUploadState((current) => ({ ...current, phase: "uploading", percent }));
      });
      if (target === "cover") updateForm({ cover_image_url: url });
      if (target === "og") updateForm({ og_image_url: url });
      return url;
    } catch (err) {
      setError(err?.message || "Upload failed.");
      return "";
    } finally {
      setUploadState(null);
    }
  };

  const resetEditor = () => {
    setEditingId(null);
    setForm(createEmptyForm());
    setActiveLocale("fa");
    setDirty(false);
    localStorage.removeItem("blog-editor-draft");
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      const payload = {
        ...form,
        scheduled_at: form.scheduled_at ? new Date(form.scheduled_at).toISOString() : null,
        tags: Array.isArray(form.tags) ? form.tags : [],
        translations: form.translations.filter((item) => item.title.trim())
      };
      const response = await fetchJSON(editingId ? `/api/admin/blogs/${editingId}` : "/api/admin/blogs", {
        method: editingId ? "PUT" : "POST",
        body: JSON.stringify(payload)
      });
      const saved = response.data;
      setEditingId(saved.id);
      setForm(normalizeBlog(saved));
      setDirty(false);
      localStorage.removeItem("blog-editor-draft");
      await loadBlogs();
    } catch (err) {
      setError(err?.message || "Could not save the article.");
    } finally {
      setSaving(false);
    }
  };

  const editBlog = async (id) => {
    setError("");
    try {
      const response = await fetchJSON(`/api/admin/blogs/${id}`);
      setEditingId(id);
      setForm(normalizeBlog(response.data));
      setActiveLocale(response.data.translations?.[0]?.locale || "fa");
      setDirty(false);
      window.scrollTo({ top: 0, behavior: "smooth" });
    } catch (err) {
      setError(err?.message || "Could not load the article.");
    }
  };

  const duplicateBlog = (blog) => {
    const copy = normalizeBlog(blog);
    copy.status = "draft";
    copy.published_at = null;
    copy.scheduled_at = "";
    copy.translations = copy.translations.map((item) => ({
      ...item,
      id: undefined,
      slug: item.title ? `${item.slug || "article"}-copy` : "",
      translation_status: "draft"
    }));
    setEditingId(null);
    setForm(copy);
    setActiveLocale(copy.translations.find((item) => item.title)?.locale || "fa");
    setDirty(true);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const deleteBlog = async (id) => {
    if (!window.confirm("Delete this article?")) return;
    try {
      await fetchJSON(`/api/admin/blogs/${id}`, { method: "DELETE" });
      if (editingId === id) resetEditor();
      await loadBlogs();
    } catch (err) {
      setError(err?.message || "Could not delete the article.");
    }
  };

  const filteredBlogs = useMemo(() => {
    const query = search.trim().toLowerCase();
    return blogs.filter((blog) => {
      if (statusFilter !== "all" && blog.status !== statusFilter) return false;
      if (!query) return true;
      return (blog.translations || []).some((item) => `${item.title} ${item.slug}`.toLowerCase().includes(query));
    });
  }, [blogs, search, statusFilter]);

  const publicPath = translation.slug ? `/${activeLocale}/blogs/${translation.slug}` : "";
  const seoTitle = translation.seo_title || translation.title || "Article title";
  const seoDescription = translation.seo_description || translation.excerpt || "Article description";

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit} className="space-y-6">
        <section className="panel-card">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-primary/10 pb-5">
            <div>
              <p className="text-xs font-semibold uppercase text-primary/50">Blog publishing</p>
              <h2 className="mt-1 font-display text-2xl">{editingId ? "Edit article" : "New article"}</h2>
            </div>
            <div className="flex items-center gap-2">
              <button type="button" title="Preview" onClick={() => setPreviewOpen(true)} className="inline-flex h-10 w-10 items-center justify-center border border-primary/20 bg-white"><Eye size={18} /></button>
              <button type="button" onClick={resetEditor} className="inline-flex items-center gap-2 border border-primary/20 px-4 py-2 text-sm font-semibold"><FilePlus2 size={17} /> New</button>
              <button type="submit" disabled={saving} className="bg-primary px-5 py-2.5 text-sm font-semibold text-white disabled:opacity-50">{saving ? "Saving..." : "Save article"}</button>
            </div>
          </div>

          <div className="mt-5 grid gap-4 lg:grid-cols-4">
            <label className="text-xs font-semibold text-primary/65">Status
              <select value={form.status} onChange={(event) => updateForm({ status: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm">
                <option value="draft">Draft</option><option value="scheduled">Scheduled</option><option value="published">Published</option><option value="archived">Archived</option>
              </select>
            </label>
            <label className="text-xs font-semibold text-primary/65">Author
              <input value={form.author_name} onChange={(event) => updateForm({ author_name: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm" />
            </label>
            <label className="text-xs font-semibold text-primary/65">Category slug
              <input value={form.category_slug} onChange={(event) => updateForm({ category_slug: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm" />
            </label>
            <label className="text-xs font-semibold text-primary/65">Tags
              <input value={(form.tags || []).join(", ")} onChange={(event) => updateForm({ tags: event.target.value.split(",").map((item) => item.trim()).filter(Boolean) })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm" />
            </label>
            {form.status === "scheduled" && <label className="text-xs font-semibold text-primary/65">Publish at
              <input required type="datetime-local" value={form.scheduled_at} onChange={(event) => updateForm({ scheduled_at: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm" />
            </label>}
            <label className="flex items-center gap-2 pt-7 text-sm font-semibold"><input type="checkbox" checked={form.is_featured} onChange={(event) => updateForm({ is_featured: event.target.checked })} /> Featured article</label>
          </div>

          <div className="mt-5 grid gap-4 md:grid-cols-2">
            {[{ key: "cover", label: "Featured image", value: form.cover_image_url }, { key: "og", label: "Open Graph image", value: form.og_image_url }].map((image) => (
              <label key={image.key} className="border border-primary/10 bg-white p-4 text-xs font-semibold text-primary/65">{image.label}
                <div className="mt-3 flex items-center gap-4">
                  <input type="file" accept="image/png,image/jpeg,image/webp" disabled={Boolean(uploadState)} onChange={(event) => { uploadImage(event.target.files?.[0], image.key); event.target.value = ""; }} className="min-w-0 flex-1 text-sm" />
                  {image.value && <img src={resolveImageUrl(image.value)} alt="" className="h-16 w-24 object-cover" />}
                </div>
                {uploadState?.target === image.key && <p className="mt-3 text-[11px] font-medium text-primary/55">
                  {uploadPhaseLabel(uploadState)}
                  {uploadState.optimizedSize ? ` · ${formatBytes(uploadState.originalSize)} → ${formatBytes(uploadState.optimizedSize)}` : ""}
                </p>}
              </label>
            ))}
          </div>
        </section>

        <section className="panel-card">
          <div className="flex border-b border-primary/10">
            {locales.map((locale) => {
              const item = form.translations.find((entry) => entry.locale === locale);
              return <button key={locale} type="button" onClick={() => setActiveLocale(locale)} className={`relative min-w-24 px-5 py-3 text-sm font-semibold uppercase ${activeLocale === locale ? "text-primary" : "text-primary/45"}`}>
                {locale}<span className={`absolute bottom-0 left-0 h-0.5 w-full ${activeLocale === locale ? "bg-accent" : "bg-transparent"}`} />{item?.title && <span className="ml-2 inline-block h-1.5 w-1.5 rounded-full bg-emerald-500" />}
              </button>;
            })}
          </div>

          <div className="mt-6 grid gap-5 lg:grid-cols-[minmax(0,1fr)_280px]">
            <div className="space-y-4" dir={activeLocale === "fa" || activeLocale === "ar" ? "rtl" : "ltr"}>
              <label className="block text-xs font-semibold text-primary/65">Title
                <input value={translation.title} onChange={(event) => updateTranslation({ title: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-4 py-3 text-base" />
              </label>
              <label className="block text-xs font-semibold text-primary/65">Excerpt
                <textarea rows="3" value={translation.excerpt} onChange={(event) => updateTranslation({ excerpt: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-4 py-3 text-sm" />
              </label>
              <div>
                <div className="mb-2 flex items-center justify-between text-xs font-semibold text-primary/65"><span>Content</span><span className="font-normal text-primary/45">{stats.words} words · {stats.minutes} min · {stats.links} links</span></div>
                {uploadState?.target === "inline" && <p className="mb-2 border border-primary/10 bg-white px-3 py-2 text-xs font-medium text-primary/60">
                  {uploadPhaseLabel(uploadState)}
                  {uploadState.optimizedSize ? ` · ${formatBytes(uploadState.originalSize)} → ${formatBytes(uploadState.optimizedSize)}` : ""}
                </p>}
                <Suspense fallback={<div className="h-72 animate-pulse bg-primary/5" />}>
                  <RichTextEditor value={{ html: translation.content_html, json: translation.content_json }} locale={activeLocale} onUploadImage={(file) => uploadImage(file, "inline")} onChange={({ html, json }) => updateTranslation({ content_html: html, content_json: json })} />
                </Suspense>
              </div>
            </div>

            <aside className="space-y-4 border-l border-primary/10 pl-5">
              <label className="block text-xs font-semibold text-primary/65">Translation
                <select value={translation.translation_status} onChange={(event) => updateTranslation({ translation_status: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm"><option value="draft">Draft</option><option value="published">Published</option></select>
              </label>
              <label className="block text-xs font-semibold text-primary/65">Slug
                <input value={translation.slug} placeholder="Generated from title" onChange={(event) => updateTranslation({ slug: event.target.value })} dir="ltr" className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-left text-sm" />
              </label>
              <label className="block text-xs font-semibold text-primary/65">Featured image alt
                <input value={translation.featured_image_alt} onChange={(event) => updateTranslation({ featured_image_alt: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm" />
              </label>
              <label className="block text-xs font-semibold text-primary/65">SEO title
                <input value={translation.seo_title} onChange={(event) => updateTranslation({ seo_title: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm" />
                <span className="mt-1 block text-right text-[11px] font-normal text-primary/45">{translation.seo_title.length}/70</span>
              </label>
              <label className="block text-xs font-semibold text-primary/65">Meta description
                <textarea rows="4" value={translation.seo_description} onChange={(event) => updateTranslation({ seo_description: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm" />
                <span className="mt-1 block text-right text-[11px] font-normal text-primary/45">{translation.seo_description.length}/180</span>
              </label>
              <label className="block text-xs font-semibold text-primary/65">Canonical override
                <input value={translation.canonical_url} onChange={(event) => updateTranslation({ canonical_url: event.target.value })} dir="ltr" className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-left text-sm" />
              </label>
              <label className="block text-xs font-semibold text-primary/65">Robots
                <select value={translation.robots} onChange={(event) => updateTranslation({ robots: event.target.value })} className="mt-2 w-full border border-primary/20 bg-white px-3 py-2.5 text-sm"><option value="index,follow">index,follow</option><option value="noindex,follow">noindex,follow</option><option value="noindex,nofollow">noindex,nofollow</option></select>
              </label>
            </aside>
          </div>

          <div className="mt-6 border-t border-primary/10 pt-5">
            <p className="text-xs font-semibold uppercase text-primary/45">Search preview</p>
            <div className="mt-3 max-w-2xl bg-white p-4 shadow-sm" dir={activeLocale === "fa" || activeLocale === "ar" ? "rtl" : "ltr"}>
              <p className="text-xs text-emerald-700" dir="ltr">sangehassan.com{publicPath || `/${activeLocale}/blogs/...`}</p>
              <p className="mt-1 text-lg text-blue-800">{seoTitle}</p>
              <p className="mt-1 text-sm text-primary/65">{seoDescription}</p>
            </div>
          </div>
          {error && <p className="mt-5 border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</p>}
        </section>
      </form>

      <section className="panel-card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h3 className="font-display text-xl">Articles</h3>
          <div className="flex gap-2">
            <label className="relative"><Search size={16} className="absolute left-3 top-2.5 text-primary/40" /><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search" className="border border-primary/20 bg-white py-2 pl-9 pr-3 text-sm" /></label>
            <select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)} className="border border-primary/20 bg-white px-3 py-2 text-sm"><option value="all">All statuses</option><option value="draft">Draft</option><option value="scheduled">Scheduled</option><option value="published">Published</option><option value="archived">Archived</option></select>
          </div>
        </div>
        {loading ? <p className="mt-5 text-sm text-primary/60">Loading...</p> : <div className="mt-5 overflow-x-auto">
          <table className="w-full min-w-[720px] text-left text-sm"><thead className="border-b border-primary/10 text-xs uppercase text-primary/45"><tr><th className="py-3">Title</th><th>Status</th><th>Languages</th><th>Updated</th><th className="text-right">Actions</th></tr></thead><tbody>
            {filteredBlogs.map((blog) => {
              const primary = blog.translations?.find((item) => item.locale === "fa") || blog.translations?.[0];
              return <tr key={blog.id} className="border-b border-primary/10"><td className="py-4"><p className="font-semibold">{primary?.title || "Untitled"}</p><p className="text-xs text-primary/45">{primary?.slug}</p></td><td><span className="border border-primary/15 px-2 py-1 text-xs capitalize">{blog.status}</span></td><td className="uppercase text-primary/60">{(blog.translations || []).map((item) => item.locale).join(" · ")}</td><td className="text-xs text-primary/55">{new Date(blog.updated_at || blog.created_at).toLocaleDateString()}</td><td><div className="flex justify-end gap-1"><button title="Edit" onClick={() => editBlog(blog.id)} className="inline-flex h-9 w-9 items-center justify-center border border-primary/15"><Pencil size={16} /></button><button title="Duplicate" onClick={() => duplicateBlog(blog)} className="inline-flex h-9 w-9 items-center justify-center border border-primary/15"><FilePlus2 size={16} /></button>{primary?.slug && primary.translation_status === "published" && <a title="Open" target="_blank" rel="noreferrer" href={`/${primary.locale}/blogs/${primary.slug}`} className="inline-flex h-9 w-9 items-center justify-center border border-primary/15"><ExternalLink size={16} /></a>}<button title="Delete" onClick={() => deleteBlog(blog.id)} className="inline-flex h-9 w-9 items-center justify-center border border-red-200 text-red-600"><Trash2 size={16} /></button></div></td></tr>;
            })}
          </tbody></table>
        </div>}
      </section>

      {previewOpen && <div className="fixed inset-0 z-[100] overflow-y-auto bg-black/55 p-4 md:p-10"><div className="mx-auto max-w-4xl bg-white shadow-2xl"><div className="sticky top-0 flex items-center justify-between border-b border-primary/10 bg-white px-5 py-3"><strong>{translation.title || "Preview"}</strong><button title="Close" onClick={() => setPreviewOpen(false)} className="inline-flex h-9 w-9 items-center justify-center"><X size={20} /></button></div>{form.cover_image_url && <img src={resolveImageUrl(form.cover_image_url)} alt={translation.featured_image_alt} className="aspect-[16/7] w-full object-cover" />}<article className="blog-preview-content mx-auto max-w-3xl px-6 py-10" dir={activeLocale === "fa" || activeLocale === "ar" ? "rtl" : "ltr"}><h1 className="font-display text-4xl">{translation.title}</h1><p className="mt-4 text-lg text-primary/65">{translation.excerpt}</p><div className="mt-8" dangerouslySetInnerHTML={{ __html: translation.content_html }} /></article></div></div>}
    </div>
  );
}
