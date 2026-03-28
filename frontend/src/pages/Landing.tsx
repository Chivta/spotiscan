import { useNavigate } from "react-router-dom";
import Aurora from "../components/react-bits/Aurora";
import PlaylistScanner from "../components/PlaylistScanner";

export default function Landing() {
  const navigate = useNavigate();

  return (
    <div style={{ position: "relative", minHeight: "100vh", width: "100%", overflow: "hidden" }}>
      <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, zIndex: 0, pointerEvents: "none" }}>
        <Aurora colorStops={["#0D4F1C", "#1DB954", "#90EE90"]} blend={0.5} amplitude={1.0} speed={0.5} />
      </div>

      {/* Top bar */}
      <div style={{
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
        <span style={{ fontFamily: "'Outfit', sans-serif", fontWeight: 800, color: "#fff", fontSize: "1.1rem" }}>
          SpotiScan
        </span>
        <div style={{ display: "flex", gap: 10 }}>
          <button
            onClick={() => navigate("/login")}
            style={{
              padding: "7px 18px",
              background: "transparent",
              color: "rgba(255, 255, 255, 0.8)",
              border: "1px solid rgba(255, 255, 255, 0.25)",
              borderRadius: 50,
              fontWeight: 600,
              fontSize: 13,
              cursor: "pointer",
              transition: "all 0.2s ease",
            }}
            onMouseEnter={e => { e.currentTarget.style.borderColor = "rgba(255,255,255,0.6)"; e.currentTarget.style.color = "#fff"; }}
            onMouseLeave={e => { e.currentTarget.style.borderColor = "rgba(255,255,255,0.25)"; e.currentTarget.style.color = "rgba(255,255,255,0.8)"; }}
          >
            Sign in
          </button>
          <button
            onClick={() => navigate("/signup")}
            style={{
              padding: "7px 18px",
              background: "#1DB954",
              color: "#000",
              border: "none",
              borderRadius: 50,
              fontWeight: 700,
              fontSize: 13,
              cursor: "pointer",
              transition: "opacity 0.2s ease",
            }}
            onMouseEnter={e => { e.currentTarget.style.opacity = "0.85"; }}
            onMouseLeave={e => { e.currentTarget.style.opacity = "1"; }}
          >
            Create account
          </button>
        </div>
      </div>

      <div style={{
        position: "relative",
        zIndex: 1,
        width: "100%",
        padding: "88px 24px 60px",
        boxSizing: "border-box",
      }}>
        {/* Hero */}
        <div style={{ textAlign: "center", marginBottom: 48 }}>
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
            margin: "0 auto 12px",
          }}>
            Scan your Spotify playlists to find and remove tracks by Russian artists.
          </p>

          <p style={{
            fontSize: 14,
            color: "rgba(255, 255, 255, 0.35)",
            maxWidth: 400,
            lineHeight: 1.6,
            margin: "0 auto",
          }}>
            Paste a playlist URL below — no sign-in required for a few scans.
          </p>
        </div>

        {/* Scanner */}
        <div style={{ maxWidth: 800, margin: "0 auto" }}>
          <PlaylistScanner />
        </div>
      </div>
    </div>
  );
}
