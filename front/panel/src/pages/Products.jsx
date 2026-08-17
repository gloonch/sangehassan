import { useEffect, useMemo, useState } from "react";
import { CheckCircle2, GripVertical, LoaderCircle } from "lucide-react";
import ReorderableImageGrid from "../components/ReorderableImageGrid";
import { moveImage, selectedIndexAfterImageMove } from "../lib/imageOrderDraft";
import { moveProduct } from "../lib/productOrder";
import { useTranslation } from "../lib/i18n";
import { API_BASE, fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { useImageOrderDraft } from "../lib/useImageOrderDraft";

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
  video_url: "",
  category_id: "",
  is_popular: false,
  is_active: true,
  is_indexable: true,
  sample_available: true,
  aliases: [],
  variants: [],
  mines: [],
  finishes: [],
  tone: [],
  use_case_application: [],
  use_case_form: [],
  pattern: [],
  availability: [],
  terms: [],
  term_ids: []
};

const emptySelectedListInputs = {
  variants: "",
  mines: "",
  finishes: "",
  tone: "",
  use_case_application: "",
  use_case_form: "",
  pattern: "",
  availability: ""
};

const hasSEOContent = (value) => String(value || "").replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim().length >= 80;

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
  const [descriptionLang, setDescriptionLang] = useState("en");
  const [newListInputs, setNewListInputs] = useState({ aliases: "" });
  const [selectedListInputs, setSelectedListInputs] = useState(emptySelectedListInputs);
  const [productTerms, setProductTerms] = useState([]);
  const [uploadingVideo, setUploadingVideo] = useState(false);
  const [draggedProductID, setDraggedProductID] = useState(null);
  const [dropTarget, setDropTarget] = useState(null);
  const [orderStatus, setOrderStatus] = useState("idle");
  const { imageOrderDraft, stageImageOrder, resetImageOrderDraft } = useImageOrderDraft();

  const metaListFields = useMemo(() => ([
    { key: "variants", label: t("panelProductMeta.variants") },
    { key: "mines", label: t("panelProductMeta.mines") },
    { key: "finishes", label: t("panelProductMeta.finishes") },
    { key: "tone", label: t("panelProductMeta.tone") },
    { key: "use_case_application", label: t("panelProductMeta.useCaseApplications") },
    { key: "use_case_form", label: t("panelProductMeta.useCaseForms") },
    { key: "pattern", label: t("panelProductMeta.pattern") },
    { key: "availability", label: t("panelProductMeta.availability"), single: true }
  ]), [t]);
  const aliasField = useMemo(() => ({ key: "aliases", label: t("panelProductMeta.aliases") }), [t]);

  const getTermLabel = (term) => {
    if (!term) return "";
    if (lang === "fa") return term.label_fa || term.label_en || "";
    if (lang === "ar") return term.label_ar || term.label_en || "";
    return term.label_en || "";
  };

  const getStoredTermValue = (term) => term?.label_fa || term?.label_en || term?.label_ar || term?.key || "";
  const normalizeTermValue = (value) => String(value || "").trim().toLowerCase();

  const termsByTaxonomy = useMemo(() => {
    const grouped = Object.fromEntries(metaListFields.map((field) => [field.key, []]));
    for (const term of productTerms) {
      if (term.is_active === false) continue;
      if (!grouped[term.taxonomy]) continue;
      grouped[term.taxonomy].push(term);
    }
    return grouped;
  }, [metaListFields, productTerms]);

  const findTermForValue = (field, value) => {
    const needle = normalizeTermValue(value);
    if (!needle) return null;
    return (termsByTaxonomy[field] || []).find((term) =>
      [term.key, term.label_en, term.label_fa, term.label_ar].some((candidate) => normalizeTermValue(candidate) === needle)
    );
  };

  const getMetaListItemLabel = (field, value) => getTermLabel(findTermForValue(field, value)) || value;

  const getValuesFromTerms = (item, field) => {
    const termValues = (item.terms || [])
      .filter((term) => term.taxonomy === field)
      .map((term) => getStoredTermValue(term))
      .filter(Boolean);
    return termValues.length ? termValues : item[field] || [];
  };

  const getPayloadTermIDs = () => {
    const metaTaxonomies = new Set(metaListFields.map((field) => field.key));
    const preservedTermIDs = (form.terms || [])
      .filter((term) => !metaTaxonomies.has(term.taxonomy))
      .map((term) => term.id)
      .filter(Boolean);
    const selectedTermIDs = metaListFields
      .flatMap((field) => (form[field.key] || []).map((value) => findTermForValue(field.key, value)?.id))
      .filter(Boolean);
    return [...new Set([...preservedTermIDs, ...selectedTermIDs])];
  };

  const loadData = async () => {
    try {
      const [productRes, categoryRes, termRes] = await Promise.all([
        fetchJSON("/api/admin/products"),
        fetchJSON("/api/admin/categories"),
        fetchJSON("/api/admin/product-terms")
      ]);
      setProducts(productRes.data || []);
      setCategories(categoryRes.data || []);
      setProductTerms(termRes.data || []);
      setError("");
    } catch (err) {
      setProducts([]);
      setCategories([]);
      setProductTerms([]);
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

  const clearProductDrag = () => {
    setDraggedProductID(null);
    setDropTarget(null);
  };

  const handleProductDragStart = (event, productID) => {
    if (filterCategory !== "all" || orderStatus === "saving") {
      event.preventDefault();
      return;
    }
    setOrderStatus("idle");
    setDraggedProductID(productID);
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", String(productID));
  };

  const handleProductDragOver = (event, targetID) => {
    if (filterCategory !== "all" || draggedProductID == null || orderStatus === "saving") return;
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    const bounds = event.currentTarget.getBoundingClientRect();
    const placement = event.clientY < bounds.top + bounds.height / 2 ? "before" : "after";
    setDropTarget({ id: targetID, placement });
  };

  const handleProductDrop = async (event, targetID) => {
    event.preventDefault();
    const placement = dropTarget?.id === targetID ? dropTarget.placement : "before";
    const nextProducts = moveProduct(products, draggedProductID, targetID, placement);
    clearProductDrag();
    if (nextProducts === products) return;

    const previousProducts = products;
    setProducts(nextProducts);
    setOrderStatus("saving");
    setError("");
    try {
      await fetchJSON("/api/admin/products/order", {
        method: "PUT",
        body: JSON.stringify({ product_ids: nextProducts.map((product) => product.id) })
      });
      setOrderStatus("saved");
    } catch (_) {
      setProducts(previousProducts);
      setOrderStatus("error");
      setError(t("panelProducts.orderSaveFailed"));
    }
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
      const currentImages = form.image_urls || [];
      const nextImages = [...currentImages, ...uploads].filter(Boolean);
      if (imageOrderDraft) {
        stageImageOrder(currentImages, nextImages);
      }
      setForm((prev) => ({
        ...prev,
        image_urls: nextImages,
        image_url: nextImages[0] || prev.image_url
      }));
    } catch (err) {
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  const handleVideoUpload = async (file) => {
    if (!file) return;
    setError("");
    setUploadingVideo(true);
    try {
      const formData = new FormData();
      formData.append("file", file);
      const res = await fetch(`${API_BASE}/api/admin/upload/product`, {
        method: "POST",
        body: formData,
        credentials: "include",
      });
      if (!res.ok) throw new Error("Upload failed");
      const data = await res.json();
      const videoUrl = data?.data?.video_url || data?.data?.file_url || "";
      if (!videoUrl) throw new Error("Upload failed");
      setForm((prev) => ({ ...prev, video_url: videoUrl }));
    } catch (_) {
      setError(t("messages.error"));
    } finally {
      setUploadingVideo(false);
    }
  };

  const handleRemoveImage = (index) => {
    const currentImages = form.image_urls || [];
    const nextImages = currentImages.filter((_, idx) => idx !== index);
    if (imageOrderDraft) {
      stageImageOrder(currentImages, nextImages);
    }
    setForm((prev) => ({
      ...prev,
      image_urls: nextImages,
      image_url: nextImages[0] || ""
    }));
    if (selectedImageIndex === index) {
      setSelectedImageIndex(0);
    }
  };

  const handleMoveImage = (index, direction) => {
    const currentImages = form.image_urls || [];
    const nextImages = moveImage(currentImages, index, direction);
    if (nextImages === currentImages) return;
    stageImageOrder(currentImages, nextImages);
    setForm((prev) => ({
      ...prev,
      image_urls: nextImages,
      image_url: nextImages[0] || ""
    }));
    setSelectedImageIndex((current) => selectedIndexAfterImageMove(current, index, direction));
  };

  const addListItem = (field) => {
    const value = String(newListInputs[field] || "").trim();
    if (!value) return;
    setForm((prev) => {
      const existing = Array.isArray(prev[field]) ? prev[field] : [];
      if (existing.includes(value)) return prev;
      return { ...prev, [field]: [...existing, value] };
    });
    setNewListInputs((prev) => ({ ...prev, [field]: "" }));
  };

  const addSelectedListItem = (field) => {
    const value = String(selectedListInputs[field] || "").trim();
    if (!value) return;
    setForm((prev) => {
      const fieldConfig = metaListFields.find((item) => item.key === field);
      const existing = fieldConfig?.single ? [] : Array.isArray(prev[field]) ? prev[field] : [];
      if (existing.some((item) => normalizeTermValue(item) === normalizeTermValue(value))) return prev;
      return { ...prev, [field]: [...existing, value] };
    });
    setSelectedListInputs((prev) => ({ ...prev, [field]: "" }));
  };

  const removeListItem = (field, index) => {
    setForm((prev) => {
      const existing = Array.isArray(prev[field]) ? prev[field] : [];
      return { ...prev, [field]: existing.filter((_, idx) => idx !== index) };
    });
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
      is_active: form.is_active !== false,
      is_indexable: form.is_indexable !== false,
      sample_available: form.sample_available !== false,
      image_url: images[0] || form.image_url || "",
      image_urls: images,
      video_url: form.video_url || ""
    };
    payload.term_ids = getPayloadTermIDs();

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
      setNewListInputs({ aliases: "" });
      setSelectedListInputs(emptySelectedListInputs);
      resetImageOrderDraft();
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
        description_html_en: item.description_html_en || item.description_html || item.description || "",
        description_html_fa: item.description_html_fa || "",
        description_html_ar: item.description_html_ar || "",
        short_description_html_en: item.short_description_html_en || item.short_description_html || "",
        short_description_html_fa: item.short_description_html_fa || "",
        short_description_html_ar: item.short_description_html_ar || "",
        price: item.price || "",
        image_url: item.image_url || "",
        image_urls: images,
        video_url: item.video_url || "",
        category_id: item.category_id ? String(item.category_id) : "",
        is_popular: Boolean(item.is_popular),
        is_active: item.is_active !== false,
        is_indexable: item.is_indexable !== false,
        sample_available: item.sample_available !== false,
        aliases: item.aliases || [],
        variants: getValuesFromTerms(item, "variants"),
        mines: getValuesFromTerms(item, "mines"),
        finishes: getValuesFromTerms(item, "finishes"),
        tone: getValuesFromTerms(item, "tone"),
        use_case_application: getValuesFromTerms(item, "use_case_application"),
        use_case_form: getValuesFromTerms(item, "use_case_form"),
        pattern: getValuesFromTerms(item, "pattern"),
        availability: getValuesFromTerms(item, "availability"),
        terms: item.terms || [],
        term_ids: item.term_ids || []
      });
      setSelectedImageIndex(0);
      setDescriptionLang("en");
      setSelectedListInputs(emptySelectedListInputs);
      resetImageOrderDraft(images);
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
                  {/* Tags & use-cases removed */}
                </div>
              </div>
              <div className="rounded-2xl border border-primary/15 bg-primary/5 p-4 md:col-span-2">
                <p className="text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("panelProductMeta.detailsTitle")}
                </p>
                <div className="mt-4 grid gap-4 md:grid-cols-2">
                  {metaListFields.map((field) => {
                    const options = termsByTaxonomy[field.key] || [];
                    return (
                      <div key={field.key} className="space-y-2">
                        <label className="block text-sm font-semibold text-primary">{field.label}</label>
                        <div className="flex items-center gap-2">
                          <select
                            className="w-full rounded-xl border border-primary/20 bg-white px-3 py-2 text-sm"
                            value={selectedListInputs[field.key] || ""}
                            onChange={(event) =>
                              setSelectedListInputs((prev) => ({ ...prev, [field.key]: event.target.value }))
                            }
                            disabled={options.length === 0}
                          >
                            <option value="">
                              {options.length === 0 ? t("panelProductMeta.noTerms") : t("panelProductMeta.selectTerm")}
                            </option>
                            {options.map((term) => {
                              const value = getStoredTermValue(term);
                              return (
                                <option key={term.id} value={value}>
                                  {getTermLabel(term)}
                                </option>
                              );
                            })}
                          </select>
                          <button
                            type="button"
                            onClick={() => addSelectedListItem(field.key)}
                            disabled={options.length === 0}
                            className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
                          >
                            {t("actions.add")}
                          </button>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          {(form[field.key] || []).map((item, index) => (
                            <span
                              key={`${field.key}-${item}-${index}`}
                              className="flex items-center gap-2 rounded-full bg-white px-3 py-1 text-xs font-semibold text-primary/70 shadow-sm"
                            >
                              {getMetaListItemLabel(field.key, item)}
                              <button
                                type="button"
                                onClick={() => removeListItem(field.key, index)}
                                className="text-primary/50 transition hover:text-primary"
                              >
                                ×
                              </button>
                            </span>
                          ))}
                          {(form[field.key] || []).length === 0 && (
                            <p className="text-xs text-primary/50">{t("messages.empty")}</p>
                          )}
                        </div>
                      </div>
                    );
                  })}
                  <div className="space-y-2">
                    <label className="block text-sm font-semibold text-primary">{aliasField.label}</label>
                    <div className="flex items-center gap-2">
                      <input
                        type="text"
                        className="w-full rounded-xl border border-primary/20 bg-white px-3 py-2 text-sm"
                        placeholder={t("panelProductMeta.newTermPlaceholder")}
                        value={newListInputs.aliases || ""}
                        onChange={(event) =>
                          setNewListInputs((prev) => ({ ...prev, aliases: event.target.value }))
                        }
                        onKeyDown={(event) => {
                          if (event.key === "Enter") {
                            event.preventDefault();
                            addListItem("aliases");
                          }
                        }}
                      />
                      <button
                        type="button"
                        onClick={() => addListItem("aliases")}
                        className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
                      >
                        {t("actions.addNew")}
                      </button>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {(form.aliases || []).map((item, index) => (
                        <span
                          key={`aliases-${item}-${index}`}
                          className="flex items-center gap-2 rounded-full bg-white px-3 py-1 text-xs font-semibold text-primary/70 shadow-sm"
                        >
                          {item}
                          <button
                            type="button"
                            onClick={() => removeListItem("aliases", index)}
                            className="text-primary/50 transition hover:text-primary"
                          >
                            ×
                          </button>
                        </span>
                      ))}
                      {(form.aliases || []).length === 0 && (
                        <p className="text-xs text-primary/50">{t("messages.empty")}</p>
                      )}
                    </div>
                  </div>
                </div>
              </div>
              <div className="block text-xs font-semibold uppercase tracking-wide text-primary/70 md:col-span-2">
                <label htmlFor="product-images-upload">{t("form.images")}</label>
                <div className="mt-2 space-y-3">
                  <input
                    id="product-images-upload"
                    type="file"
                    multiple
                    accept="image/*"
                    className="w-full text-sm text-primary/70 file:mr-4 file:rounded-full file:border-0 file:bg-primary/10 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20"
                    onChange={(e) => handleImageUpload(e.target.files)}
                  />
                  <ReorderableImageGrid
                    images={form.image_urls}
                    selectedIndex={selectedImageIndex}
                    onSelect={setSelectedImageIndex}
                    onRemove={handleRemoveImage}
                    onMove={handleMoveImage}
                    previewClassName="h-64"
                    thumbnailClassName="h-16"
                    gridClassName="grid grid-cols-3 gap-3 sm:grid-cols-4 md:grid-cols-6"
                    labels={{
                      empty: t("messages.empty"),
                      remove: t("actions.delete"),
                      moveLeft: t("actions.moveLeft"),
                      moveRight: t("actions.moveRight")
                    }}
                  />
                </div>
              </div>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70 md:col-span-2">
                {t("form.video")}
                <div className="mt-2 space-y-3">
                  <input
                    type="file"
                    accept="video/mp4,video/webm,video/quicktime,video/x-m4v"
                    className="w-full text-sm text-primary/70 file:mr-4 file:rounded-full file:border-0 file:bg-primary/10 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20"
                    disabled={uploadingVideo}
                    onChange={(event) => handleVideoUpload(event.target.files?.[0])}
                  />
                  {form.video_url ? (
                    <div className="space-y-3">
                      <video className="h-64 w-full rounded-2xl border border-primary/15 bg-primary/5 object-cover" controls preload="metadata">
                        <source src={resolveImageUrl(form.video_url)} />
                      </video>
                      <button
                        type="button"
                        onClick={() => setForm((prev) => ({ ...prev, video_url: "" }))}
                        className="rounded-full border border-primary/20 px-4 py-1 text-xs font-semibold text-primary/70"
                      >
                        {t("actions.delete")}
                      </button>
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
              <label className="flex items-center gap-3 text-xs font-semibold uppercase tracking-wide text-primary/70">
                <input type="checkbox" checked={form.is_active} onChange={(event) => setForm({ ...form, is_active: event.target.checked })} className="h-4 w-4 rounded border-primary/20" />
                {t("form.active")}
              </label>
              <label className="flex items-center gap-3 text-xs font-semibold uppercase tracking-wide text-primary/70">
                <input type="checkbox" checked={form.is_indexable} onChange={(event) => setForm({ ...form, is_indexable: event.target.checked })} className="h-4 w-4 rounded border-primary/20" />
                {t("form.indexable")}
              </label>
              <label className="flex items-center gap-3 text-xs font-semibold uppercase tracking-wide text-primary/70">
                <input type="checkbox" checked={form.sample_available} onChange={(event) => setForm({ ...form, sample_available: event.target.checked })} className="h-4 w-4 rounded border-primary/20" />
                {t("form.sampleAvailable")}
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
                  setNewListInputs({ aliases: "" });
                  setSelectedListInputs(emptySelectedListInputs);
                  resetImageOrderDraft();
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

        <div className="mb-4 flex flex-wrap items-center justify-between gap-2 rounded-xl border border-primary/10 bg-primary/[0.03] px-4 py-3 text-xs text-primary/65">
          <p>
            {filterCategory === "all"
              ? t("panelProducts.orderHint")
              : t("panelProducts.reorderAllCategories")}
          </p>
          {orderStatus === "saving" && (
            <span className="flex items-center gap-2 font-semibold text-primary">
              <LoaderCircle className="h-4 w-4 animate-spin" aria-hidden="true" />
              {t("panelProducts.orderSaving")}
            </span>
          )}
          {orderStatus === "saved" && (
            <span className="flex items-center gap-2 font-semibold text-emerald-700">
              <CheckCircle2 className="h-4 w-4" aria-hidden="true" />
              {t("panelProducts.orderSaved")}
            </span>
          )}
          {orderStatus === "error" && (
            <span className="font-semibold text-red-600">{t("panelProducts.orderSaveFailed")}</span>
          )}
        </div>

        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : filteredProducts.length === 0 ? (
          <p className="text-sm text-primary/70">{t("panelProducts.empty")}</p>
        ) : (
          <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
            {filteredProducts.map((product) => {
              const localeReadiness = ["fa", "en", "ar"].map((locale) => ({
                locale,
                ready: Boolean(product[`title_${locale}`]) && hasSEOContent(
                  product[`description_html_${locale}`] || product[`short_description_html_${locale}`]
                )
              }));
              return (
              <div
                key={product.id}
                onDragOver={(event) => handleProductDragOver(event, product.id)}
                onDrop={(event) => handleProductDrop(event, product.id)}
                className={`relative flex flex-wrap items-center justify-between gap-4 rounded-xl border bg-white/80 px-4 py-3 transition ${draggedProductID === product.id ? "opacity-50" : ""} ${dropTarget?.id === product.id && dropTarget.placement === "before" ? "border-t-4 border-t-accent" : "border-primary/10"} ${dropTarget?.id === product.id && dropTarget.placement === "after" ? "border-b-4 border-b-accent" : ""}`}
              >
                <div className="flex items-center gap-3">
                  <button
                    type="button"
                    draggable={filterCategory === "all" && orderStatus !== "saving"}
                    onDragStart={(event) => handleProductDragStart(event, product.id)}
                    onDragEnd={clearProductDrag}
                    disabled={filterCategory !== "all" || orderStatus === "saving"}
                    className="flex h-10 w-8 shrink-0 cursor-grab items-center justify-center rounded-lg border border-primary/10 text-primary/45 transition hover:border-accent/40 hover:text-accent active:cursor-grabbing disabled:cursor-not-allowed disabled:opacity-30"
                    aria-label={`${t("panelProducts.dragProduct")} ${product.title_fa || product.title_en}`}
                    title={t("panelProducts.dragProduct")}
                  >
                    <GripVertical className="h-5 w-5" aria-hidden="true" />
                  </button>
                  <span className="flex h-8 min-w-8 items-center justify-center rounded-full bg-primary/5 px-2 text-xs font-bold tabular-nums text-primary/60" title={t("panelProducts.position")}>
                    {products.findIndex((item) => item.id === product.id) + 1}
                  </span>
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
                    <div className="mt-2 flex flex-wrap gap-1.5">
                      {localeReadiness.map((item) => (
                        <span key={item.locale} className={`rounded-full border px-2 py-0.5 text-[10px] font-semibold uppercase ${item.ready ? "border-emerald-200 bg-emerald-50 text-emerald-700" : "border-amber-200 bg-amber-50 text-amber-700"}`}>
                          {item.locale} {item.ready ? "SEO ready" : "needs content"}
                        </span>
                      ))}
                    </div>
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
                {product.video_url && (
                  <span className="rounded-full border border-primary/15 px-3 py-1 text-xs font-semibold text-primary/70">
                    {t("form.video")}
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
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
}
