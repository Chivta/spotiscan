import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";

export default function Landing({ authenticated }) {
  const navigate = useNavigate();

  useEffect(() => {
    if (authenticated) {
      navigate("/dashboard", { replace: true });
    }
  }, [authenticated, navigate]);

  return (
    <main style={{ textAlign: "center", marginTop: "3rem" }}>
      <h1>Welcome to SpotiScan</h1>
      <p>Analyze your Spotify playlists and discover artists from Russia and more!</p>
      <ul style={{ listStyle: "none", padding: 0 }}>
        <li>🔍 Scan your playlists for Russian artists</li>
        <li>💩 Clean your music library from shit</li>
        <li>🔒 100% privacy, no data stored</li>
      </ul>
      <a href="/api/auth/start" style={{
        display: "inline-block",
        marginTop: "2rem",
        padding: "1rem 2rem",
        background: "#1DB954",
        color: "white",
        borderRadius: "30px",
        fontWeight: "bold",
        textDecoration: "none",
        fontSize: "1.2rem"
      }}>
        Sign up with Spotify
      </a>
    </main>
  );
}
