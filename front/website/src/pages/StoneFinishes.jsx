import { useTranslation } from "../lib/i18n";

export default function StoneFinishes() {
  const { t } = useTranslation();

  return (
    <section className="section-shell py-16">
      <div className="glass-panel rounded-3xl p-10">
        <h1 className="font-display text-3xl">{t("nav.stoneFinishes")}</h1>
        <p className="mt-3 text-sm text-primary/70">{t("sections.comingSoon")}</p>
      </div>
    </section>
  );
}
