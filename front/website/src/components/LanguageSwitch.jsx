import { useEffect, useMemo, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { getLanguageTarget } from "../lib/languageRoutes";
import { usePrerenderData } from "../lib/prerenderData";

const options = [
  { code: "fa", label: "fa" },
  { code: "en", label: "en" },
  { code: "ar", label: "ar" }
];

export default function LanguageSwitch({ tone = "default" }) {
  const { lang, setLang, t } = useTranslation();
  const location = useLocation();
  const prerenderBlog = usePrerenderData("blog");
  const prerenderAlternates = useMemo(
    () => Object.fromEntries((prerenderBlog?.translations || []).map((item) => [item.locale, `/${item.locale}/blogs/${item.slug}`])),
    [prerenderBlog?.translations]
  );
  const [runtimeAlternates, setRuntimeAlternates] = useState({});
  const isLightTone = tone === "light";
  const isMenuTone = tone === "menu";

  useEffect(() => {
    const updateAlternates = (event) => setRuntimeAlternates(event.detail || {});
    setRuntimeAlternates(typeof window !== "undefined" ? window.__SH_BLOG_ALTERNATES__ || {} : {});
    window.addEventListener("sh:blog-alternates", updateAlternates);
    return () => window.removeEventListener("sh:blog-alternates", updateAlternates);
  }, [location.pathname]);

  const blogAlternates = Object.keys(runtimeAlternates).length ? runtimeAlternates : prerenderAlternates;

  return (
    <div
      className={`inline-flex items-center rounded-full px-3 py-2 text-[11px] font-semibold uppercase tracking-[0.16em] shadow-sm transition ${isLightTone
          ? "text-sand hover:bg-sand/25"
          : isMenuTone
            ? "bg-primary/90 text-sand hover:bg-primary"
            : "text-accent shadow-none"
        }`}
      aria-label="Language"
    >
      {options.map((option, index) => {
        const isActive = lang === option.code;
        const target = getLanguageTarget({
          pathname: location.pathname,
          search: location.search,
          nextLang: option.code,
          blogAlternates
        });
        const controlClass = `px-1 leading-none transition ${isActive ? "opacity-100" : "opacity-60 hover:opacity-100"}`;
        return (
          <span key={option.code} className="inline-flex items-center">
            {target ? (
              <Link
                to={target}
                state={location.pathname.match(/^\/products\/[^/]+$/) ? { catalogRouteKind: "product" } : location.state}
                onClick={() => setLang(option.code)}
                className={controlClass}
                aria-label={t(`language.${option.code}`)}
                aria-current={isActive ? "page" : undefined}
              >
                {option.label}
              </Link>
            ) : (
              <button
                type="button"
                onClick={() => setLang(option.code)}
                className={controlClass}
                aria-label={t(`language.${option.code}`)}
                aria-pressed={isActive}
              >
                {option.label}
              </button>
            )}
            {index < options.length - 1 ? <span className="px-1 opacity-45">|</span> : null}
          </span>
        );
      })}
    </div>
  );
}
