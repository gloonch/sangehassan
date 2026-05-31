import { useEffect, useLayoutEffect } from "react";

const configuredSiteUrl = (import.meta.env.VITE_SITE_URL || "").replace(/\/+$/, "");
const useIsomorphicLayoutEffect = typeof window === "undefined" ? useEffect : useLayoutEffect;

export const getSiteOrigin = () => {
  if (configuredSiteUrl) return configuredSiteUrl;
  if (typeof window === "undefined") return "";
  return window.location.origin;
};

export const getCanonicalUrl = (path = "/") => {
  if (/^https?:\/\//i.test(path)) return path;

  const origin = getSiteOrigin();
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return origin ? `${origin}${normalizedPath}` : normalizedPath;
};

export const getAbsoluteUrl = (url = "") => {
  if (!url) return "";
  if (/^https?:\/\//i.test(url)) return url;
  return getCanonicalUrl(url.startsWith("/") ? url : `/${url}`);
};

export const usePageSeo = ({
  title,
  description,
  path,
  lang = "en",
  locale = "en_US",
  image = "",
  type = "website",
  robots = "index,follow",
  jsonLd = null,
  jsonLdId = ""
}) => {
  useIsomorphicLayoutEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return undefined;
    if (!title || !description || !path) return undefined;

    const pageUrl = getCanonicalUrl(path);
    const imageUrl = getAbsoluteUrl(image);
    const cleanups = [];

    const upsertMeta = (selector, createAttrs, value) => {
      if (!value) return;

      let el = document.head.querySelector(selector);
      const created = !el;

      if (!el) {
        el = document.createElement("meta");
        Object.entries(createAttrs).forEach(([key, attrValue]) => {
          el.setAttribute(key, attrValue);
        });
        document.head.appendChild(el);
      }

      const prevContent = el.getAttribute("content");
      el.setAttribute("content", value);

      cleanups.push(() => {
        if (created) {
          el.remove();
          return;
        }
        if (prevContent === null) {
          el.removeAttribute("content");
          return;
        }
        el.setAttribute("content", prevContent);
      });
    };

    const upsertCanonical = (href) => {
      let el = document.head.querySelector('link[rel="canonical"]');
      const created = !el;

      if (!el) {
        el = document.createElement("link");
        el.setAttribute("rel", "canonical");
        document.head.appendChild(el);
      }

      const prevHref = el.getAttribute("href");
      el.setAttribute("href", href);

      cleanups.push(() => {
        if (created) {
          el.remove();
          return;
        }
        if (prevHref === null) {
          el.removeAttribute("href");
          return;
        }
        el.setAttribute("href", prevHref);
      });
    };

    const upsertJsonLd = (payload, id) => {
      if (!payload || !id) return;

      let script = document.getElementById(id);
      const created = !script;

      if (!script) {
        script = document.createElement("script");
        script.setAttribute("id", id);
        script.setAttribute("type", "application/ld+json");
        document.head.appendChild(script);
      }

      const prevContent = script.textContent;
      script.textContent = JSON.stringify(payload);

      cleanups.push(() => {
        if (created) {
          script.remove();
          return;
        }
        script.textContent = prevContent;
      });
    };

    document.title = title;
    upsertCanonical(pageUrl);
    upsertMeta('meta[name="description"]', { name: "description" }, description);
    upsertMeta('meta[name="robots"]', { name: "robots" }, robots);
    upsertMeta('meta[property="og:type"]', { property: "og:type" }, type);
    upsertMeta('meta[property="og:title"]', { property: "og:title" }, title);
    upsertMeta('meta[property="og:description"]', { property: "og:description" }, description);
    upsertMeta('meta[property="og:url"]', { property: "og:url" }, pageUrl);
    upsertMeta('meta[property="og:locale"]', { property: "og:locale" }, locale);
    upsertMeta('meta[property="og:image"]', { property: "og:image" }, imageUrl);
    upsertMeta('meta[name="twitter:card"]', { name: "twitter:card" }, imageUrl ? "summary_large_image" : "summary");
    upsertMeta('meta[name="twitter:title"]', { name: "twitter:title" }, title);
    upsertMeta('meta[name="twitter:description"]', { name: "twitter:description" }, description);
    upsertMeta('meta[name="twitter:image"]', { name: "twitter:image" }, imageUrl);
    upsertJsonLd(jsonLd, jsonLdId);

    return () => {
      cleanups.reverse().forEach((fn) => fn());
    };
  }, [description, image, jsonLd, jsonLdId, lang, locale, path, robots, title, type]);
};
