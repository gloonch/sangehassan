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
