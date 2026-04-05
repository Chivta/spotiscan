import * as React from "react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import Aurora from "../components/react-bits/Aurora";

interface AuthPageProps {
  initialMode?: "login" | "signup";
}

type Mode = "login" | "signup";

export default function AuthPage({ initialMode = "signup" }: AuthPageProps) {
  const [mode, setMode] = useState<Mode>(initialMode);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleModeSwitch = (newMode: Mode) => {
    setMode(newMode);
    setEmail("");
    setPassword("");
    setError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const endpoint = mode === "login" ? "/api/auth/login" : "/api/auth/signup";
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ Email: email, Password: password }),
      });
      if (res.ok) {
        navigate("/", { replace: true });
      } else {
        const body = await res.json().catch(() => null);
        const code = body?.code;
        if (res.status === 401) {
          setError("Incorrect email or password.");
        } else if (res.status === 409 || code === "EMAIL_EXISTS") {
          setError("An account with this email already exists.");
        } else if (res.status === 400 || code === "BAD_REQUEST") {
          setError("Please check your email and password.");
        } else {
          setError("Something went wrong. Please try again.");
        }
      }
    } catch {
      setError("Something went wrong. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const cardStyle: React.CSSProperties = {
    background: "rgba(255, 255, 255, 0.05)",
    backdropFilter: "blur(10px)",
    border: "1px solid rgba(255, 255, 255, 0.1)",
    borderRadius: 16,
    padding: 32,
    maxWidth: 400,
    width: "100%",
    boxSizing: "border-box",
  };

  const inputStyle: React.CSSProperties = {
    background: "rgba(255, 255, 255, 0.1)",
    border: "1px solid rgba(255, 255, 255, 0.2)",
    borderRadius: 8,
    color: "#fff",
    padding: "12px 16px",
    fontSize: 16,
    width: "100%",
    outline: "none",
    boxSizing: "border-box",
  };

  return (
    <div style={{ position: "relative", minHeight: "100vh", width: "100%", overflow: "hidden" }}>
      <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, zIndex: 0, pointerEvents: "none" }}>
        <Aurora colorStops={["#0D4F1C", "#1DB954", "#90EE90"]} blend={0.5} amplitude={1.0} speed={0.5} />
      </div>
      <div style={{
        position: "relative",
        zIndex: 1,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        minHeight: "100vh",
        padding: 24,
        boxSizing: "border-box",
      }}>
        <div style={cardStyle}>
          <h1 style={{
            fontFamily: "'Outfit', sans-serif",
            fontWeight: 800,
            fontSize: "2rem",
            color: "#fff",
            textAlign: "center",
            marginBottom: 24,
          }}>
            SpotiScan
          </h1>

          {/* Pill toggle */}
          <div style={{
            display: "flex",
            gap: 4,
            background: "rgba(255,255,255,0.05)",
            borderRadius: 50,
            padding: 4,
            marginBottom: 24,
          }}>
            {(["signup", "login"] as Mode[]).map(m => (
              <button
                key={m}
                onClick={() => handleModeSwitch(m)}
                style={{
                  flex: 1,
                  background: mode === m ? "#1DB954" : "rgba(255, 255, 255, 0.08)",
                  color: mode === m ? "#000" : "rgba(255, 255, 255, 0.5)",
                  fontWeight: mode === m ? 600 : 400,
                  border: "none",
                  borderRadius: 50,
                  padding: "8px 24px",
                  cursor: "pointer",
                  fontSize: 14,
                  transition: "all 0.2s ease",
                }}
              >
                {m === "login" ? "Log In" : "Sign Up"}
              </button>
            ))}
          </div>

          <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: 16 }}>
            <input
              type="email"
              placeholder="Email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              style={inputStyle}
              required
              disabled={loading}
            />
            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              style={inputStyle}
              required
              disabled={loading}
            />

            {error && (
              <div style={{ color: "#e74c3c", fontSize: 14, textAlign: "center" }}>
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              style={{
                background: "#1DB954",
                color: "#000",
                borderRadius: 50,
                fontWeight: 600,
                fontSize: 14,
                padding: "12px 24px",
                width: "100%",
                border: "none",
                cursor: loading ? "not-allowed" : "pointer",
                opacity: loading ? 0.7 : 1,
              }}
            >
              {loading
                ? (mode === "login" ? "Logging in…" : "Signing up…")
                : (mode === "login" ? "Log In" : "Sign Up")}
            </button>
          </form>

          <p style={{ marginTop: 20, textAlign: "center", fontSize: 14, color: "rgba(255,255,255,0.4)" }}>
            {mode === "signup" ? "Already have an account? " : "Don't have an account? "}
            <span
              onClick={() => handleModeSwitch(mode === "signup" ? "login" : "signup")}
              style={{
                color: "#1DB954",
                cursor: "pointer",
                fontWeight: 600,
                textDecoration: "underline",
                textUnderlineOffset: 2,
              }}
            >
              {mode === "signup" ? "Log in" : "Sign up"}
            </span>
          </p>
        </div>
      </div>
    </div>
  );
}
