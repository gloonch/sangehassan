import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

export default function Dashboard() {
  const { t } = useTranslation();
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      try {
        const res = await fetchJSON("/api/admin/dashboard");
        if (!mounted) return;
        setStats(res.data || null);
        setError("");
      } catch (err) {
        if (!mounted) return;
        setStats(null);
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

  const templates = useMemo(() => stats?.templates || [], [stats]);
  const categoryUsage = useMemo(() => stats?.category_usage || [], [stats]);

  return (
    <div className="space-y-6">
      <div className="panel-card">
        <h2 className="font-display text-2xl">{t("dashboard.overviewTitle")}</h2>
        <p className="mt-2 text-sm text-primary/70">{t("dashboard.overviewSubtitle")}</p>
        {loading ? (
          <p className="mt-4 text-sm text-primary/60">{t("messages.loading")}</p>
        ) : error ? (
          <p className="mt-4 text-sm text-red-500">{error}</p>
        ) : (
          <div className="mt-6 grid gap-4 md:grid-cols-4">
            <div className="rounded-2xl border border-primary/10 bg-white/70 px-4 py-3">
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("dashboard.productsCount")}</p>
              <p className="mt-2 text-2xl font-semibold text-primary">{stats?.products_count ?? 0}</p>
            </div>
            <div className="rounded-2xl border border-primary/10 bg-white/70 px-4 py-3">
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("dashboard.categoriesCount")}</p>
              <p className="mt-2 text-2xl font-semibold text-primary">{stats?.categories_count ?? 0}</p>
            </div>
            <div className="rounded-2xl border border-primary/10 bg-white/70 px-4 py-3">
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("dashboard.templatesCount")}</p>
              <p className="mt-2 text-2xl font-semibold text-primary">{stats?.templates_count ?? 0}</p>
            </div>
            <div className="rounded-2xl border border-primary/10 bg-white/70 px-4 py-3">
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("dashboard.blogsCount")}</p>
              <p className="mt-2 text-2xl font-semibold text-primary">{stats?.blogs_count ?? 0}</p>
            </div>
          </div>
        )}
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.4fr_1fr]">
        <div className="panel-card">
          <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("dashboard.categoryUsageTitle")}</p>
          {loading ? (
            <p className="mt-4 text-sm text-primary/60">{t("messages.loading")}</p>
          ) : categoryUsage.length === 0 ? (
            <p className="mt-4 text-sm text-primary/60">{t("messages.empty")}</p>
          ) : (
            <div className="mt-4 space-y-3">
              {categoryUsage.map((category) => (
                <div
                  key={category.id}
                  className="flex items-center justify-between rounded-xl border border-primary/10 bg-white/70 px-4 py-3"
                >
                  <div>
                    <p className="text-sm font-semibold text-primary">{category.title_en}</p>
                    <p className="text-xs text-primary/60">{category.title_fa} • {category.title_ar}</p>
                  </div>
                  <span className="rounded-full bg-accent/15 px-3 py-1 text-xs font-semibold text-accent">
                    {category.product_count}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="panel-card">
          <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("dashboard.templatesTitle")}</p>
          {loading ? (
            <p className="mt-4 text-sm text-primary/60">{t("messages.loading")}</p>
          ) : templates.length === 0 ? (
            <p className="mt-4 text-sm text-primary/60">{t("messages.empty")}</p>
          ) : (
            <div className="mt-4 space-y-3">
              {templates.map((template) => (
                <div
                  key={template.id}
                  className="flex items-center gap-3 rounded-xl border border-primary/10 bg-white/70 px-4 py-3"
                >
                  <div className="h-12 w-12 overflow-hidden rounded-lg border border-primary/10 bg-primary/5">
                    {template.image_url ? (
                      <img
                        src={resolveImageUrl(template.image_url)}
                        alt={template.name}
                        className="h-full w-full object-cover"
                      />
                    ) : null}
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-primary">{template.name}</p>
                    <p className="text-xs text-primary/60">
                      {template.is_active ? t("form.active") : t("form.inactive")}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
