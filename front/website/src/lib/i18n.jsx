import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import en from "@shared/i18n/en.json";
import fa from "@shared/i18n/fa.json";
import ar from "@shared/i18n/ar.json";

const dictionaries = { en, fa, ar };
const supportedLangs = Object.keys(dictionaries);
const LanguageContext = createContext({
  lang: "en",
  setLang: () => {},
  t: (key) => key
});

const getValue = (obj, path) => {
  return path.split(".").reduce((acc, part) => (acc ? acc[part] : undefined), obj);
};

export const LanguageProvider = ({ children }) => {
  const [lang, setLang] = useState("en");
  const setSafeLang = useCallback((nextLang) => {
    setLang(supportedLangs.includes(nextLang) ? nextLang : "en");
  }, []);

  useEffect(() => {
    document.documentElement.lang = lang;
    document.documentElement.dir = lang === "fa" || lang === "ar" ? "rtl" : "ltr";
  }, [lang]);

  const value = useMemo(() => {
    const t = (key) => {
      const current = getValue(dictionaries[lang], key);
      if (current !== undefined) {
        return current;
      }
      const fallback = getValue(dictionaries.en, key);
      return fallback !== undefined ? fallback : key;
    };
    return { lang, setLang: setSafeLang, t };
  }, [lang, setSafeLang]);

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>;
};

export const useTranslation = () => useContext(LanguageContext);
