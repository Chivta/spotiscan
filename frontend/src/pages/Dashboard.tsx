import Aurora from "../components/react-bits/Aurora";
import PlaylistScanner from "../components/PlaylistScanner";

export default function Dashboard() {
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
          <PlaylistScanner />
        </div>
      </div>
    </div>
  );
}
