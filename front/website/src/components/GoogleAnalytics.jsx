import { useEffect, useRef } from "react";
import { useLocation } from "react-router-dom";

const GA_MEASUREMENT_ID = "G-QB0SMJ2BXJ";

const loadAnalytics = () => {
  if (document.getElementById("sangehassan-google-analytics")) return;
  window.dataLayer = window.dataLayer || [];
  window.gtag = window.gtag || function gtag() { window.dataLayer.push(arguments); };
  window.gtag("js", new Date());
  window.gtag("config", GA_MEASUREMENT_ID);

  const script = document.createElement("script");
  script.id = "sangehassan-google-analytics";
  script.async = true;
  script.src = `https://www.googletagmanager.com/gtag/js?id=${GA_MEASUREMENT_ID}`;
  document.head.appendChild(script);
};

export default function GoogleAnalytics() {
  const location = useLocation();
  const previousPath = useRef(null);

  useEffect(() => {
    const start = () => loadAnalytics();
    const timeoutId = window.setTimeout(start, 10000);
    window.addEventListener("pointerdown", start, { once: true, passive: true });
    window.addEventListener("keydown", start, { once: true });

    return () => {
      window.clearTimeout(timeoutId);
      window.removeEventListener("pointerdown", start);
      window.removeEventListener("keydown", start);
    };
  }, []);

  useEffect(() => {
    const pagePath = `${location.pathname}${location.search}${location.hash}`;

    if (previousPath.current === null) {
      previousPath.current = pagePath;
      return undefined;
    }

    if (previousPath.current === pagePath) return undefined;
    previousPath.current = pagePath;

    const timeoutId = window.setTimeout(() => {
      if (typeof window.gtag !== "function") return;

      window.gtag("config", GA_MEASUREMENT_ID, {
        page_path: pagePath,
        page_location: window.location.href,
        page_title: document.title
      });
    }, 0);

    return () => window.clearTimeout(timeoutId);
  }, [location.hash, location.pathname, location.search]);

  return null;
}
