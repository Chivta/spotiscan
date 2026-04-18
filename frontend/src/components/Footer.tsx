import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";

export default function Footer() {
  const { lang } = useLanguage();
  const tx = translations[lang];

  return (
    <footer style={{
      position: "relative",
      zIndex: 1,
      textAlign: "center",
      padding: "24px 16px",
      color: "rgba(255,255,255,0.35)",
      fontSize: 13,
    }}>
      <a
        href="https://t.me/ukbotsup"
        target="_blank"
        rel="noopener noreferrer"
        style={{
          color: "rgba(255,255,255,0.45)",
          textDecoration: "none",
          transition: "color 0.2s ease",
        }}
        onMouseEnter={e => { e.currentTarget.style.color = "rgba(255,255,255,0.85)"; }}
        onMouseLeave={e => { e.currentTarget.style.color = "rgba(255,255,255,0.45)"; }}
      >
        {tx.feedback} · @ukbotsup
      </a>
    </footer>
  );
}
