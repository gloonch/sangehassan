import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

export default function Ads() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
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
    load();
    return () => {
      active = false;
    };
  }, [t]);

  return (
    <section className="section-shell py-16">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h1 className="font-display text-3xl">{t("ads.title")}</h1>
          <p className="text-sm text-primary/70">{t("ads.subtitle")}</p>
        </div>
        <button
          type="button"
          onClick={() => navigate("/ads/new")}
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
            {items.map((ad) => (
              <Link
                to={`/ads/${ad.id}`}
                key={ad.id}
                className="group rounded-2xl border border-primary/10 bg-white/80 p-4 shadow-sm transition hover:-translate-y-1 hover:shadow-lg"
              >
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
                      {ad.price_amount} {ad.price_unit}
                    </span>
                  ) : (
                    <span className="rounded-full bg-primary/10 px-3 py-1 text-xs">{t("ads.viewDetails")}</span>
                  )}
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
