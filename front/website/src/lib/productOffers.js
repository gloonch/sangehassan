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

const unitLabels = {
  en: "per m²",
  fa: "به‌ازای هر مترمربع",
  ar: "لكل متر مربع"
};

const schemaAvailabilityByKey = {
  "in-stock": "https://schema.org/InStock",
  "available-on-order": "https://schema.org/PreOrder",
  limited: "https://schema.org/LimitedAvailability",
  unavailable: "https://schema.org/OutOfStock"
};

export const getProductOfferPrice = (product) => {
  if (!product?.is_popular) return 0;
  const value = typeof product.price === "number" ? product.price : Number(product.price);
  return Number.isFinite(value) && value > 0 ? Math.round(value) : 0;
};

export const formatOfferPrice = (price, lang = "en", options = {}) => {
  const value = typeof price === "number" ? price : Number(price);
  if (!Number.isFinite(value) || value <= 0) return "";

  const { withPrefix = true, withUnit = true } = options;
  const locale = numberLocales[lang] || numberLocales.en;
  const formatted = Math.round(value).toLocaleString(locale);
  const label = currencyLabels[lang] || currencyLabels.en;
  const unit = withUnit ? ` ${unitLabels[lang] || unitLabels.en}` : "";
  const priceText = `${formatted} ${label}${unit}`;

  if (!withPrefix) return priceText;
  return `${fromLabels[lang] || fromLabels.en} ${priceText}`;
};

export const getProductOfferStructuredData = (product, url = "") => {
  const priceToman = getProductOfferPrice(product);
  if (priceToman <= 0) return undefined;

  const priceRial = priceToman * 10;
  const availabilityKey = (product?.terms || []).find((term) => term.taxonomy === "availability")?.key;
  const availability = schemaAvailabilityByKey[availabilityKey];
  const offer = {
    "@type": "Offer",
    price: String(priceRial),
    priceCurrency: "IRR",
    priceSpecification: {
      "@type": "UnitPriceSpecification",
      price: String(priceRial),
      priceCurrency: "IRR",
      unitCode: "MTK",
      referenceQuantity: {
        "@type": "QuantitativeValue",
        value: 1,
        unitCode: "MTK"
      }
    }
  };

  if (url) offer.url = url;
  if (availability) offer.availability = availability;
  return offer;
};
