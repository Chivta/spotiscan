import { createContext, useContext, useState, useCallback } from "react";
import type { ReactNode } from "react";
import type { Lang } from "../i18n";

export type { Lang };

const LS_KEY = "spotiscan_lang";

function detectLang(): Lang {
  try {
    const stored = localStorage.getItem(LS_KEY);
    if (stored === "uk" || stored === "en") return stored;
  } catch {
    // localStorage unavailable (privacy mode, SSR, etc.)
  }
  return navigator.language?.startsWith("uk") ? "uk" : "en";
}

interface LanguageContextValue {
  lang: Lang;
  setLang: (lang: Lang) => void;
}

const LanguageContext = createContext<LanguageContextValue>({
  lang: "en",
  setLang: () => {},
});

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(detectLang);

  const setLang = useCallback((next: Lang) => {
    try {
      localStorage.setItem(LS_KEY, next);
    } catch {
      // Ignore if localStorage is unavailable
    }
    setLangState(next);
  }, []);

  return (
    <LanguageContext.Provider value={{ lang, setLang }}>
      {children}
    </LanguageContext.Provider>
  );
}

export function useLanguage() {
  return useContext(LanguageContext);
}
