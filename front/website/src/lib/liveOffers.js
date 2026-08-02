const localeByLanguage = {
  en: "en-US",
  fa: "fa-IR",
  ar: "ar-SA"
};

const interpolate = (template, values) => Object.entries(values).reduce(
  (message, [key, value]) => message.replaceAll(`{${key}}`, value),
  template
);

const translatedLabel = (t, key) => {
  const value = t(key);
  return value && value !== key ? value : "";
};

export const formatLiveOfferQuantity = (quantity, lang) => {
  const numericQuantity = Number(quantity);
  if (!Number.isFinite(numericQuantity)) return "";
  return new Intl.NumberFormat(localeByLanguage[lang] || localeByLanguage.en, {
    maximumFractionDigits: 2
  }).format(numericQuantity);
};

export const renderLiveOfferMessage = (item, lang, t) => {
  const stoneName = String(item?.stoneName || item?.title || "").trim();
  const productType = translatedLabel(t, `ads.liveFeed.productTypes.${item?.productType || ""}`);
  const product = productType
    ? interpolate(t("ads.liveFeed.productTemplate"), { productType, stoneName })
    : stoneName;

  if (!product) return t("ads.liveFeed.loading");

  const quantity = formatLiveOfferQuantity(item?.quantity, lang);
  const unit = translatedLabel(t, `ads.liveFeed.units.${item?.unit || ""}`);
  if (!quantity || !unit) {
    return interpolate(t("ads.liveFeed.withoutQuantity"), { product });
  }

  return interpolate(t("ads.liveFeed.withQuantity"), { quantity, unit, product });
};
