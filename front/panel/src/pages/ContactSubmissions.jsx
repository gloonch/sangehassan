import { useEffect, useState } from "react";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

const formatDateTime = (value, lang) => {
  if (!value) return "-";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return "-";
  const locale = lang === "fa" ? "fa-IR" : lang === "ar" ? "ar-SA" : "en-US";
  return parsed.toLocaleString(locale);
};

export default function ContactSubmissions() {
  const { t, lang } = useTranslation();
  const [items, setItems] = useState([]);
  const [selectedId, setSelectedId] = useState(null);
  const [selectedDetail, setSelectedDetail] = useState(null);
  const [loading, setLoading] = useState(true);
  const [detailLoading, setDetailLoading] = useState(false);
  const [error, setError] = useState("");

  const selected = selectedDetail || items.find((item) => Number(item.id) === Number(selectedId)) || items[0] || null;

  useEffect(() => {
    let active = true;
    const loadItems = async () => {
      setLoading(true);
      try {
        const res = await fetchJSON("/api/admin/contact-submissions?limit=100");
        const data = Array.isArray(res?.data) ? res.data : [];
        if (!active) return;
        setItems(data);
        setError("");
        setSelectedId((currentId) => currentId || data[0]?.id || null);
      } catch (_) {
        if (!active) return;
        setItems([]);
        setError(t("messages.error"));
      } finally {
        if (active) setLoading(false);
      }
    };
    loadItems();
    return () => {
      active = false;
    };
  }, [t]);

  useEffect(() => {
    if (!selectedId) {
      setSelectedDetail(null);
      return undefined;
    }

    let active = true;
    const loadDetail = async () => {
      setDetailLoading(true);
      try {
        const res = await fetchJSON(`/api/admin/contact-submissions/${selectedId}`);
        if (active) setSelectedDetail(res?.data || null);
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

  return (
    <div className="grid gap-6 lg:grid-cols-[22rem_minmax(0,1fr)]">
      <section className="panel-card">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="font-display text-xl">{t("panelContactSubmissions.title")}</h2>
          <span className="rounded-full bg-accent/10 px-3 py-1 text-xs font-semibold text-accent">
            {items.length} {t("panelContactSubmissions.countLabel")}
          </span>
        </div>

        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : error ? (
          <p className="text-sm text-red-500">{error}</p>
        ) : items.length === 0 ? (
          <p className="text-sm text-primary/70">{t("panelContactSubmissions.empty")}</p>
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
                  <p className="truncate text-sm font-semibold">{item.full_name || `#${item.id}`}</p>
                  <span className="rounded-full border border-current/20 px-2 py-0.5 text-[10px] font-semibold">
                    #{item.id}
                  </span>
                </div>
                <p className="mt-2 truncate text-xs opacity-75" dir="ltr">
                  {item.phone_e164 || item.phone_number || "-"}
                </p>
                <p className="mt-1 text-xs opacity-60">{formatDateTime(item.created_at, lang)}</p>
              </button>
            ))}
          </div>
        )}
      </section>

      <section className="panel-card min-h-[28rem]">
        {!selected ? (
          <p className="text-sm text-primary/70">{t("panelContactSubmissions.empty")}</p>
        ) : detailLoading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : (
          <>
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <p className="text-xs uppercase tracking-[0.25em] text-primary/55">
                  {t("panelContactSubmissions.detailKicker")} #{selected.id}
                </p>
                <h3 className="mt-2 font-display text-2xl">{selected.full_name}</h3>
                <p className="mt-2 text-sm text-primary/65">{formatDateTime(selected.created_at, lang)}</p>
              </div>
              <span className="rounded-full bg-primary/10 px-3 py-1 text-xs font-semibold uppercase text-primary/75">
                {selected.source || "footer"}
              </span>
            </div>

            <div className="mt-6 grid gap-3 md:grid-cols-2">
              <Info label={t("panelContactSubmissions.name")} value={selected.full_name || "-"} />
              <Info
                label={t("panelContactSubmissions.phone")}
                value={
                  selected.phone_e164 ? (
                    <a href={`tel:${selected.phone_e164}`} dir="ltr" className="transition hover:text-accent">
                      {selected.phone_e164}
                    </a>
                  ) : "-"
                }
              />
              <Info
                label={t("panelContactSubmissions.email")}
                value={
                  selected.email ? (
                    <a href={`mailto:${selected.email}`} className="transition hover:text-accent">
                      {selected.email}
                    </a>
                  ) : "-"
                }
              />
              <Info label={t("panelContactSubmissions.country")} value={`${selected.country_iso || "-"} ${selected.country_code || ""}`} />
            </div>

            <div className="mt-6 rounded-2xl border border-primary/10 bg-white/75 p-5">
              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-primary/55">
                {t("panelContactSubmissions.message")}
              </p>
              <p className="mt-3 whitespace-pre-wrap text-sm leading-7 text-primary/80">{selected.message || "-"}</p>
            </div>
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
