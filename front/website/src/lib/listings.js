export const PRICE_UNIT_VALUES = ["per_ton", "per_meter", "per_cubic_meter", "total", "negotiable"];

export function formatPriceUnit(unit, t) {
  if (!unit) return "—";
  const key = `ads.priceUnitOptions.${unit}`;
  const translated = t(key);
  if (translated && translated !== key) return translated;
  return String(unit).replaceAll("_", " ");
}

export function formatPriceValue(amount, unit, t) {
  if (amount === undefined || amount === null || amount === "") return "—";
  return `${amount} ${formatPriceUnit(unit, t)}`.trim();
}

export function getLocalizedProductTitle(product, lang = "en") {
  if (!product) return "";
  return product[`title_${lang}`] || product.title_en || product.title_fa || product.title_ar || "";
}

export function getListingProductTitle(ad, lang = "en") {
  return getLocalizedProductTitle(ad?.product, lang);
}

export function getListingProductPath(product, lang = "en") {
  if (!product?.slug) return "";
  const safeLang = ["en", "fa", "ar"].includes(lang) ? lang : "en";
  return `/${safeLang}/products/${product.slug}`;
}

export function getListingCoverImageUrl(ad) {
  if (ad?.product?.image_url) return ad.product.image_url;
  const images = Array.isArray(ad?.images) ? ad.images : [];
  for (let i = images.length - 1; i >= 0; i -= 1) {
    const imageUrl = images[i]?.image_url;
    if (imageUrl) return imageUrl;
  }
  return "";
}

const quantityLocaleByLanguage = {
  en: "en-US",
  fa: "fa-IR",
  ar: "ar-SA"
};

export function getListingQuantityUnit(form) {
  return form === "block" ? "ton" : "square_meter";
}

export function getListingQuantityFieldLabel(form, t) {
  return t(form === "block" ? "ads.form.quantityBlock" : "ads.form.quantityFinished");
}

export function formatListingQuantity(ad, lang, t) {
  if (ad?.tonnage === undefined || ad?.tonnage === null || ad?.tonnage === "") return "—";
  const value = Number(ad?.tonnage);
  if (!Number.isFinite(value)) return "—";
  const unit = getListingQuantityUnit(ad?.form);
  const unitKey = `ads.liveFeed.units.${unit}`;
  const unitLabel = t(unitKey);
  const formattedValue = new Intl.NumberFormat(quantityLocaleByLanguage[lang] || quantityLocaleByLanguage.en, {
    maximumFractionDigits: 2
  }).format(value);
  return `${formattedValue} ${unitLabel === unitKey ? "" : unitLabel}`.trim();
}
