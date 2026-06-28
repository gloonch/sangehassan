const numberLocales = {
  en: "en-US",
  fa: "fa-IR",
  ar: "ar-SA"
};

const currencyLabels = {
  en: "Toman",
  fa: "تومان",
  ar: "تومان"
};

const fromLabels = {
  en: "From",
  fa: "از",
  ar: "من"
};

export const getProductOfferPrice = (product) => {
  if (!product?.is_popular) return 0;
  const value = typeof product.price === "number" ? product.price : Number(product.price);
  return Number.isFinite(value) && value > 0 ? Math.round(value) : 0;
};

export const formatOfferPrice = (price, lang = "en", options = {}) => {
  const value = typeof price === "number" ? price : Number(price);
  if (!Number.isFinite(value) || value <= 0) return "";

  const { withPrefix = true } = options;
  const locale = numberLocales[lang] || numberLocales.en;
  const formatted = Math.round(value).toLocaleString(locale);
  const label = currencyLabels[lang] || currencyLabels.en;
  const priceText = `${formatted} ${label}`;

  if (!withPrefix) return priceText;
  return `${fromLabels[lang] || fromLabels.en} ${priceText}`;
};
