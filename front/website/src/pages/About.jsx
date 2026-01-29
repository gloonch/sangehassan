import { Link } from "react-router-dom";

const PALETTE = {
  primary: "#083A4F",
  accent: "#A58D66",
  bg: "#E5E1DD",
};

function SectionTitle({ children, subtitle }) {
  return (
    <div className="mb-5">
      <div className="flex items-center gap-3">
        <span className="h-[2px] w-7 rounded-full bg-[--accent]" />
        <h2 className="font-display text-xl md:text-2xl font-semibold tracking-tight text-[--accent]">
          {children}
        </h2>
      </div>
      {subtitle ? (
        <p className="mt-2 text-sm md:text-[15px] leading-relaxed text-[--primary]/80">
          {subtitle}
        </p>
      ) : null}
    </div>
  );
}

function GlassCard({ title, children }) {
  return (
    <article
      className="
        group relative overflow-hidden rounded-2xl border border-[--primary]/10
        bg-white/55 backdrop-blur-md shadow-[0_18px_40px_rgba(8,58,79,0.10)]
        transition-transform duration-300 hover:-translate-y-0.5
      "
    >
      <div className="pointer-events-none absolute -top-10 -left-10 h-32 w-32 rounded-full bg-[--accent]/15 blur-2xl opacity-70" />
      <div className="pointer-events-none absolute -bottom-12 -right-12 h-40 w-40 rounded-full bg-[--primary]/10 blur-2xl opacity-60" />

      <div className="p-5">
        <div className="min-w-0">
          <h3 className="text-[15px] md:text-base font-semibold text-[--primary]">
            {title}
          </h3>
          <div className="mt-2 text-sm md:text-[15px] leading-relaxed text-[--primary]/85">
            {children}
          </div>
        </div>

        <div className="mt-4 h-[1px] w-full bg-gradient-to-l from-transparent via-[--accent]/40 to-transparent opacity-70" />
      </div>
    </article>
  );
}

export default function About() {
  return (
    <main
      className="min-h-screen"
      style={{
        "--primary": PALETTE.primary,
        "--accent": PALETTE.accent,
        "--bg": PALETTE.bg,
        backgroundColor: PALETTE.bg,
        color: PALETTE.primary,
      }}
    >
      <div
        className="pointer-events-none fixed inset-0 opacity-[0.18]"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(8,58,79,0.06) 1px, transparent 1px), linear-gradient(to bottom, rgba(8,58,79,0.06) 1px, transparent 1px)",
          backgroundSize: "44px 44px",
        }}
      />

      <div className="relative mx-auto max-w-6xl px-4 sm:px-6 lg:px-8 py-10 md:py-14">
        <nav className="text-sm text-[--primary]/70 mb-6 flex items-center gap-2">
          <Link to="/" className="hover:text-[--primary] transition-colors">
            خانه
          </Link>
          <span className="opacity-50">›</span>
          <span className="font-semibold text-[--primary]">درباره ما</span>
        </nav>

        <section
          className="
            relative overflow-hidden rounded-3xl border border-[--primary]/10
            bg-white/55 backdrop-blur-md
            shadow-[0_22px_60px_rgba(8,58,79,0.12)]
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

          <div className="relative grid grid-cols-1 lg:grid-cols-12 gap-8 p-6 sm:p-8 md:p-10">
            <div className="lg:col-span-7">
              <div className="flex items-start gap-4">
                <div className="hidden sm:block mt-2 h-16 w-[3px] rounded-full bg-[--accent]" />
                <div>
                  <h1 className="font-display text-3xl md:text-4xl font-bold tracking-tight text-[--accent]">
                    نگاه ما
                  </h1>

                  <p className="mt-5 text-sm md:text-[15.5px] leading-relaxed text-[--primary]/90">
                    مجموعه ما با تمرکز بر تأمین و آماده‌سازی تخصصی سنگ‌های طبیعی در صنعت ساختمان فعالیت می‌کند.
                    باور ما این است که انتخاب سنگ مناسب برای هر پروژه، یک تصمیم مهم است؛ برای همین تلاش می‌کنیم مسیر
                    انتخاب، فرآوری و تحویل را برای مشتریان شفاف، مطمئن و ساده کنیم.
                  </p>

                  <div className="mt-6 flex flex-wrap gap-3">
                    <Link
                      to="/products"
                      className="
                        inline-flex items-center justify-center rounded-xl px-4 py-2.5
                        bg-[--primary] text-[--bg] font-semibold text-sm
                        shadow-[0_18px_35px_rgba(8,58,79,0.25)]
                        hover:opacity-95 transition
                      "
                    >
                      مشاهده محصولات
                    </Link>
                  </div>
                </div>
              </div>
            </div>

            <div className="lg:col-span-5">
              <div
                className="
                  relative h-full overflow-hidden rounded-2xl border border-[--primary]/10
                  bg-white/40 backdrop-blur-md p-6
                "
              >
                <div className="pointer-events-none absolute inset-0 opacity-70">
                  <div className="absolute -top-10 -left-10 h-28 w-28 rounded-full bg-[--accent]/20 blur-2xl" />
                  <div className="absolute -bottom-10 -right-10 h-32 w-32 rounded-full bg-[--primary]/10 blur-2xl" />
                </div>

                <div className="relative">

                  <p className="mt-3 text-[15px] leading-relaxed text-[--primary] font-semibold">
                    کیفیت فقط «سنگ خوب» نیست، کیفیت یعنی انتخاب درست، پرداخت مناسب، تحویل دقیق، و پاسخ‌گویی بعد از خرید.
                  </p>

                  <ul className="mt-5 space-y-3 text-sm text-[--primary]/85">
                    <li className="flex gap-3">
                      <span className="mt-1.5 h-1.5 w-1.5 rounded-full bg-[--accent]" />
                      مشاوره برای انتخاب سنگ متناسب با کاربرد و فضا
                    </li>
                    <li className="flex gap-3">
                      <span className="mt-1.5 h-1.5 w-1.5 rounded-full bg-[--accent]" />
                      پرداخت و برش استاندارد برای اجرای تمیز و بادوام
                    </li>
                    <li className="flex gap-3">
                      <span className="mt-1.5 h-1.5 w-1.5 rounded-full bg-[--accent]" />
                      تحویل ایمن با بسته‌بندی حرفه‌ای و زمان‌بندی دقیق
                    </li>
                  </ul>

                  <div className="mt-6 h-[1px] w-full bg-gradient-to-l from-transparent via-[--accent]/45 to-transparent" />

                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="mt-10 md:mt-14">
          <SectionTitle subtitle="چند سرویس کلیدی که مسیر انتخاب تا تحویل را برای پروژه‌ها ساده‌تر می‌کند.">
            خدمات ما
          </SectionTitle>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <GlassCard title="۱) مشاوره تخصصی">
              تیم کارشناسی ما با توجه به نیاز پروژه، سلیقه و بودجه شما، گزینه‌های مناسب را پیشنهاد می‌دهد تا انتخاب نهایی دقیق‌تر و کم‌ریسک‌تر باشد.
            </GlassCard>

            <GlassCard title="۲) برش و سفارشی‌سازی">
              با تجهیزات فرآوری، امکان آماده‌سازی سنگ بر اساس ابعاد و طراحی موردنظر فراهم است؛ از برش و پرداخت تا اجرای جزئیات موردنیاز پروژه‌های معماری و دکوراسیون.
            </GlassCard>

            <GlassCard title="۳) حمل‌ونقل سریع و ایمن">
              ارسال محصولات با بسته‌بندی حرفه‌ای و رعایت ایمنی انجام می‌شود تا سنگ‌ها سالم و بدون آسیب به مقصد برسند؛ در مسیرهای داخلی و در صورت نیاز بین‌المللی.
            </GlassCard>

            <GlassCard title="۴) نظارت و پشتیبانی پروژه">
              همراهی کارشناسان از مرحله انتخاب تا اجرای نهایی، کمک می‌کند کیفیت خروجی مطابق استانداردهای مورد انتظار حفظ شود.
            </GlassCard>

            <div className="md:col-span-2">
              <GlassCard title="۵) خدمات پس از فروش">
                ارتباط ما بعد از تحویل تمام نمی‌شود؛ اگر راهنمایی یا پیگیری لازم باشد، برای رفع ابهام‌ها و مسائل احتمالی کنار شما هستیم.
              </GlassCard>
            </div>
          </div>
        </section>

        <section className="mt-10 md:mt-14">
          <SectionTitle>ماموریت ما</SectionTitle>

          <div className="relative overflow-hidden rounded-3xl border border-[--primary]/10 bg-white/55 backdrop-blur-md p-6 md:p-7">
            <div className="pointer-events-none absolute inset-0">
              <div className="absolute -bottom-16 -left-16 h-64 w-64 rounded-full bg-[--primary]/10 blur-3xl" />
              <div className="absolute -top-10 -right-10 h-56 w-56 rounded-full bg-[--accent]/14 blur-3xl" />
            </div>

            <p className="relative text-sm md:text-[15.5px] leading-relaxed text-[--primary]/90">
              تبدیل فرآیند تأمین سنگ به تجربه‌ای{" "}
              <strong className="font-semibold text-[--primary]">آسان، شفاف و رضایت‌بخش</strong>{" "}
              برای همه مشتریان؛ از پروژه‌های کوچک تا پروژه‌های بزرگ ساختمانی.
            </p>

            <div className="relative mt-6 flex flex-wrap gap-3">
              <Link
                to="/stone-finishes"
                className="
                  inline-flex items-center justify-center rounded-xl px-4 py-2.5
                  border border-[--accent]/45 bg-[--accent]/10 text-[--primary]
                  font-semibold text-sm hover:bg-[--accent]/14 transition
                "
              >
                انواع پرداخت و برش سنگ
              </Link>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
