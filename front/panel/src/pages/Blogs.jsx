import { useEffect, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";

const emptyForm = {
  title: "",
  excerpt: "",
  content: "",
  cover_image_url: ""
};

export default function Blogs() {
  const { t } = useTranslation();
  const [blogs, setBlogs] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [formOpen, setFormOpen] = useState(true);

  const loadBlogs = async () => {
    try {
      const response = await fetchJSON("/api/admin/blogs");
      setBlogs(response.data || []);
      setError("");
    } catch (err) {
      setBlogs([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBlogs();
  }, []);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    try {
      if (editingId) {
        await fetchJSON(`/api/admin/blogs/${editingId}`, {
          method: "PUT",
          body: JSON.stringify(form)
        });
      } else {
        await fetchJSON("/api/admin/blogs", {
          method: "POST",
          body: JSON.stringify(form)
        });
      }
      setForm(emptyForm);
      setEditingId(null);
      loadBlogs();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const handleEdit = (blog) => {
    setEditingId(blog.id);
    setForm({
      title: blog.title || "",
      excerpt: blog.excerpt || "",
      content: blog.content || "",
      cover_image_url: blog.cover_image_url || ""
    });
  };

  const handleDelete = async (id) => {
    try {
      await fetchJSON(`/api/admin/blogs/${id}`, { method: "DELETE" });
      loadBlogs();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  return (
    <div className="space-y-6">
      <section className="panel-card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-display text-xl">{t("panelBlogs.title")}</h2>
          <button
            type="button"
            onClick={() => setFormOpen((prev) => !prev)}
            className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
          >
            {formOpen ? t("actions.hideForm") : t("actions.showForm")}
          </button>
        </div>

        {formOpen && (
          <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
            <div className="grid gap-4 md:grid-cols-2">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70 md:col-span-2">
                {t("form.title")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.title}
                  onChange={(event) => setForm({ ...form, title: event.target.value })}
                  required
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70 md:col-span-2">
                {t("form.excerpt")}
                <textarea
                  rows="3"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.excerpt}
                  onChange={(event) => setForm({ ...form, excerpt: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70 md:col-span-2">
                {t("form.content")}
                <textarea
                  rows="5"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.content}
                  onChange={(event) => setForm({ ...form, content: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.coverImageUrl")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.cover_image_url}
                  onChange={(event) => setForm({ ...form, cover_image_url: event.target.value })}
                />
              </label>
            </div>

            {error && <p className="text-sm text-red-500">{error}</p>}

            <div className="flex items-center gap-3">
              <button
                type="submit"
                className="rounded-full bg-primary px-5 py-2 text-xs font-semibold text-sand"
              >
                {editingId ? t("actions.update") : t("actions.create")}
              </button>
              {editingId && (
                <button
                  type="button"
                  className="rounded-full border border-primary/20 px-5 py-2 text-xs font-semibold text-primary/70"
                  onClick={() => {
                    setEditingId(null);
                    setForm(emptyForm);
                  }}
                >
                  {t("actions.cancel")}
                </button>
              )}
            </div>
          </form>
        )}
      </section>

      <section className="panel-card">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="font-display text-xl">{t("panelBlogs.title")}</h3>
        </div>

        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : blogs.length === 0 ? (
          <p className="text-sm text-primary/70">{t("panelBlogs.empty")}</p>
        ) : (
          <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
            {blogs.map((blog) => (
              <div
                key={blog.id}
                className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
              >
                <div>
                  <p className="text-sm font-semibold text-primary">{blog.title}</p>
                  <p className="text-xs text-primary/60">{blog.excerpt}</p>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => handleEdit(blog)}
                    className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold text-primary/70"
                  >
                    {t("actions.edit")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(blog.id)}
                    className="rounded-full border border-red-200 px-3 py-1 text-xs font-semibold text-red-500"
                  >
                    {t("actions.delete")}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
