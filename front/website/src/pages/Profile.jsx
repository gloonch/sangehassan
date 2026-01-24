import { useTranslation } from "../lib/i18n";

export default function Profile() {
  const { t } = useTranslation();

  return (
    <section className="section-shell py-16">
      <div className="glass-panel rounded-3xl p-10">
        <h1 className="font-display text-3xl">{t("profile.title")}</h1>
        <p className="mt-2 text-sm text-primary/70">{t("profile.subtitle")}</p>
      </div>
    </section>
  );
}
