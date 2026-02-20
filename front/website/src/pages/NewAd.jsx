import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

export default function NewAd() {
  const { t } = useTranslation();
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
    description: "",
    extra_props: "{}",
    images: ""
  });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [saving, setSaving] = useState(false);

  const update = (key, value) => setForm((prev) => ({ ...prev, [key]: value }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setSaving(true);
    try {
      let extra = {};
      try {
        extra = form.extra_props ? JSON.parse(form.extra_props) : {};
      } catch {
        setError("Invalid JSON in extra properties");
        setSaving(false);
        return;
      }
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
        images: form.images
          ? form.images
              .split(",")
              .map((s) => s.trim())
              .filter(Boolean)
          : []
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
              <option value="per_ton">per_ton</option>
              <option value="total">total</option>
              <option value="negotiable">negotiable</option>
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
            <textarea
              value={form.extra_props}
              onChange={(e) => update("extra_props", e.target.value)}
              className={`${inputClass} font-mono text-xs`}
              rows={3}
            />
          </Field>
          <Field label="Images (comma separated URLs)" full>
            <input value={form.images} onChange={(e) => update("images", e.target.value)} className={inputClass} />
          </Field>

          <div className="md:col-span-2 flex items-center gap-3">
            <button
              type="submit"
              disabled={saving}
              className="rounded-full bg-primary px-5 py-2 text-sm font-semibold text-sand hover:bg-primary/90 disabled:opacity-60"
            >
              {saving ? t("messages.loading") : t("ads.create")}
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
