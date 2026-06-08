import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { API_BASE, fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const MAX_VIDEO_UPLOAD_BYTES = 90 * 1024 * 1024;

const formatFileSize = (bytes) => `${Math.round(bytes / 1024 / 1024)} MB`;

const emptyForm = {
  description_en: "",
  description_fa: "",
  description_ar: "",
  cover_image_url: "",
  video_url: "",
  gallery_images: [],
  product_ids: [],
  sort_order: 0
};

const uploadErrorMessage = (error, t) => {
  const rawMessage = error?.message || "";
  if (/failed to fetch|network|load failed|connection/i.test(rawMessage)) {
    return t("panelProjects.uploadConnectionLost");
  }
  return rawMessage || t("messages.error");
};

export default function Projects() {
  const { t, lang } = useTranslation();
  const [projects, setProjects] = useState([]);
  const [products, setProducts] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [formOpen, setFormOpen] = useState(true);
  const [productSearch, setProductSearch] = useState("");
  const [uploadingCover, setUploadingCover] = useState(false);
  const [uploadingGallery, setUploadingGallery] = useState(false);
  const [uploadingVideo, setUploadingVideo] = useState(false);

  const loadProjects = async () => {
    try {
      const [projectResponse, productResponse] = await Promise.all([
        fetchJSON("/api/admin/projects"),
        fetchJSON("/api/admin/products")
      ]);
      setProjects(projectResponse.data || []);
      setProducts(productResponse.data || []);
      setError("");
    } catch (_) {
      setProjects([]);
      setProducts([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProjects();
  }, []);

  const getProductTitle = (product) => {
    if (!product) return "";
    if (lang === "fa") return product.title_fa || product.title_en || product.title_ar || "";
    if (lang === "ar") return product.title_ar || product.title_en || product.title_fa || "";
    return product.title_en || product.title_fa || product.title_ar || "";
  };

  const getProductCategoryTitle = (product) => {
    const category = product?.category;
    if (!category) return "";
    if (lang === "fa") return category.title_fa || category.title_en || category.title_ar || "";
    if (lang === "ar") return category.title_ar || category.title_en || category.title_fa || "";
    return category.title_en || category.title_fa || category.title_ar || "";
  };

  const selectedProductIds = useMemo(
    () => (Array.isArray(form.product_ids) ? form.product_ids.map(Number).filter(Boolean) : []),
    [form.product_ids]
  );

  const selectedProducts = useMemo(() => {
    const selected = new Set(selectedProductIds);
    return products.filter((product) => selected.has(Number(product.id)));
  }, [products, selectedProductIds]);

  const filteredProductOptions = useMemo(() => {
    const selected = new Set(selectedProductIds);
    const query = productSearch.trim().toLowerCase();

    return products
      .filter((product) => {
        if (selected.has(Number(product.id))) return false;
        if (!query) return true;
        const haystack = [
          product.title_en,
          product.title_fa,
          product.title_ar,
          product.slug,
          product.category?.title_en,
          product.category?.title_fa,
          product.category?.title_ar
        ]
          .filter(Boolean)
          .join(" ")
          .toLowerCase();
        return haystack.includes(query);
      })
      .slice(0, 40);
  }, [productSearch, products, selectedProductIds]);

  const addProductToProject = (productId) => {
    const normalizedId = Number(productId);
    if (!normalizedId) return;
    setForm((prev) => {
      const current = Array.isArray(prev.product_ids) ? prev.product_ids.map(Number).filter(Boolean) : [];
      if (current.includes(normalizedId)) return prev;
      return { ...prev, product_ids: [...current, normalizedId] };
    });
  };

  const removeProductFromProject = (productId) => {
    const normalizedId = Number(productId);
    setForm((prev) => ({
      ...prev,
      product_ids: (Array.isArray(prev.product_ids) ? prev.product_ids : [])
        .map(Number)
        .filter((id) => id && id !== normalizedId)
    }));
  };

  const uploadFile = async (file) => {
    const formData = new FormData();
    formData.append("file", file);

    let res;
    try {
      res = await fetch(`${API_BASE}/api/admin/upload/project`, {
        method: "POST",
        body: formData,
        credentials: "include"
      });
    } catch (error) {
      throw new Error(uploadErrorMessage(error, t));
    }
    if (!res.ok) {
      const message = res.status === 413 ? t("panelProjects.uploadTooLarge") : t("panelProjects.uploadFailed");
      throw new Error(message);
    }
    const data = await res.json();
    return data?.data || {};
  };

  const handleCoverUpload = async (file) => {
    if (!file) return;
    setUploadingCover(true);
    setError("");
    try {
      const uploaded = await uploadFile(file);
      const imageUrl = uploaded?.image_url || "";
      if (!imageUrl) throw new Error("Upload failed");
      setForm((prev) => ({ ...prev, cover_image_url: imageUrl }));
    } catch (_) {
      setError(t("messages.error"));
    } finally {
      setUploadingCover(false);
    }
  };

  const handleGalleryUpload = async (files) => {
    if (!files || files.length === 0) return;
    const selected = Array.from(files);

    setUploadingGallery(true);
    setError("");
    try {
      const uploaded = await Promise.all(selected.map((file) => uploadFile(file)));
      setForm((prev) => ({
        ...prev,
        gallery_images: [...prev.gallery_images, ...uploaded.map((item) => item?.image_url || "").filter(Boolean)]
      }));
    } catch (_) {
      setError(t("messages.error"));
    } finally {
      setUploadingGallery(false);
    }
  };

  const handleRemoveGalleryImage = (index) => {
    setForm((prev) => ({
      ...prev,
      gallery_images: prev.gallery_images.filter((_, idx) => idx !== index)
    }));
  };

  const handleVideoUpload = async (file) => {
    if (!file) return;
    if (file.size > MAX_VIDEO_UPLOAD_BYTES) {
      setError(t("panelProjects.videoTooLarge").replace("{size}", formatFileSize(MAX_VIDEO_UPLOAD_BYTES)));
      return;
    }
    setUploadingVideo(true);
    setError("");
    try {
      const uploaded = await uploadFile(file);
      const videoUrl = uploaded?.video_url || uploaded?.file_url || "";
      if (!videoUrl) {
        throw new Error("Upload failed");
      }
      setForm((prev) => ({ ...prev, video_url: videoUrl }));
    } catch (error) {
      setError(uploadErrorMessage(error, t));
    } finally {
      setUploadingVideo(false);
    }
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (!form.cover_image_url) {
      setError(t("panelProjects.coverRequired"));
      return;
    }

    const payload = {
      description_en: form.description_en,
      description_fa: form.description_fa,
      description_ar: form.description_ar,
      cover_image_url: form.cover_image_url,
      video_url: form.video_url,
      gallery_images: form.gallery_images,
      product_ids: selectedProductIds,
      sort_order: Number(form.sort_order) || 0
    };

    setError("");
    try {
      if (editingId) {
        await fetchJSON(`/api/admin/projects/${editingId}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
      } else {
        await fetchJSON("/api/admin/projects", {
          method: "POST",
          body: JSON.stringify(payload)
        });
      }

      setForm(emptyForm);
      setEditingId(null);
      loadProjects();
    } catch (_) {
      setError(t("messages.error"));
    }
  };

  const handleEdit = async (project) => {
    setEditingId(project.id);
    setLoading(true);
    setError("");
    try {
      const response = await fetchJSON(`/api/admin/projects/${project.id}`);
      const item = response.data || project;
      const productIds = Array.isArray(item.product_ids)
        ? item.product_ids
        : Array.isArray(item.used_products)
          ? item.used_products.map((product) => product.id)
          : [];
      setForm({
        description_en: item.description_en || item.description || "",
        description_fa: item.description_fa || "",
        description_ar: item.description_ar || "",
        cover_image_url: item.cover_image_url || "",
        video_url: item.video_url || "",
        gallery_images: item.gallery_images || [],
        product_ids: productIds.map(Number).filter(Boolean),
        sort_order: item.sort_order || 0
      });
    } catch (_) {
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await fetchJSON(`/api/admin/projects/${id}`, { method: "DELETE" });
      loadProjects();
    } catch (_) {
      setError(t("messages.error"));
    }
  };

  return (
    <div className="space-y-6">
      <section className="panel-card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-display text-xl">{t("panelProjects.title")}</h2>
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
            <div className="grid gap-4 md:grid-cols-3">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.descriptionEn")}
                <textarea
                  rows={4}
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.description_en}
                  onChange={(event) => setForm((prev) => ({ ...prev, description_en: event.target.value }))}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.descriptionFa")}
                <textarea
                  rows={4}
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.description_fa}
                  onChange={(event) => setForm((prev) => ({ ...prev, description_fa: event.target.value }))}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.descriptionAr")}
                <textarea
                  rows={4}
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.description_ar}
                  onChange={(event) => setForm((prev) => ({ ...prev, description_ar: event.target.value }))}
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-3">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.coverImageUrl")}
                <input
                  type="file"
                  accept="image/*"
                  className="mt-2 w-full text-sm text-primary/70 file:mr-4 file:rounded-full file:border-0 file:bg-primary/10 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20"
                  disabled={uploadingCover}
                  onChange={(event) => handleCoverUpload(event.target.files?.[0])}
                />
                {form.cover_image_url ? (
                  <div className="mt-3 h-24 w-24 overflow-hidden rounded-xl border border-primary/15 bg-primary/5">
                    <img
                      src={resolveImageUrl(form.cover_image_url)}
                      alt="Cover"
                      className="h-full w-full object-cover"
                    />
                  </div>
                ) : null}
              </label>

              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.video")}
                <input
                  type="file"
                  accept="video/mp4,video/webm,video/quicktime,video/x-m4v"
                  className="mt-2 w-full text-sm text-primary/70 file:mr-4 file:rounded-full file:border-0 file:bg-primary/10 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20"
                  disabled={uploadingVideo}
                  onChange={(event) => handleVideoUpload(event.target.files?.[0])}
                />
                {form.video_url ? (
                  <div className="mt-3 space-y-2">
                    <video className="h-24 w-24 rounded-xl border border-primary/15 bg-primary/5 object-cover" controls preload="metadata">
                      <source src={resolveImageUrl(form.video_url)} />
                    </video>
                    <button
                      type="button"
                      className="rounded-full border border-primary/20 px-3 py-1 text-[10px] font-semibold text-primary/70"
                      onClick={() => setForm((prev) => ({ ...prev, video_url: "" }))}
                    >
                      {t("actions.delete")}
                    </button>
                  </div>
                ) : null}
              </label>

              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.orderIndex")}
                <input
                  type="number"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.sort_order}
                  onChange={(event) => setForm((prev) => ({ ...prev, sort_order: event.target.value }))}
                />
              </label>
            </div>

            <div className="rounded-2xl border border-primary/15 bg-primary/5 p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-wide text-primary/70">
                    {t("panelProjects.usedProducts")}
                  </p>
                  <p className="mt-1 text-xs text-primary/50">{t("panelProjects.usedProductsHint")}</p>
                </div>
                <span className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-primary/60">
                  {selectedProductIds.length}
                </span>
              </div>

              <div className="mt-4 grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)_auto]">
                <input
                  type="search"
                  className="w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  placeholder={t("panelProjects.productSearchPlaceholder")}
                  value={productSearch}
                  onChange={(event) => setProductSearch(event.target.value)}
                />
                <select
                  className="w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value=""
                  onChange={(event) => addProductToProject(event.target.value)}
                >
                  <option value="">
                    {filteredProductOptions.length
                      ? t("panelProjects.selectProduct")
                      : t("panelProjects.noProductOptions")}
                  </option>
                  {filteredProductOptions.map((product) => {
                    const categoryTitle = getProductCategoryTitle(product);
                    return (
                      <option key={product.id} value={product.id}>
                        {getProductTitle(product)}
                        {categoryTitle ? ` - ${categoryTitle}` : ""}
                      </option>
                    );
                  })}
                </select>
                <button
                  type="button"
                  className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
                  onClick={() => addProductToProject(filteredProductOptions[0]?.id)}
                  disabled={!filteredProductOptions.length}
                >
                  {t("panelProjects.addProduct")}
                </button>
              </div>

              <div className="mt-4 grid gap-2 md:grid-cols-2">
                {selectedProducts.map((product) => (
                  <div
                    key={product.id}
                    className="flex items-center justify-between gap-3 rounded-xl border border-primary/10 bg-white px-3 py-2"
                  >
                    <div className="flex min-w-0 items-center gap-3">
                      <div className="h-10 w-10 flex-shrink-0 overflow-hidden rounded-lg border border-primary/10 bg-primary/5">
                        {product.image_url ? (
                          <img
                            src={resolveImageUrl(product.image_url)}
                            alt={getProductTitle(product)}
                            className="h-full w-full object-cover"
                          />
                        ) : null}
                      </div>
                      <div className="min-w-0">
                        <p className="truncate text-sm font-semibold text-primary">{getProductTitle(product)}</p>
                        <p className="truncate text-xs text-primary/45">{getProductCategoryTitle(product) || product.slug}</p>
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => removeProductFromProject(product.id)}
                      className="rounded-full border border-primary/15 px-2 py-1 text-[10px] font-semibold text-primary/55"
                    >
                      {t("actions.delete")}
                    </button>
                  </div>
                ))}
                {selectedProducts.length === 0 && (
                  <p className="text-xs text-primary/50">{t("panelProjects.noProductsSelected")}</p>
                )}
              </div>
            </div>

            <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
              {t("form.images")} ({form.gallery_images.length})
              <input
                type="file"
                accept="image/*"
                multiple
                className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                disabled={uploadingGallery}
                onChange={(event) => handleGalleryUpload(event.target.files)}
              />
            </label>

            {form.gallery_images.length > 0 && (
              <div className="rounded-2xl border border-primary/10 bg-sand/60 p-4">
                <div className="grid grid-cols-3 gap-2 md:grid-cols-5">
                  {form.gallery_images.map((url, index) => (
                    <div key={`${url}-${index}`} className="relative overflow-hidden rounded-xl border border-primary/15">
                      <img src={resolveImageUrl(url)} alt="Gallery" className="h-20 w-full object-cover" />
                      <button
                        type="button"
                        onClick={() => handleRemoveGalleryImage(index)}
                        className="absolute right-1 top-1 rounded-full bg-white/95 px-1.5 text-[10px] font-semibold text-primary"
                      >
                        x
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}

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
                    setError("");
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
        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : projects.length === 0 ? (
          <p className="text-sm text-primary/70">{t("panelProjects.empty")}</p>
        ) : (
          <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
            {projects.map((project) => (
              <div
                key={project.id}
                className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
              >
                <div className="flex items-center gap-3">
                  <div className="h-14 w-14 overflow-hidden rounded-lg border border-primary/15 bg-primary/5">
                    {project.cover_image_url ? (
                      <img
                        src={resolveImageUrl(project.cover_image_url)}
                        alt="Cover"
                        className="h-full w-full object-cover"
                      />
                    ) : null}
                  </div>
                  <div>
                    <p className="text-xs text-primary/60">#{project.id}</p>
                    <p className="line-clamp-2 max-w-lg text-sm text-primary/80">
                      {(lang === "fa"
                        ? project.description_fa || project.description_en || project.description_ar
                        : lang === "ar"
                          ? project.description_ar || project.description_en || project.description_fa
                          : project.description_en || project.description_fa || project.description_ar) || t("projects.noDescription")}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-2 text-xs text-primary/60">
                  <span>{t("form.orderIndex")}: {project.sort_order || 0}</span>
                  <span>•</span>
                  <span>{project.gallery_count || 0} {t("panelProjects.galleryCount")}</span>
                  {project.video_url ? (
                    <>
                      <span>•</span>
                      <span>{t("form.video")}</span>
                    </>
                  ) : null}
                </div>

                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => handleEdit(project)}
                    className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold text-primary/70"
                  >
                    {t("actions.edit")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(project.id)}
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
