import { createContext, useContext, useEffect, useMemo, useState } from "react";
import en from "@shared/i18n/en.json";
import fa from "@shared/i18n/fa.json";
import ar from "@shared/i18n/ar.json";

const dictionaries = { en, fa, ar };
const supportedLangs = Object.keys(dictionaries);
const configuredDefaultLang = String(import.meta.env.VITE_DEFAULT_LANG || "en").toLowerCase();
const defaultLang = supportedLangs.includes(configuredDefaultLang) ? configuredDefaultLang : "en";
const LanguageContext = createContext({
  lang: defaultLang,
  setLang: () => { },
  t: (key) => key
});

const getValue = (obj, path) => {
  return path.split(".").reduce((acc, part) => (acc ? acc[part] : undefined), obj);
};

export const LanguageProvider = ({ children }) => {
  const [lang, setLang] = useState(() => {
    const stored = window.localStorage.getItem("lang");
    if (stored && supportedLangs.includes(stored)) return stored;
    return defaultLang;
  });

  useEffect(() => {
    window.localStorage.setItem("lang", lang);
    document.documentElement.lang = lang;
    document.documentElement.dir = lang === "fa" || lang === "ar" ? "rtl" : "ltr";
  }, [lang]);

  const value = useMemo(() => {
    const setSafeLang = (nextLang) => {
      setLang(supportedLangs.includes(nextLang) ? nextLang : defaultLang);
    };
    const t = (key) => {
      const current = getValue(dictionaries[lang], key);
      if (current !== undefined) {
        return current;
      }
      const fallback = getValue(dictionaries.en, key);
      return fallback !== undefined ? fallback : key;
    };
    return { lang, setLang: setSafeLang, t };
  }, [lang]);

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>
};

export const useTranslation = () => useContext(LanguageContext);
