import { useEffect, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { API_BASE, fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const emptyForm = {
  title_en: "",
  title_fa: "",
  title_ar: "",
  description: "",
  price: "",
  image_url: "",
  category_id: "",
  is_popular: false
};

export default function Products() {
  const { t } = useTranslation();
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadData = async () => {
    try {
      const [productRes, categoryRes] = await Promise.all([
        fetchJSON("/api/admin/products"),
        fetchJSON("/api/admin/categories")
      ]);
      setProducts(productRes.data || []);
      setCategories(categoryRes.data || []);
      setError("");
    } catch (err) {
      setProducts([]);
      setCategories([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    const payload = {
      ...form,
      price: form.price ? Number(form.price) : 0,
      category_id: Number(form.category_id),
      is_popular: Boolean(form.is_popular)
    };

    try {
      if (editingId) {
        await fetchJSON(`/api/admin/products/${editingId}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
      } else {
        await fetchJSON("/api/admin/products", {
          method: "POST",
          body: JSON.stringify(payload)
        });
      }
      setForm(emptyForm);
      setEditingId(null);
      loadData();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const handleEdit = (product) => {
    setEditingId(product.id);
    setForm({
      title_en: product.title_en || "",
      title_fa: product.title_fa || "",
      title_ar: product.title_ar || "",
      description: product.description || "",
      price: product.price || "",
      image_url: product.image_url || "",
      category_id: product.category_id ? String(product.category_id) : "",
      is_popular: Boolean(product.is_popular)
    });
  };

  const handleDelete = async (id) => {
    try {
      await fetchJSON(`/api/admin/products/${id}`, { method: "DELETE" });
      loadData();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  return (
    <div className="grid gap-6 lg:grid-cols-[1.1fr_1.9fr]">
      <form className="panel-card space-y-4" onSubmit={handleSubmit}>
        <h2 className="font-display text-xl">{t("panelProducts.title")}</h2>

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
          {t("form.description")}
          <textarea
            rows="3"
            className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
            value={form.description}
            onChange={(event) => setForm({ ...form, description: event.target.value })}
          />
        </label>
        <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
          {t("form.price")}
          <input
            type="number"
            className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
            value={form.price}
            onChange={(event) => setForm({ ...form, price: event.target.value })}
          />
        </label>
        <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
          {t("form.imageUrl")}
          <div className="mt-2 flex items-center gap-2">
            <input
              type="file"
              accept="image/*"
              className="w-full text-sm text-primary/70 file:mr-4 file:rounded-full file:border-0 file:bg-primary/10 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20"
              onChange={async (e) => {
                const file = e.target.files?.[0];
                if (!file) return;

                const formData = new FormData();
                formData.append("file", file);

                try {
                  setLoading(true);
                  const res = await fetch(`${API_BASE}/api/admin/upload/product`, {
                    method: "POST",
                    body: formData,
                    credentials: "include",
                  });
                  if (!res.ok) throw new Error("Upload failed");
                  const data = await res.json();
                  setForm({ ...form, image_url: data?.data?.image_url || "" });
                } catch (err) {
                  setError(t("messages.error"));
                } finally {
                  setLoading(false);
                }
              }}
            />
            {form.image_url && (
              <div className="h-10 w-10 overflow-hidden rounded-lg border border-primary/20 bg-primary/5">
                <img src={resolveImageUrl(form.image_url)} alt="Preview" className="h-full w-full object-cover" />
              </div>
            )}
          </div>
        </label>
        <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
          {t("form.category")}
          <select
            className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
            value={form.category_id}
            onChange={(event) => setForm({ ...form, category_id: event.target.value })}
            required
          >
            <option value="">{t("form.category")}</option>
            {categories.map((category) => (
              <option key={category.id} value={category.id}>
                {category.title_en}
              </option>
            ))}
          </select>
        </label>
        <label className="flex items-center gap-3 text-xs font-semibold uppercase tracking-wide text-primary/70">
          <input
            type="checkbox"
            checked={form.is_popular}
            onChange={(event) => setForm({ ...form, is_popular: event.target.checked })}
            className="h-4 w-4 rounded border-primary/20"
          />
          {t("form.popular")}
        </label>

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

      <div className="panel-card">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="font-display text-xl">{t("panelProducts.title")}</h3>
        </div>

        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : products.length === 0 ? (
          <p className="text-sm text-primary/70">{t("panelProducts.empty")}</p>
        ) : (
          <div className="space-y-3">
            {products.map((product) => (
              <div
                key={product.id}
                className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
              >
                <div>
                  <p className="text-sm font-semibold text-primary">{product.title_en}</p>
                  <p className="text-xs text-primary/60">{product.title_fa} • {product.title_ar}</p>
                  <p className="text-xs text-primary/40">{product.category?.title_en}</p>
                </div>
                {product.is_popular && (
                  <span className="rounded-full bg-accent/20 px-3 py-1 text-xs font-semibold text-accent">
                    {t("form.popular")}
                  </span>
                )}
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => handleEdit(product)}
                    className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold text-primary/70"
                  >
                    {t("actions.edit")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(product.id)}
                    className="rounded-full border border-red-200 px-3 py-1 text-xs font-semibold text-red-500"
                  >
                    {t("actions.delete")}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
