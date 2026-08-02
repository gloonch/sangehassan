import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";
import { resolveImageUrl } from "../lib/assets";
import { PRICE_UNIT_VALUES, formatPriceUnit, getListingQuantityFieldLabel, getLocalizedProductTitle } from "../lib/listings";
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
    product_id: "",
    form: "",
    tonnage: "",
    province: "",
    city: "",
    price_amount: "",
    price_unit: "per_ton",
    description: ""
  });
  const [extraRows, setExtraRows] = useState([createExtraRow()]);
  const [productQuery, setProductQuery] = useState("");
  const [productOptions, setProductOptions] = useState([]);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [productLoading, setProductLoading] = useState(false);
  const [productToast, setProductToast] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [saving, setSaving] = useState(false);

  usePageSeo({
    title: `${t("ads.create")} | SangeHassan`,
    description: t("ads.subtitle"),
    path: "/ads/new",
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    robots: "noindex,nofollow,noarchive"
  });

  const update = (key, value) => setForm((prev) => ({ ...prev, [key]: value }));
  const updateFormType = (value) => setForm((prev) => ({
    ...prev,
    form: value,
    price_unit: ["per_ton", "per_meter"].includes(prev.price_unit)
      ? value === "block" ? "per_ton" : "per_meter"
      : prev.price_unit
  }));

  useEffect(() => {
    const query = productQuery.trim();
    if (query.length < 2) {
      setProductOptions([]);
      setProductLoading(false);
      return undefined;
    }

    let active = true;
    setProductLoading(true);
    const timer = window.setTimeout(async () => {
      try {
        const res = await fetchJSON(`/api/products?limit=12&q=${encodeURIComponent(query)}`);
        const data = res?.data || res;
        if (active) setProductOptions(Array.isArray(data) ? data : []);
      } catch (_) {
        if (active) setProductOptions([]);
      } finally {
        if (active) setProductLoading(false);
      }
    }, 220);

    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [productQuery]);

  useEffect(() => {
    if (!productToast) return undefined;
    const timer = window.setTimeout(() => setProductToast(""), 10000);
    return () => window.clearTimeout(timer);
  }, [productToast]);

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
      if (key.toLowerCase() === "recommended_use") {
        return { error: t("ads.form.recommendedUseUnsupported") };
      }
      extra[key] = value;
    }
    return { extra };
  };

  const selectProduct = (product) => {
    setSelectedProduct(product);
    update("product_id", product?.id ? String(product.id) : "");
    setProductQuery(getLocalizedProductTitle(product, lang));
    setProductOptions([]);
    setProductToast("");
  };

  const requestNewProduct = async () => {
    const query = productQuery.trim();
    if (!query) {
      setProductToast(t("ads.form.productRequestNeedsQuery"));
      return;
    }
    try {
      await fetchJSON("/api/ads/product-requests", {
        method: "POST",
        body: JSON.stringify({ query })
      });
    } catch (_) {
      // The user-facing next step is still the same: call support.
    }
    setProductToast(t("ads.form.productRequestToast").replace("{phone}", t("footer.phoneValue")));
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
      const productID = Number(form.product_id || selectedProduct?.id || 0);
      if (!productID) {
        setError(t("ads.form.productRequired"));
        setSaving(false);
        return;
      }
      const payload = {
        title: form.title || null,
        product_id: productID,
        form: form.form || null,
        tonnage: form.tonnage ? Number(form.tonnage) : null,
        province: form.province || null,
        city: form.city || null,
        price_amount: form.price_amount ? Number(form.price_amount) : null,
        price_unit: form.price_unit || null,
        description: form.description || null,
        extra_props: extra
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
          <Field label={t("ads.form.product")} full>
            <div className="relative normal-case tracking-normal">
              <input
                value={productQuery}
                onChange={(e) => {
                  setProductQuery(e.target.value);
                  setSelectedProduct(null);
                  update("product_id", "");
                }}
                placeholder={t("ads.form.productSearchPlaceholder")}
                className={inputClass}
                autoComplete="off"
              />
              {productLoading && (
                <span className="absolute right-3 top-5 h-3 w-3 animate-spin rounded-full border border-primary/40 border-t-transparent" />
              )}
              {productQuery.trim().length >= 2 && productOptions.length > 0 && !selectedProduct && (
                <div className="absolute z-20 mt-2 max-h-72 w-full overflow-y-auto rounded-xl border border-primary/15 bg-white p-2 shadow-xl">
                  {productOptions.map((product) => {
                    const title = getLocalizedProductTitle(product, lang);
                    return (
                      <button
                        type="button"
                        key={product.id}
                        onClick={() => selectProduct(product)}
                        className="flex w-full items-center gap-3 rounded-lg px-2 py-2 text-start hover:bg-primary/5"
                      >
                        <span className="h-10 w-10 shrink-0 overflow-hidden rounded-lg bg-primary/10">
                          {product.image_url ? (
                            <img
                              src={resolveImageUrl(product.image_url)}
                              alt={title}
                              className="h-full w-full object-cover"
                              loading="lazy"
                            />
                          ) : null}
                        </span>
                        <span className="min-w-0">
                          <span className="block truncate text-sm font-semibold text-primary">{title}</span>
                          <span className="block truncate text-xs text-primary/55">{product.slug}</span>
                        </span>
                      </button>
                    );
                  })}
                </div>
              )}
              {productQuery.trim().length >= 2 && !productLoading && productOptions.length === 0 && !selectedProduct && (
                <div className="mt-2 rounded-xl border border-primary/10 bg-primary/5 px-3 py-3 text-xs text-primary/70">
                  <p>{t("ads.form.productNoResults")}</p>
                  <button
                    type="button"
                    onClick={requestNewProduct}
                    className="mt-2 rounded-full bg-primary px-3 py-1.5 text-[11px] font-semibold text-sand hover:bg-primary/90"
                  >
                    {t("ads.form.requestNewProduct")}
                  </button>
                </div>
              )}
              {selectedProduct && (
                <div className="mt-3 flex items-center gap-3 rounded-xl border border-primary/10 bg-white/80 p-3">
                  <span className="h-16 w-16 shrink-0 overflow-hidden rounded-lg bg-primary/10">
                    {selectedProduct.image_url ? (
                      <img
                        src={resolveImageUrl(selectedProduct.image_url)}
                        alt={getLocalizedProductTitle(selectedProduct, lang)}
                        className="h-full w-full object-cover"
                        loading="lazy"
                      />
                    ) : null}
                  </span>
                  <span className="min-w-0">
                    <span className="block text-sm font-semibold text-primary">
                      {getLocalizedProductTitle(selectedProduct, lang)}
                    </span>
                    <span className="block text-xs text-primary/55">{t("ads.form.productCoverHint")}</span>
                  </span>
                </div>
              )}
              {productToast && (
                <div className="mt-3 rounded-xl border border-accent/20 bg-accent/10 px-3 py-2 text-xs font-semibold text-primary">
                  {productToast}
                </div>
              )}
            </div>
          </Field>
          <Field label={t("ads.form.form")}>
            <select value={form.form} onChange={(e) => updateFormType(e.target.value)} className={inputClass}>
              <option value="">—</option>
              <option value="block">{t("ads.formOptions.block")}</option>
              <option value="finished">{t("ads.formOptions.finished")}</option>
            </select>
          </Field>
          <Field label={getListingQuantityFieldLabel(form.form, t)}>
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
