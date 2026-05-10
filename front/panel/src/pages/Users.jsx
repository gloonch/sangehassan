import { useEffect, useState } from "react";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

const formatDateTime = (value) => {
  if (!value) return "-";
  const dt = new Date(value);
  if (Number.isNaN(dt.getTime())) return "-";
  return dt.toLocaleString();
};

export default function Users() {
  const { t } = useTranslation();
  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      try {
        const res = await fetchJSON("/api/admin/users?limit=100");
        if (!mounted) return;
        const data = res.data || {};
        setItems(data.items || []);
        setTotal(data.total || 0);
        setError("");
      } catch (err) {
        if (!mounted) return;
        setItems([]);
        setTotal(0);
        setError(t("messages.error"));
      } finally {
        if (mounted) setLoading(false);
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, [t]);

  return (
    <section className="panel-card">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h2 className="font-display text-xl">{t("panelUsers.title")}</h2>
        <span className="rounded-full bg-accent/10 px-3 py-1 text-xs font-semibold text-accent">
          {total} {t("panelUsers.countLabel")}
        </span>
      </div>

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : error ? (
        <p className="text-sm text-red-500">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-primary/70">{t("panelUsers.empty")}</p>
      ) : (
        <div className="max-h-[720px] space-y-3 overflow-y-auto pr-2">
          {items.map((user) => (
            <div
              key={user.id}
              className="rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
            >
              <p className="text-sm font-semibold text-primary">{user.full_name || "-"}</p>
              <p className="mt-1 text-xs text-primary/70">{t("panelUsers.email")}: {user.email || "-"}</p>
              <p className="mt-1 text-xs text-primary/70">{t("panelUsers.phone")}: {user.phone || "-"}</p>
              <p className="mt-1 text-xs text-primary/70">{t("panelUsers.role")}: {user.role || "-"}</p>
              <p className="mt-1 text-xs text-primary/60">
                {t("panelUsers.registeredAt")}: {formatDateTime(user.created_at)}
              </p>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
