import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { API_BASE, fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const emptyForm = {
  title_en: "",
  title_fa: "",
  title_ar: "",
  description_html_en: "",
  description_html_fa: "",
  description_html_ar: "",
  short_description_html_en: "",
  short_description_html_fa: "",
  short_description_html_ar: "",
  price: "",
  image_url: "",
  image_urls: [],
  category_id: "",
  is_popular: false,
  term_ids: []
};

export default function Products() {
  const { t, lang } = useTranslation();
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filterCategory, setFilterCategory] = useState("all");
  const [selectedImageIndex, setSelectedImageIndex] = useState(0);
  const [formOpen, setFormOpen] = useState(true);
  const [termsByTaxonomy, setTermsByTaxonomy] = useState({});
  const [newTermLabels, setNewTermLabels] = useState({});
  const [descriptionLang, setDescriptionLang] = useState("en");

	  const termTaxonomies = useMemo(() => ([
	    { taxonomy: "stone_type", label: t("panelProductMeta.stoneType"), multiple: false },
	    { taxonomy: "visual_impact", label: t("panelProductMeta.visualImpact"), multiple: false },
	    { taxonomy: "tone", label: t("panelProductMeta.tone"), multiple: true },
	    { taxonomy: "pattern", label: t("panelProductMeta.pattern"), multiple: true },
	    { taxonomy: "use_case_space", label: t("panelProductMeta.useCaseSpaces"), multiple: true },
	    { taxonomy: "use_case_form", label: t("panelProductMeta.useCaseForms"), multiple: true },
	    { taxonomy: "use_case_application", label: t("panelProductMeta.useCaseApplications"), multiple: true },
	    { taxonomy: "use_case_project_type", label: t("panelProductMeta.useCaseProjects"), multiple: true },
	    { taxonomy: "use_case_special", label: t("panelProductMeta.useCaseSpecial"), multiple: true },
	  ]), [t]);

  const getTermLabel = (term) => {
    if (!term) return "";
    if (lang === "fa") return term.label_fa || term.label_en || "";
    if (lang === "ar") return term.label_ar || term.label_en || "";
    return term.label_en || "";
  };

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

  const loadTerms = async () => {
    try {
      const results = await Promise.all(
        termTaxonomies.map((entry) =>
          fetchJSON(`/api/admin/product-terms?taxonomy=${encodeURIComponent(entry.taxonomy)}`)
        )
      );
      const grouped = {};
      termTaxonomies.forEach((entry, index) => {
        grouped[entry.taxonomy] = results[index]?.data || [];
      });
      setTermsByTaxonomy(grouped);
    } catch (err) {
      setTermsByTaxonomy({});
    }
  };

  useEffect(() => {
    loadData();
    loadTerms();
  }, [termTaxonomies]);

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

  const toggleTerm = (taxonomy, termId) => {
    const config = termTaxonomies.find((entry) => entry.taxonomy === taxonomy);
    const isMultiple = Boolean(config?.multiple);
    const termIdsInTax = (termsByTaxonomy[taxonomy] || []).map((term) => term.id);

    setForm((prev) => {
      const current = new Set(prev.term_ids || []);
      const isSelected = current.has(termId);

      if (!isMultiple && !isSelected) {
        for (const existingId of termIdsInTax) current.delete(existingId);
      }

      if (isSelected) {
        current.delete(termId);
      } else {
        current.add(termId);
      }

      return { ...prev, term_ids: Array.from(current) };
    });
  };

  const createTerm = async (taxonomy) => {
    const label = String(newTermLabels[taxonomy] || "").trim();
    if (!label) return;

    setError("");
    try {
      await fetchJSON("/api/admin/product-terms", {
        method: "POST",
        body: JSON.stringify({ taxonomy, label_en: label })
      });
      setNewTermLabels((prev) => ({ ...prev, [taxonomy]: "" }));
      loadTerms();
    } catch (err) {
      setError(t("messages.error"));
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
      setDescriptionLang("en");
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
      const termIds = item.term_ids?.length
        ? item.term_ids
        : item.terms?.length
          ? item.terms.map((term) => term.id)
          : [];
      setForm({
        title_en: item.title_en || "",
        title_fa: item.title_fa || "",
        title_ar: item.title_ar || "",
        description_html_en: item.description_html_en || item.description_html || item.description || "",
        description_html_fa: item.description_html_fa || "",
        description_html_ar: item.description_html_ar || "",
        short_description_html_en: item.short_description_html_en || item.short_description_html || "",
        short_description_html_fa: item.short_description_html_fa || "",
        short_description_html_ar: item.short_description_html_ar || "",
        price: item.price || "",
        image_url: item.image_url || "",
        image_urls: images,
        category_id: item.category_id ? String(item.category_id) : "",
        is_popular: Boolean(item.is_popular),
        term_ids: termIds
      });
      setSelectedImageIndex(0);
      setDescriptionLang("en");
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
              <div className="rounded-2xl border border-primary/15 bg-primary/5 p-4 md:col-span-2">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <p className="text-xs font-semibold uppercase tracking-wide text-primary/70">
                    {t("form.description")}
                  </p>
                  <div className="flex items-center gap-2">
                    {["en", "fa", "ar"].map((code) => (
                      <button
                        key={code}
                        type="button"
                        onClick={() => setDescriptionLang(code)}
                        className={`rounded-full border px-3 py-1 text-xs font-semibold transition ${
                          descriptionLang === code
                            ? "border-accent bg-accent text-white"
                            : "border-primary/20 bg-white text-primary/70 hover:border-primary/40"
                        }`}
                      >
                        {code.toUpperCase()}
                      </button>
                    ))}
                  </div>
                </div>

                <label className="mt-4 block text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("form.description")} (HTML)
                  <textarea
                    rows="4"
                    className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                    value={form[`description_html_${descriptionLang}`] || ""}
                    onChange={(event) =>
                      setForm({ ...form, [`description_html_${descriptionLang}`]: event.target.value })
                    }
                  />
                </label>

                <label className="mt-4 block text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("form.shortDescription")} (HTML)
                  <textarea
                    rows="2"
                    className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                    value={form[`short_description_html_${descriptionLang}`] || ""}
                    onChange={(event) =>
                      setForm({ ...form, [`short_description_html_${descriptionLang}`]: event.target.value })
                    }
                  />
                </label>
              </div>

              <div className="rounded-2xl border border-primary/15 bg-primary/5 p-4 md:col-span-2">
                <p className="text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("panelProductMeta.tagsTitle")}
                </p>
                <div className="mt-4 space-y-6">
                  {termTaxonomies.map((entry) => (
                    <div key={entry.taxonomy} className="space-y-3">
                      <div className="flex flex-wrap items-center justify-between gap-3">
                        <p className="text-sm font-semibold text-primary">
                          {entry.label}
                          {!entry.multiple && (
                            <span className="ml-2 text-xs font-semibold text-primary/50">
                              ({t("panelProductMeta.singleSelect")})
                            </span>
                          )}
                        </p>
                        <div className="flex flex-wrap items-center gap-2">
                          <input
                            type="text"
                            className="w-56 rounded-xl border border-primary/20 bg-white px-3 py-2 text-sm"
                            placeholder={t("panelProductMeta.newTermPlaceholder")}
                            value={newTermLabels[entry.taxonomy] || ""}
                            onChange={(event) =>
                              setNewTermLabels((prev) => ({ ...prev, [entry.taxonomy]: event.target.value }))
                            }
                          />
                          <button
                            type="button"
                            onClick={() => createTerm(entry.taxonomy)}
                            className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
                          >
                            {t("panelProductMeta.addTerm")}
                          </button>
                        </div>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        {(termsByTaxonomy[entry.taxonomy] || []).map((term) => {
                          const selected = form.term_ids?.includes(term.id);
                          return (
                            <button
                              key={`${entry.taxonomy}-${term.id}`}
                              type="button"
                              onClick={() => toggleTerm(entry.taxonomy, term.id)}
                              className={`rounded-full border px-3 py-1 text-xs font-semibold transition ${
                                selected
                                  ? "border-accent bg-accent text-white"
                                  : "border-primary/20 bg-white text-primary/70 hover:border-primary/40"
                              }`}
                            >
                              {getTermLabel(term)}
                            </button>
                          );
                        })}
                        {(termsByTaxonomy[entry.taxonomy] || []).length === 0 && (
                          <p className="text-xs text-primary/60">{t("panelProductMeta.noTerms")}</p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
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
                    setDescriptionLang("en");
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
