import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";
import { formatPriceValue } from "../lib/listings";
import { resolveImageUrl } from "../lib/assets";
import { usePageSeo } from "../lib/seo";

const getLatestImageUrl = (ad) => {
  const images = Array.isArray(ad?.images) ? ad.images : [];
  for (let i = images.length - 1; i >= 0; i -= 1) {
    const imageUrl = images[i]?.image_url;
    if (imageUrl) return imageUrl;
  }
  return "";
};

export default function Ads() {
  const { t, lang } = useTranslation();
  const navigate = useNavigate();
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const latestImage = useMemo(() => items.map(getLatestImageUrl).find(Boolean) || "", [items]);
  const jsonLd = useMemo(
    () => ({
      "@context": "https://schema.org",
      "@type": "CollectionPage",
      inLanguage: lang,
      name: t("ads.title"),
      description: t("ads.subtitle"),
      url: typeof window !== "undefined" ? `${window.location.origin}/ads` : undefined
    }),
    [lang, t]
  );

  usePageSeo({
    title: `${t("ads.title")} | SangeHassan`,
    description: t("ads.subtitle"),
    path: "/ads",
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    image: latestImage ? resolveImageUrl(latestImage) : "",
    jsonLdId: "ads-jsonld",
    jsonLd
  });

  useEffect(() => {
    let active = true;
    const restore = () => {
      try {
        const stored = sessionStorage.getItem("sh_me");
        if (stored) {
          const parsed = JSON.parse(stored);
          if (parsed?.id) setIsAuthenticated(true);
        }
      } catch (_) {
        /* ignore */
      }
    };

    const loadMe = async () => {
      try {
        const res = await fetchJSON("/api/v1/me");
        const me = res?.data || res;
        if (!active) return;
        setIsAuthenticated(true);
        sessionStorage.setItem("sh_me", JSON.stringify(me));
      } catch (_) {
        if (!active) return;
        setIsAuthenticated(false);
      }
    };

    const load = async () => {
      try {
        const res = await fetchJSON("/api/ads");
        const data = res?.data || res;
        if (!active) return;
        setItems(Array.isArray(data) ? data : []);
      } catch (err) {
        if (!active) return;
        setError(err?.message || t("messages.error"));
      } finally {
        if (active) setLoading(false);
      }
    };
    restore();
    loadMe();
    load();
    return () => {
      active = false;
    };
  }, [t]);

  const handleCreate = () => {
    if (isAuthenticated) {
      navigate("/ads/new");
      return;
    }
    sessionStorage.setItem("sh_after_login", "/ads/new");
    navigate("/login");
  };

  return (
    <section className="section-shell py-16">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h1 className="font-display text-3xl">{t("ads.title")}</h1>
          <p className="text-sm text-primary/70">{t("ads.subtitle")}</p>
        </div>
        <button
          type="button"
          onClick={handleCreate}
          className="rounded-full bg-primary px-4 py-2 text-xs font-semibold text-sand shadow hover:bg-primary/90"
        >
          {t("ads.create")}
        </button>
      </div>
      <div className="mt-4 rounded-2xl border border-primary/15 bg-primary/5 px-4 py-3 text-sm text-primary/80">
        {t("ads.privacyNote")}
      </div>

      <div className="mt-8">
        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : error ? (
          <p className="text-sm text-red-600">{error}</p>
        ) : items.length === 0 ? (
          <p className="text-sm text-primary/70">{t("ads.empty")}</p>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {items.map((ad) => {
              const latestImageUrl = getLatestImageUrl(ad);
              return (
                <Link
                  to={`/ads/${ad.id}`}
                  key={ad.id}
                  className="group overflow-hidden rounded-2xl border border-primary/10 bg-white/80 shadow-sm transition hover:-translate-y-1 hover:shadow-lg"
                >
                  <div className="relative aspect-[4/3] bg-primary/5">
                    {latestImageUrl ? (
                      <img
                        src={resolveImageUrl(latestImageUrl)}
                        alt={ad.title || ad.stone_type || t("ads.title")}
                        loading="lazy"
                        className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.03]"
                      />
                    ) : (
                      <div className="flex h-full w-full items-center justify-center px-4 text-center text-xs font-semibold text-primary/50">
                        {t("productDetail.noImages")}
                      </div>
                    )}
                  </div>
                  <div className="p-4">
                    <div className="flex items-center justify-between text-xs text-primary/60">
                      <span>{ad.form || "—"}</span>
                      <span>{ad.city || ad.province || "—"}</span>
                    </div>
                    <h3 className="mt-2 text-lg font-semibold text-primary group-hover:text-accent">
                      {ad.title || ad.stone_type || t("ads.title")}
                    </h3>
                    <p className="mt-1 text-sm text-primary/70 line-clamp-2">{ad.description || " "}</p>
                    <div className="mt-3 flex flex-wrap items-center gap-2 text-sm font-semibold text-primary">
                      {ad.tonnage ? <span>{ad.tonnage} t</span> : null}
                      {ad.price_amount ? (
                        <span className="rounded-full bg-primary/10 px-3 py-1 text-xs">
                          {formatPriceValue(ad.price_amount, ad.price_unit, t)}
                        </span>
                      ) : (
                        <span className="rounded-full bg-primary/10 px-3 py-1 text-xs">{t("ads.viewDetails")}</span>
                      )}
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </div>
    </section>
  );
}
