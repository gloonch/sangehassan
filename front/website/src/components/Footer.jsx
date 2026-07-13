import { useMemo, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { getContactHref, getContactPhoneItems } from "../lib/contact";
import { fetchJSON } from "../lib/api";
import { PHONE_COUNTRIES, getPhoneCountry, isValidLocalPhoneNumber, normalizeLocalPhoneNumber } from "../lib/phoneCountries";
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

export default function Footer() {
  const { t, lang } = useTranslation();
  const isRTL = lang === "fa" || lang === "ar";
  const phoneItems = getContactPhoneItems(t("footer.phoneValue"));
  const [contactForm, setContactForm] = useState({
    fullName: "",
    email: "",
    countryISO: "IR",
    phoneNumber: "",
    message: ""
  });
  const [contactStatus, setContactStatus] = useState({ type: "", message: "" });
  const [contactSubmitting, setContactSubmitting] = useState(false);
  const selectedCountry = useMemo(() => getPhoneCountry(contactForm.countryISO), [contactForm.countryISO]);

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
  ]
    .map((item) => ({ ...item, href: getContactHref(item.hrefKey, item.value) }))
    .filter((item) => item.href);

  const firstColumnAlign = isRTL ? "md:items-end md:text-right" : "md:items-start md:text-left";
  const lastColumnAlign = isRTL ? "md:items-start md:text-left" : "md:items-end md:text-right";
  const socialJustify = isRTL ? "justify-center md:justify-end" : "justify-center md:justify-start";
  const updateContactForm = (key, value) => {
    setContactStatus({ type: "", message: "" });
    setContactForm((prev) => ({ ...prev, [key]: value }));
  };

  const handleContactSubmit = async (event) => {
    event.preventDefault();
    const fullName = contactForm.fullName.trim();
    const email = contactForm.email.trim();
    const message = contactForm.message.trim();
    const phoneNumber = normalizeLocalPhoneNumber(contactForm.phoneNumber);

    if (!fullName || !message || !isValidLocalPhoneNumber(contactForm.phoneNumber, selectedCountry)) {
      setContactStatus({ type: "error", message: t("footerContact.validation") });
      return;
    }

    if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setContactStatus({ type: "error", message: t("footerContact.emailInvalid") });
      return;
    }

    setContactSubmitting(true);
    try {
      await fetchJSON("/api/contact-submissions", {
        method: "POST",
        body: JSON.stringify({
          full_name: fullName,
          email,
          country_iso: selectedCountry.iso,
          country_code: selectedCountry.code,
          phone_number: phoneNumber,
          message,
          source: "footer"
        })
      });
      setContactForm({ fullName: "", email: "", countryISO: selectedCountry.iso, phoneNumber: "", message: "" });
      setContactStatus({ type: "success", message: t("footerContact.success") });
    } catch (_) {
      setContactStatus({ type: "error", message: t("footerContact.error") });
    } finally {
      setContactSubmitting(false);
    }
  };

  return (
    <footer className="overflow-hidden bg-primary text-sand">
      <div className="section-shell py-5 md:py-13">
        <div className="grid grid-cols-1 items-center gap-4 md:grid-cols-[1.2fr_1fr_1.2fr] md:gap-10">
          <div className={`flex flex-col items-center gap-2.5 md:gap-4 ${firstColumnAlign}`}>
            <img
              src={logoWhiteImage}
              alt=""
              width="1070"
              height="710"
              aria-hidden="true"
              className="h-auto w-36 opacity-[0.18] md:w-48"
            />
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

          <div className="relative min-h-[13.5rem] overflow-visible md:min-h-[12.5rem]">
            <form
              onSubmit={handleContactSubmit}
              className="relative mx-auto flex w-full max-w-sm flex-col gap-2 rounded-md bg-primary/35 p-3 backdrop-blur-sm"
            >
              <div className="grid grid-cols-2 gap-2">
                <label className="sr-only" htmlFor="footer-contact-name">
                  {t("footerContact.name")}
                </label>
                <input
                  id="footer-contact-name"
                  value={contactForm.fullName}
                  onChange={(event) => updateContactForm("fullName", event.target.value)}
                  placeholder={t("footerContact.name")}
                  className="h-8 min-w-0 rounded-md bg-sand/95 px-2 text-[11px] font-semibold text-primary outline-none placeholder:text-primary/45 focus:bg-white"
                />
                <label className="sr-only" htmlFor="footer-contact-email">
                  {t("footerContact.email")}
                </label>
                <input
                  id="footer-contact-email"
                  value={contactForm.email}
                  onChange={(event) => updateContactForm("email", event.target.value)}
                  placeholder={t("footerContact.email")}
                  className="h-8 min-w-0 rounded-md bg-sand/95 px-2 text-[11px] font-semibold text-primary outline-none placeholder:text-primary/45 focus:bg-white"
                />
              </div>
              <div className="grid grid-cols-[6.5rem_minmax(0,1fr)] gap-2">
                <label className="sr-only" htmlFor="footer-contact-country">
                  {t("footerContact.countryCode")}
                </label>
                <div>
                  <select
                    id="footer-contact-country"
                    value={contactForm.countryISO}
                    onChange={(event) => updateContactForm("countryISO", event.target.value)}
                    className="h-8 w-full rounded-md bg-sand/95 px-2 text-[11px] font-semibold text-primary outline-none focus:bg-white"
                  >
                    {PHONE_COUNTRIES.map((country) => (
                      <option key={country.iso} value={country.iso}>
                        {country.code} {country.label}
                      </option>
                    ))}
                  </select>
                </div>
                <label className="sr-only" htmlFor="footer-contact-phone">
                  {t("footerContact.phone")}
                </label>
                <input
                  id="footer-contact-phone"
                  inputMode="numeric"
                  value={contactForm.phoneNumber}
                  onChange={(event) => updateContactForm("phoneNumber", event.target.value.replace(/\D/g, ""))}
                  placeholder={t("footerContact.phone")}
                  className="h-8 min-w-0 rounded-md bg-sand/95 px-2 text-[11px] font-semibold text-primary outline-none placeholder:text-primary/45 focus:bg-white"
                  dir="ltr"
                />
              </div>
              <label className="sr-only" htmlFor="footer-contact-message">
                {t("footerContact.message")}
              </label>
              <textarea
                id="footer-contact-message"
                value={contactForm.message}
                onChange={(event) => updateContactForm("message", event.target.value)}
                placeholder={t("footerContact.message")}
                rows="2"
                className="min-h-[3.25rem] resize-none rounded-md bg-sand/95 px-2 py-1.5 text-[11px] font-semibold leading-5 text-primary outline-none placeholder:text-primary/45 focus:bg-white"
              />
              <div className="flex items-center justify-between gap-2">
                <p
                  className={`min-h-4 flex-1 truncate text-[10px] ${
                    contactStatus.type === "error" ? "text-red-100" : "text-sand/72"
                  }`}
                >
                  {contactStatus.message || t("footerContact.hint")}
                </p>
                <button
                  type="submit"
                  disabled={contactSubmitting}
                  className="h-8 shrink-0 rounded-full bg-sand px-3 text-[10px] font-bold text-primary transition hover:bg-white disabled:opacity-60"
                >
                  {contactSubmitting ? t("messages.loading") : t("footerContact.submit")}
                </button>
              </div>
            </form>
          </div>

          <div className={`flex flex-col items-center gap-2.5 md:gap-4 ${lastColumnAlign}`}>
            <p className="font-display text-lg leading-tight text-sand/92 md:text-2xl">Let&apos;s build in stone</p>
            {phoneItems.length > 0 && (
              <div
                className={`flex w-full flex-col items-center justify-center gap-1.5 md:gap-2 ${
                  isRTL ? "md:items-start" : "md:items-end"
                }`}
                aria-label={t("footer.phoneLabel")}
              >
                {phoneItems.map((item) => (
                  <a
                    key={item.normalized}
                    href={`tel:${item.normalized}`}
                    className="inline-flex h-8 items-center gap-1.5 px-2.5 text-[11px] font-semibold text-sand/82 transition hover:text-sand md:text-xs"
                    dir="ltr"
                  >
                    <PhoneIcon className="h-3.5 w-3.5" />
                    <span>{item.value}</span>
                  </a>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </footer>
  );
}
