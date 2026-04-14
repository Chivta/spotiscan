import { useState } from "react";
import Aurora from "../components/react-bits/Aurora";
import PlaylistScanner from "../components/PlaylistScanner";
import ArtistSuggestions from "../components/ArtistSuggestions";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";

type Tab = "scanner" | "suggestions";

export default function Dashboard() {
  const { lang } = useLanguage();
  const tx = translations[lang];
  const [tab, setTab] = useState<Tab>("scanner");
  const [deletePrefillName, setDeletePrefillName] = useState<string | undefined>(undefined);

  const handleNotRussian = (artistName: string) => {
    setDeletePrefillName(artistName);
    setTab("suggestions");
  };

  return (
    <div style={{ position: "relative", minHeight: "100vh", width: "100%", overflow: "hidden" }}>
      <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, zIndex: 0, pointerEvents: "none" }}>
        <Aurora colorStops={["#0D4F1C", "#1DB954", "#90EE90"]} blend={0.5} amplitude={1.0} speed={0.5} />
      </div>

      <div style={{
        position: "relative",
        zIndex: 1,
        width: "100%",
        padding: "88px 24px 40px",
        boxSizing: "border-box",
        overflow: "auto",
      }}>
        <div style={{ maxWidth: 800, margin: "0 auto" }}>
          {/* Tab bar */}
          <div style={{
            display: "flex",
            gap: 4,
            marginBottom: 24,
            background: "rgba(255,255,255,0.05)",
            border: "1px solid rgba(255,255,255,0.1)",
            borderRadius: 50,
            padding: 4,
            width: "fit-content",
          }}>
            {(["scanner", "suggestions"] as Tab[]).map(t => (
              <button
                key={t}
                onClick={() => setTab(t)}
                style={{
                  padding: "8px 20px",
                  borderRadius: 50,
                  border: "none",
                  fontSize: 13,
                  fontWeight: 600,
                  cursor: "pointer",
                  transition: "all 0.15s ease",
                  background: tab === t ? "#1DB954" : "transparent",
                  color: tab === t ? "#000" : "rgba(255,255,255,0.55)",
                }}
              >
                {t === "scanner" ? tx.tabScanner : tx.tabSuggestions}
              </button>
            ))}
          </div>

          {/* Scanner is always mounted to preserve scan results; hidden when not active */}
          <div style={{ display: tab === "scanner" ? "block" : "none" }}>
            <PlaylistScanner onNotRussian={handleNotRussian} />
          </div>

          {tab === "suggestions" && (
            <ArtistSuggestions
              deletePrefillName={deletePrefillName}
              initialTab={deletePrefillName !== undefined ? "delete" : "insert"}
            />
          )}
        </div>
      </div>
    </div>
  );
}
