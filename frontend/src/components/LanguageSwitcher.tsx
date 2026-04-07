import { useState, useEffect, useRef } from "react";
import { useLanguage } from "../context/LanguageContext";
import type { Lang } from "../context/LanguageContext";

const LANGUAGES: { code: Lang; label: string; flag: string }[] = [
  { code: "en", label: "English", flag: "/countries/sh.svg" },
  { code: "uk", label: "Українська", flag: "/countries/ua.svg" },
];

const FLAG_STYLE = { width: 28, height: 21, borderRadius: 2, objectFit: "cover" as const, display: "block", flexShrink: 0 };

export default function LanguageSwitcher() {
  const { lang, setLang } = useLanguage();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  const current = LANGUAGES.find(l => l.code === lang)!;

  return (
    <div ref={ref} style={{ position: "relative" }}>
      <button
        onClick={() => setOpen(o => !o)}
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          background: "transparent",
          border: "1px solid rgba(255,255,255,0.2)",
          borderRadius: 50,
          color: "rgba(255,255,255,0.8)",
          padding: "5px 8px",
          cursor: "pointer",
          transition: "border-color 0.2s ease, color 0.2s ease",
        }}
        onMouseEnter={e => { e.currentTarget.style.borderColor = "rgba(255,255,255,0.5)"; e.currentTarget.style.color = "#fff"; }}
        onMouseLeave={e => { e.currentTarget.style.borderColor = "rgba(255,255,255,0.2)"; e.currentTarget.style.color = "rgba(255,255,255,0.8)"; }}
      >
        <img src={current.flag} alt={current.label} style={FLAG_STYLE} />
        <svg width="10" height="10" viewBox="0 0 10 10" style={{ opacity: 0.6, transition: "transform 0.2s", transform: open ? "rotate(180deg)" : "none", flexShrink: 0 }}>
          <path d="M1 3l4 4 4-4" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </button>

      {open && (
        <div style={{
          position: "absolute",
          top: "calc(100% + 6px)",
          right: 0,
          background: "rgba(18,18,18,0.97)",
          border: "1px solid rgba(255,255,255,0.12)",
          borderRadius: 10,
          overflow: "hidden",
          minWidth: 148,
          zIndex: 100,
          boxShadow: "0 8px 24px rgba(0,0,0,0.5)",
        }}>
          {LANGUAGES.map(l => (
            <button
              key={l.code}
              onClick={() => { setLang(l.code); setOpen(false); }}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 10,
                width: "100%",
                background: l.code === lang ? "rgba(255,255,255,0.08)" : "transparent",
                border: "none",
                padding: "10px 14px",
                color: l.code === lang ? "#fff" : "rgba(255,255,255,0.65)",
                fontSize: 13,
                fontWeight: l.code === lang ? 600 : 400,
                cursor: l.code === lang ? "default" : "pointer",
                textAlign: "left",
                transition: "background 0.15s ease",
              }}
              onMouseEnter={e => { if (l.code !== lang) e.currentTarget.style.background = "rgba(255,255,255,0.05)"; }}
              onMouseLeave={e => { if (l.code !== lang) e.currentTarget.style.background = "transparent"; }}
            >
              <img src={l.flag} alt={l.label} style={FLAG_STYLE} />
              {l.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
