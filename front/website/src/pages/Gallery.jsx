import { useTranslation } from "../lib/i18n";
import { usePageSeo } from "../lib/seo";

export default function Gallery() {
  const { t, lang } = useTranslation();

  usePageSeo({
    title: `${t("nav.gallery")} | SangeHassan`,
    description: t("sections.comingSoon"),
    path: "/gallery",
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    robots: "noindex,follow"
  });

  return (
    <section className="section-shell py-16">
      <div className="glass-panel rounded-3xl p-10">
        <h1 className="font-display text-3xl">{t("nav.gallery")}</h1>
        <p className="mt-3 text-sm text-primary/70">{t("sections.comingSoon")}</p>
      </div>
    </section>
  );
}
