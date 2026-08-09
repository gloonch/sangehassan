import { DEFAULT_LANGUAGE } from "./languageConfig.js";

export const BLOG_LOCALES = ["en", "fa", "ar"];

export function normalizeBlogLocales(locales = []) {
  const available = new Set(Array.isArray(locales) ? locales : []);
  return BLOG_LOCALES.filter((locale) => available.has(locale));
}

export function defaultBlogLocale(locales = []) {
  const available = normalizeBlogLocales(locales);
  if (available.includes(DEFAULT_LANGUAGE)) return DEFAULT_LANGUAGE;
  return available[0] || "";
}

export function blogHubAlternates(locales = []) {
  const available = normalizeBlogLocales(locales);
  const fallback = defaultBlogLocale(available);
  const alternates = available.map((locale) => ({ lang: locale, path: `/${locale}/blogs` }));
  if (fallback) alternates.push({ lang: "x-default", path: `/${fallback}/blogs` });
  return alternates;
}
