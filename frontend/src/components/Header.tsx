import { useNavigate, useLocation } from "react-router-dom";
import type { User } from "../types/models";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import LanguageSwitcher from "./LanguageSwitcher";

interface HeaderProps {
  user: User;
  onLogout: () => void;
}

const btnStyle: React.CSSProperties = {
  background: "transparent",
  border: "1px solid rgba(255, 255, 255, 0.2)",
  borderRadius: 50,
  color: "rgba(255, 255, 255, 0.7)",
  padding: "6px 16px",
  fontSize: 13,
  cursor: "pointer",
};

const btnEnter = (e: React.MouseEvent<HTMLElement>) => {
  e.currentTarget.style.borderColor = "rgba(255,255,255,0.5)";
  e.currentTarget.style.color = "#fff";
};
const btnLeave = (e: React.MouseEvent<HTMLElement>) => {
  e.currentTarget.style.borderColor = "rgba(255,255,255,0.2)";
  e.currentTarget.style.color = "rgba(255,255,255,0.7)";
};

export default function Header({ user, onLogout }: HeaderProps) {
  const { lang } = useLanguage();
  const tx = translations[lang];
  const navigate = useNavigate();
  const location = useLocation();
  const isAdmin = user.userRole === "admin";

  return (
    <>
      <style>{`
        .header-email { display: inline; }
        @media (max-width: 600px) {
          .header-email { display: none; }
          .header-nav { gap: 6px !important; }
          .header-nav-btn { padding: 6px 10px !important; }
        }
      `}</style>
      <header style={{
        position: "fixed",
        top: 0,
        left: 0,
        right: 0,
        height: 56,
        padding: "0 16px",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        background: "rgba(0, 0, 0, 0.3)",
        backdropFilter: "blur(8px)",
        zIndex: 10,
        boxSizing: "border-box",
      }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <img src="/logo.png" alt="" aria-hidden={true} style={{ height: 32, width: 32 }} />
          <span style={{
            fontFamily: "'Outfit', sans-serif",
            fontWeight: 800,
            color: "#fff",
            fontSize: "1.1rem",
          }}>
            RuScan
          </span>
        </div>

        <div className="header-nav" style={{ display: "flex", alignItems: "center", gap: 10 }}>
          {user.Email && (
            <span className="header-email" style={{ color: "rgba(255, 255, 255, 0.4)", fontSize: 13 }}>
              {user.Email}
            </span>
          )}

          {isAdmin && (
            <button
              className="header-nav-btn"
              onClick={() => navigate(location.pathname === "/admin" ? "/dashboard" : "/admin")}
              style={btnStyle}
              onMouseEnter={btnEnter}
              onMouseLeave={btnLeave}
            >
              {location.pathname === "/admin" ? tx.tabScanner : tx.adminTitle}
            </button>
          )}

          <LanguageSwitcher />

          <button
            className="header-nav-btn"
            onClick={onLogout}
            style={btnStyle}
            onMouseEnter={btnEnter}
            onMouseLeave={btnLeave}
          >
            {tx.logOut}
          </button>
        </div>
      </header>
    </>
  );
}
