import { useTranslation } from "../lib/i18n";

export default function Dashboard() {
  const { t } = useTranslation();

  return (
    <div className="grid gap-6 md:grid-cols-3">
      <div className="panel-card md:col-span-3">
        <h2 className="font-display text-2xl">{t("dashboard.overviewTitle")}</h2>
        <p className="mt-2 text-sm text-primary/70">{t("dashboard.overviewSubtitle")}</p>
      </div>
      <div className="panel-card">
        <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("panelProducts.title")}</p>
        <p className="mt-3 text-sm text-primary/60">{t("messages.empty")}</p>
      </div>
      <div className="panel-card">
        <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("categories.title")}</p>
        <p className="mt-3 text-sm text-primary/60">{t("messages.empty")}</p>
      </div>
      <div className="panel-card">
        <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("panelBlogs.title")}</p>
        <p className="mt-3 text-sm text-primary/60">{t("messages.empty")}</p>
      </div>
    </div>
  );
}
