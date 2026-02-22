import { Link } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import enDict from "@shared/i18n/en.json";
import faDict from "@shared/i18n/fa.json";
import arDict from "@shared/i18n/ar.json";

const PALETTE = {
  primary: "#083A4F",
  accent: "#A58D66",
  bg: "#E5E1DD",
};

const ABOUT_COPY = {
  en: enDict.about,
  fa: faDict.about,
  ar: arDict.about,
};

function SectionTitle({ children, subtitle }) {
  return (
    <div className="mb-7 md:mb-8">
      <div className="inline-flex items-center gap-3 rounded-full border border-[--accent]/20 bg-white/45 px-3 py-1.5">
        <span className="h-[2px] w-8 rounded-full bg-[--accent]" />
        <h2 className="font-display text-[22px] md:text-[28px] font-semibold tracking-tight text-[--accent]">
          {children}
        </h2>
      </div>
      {subtitle ? (
        <p className="mt-3 text-[15px] md:text-[16px] leading-8 text-[--primary]/85">
          {subtitle}
        </p>
      ) : null}
    </div>
  );
}

function GlassCard({ title, badge, children }) {
  return (
    <article
      className="
        group relative overflow-hidden rounded-2xl border border-[--primary]/10
        bg-white/60 backdrop-blur-md shadow-[0_10px_30px_rgba(8,58,79,0.08)]
        transition-all duration-300 hover:-translate-y-0.5 hover:shadow-[0_14px_34px_rgba(8,58,79,0.10)]
      "
    >
      <div className="pointer-events-none absolute -top-10 -left-10 h-32 w-32 rounded-full bg-[--accent]/10 blur-3xl opacity-70" />
      <div className="pointer-events-none absolute -bottom-12 -right-12 h-40 w-40 rounded-full bg-[--primary]/10 blur-3xl opacity-60" />

      {badge ? (
        <span className="absolute top-4 right-4 inline-flex h-7 min-w-7 items-center justify-center rounded-full border border-[--accent]/35 bg-[--accent]/10 px-2 text-xs font-semibold text-[--accent]">
          {badge}
        </span>
      ) : null}

      <div className="p-6 md:p-7">
        <div className="min-w-0">
          <h3 className="pr-10 text-[15px] font-semibold text-[--primary]">
            {title}
          </h3>
          <div className="mt-3 text-[15px] md:text-[16px] leading-8 text-[--primary]/85">
            {children}
          </div>
        </div>

        <div className="mt-5 h-[1px] w-full bg-gradient-to-l from-transparent via-[--accent]/35 to-transparent opacity-70" />
      </div>
    </article>
  );
}

export default function About() {
  const { t, lang } = useTranslation();
  const about = ABOUT_COPY[lang] || ABOUT_COPY.en;
  const heroBulletsRaw = about?.hero?.bullets;
  const heroBullets = Array.isArray(heroBulletsRaw) ? heroBulletsRaw : [];
  const serviceCardsRaw = about?.services?.cards;
  const serviceCards = Array.isArray(serviceCardsRaw) ? serviceCardsRaw : [];

  return (
    <main
      key={lang}
      className="relative"
      style={{
        "--primary": PALETTE.primary,
        "--accent": PALETTE.accent,
        "--bg": PALETTE.bg,
        backgroundColor: PALETTE.bg,
        color: PALETTE.primary,
      }}
    >
      <div
        className="pointer-events-none fixed inset-0 opacity-[0.14]"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(8,58,79,0.06) 1px, transparent 1px), linear-gradient(to bottom, rgba(8,58,79,0.06) 1px, transparent 1px)",
          backgroundSize: "48px 48px",
        }}
      />

      <div
        className="pointer-events-none fixed inset-0 opacity-[0.35]"
        style={{
          backgroundImage:
            "radial-gradient(circle at 12% 18%, rgba(165,141,102,0.18), transparent 48%), radial-gradient(circle at 88% 80%, rgba(8,58,79,0.12), transparent 52%)",
        }}
      />

      <div className="relative mx-auto max-w-6xl px-4 sm:px-6 lg:px-8 pt-8 pb-16 md:pt-12 md:pb-20">
        <nav className="mb-4 flex items-center gap-1.5 text-xs md:text-sm text-[--primary]/55">
          <Link to="/" className="hover:text-[--primary] transition-colors">
            {t("nav.home")}
          </Link>
          <span className="opacity-45">›</span>
          <span className="font-semibold text-[--primary]">{t("nav.about")}</span>
        </nav>

        <section
          className="
            relative overflow-hidden rounded-3xl border border-[--primary]/10
            bg-white/55 backdrop-blur-md
            shadow-[0_26px_72px_rgba(8,58,79,0.16)]
          "
        >
          <div
            className="pointer-events-none absolute inset-0 opacity-100"
            style={{
              backgroundImage:
                "linear-gradient(135deg, rgba(165,141,102,0.18) 0%, rgba(165,141,102,0.10) 18%, rgba(255,255,255,0.0) 45%)",
            }}
          />

          <div className="pointer-events-none absolute -top-24 -right-24 h-72 w-72 rounded-full bg-[--primary]/10 blur-3xl" />
          <div className="pointer-events-none absolute -bottom-28 -left-24 h-80 w-80 rounded-full bg-[--accent]/18 blur-3xl" />

          <div className="relative grid grid-cols-1 lg:grid-cols-12">
            <div className="lg:col-span-7 p-6 sm:p-8 md:p-10 lg:p-11">
              <div className="flex items-start gap-4">
                <div className="hidden sm:block mt-2 h-16 w-[3px] rounded-full bg-[--accent]" />
                <div>
                  <h1 className="font-display text-4xl md:text-5xl font-bold tracking-tight text-[--accent]">
                    {about?.hero?.title}
                  </h1>

                  <p className="mt-6 text-[15px] md:text-[16px] leading-8 text-[--primary]/85">
                    {about?.hero?.body}
                  </p>

                  <div className="mt-6 flex flex-wrap gap-3">
                    <Link
                      to="/products"
                      className="
                        inline-flex items-center justify-center rounded-xl px-5 py-3
                        bg-[--primary] text-[--bg] font-semibold text-sm
                        shadow-[0_18px_36px_rgba(8,58,79,0.24)]
                        transition hover:translate-y-[-1px] hover:opacity-95 hover:shadow-[0_22px_40px_rgba(8,58,79,0.28)]
                      "
                    >
                      {about?.hero?.ctaPrimary}
                    </Link>
                  </div>
                </div>
              </div>
            </div>

            <div className="lg:col-span-5 border-t border-[--primary]/10 bg-white/38 lg:border-t-0 lg:border-r">
              <div className="relative h-full p-6 sm:p-8 md:p-10 lg:p-11">
                <div className="pointer-events-none absolute inset-0 opacity-70">
                  <div className="absolute -top-10 -left-10 h-28 w-28 rounded-full bg-[--accent]/16 blur-3xl" />
                  <div className="absolute -bottom-10 -right-10 h-32 w-32 rounded-full bg-[--primary]/12 blur-3xl" />
                </div>

                <div className="relative">
                  <p className="text-[15px] md:text-[16px] leading-8 text-[--primary] font-semibold">
                    {about?.hero?.quality}
                  </p>

                  <ul className="mt-6 space-y-4 text-[15px] md:text-[16px] leading-8 text-[--primary]/85">
                    {heroBullets.map((item) => (
                      <li key={item} className="flex gap-3">
                        <span className="mt-3 h-1.5 w-1.5 rounded-full bg-[--accent]" />
                        {item}
                      </li>
                    ))}
                  </ul>

                  <div className="mt-7 h-[1px] w-full bg-gradient-to-l from-transparent via-[--accent]/40 to-transparent" />
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="mt-16 md:mt-20">
          <SectionTitle subtitle={about?.services?.subtitle}>
            {about?.services?.title}
          </SectionTitle>

          <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
            {serviceCards.map((card, index) => {
              const isLastOdd = index === serviceCards.length - 1 && serviceCards.length % 2 === 1;
              return (
                <div key={`${card.badge}-${card.title}`} className={isLastOdd ? "md:col-span-2" : undefined}>
                  <GlassCard title={card.title} badge={card.badge}>
                    {card.body}
                  </GlassCard>
                </div>
              );
            })}
          </div>
        </section>

        <section className="mt-16 md:mt-20">
          <SectionTitle>{about?.mission?.title}</SectionTitle>

          <div className="relative overflow-hidden rounded-2xl border border-[--primary]/10 bg-white/60 backdrop-blur-md p-6 md:p-7 shadow-[0_10px_30px_rgba(8,58,79,0.08)]">
            <div className="pointer-events-none absolute inset-0">
              <div className="absolute -bottom-16 -left-16 h-64 w-64 rounded-full bg-[--primary]/10 blur-3xl" />
              <div className="absolute -top-10 -right-10 h-56 w-56 rounded-full bg-[--accent]/14 blur-3xl" />
            </div>

            <p className="relative text-[15px] md:text-[16px] leading-8 text-[--primary]/85">
              {about?.mission?.before}{" "}
              <strong className="font-semibold text-[--primary]">{about?.mission?.emphasis}</strong>{" "}
              {about?.mission?.after}
            </p>
          </div>
        </section>
      </div>
    </main>
  );
}
