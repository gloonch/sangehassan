import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";

const slugify = (value) =>
  value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)+/g, "");

const emptyForm = {
  title_en: "",
  title_fa: "",
  title_ar: ""
};

export default function Categories() {
  const { t } = useTranslation();
  const [categories, setCategories] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [formOpen, setFormOpen] = useState(true);

  const slugPreview = useMemo(() => slugify(form.title_en), [form.title_en]);

  const loadCategories = async () => {
    try {
      const response = await fetchJSON("/api/admin/categories");
      setCategories(response.data || []);
      setError("");
    } catch (err) {
      setCategories([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCategories();
  }, []);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    try {
      if (editingId) {
        await fetchJSON(`/api/admin/categories/${editingId}`, {
          method: "PUT",
          body: JSON.stringify(form)
        });
      } else {
        await fetchJSON("/api/admin/categories", {
          method: "POST",
          body: JSON.stringify(form)
        });
      }
      setForm(emptyForm);
      setEditingId(null);
      loadCategories();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const handleEdit = (category) => {
    setEditingId(category.id);
    setForm({
      title_en: category.title_en,
      title_fa: category.title_fa,
      title_ar: category.title_ar
    });
  };

  const handleDelete = async (id) => {
    try {
      await fetchJSON(`/api/admin/categories/${id}`, { method: "DELETE" });
      loadCategories();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  return (
    <div className="space-y-6">
      <section className="panel-card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-display text-xl">{t("categories.title")}</h2>
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
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.titleEn")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.title_en}
                  onChange={(event) => setForm({ ...form, title_en: event.target.value })}
                  required
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.titleFa")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.title_fa}
                  onChange={(event) => setForm({ ...form, title_fa: event.target.value })}
                  required
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.titleAr")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.title_ar}
                  onChange={(event) => setForm({ ...form, title_ar: event.target.value })}
                  required
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.slug")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/10 bg-primary/5 px-4 py-3 text-sm"
                  value={slugPreview}
                  readOnly
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
          <h3 className="font-display text-xl">{t("categories.title")}</h3>
        </div>

        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : categories.length === 0 ? (
          <p className="text-sm text-primary/70">{t("categories.empty")}</p>
        ) : (
          <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
            {categories.map((category) => (
              <div
                key={category.id}
                className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
              >
                <div>
                  <p className="text-sm font-semibold text-primary">{category.title_en}</p>
                  <p className="text-xs text-primary/60">{category.title_fa} • {category.title_ar}</p>
                  <p className="text-xs text-primary/40">{category.slug}</p>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => handleEdit(category)}
                    className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold text-primary/70"
                  >
                    {t("actions.edit")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(category.id)}
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
