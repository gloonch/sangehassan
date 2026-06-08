import { useTranslation } from "../lib/i18n";
import logoWhiteImage from "@shared/assets/logo_white.png";

const iconBaseProps = {
  xmlns: "http://www.w3.org/2000/svg",
  fill: "none",
  viewBox: "0 0 24 24",
  strokeWidth: 1.6,
  stroke: "currentColor",
  strokeLinecap: "round",
  strokeLinejoin: "round"
};

const LinkedInIcon = ({ className }) => (
  <svg {...iconBaseProps} className={className} aria-hidden="true">
    <rect x="2" y="9" width="4" height="12" />
    <circle cx="4" cy="4" r="2" />
    <path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6z" />
  </svg>
);

const TelegramIcon = ({ className }) => (
  <svg {...iconBaseProps} className={className} aria-hidden="true">
    <path d="M22 2L11 13" />
    <path d="M22 2L15 22L11 13L2 9L22 2Z" />
  </svg>
);

const InstagramIcon = ({ className }) => (
  <svg {...iconBaseProps} className={className} aria-hidden="true">
    <rect x="2" y="2" width="20" height="20" rx="5" ry="5" />
    <circle cx="12" cy="12" r="3.5" />
    <circle cx="17.5" cy="6.5" r="0.5" />
  </svg>
);

const PhoneIcon = ({ className }) => (
  <svg {...iconBaseProps} className={className} aria-hidden="true">
    <path d="M6.5 3h4l2 5l-2.5 2.5a16 16 0 0 0 3.5 3.5L16 11.5l5 2v4a2 2 0 0 1-2 2C10.7 19.5 4.5 13.3 4.5 5a2 2 0 0 1 2-2z" />
  </svg>
);

const toLatinDigits = (value) =>
  value
    .replace(/[۰-۹]/g, (digit) => String("۰۱۲۳۴۵۶۷۸۹".indexOf(digit)))
    .replace(/[٠-٩]/g, (digit) => String("٠١٢٣٤٥٦٧٨٩".indexOf(digit)));

const normalizePhone = (value) => {
  if (!value || typeof value !== "string") return "";
  const cleaned = toLatinDigits(value).replace(/[^\d+]/g, "");
  if (!cleaned) return "";
  if (cleaned.startsWith("+")) {
    return `+${cleaned.slice(1).replace(/\+/g, "")}`;
  }
  return cleaned.replace(/\+/g, "");
};

const getContactHref = (key, value) => {
  if (!value || typeof value !== "string") return "";
  if (key === "phone") {
    const normalized = value.startsWith("tel:") ? value.slice(4) : normalizePhone(value);
    return normalized ? `tel:${normalized}` : "";
  }
  if (value.startsWith("http://") || value.startsWith("https://")) return value;
  const lower = value.toLowerCase();
  if (
    lower.includes("linkedin.com") ||
    lower.includes("instagram.com") ||
    lower.includes("t.me") ||
    lower.includes("telegram.me")
  ) {
    return `https://${value.replace(/^https?:\/\//i, "")}`;
  }
  return "";
};

export default function Footer() {
  const { t, lang } = useTranslation();
  const isRTL = lang === "fa" || lang === "ar";

  const socialItems = [
    {
      key: "linkedin-main",
      label: t("footer.linkedinLabel"),
      value: t("footer.linkedinValue"),
      Icon: LinkedInIcon,
      hrefKey: "linkedin"
    },
    {
      key: "instagram",
      label: t("footer.instagramLabel"),
      value: t("footer.instagramValue"),
      Icon: InstagramIcon,
      hrefKey: "instagram"
    },
    {
      key: "telegram",
      label: t("footer.telegramLabel"),
      value: t("footer.telegramValue"),
      Icon: TelegramIcon,
      hrefKey: "telegram"
    },
    {
      key: "phone",
      label: t("footer.phoneLabel"),
      value: t("footer.phoneValue"),
      Icon: PhoneIcon,
      hrefKey: "phone"
    }
  ]
    .map((item) => ({ ...item, href: getContactHref(item.hrefKey, item.value) }))
    .filter((item) => item.href);

  const firstColumnAlign = isRTL ? "md:items-end md:text-right" : "md:items-start md:text-left";
  const lastColumnAlign = isRTL ? "md:items-start md:text-left" : "md:items-end md:text-right";
  const socialJustify = isRTL ? "justify-center md:justify-end" : "justify-center md:justify-start";

  return (
    <footer className="mt-12 overflow-hidden bg-primary text-sand md:mt-24">
      <div className="section-shell py-5 md:py-13">
        <div className="grid grid-cols-1 items-center gap-4 md:grid-cols-[1.2fr_1fr_1.2fr] md:gap-10">
          <div className={`flex flex-col items-center gap-2.5 md:gap-4 ${firstColumnAlign}`}>
            <p className="text-xs tracking-[0.04em] text-sand/82 md:text-sm">
              <span className="font-semibold text-sand">Stay Connected</span>
            </p>
            <div className={`flex w-full flex-wrap items-center gap-1.5 md:gap-2 ${socialJustify}`}>
              {socialItems.map(({ key, label, href, Icon }, index) => {
                const isExternal = href.startsWith("http");
                return (
                  <span key={key} className="inline-flex items-center gap-1.5 md:gap-2">
                    <a
                      href={href}
                      className="inline-flex h-7 w-7 items-center justify-center text-sand/80 transition hover:text-sand md:h-8 md:w-8"
                      aria-label={label}
                      {...(isExternal ? { target: "_blank", rel: "noreferrer" } : {})}
                    >
                      <Icon className="h-3.5 w-3.5 md:h-4 md:w-4" />
                    </a>
                    {index < socialItems.length - 1 ? <span className="text-sand/35">|</span> : null}
                  </span>
                );
              })}
            </div>
          </div>

          <div className="relative h-14 overflow-visible md:h-40 lg:h-48">
            <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
              <img
                src={logoWhiteImage}
                alt=""
                aria-hidden="true"
                className="h-auto w-[120%] max-w-none opacity-[0.14] md:w-[200%] lg:w-[225%]"
              />
            </div>
          </div>

          <div className={`flex flex-col items-center gap-2.5 md:gap-4 ${lastColumnAlign}`}>
            <p className="font-display text-lg leading-tight text-sand/92 md:text-2xl">Let&apos;s build in stone</p>
          </div>
        </div>
      </div>
    </footer>
  );
}
