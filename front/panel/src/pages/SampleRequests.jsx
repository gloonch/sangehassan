import { useEffect, useMemo, useState } from "react";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { useTranslation } from "../lib/i18n";

const statuses = ["PENDING", "CONFIRMED", "REJECTED", "SHIPPED", "DELIVERED", "CANCELED"];

const formatDateTime = (value, lang) => {
  if (!value) return "-";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return "-";
  const locale = lang === "fa" ? "fa-IR" : lang === "ar" ? "ar-SA" : "en-US";
  return parsed.toLocaleString(locale);
};

const formatMoney = (value, lang) => {
  const locale = lang === "fa" ? "fa-IR" : lang === "ar" ? "ar-SA" : "en-US";
  return new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(Number(value) || 0);
};

const itemTitle = (item, lang) => {
  if (!item) return "";
  if (lang === "fa") return item.product_title_fa || item.product_title_en || item.product_title_ar || "";
  if (lang === "ar") return item.product_title_ar || item.product_title_en || item.product_title_fa || "";
  return item.product_title_en || item.product_title_fa || item.product_title_ar || "";
};

const groupItems = (items = []) => {
  const map = new Map();
  for (const item of Array.isArray(items) ? items : []) {
    const key = item.box_index || 1;
    if (!map.has(key)) map.set(key, []);
    map.get(key).push(item);
  }
  return [...map.entries()].map(([boxIndex, stones]) => [
    boxIndex,
    stones.sort((a, b) => (a.slot_index || 0) - (b.slot_index || 0))
  ]);
};

export default function SampleRequests() {
  const { t, lang } = useTranslation();
  const [items, setItems] = useState([]);
  const [selectedId, setSelectedId] = useState(null);
  const [selectedDetail, setSelectedDetail] = useState(null);
  const [loading, setLoading] = useState(true);
  const [detailLoading, setDetailLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("PENDING");
  const [adminNote, setAdminNote] = useState("");

  const selected = selectedDetail || items.find((item) => Number(item.id) === Number(selectedId)) || items[0] || null;
  const boxes = useMemo(() => groupItems(selected?.items), [selected]);

  const loadItems = async () => {
    setLoading(true);
    try {
      const res = await fetchJSON("/api/admin/sample-requests?limit=100");
      const data = res?.data || [];
      setItems(Array.isArray(data) ? data : []);
      setError("");
      if (!selectedId && data?.[0]?.id) setSelectedId(data[0].id);
    } catch (_) {
      setItems([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadItems();
  }, []);

  useEffect(() => {
    if (!selectedId) {
      setSelectedDetail(null);
      return undefined;
    }
    let active = true;
    const loadDetail = async () => {
      setDetailLoading(true);
      try {
        const res = await fetchJSON(`/api/admin/sample-requests/${selectedId}`);
        const data = res?.data || null;
        if (!active) return;
        setSelectedDetail(data);
        setStatus(data?.status || "PENDING");
        setAdminNote(data?.admin_note || "");
      } catch (_) {
        if (active) setError(t("messages.error"));
      } finally {
        if (active) setDetailLoading(false);
      }
    };
    loadDetail();
    return () => {
      active = false;
    };
  }, [selectedId, t]);

  const handleStatusUpdate = async (event) => {
    event.preventDefault();
    if (!selected) return;
    setSaving(true);
    setError("");
    try {
      await fetchJSON(`/api/admin/sample-requests/${selected.id}/status`, {
        method: "PUT",
        body: JSON.stringify({ status, admin_note: adminNote })
      });
      await loadItems();
      const res = await fetchJSON(`/api/admin/sample-requests/${selected.id}`);
      setSelectedDetail(res?.data || null);
    } catch (_) {
      setError(t("messages.error"));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="grid gap-6 lg:grid-cols-[22rem_minmax(0,1fr)]">
      <section className="panel-card">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="font-display text-xl">{t("panelSampleRequests.title")}</h2>
          <span className="rounded-full bg-accent/10 px-3 py-1 text-xs font-semibold text-accent">
            {items.length} {t("panelSampleRequests.countLabel")}
          </span>
        </div>

        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : error ? (
          <p className="text-sm text-red-500">{error}</p>
        ) : items.length === 0 ? (
          <p className="text-sm text-primary/70">{t("panelSampleRequests.empty")}</p>
        ) : (
          <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
            {items.map((item) => (
              <button
                key={item.id}
                type="button"
                onClick={() => setSelectedId(item.id)}
                className={`w-full rounded-xl border px-4 py-3 text-left transition ${
                  Number(selectedId) === Number(item.id)
                    ? "border-primary bg-primary text-sand"
                    : "border-primary/10 bg-white/80 text-primary hover:border-primary/35"
                }`}
              >
                <div className="flex items-center justify-between gap-2">
                  <p className="text-sm font-semibold">#{item.id}</p>
                  <span className="rounded-full border border-current/20 px-2 py-0.5 text-[10px] font-semibold">
                    {t(`panelSampleRequests.status.${item.status}`)}
                  </span>
                </div>
                <p className="mt-2 text-xs opacity-75">
                  {item.user?.full_name || item.user?.phone || item.user?.email || "-"}
                </p>
                <p className="mt-1 text-xs opacity-60">
                  {item.box_count} {t("panelSampleRequests.boxes")} / {formatDateTime(item.created_at, lang)}
                </p>
              </button>
            ))}
          </div>
        )}
      </section>

      <section className="panel-card min-h-[28rem]">
        {!selected ? (
          <p className="text-sm text-primary/70">{t("panelSampleRequests.empty")}</p>
        ) : detailLoading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : (
          <>
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <p className="text-xs uppercase tracking-[0.25em] text-primary/55">{t("panelSampleRequests.detailKicker")} #{selected.id}</p>
                <h3 className="mt-2 font-display text-2xl">{t("panelSampleRequests.detailTitle")}</h3>
                <p className="mt-2 text-sm text-primary/65">{formatDateTime(selected.created_at, lang)}</p>
              </div>
              <span className="rounded-full bg-primary/10 px-3 py-1 text-xs font-semibold text-primary/75">
                {t(`panelSampleRequests.status.${selected.status}`)}
              </span>
            </div>

            <div className="mt-6 grid gap-3 md:grid-cols-2">
              <Info label={t("panelSampleRequests.customer")} value={selected.user?.full_name || selected.user?.email || "-"} />
              <Info label={t("panelSampleRequests.phone")} value={selected.phone_snapshot || selected.user?.phone || "-"} />
              <Info label={t("panelSampleRequests.address")} value={selected.address_snapshot || "-"} />
              <Info label={t("panelSampleRequests.shipping")} value={t(`sampleRequest.shipping.${selected.shipping_method}`)} />
              <Info label={t("panelSampleRequests.boxes")} value={selected.box_count} />
              <Info label={t("panelSampleRequests.total")} value={`${formatMoney(selected.total_price_toman, lang)} ${t("sampleRequest.toman")}`} />
            </div>

            <div className="mt-6 space-y-4">
              {boxes.map(([boxIndex, stones]) => (
                <div key={boxIndex} className="rounded-2xl border border-primary/10 bg-white/70 p-4">
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-primary/55">
                    {t("sampleRequest.box")} {boxIndex}
                  </p>
                  <div className="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                    {stones.map((stone) => (
                      <div key={stone.id} className="overflow-hidden rounded-xl border border-primary/10 bg-primary/5">
                        {stone.product_image_url ? (
                          <img
                            src={resolveImageUrl(stone.product_image_url)}
                            alt={itemTitle(stone, lang)}
                            className="h-28 w-full object-cover"
                            loading="lazy"
                          />
                        ) : (
                          <div className="flex h-28 items-center justify-center text-xs text-primary/50">
                            {t("productDetail.noImages")}
                          </div>
                        )}
                        <div className="p-2">
                          <p className="truncate text-xs font-semibold text-primary">{itemTitle(stone, lang)}</p>
                          <p className="mt-1 truncate text-[11px] text-primary/55">
                            {stone.category_title_fa || stone.category_title_en || stone.category_title_ar || "-"}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>

            <form onSubmit={handleStatusUpdate} className="mt-6 rounded-2xl border border-primary/10 bg-primary/5 p-4">
              <div className="grid gap-3 md:grid-cols-[14rem_minmax(0,1fr)]">
                <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("panelSampleRequests.statusLabel")}
                  <select value={status} onChange={(event) => setStatus(event.target.value)} className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm">
                    {statuses.map((item) => (
                      <option key={item} value={item}>{t(`panelSampleRequests.status.${item}`)}</option>
                    ))}
                  </select>
                </label>
                <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("panelSampleRequests.adminNote")}
                  <textarea value={adminNote} onChange={(event) => setAdminNote(event.target.value)} rows="3" className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm" />
                </label>
              </div>
              <button type="submit" disabled={saving} className="mt-4 rounded-full bg-primary px-5 py-2 text-xs font-semibold text-sand disabled:opacity-60">
                {saving ? t("messages.loading") : t("panelSampleRequests.updateStatus")}
              </button>
            </form>
          </>
        )}
      </section>
    </div>
  );
}

function Info({ label, value }) {
  return (
    <div className="rounded-xl border border-primary/10 bg-white/70 px-4 py-3">
      <p className="text-[11px] uppercase tracking-[0.16em] text-primary/50">{label}</p>
      <p className="mt-1 break-words text-sm font-semibold text-primary/85">{value}</p>
    </div>
  );
}
