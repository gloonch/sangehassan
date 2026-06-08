import { useTranslation } from "../lib/i18n";

const options = [
  { code: "fa", label: "fa" },
  { code: "en", label: "en" },
  { code: "ar", label: "ar" }
];

export default function LanguageSwitch({ tone = "default" }) {
  const { lang, setLang, t } = useTranslation();
  const isLightTone = tone === "light";
  const isMenuTone = tone === "menu";

  return (
    <div
      className={`inline-flex items-center rounded-full px-3 py-2 text-[11px] font-semibold uppercase tracking-[0.16em] shadow-sm transition ${isLightTone
          ? "text-sand hover:bg-sand/25"
          : isMenuTone
            ? "bg-primary/90 text-sand hover:bg-primary"
            : "text-accent shadow-none"
        }`}
      aria-label="Language"
    >
      {options.map((option, index) => {
        const isActive = lang === option.code;
        return (
          <span key={option.code} className="inline-flex items-center">
            <button
              type="button"
              onClick={() => setLang(option.code)}
              className={`px-1 leading-none transition ${isActive ? "opacity-100" : "opacity-60 hover:opacity-100"
                }`}
              aria-label={t(`language.${option.code}`)}
              aria-pressed={isActive}
            >
              {option.label}
            </button>
            {index < options.length - 1 ? <span className="px-1 opacity-45">|</span> : null}
          </span>
        );
      })}
    </div>
  );
}
