import type { User } from "../types/models";

interface HeaderProps {
  user: User;
  onLogout: () => void;
}

export default function Header({ user, onLogout }: HeaderProps) {
  return (
    <header style={{
      position: "fixed",
      top: 0,
      left: 0,
      right: 0,
      height: 56,
      padding: "0 24px",
      display: "flex",
      alignItems: "center",
      justifyContent: "space-between",
      background: "rgba(0, 0, 0, 0.3)",
      backdropFilter: "blur(8px)",
      zIndex: 10,
      boxSizing: "border-box",
    }}>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <img src="/logo.png" alt="SpotiScan" style={{ height: 32, width: 32 }} />
        <span style={{
          fontFamily: "'Outfit', sans-serif",
          fontWeight: 800,
          color: "#fff",
          fontSize: "1.1rem",
        }}>
          SpotiScan
        </span>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        {user.Email && (
          <span style={{ color: "rgba(255, 255, 255, 0.4)", fontSize: 13 }}>
            {user.Email}
          </span>
        )}
        <button
          onClick={onLogout}
          style={{
            background: "transparent",
            border: "1px solid rgba(255, 255, 255, 0.2)",
            borderRadius: 50,
            color: "rgba(255, 255, 255, 0.7)",
            padding: "6px 16px",
            fontSize: 13,
            cursor: "pointer",
          }}
          onMouseEnter={e => {
            e.currentTarget.style.borderColor = "rgba(255,255,255,0.5)";
            e.currentTarget.style.color = "#fff";
          }}
          onMouseLeave={e => {
            e.currentTarget.style.borderColor = "rgba(255,255,255,0.2)";
            e.currentTarget.style.color = "rgba(255,255,255,0.7)";
          }}
        >
          Log out
        </button>
      </div>
    </header>
  );
}
