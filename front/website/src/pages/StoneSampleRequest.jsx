import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { Check, Info, PackageCheck, Plus, Search, Trash2, X } from "lucide-react";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { useTranslation } from "../lib/i18n";
import { usePageSeo } from "../lib/seo";

const SAMPLES_PER_BOX = 4;
const MAX_BOXES = 4;
const FALLBACK_PRICE_PER_BOX = 4000000;

const getLocalizedTitle = (item, lang) => {
  if (!item) return "";
  if (lang === "fa") return item.title_fa || item.title_en || item.title_ar || "";
  if (lang === "ar") return item.title_ar || item.title_en || item.title_fa || "";
  return item.title_en || item.title_fa || item.title_ar || "";
};

const getRequestItemTitle = (item, lang) => {
  if (!item) return "";
  if (lang === "fa") return item.product_title_fa || item.product_title_en || item.product_title_ar || "";
  if (lang === "ar") return item.product_title_ar || item.product_title_en || item.product_title_fa || "";
  return item.product_title_en || item.product_title_fa || item.product_title_ar || "";
};

const formatToman = (value, lang) => {
  const locale = lang === "fa" ? "fa-IR" : lang === "ar" ? "ar-SA" : "en-US";
  return new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(Number(value) || 0);
};

const chunkSelected = (selected) => {
  const boxes = [];
  for (let index = 0; index < selected.length; index += SAMPLES_PER_BOX) {
    boxes.push(selected.slice(index, index + SAMPLES_PER_BOX));
  }
  return boxes;
};

export default function StoneSampleRequest() {
  const { t, lang } = useTranslation();
  const [selectedStones, setSelectedStones] = useState([]);
  const [categories, setCategories] = useState([]);
  const [categoryLoading, setCategoryLoading] = useState(true);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [replaceIndex, setReplaceIndex] = useState(null);
  const [activeCategoryId, setActiveCategoryId] = useState("");
  const [search, setSearch] = useState("");
  const [products, setProducts] = useState([]);
  const [productsTotal, setProductsTotal] = useState(0);
  const [productsOffset, setProductsOffset] = useState(0);
  const [productsLoading, setProductsLoading] = useState(false);
  const [options, setOptions] = useState(null);
  const [optionsLoading, setOptionsLoading] = useState(true);
  const [addressMode, setAddressMode] = useState("new");
  const [addressText, setAddressText] = useState("");
  const [phoneMode, setPhoneMode] = useState("new");
  const [phoneText, setPhoneText] = useState("");
  const [shippingMethod, setShippingMethod] = useState("OPERATOR_COORDINATION");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [successRequest, setSuccessRequest] = useState(null);
  const [infoOpen, setInfoOpen] = useState(false);

  usePageSeo({
    title: `${t("sampleRequest.title")} | SangeHassan`,
    description: t("sampleRequest.subtitle"),
    path: "/stone-sample-request",
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    robots: "noindex,nofollow,noarchive"
  });

  useEffect(() => {
    let active = true;
    const load = async () => {
      try {
        const [categoryRes, optionsRes] = await Promise.all([
          fetchJSON("/api/sample-categories"),
          fetchJSON("/api/v1/sample-request-options")
        ]);
        if (!active) return;
        const loadedCategories = categoryRes?.data || [];
        const loadedOptions = optionsRes?.data || null;
        setCategories(Array.isArray(loadedCategories) ? loadedCategories : []);
        setOptions(loadedOptions);
        const addresses = Array.isArray(loadedOptions?.addresses) ? loadedOptions.addresses : [];
        const phones = Array.isArray(loadedOptions?.phones) ? loadedOptions.phones : [];
        if (addresses[0]) {
          setAddressMode(`id:${addresses[0].id}`);
          setAddressText(addresses[0].address_text || "");
        }
        if (phones[0]) {
          const value = phones[0].id ? `id:${phones[0].id}` : `profile:${phones[0].phone}`;
          setPhoneMode(value);
          setPhoneText(phones[0].phone || "");
        }
        if (loadedOptions?.shipping_methods?.[0]?.value) {
          setShippingMethod(loadedOptions.shipping_methods[0].value);
        }
      } catch (err) {
        if (active) setError(err?.message || t("messages.error"));
      } finally {
        if (active) {
          setCategoryLoading(false);
          setOptionsLoading(false);
        }
      }
    };
    load();
    return () => {
      active = false;
    };
  }, [t]);

  useEffect(() => {
    if (!pickerOpen || !activeCategoryId) return undefined;
    let active = true;
    const load = async () => {
      setProductsLoading(true);
      try {
        const params = new URLSearchParams({
          category_id: activeCategoryId,
          limit: "9",
          offset: String(productsOffset)
        });
        if (search.trim()) params.set("q", search.trim());
        const res = await fetchJSON(`/api/sample-products?${params.toString()}`);
        const data = res?.data || {};
        if (!active) return;
        setProducts(Array.isArray(data.items) ? data.items : []);
        setProductsTotal(Number(data.total) || 0);
      } catch (err) {
        if (active) {
          setProducts([]);
          setProductsTotal(0);
          setError(err?.message || t("messages.error"));
        }
      } finally {
        if (active) setProductsLoading(false);
      }
    };
    load();
    return () => {
      active = false;
    };
  }, [activeCategoryId, pickerOpen, productsOffset, search, t]);

  const pricePerBox = options?.price_per_box_toman || FALLBACK_PRICE_PER_BOX;
  const selectedIds = useMemo(() => new Set(selectedStones.map((stone) => Number(stone.id))), [selectedStones]);
  const selectedBoxes = useMemo(() => chunkSelected(selectedStones), [selectedStones]);
  const completedBoxCount = selectedStones.length > 0 && selectedStones.length % SAMPLES_PER_BOX === 0
    ? selectedStones.length / SAMPLES_PER_BOX
    : Math.floor(selectedStones.length / SAMPLES_PER_BOX);
  const totalPrice = completedBoxCount * pricePerBox;
  const canSubmit =
    selectedStones.length >= SAMPLES_PER_BOX &&
    selectedStones.length <= MAX_BOXES * SAMPLES_PER_BOX &&
    selectedStones.length % SAMPLES_PER_BOX === 0 &&
    addressText.trim() &&
    phoneText.trim() &&
    shippingMethod;
  const maxReached = selectedStones.length >= MAX_BOXES * SAMPLES_PER_BOX;

  const visibleBoxCount = useMemo(() => {
    const filled = Math.max(1, Math.ceil(selectedStones.length / SAMPLES_PER_BOX));
    if (selectedStones.length > 0 && selectedStones.length % SAMPLES_PER_BOX === 0 && !maxReached) {
      return Math.min(MAX_BOXES, filled + 1);
    }
    return Math.min(MAX_BOXES, filled);
  }, [maxReached, selectedStones.length]);

  const openPicker = (index = null) => {
    if (index === null && maxReached) return;
    setReplaceIndex(index);
    setPickerOpen(true);
    setProductsOffset(0);
  };

  const closePicker = () => {
    setPickerOpen(false);
    setReplaceIndex(null);
    setSearch("");
    setProductsOffset(0);
  };

  const selectProduct = (product) => {
    const productID = Number(product.id);
    const replacingSame = replaceIndex !== null && Number(selectedStones[replaceIndex]?.id) === productID;
    if (selectedIds.has(productID) && !replacingSame) return;

    setSelectedStones((current) => {
      if (replaceIndex !== null) {
        return current.map((item, index) => (index === replaceIndex ? product : item));
      }
      if (current.length >= MAX_BOXES * SAMPLES_PER_BOX) return current;
      return [...current, product];
    });
    setSuccessRequest(null);
    closePicker();
  };

  const removeStone = (index) => {
    setSelectedStones((current) => current.filter((_, itemIndex) => itemIndex !== index));
    setSuccessRequest(null);
  };

  const handleAddressSelect = (value) => {
    setAddressMode(value);
    if (value === "new") {
      setAddressText("");
      return;
    }
    const id = Number(value.replace("id:", ""));
    const address = options?.addresses?.find((item) => Number(item.id) === id);
    setAddressText(address?.address_text || "");
  };

  const handlePhoneSelect = (value) => {
    setPhoneMode(value);
    if (value === "new") {
      setPhoneText("");
      return;
    }
    if (value.startsWith("profile:")) {
      setPhoneText(value.replace("profile:", ""));
      return;
    }
    const id = Number(value.replace("id:", ""));
    const phone = options?.phones?.find((item) => Number(item.id) === id);
    setPhoneText(phone?.phone || "");
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (!canSubmit) {
      setError(t("sampleRequest.validation"));
      return;
    }
    setSubmitting(true);
    setError("");
    try {
      const boxes = selectedBoxes.map((box) => box.map((stone) => Number(stone.id)));
      const selectedAddressID = addressMode.startsWith("id:") ? Number(addressMode.replace("id:", "")) : null;
      const selectedPhoneID = phoneMode.startsWith("id:") ? Number(phoneMode.replace("id:", "")) : null;
      const savedAddress = options?.addresses?.find((item) => Number(item.id) === selectedAddressID);
      const savedPhone = options?.phones?.find((item) => Number(item.id) === selectedPhoneID);
      const addressID = savedAddress?.address_text === addressText ? selectedAddressID : null;
      const phoneID = savedPhone?.phone === phoneText ? selectedPhoneID : null;
      const res = await fetchJSON("/api/v1/sample-requests", {
        method: "POST",
        body: JSON.stringify({
          boxes,
          address_id: addressID,
          address_text: addressText,
          phone_id: phoneID,
          phone: phoneText,
          shipping_method: shippingMethod
        })
      });
      setSuccessRequest(res?.data || null);
    } catch (err) {
      setError(err?.message || t("messages.error"));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section className="section-shell pb-16 pt-10">
      <h1 className="sr-only">{t("sampleRequest.title")}</h1>

      <div className="mx-auto max-w-5xl">
        <div className="flex justify-end">
          <button
            type="button"
            onClick={() => setInfoOpen((open) => !open)}
            aria-expanded={infoOpen}
            aria-controls="sample-request-info"
            aria-label={t("sampleRequest.infoButtonLabel")}
            className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-primary/15 bg-white/75 text-primary shadow-sm transition hover:border-primary/30 hover:bg-white"
          >
            <Info className="h-4 w-4" />
          </button>
        </div>
        {infoOpen ? <SampleRequestInfo pricePerBox={pricePerBox} lang={lang} t={t} /> : null}
      </div>

      <form onSubmit={handleSubmit} className="mt-4 grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="space-y-6">
          <div className="glass-panel rounded-3xl p-5 md:p-6">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.22em] text-primary/55">{t("sampleRequest.builderKicker")}</p>
                <h2 className="mt-1 font-display text-2xl">{t("sampleRequest.builderTitle")}</h2>
              </div>
              <span className="rounded-full border border-primary/15 bg-white/60 px-3 py-1 text-xs font-semibold text-primary/70">
                {selectedStones.length}/{MAX_BOXES * SAMPLES_PER_BOX}
              </span>
            </div>

            <div className="mt-6 space-y-4">
              {Array.from({ length: visibleBoxCount }).map((_, boxIndex) => (
                <SampleBox
                  key={boxIndex}
                  boxIndex={boxIndex}
                  selectedStones={selectedStones}
                  activeIndex={selectedStones.length}
                  lang={lang}
                  onPick={() => openPicker(null)}
                  onReplace={openPicker}
                  onRemove={removeStone}
                  t={t}
                />
              ))}
            </div>
            {maxReached ? (
              <p className="mt-4 rounded-2xl border border-primary/10 bg-primary/5 px-4 py-3 text-xs font-semibold text-primary/70">
                {t("sampleRequest.maxReached")}
              </p>
            ) : null}
          </div>

          <div className="glass-panel rounded-3xl p-5 md:p-6">
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-primary/55">{t("sampleRequest.deliveryKicker")}</p>
            <h2 className="mt-1 font-display text-2xl">{t("sampleRequest.deliveryTitle")}</h2>
            {optionsLoading ? (
              <p className="mt-4 text-sm text-primary/60">{t("messages.loading")}</p>
            ) : (
              <div className="mt-5 grid gap-4 md:grid-cols-2">
                <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70 md:col-span-2">
                  {t("sampleRequest.address")}
                  <select
                    value={addressMode}
                    onChange={(event) => handleAddressSelect(event.target.value)}
                    className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm normal-case tracking-normal text-primary"
                  >
                    {options?.addresses?.map((address) => (
                      <option key={address.id} value={`id:${address.id}`}>
                        {address.label || address.address_text}
                      </option>
                    ))}
                    <option value="new">{t("sampleRequest.addNewAddress")}</option>
                  </select>
                  <textarea
                    rows="3"
                    value={addressText}
                    onChange={(event) => setAddressText(event.target.value)}
                    placeholder={t("sampleRequest.addressPlaceholder")}
                    className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm normal-case leading-7 tracking-normal text-primary"
                  />
                </label>

                <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("sampleRequest.phone")}
                  <select
                    value={phoneMode}
                    onChange={(event) => handlePhoneSelect(event.target.value)}
                    className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm normal-case tracking-normal text-primary"
                  >
                    {options?.phones?.map((phone, index) => {
                      const value = phone.id ? `id:${phone.id}` : `profile:${phone.phone}`;
                      return (
                        <option key={`${value}-${index}`} value={value}>
                          {phone.label || phone.phone}
                        </option>
                      );
                    })}
                    <option value="new">{t("sampleRequest.addNewPhone")}</option>
                  </select>
                  <input
                    type="tel"
                    value={phoneText}
                    onChange={(event) => setPhoneText(event.target.value)}
                    placeholder={t("sampleRequest.phonePlaceholder")}
                    className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm normal-case tracking-normal text-primary"
                  />
                </label>

                <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("sampleRequest.shippingMethod")}
                  <select
                    value={shippingMethod}
                    onChange={(event) => setShippingMethod(event.target.value)}
                    className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm normal-case tracking-normal text-primary"
                  >
                    {options?.shipping_methods?.map((method) => (
                      <option key={method.value} value={method.value}>
                        {t(`sampleRequest.shipping.${method.value}`)}
                      </option>
                    ))}
                  </select>
                </label>
              </div>
            )}
          </div>
        </div>

        <aside className="glass-panel h-max rounded-3xl p-5 md:p-6">
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-primary/55">{t("sampleRequest.summaryKicker")}</p>
          <h2 className="mt-1 font-display text-2xl">{t("sampleRequest.summaryTitle")}</h2>
          <div className="mt-5 space-y-3 text-sm text-primary/75">
            <SummaryRow label={t("sampleRequest.summary.boxes")} value={completedBoxCount} />
            <SummaryRow label={t("sampleRequest.summary.stones")} value={selectedStones.length} />
            <SummaryRow label={t("sampleRequest.summary.pricePerBox")} value={`${formatToman(pricePerBox, lang)} ${t("sampleRequest.toman")}`} />
            <SummaryRow label={t("sampleRequest.summary.shipping")} value={t("sampleRequest.shippingPending")} />
          </div>

          {selectedBoxes.length > 0 ? (
            <div className="mt-5 space-y-3">
              {selectedBoxes.map((box, index) => (
                <div key={index} className="rounded-2xl border border-primary/10 bg-white/55 p-3">
                  <p className="text-xs font-semibold text-primary/65">{t("sampleRequest.box")} {index + 1}</p>
                  <ul className="mt-2 space-y-1 text-xs text-primary/75">
                    {box.map((stone) => (
                      <li key={stone.id} className="truncate">{getLocalizedTitle(stone, lang)}</li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          ) : null}

          <div className="mt-5 border-t border-primary/10 pt-4">
            <SummaryRow label={t("sampleRequest.summary.total")} value={`${formatToman(totalPrice, lang)} ${t("sampleRequest.toman")}`} strong />
          </div>

          {error ? <p className="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-600">{error}</p> : null}
          {successRequest ? (
            <div className="mt-4 rounded-2xl border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
              <div className="flex items-start gap-2">
                <Check className="mt-0.5 h-4 w-4" />
                <div>
                  <p className="font-semibold">{t("sampleRequest.successTitle")}</p>
                  <p className="mt-1 text-xs leading-6">{t("sampleRequest.successMessage")}</p>
                  <Link to="/profile" className="mt-3 inline-flex rounded-full bg-primary px-4 py-2 text-xs font-semibold text-sand">
                    {t("sampleRequest.viewInProfile")}
                  </Link>
                </div>
              </div>
            </div>
          ) : null}

          <button
            type="submit"
            disabled={!canSubmit || submitting}
            className="mt-5 w-full rounded-full bg-primary px-5 py-3 text-sm font-semibold text-sand transition hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {submitting ? t("messages.loading") : t("sampleRequest.submit")}
          </button>
          {!canSubmit ? <p className="mt-3 text-xs leading-6 text-primary/55">{t("sampleRequest.validation")}</p> : null}
        </aside>
      </form>

      {pickerOpen ? (
        <ProductPickerModal
          categories={categories}
          categoryLoading={categoryLoading}
          activeCategoryId={activeCategoryId}
          setActiveCategoryId={(value) => {
            setActiveCategoryId(value);
            setProductsOffset(0);
          }}
          search={search}
          setSearch={(value) => {
            setSearch(value);
            setProductsOffset(0);
          }}
          products={products}
          productsLoading={productsLoading}
          productsTotal={productsTotal}
          productsOffset={productsOffset}
          setProductsOffset={setProductsOffset}
          selectedStones={selectedStones}
          replaceIndex={replaceIndex}
          lang={lang}
          t={t}
          onSelect={selectProduct}
          onClose={closePicker}
        />
      ) : null}
    </section>
  );
}

function SampleRequestInfo({ pricePerBox, lang, t }) {
  const statuses = ["PENDING", "CONFIRMED", "SHIPPED", "DELIVERED", "REJECTED", "CANCELED"];
  return (
    <div id="sample-request-info" className="mt-3 rounded-3xl border border-primary/10 bg-white/75 p-5 shadow-sm backdrop-blur md:p-6">
      <div className="flex items-start gap-3">
        <span className="mt-1 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-sand">
          <Info className="h-4 w-4" />
        </span>
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-primary/55">{t("sampleRequest.infoPanel.kicker")}</p>
          <h2 className="mt-1 font-display text-xl text-primary md:text-2xl">{t("sampleRequest.infoPanel.title")}</h2>
          <p className="mt-2 max-w-3xl text-sm leading-7 text-primary/70">{t("sampleRequest.infoPanel.body")}</p>
        </div>
      </div>

      <dl className="mt-5 grid gap-3 text-xs font-semibold text-primary/70 sm:grid-cols-4">
        <div className="rounded-2xl border border-primary/10 bg-sand/35 px-4 py-3">
          <dt className="uppercase tracking-[0.18em] text-primary/45">{t("sampleRequest.info.samples")}</dt>
          <dd className="mt-1 text-primary/80">{t("sampleRequest.info.samplesValue")}</dd>
        </div>
        <div className="rounded-2xl border border-primary/10 bg-sand/35 px-4 py-3">
          <dt className="uppercase tracking-[0.18em] text-primary/45">{t("sampleRequest.info.limit")}</dt>
          <dd className="mt-1 text-primary/80">{t("sampleRequest.info.limitValue")}</dd>
        </div>
        <div className="rounded-2xl border border-primary/10 bg-sand/35 px-4 py-3">
          <dt className="uppercase tracking-[0.18em] text-primary/45">{t("sampleRequest.info.price")}</dt>
          <dd className="mt-1 text-primary/80">{`${formatToman(pricePerBox, lang)} ${t("sampleRequest.toman")}`}</dd>
        </div>
        <div className="rounded-2xl border border-primary/10 bg-sand/35 px-4 py-3">
          <dt className="uppercase tracking-[0.18em] text-primary/45">{t("sampleRequest.info.status")}</dt>
          <dd className="mt-1 text-primary/80">{t("sampleRequest.info.statusValue")}</dd>
        </div>
      </dl>

      <div className="mt-5 rounded-2xl border border-primary/10 bg-primary/5 px-4 py-4">
        <p className="text-sm font-semibold text-primary">{t("sampleRequest.infoPanel.statusTitle")}</p>
        <div className="mt-3 grid gap-3 md:grid-cols-2">
          {statuses.map((status) => (
            <div key={status} className="text-xs leading-6 text-primary/70">
              <span className="font-semibold text-primary">{t(`sampleRequest.status.${status}`)}</span>
              <span className="mx-2 text-primary/30">/</span>
              <span>{t(`sampleRequest.infoPanel.statusDescriptions.${status}`)}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function SampleBox({ boxIndex, selectedStones, activeIndex, lang, onPick, onReplace, onRemove, t }) {
  return (
    <div className="rounded-3xl border border-primary/10 bg-white/50 p-3">
      <div className="mb-3 flex items-center justify-between px-1">
        <p className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.18em] text-primary/55">
          <PackageCheck className="h-4 w-4" />
          {t("sampleRequest.box")} {boxIndex + 1}
        </p>
      </div>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {Array.from({ length: SAMPLES_PER_BOX }).map((_, slotIndex) => {
          const globalIndex = boxIndex * SAMPLES_PER_BOX + slotIndex;
          const stone = selectedStones[globalIndex];
          const isActive = !stone && globalIndex === activeIndex && selectedStones.length < MAX_BOXES * SAMPLES_PER_BOX;
          if (stone) {
            return (
              <div key={globalIndex} className="group relative aspect-square overflow-hidden rounded-2xl border border-primary/10 bg-primary/5">
                {stone.image_url ? (
                  <img src={resolveImageUrl(stone.image_url)} alt={getLocalizedTitle(stone, lang)} className="h-full w-full object-cover" />
                ) : (
                  <div className="flex h-full w-full items-center justify-center text-xs text-primary/50">{t("productDetail.noImages")}</div>
                )}
                <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/70 via-black/30 to-transparent p-3">
                  <p className="line-clamp-2 text-xs font-semibold leading-5 text-white">{getLocalizedTitle(stone, lang)}</p>
                </div>
                <div className="absolute right-2 top-2 flex gap-1 opacity-100 sm:opacity-0 sm:transition sm:group-hover:opacity-100">
                  <button
                    type="button"
                    onClick={() => onReplace(globalIndex)}
                    className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-white/90 text-primary shadow"
                    aria-label={t("sampleRequest.replace")}
                  >
                    <Search className="h-4 w-4" />
                  </button>
                  <button
                    type="button"
                    onClick={() => onRemove(globalIndex)}
                    className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-white/90 text-red-600 shadow"
                    aria-label={t("sampleRequest.remove")}
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            );
          }
          return (
            <button
              key={globalIndex}
              type="button"
              onClick={isActive ? onPick : undefined}
              disabled={!isActive}
              className={`flex aspect-square flex-col items-center justify-center rounded-2xl border border-dashed p-3 text-center transition ${
                isActive
                  ? "border-primary/35 bg-primary/5 text-primary hover:border-primary/60 hover:bg-primary/10"
                  : "border-primary/10 bg-white/35 text-primary/30"
              }`}
            >
              {isActive ? <Plus className="h-7 w-7" /> : <span className="h-7 w-7" />}
              <span className="mt-2 text-[11px] font-semibold leading-4">{isActive ? t("sampleRequest.pickStone") : t("sampleRequest.emptySlot")}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}

function ProductPickerModal({
  categories,
  categoryLoading,
  activeCategoryId,
  setActiveCategoryId,
  search,
  setSearch,
  products,
  productsLoading,
  productsTotal,
  productsOffset,
  setProductsOffset,
  selectedStones,
  replaceIndex,
  lang,
  t,
  onSelect,
  onClose
}) {
  const selectedIDs = new Set(selectedStones.map((stone) => Number(stone.id)));
  const replacingID = replaceIndex !== null ? Number(selectedStones[replaceIndex]?.id) : null;
  const canGoPrev = productsOffset > 0;
  const canGoNext = productsOffset + 9 < productsTotal;

  return (
    <div className="fixed inset-0 z-[10000] flex items-center justify-center bg-primary/45 px-4 py-6 backdrop-blur-sm">
      <div className="flex max-h-[90dvh] w-full max-w-5xl flex-col overflow-hidden rounded-3xl border border-white/40 bg-sand shadow-2xl">
        <div className="flex items-center justify-between gap-4 border-b border-primary/10 px-5 py-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-primary/55">{t("sampleRequest.modalKicker")}</p>
            <h2 className="mt-1 font-display text-2xl">{t("sampleRequest.modalTitle")}</h2>
          </div>
          <button type="button" onClick={onClose} className="inline-flex h-10 w-10 items-center justify-center rounded-full hover:bg-primary/10" aria-label={t("actions.close")}>
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="grid min-h-0 flex-1 gap-0 overflow-hidden md:grid-cols-[16rem_minmax(0,1fr)]">
          <aside className="min-h-0 overflow-y-auto border-b border-primary/10 p-4 md:border-b-0 md:border-r">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-primary/55">{t("sampleRequest.categories")}</p>
            {categoryLoading ? (
              <p className="mt-3 text-sm text-primary/60">{t("messages.loading")}</p>
            ) : (
              <div className="mt-3 flex gap-2 overflow-x-auto md:flex-col md:overflow-visible">
                {categories.map((category) => (
                  <button
                    key={category.id}
                    type="button"
                    onClick={() => setActiveCategoryId(String(category.id))}
                    className={`shrink-0 rounded-full border px-4 py-2 text-xs font-semibold transition md:w-full md:text-start ${
                      String(category.id) === String(activeCategoryId)
                        ? "border-primary bg-primary text-sand"
                        : "border-primary/15 bg-white/55 text-primary/70 hover:border-primary/45"
                    }`}
                  >
                    {getLocalizedTitle(category, lang)}
                  </button>
                ))}
              </div>
            )}
          </aside>

          <main className="min-h-0 overflow-y-auto p-4">
            <label className="sr-only" htmlFor="sample-product-search">{t("products.searchLabel")}</label>
            <div className="relative">
              <input
                id="sample-product-search"
                type="search"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder={t("sampleRequest.searchPlaceholder")}
                className="w-full rounded-full border border-primary/20 bg-white/75 px-4 py-3 pr-10 text-sm font-semibold text-primary outline-none focus:border-primary/60"
              />
              <Search className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-primary/45" />
            </div>

            {!activeCategoryId ? (
              <div className="mt-6 rounded-2xl border border-primary/10 bg-white/55 px-4 py-8 text-center text-sm text-primary/60">
                {t("sampleRequest.chooseCategory")}
              </div>
            ) : productsLoading ? (
              <div className="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {Array.from({ length: 9 }).map((_, index) => (
                  <div key={index} className="aspect-square animate-pulse rounded-2xl bg-primary/10" />
                ))}
              </div>
            ) : products.length === 0 ? (
              <div className="mt-6 rounded-2xl border border-primary/10 bg-white/55 px-4 py-8 text-center text-sm text-primary/60">
                {t("sampleRequest.emptyProducts")}
              </div>
            ) : (
              <div className="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {products.map((product) => {
                  const alreadySelected = selectedIDs.has(Number(product.id)) && replacingID !== Number(product.id);
                  return (
                    <button
                      key={product.id}
                      type="button"
                      onClick={() => onSelect(product)}
                      disabled={alreadySelected}
                      className="group relative aspect-square overflow-hidden rounded-2xl border border-primary/10 bg-primary/5 text-left transition hover:-translate-y-0.5 disabled:cursor-not-allowed disabled:opacity-55"
                    >
                      {product.image_url ? (
                        <img src={resolveImageUrl(product.image_url)} alt={getLocalizedTitle(product, lang)} className="h-full w-full object-cover transition duration-500 group-hover:scale-105" />
                      ) : (
                        <div className="flex h-full w-full items-center justify-center text-sm text-primary/50">{t("productDetail.noImages")}</div>
                      )}
                      <div className="absolute inset-0 bg-gradient-to-t from-black/75 via-black/20 to-transparent" />
                      <div className="absolute inset-x-0 bottom-0 p-4 text-white">
                        <h3 className="text-center font-display text-lg leading-tight drop-shadow">{getLocalizedTitle(product, lang)}</h3>
                        {alreadySelected ? (
                          <p className="mx-auto mt-2 w-max rounded-full bg-white/20 px-3 py-1 text-[11px] font-semibold backdrop-blur">
                            {t("sampleRequest.alreadySelected")}
                          </p>
                        ) : null}
                      </div>
                    </button>
                  );
                })}
              </div>
            )}

            <div className="mt-5 flex items-center justify-between gap-3">
              <button
                type="button"
                disabled={!canGoPrev}
                onClick={() => setProductsOffset(Math.max(0, productsOffset - 9))}
                className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70 disabled:opacity-40"
              >
                {t("actions.prev")}
              </button>
              <span className="text-xs font-semibold text-primary/50">{productsTotal} {t("sampleRequest.products")}</span>
              <button
                type="button"
                disabled={!canGoNext}
                onClick={() => setProductsOffset(productsOffset + 9)}
                className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70 disabled:opacity-40"
              >
                {t("actions.next")}
              </button>
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}

function SummaryRow({ label, value, strong = false }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <span className="text-primary/55">{label}</span>
      <span className={`${strong ? "text-base font-bold text-primary" : "font-semibold text-primary/85"} text-end`}>{value}</span>
    </div>
  );
}
