import { useEffect, useState } from "react";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

const formatDateTime = (value) => {
  if (!value) return "-";
  const dt = new Date(value);
  if (Number.isNaN(dt.getTime())) return "-";
  return dt.toLocaleString();
};

export default function Ads() {
  const { t } = useTranslation();
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadAds = async () => {
    try {
      const res = await fetchJSON("/api/admin/ads?limit=100");
      setItems(res.data || []);
      setError("");
    } catch (err) {
      setItems([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    let mounted = true;
    (async () => {
      await loadAds();
      if (!mounted) return;
    })();
    return () => {
      mounted = false;
    };
  }, [t]);

  const handleDelete = async (id) => {
    try {
      await fetchJSON(`/api/admin/ads/${id}`, { method: "DELETE" });
      loadAds();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  return (
    <section className="panel-card">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h2 className="font-display text-xl">{t("panelAds.title")}</h2>
        <span className="rounded-full bg-accent/10 px-3 py-1 text-xs font-semibold text-accent">
          {items.length} {t("panelAds.countLabel")}
        </span>
      </div>

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : error ? (
        <p className="text-sm text-red-500">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-primary/70">{t("panelAds.empty")}</p>
      ) : (
        <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
          {items.map((ad) => (
            <div
              key={ad.id}
              className="rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-sm font-semibold text-primary">
                  {ad.title || ad.stone_type || `#${ad.id}`}
                </p>
                <span className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold text-primary/70">
                  {ad.status || "-"}
                </span>
              </div>
              <p className="mt-2 text-xs text-primary/70">
                {t("panelAds.author")}: {ad.author?.full_name || ad.author?.email || "-"}
              </p>
              <p className="mt-1 text-xs text-primary/70">
                {t("panelAds.authorContact")}: {ad.author?.phone || ad.author?.email || "-"}
              </p>
              <p className="mt-1 text-xs text-primary/60">
                {t("panelAds.createdAt")}: {formatDateTime(ad.created_at)}
              </p>
              <div className="mt-3 flex justify-end">
                <button
                  type="button"
                  onClick={() => handleDelete(ad.id)}
                  className="rounded-full border border-red-200 px-3 py-1 text-xs font-semibold text-red-500"
                >
                  {t("actions.delete")}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
