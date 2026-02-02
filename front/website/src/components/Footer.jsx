import { useTranslation } from "../lib/i18n";

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

const EmailIcon = ({ className }) => (
  <svg {...iconBaseProps} className={className} aria-hidden="true">
    <rect x="2" y="4" width="20" height="16" rx="2" />
    <path d="M2 6l10 7l10-7" />
  </svg>
);

const getContactHref = (key, value) => {
  if (!value || typeof value !== "string") return "";
  if (key === "email") return `mailto:${value}`;
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
  const { t } = useTranslation();
  const addressLinesRaw = t("footer.addressLines");
  const addressLines = Array.isArray(addressLinesRaw)
    ? addressLinesRaw
    : [t("footer.addressValue")].filter(Boolean);
  const socialItems = [
    {
      key: "linkedin",
      label: t("footer.linkedinLabel"),
      value: t("footer.linkedinValue"),
      Icon: LinkedInIcon
    },
    {
      key: "telegram",
      label: t("footer.telegramLabel"),
      value: t("footer.telegramValue"),
      Icon: TelegramIcon
    },
    {
      key: "instagram",
      label: t("footer.instagramLabel"),
      value: t("footer.instagramValue"),
      Icon: InstagramIcon
    },
    {
      key: "email",
      label: t("footer.emailLabel"),
      value: t("footer.emailValue"),
      Icon: EmailIcon
    }
  ].filter((item) => item.value);
  const orderedSocialItems = [...socialItems].sort(
    (a, b) => a.value.length - b.value.length || a.label.localeCompare(b.label)
  );

  return (
    <footer className="mt-24 bg-primary text-sand">
      <div className="section-shell flex min-h-[40vh] flex-col justify-between gap-10 py-16">
        <div className="flex flex-col gap-4">
          <p className="font-display text-2xl">{t("footer.title")}</p>
          <p className="max-w-2xl text-sand/80">{t("footer.subtitle")}</p>
        </div>

        <div className="grid gap-8 md:grid-cols-2 justify-items-center">
          <div className="space-y-2 text-sm">
            <p className="font-semibold uppercase tracking-wide">{t("footer.addressLabel")}</p>
            <ul className="space-y-1 text-sand/80">
              {addressLines.map((line, index) => (
                <li key={`${index}-${line}`}>{line}</li>
              ))}
            </ul>
            <p className="font-semibold uppercase tracking-wide">{t("footer.phoneLabel")}</p>
            <p>{t("footer.phoneValue")}</p>
            <p className="font-semibold uppercase tracking-wide">{t("footer.whatsappLabel")}</p>
            <p>{t("footer.whatsappValue")}</p>
          </div>

          <div className="space-y-4 text-sm">
            {orderedSocialItems.map(({ key, label, value, Icon }) => {
              const href = getContactHref(key, value);
              const isExternal = href.startsWith("http");
              return (
                <div key={key} className="flex items-start gap-3">
                  <Icon className="mt-1 h-4 w-4 text-sand/70" />
                  <div className="space-y-1">
                    <p className="font-semibold uppercase tracking-wide">{label}</p>
                    {href ? (
                      <a
                        href={href}
                        className="text-sand/80 transition hover:text-sand break-words"
                        dir="ltr"
                        {...(isExternal ? { target: "_blank", rel: "noreferrer" } : {})}
                      >
                        {value}
                      </a>
                    ) : (
                      <p className="text-sand/80 break-words" dir="ltr">
                        {value}
                      </p>
                    )}
                  </div>
                </div>
              );
            })}
          </div>


        </div>
      </div>
    </footer>
  );
}
