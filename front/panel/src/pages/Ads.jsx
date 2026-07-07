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
  const [productRequests, setProductRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadAds = async () => {
    try {
      const [adsRes, productRequestsRes] = await Promise.all([
        fetchJSON("/api/admin/ads?limit=100"),
        fetchJSON("/api/admin/product-requests?limit=50")
      ]);
      setItems(adsRes.data || []);
      setProductRequests(productRequestsRes.data || []);
      setError("");
    } catch (err) {
      setItems([]);
      setProductRequests([]);
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
      ) : items.length === 0 && productRequests.length === 0 ? (
        <p className="text-sm text-primary/70">{t("panelAds.empty")}</p>
      ) : (
        <>
          <div className="mb-5 rounded-xl border border-primary/10 bg-white/80 px-4 py-3">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <p className="text-sm font-semibold text-primary">{t("panelAds.productRequests")}</p>
              <span className="rounded-full bg-primary/10 px-3 py-1 text-xs font-semibold text-primary/70">
                {productRequests.length}
              </span>
            </div>
            {productRequests.length === 0 ? (
              <p className="text-xs text-primary/60">{t("panelAds.productRequestsEmpty")}</p>
            ) : (
              <div className="space-y-2">
                {productRequests.map((request) => (
                  <div key={request.id} className="rounded-lg border border-primary/10 bg-primary/5 px-3 py-2">
                    <p className="text-sm font-semibold text-primary">
                      {t("panelAds.requestedProduct")}: {request.query}
                    </p>
                    <p className="mt-1 text-xs text-primary/70">
                      {t("panelAds.authorContact")}: {request.user?.phone || request.user?.email || "-"}
                    </p>
                    <p className="mt-1 text-xs text-primary/60">
                      {t("panelAds.createdAt")}: {formatDateTime(request.created_at)}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
            {items.map((ad) => (
              <div
                key={ad.id}
                className="rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <p className="text-sm font-semibold text-primary">
                    {ad.title || ad.product?.title_en || ad.stone_type || `#${ad.id}`}
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
        </>
      )}
    </section>
  );
}
