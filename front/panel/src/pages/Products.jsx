import { useEffect, useMemo, useState } from "react";
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
  image_urls: [],
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
  const [filterCategory, setFilterCategory] = useState("all");
  const [selectedImageIndex, setSelectedImageIndex] = useState(0);
  const [formOpen, setFormOpen] = useState(true);

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

  const filteredProducts = useMemo(() => {
    if (filterCategory === "all") return products;
    return products.filter((product) => String(product.category_id || "") === filterCategory);
  }, [filterCategory, products]);

  const getImageCount = (product) => {
    if (typeof product.image_count === "number") return product.image_count;
    return product.image_url ? 1 : 0;
  };

  const handleImageUpload = async (files) => {
    if (!files || files.length === 0) return;
    setError("");
    setLoading(true);
    try {
      const uploads = await Promise.all(
        Array.from(files).map(async (file) => {
          const formData = new FormData();
          formData.append("file", file);
          const res = await fetch(`${API_BASE}/api/admin/upload/product`, {
            method: "POST",
            body: formData,
            credentials: "include",
          });
          if (!res.ok) throw new Error("Upload failed");
          const data = await res.json();
          return data?.data?.image_url || "";
        })
      );
      setForm((prev) => {
        const nextImages = [...(prev.image_urls || []), ...uploads].filter(Boolean);
        return {
          ...prev,
          image_urls: nextImages,
          image_url: nextImages[0] || prev.image_url
        };
      });
    } catch (err) {
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveImage = (index) => {
    setForm((prev) => {
      const nextImages = (prev.image_urls || []).filter((_, idx) => idx !== index);
      return {
        ...prev,
        image_urls: nextImages,
        image_url: nextImages[0] || ""
      };
    });
    if (selectedImageIndex === index) {
      setSelectedImageIndex(0);
    }
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    const images = form.image_urls?.length ? form.image_urls : form.image_url ? [form.image_url] : [];
    const payload = {
      ...form,
      price: form.price ? Number(form.price) : 0,
      category_id: form.category_id ? Number(form.category_id) : 0,
      is_popular: Boolean(form.is_popular),
      image_url: images[0] || form.image_url || "",
      image_urls: images
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
      setSelectedImageIndex(0);
      loadData();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const handleEdit = async (product) => {
    setEditingId(product.id);
    setLoading(true);
    try {
      const res = await fetchJSON(`/api/admin/products/${product.id}`);
      const item = res.data || product;
      const images = item.images?.length ? item.images : item.image_url ? [item.image_url] : [];
      setForm({
        title_en: item.title_en || "",
        title_fa: item.title_fa || "",
        title_ar: item.title_ar || "",
        description: item.description || "",
        price: item.price || "",
        image_url: item.image_url || "",
        image_urls: images,
        category_id: item.category_id ? String(item.category_id) : "",
        is_popular: Boolean(item.is_popular)
      });
      setSelectedImageIndex(0);
    } catch (err) {
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
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
    <div className="space-y-6">
      <section className="panel-card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-display text-xl">{t("panelProducts.title")}</h2>
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
                {t("form.price")}
                <input
                  type="number"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.price}
                  onChange={(event) => setForm({ ...form, price: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70 md:col-span-2">
                {t("form.description")}
                <textarea
                  rows="3"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.description}
                  onChange={(event) => setForm({ ...form, description: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70 md:col-span-2">
                {t("form.images")}
                <div className="mt-2 space-y-3">
                  <input
                    type="file"
                    multiple
                    accept="image/*"
                    className="w-full text-sm text-primary/70 file:mr-4 file:rounded-full file:border-0 file:bg-primary/10 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20"
                    onChange={(e) => handleImageUpload(e.target.files)}
                  />
                  {form.image_urls?.length ? (
                    <div className="space-y-3">
                      <div className="overflow-hidden rounded-2xl border border-primary/15 bg-primary/5">
                        <img
                          src={resolveImageUrl(form.image_urls[selectedImageIndex] || form.image_urls[0])}
                          alt="Preview"
                          className="h-64 w-full object-cover"
                        />
                      </div>
                      <div className="flex gap-2 overflow-x-auto pb-1">
                        {form.image_urls.map((url, index) => (
                          <div key={`${url}-${index}`} className="relative">
                            <button
                              type="button"
                              onClick={() => setSelectedImageIndex(index)}
                              className={`h-16 w-20 overflow-hidden rounded-xl border ${
                                selectedImageIndex === index ? "border-accent" : "border-primary/20"
                              }`}
                            >
                              <img
                                src={resolveImageUrl(url)}
                                alt={`Preview-${index}`}
                                className="h-full w-full object-cover"
                              />
                            </button>
                            <button
                              type="button"
                              onClick={() => handleRemoveImage(index)}
                              className="absolute -right-2 -top-2 rounded-full bg-white/90 px-2 py-1 text-[10px] font-semibold text-primary shadow"
                            >
                              ×
                            </button>
                          </div>
                        ))}
                      </div>
                    </div>
                  ) : (
                    <p className="text-xs text-primary/50">{t("messages.empty")}</p>
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
                    setSelectedImageIndex(0);
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
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <h3 className="font-display text-xl">{t("panelProducts.title")}</h3>
            <span className="rounded-full bg-accent/15 px-3 py-1 text-xs font-semibold text-accent">
              {filterCategory === "all" ? products.length : filteredProducts.length}
            </span>
          </div>
          <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("panelProducts.filterLabel")}
            <select
              className="rounded-full border border-primary/20 bg-white px-3 py-2 text-xs"
              value={filterCategory}
              onChange={(event) => setFilterCategory(event.target.value)}
            >
              <option value="all">{t("panelProducts.allCategories")}</option>
              {categories.map((category) => (
                <option key={category.id} value={String(category.id)}>
                  {category.title_en}
                </option>
              ))}
            </select>
          </div>
        </div>

        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : filteredProducts.length === 0 ? (
          <p className="text-sm text-primary/70">{t("panelProducts.empty")}</p>
        ) : (
          <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
            {filteredProducts.map((product) => (
              <div
                key={product.id}
                className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
              >
                <div className="flex items-center gap-3">
                  <div className="h-12 w-12 overflow-hidden rounded-xl border border-primary/20 bg-primary/5">
                    {product.image_url ? (
                      <img
                        src={resolveImageUrl(product.image_url)}
                        alt={product.title_en}
                        className="h-full w-full object-cover"
                      />
                    ) : null}
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-primary">{product.title_en}</p>
                    <p className="text-xs text-primary/60">{product.title_fa} • {product.title_ar}</p>
                    <p className="text-xs text-primary/40">{product.category?.title_en}</p>
                  </div>
                </div>
                {product.is_popular && (
                  <span className="rounded-full bg-accent/20 px-3 py-1 text-xs font-semibold text-accent">
                    {t("form.popular")}
                  </span>
                )}
                <span className="rounded-full border border-primary/15 px-3 py-1 text-xs font-semibold text-primary/70">
                  {getImageCount(product)} {t("panelProducts.imageCount")}
                </span>
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
      </section>
    </div>
  );
}
