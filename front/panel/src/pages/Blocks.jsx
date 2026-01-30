import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { API_BASE, fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const emptyForm = {
  title_en: "",
  title_fa: "",
  title_ar: "",
  slug: "",
  stone_type: "",
  quarry: "",
  dimensions: "",
  weight_ton: "",
  status: "available",
  description: "",
  image_url: "",
  image_urls: [],
  is_featured: false
};

const statusOptions = ["available", "reserved", "sold"];

export default function Blocks() {
  const { t } = useTranslation();
  const [blocks, setBlocks] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedImageIndex, setSelectedImageIndex] = useState(0);
  const [formOpen, setFormOpen] = useState(true);

  const loadData = async () => {
    try {
      const res = await fetchJSON("/api/admin/blocks");
      setBlocks(res.data || []);
      setError("");
    } catch (err) {
      setBlocks([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const getImageCount = (block) => {
    if (typeof block.image_count === "number") return block.image_count;
    return block.image_url ? 1 : 0;
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
          const res = await fetch(`${API_BASE}/api/admin/upload/block`, {
            method: "POST",
            body: formData,
            credentials: "include"
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
      weight_ton: form.weight_ton ? Number(form.weight_ton) : 0,
      image_url: images[0] || form.image_url || "",
      image_urls: images,
      is_featured: Boolean(form.is_featured)
    };

    try {
      if (editingId) {
        await fetchJSON(`/api/admin/blocks/${editingId}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
      } else {
        await fetchJSON("/api/admin/blocks", {
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

  const handleEdit = async (block) => {
    setEditingId(block.id);
    setLoading(true);
    try {
      const res = await fetchJSON(`/api/admin/blocks/${block.id}`);
      const item = res.data || block;
      const images = item.images?.length ? item.images : item.image_url ? [item.image_url] : [];
      setForm({
        title_en: item.title_en || "",
        title_fa: item.title_fa || "",
        title_ar: item.title_ar || "",
        slug: item.slug || "",
        stone_type: item.stone_type || "",
        quarry: item.quarry || "",
        dimensions: item.dimensions || "",
        weight_ton: item.weight_ton || "",
        status: item.status || "available",
        description: item.description || "",
        image_url: item.image_url || "",
        image_urls: images,
        is_featured: Boolean(item.is_featured)
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
      await fetchJSON(`/api/admin/blocks/${id}`, { method: "DELETE" });
      loadData();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const listItems = useMemo(() => blocks, [blocks]);

  const statusLabel = (status) => {
    if (status === "available") return t("blocks.statusAvailable");
    if (status === "reserved") return t("blocks.statusReserved");
    if (status === "sold") return t("blocks.statusSold");
    return status;
  };

  return (
    <div className="space-y-6">
      <section className="panel-card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-display text-xl">{t("panelBlocks.title")}</h2>
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
                {t("form.slug")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.slug}
                  onChange={(event) => setForm({ ...form, slug: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.stoneType")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.stone_type}
                  onChange={(event) => setForm({ ...form, stone_type: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.quarry")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.quarry}
                  onChange={(event) => setForm({ ...form, quarry: event.target.value })}
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-3">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.dimensions")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.dimensions}
                  onChange={(event) => setForm({ ...form, dimensions: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.weightTon")}
                <input
                  type="number"
                  step="0.01"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.weight_ton}
                  onChange={(event) => setForm({ ...form, weight_ton: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.status")}
                <select
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.status}
                  onChange={(event) => setForm({ ...form, status: event.target.value })}
                >
                  {statusOptions.map((status) => (
                    <option key={status} value={status}>
                      {t(`blocks.status${status.charAt(0).toUpperCase()}${status.slice(1)}`)}
                    </option>
                  ))}
                </select>
              </label>
            </div>

            <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
              {t("form.description")}
              <textarea
                rows={4}
                className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                value={form.description}
                onChange={(event) => setForm({ ...form, description: event.target.value })}
              />
            </label>

            <div className="grid gap-4 md:grid-cols-[2fr_1fr]">
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.imageUrl")}
                <input
                  type="text"
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  value={form.image_url}
                  onChange={(event) => setForm({ ...form, image_url: event.target.value })}
                />
              </label>
              <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                {t("form.images")}
                <input
                  type="file"
                  accept="image/*"
                  multiple
                  className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                  onChange={(event) => handleImageUpload(event.target.files)}
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-[2fr_1fr]">
              <div className="rounded-2xl border border-primary/10 bg-sand/60 p-4">
                <p className="text-xs font-semibold uppercase tracking-wide text-primary/60">
                  {t("panelBlocks.imagesLabel")}
                </p>
                {form.image_urls.length ? (
                  <div className="mt-4">
                    <img
                      src={resolveImageUrl(form.image_urls[selectedImageIndex])}
                      alt=""
                      className="h-40 w-full rounded-xl object-cover"
                    />
                    <div className="mt-3 flex flex-wrap gap-2">
                      {form.image_urls.map((url, index) => (
                        <button
                          key={`${url}-${index}`}
                          type="button"
                          onClick={() => setSelectedImageIndex(index)}
                          className={`relative h-14 w-14 overflow-hidden rounded-lg border ${
                            selectedImageIndex === index ? "border-accent" : "border-primary/10"
                          }`}
                        >
                          <img src={resolveImageUrl(url)} alt="" className="h-full w-full object-cover" />
                          <span
                            role="button"
                            tabIndex={0}
                            onClick={(event) => {
                              event.stopPropagation();
                              handleRemoveImage(index);
                            }}
                            onKeyDown={(event) => {
                              if (event.key === "Enter") {
                                event.stopPropagation();
                                handleRemoveImage(index);
                              }
                            }}
                            className="absolute right-1 top-1 rounded-full bg-white/90 px-1 text-[10px]"
                          >
                            ✕
                          </span>
                        </button>
                      ))}
                    </div>
                  </div>
                ) : (
                  <p className="mt-3 text-sm text-primary/60">{t("messages.empty")}</p>
                )}
              </div>

              <label className="flex items-center gap-3 rounded-2xl border border-primary/10 bg-white p-4 text-sm">
                <input
                  type="checkbox"
                  className="h-4 w-4"
                  checked={form.is_featured}
                  onChange={(event) => setForm({ ...form, is_featured: event.target.checked })}
                />
                <span className="font-semibold text-primary/80">{t("form.featured")}</span>
              </label>
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
          <p className="text-sm text-primary/70">{t("panelBlocks.empty")}</p>
        )}

        <div className="space-y-3">
          {listItems.map((block) => (
            <div
              key={block.id}
              className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-primary/10 bg-white px-5 py-4"
            >
              <div className="flex min-w-0 items-center gap-4">
                <div className="relative h-16 w-16 overflow-hidden rounded-xl bg-primary/5">
                  {block.image_url ? (
                    <img
                      src={resolveImageUrl(block.image_url)}
                      alt=""
                      className="h-full w-full object-cover"
                    />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center text-xs text-primary/50">
                      {t("panelBlocks.imagesLabel")}
                    </div>
                  )}
                  {getImageCount(block) > 0 && (
                    <span className="absolute right-1 top-1 rounded-full bg-white/90 px-1.5 text-[10px] font-semibold text-primary">
                      {getImageCount(block)}
                    </span>
                  )}
                </div>
                <div className="min-w-0">
                  <p className="truncate text-sm font-semibold text-primary">{block.title_fa || block.title_en}</p>
                  <p className="text-xs text-primary/60">
                    {(block.stone_type || t("blocks.stoneType"))} · {statusLabel(block.status)}
                  </p>
                </div>
              </div>

              <div className="flex flex-wrap items-center gap-2 text-xs text-primary/70">
                {block.weight_ton ? <span>{t("blocks.weightTon")}: {block.weight_ton}</span> : null}
                {block.dimensions ? <span>{t("blocks.dimensions")}: {block.dimensions}</span> : null}
              </div>

              <div className="flex items-center gap-2">
                <button
                  type="button"
                  className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
                  onClick={() => handleEdit(block)}
                >
                  {t("actions.edit")}
                </button>
                <button
                  type="button"
                  className="rounded-full border border-red-200 px-4 py-2 text-xs font-semibold text-red-500"
                  onClick={() => handleDelete(block.id)}
                >
                  {t("actions.delete")}
                </button>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
