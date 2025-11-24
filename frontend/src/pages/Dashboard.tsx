import * as React from "react";
import { useState } from "react";
import Aurora from "../components/Aurora";
import AnimatedList from "../components/AnimatedList";
import type { Artist, Track, Playlist, RuContent } from "../types/models";
// Ensure JSX namespace is available for TS
/// <reference types="react/next" />


const SPOTIFY_PLAYLIST_BASE = "https://open.spotify.com/playlist/";

// Helper to extract playlist ID from various formats
const extractPlaylistId = (input: string): string => {
  // If it's already just an ID (alphanumeric, typically 22 chars)
  if (/^[A-Za-z0-9]{22}$/.test(input.trim())) {
    return input.trim();
  }
  // Extract from full URL or URI
  const match = /(?:playlist\/|spotify:playlist:)([A-Za-z0-9]+)/.exec(input);
  return match ? match[1] : input.trim();
};

const Dashboard = () => {
  const [playlistId, setPlaylistId] = useState("");
  const [ruContent, setRuContent] = useState<RuContent | null>(null);
  const [selected, setSelected] = useState(new Set<string>());
  const [loading, setLoading] = useState<null | string>(null); // null | 'playlist' | 'liked'
  const [error, setError] = useState<string | null>(null);
  const [lastPlaylistId, setLastPlaylistId] = useState<string | null>(null);
  const [userPlaylists, setUserPlaylists] = useState<Playlist[]>([]);
  const [playlistsLoading, setPlaylistsLoading] = useState(true);
  const [playlistSearch, setPlaylistSearch] = useState("");

  // No longer needed: selectedPlaylist

  // Fetch user's top playlists on mount
  React.useEffect(() => {
    const fetchPlaylists = async () => {
      try {
        const response = await fetch("/api/user/playlists", { credentials: "include" });
        if (!response.ok) {
          throw new Error(`Failed to fetch playlists: ${response.status}`);
        }
        const data: Playlist[] = await response.json();
        setUserPlaylists(data);
      } catch (e: any) {
        console.error("Error fetching playlists:", e.message);
        setUserPlaylists([]);
      } finally {
        setPlaylistsLoading(false);
      }
    };
    fetchPlaylists();
  }, []);

  // Handle input change - auto-extract ID from pasted URLs
  const handlePlaylistInput = (value: string) => {
    const extracted = extractPlaylistId(value);
    setPlaylistId(extracted);
  };

  // Filter playlists based on search term
  const filteredPlaylists = userPlaylists.filter(playlist =>
    playlist.Name.toLowerCase().includes(playlistSearch.toLowerCase())
  );

  // Real fetch for playlist scan
  const fetchPlaylistRuContent = async (id: string): Promise<RuContent> => {
    setError(null);
    setLoading("playlist");
    if (!id || !/^[A-Za-z0-9]+$/.test(id)) {
      throw new Error("Invalid playlist ID format");
    }
    const response = await fetch(`/api/playlist/${encodeURIComponent(id)}/rucontent`);
    if (!response.ok) {
      throw new Error(`HTTP error status: ${response.status}`);
    }
    return await response.json();
  };

  const handleScanPlaylist = async () => {
    if (!playlistId || !/^[A-Za-z0-9]+$/.test(playlistId)) {
      setError("Invalid playlist ID format");
      return;
    }
    if (playlistId === lastPlaylistId) {
      setError("This playlist has already been scanned.");
      return;
    }
    try {
      const data = await fetchPlaylistRuContent(playlistId);
      setRuContent(data);
      setSelected(new Set());
      setLastPlaylistId(playlistId);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(null);
    }
  };
  const handleScanLiked = async () => {
    setError(null);
    setLoading("liked");
    try {
      const response = await fetch("/api/user/liked-songs/rucontent");
      if (!response.ok) throw new Error(`HTTP error status: ${response.status}`);
      const data: RuContent = await response.json();
      setRuContent(data);
      setSelected(new Set());
      setLastPlaylistId(null);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(null);
    }
  };

  // Track selection logic
  const handleTrackToggle = (id: string) => {
    setSelected((prev: Set<string>) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };
  const handleTrackSelectAll = (checked: boolean) => {
    if (ruContent && checked) setSelected(new Set((ruContent.Tracks ?? []).map((t: Track) => t.ID)));
    else setSelected(new Set());
  };

  // Delete selected tracks from playlist or liked songs
  const handleDeleteTracks = async (from: 'playlist' | 'liked') => {
    setLoading('delete');
    setError(null);
    try {
      let endpoint = '';
      let bodyTracks: Track[] = [];
      if (from === 'playlist') {
        if (!playlistId) throw new Error('No playlist selected');
        endpoint = `/api/playlist/${encodeURIComponent(playlistId)}/rucontent`;
      } else {
        endpoint = '/api/user/liked-songs/rucontent';
      }
      // Find selected tracks
      bodyTracks = (ruContent?.Tracks ?? []).filter((t: Track) => selected.has(t.ID));
      const response = await fetch(endpoint, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(bodyTracks),
      });
      if (!response.ok) throw new Error(`HTTP error status: ${response.status}`);
      // Remove deleted tracks from UI
      if (ruContent) {
        const filtered = (ruContent.Tracks ?? []).filter((t: Track) => !selected.has(t.ID));
        setRuContent({ ...ruContent, Tracks: filtered });
      }
      setSelected(new Set());
    } catch (e: any) {
      setError(e.message || 'Failed to delete selected tracks.');
    } finally {
      setLoading(null);
    }
  };

  // Helper: set of Russian artist IDs for quick lookup
  const ruArtistIds = new Set((ruContent?.Artists ?? []).map((a: Artist) => a.ID));

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

  const deleteButtonStyle: React.CSSProperties = {
    ...buttonStyle,
    background: "#e74c3c",
    color: "#fff",
  };

  return (
    <div style={{ position: "relative", minHeight: "100vh", width: "100%", overflow: "hidden" }}>
      {/* Sign Out Button - fixed top right */}
      <button
        onClick={async () => {
          await fetch("/api/signout", { method: "POST", credentials: "include" });
          window.location.href = "/";
        }}
        style={{
          position: "fixed",
          top: 12,
          right: 16,
          zIndex: 10,
          padding: "8px 16px",
          background: "#caffdcff",
          color: "#000",
          border: "none",
          borderRadius: 50,
          fontWeight: 600,
          fontSize: 12,
          cursor: "pointer",
          transition: "all 0.2s ease",
          boxShadow: "0 2px 8px 0 rgba(0,0,0,0.08)",
        }}
      >
        Sign Out
      </button>
      {/* Aurora Background */}
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

      {/* Content */}
      <div style={{
        position: "relative",
        zIndex: 1,
        width: "100%",
        height: "100%",
        padding: "40px 24px",
        boxSizing: "border-box",
        overflow: "auto",
      }}>
        <h1 style={{
          fontSize: "2.5rem",
          fontFamily: "'Outfit', sans-serif",
          fontWeight: 800,
          marginBottom: 32,
          textAlign: "center",
          color: "#fff",
        }}>
          Music Dashboard
        </h1>

        {/* Two Column Layout */}
        <div style={{
          display: "grid",
          gridTemplateColumns: "340px 1fr",
          gap: 24,
          alignItems: "start",
        }}>
          {/* Left Column - Controls */}
          <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
            {/* Scan Controls */}
            <div style={cardStyle}>
              <h3 style={{ color: "#fff", marginBottom: 16, fontSize: "1.1rem" }}>Scan Playlist</h3>

              {/* Show selected playlist name */}
              {playlistId && (() => {
                const selectedPlaylist = userPlaylists.find(p => p.ID === playlistId);
                return selectedPlaylist ? (
                  <div style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 10,
                    padding: "10px 12px",
                    marginBottom: 12,
                    background: "rgba(29, 185, 84, 0.15)",
                    border: "1px solid rgba(29, 185, 84, 0.3)",
                    borderRadius: 8,
                    animation: "fadeIn 0.3s ease",
                  }}>
                    {selectedPlaylist.ImageURL ? (
                      <img
                        src={selectedPlaylist.ImageURL}
                        alt={selectedPlaylist.Name}
                        style={{ width: 48, height: 48, borderRadius: 4, flexShrink: 0 }}
                      />
                    ) : (
                      <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="#666" style={{ borderRadius: 4, flexShrink: 0 }}>
                        <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/>
                      </svg>
                    )}
                    <div style={{ flex: 1, overflow: "hidden" }}>
                      <div style={{ color: "#1DB954", fontSize: 12, fontWeight: 500 }}>Selected</div>
                      <div style={{ color: "#fff", fontSize: 13, fontWeight: 600, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                        {selectedPlaylist.Name}
                      </div>
                    </div>
                    <button
                      onClick={() => setPlaylistId("")}
                      style={{
                        background: "transparent",
                        border: "none",
                        color: "rgba(255,255,255,0.5)",
                        cursor: "pointer",
                        padding: 4,
                        fontSize: 18,
                        lineHeight: 1,
                      }}
                    >
                      ×
                    </button>
                  </div>
                ) : null;
              })()}

              {/* URL prefix - shows only when ID is entered */}
              {playlistId && (
                <div style={{
                  color: "rgba(255, 255, 255, 0.5)",
                  fontSize: 12,
                  marginBottom: 6,
                  animation: "fadeIn 0.2s ease",
                }}>
                  {SPOTIFY_PLAYLIST_BASE}
                </div>
              )}
              <input
                type="text"
                placeholder="Paste playlist URL..."
                value={playlistId}
                onChange={e => handlePlaylistInput((e.target as HTMLInputElement).value)}
                style={{
                  ...inputStyle,
                  width: "100%",
                  marginBottom: 12,
                  borderColor: playlistId ? "#1DB954" : "rgba(255, 255, 255, 0.2)",
                  transition: "border-color 0.2s ease",
                }}
                disabled={loading === "playlist"}
              />
              <button
                onClick={handleScanPlaylist}
                disabled={loading === "playlist" || !playlistId}
                style={{
                  ...buttonStyle,
                  width: "100%",
                  opacity: (loading === "playlist" || !playlistId) ? 0.5 : 1,
                  cursor: (loading === "playlist" || !playlistId) ? "not-allowed" : "pointer",
                }}
              >
                {loading === "playlist" ? "Scanning..." : "Scan Playlist"}
              </button>

              <div style={{ textAlign: "center", margin: "16px 0" }}>
                <span style={{ color: "rgba(255,255,255,0.4)", fontSize: 13 }}>or</span>
              </div>

              <button
                onClick={handleScanLiked}
                disabled={loading === "liked"}
                style={{
                  ...buttonStyle,
                  width: "100%",
                  background: "transparent",
                  border: "1px solid #1DB954",
                  color: "#1DB954",
                  opacity: loading === "liked" ? 0.5 : 1,
                }}
              >
                {loading === "liked" ? "Scanning..." : "Scan Liked Songs"}
              </button>
            </div>

            {/* Your Playlists */}
            <div style={cardStyle}>
              <h3 style={{ color: "#fff", marginBottom: 12, fontSize: "1.1rem" }}>Your Playlists</h3>
              {!playlistsLoading && userPlaylists.length > 0 && (
                <input
                  type="text"
                  placeholder="Search playlists..."
                  value={playlistSearch}
                  onChange={e => setPlaylistSearch(e.target.value)}
                  style={{
                    ...inputStyle,
                    width: "100%",
                    marginBottom: 12,
                    fontSize: 13,
                  }}
                />
              )}
              {playlistsLoading ? (
                <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "20px 0" }}>
                  Loading playlists...
                </div>
              ) : userPlaylists.length === 0 ? (
                <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "20px 0" }}>
                  No playlists found
                </div>
              ) : filteredPlaylists.length === 0 ? (
                <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "20px 0" }}>
                  No playlists match "{playlistSearch}"
                </div>
              ) : (
                <div style={{ marginLeft: "-16px" }}>
                  <AnimatedList
                    items={filteredPlaylists.map((p: Playlist) => p.ID)}
                    showGradients={false}
                    displayScrollbar={true}
                    enableArrowNavigation={true}
                    maxHeight={400}
                    children={(itemId: string) => {
                    const playlist = filteredPlaylists.find((p: Playlist) => p.ID === itemId);
                    if (!playlist) return null;
                    const isSelected = playlistId === itemId;
                    return (
                      <button
                        onClick={() => setPlaylistId(playlist.ID)}
                        style={{
                          display: "flex",
                          alignItems: "center",
                          gap: 12,
                          padding: 10,
                          background: isSelected ? "rgba(29, 185, 84, 0.2)" : "rgba(255, 255, 255, 0.05)",
                          border: isSelected ? "2px solid #1DB954" : "1px solid rgba(255, 255, 255, 0.1)",
                          borderRadius: 10,
                          cursor: "pointer",
                          transition: "all 0.2s ease",
                          textAlign: "left",
                          width: "100%",
                        }}
                      >
                        <img
                          src={playlist.ImageURL}
                          alt={playlist.Name}
                          style={{
                            width: 48,
                            height: 48,
                            borderRadius: 6,
                            objectFit: "cover",
                            flexShrink: 0,
                          }}
                          onError={e => {
                            (e.target as HTMLImageElement).src = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='48' height='48' viewBox='0 0 24 24' fill='%23666'%3E%3Cpath d='M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z'/%3E%3C/svg%3E";
                          }}
                        />
                        <div style={{ overflow: "hidden", flex: 1 }}>
                          <div style={{
                            color: "#fff",
                            fontSize: 13,
                            fontWeight: 500,
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap",
                          }}>
                            {playlist.Name}
                          </div>
                          <div style={{
                            color: "rgba(255,255,255,0.5)",
                            fontSize: 11,
                            marginTop: 2,
                          }}>
                            {playlist.TrackCount} tracks
                          </div>
                        </div>
                      </button>
                    );
                  }}
                  />
                </div>
              )}
            </div>
          </div>

          {/* Right Column - Results */}
          <div>
            {/* Error Message */}
            {error && (
              <div style={{
                ...cardStyle,
                background: "rgba(231, 76, 60, 0.1)",
                border: "1px solid rgba(231, 76, 60, 0.3)",
                color: "#e74c3c",
                marginBottom: 20,
                textAlign: "center",
              }}>
                {error}
              </div>
            )}

            {/* Results */}
            {ruContent ? (
              <div style={cardStyle}>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20, flexWrap: "wrap", gap: 12 }}>
                  <h3 style={{ color: "#fff", margin: 0, fontSize: "1.1rem" }}>
                    Tracks with Russian Artists ({(ruContent.Tracks ?? []).length})
                  </h3>
                  <div style={{ display: "flex", gap: 12, alignItems: "center" }}>
                    {/* Only show Select All if AbleToDelete is true and there are tracks */}
                    {ruContent.AbleToDelete && (ruContent.Tracks ?? []).length > 0 && (
                      <label style={{ color: "rgba(255,255,255,0.7)", fontSize: 14, cursor: "pointer", display: "flex", alignItems: "center", gap: 8 }}>
                        <input
                          type="checkbox"
                          checked={(ruContent.Tracks ?? []).length > 0 && (ruContent.Tracks ?? []).every((t: Track) => selected.has(t.ID))}
                          onChange={e => handleTrackSelectAll((e.target as HTMLInputElement).checked)}
                          style={{ accentColor: "#1DB954" }}
                        />
                        Select All
                      </label>
                    )}
                    {/* Only show delete button if AbleToDelete is true and there are tracks */}
                    {ruContent.AbleToDelete && (ruContent.Tracks ?? []).length > 0 && (
                      <button
                        onClick={() => handleDeleteTracks(lastPlaylistId ? 'playlist' : 'liked')}
                        disabled={selected.size === 0 || loading === 'delete'}
                        style={{
                          ...deleteButtonStyle,
                          opacity: (selected.size === 0 || loading === 'delete') ? 0.5 : 1,
                          cursor: (selected.size === 0 || loading === 'delete') ? "not-allowed" : "pointer",
                        }}
                      >
                        {loading === 'delete' ? 'Deleting...' : `Delete (${selected.size})`}
                      </button>
                    )}
                  </div>
                </div>

                <AnimatedList
                  items={(ruContent.Tracks ?? []).map((t: Track) => t.ID)}
                  showGradients={false}
                  displayScrollbar={true}
                  enableArrowNavigation={true}
                  children={(trackId: string) => {
                    const track = (ruContent.Tracks ?? []).find((t: Track) => t.ID === trackId);
                    if (!track) return null;
                    const isSelected = selected.has(track.ID);
                    return (
                      <div style={{
                        padding: 16,
                        background: isSelected ? "rgba(231, 76, 60, 0.15)" : "rgba(255, 255, 255, 0.03)",
                        border: isSelected ? "1px solid rgba(231, 76, 60, 0.3)" : "1px solid rgba(255, 255, 255, 0.05)",
                        borderRadius: 10,
                        transition: "all 0.2s ease",
                      }}>
                        <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer', gap: 12 }}>
                          {ruContent.AbleToDelete && (
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={(e) => {
                                e.stopPropagation();
                                handleTrackToggle(track.ID);
                              }}
                              style={{ accentColor: "#1DB954", width: 18, height: 18, flexShrink: 0 }}
                            />
                          )}
                          {track.ImageURL && (
                            <img
                              src={track.ImageURL}
                              alt={track.Name}
                              style={{
                                width: 48,
                                height: 48,
                                borderRadius: 4,
                                objectFit: "cover",
                                flexShrink: 0,
                              }}
                              onError={(e) => {
                                (e.target as HTMLImageElement).style.display = 'none';
                              }}
                            />
                          )}
                          <div style={{ flex: 1, minWidth: 0 }}>
                            <div style={{ fontWeight: 600, color: "#fff", fontSize: 15, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{track.Name}</div>
                            <div style={{ marginTop: 4, fontSize: 13 }}>
                              {(track.Artists ?? []).map((artist: Artist, idx: number) => (
                                <span key={artist.ID}>
                                  <span
                                    style={{
                                      color: ruArtistIds.has(artist.ID) ? "#e74c3c" : "rgba(255,255,255,0.6)",
                                      fontWeight: ruArtistIds.has(artist.ID) ? 600 : 400,
                                    }}
                                  >
                                    {artist.Name}
                                  </span>
                                  {idx < (track.Artists ?? []).length - 1 && <span style={{ color: "rgba(255,255,255,0.3)", margin: "0 6px" }}>•</span>}
                                </span>
                              ))}
                            </div>
                          </div>
                        </label>
                      </div>
                    );
                  }}
                />

                {(ruContent.Tracks ?? []).length === 0 && (
                  <div style={{ textAlign: "center", color: "rgba(255,255,255,0.5)", padding: "40px 0" }}>
                    No Russian tracks found. Your library is clean!
                  </div>
                )}
              </div>
            ) : (
              <div style={{
                ...cardStyle,
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                justifyContent: "center",
                minHeight: 400,
                color: "rgba(255,255,255,0.5)",
              }}>
                <svg width="64" height="64" viewBox="0 0 24 24" fill="currentColor" style={{ opacity: 0.3, marginBottom: 16 }}>
                  <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/>
                </svg>
                <p style={{ fontSize: 16, marginBottom: 8 }}>No scan results yet</p>
                <p style={{ fontSize: 13, opacity: 0.7 }}>Select a playlist or scan your liked songs to get started</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
