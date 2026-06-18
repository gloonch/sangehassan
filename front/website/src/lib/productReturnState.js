const CATALOG_RETURN_STATE_KEY = "sh_catalog_product_return_state";
const LEGACY_PRODUCTS_RETURN_STATE_KEY = "sh_products_return_state";
const RETURN_STATE_MAX_AGE_MS = 30 * 60 * 1000;

const readJSON = (key) => {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.sessionStorage.getItem(key);
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    const ts = Number(parsed?.ts || 0);
    if (!ts || Date.now() - ts > RETURN_STATE_MAX_AGE_MS) {
      window.sessionStorage.removeItem(key);
      return null;
    }
    return parsed;
  } catch (_) {
    return null;
  }
};

export const currentCatalogReturnPath = (location) => `${location?.pathname || ""}${location?.search || ""}${location?.hash || ""}`;

export const currentHistoryKey = () => {
  if (typeof window === "undefined") return "";
  return String(window.history.state?.key || "");
};

export const writeCatalogProductReturnState = ({ path, scrollY, productCount, productSlug, lang }) => {
  if (typeof window === "undefined" || !path) return;
  try {
    window.sessionStorage.setItem(
      CATALOG_RETURN_STATE_KEY,
      JSON.stringify({
        path,
        scrollY: Math.max(0, Number(scrollY) || 0),
        productCount: Math.max(0, Number(productCount) || 0),
        productSlug: productSlug || "",
        lang: lang || "",
        historyKey: currentHistoryKey(),
        ts: Date.now()
      })
    );
  } catch (_) {
    // ignore transient storage failures
  }
};

export const readCatalogProductReturnState = (expectedPath = "") => {
  const parsed = readJSON(CATALOG_RETURN_STATE_KEY);
  if (!parsed?.path) return null;
  if (expectedPath && parsed.path !== expectedPath) return null;
  return {
    path: parsed.path,
    scrollY: Math.max(0, Number(parsed.scrollY) || 0),
    productCount: Math.max(0, Number(parsed.productCount) || 0),
    productSlug: parsed.productSlug || "",
    historyKey: parsed.historyKey || "",
    lang: parsed.lang || ""
  };
};

export const clearCatalogProductReturnState = () => {
  if (typeof window === "undefined") return;
  try {
    window.sessionStorage.removeItem(CATALOG_RETURN_STATE_KEY);
  } catch (_) {
    // ignore transient storage failures
  }
};

export const hasLegacyProductsReturnState = () => Boolean(readJSON(LEGACY_PRODUCTS_RETURN_STATE_KEY));
