import { useTranslation } from "../lib/i18n";
import irFlag from "@shared/assets/flags/ir.svg";
import saFlag from "@shared/assets/flags/sa.svg";
import ukFlag from "@shared/assets/flags/uk.svg";

const options = [
  { code: "en", flag: ukFlag },
  { code: "fa", flag: irFlag },
  { code: "ar", flag: saFlag }
];

export default function LanguageSwitch() {
  const { lang, setLang, t } = useTranslation();

  return (
    <div className="inline-flex items-center rounded-full border border-primary/20 bg-white/70 p-1 text-xs shadow-sm">
      {options.map((option) => {
        const isActive = lang === option.code;
        return (
          <button
            key={option.code}
            type="button"
            onClick={() => setLang(option.code)}
            className={`flex items-center gap-2 rounded-full px-3 py-1.5 transition ${isActive ? "bg-primary text-white" : "text-primary/70 hover:text-primary"
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
