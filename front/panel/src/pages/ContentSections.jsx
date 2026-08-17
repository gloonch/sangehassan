import { useEffect, useMemo, useState } from "react";
import { defaultLandingDesktopImages } from "@shared/landingSlides";
import ReorderableImageGrid from "../components/ReorderableImageGrid";
import { moveImage, selectedIndexAfterImageMove } from "../lib/imageOrderDraft";
import { useTranslation } from "../lib/i18n";
import { API_BASE, fetchJSON } from "../lib/api";
import { useImageOrderDraft } from "../lib/useImageOrderDraft";

const emptyForm = {
  page: "",
  key: "",
  title_en: "",
  title_fa: "",
  title_ar: "",
  subtitle_en: "",
  subtitle_fa: "",
  subtitle_ar: "",
  description_en: "",
  description_fa: "",
  description_ar: "",
  cta_label_en: "",
  cta_label_fa: "",
  cta_label_ar: "",
  cta_href: "",
  order_index: 0,
  is_active: true,
  image_urls: []
};

const defaultSlideImages = (page, key) => {
  if (page !== "home") return [];
  return defaultLandingDesktopImages[key] || [];
};

export default function ContentSections() {
  const { t } = useTranslation();
  const [sections, setSections] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedImageIndex, setSelectedImageIndex] = useState(0);
  const [formOpen, setFormOpen] = useState(true);
  const { imageOrderDraft, stageImageOrder, resetImageOrderDraft } = useImageOrderDraft();

  const loadData = async () => {
    try {
      const res = await fetchJSON("/api/admin/content-sections");
      setSections(res.data || []);
      setError("");
    } catch (err) {
      setSections([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleImageUpload = async (files) => {
    if (!files || files.length === 0) return;
    setError("");
    setLoading(true);
    try {
      const uploads = await Promise.all(
        Array.from(files).map(async (file) => {
          const formData = new FormData();
          formData.append("file", file);
          const res = await fetch(`${API_BASE}/api/admin/upload/content`, {
            method: "POST",
            body: formData,
            credentials: "include"
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
        image_urls: nextImages
      }));
    } catch (err) {
      setError(t("messages.error"));
    } finally {
      setLoading(false);
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
      image_urls: nextImages
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
      image_urls: nextImages
    }));
    setSelectedImageIndex((current) => selectedIndexAfterImageMove(current, index, direction));
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    const payload = {
      ...form,
      order_index: Number(form.order_index) || 0,
      image_urls: form.image_urls || []
    };
    try {
      if (editingId) {
        await fetchJSON(`/api/admin/content-sections/${editingId}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
      } else {
        await fetchJSON("/api/admin/content-sections", {
          method: "POST",
          body: JSON.stringify(payload)
        });
      }
      setForm(emptyForm);
      setEditingId(null);
      setSelectedImageIndex(0);
      resetImageOrderDraft();
      loadData();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const handleEdit = async (section) => {
    setEditingId(section.id);
    setFormOpen(true);
    setLoading(true);
    try {
      const res = await fetchJSON(`/api/admin/content-sections/${section.id}`);
      const item = res.data || section;
      setForm({
        page: item.page || "",
        key: item.key || "",
        title_en: item.title_en || "",
        title_fa: item.title_fa || "",
        title_ar: item.title_ar || "",
        subtitle_en: item.subtitle_en || "",
        subtitle_fa: item.subtitle_fa || "",
        subtitle_ar: item.subtitle_ar || "",
        description_en: item.description_en || "",
        description_fa: item.description_fa || "",
        description_ar: item.description_ar || "",
        cta_label_en: item.cta_label_en || "",
        cta_label_fa: item.cta_label_fa || "",
        cta_label_ar: item.cta_label_ar || "",
        cta_href: item.cta_href || "",
        order_index: item.order_index || 0,
        is_active: Boolean(item.is_active),
        image_urls: item.images || []
      });
      setSelectedImageIndex(0);
      resetImageOrderDraft(item.images || []);
    } catch (err) {
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await fetchJSON(`/api/admin/content-sections/${id}`, { method: "DELETE" });
      loadData();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const listItems = useMemo(() => sections, [sections]);
  const configuredFormImages = form.image_urls || [];
  const fallbackFormImages = defaultSlideImages(form.page, form.key);
  const usingDefaultFormImages = configuredFormImages.length === 0 && fallbackFormImages.length > 0;
  const effectiveFormImages = configuredFormImages.length ? configuredFormImages : fallbackFormImages;

  return (
    <div className="space-y-6">
      <section className="panel-card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-display text-xl">{t("panelContent.title")}</h2>
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
                {t("form.page")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.page}
                  onChange={(event) => setForm({ ...form, page: event.target.value })}
                  required
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.sectionKey")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.key}
                  onChange={(event) => {
                    setForm({ ...form, key: event.target.value });
                    setSelectedImageIndex(0);
                  }}
                  required
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-3">
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
            </div>

            <div className="grid gap-4 md:grid-cols-3">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.subtitle")} EN
                <textarea
                  rows={2}
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.subtitle_en}
                  onChange={(event) => setForm({ ...form, subtitle_en: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.subtitle")} FA
                <textarea
                  rows={2}
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.subtitle_fa}
                  onChange={(event) => setForm({ ...form, subtitle_fa: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.subtitle")} AR
                <textarea
                  rows={2}
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.subtitle_ar}
                  onChange={(event) => setForm({ ...form, subtitle_ar: event.target.value })}
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-3">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.description")} EN
                <textarea
                  rows={4}
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.description_en}
                  onChange={(event) => setForm({ ...form, description_en: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.description")} FA
                <textarea
                  rows={4}
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.description_fa}
                  onChange={(event) => setForm({ ...form, description_fa: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.description")} AR
                <textarea
                  rows={4}
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.description_ar}
                  onChange={(event) => setForm({ ...form, description_ar: event.target.value })}
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-3">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.ctaLabel")} EN
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.cta_label_en}
                  onChange={(event) => setForm({ ...form, cta_label_en: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.ctaLabel")} FA
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.cta_label_fa}
                  onChange={(event) => setForm({ ...form, cta_label_fa: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.ctaLabel")} AR
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.cta_label_ar}
                  onChange={(event) => setForm({ ...form, cta_label_ar: event.target.value })}
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-[2fr_1fr]">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.ctaHref")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.cta_href}
                  onChange={(event) => setForm({ ...form, cta_href: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.orderIndex")}
                <input
                  type="number"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.order_index}
                  onChange={(event) => setForm({ ...form, order_index: event.target.value })}
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-[2fr_1fr]">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("panelContent.uploadSlides")}
                <input
                  type="file"
                  accept="image/*"
                  multiple
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  onChange={(event) => handleImageUpload(event.target.files)}
                />
                <span className="mt-2 block text-[11px] font-normal normal-case tracking-normal text-primary/55">
                  {t("panelContent.uploadHint")}
                </span>
              </label>
              <label className="flex items-center gap-3 rounded-2xl border border-primary/10 bg-white p-4 text-sm">
                <input
                  type="checkbox"
                  className="h-4 w-4"
                  checked={form.is_active}
                  onChange={(event) => setForm({ ...form, is_active: event.target.checked })}
                />
                <span className="font-semibold text-primary/80">{t("form.isActive")}</span>
              </label>
            </div>

            <div className="rounded-2xl border border-primary/10 bg-sand/60 p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-wide text-primary/60">
                    {t("panelContent.mediaTitle")}
                  </p>
                  <p className="mt-1 text-xs text-primary/55">
                    {usingDefaultFormImages
                      ? t("panelContent.defaultSlidesHint")
                      : configuredFormImages.length
                        ? t("panelContent.customSlidesHint")
                        : t("panelContent.noSlides")}
                  </p>
                </div>
                <span className={`rounded-full px-3 py-1 text-[11px] font-semibold ${usingDefaultFormImages ? "bg-primary/10 text-primary/70" : configuredFormImages.length ? "bg-emerald-100 text-emerald-700" : "bg-amber-100 text-amber-700"}`}>
                  {usingDefaultFormImages
                    ? t("panelContent.defaultSlides")
                    : configuredFormImages.length
                      ? t("panelContent.customSlides")
                      : t("panelContent.noSlides")}
                  {` • ${effectiveFormImages.length}`}
                </span>
              </div>
              <div className="mt-4">
                <ReorderableImageGrid
                  images={effectiveFormImages}
                  selectedIndex={selectedImageIndex}
                  onSelect={setSelectedImageIndex}
                  onRemove={usingDefaultFormImages ? undefined : handleRemoveImage}
                  onMove={usingDefaultFormImages ? undefined : handleMoveImage}
                  previewClassName="h-40"
                  thumbnailClassName="h-14"
                  gridClassName="grid grid-cols-3 gap-3 sm:grid-cols-4 md:grid-cols-6"
                  showOrderNumbers
                  labels={{
                    empty: t("messages.empty"),
                    remove: t("actions.delete"),
                    moveLeft: t("actions.moveLeft"),
                    moveRight: t("actions.moveRight"),
                    slide: t("panelContent.slide")
                  }}
                />
              </div>
            </div>

            {error && <p className="text-sm text-red-500">{error}</p>}

            <div className="flex flex-wrap gap-3">
              <button
                type="submit"
                className="rounded-full bg-primary px-6 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-sand"
              >
                {editingId ? t("actions.update") : t("actions.create")}
              </button>
              {editingId && (
                <button
                  type="button"
                  className="rounded-full border border-primary/20 px-6 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-primary"
                  onClick={() => {
                    setEditingId(null);
                    setForm(emptyForm);
                    setSelectedImageIndex(0);
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

      <section className="panel-card max-h-[720px] overflow-y-auto">
        {loading && <p className="text-sm text-primary/70">{t("messages.loading")}</p>}
        {!loading && listItems.length === 0 && (
          <p className="text-sm text-primary/70">{t("panelContent.empty")}</p>
        )}

        <div className="space-y-3">
          {listItems.map((section) => {
            const configuredImages = section.images || [];
            const fallbackImages = defaultSlideImages(section.page, section.key);
            const usingDefaults = configuredImages.length === 0 && fallbackImages.length > 0;
            const visibleImages = configuredImages.length ? configuredImages : fallbackImages;
            return (
              <div
                key={section.id}
                className="rounded-2xl border border-primary/10 bg-white px-5 py-4"
              >
                <div className="flex flex-wrap items-center justify-between gap-4">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-semibold text-primary">
                      {section.page} / {section.key}
                    </p>
                    <p className="text-xs text-primary/60">
                      {section.title_fa || section.title_en}
                    </p>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-primary/70">
                    <span>{t("form.orderIndex")}: {section.order_index}</span>
                    <span>{section.is_active ? t("form.active") : t("form.inactive")}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
                      onClick={() => handleEdit(section)}
                    >
                      {t("actions.edit")}
                    </button>
                    <button
                      type="button"
                      className="rounded-full border border-red-200 px-4 py-2 text-xs font-semibold text-red-500"
                      onClick={() => handleDelete(section.id)}
                    >
                      {t("actions.delete")}
                    </button>
                  </div>
                </div>

                <div className="mt-4 border-t border-primary/10 pt-4">
                  <div className="mb-3 flex flex-wrap items-center justify-between gap-2 text-xs">
                    <span className="font-semibold text-primary/70">{t("panelContent.currentSlides")}</span>
                    <span className={usingDefaults ? "text-primary/55" : configuredImages.length ? "text-emerald-700" : "text-amber-700"}>
                      {usingDefaults
                        ? t("panelContent.defaultSlides")
                        : configuredImages.length
                          ? t("panelContent.customSlides")
                          : t("panelContent.noSlides")}
                      {` • ${visibleImages.length}`}
                    </span>
                  </div>
                  <ReorderableImageGrid
                    images={visibleImages}
                    showPreview={false}
                    showOrderNumbers
                    thumbnailClassName="h-20"
                    gridClassName="grid grid-cols-3 gap-3 sm:grid-cols-5 lg:grid-cols-8"
                    labels={{
                      empty: t("panelContent.noSlides"),
                      slide: t("panelContent.slide")
                    }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      </section>
    </div>
  );
}
