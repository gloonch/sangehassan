const localizedBlogPattern = /^\/(en|fa|ar)\/blogs(?:\/.*)?$/;
const localizedCatalogPattern = /^\/(?:en|fa|ar)\/products(\/.*)?$/;

export function getLanguageTarget({ pathname, search = "", nextLang, blogAlternates = {} }) {
  const currentBlogLocale = String(pathname || "").match(localizedBlogPattern)?.[1];
  if (currentBlogLocale) {
    if (nextLang === currentBlogLocale) return `${pathname}${search}`;
    return blogAlternates[nextLang] || `/${nextLang}/blogs`;
  }

  const catalogMatch = String(pathname || "").match(localizedCatalogPattern);
  if (catalogMatch) return `/${nextLang}/products${catalogMatch[1] || ""}${search}`;

  const legacyProductMatch = String(pathname || "").match(/^\/products\/([^/]+)$/);
  if (legacyProductMatch) return `/${nextLang}/products/${legacyProductMatch[1]}`;
  if (pathname === "/products") return `/${nextLang}/products`;
  return "";
}
