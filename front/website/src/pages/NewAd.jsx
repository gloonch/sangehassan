import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";
import { PRICE_UNIT_VALUES, formatPriceUnit } from "../lib/listings";
import { usePageSeo } from "../lib/seo";

let extraRowId = 0;

const createExtraRow = () => {
  extraRowId += 1;
  return { id: extraRowId, key: "", value: "" };
};

export default function NewAd() {
  const { t, lang } = useTranslation();
  const navigate = useNavigate();
  const [form, setForm] = useState({
    title: "",
    stone_type: "",
    form: "",
    tonnage: "",
    province: "",
    city: "",
    price_amount: "",
    price_unit: "per_ton",
    description: ""
  });
  const [extraRows, setExtraRows] = useState([createExtraRow()]);
  const [images, setImages] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);

  usePageSeo({
    title: `${t("ads.create")} | SangeHassan`,
    description: t("ads.subtitle"),
    path: "/ads/new",
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    robots: "noindex,nofollow,noarchive"
  });

  const update = (key, value) => setForm((prev) => ({ ...prev, [key]: value }));

  useEffect(() => {
    return () => {
      images.forEach((item) => {
        if (item?.preview) {
          URL.revokeObjectURL(item.preview);
        }
      });
    };
  }, [images]);

  const updateExtraRow = (id, key, value) => {
    setExtraRows((prev) => prev.map((row) => (row.id === id ? { ...row, [key]: value } : row)));
  };

  const addExtraRow = () => setExtraRows((prev) => [...prev, createExtraRow()]);

  const removeExtraRow = (id) => {
    setExtraRows((prev) => {
      const next = prev.filter((row) => row.id !== id);
      return next.length > 0 ? next : [createExtraRow()];
    });
  };

  const buildExtraProps = () => {
    const extra = {};
    for (const row of extraRows) {
      const key = row.key.trim();
      const value = row.value.trim();
      if (!key && !value) continue;
      if (!key || !value) {
        return { error: t("ads.form.extraInvalidRow") };
      }
      extra[key] = value;
    }
    return { extra };
  };

  const handleImageChange = (e) => {
    const selected = Array.from(e.target.files || []).map((file, index) => ({
      id: `${file.name}-${file.size}-${file.lastModified}-${index}`,
      file,
      preview: URL.createObjectURL(file)
    }));
    setImages(selected);
  };

  const uploadImages = async () => {
    if (!images.length) return [];
    setUploading(true);
    const uploaded = [];
    for (const item of images) {
      const body = new FormData();
      body.append("file", item.file);
      const res = await fetchJSON("/api/ads/upload-image", {
        method: "POST",
        body
      });
      const data = res?.data || res;
      if (data?.image_url) uploaded.push(data.image_url);
    }
    setUploading(false);
    return uploaded;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setSaving(true);
    try {
      const { extra, error: extraError } = buildExtraProps();
      if (extraError) {
        setError(extraError);
        setSaving(false);
        return;
      }
      const uploadedImages = await uploadImages();
      const payload = {
        title: form.title || null,
        stone_type: form.stone_type || null,
        form: form.form || null,
        tonnage: form.tonnage ? Number(form.tonnage) : null,
        province: form.province || null,
        city: form.city || null,
        price_amount: form.price_amount ? Number(form.price_amount) : null,
        price_unit: form.price_unit || null,
        description: form.description || null,
        extra_props: extra,
        images: uploadedImages
      };
      const res = await fetchJSON("/api/ads", {
        method: "POST",
        body: JSON.stringify(payload)
      });
      const created = res?.data || res;
      setSuccess(t("ads.created"));
      setTimeout(() => navigate(`/ads/${created.id}`), 500);
    } catch (err) {
      if (err?.status === 401) {
        sessionStorage.setItem("sh_after_login", "/ads/new");
        setError(t("ads.loginRequired"));
        navigate("/login");
      } else {
        setError(err?.message || t("messages.error"));
      }
    } finally {
      setUploading(false);
      setSaving(false);
    }
  };

  return (
    <section className="section-shell py-16">
      <div className="glass-panel rounded-3xl p-8">
        <h1 className="font-display text-3xl">{t("ads.create")}</h1>
        <p className="text-sm text-primary/70">{t("ads.subtitle")}</p>
        <div className="mt-4 rounded-2xl border border-primary/15 bg-primary/5 px-4 py-3 text-sm text-primary/80">
          {t("ads.privacyNote")}
        </div>

        <form className="mt-6 grid gap-4 md:grid-cols-2" onSubmit={handleSubmit}>
          <Field label={t("ads.form.title")}>
            <input value={form.title} onChange={(e) => update("title", e.target.value)} className={inputClass} />
          </Field>
          <Field label={t("ads.form.stoneType")}>
            <input value={form.stone_type} onChange={(e) => update("stone_type", e.target.value)} className={inputClass} />
          </Field>
          <Field label={t("ads.form.form")}>
            <select value={form.form} onChange={(e) => update("form", e.target.value)} className={inputClass}>
              <option value="">—</option>
              <option value="block">Block</option>
              <option value="finished">Finished</option>
            </select>
          </Field>
          <Field label={t("ads.form.tonnage")}>
            <input
              value={form.tonnage}
              onChange={(e) => update("tonnage", e.target.value)}
              className={inputClass}
              type="number"
              step="0.01"
            />
          </Field>
          <Field label={t("ads.form.province")}>
            <input value={form.province} onChange={(e) => update("province", e.target.value)} className={inputClass} />
          </Field>
          <Field label={t("ads.form.city")}>
            <input value={form.city} onChange={(e) => update("city", e.target.value)} className={inputClass} />
          </Field>
          <Field label={t("ads.form.price")}>
            <input
              value={form.price_amount}
              onChange={(e) => update("price_amount", e.target.value)}
              className={inputClass}
              type="number"
              step="0.01"
            />
          </Field>
          <Field label={t("ads.form.priceUnit")}>
            <select value={form.price_unit} onChange={(e) => update("price_unit", e.target.value)} className={inputClass}>
              {PRICE_UNIT_VALUES.map((unitValue) => (
                <option key={unitValue} value={unitValue}>
                  {formatPriceUnit(unitValue, t)}
                </option>
              ))}
            </select>
          </Field>
          <Field label={t("ads.form.description")} full>
            <textarea
              value={form.description}
              onChange={(e) => update("description", e.target.value)}
              className={inputClass}
              rows={3}
            />
          </Field>
          <Field label={t("ads.form.extra")} full>
            <div className="space-y-2">
              {extraRows.map((row) => (
                <div key={row.id} className="grid gap-2 md:grid-cols-[1fr_1fr_auto] md:items-center">
                  <input
                    value={row.key}
                    onChange={(e) => updateExtraRow(row.id, "key", e.target.value)}
                    placeholder={t("ads.form.extraKeyPlaceholder")}
                    className={inputClass}
                  />
                  <input
                    value={row.value}
                    onChange={(e) => updateExtraRow(row.id, "value", e.target.value)}
                    placeholder={t("ads.form.extraValuePlaceholder")}
                    className={inputClass}
                  />
                  <button
                    type="button"
                    onClick={() => removeExtraRow(row.id)}
                    className="h-fit rounded-full border border-primary/20 px-3 py-2 text-[11px] font-semibold text-primary hover:border-primary/40"
                  >
                    {t("ads.form.removeRow")}
                  </button>
                </div>
              ))}
            </div>
            <button
              type="button"
              onClick={addExtraRow}
              className="mt-1 inline-flex h-fit w-fit rounded-full bg-primary/10 px-3 py-2 text-[11px] font-semibold text-primary hover:bg-primary/20"
            >
              {t("ads.form.addRow")}
            </button>
            <p className="text-[11px] font-medium normal-case tracking-normal text-primary/60">
              {t("ads.form.extraHint")}
            </p>
          </Field>
          <Field label={t("ads.form.images")} full>
            <input
              type="file"
              multiple
              accept=".png,.jpg,.jpeg,.webp"
              onChange={handleImageChange}
              className={`${inputClass} file:mr-3 file:rounded-full file:border-0 file:bg-primary file:px-3 file:py-1.5 file:text-xs file:font-semibold file:text-sand`}
            />
            <p className="text-[11px] font-medium text-primary/60">{t("ads.form.imagesHint")}</p>
            {images.length > 0 && (
              <div className="mt-2 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
                {images.map((item) => (
                  <figure key={item.id} className="rounded-xl border border-primary/10 bg-white/80 p-2">
                    <img
                      src={item.preview}
                      alt={item.file.name}
                      className="h-20 w-full rounded-lg object-cover"
                    />
                    <figcaption className="mt-1 truncate text-[11px] normal-case tracking-normal text-primary/75">
                      {item.file.name}
                    </figcaption>
                  </figure>
                ))}
              </div>
            )}
          </Field>

          <div className="md:col-span-2 flex items-center gap-3">
            <button
              type="submit"
              disabled={saving || uploading}
              className="rounded-full bg-primary px-5 py-2 text-sm font-semibold text-sand hover:bg-primary/90 disabled:opacity-60"
            >
              {saving || uploading ? t("messages.loading") : t("ads.create")}
            </button>
            {success && <span className="text-xs font-semibold text-green-700">{success}</span>}
            {error && <span className="text-xs font-semibold text-red-600">{error}</span>}
          </div>
        </form>
      </div>
    </section>
  );
}

function Field({ label, children, full }) {
  return (
    <label className={`flex flex-col gap-2 text-xs font-semibold uppercase tracking-wide text-primary/70 ${full ? "md:col-span-2" : ""}`}>
      {label}
      {children}
    </label>
  );
}

const inputClass =
  "mt-1 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm focus:border-primary focus:outline-none";
