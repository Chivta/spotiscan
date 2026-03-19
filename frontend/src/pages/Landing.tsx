import { useNavigate } from "react-router-dom";
import Aurora from "../components/react-bits/Aurora";

export default function Landing() {
  const navigate = useNavigate();

  return (
    <div style={{ position: "relative", minHeight: "100vh", width: "100%", overflow: "hidden" }}>
      <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, zIndex: 0, pointerEvents: "none" }}>
        <Aurora colorStops={["#0D4F1C", "#1DB954", "#90EE90"]} blend={0.5} amplitude={1.0} speed={0.5} />
      </div>

      <div style={{
        position: "relative",
        zIndex: 1,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        minHeight: "100vh",
        padding: "24px",
        boxSizing: "border-box",
        textAlign: "center",
      }}>
        <h1 style={{
          fontFamily: "'Outfit', sans-serif",
          fontWeight: 800,
          fontSize: "clamp(2.5rem, 6vw, 4rem)",
          color: "#fff",
          marginBottom: 16,
          lineHeight: 1.1,
        }}>
          SpotiScan
        </h1>

        <p style={{
          fontSize: "clamp(1rem, 2.5vw, 1.25rem)",
          color: "rgba(255, 255, 255, 0.65)",
          maxWidth: 520,
          lineHeight: 1.6,
          marginBottom: 12,
        }}>
          Scan your Spotify playlists and liked songs to find and remove tracks by Russian artists.
        </p>

        <p style={{
          fontSize: 14,
          color: "rgba(255, 255, 255, 0.35)",
          maxWidth: 400,
          lineHeight: 1.6,
          marginBottom: 40,
        }}>
          Placeholder — more details about how SpotiScan works will go here.
        </p>

        <div style={{ display: "flex", gap: 12, flexWrap: "wrap", justifyContent: "center" }}>
          <button
            onClick={() => navigate("/signup")}
            style={{
              padding: "14px 36px",
              background: "#1DB954",
              color: "#000",
              border: "none",
              borderRadius: 50,
              fontWeight: 700,
              fontSize: 16,
              cursor: "pointer",
              transition: "opacity 0.2s ease",
            }}
            onMouseEnter={e => { e.currentTarget.style.opacity = "0.85"; }}
            onMouseLeave={e => { e.currentTarget.style.opacity = "1"; }}
          >
            Get Started
          </button>

        </div>
      </div>
    </div>
  );
}
