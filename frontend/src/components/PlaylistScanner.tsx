import * as React from "react";
import { useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import AnimatedList from "./react-bits/AnimatedList";
import type { Artist, Track, RuContent } from "../types/models";

const SPOTIFY_PLAYLIST_BASE = "https://open.spotify.com/playlist/";
const BASE62_ID_REGEX = /^[A-Za-z0-9]{22}$/;

const extractPlaylistId = (input: string): string => {
  if (BASE62_ID_REGEX.test(input.trim())) return input.trim();
  const match = /(?:playlist\/|spotify:playlist:)([A-Za-z0-9]{22})/.exec(input);
  return match ? match[1] : input.trim();
};

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

type ErrorState = { message: string; type: "warning" | "error" | "auth" };

export default function PlaylistScanner() {
  const navigate = useNavigate();
  const [playlistId, setPlaylistId] = useState("");
  const [ruContent, setRuContent] = useState<RuContent | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<ErrorState | null>(null);
  const [lastPlaylistId, setLastPlaylistId] = useState<string | null>(null);
  const [resultTab, setResultTab] = useState<"tracks" | "artists">("tracks");
  const [tracksSearch, setTracksSearch] = useState("");
  const [artistsSearch, setArtistsSearch] = useState("");
  const [scanCache, setScanCache] = useState<Record<string, RuContent>>({});
  const [fromCache, setFromCache] = useState(false);

  const handlePlaylistInput = (value: string) => {
    setPlaylistId(extractPlaylistId(value));
  };

  const fetchPlaylistRuContent = async (id: string): Promise<RuContent> => {
    if (!id || !BASE62_ID_REGEX.test(id)) throw new Error("Invalid playlist ID");
    const response = await fetch(`/api/playlist/${encodeURIComponent(id)}/rucontent`);
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

  const handleScanPlaylist = async (targetId?: string, forceRefresh = false) => {
    const id = targetId || playlistId;
    if (!id || !BASE62_ID_REGEX.test(id)) {
      setError({ message: "Invalid playlist ID. Check the URL or ID and try again.", type: "warning" });
      return;
    }
    setError(null);

    if (!forceRefresh && scanCache[id]) {
      setRuContent(scanCache[id]);
      setLastPlaylistId(id);
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
      const data = await fetchPlaylistRuContent(id);
      setRuContent(data);
      setLastPlaylistId(id);
      setScanCache(prev => ({ ...prev, [id]: data }));
      setResultTab("tracks");
      setTracksSearch("");
      setArtistsSearch("");
    } catch (e: any) {
      const code = e.code as string | undefined;
      if (code === "ANON_QUOTA_EXCEEDED") {
        setError({ message: "You've used your trial scans. Sign in to keep scanning, it's free!", type: "auth" });
        return;
      }
      const messages: Record<string, string> = {
        PLAYLIST_NOT_FOUND: "Playlist not found. Check the URL or ID and try again.",
        BAD_REQUEST: "Invalid request. Please check your input.",
        DATABASE_ERROR: "A server error occurred. Please try again later.",
        SPOTIFY_API_ERROR: "Failed to communicate with Spotify. Please try again later.",
        INTERNAL_ERROR: "An unexpected error occurred. Please try again later.",
      };
      const isWarning = code === "PLAYLIST_NOT_FOUND" || code === "BAD_REQUEST";
      setError({
        message: code && messages[code] ? messages[code] : (e.message || "Something went wrong"),
        type: isWarning ? "warning" : "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const ruArtistIds = new Set((ruContent?.Artists ?? []).map((a: Artist) => a.ID));

  const filteredTracks = (ruContent?.Tracks ?? []).filter((track: Track) => {
    if (!tracksSearch.trim()) return true;
    const searchLower = tracksSearch.toLowerCase();
    return (
      track.Name.toLowerCase().includes(searchLower) ||
      (track.Artists ?? []).some((a: Artist) => a.Name.toLowerCase().includes(searchLower))
    );
  });

  const filteredTracksMap = useMemo(
    () => new Map(filteredTracks.map((t: Track) => [t.ID, t])),
    [filteredTracks],
  );

  const artistTrackCount = useMemo(() => {
    const counts = new Map<string, number>();
    for (const track of ruContent?.Tracks ?? []) {
      for (const artist of track.Artists ?? []) {
        counts.set(artist.ID, (counts.get(artist.ID) ?? 0) + 1);
      }
    }
    return counts;
  }, [ruContent?.Tracks]);

  const filteredArtists = (ruContent?.Artists ?? []).filter((artist: Artist) => {
    if (!artistsSearch.trim()) return true;
    return artist.Name.toLowerCase().includes(artistsSearch.toLowerCase());
  });

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
      {/* Scan Controls */}
      <div style={cardStyle}>
        <h3 style={{ color: "#fff", marginBottom: 16, fontSize: "1.1rem" }}>Scan Playlist</h3>
        {playlistId && (
          <div style={{ color: "rgba(255, 255, 255, 0.5)", fontSize: 12, marginBottom: 6, animation: "fadeIn 0.2s ease" }}>
            {SPOTIFY_PLAYLIST_BASE}
          </div>
        )}
        <input
          type="text"
          placeholder="Paste playlist URL or ID..."
          value={playlistId}
          onChange={e => handlePlaylistInput((e.target as HTMLInputElement).value)}
          style={{
            ...inputStyle,
            width: "100%",
            marginBottom: 12,
            borderColor: playlistId ? "#1DB954" : "rgba(255, 255, 255, 0.2)",
            transition: "border-color 0.2s ease",
          }}
          disabled={loading}
        />
        <button
          onClick={() => handleScanPlaylist()}
          disabled={loading || !playlistId}
          style={{
            ...buttonStyle,
            width: "100%",
            opacity: loading || !playlistId ? 0.5 : 1,
            cursor: loading || !playlistId ? "not-allowed" : "pointer",
          }}
        >
          {loading ? "Scanning..." : "Scan Playlist"}
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
          <p style={{ margin: 0 }}>{error.message}</p>
          {error.type === "auth" && (
            <div style={{ marginTop: 16, display: "flex", justifyContent: "center" }}>
              <button
                onClick={() => navigate("/signup")}
                style={buttonStyle}
                onMouseEnter={e => { e.currentTarget.style.opacity = "0.85"; }}
                onMouseLeave={e => { e.currentTarget.style.opacity = "1"; }}
              >
                Create account
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
          Showing results from a previous scan.{" "}
          <span
            onClick={() => handleScanPlaylist(lastPlaylistId ?? undefined, true)}
            style={{ textDecoration: "underline", cursor: "pointer", fontWeight: 600 }}
          >
            Rescan?
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
          <p style={{ color: "rgba(255, 255, 255, 0.6)", margin: 0 }}>Scanning...</p>
        </div>
      )}

      {/* Results */}
      {ruContent && (
        <div style={cardStyle}>
          {/* Loading spinner (rescan) */}
          {loading && (
            <div style={{ display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", minHeight: 300, gap: 20 }}>
              <div className="loader" />
              <p style={{ color: "rgba(255, 255, 255, 0.6)", margin: 0 }}>Scanning...</p>
            </div>
          )}

          {!loading && (
            <>
              {/* Tab Navigation */}
              <div style={{ display: "flex", gap: 12, marginBottom: 20, borderBottom: "1px solid rgba(255, 255, 255, 0.1)" }}>
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
                  Tracks ({(ruContent?.Tracks ?? []).length})
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
                  Artists ({ruContent?.Artists?.length ?? 0})
                </button>
              </div>

              {/* Tracks Tab */}
              {resultTab === "tracks" && (
                <div>
                  {(ruContent?.Tracks ?? []).length > 0 && (
                    <input
                      type="text"
                      placeholder="Search tracks..."
                      value={tracksSearch}
                      onChange={e => setTracksSearch(e.target.value)}
                      style={{ ...inputStyle, width: "100%", marginBottom: 12 }}
                    />
                  )}
                  <h3 style={{ color: "#fff", margin: "0 0 16px 0", fontSize: "1.1rem" }}>
                    Tracks with Russian Artists ({filteredTracks.length})
                  </h3>
                  <AnimatedList
                    items={filteredTracks.map((t: Track) => t.ID)}
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
                            <div style={{ fontWeight: 600, color: "#fff", fontSize: 15, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                              {track.Name}
                            </div>
                            <div style={{ marginTop: 4, fontSize: 13 }}>
                              {(track.Artists ?? []).map((artist: Artist, idx: number) => (
                                <span key={artist.ID}>
                                  {ruArtistIds.has(artist.ID) ? (
                                    <a
                                      href={artist.URL}
                                      target="_blank"
                                      rel="noopener noreferrer"
                                      style={{ color: "#e74c3c", fontWeight: 600, textDecoration: "underline", textUnderlineOffset: 2, cursor: "pointer", transition: "color 0.2s ease" }}
                                      onMouseEnter={e => { e.currentTarget.style.color = "#ff6b5a"; }}
                                      onMouseLeave={e => { e.currentTarget.style.color = "#e74c3c"; }}
                                    >
                                      {artist.Name}
                                    </a>
                                  ) : (
                                    <a
                                      href={artist.URL}
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
                      No Russian tracks found. This playlist is clean!
                    </div>
                  )}
                  {filteredTracks.length === 0 && (ruContent?.Tracks ?? []).length > 0 && (
                    <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "40px 0" }}>
                      No tracks match your search
                    </div>
                  )}
                </div>
              )}

              {/* Artists Tab */}
              {resultTab === "artists" && (
                <div>
                  {(ruContent?.Artists ?? []).length > 0 && (
                    <input
                      type="text"
                      placeholder="Search artists..."
                      value={artistsSearch}
                      onChange={e => setArtistsSearch(e.target.value)}
                      style={{ ...inputStyle, width: "100%", marginBottom: 12 }}
                    />
                  )}
                  <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                    {filteredArtists.length === 0 && (ruContent?.Artists ?? []).length === 0 && (
                      <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "40px 0" }}>No artists found</div>
                    )}
                    {filteredArtists.length === 0 && (ruContent?.Artists ?? []).length > 0 && (
                      <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "40px 0" }}>No artists match your search</div>
                    )}
                    {filteredArtists
                      .map((artist: Artist) => ({
                        artist,
                        trackCount: artistTrackCount.get(artist.ID) ?? 0,
                      }))
                      .sort((a, b) => b.trackCount - a.trackCount)
                      .map(({ artist, trackCount }) => (
                        <div
                          key={artist.ID}
                          style={{
                            padding: 16,
                            background: "rgba(255, 255, 255, 0.03)",
                            border: "1px solid rgba(255, 255, 255, 0.05)",
                            borderRadius: 10,
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "space-between",
                            gap: 12,
                          }}
                        >
                          <a
                            href={artist.URL}
                            target="_blank"
                            rel="noopener noreferrer"
                            style={{ flex: 1, color: "#fff", textDecoration: "none", fontWeight: 600, fontSize: 16, cursor: "pointer", transition: "all 0.2s ease", display: "flex", alignItems: "center", gap: 8 }}
                            onMouseEnter={e => { e.currentTarget.style.color = "#1DB954"; e.currentTarget.style.textDecoration = "underline"; }}
                            onMouseLeave={e => { e.currentTarget.style.color = "#fff"; e.currentTarget.style.textDecoration = "none"; }}
                          >
                            {artist.Name}
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ flexShrink: 0 }}>
                              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
                              <polyline points="15 3 21 3 21 9"></polyline>
                              <line x1="10" y1="14" x2="21" y2="3"></line>
                            </svg>
                          </a>
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
                            }}
                            onMouseEnter={e => { e.currentTarget.style.background = "rgba(29, 185, 84, 0.3)"; }}
                            onMouseLeave={e => { e.currentTarget.style.background = "rgba(29, 185, 84, 0.2)"; }}
                          >
                            {trackCount} track{trackCount !== 1 ? "s" : ""}
                          </button>
                        </div>
                      ))}
                  </div>
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
          <p style={{ fontSize: 16, marginBottom: 8, margin: "0 0 8px 0" }}>No scan results yet</p>
          <p style={{ fontSize: 13, opacity: 0.7, margin: 0 }}>Enter a playlist ID or URL to get started</p>
        </div>
      )}
    </div>
  );
}
