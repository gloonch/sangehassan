import { Link } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import ProductMarquee from "../components/ProductMarquee";

export default function Home() {
  const { t } = useTranslation();
  const popularItems = t("home.popularItems");

  return (
    <div className="bg-quiet-grid">
      <section>
        <div className="section-shell flex h-[70vh] min-h-[520px] items-center">
          <div className="grid gap-6 md:grid-cols-[1.2fr_0.8fr]">
            <div className="space-y-6">
              <p className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("hero.kicker")}</p>
              <h1 className="font-display text-4xl leading-tight md:text-5xl">
                {t("hero.title")}
              </h1>
              <p className="max-w-xl text-lg text-primary/70">{t("hero.subtitle")}</p>
              <div className="flex flex-wrap gap-3">
                <Link
                  to="/products"
                  className="rounded-full bg-primary px-6 py-3 text-sm font-semibold text-sand transition hover:bg-primary/90"
                >
                  {t("hero.ctaPrimary")}
                </Link>
                <Link
                  to="/contact"
                  className="rounded-full border border-primary/30 px-6 py-3 text-sm font-semibold text-primary/80 transition hover:border-primary/60"
                >
                  {t("hero.ctaSecondary")}
                </Link>
              </div>
              <p className="text-xs uppercase tracking-[0.3em] text-primary/60">{t("hero.scrollHint")}</p>
            </div>
            <div className="relative hidden h-full md:block">
              <div className="absolute right-0 top-8 h-56 w-56 rounded-[32px] bg-accent/20" />
              <div className="absolute bottom-8 left-0 h-72 w-60 rounded-[32px] bg-primary/10" />
              <div className="glass-panel absolute inset-6 rounded-[36px]" />
            </div>
          </div>
        </div>
      </section>

      <section className="section-shell pb-16 pt-8">
        <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("home.popularTitle")}</p>
            <h2 className="mt-2 font-display text-3xl">{t("home.popularSubtitle")}</h2>
          </div>
          <Link to="/products" className="text-sm font-semibold text-accent">
            {t("home.viewAll")}
          </Link>
        </div>
        <ProductMarquee items={Array.isArray(popularItems) ? popularItems : []} />
      </section>
    </div>
  );
}
