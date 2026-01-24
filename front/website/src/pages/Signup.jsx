import { useTranslation } from "../lib/i18n";

export default function Signup() {
  const { t } = useTranslation();

  return (
    <section className="section-shell py-16">
      <div className="max-w-xl rounded-3xl bg-white/80 p-8 shadow-xl">
        <h1 className="font-display text-3xl">{t("auth.signupTitle")}</h1>
        <p className="mt-2 text-sm text-primary/70">{t("auth.signupSubtitle")}</p>

        <form className="mt-6 space-y-4">
          <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("auth.fullName")}
            <input
              type="text"
              className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
              placeholder={t("auth.fullName")}
            />
          </label>
          <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("auth.email")}
            <input
              type="email"
              className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
              placeholder={t("auth.email")}
            />
          </label>
          <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("auth.password")}
            <input
              type="password"
              className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
              placeholder={t("auth.password")}
            />
          </label>
          <button
            type="button"
            className="w-full rounded-full bg-primary px-6 py-3 text-sm font-semibold text-sand"
          >
            {t("auth.signupButton")}
          </button>
        </form>
      </div>
    </section>
  );
}
