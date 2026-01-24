import { useTranslation } from "../lib/i18n";

export default function Footer() {
  const { t } = useTranslation();

  return (
    <footer className="mt-24 bg-primary text-sand">
      <div className="section-shell flex min-h-[40vh] flex-col justify-between gap-10 py-16">
        <div className="flex flex-col gap-4">
          <p className="font-display text-2xl">{t("footer.title")}</p>
          <p className="max-w-2xl text-sand/80">{t("footer.subtitle")}</p>
        </div>

        <div className="grid gap-8 md:grid-cols-3">
          <div className="space-y-2 text-sm">
            <p className="font-semibold uppercase tracking-wide">{t("footer.addressLabel")}</p>
            <p>{t("footer.addressValue")}</p>
            <p className="font-semibold uppercase tracking-wide">{t("footer.phoneLabel")}</p>
            <p>{t("footer.phoneValue")}</p>
            <p className="font-semibold uppercase tracking-wide">{t("footer.whatsappLabel")}</p>
            <p>{t("footer.whatsappValue")}</p>
          </div>

          <div className="space-y-2 text-sm">
            <p className="font-semibold uppercase tracking-wide">{t("footer.linkedinLabel")}</p>
            <p>{t("footer.linkedinValue")}</p>
            <p className="font-semibold uppercase tracking-wide">{t("footer.emailLabel")}</p>
            <p>{t("footer.emailValue")}</p>
          </div>

          <div className="space-y-3 text-sm">
            <p className="font-semibold uppercase tracking-wide">{t("footer.aboutTitle")}</p>
            <p className="text-sand/80">{t("footer.aboutText")}</p>
          </div>
        </div>
      </div>
    </footer>
  );
}
