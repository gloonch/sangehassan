import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";

const taxonomies = [
  { key: "variants", labelKey: "panelProductMeta.variants" },
  { key: "mines", labelKey: "panelProductMeta.mines" },
  { key: "finishes", labelKey: "panelProductMeta.finishes" }
];

const emptyForm = {
  key: "",
  label_en: "",
  label_fa: "",
  label_ar: "",
  link_url: ""
};

const slugify = (value) =>
  value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)+/g, "");

export default function ProductTerms() {
  const { t } = useTranslation();
  const [terms, setTerms] = useState([]);
  const [activeTaxonomy, setActiveTaxonomy] = useState("variants");
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [formOpen, setFormOpen] = useState(true);

  const slugPreview = useMemo(() => form.key || slugify(form.label_en), [form.key, form.label_en]);
  const visibleTerms = useMemo(
    () => terms.filter((term) => term.taxonomy === activeTaxonomy),
    [activeTaxonomy, terms]
  );

  const loadTerms = async () => {
    try {
      const response = await fetchJSON("/api/admin/product-terms");
      setTerms(response.data || []);
      setError("");
    } catch (err) {
      setTerms([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTerms();
  }, []);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    try {
      await fetchJSON("/api/admin/product-terms", {
        method: "POST",
        body: JSON.stringify({
          taxonomy: activeTaxonomy,
          key: slugPreview,
          label_en: form.label_en,
          label_fa: form.label_fa,
          label_ar: form.label_ar,
          link_url: form.link_url
        })
      });
      setForm(emptyForm);
      setEditingId(null);
      loadTerms();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const handleEdit = (term) => {
    setActiveTaxonomy(term.taxonomy);
    setEditingId(term.id);
    setForm({
      key: term.key || "",
      label_en: term.label_en || "",
      label_fa: term.label_fa || "",
      label_ar: term.label_ar || "",
      link_url: term.link_url || ""
    });
    setFormOpen(true);
  };

  const handleDelete = async (id) => {
    try {
      await fetchJSON(`/api/admin/product-terms/${id}`, { method: "DELETE" });
      loadTerms();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const resetForm = () => {
    setEditingId(null);
    setForm(emptyForm);
  };

  const selectTaxonomy = (key) => {
    setActiveTaxonomy(key);
    resetForm();
  };

  return (
    <div className="space-y-6">
      <section className="panel-card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-display text-xl">{t("panelProductTerms.title")}</h2>
          <button
            type="button"
            onClick={() => setFormOpen((prev) => !prev)}
            className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
          >
            {formOpen ? t("actions.hideForm") : t("actions.showForm")}
          </button>
        </div>

        <div className="mt-5 flex flex-wrap gap-2">
          {taxonomies.map((item) => (
            <button
              key={item.key}
              type="button"
              onClick={() => selectTaxonomy(item.key)}
              className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${
                activeTaxonomy === item.key
                  ? "border-primary bg-primary text-sand"
                  : "border-primary/20 bg-white text-primary/70 hover:border-primary/45"
              }`}
            >
              {t(item.labelKey)}
            </button>
          ))}
        </div>

        {formOpen && (
          <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
            <div className="grid gap-4 md:grid-cols-2">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.titleEn")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.label_en}
                  onChange={(event) => setForm({ ...form, label_en: event.target.value })}
                  required
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.titleFa")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.label_fa}
                  onChange={(event) => setForm({ ...form, label_fa: event.target.value })}
                  required
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.titleAr")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.label_ar}
                  onChange={(event) => setForm({ ...form, label_ar: event.target.value })}
                  required
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.slug")}
                <input
                  type="text"
                  className={`mt-2 w-full rounded-xl border border-primary/20 px-4 py-3 text-sm ${
                    editingId ? "bg-primary/5" : "bg-white"
                  }`}
                  value={slugPreview}
                  onChange={(event) => setForm({ ...form, key: slugify(event.target.value) })}
                  readOnly={Boolean(editingId)}
                  required
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70 md:col-span-2">
                {t("form.linkUrl")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.link_url}
                  onChange={(event) => setForm({ ...form, link_url: event.target.value })}
                  placeholder="/products?category=travertine"
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
                  onClick={resetForm}
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
          <h3 className="font-display text-xl">
            {t(taxonomies.find((item) => item.key === activeTaxonomy)?.labelKey || "panelProductTerms.title")}
          </h3>
        </div>

        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : visibleTerms.length === 0 ? (
          <p className="text-sm text-primary/70">{t("panelProductTerms.empty")}</p>
        ) : (
          <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
            {visibleTerms.map((term) => (
              <div
                key={term.id}
                className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
              >
                <div>
                  <p className="text-sm font-semibold text-primary">{term.label_en}</p>
                  <p className="text-xs text-primary/60">{term.label_fa} • {term.label_ar}</p>
                  <p className="text-xs text-primary/40">{term.key}</p>
                  {term.link_url ? (
                    <p className="mt-1 max-w-xl truncate text-xs text-primary/50">{term.link_url}</p>
                  ) : null}
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => handleEdit(term)}
                    className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold text-primary/70"
                  >
                    {t("actions.edit")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(term.id)}
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
