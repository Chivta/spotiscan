import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Aurora from "../components/Aurora";
import GradientText from "../components/GradientText";

interface LandingProps {
  authenticated: boolean;
}

export default function Landing({ authenticated }: LandingProps) {
  const navigate = useNavigate();

  useEffect(() => {
    if (authenticated) {
      navigate("/dashboard", { replace: true });
    }
  }, [authenticated, navigate]);

  return (
    <div style={{ position: "relative", minHeight: "100vh", width: "100vw", overflow: "hidden" }}>
      <div style={{
        position: "fixed",
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        zIndex: 0,
        pointerEvents: "none"
      }}>
        <Aurora
          colorStops={["#0D4F1C", "#1DB954", "#90EE90"]}
          blend={0.5}
          amplitude={1.0}
          speed={0.5}
        />
      </div>
      <main style={{
        textAlign: "center",
        marginTop: "15vh",
        position: "relative",
        zIndex: 1,
        padding: "0 2rem"
      }}>
        <h1 style={{
          fontSize: "5rem",
          fontFamily: "'Outfit', sans-serif",
          fontWeight: "900",
          marginBottom: "1.5rem",
          color: "#fff",
          letterSpacing: "-0.02em",
          textShadow: "0 0 40px rgba(29, 185, 84, 0.5), 0 0 80px rgba(29, 185, 84, 0.3)"
        }}>
          <GradientText
            colors={["#1DB954", "#4ade80", "#86efac", "#4ade80", "#1DB954"]}
            animationSpeed={3}
            showBorder={false}
          >SpotiScan</GradientText>
        </h1>
        <p style={{
          fontSize: "1.25rem",
          color: "rgba(255, 255, 255, 0.8)",
          maxWidth: "600px",
          margin: "0 auto 2.5rem",
          lineHeight: "1.6"
        }}>
          Take control of your music library. Scan your Spotify playlists and liked songs
          to discover and remove tracks that don't belong.
        </p>

        <div style={{
          display: "flex",
          justifyContent: "center",
          gap: "2rem",
          marginBottom: "3rem",
          flexWrap: "wrap"
        }}>
          <div style={{
            background: "rgba(255, 255, 255, 0.05)",
            backdropFilter: "blur(10px)",
            borderRadius: "16px",
            padding: "1.5rem 2rem",
            border: "1px solid rgba(255, 255, 255, 0.1)",
            minWidth: "200px"
          }}>
            <div style={{ fontSize: "2rem", marginBottom: "0.5rem" }}>🔍</div>
            <h3 style={{ color: "#fff", margin: "0 0 0.5rem", fontSize: "1.1rem" }}>Deep Scan</h3>
            <p style={{ color: "rgba(255, 255, 255, 0.6)", margin: 0, fontSize: "0.9rem" }}>
              Analyze all your playlists and saved tracks
            </p>
          </div>
          <div style={{
            background: "rgba(255, 255, 255, 0.05)",
            backdropFilter: "blur(10px)",
            borderRadius: "16px",
            padding: "1.5rem 2rem",
            border: "1px solid rgba(255, 255, 255, 0.1)",
            minWidth: "200px"
          }}>
            <div style={{ fontSize: "2rem", marginBottom: "0.5rem" }}>🧹</div>
            <h3 style={{ color: "#fff", margin: "0 0 0.5rem", fontSize: "1.1rem" }}>One-Click Clean</h3>
            <p style={{ color: "rgba(255, 255, 255, 0.6)", margin: 0, fontSize: "0.9rem" }}>
              Remove unwanted tracks instantly
            </p>
          </div>
          <div style={{
            background: "rgba(255, 255, 255, 0.05)",
            backdropFilter: "blur(10px)",
            borderRadius: "16px",
            padding: "1.5rem 2rem",
            border: "1px solid rgba(255, 255, 255, 0.1)",
            minWidth: "200px"
          }}>
            <div style={{ fontSize: "2rem", marginBottom: "0.5rem" }}>🔒</div>
            <h3 style={{ color: "#fff", margin: "0 0 0.5rem", fontSize: "1.1rem" }}>Privacy First</h3>
            <p style={{ color: "rgba(255, 255, 255, 0.6)", margin: 0, fontSize: "0.9rem" }}>
              No data stored, secure authentication
            </p>
          </div>
        </div>

        <a href="/api/auth/start" style={{
          display: "inline-flex",
          alignItems: "center",
          gap: "0.75rem",
          padding: "1rem 2.5rem",
          background: "#1DB954",
          color: "#000",
          borderRadius: "50px",
          fontWeight: "600",
          textDecoration: "none",
          fontSize: "1.1rem",
          transition: "all 0.2s ease",
          boxShadow: "0 4px 20px rgba(29, 185, 84, 0.4)"
        }}>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"/>
          </svg>
          Continue with Spotify
        </a>

        <p style={{
          marginTop: "1.5rem",
          color: "rgba(255, 255, 255, 0.4)",
          fontSize: "0.85rem"
        }}>
          Free to use • No credit card required
        </p>
      </main>
    </div>
  );
}
