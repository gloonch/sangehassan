import { useEffect, useRef } from "react";
import { useLocation } from "react-router-dom";

const GA_MEASUREMENT_ID = "G-QB0SMJ2BXJ";

export default function GoogleAnalytics() {
  const location = useLocation();
  const previousPath = useRef(null);

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
