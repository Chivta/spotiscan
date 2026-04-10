import * as React from "react";
import { useState, useMemo, useRef, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import AnimatedList from "./react-bits/AnimatedList";
import type { Artist, TrackArtist, Track, RuContent } from "../types/models";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";

const BASE62_ID_REGEX = /^[A-Za-z0-9]{22}$/;

type ResourceType = "playlist" | "track" | "album" | "artist";

function countryFlag(country: string): string {
  return `/countries/${country.toLowerCase().trim()}.svg`;
}

function detectResource(input: string): { type: ResourceType; id: string } | null {
  const urlMatch = /open\.spotify\.com\/(playlist|track|album|artist)\/([A-Za-z0-9]{22})/.exec(input);
  if (urlMatch) return { type: urlMatch[1] as ResourceType, id: urlMatch[2] };
  const uriMatch = /spotify:(playlist|track|album|artist):([A-Za-z0-9]{22})/.exec(input);
  if (uriMatch) return { type: uriMatch[1] as ResourceType, id: uriMatch[2] };
  if (BASE62_ID_REGEX.test(input.trim())) return { type: "playlist", id: input.trim() };
  return null;
}

const cardStyle: React.CSSProperties = {
  background: "rgba(255, 255, 255, 0.05)",
  backdropFilter: "blur(10px)",
  border: "1px solid rgba(255, 255, 255, 0.1)",
  borderRadius: 16,
  padding: 24,
};

const inputStyle: React.CSSProperties = {
  flex: 1,
  padding: "12px 16px",
  fontSize: 16,
  background: "rgba(255, 255, 255, 0.1)",
  border: "1px solid rgba(255, 255, 255, 0.2)",
  borderRadius: 8,
  color: "#fff",
  outline: "none",
};

const buttonStyle: React.CSSProperties = {
  padding: "12px 24px",
  background: "#1DB954",
  color: "#000",
  border: "none",
  borderRadius: 50,
  fontWeight: 600,
  fontSize: 14,
  cursor: "pointer",
  transition: "all 0.2s ease",
};

type ErrorKey = "invalidInput" | "anonQuotaExceeded" | "notFound" | "badRequest" | "databaseError" | "spotifyApiError" | "internalError" | "tooManyRequests" | "unauthorized" | "somethingWentWrong";
type ErrorState = { key: ErrorKey; type: "warning" | "error" | "auth" };

function artistDesc(artist: Artist, lang: "uk" | "en"): string {
  if (lang === "uk") return artist.DescriptionUA || artist.DescriptionEN;
  return artist.DescriptionEN || artist.DescriptionUA;
}

export default function PlaylistScanner() {
  const navigate = useNavigate();
  const { lang } = useLanguage();
  const tx = translations[lang];
  const [inputValue, setInputValue] = useState("");
  const [ruContent, setRuContent] = useState<RuContent | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<ErrorState | null>(null);
  const [lastInput, setLastInput] = useState<string | null>(null);
  const [resultTab, setResultTab] = useState<"tracks" | "artists">("tracks");
  const [tracksSearch, setTracksSearch] = useState("");
  const [artistsSearch, setArtistsSearch] = useState("");
  const [scanCache, setScanCache] = useState<Record<string, RuContent>>({});
  const [fromCache, setFromCache] = useState(false);

  const detected = detectResource(inputValue);

  const fetchRuContent = async (type: ResourceType, id: string): Promise<RuContent> => {
    const response = await fetch(`/api/${type}/${encodeURIComponent(id)}/rucontent`);
    if (!response.ok) {
      const body = await response.json().catch(() => null);
      const err = new Error(body?.error || "Something went wrong") as Error & { code?: string };
      err.code = body?.code;
      throw err;
    }
    return response.json().catch(() => {
      throw new Error("Invalid response from server");
    });
  };

  const handleScan = async (targetInput?: string, forceRefresh = false) => {
    const raw = targetInput ?? inputValue;
    const resource = detectResource(raw);
    if (!resource) {
      setError({ key: "invalidInput", type: "warning" });
      return;
    }
    const { type, id } = resource;
    const cacheKey = `${type}:${id}`;
    setError(null);

    if (!forceRefresh && scanCache[cacheKey]) {
      setRuContent(scanCache[cacheKey]);
      setLastInput(raw);
      setFromCache(true);
      setResultTab("tracks");
      setTracksSearch("");
      setArtistsSearch("");
      return;
    }

    setRuContent(null);
    setFromCache(false);
    setLoading(true);
    try {
      const data = await fetchRuContent(type, id);
      setRuContent(data);
      setLastInput(raw);
      setScanCache(prev => ({ ...prev, [cacheKey]: data }));
      setResultTab("tracks");
      setTracksSearch("");
      setArtistsSearch("");
    } catch (e: any) {
      const code = e.code as string | undefined;
      if (code === "ANON_QUOTA_EXCEEDED") {
        setError({ key: "anonQuotaExceeded", type: "auth" });
        return;
      }
      const codeToKey: Record<string, ErrorKey> = {
        PLAYLIST_NOT_FOUND: "notFound",
        NOT_FOUND: "notFound",
        BAD_REQUEST: "badRequest",
        DATABASE_ERROR: "databaseError",
        SPOTIFY_API_ERROR: "spotifyApiError",
        INTERNAL_ERROR: "internalError",
        TOO_MANY_REQUESTS: "tooManyRequests",
        UNAUTHORIZED: "unauthorized",
      };
      const warningCodes = new Set(["PLAYLIST_NOT_FOUND", "NOT_FOUND", "BAD_REQUEST", "TOO_MANY_REQUESTS"]);
      setError({
        key: (code && codeToKey[code]) ? codeToKey[code] : "somethingWentWrong",
        type: (code && warningCodes.has(code)) ? "warning" : "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const ruArtistIds = new Set((ruContent?.Artists ?? []).map((a: Artist) => a.SpotifyID));
  const ruArtistMap = useMemo(
    () => new Map<string, Artist>((ruContent?.Artists ?? []).map((a: Artist) => [a.SpotifyID, a])),
    [ruContent?.Artists],
  );

  type TooltipState = { artist: Artist; x: number; y: number };
  const [tooltip, setTooltip] = useState<TooltipState | null>(null);
  const hideTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => () => { if (hideTimer.current) clearTimeout(hideTimer.current); }, []);

  const showTooltip = (artist: Artist, e: React.MouseEvent) => {
    if (hideTimer.current) clearTimeout(hideTimer.current);
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    setTooltip({ artist, x: rect.left, y: rect.top });
  };

  const hideTooltip = () => {
    hideTimer.current = setTimeout(() => setTooltip(null), 120);
  };

  const filteredTracks = (ruContent?.Tracks ?? []).filter((track: Track) => {
    if (!tracksSearch.trim()) return true;
    const searchLower = tracksSearch.toLowerCase();
    return (
      track.Name.toLowerCase().includes(searchLower) ||
      (track.Artists ?? []).some((a: TrackArtist) => a.Name.toLowerCase().includes(searchLower))
    );
  });

  const filteredTracksMap = useMemo(
    () => new Map(filteredTracks.map((t: Track) => [t.SpotifyID, t])),
    [filteredTracks],
  );

  const artistTrackCount = useMemo(() => {
    const counts = new Map<string, number>();
    for (const track of ruContent?.Tracks ?? []) {
      for (const artist of track.Artists ?? []) {
        counts.set(artist.SpotifyID, (counts.get(artist.SpotifyID) ?? 0) + 1);
      }
    }
    return counts;
  }, [ruContent?.Tracks]);

  const filteredArtists = (ruContent?.Artists ?? []).filter((artist: Artist) => {
    if (!artistsSearch.trim()) return true;
    return artist.Name.toLowerCase().includes(artistsSearch.toLowerCase());
  });

  const filteredArtistsSorted = useMemo(
    () =>
      [...filteredArtists]
        .map((artist: Artist) => ({ artist, trackCount: artistTrackCount.get(artist.SpotifyID) ?? 0, desc: artistDesc(artist, lang) }))
        .sort((a, b) => b.trackCount - a.trackCount),
    [filteredArtists, artistTrackCount, lang],
  );

  const filteredArtistsSortedMap = useMemo(
    () => new Map(filteredArtistsSorted.map(entry => [entry.artist.SpotifyID, entry])),
    [filteredArtistsSorted],
  );

  const hasTracks = (ruContent?.Tracks ?? []).length > 0;
  const artistCount = ruContent?.Artists?.length ?? 0;
  const isContentEmpty = ruContent !== null && !hasTracks && artistCount === 0;

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
      {/* Artist hover tooltip */}
      {tooltip && (() => {
        const a = tooltip.artist;
        const desc = artistDesc(a, lang);
        const isPhonkers = a.Source?.toLowerCase().replace(/\s+/g, "").includes("phonkersbase") || a.Source?.toLowerCase().replace(/\s+/g, "") === "phonkers";
        const sourceHref = isPhonkers ? "https://phonkersbase.com" : a.SourceURL;
        const tooltipW = 280;
        const left = Math.min(tooltip.x, window.innerWidth - tooltipW - 12);
        const top = tooltip.y - 8;
        return (
          <div
            onMouseEnter={() => { if (hideTimer.current) clearTimeout(hideTimer.current); }}
            onMouseLeave={hideTooltip}
            style={{
              position: "fixed",
              left,
              top,
              transform: "translateY(-100%)",
              width: tooltipW,
              background: "rgba(18, 18, 18, 0.97)",
              border: "1px solid rgba(255,255,255,0.12)",
              borderRadius: 10,
              padding: "12px 14px",
              zIndex: 9999,
              boxShadow: "0 8px 32px rgba(0,0,0,0.6)",
              display: "flex",
              flexDirection: "column",
              gap: 8,
              pointerEvents: "auto",
            }}
          >
            <div style={{ display: "flex", alignItems: "center", gap: 6, flexWrap: "wrap" }}>
              <span style={{ fontWeight: 700, fontSize: 14, color: "#fff" }}>{a.Name}</span>
              {a.Confirmed && (
                <span style={{ fontSize: 10, fontWeight: 600, padding: "2px 6px", borderRadius: 4, background: "rgba(231,76,60,0.2)", border: "1px solid rgba(231,76,60,0.4)", color: "#e74c3c" }}>
                  {tx.confirmed}
                </span>
              )}
              {a.Country && (
                <span style={{ fontSize: 12, color: "rgba(255,255,255,0.45)", display: "flex", alignItems: "center", gap: 5 }}>
                  <img src={countryFlag(a.Country)} alt="" style={{ width: 18, height: 14, borderRadius: 2, objectFit: "cover", flexShrink: 0 }} onError={e => { (e.target as HTMLImageElement).style.display = "none"; }} />
                  {a.Country}
                </span>
              )}
            </div>
            {desc && (
              <p style={{ margin: 0, fontSize: 12, color: "rgba(255,255,255,0.65)", lineHeight: 1.5 }}>{desc}</p>
            )}
            {a.Source && (
              <div style={{ fontSize: 11, color: "rgba(255,255,255,0.35)", display: "flex", gap: 4 }}>
                <span>{tx.source}</span>
                {sourceHref ? (
                  <a href={sourceHref} target="_blank" rel="noopener noreferrer" style={{ color: "rgba(255,255,255,0.5)", textDecoration: "underline", textUnderlineOffset: 2 }}>
                    {a.Source}
                  </a>
                ) : (
                  <span style={{ color: "rgba(255,255,255,0.5)" }}>{a.Source}</span>
                )}
              </div>
            )}
          </div>
        );
      })()}

      {/* Scan Controls */}
      <div style={cardStyle}>
        <h3 style={{ color: "#fff", marginBottom: 16, fontSize: "1.1rem" }}>{tx.scan}</h3>
        {detected && (
          <div style={{
            display: "inline-flex",
            alignItems: "center",
            marginBottom: 8,
            padding: "3px 10px",
            background: "rgba(255, 255, 255, 0.08)",
            border: "1px solid rgba(255, 255, 255, 0.15)",
            borderRadius: 6,
            fontSize: 12,
            color: "rgba(255, 255, 255, 0.6)",
            animation: "fadeIn 0.2s ease",
          }}>
            {tx.resourceType[detected.type]}
          </div>
        )}
        <input
          type="text"
          placeholder={tx.placeholder}
          value={inputValue}
          onChange={e => setInputValue((e.target as HTMLInputElement).value)}
          onPaste={e => { e.preventDefault(); setInputValue(e.clipboardData.getData("text")); }}
          style={{
            ...inputStyle,
            width: "100%",
            marginBottom: 12,
            marginTop: detected ? 8 : 0,
            borderColor: detected ? "#1DB954" : "rgba(255, 255, 255, 0.2)",
            transition: "border-color 0.2s ease",
          }}
          disabled={loading}
        />
        <button
          onClick={() => handleScan()}
          disabled={loading || !detected}
          style={{
            ...buttonStyle,
            width: "100%",
            opacity: loading || !detected ? 0.5 : 1,
            cursor: loading || !detected ? "not-allowed" : "pointer",
          }}
        >
          {loading ? tx.scanning : tx.scan}
        </button>
      </div>

      {/* Error / Warning / Auth prompt */}
      {error && (
        <div style={{
          ...cardStyle,
          background: error.type === "warning"
            ? "rgba(241, 196, 15, 0.1)"
            : error.type === "auth"
            ? "rgba(29, 185, 84, 0.08)"
            : "rgba(231, 76, 60, 0.1)",
          border: error.type === "warning"
            ? "1px solid rgba(241, 196, 15, 0.3)"
            : error.type === "auth"
            ? "1px solid rgba(29, 185, 84, 0.3)"
            : "1px solid rgba(231, 76, 60, 0.3)",
          color: error.type === "warning" ? "#f1c40f" : error.type === "auth" ? "#1DB954" : "#e74c3c",
          textAlign: "center",
        }}>
          <p style={{ margin: 0 }}>{tx[error.key]}</p>
          {error.key === "notFound" && (
            <ul style={{ margin: "12px 0 0", paddingLeft: 20, textAlign: "left", display: "flex", flexDirection: "column", gap: 6 }}>
              {tx.notFoundHints.map((hint, i) => (
                <li key={i} style={{ lineHeight: 1.5 }}>{hint}</li>
              ))}
            </ul>
          )}
          {error.type === "auth" && (
            <div style={{ marginTop: 16, display: "flex", justifyContent: "center" }}>
              <button
                onClick={() => navigate("/signup")}
                style={buttonStyle}
                onMouseEnter={e => { e.currentTarget.style.opacity = "0.85"; }}
                onMouseLeave={e => { e.currentTarget.style.opacity = "1"; }}
              >
                {tx.createAccount}
              </button>
            </div>
          )}
        </div>
      )}

      {/* Cache Info Banner */}
      {fromCache && !loading && ruContent && (
        <div style={{
          ...cardStyle,
          background: "rgba(52, 152, 219, 0.1)",
          border: "1px solid rgba(52, 152, 219, 0.3)",
          color: "#3498db",
          textAlign: "center",
        }}>
          {tx.showingCachedResults}{" "}
          <span
            onClick={() => handleScan(lastInput ?? undefined, true)}
            style={{ textDecoration: "underline", cursor: "pointer", fontWeight: 600 }}
          >
            {tx.rescan}
          </span>
        </div>
      )}

      {/* Loading spinner (fresh scan) */}
      {loading && !ruContent && (
        <div style={{ ...cardStyle, display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", minHeight: 300, gap: 20 }}>
          <style>{`
            @keyframes l1 { to { transform: rotate(0.5turn); } }
            .loader { width: 50px; aspect-ratio: 1; border-radius: 50%; border: 8px solid; border-color: #1DB954 #0000; animation: l1 1s infinite; }
          `}</style>
          <div className="loader" />
          <p style={{ color: "rgba(255, 255, 255, 0.6)", margin: 0 }}>{tx.scanning}</p>
        </div>
      )}

      {/* Results */}
      {ruContent && (
        <div style={cardStyle}>
          {/* Loading spinner (rescan) */}
          {loading && (
            <div style={{ display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", minHeight: 300, gap: 20 }}>
              <div className="loader" />
              <p style={{ color: "rgba(255, 255, 255, 0.6)", margin: 0 }}>{tx.scanning}</p>
            </div>
          )}

          {!loading && isContentEmpty && (
            <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "40px 0", fontSize: 16 }}>
              {tx.contentClear}
            </div>
          )}

          {!loading && !isContentEmpty && (
            <>
              {/* Tab Navigation - only when tracks exist */}
              {hasTracks && <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 20, borderBottom: "1px solid rgba(255, 255, 255, 0.1)" }}>
                <div style={{ display: "flex", gap: 12 }}>
                <button
                  onClick={() => setResultTab("tracks")}
                  style={{
                    padding: "12px 16px",
                    background: "transparent",
                    border: "none",
                    borderBottom: resultTab === "tracks" ? "2px solid #1DB954" : "none",
                    color: resultTab === "tracks" ? "#1DB954" : "rgba(255, 255, 255, 0.6)",
                    cursor: "pointer",
                    fontSize: 14,
                    fontWeight: 500,
                    transition: "all 0.2s ease",
                  }}
                >
                  {tx.tracksTab((ruContent?.Tracks ?? []).length)}
                </button>
                <button
                  onClick={() => setResultTab("artists")}
                  style={{
                    padding: "12px 16px",
                    background: "transparent",
                    border: "none",
                    borderBottom: resultTab === "artists" ? "2px solid #1DB954" : "none",
                    color: resultTab === "artists" ? "#1DB954" : "rgba(255, 255, 255, 0.6)",
                    cursor: "pointer",
                    fontSize: 14,
                    fontWeight: 500,
                    transition: "all 0.2s ease",
                  }}
                >
                  {tx.artistsTab(ruContent?.Artists?.length ?? 0)}
                </button>
                </div>
                <div style={{ display: "flex", alignItems: "center", gap: 6, paddingBottom: 8, opacity: 0.45 }}>
                  <img src="/spotify.svg" alt="Spotify" style={{ width: 16, height: 16 }} />
                  <span style={{ fontSize: 11, color: "#fff", whiteSpace: "nowrap" }}>{tx.dataProvidedBySpotify}</span>
                </div>
              </div>}

              {/* Tracks Tab */}
              {hasTracks && resultTab === "tracks" && (
                <div>
                  {(ruContent?.Tracks ?? []).length > 0 && (
                    <input
                      type="text"
                      placeholder={tx.searchTracks}
                      value={tracksSearch}
                      onChange={e => setTracksSearch(e.target.value)}
                      style={{ ...inputStyle, width: "100%", marginBottom: 12 }}
                    />
                  )}
                  <h3 style={{ color: "#fff", margin: "0 0 16px 0", fontSize: "1.1rem" }}>
                    {tx.tracksWithRussianArtists(filteredTracks.length)}
                  </h3>
                  <AnimatedList
                    items={filteredTracks.map((t: Track) => t.SpotifyID)}
                    showGradients={false}
                    displayScrollbar={true}
                    enableArrowNavigation={true}
                    children={(trackId: string) => {
                      const track = filteredTracksMap.get(trackId);
                      if (!track) return null;
                      return (
                        <div style={{
                          padding: 16,
                          background: "rgba(255, 255, 255, 0.03)",
                          border: "1px solid rgba(255, 255, 255, 0.05)",
                          borderRadius: 10,
                          transition: "all 0.2s ease",
                          display: "flex",
                          alignItems: "center",
                          gap: 12,
                        }}>
                          {track.ImageURL && (
                            <img
                              src={track.ImageURL}
                              alt={track.Name}
                              style={{ width: 48, height: 48, borderRadius: 4, objectFit: "cover", flexShrink: 0 }}
                              onError={(e) => { (e.target as HTMLImageElement).style.display = "none"; }}
                            />
                          )}
                          <div style={{ flex: 1, minWidth: 0 }}>
                            <div style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                              <a
                                href={`https://open.spotify.com/track/${track.SpotifyID}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                style={{ fontWeight: 600, color: "#fff", fontSize: 15, textDecoration: "none", transition: "color 0.2s ease" }}
                                onMouseEnter={e => { e.currentTarget.style.color = "#1DB954"; }}
                                onMouseLeave={e => { e.currentTarget.style.color = "#fff"; }}
                              >
                                {track.Name}
                              </a>
                            </div>
                            <div style={{ marginTop: 4, fontSize: 13 }}>
                              {(track.Artists ?? []).map((artist: TrackArtist, idx: number) => (
                                <span key={artist.SpotifyID}>
                                  {ruArtistIds.has(artist.SpotifyID) ? (
                                    <a
                                      style={{ color: "#e74c3c", fontWeight: 600, textDecoration: "underline", textUnderlineOffset: 2, cursor: "pointer", transition: "color 0.2s ease" }}
                                      onMouseEnter={e => { e.currentTarget.style.color = "#ff6b5a"; const full = ruArtistMap.get(artist.SpotifyID); if (full) showTooltip(full, e); }}
                                      onMouseLeave={e => { e.currentTarget.style.color = "#e74c3c"; hideTooltip(); }}
                                      onClick={() => { setResultTab("artists"); setArtistsSearch(artist.Name); setTooltip(null); }}
                                    >
                                      {artist.Name}
                                    </a>
                                  ) : (
                                    <a
                                      href={`https://open.spotify.com/artist/${artist.SpotifyID}`}
                                      target="_blank"
                                      rel="noopener noreferrer"
                                      style={{ color: "rgba(255,255,255,0.6)", textDecoration: "none", cursor: "pointer", transition: "color 0.2s ease" }}
                                      onMouseEnter={e => { e.currentTarget.style.color = "rgba(255,255,255,0.9)"; }}
                                      onMouseLeave={e => { e.currentTarget.style.color = "rgba(255,255,255,0.6)"; }}
                                    >
                                      {artist.Name}
                                    </a>
                                  )}
                                  {idx < (track.Artists ?? []).length - 1 && (
                                    <span style={{ color: "rgba(255,255,255,0.3)", margin: "0 6px" }}>•</span>
                                  )}
                                </span>
                              ))}
                            </div>
                          </div>
                        </div>
                      );
                    }}
                  />
                  {filteredTracks.length === 0 && (ruContent?.Tracks ?? []).length === 0 && (
                    <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "40px 0" }}>
                      {tx.noRussianTracks}
                    </div>
                  )}
                  {filteredTracks.length === 0 && (ruContent?.Tracks ?? []).length > 0 && (
                    <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "40px 0" }}>
                      {tx.noTracksMatch}
                    </div>
                  )}
                </div>
              )}

              {/* Artists Tab */}
              {(!hasTracks || resultTab === "artists") && (
                <div>
                  {artistCount > 1 && (
                    <input
                      type="text"
                      placeholder={tx.searchArtists}
                      value={artistsSearch}
                      onChange={e => setArtistsSearch(e.target.value)}
                      style={{ ...inputStyle, width: "100%", marginBottom: 12 }}
                    />
                  )}
                  <h3 style={{ color: "#fff", margin: "0 0 16px 0", fontSize: "1.1rem" }}>
                    {tx.russianArtistsFound(filteredArtists.length)}
                  </h3>
                  <AnimatedList
                    items={filteredArtistsSorted.map(({ artist }) => artist.SpotifyID)}
                    showGradients={false}
                    displayScrollbar={true}
                    enableArrowNavigation={true}
                    children={(artistId: string) => {
                      const entry = filteredArtistsSortedMap.get(artistId);
                      if (!entry) return null;
                      const { artist, trackCount, desc } = entry;
                      return (
                        <div style={{
                          padding: 16,
                          background: "rgba(255, 255, 255, 0.03)",
                          border: "1px solid rgba(255, 255, 255, 0.05)",
                          borderRadius: 10,
                          display: "flex",
                          flexDirection: "column",
                          gap: 10,
                        }}>
                          {/* Header row: name + track count */}
                          <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 12 }}>
                            <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
                              <a
                                href={`https://open.spotify.com/artist/${artist.SpotifyID}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                style={{ color: "#fff", textDecoration: "none", fontWeight: 600, fontSize: 16, cursor: "pointer", transition: "all 0.2s ease", display: "flex", alignItems: "center", gap: 6 }}
                                onMouseEnter={e => { e.currentTarget.style.color = "#1DB954"; e.currentTarget.style.textDecoration = "underline"; }}
                                onMouseLeave={e => { e.currentTarget.style.color = "#fff"; e.currentTarget.style.textDecoration = "none"; }}
                              >
                                {artist.Name}
                                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ flexShrink: 0 }}>
                                  <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
                                  <polyline points="15 3 21 3 21 9"></polyline>
                                  <line x1="10" y1="14" x2="21" y2="3"></line>
                                </svg>
                              </a>
                              {artist.Confirmed && (
                                <span style={{ fontSize: 11, fontWeight: 600, padding: "2px 7px", borderRadius: 4, background: "rgba(231, 76, 60, 0.2)", border: "1px solid rgba(231, 76, 60, 0.4)", color: "#e74c3c", whiteSpace: "nowrap" }}>
                                  {tx.confirmed}
                                </span>
                              )}
                              {artist.Country && (
                                <span style={{ fontSize: 12, color: "rgba(255,255,255,0.45)", whiteSpace: "nowrap", display: "flex", alignItems: "center", gap: 5 }}>
                                  <img src={countryFlag(artist.Country)} alt="" style={{ width: 18, height: 14, borderRadius: 2, objectFit: "cover", flexShrink: 0 }} onError={e => { (e.target as HTMLImageElement).style.display = "none"; }} />
                                  {artist.Country}
                                </span>
                              )}
                            </div>
                            {hasTracks && (
                              <button
                                onClick={() => { setResultTab("tracks"); setTracksSearch(artist.Name); }}
                                style={{
                                  padding: "6px 12px",
                                  background: "rgba(29, 185, 84, 0.2)",
                                  border: "1px solid rgba(29, 185, 84, 0.4)",
                                  borderRadius: 6,
                                  color: "#1DB954",
                                  cursor: "pointer",
                                  fontSize: 12,
                                  fontWeight: 500,
                                  whiteSpace: "nowrap",
                                  transition: "all 0.2s ease",
                                  flexShrink: 0,
                                }}
                                onMouseEnter={e => { e.currentTarget.style.background = "rgba(29, 185, 84, 0.3)"; }}
                                onMouseLeave={e => { e.currentTarget.style.background = "rgba(29, 185, 84, 0.2)"; }}
                              >
                                {tx.trackCount(trackCount)}
                              </button>
                            )}
                          </div>

                          {/* Description */}
                          {desc && <p style={{ margin: 0, fontSize: 13, color: "rgba(255,255,255,0.65)", lineHeight: 1.5 }}>{desc}</p>}

                          {/* Source */}
                          {artist.Source && (
                            <div style={{ fontSize: 12, color: "rgba(255,255,255,0.4)", display: "flex", alignItems: "center", gap: 4 }}>
                              <span>{tx.source}</span>
                              {(() => {
                                const isPhonkers = artist.Source.toLowerCase().replace(/\s+/g, "").includes("phonkersbase") || artist.Source.toLowerCase().replace(/\s+/g, "") === "phonkers";
                                const href = isPhonkers ? "https://phonkersbase.com" : artist.SourceURL;
                                return href ? (
                                  <a
                                    href={href}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    style={{ color: "rgba(255,255,255,0.55)", textDecoration: "underline", textUnderlineOffset: 2 }}
                                  >
                                    {artist.Source}
                                  </a>
                                ) : (
                                  <span style={{ color: "rgba(255,255,255,0.55)" }}>{artist.Source}</span>
                                );
                              })()}
                            </div>
                          )}
                        </div>
                      );
                    }}
                  />
                  {filteredArtists.length === 0 && (ruContent?.Artists ?? []).length === 0 && (
                    <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "40px 0" }}>{tx.noArtistsFound}</div>
                  )}
                  {filteredArtists.length === 0 && (ruContent?.Artists ?? []).length > 0 && (
                    <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "40px 0" }}>{tx.noArtistsMatch}</div>
                  )}
                </div>
              )}
            </>
          )}
        </div>
      )}

      {/* Empty state */}
      {!ruContent && !loading && !error && (
        <div style={{ ...cardStyle, display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", minHeight: 300, color: "rgba(255,255,255,0.5)" }}>
          <svg width="64" height="64" viewBox="0 0 24 24" fill="currentColor" style={{ opacity: 0.3, marginBottom: 16 }}>
            <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z" />
          </svg>
          <p style={{ fontSize: 16, marginBottom: 8, margin: "0 0 8px 0" }}>{tx.noScanResults}</p>
          <p style={{ fontSize: 13, opacity: 0.7, margin: 0 }}>{tx.noScanResultsHint}</p>
        </div>
      )}
    </div>
  );
}
