import { useTranslation } from "../lib/i18n";
import irFlag from "@shared/assets/flags/ir.svg";
import saFlag from "@shared/assets/flags/sa.svg";
import ukFlag from "@shared/assets/flags/uk.svg";

const options = [
  { code: "en", flag: ukFlag },
  { code: "fa", flag: irFlag },
  { code: "ar", flag: saFlag }
];

export default function LanguageSwitch({ tone = "default" }) {
  const { lang, setLang, t } = useTranslation();
  const isLightTone = tone === "light";
  const isFooterTone = tone === "footer";

  return (
    <div
      className={`inline-flex items-center rounded-full p-1 text-xs shadow-sm ${
        isFooterTone
          ? "bg-transparent shadow-none"
          : isLightTone
            ? "border border-sand/35 bg-white/10 backdrop-blur-md"
            : "border border-primary/20 bg-white/70"
      }`}
    >
      {options.map((option) => {
        const isActive = lang === option.code;
        return (
          <button
            key={option.code}
            type="button"
            onClick={() => setLang(option.code)}
            className={`flex items-center gap-2 rounded-full px-3 py-1.5 transition ${
              isFooterTone
                ? isActive
                  ? "bg-sand/18 text-sand"
                  : "text-sand/70 hover:text-sand"
                : isLightTone
                ? isActive
                  ? "bg-sand/20 text-sand"
                  : "text-sand/75 hover:text-sand"
                : isActive
                  ? "bg-primary text-white"
                  : "text-primary/70 hover:text-primary"
            }`}
            aria-pressed={isActive}
          >
            <img src={option.flag} alt={t(`language.${option.code}`)} className="h-3 w-4" />
            {/* <span className="hidden md:inline">{t(`language.${option.code}`)}</span> */}
          </button>
        );
      })}
    </div>
  );
}
