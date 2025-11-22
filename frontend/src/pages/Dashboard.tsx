import * as React from "react";
import { useState } from "react";
// Ensure JSX namespace is available for TS
/// <reference types="react/next" />


// Types
export interface Artist {
  ID: string;
  Name: string;
}

export interface Track {
  ID: string;
  Name: string;
  Artists: Artist[];
}

export interface RuContent {
  RuTracks: Track[];
  RuArtists: Artist[];
}

interface ArtistListProps {
  artists: Artist[];
  selected: Set<string>;
  onToggle: (id: string) => void;
  onSelectAll: (checked: boolean) => void;
}

const ArtistList = ({ artists, selected, onToggle, onSelectAll }: ArtistListProps) => {
  if (!artists.length) return null;
  const allSelected = artists.length > 0 && artists.every((a: Artist) => selected.has(a.ID));
  return (
    <section style={{ marginTop: 32 }}>
      <label style={{ fontWeight: "bold" }}>
        <input
          type="checkbox"
          checked={allSelected}
          onChange={e => onSelectAll((e.target as HTMLInputElement).checked)}
        />
        Select All
      </label>
      <ul style={{ listStyle: "none", padding: 0, marginTop: 12 }}>
        {artists.map((artist: Artist) => (
          <li key={artist.ID} style={{ display: "flex", alignItems: "center", marginBottom: 8 }}>
            <input
              type="checkbox"
              checked={selected.has(artist.ID)}
              onChange={() => onToggle(artist.ID)}
              style={{ marginRight: 8 }}
            />
            {artist.Name}
          </li>
        ))}
      </ul>
    </section>
  );
};

interface PlaylistInputProps {
  value: string;
  onChange: (v: string) => void;
  onScan: () => void;
  loading: boolean;
}

const PlaylistInput = ({ value, onChange, onScan, loading }: PlaylistInputProps) => (
  <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 24 }}>
    <input
      type="text"
      placeholder="Enter playlist URL"
      value={value}
      onChange={e => onChange((e.target as HTMLInputElement).value)}
      style={{ flex: 1, padding: 8, fontSize: 16 }}
      disabled={loading}
    />
    <button onClick={onScan} disabled={loading || !value} style={{ padding: "8px 16px" }}>
      {loading ? "Scanning..." : "Scan Playlist"}
    </button>
  </div>
);

interface ScanButtonsProps {
  onScanLiked: () => void;
  loading: string | null;
}

const ScanButtons = ({ onScanLiked, loading }: ScanButtonsProps) => (
  <div style={{ display: "flex", flexDirection: "column", gap: 16, alignItems: "center", marginBottom: 24 }}>
    <button onClick={onScanLiked} disabled={loading === "liked"} style={{ width: 220, padding: "10px 0" }}>
      {loading === "liked" ? "Scanning..." : "Scan Liked Songs"}
    </button>
  </div>
);

const Dashboard = () => {
  const [playlistUrl, setPlaylistUrl] = useState("");
  const [ruContent, setRuContent] = useState<RuContent | null>(null);
  const [selected, setSelected] = useState(new Set<string>());
  const [loading, setLoading] = useState<null | string>(null); // null | 'playlist' | 'liked'
  const [error, setError] = useState<string | null>(null);
  const [lastPlaylistId, setLastPlaylistId] = useState<string | null>(null);


  // Real fetch for playlist scan
  const fetchPlaylistRuContent = async (playlistUrl: string): Promise<RuContent> => {
    setError(null);
    setLoading("playlist");
    // Extract playlist ID from URL
    const match = /(?:playlist\/|spotify:playlist:)([A-Za-z0-9]+)/.exec(playlistUrl);
    if (!match) throw new Error("Invalid Spotify playlist URL format");
    const playlistID = match[1];
    const response = await fetch(`/api/playlist/${encodeURIComponent(playlistID)}/rucontent`);
    if (!response.ok) {
      throw new Error(`HTTP error status: ${response.status}`);
    }
    return await response.json();
  };

  const handleScanPlaylist = async () => {
    // Extract playlist ID from URL
    const match = /(?:playlist\/|spotify:playlist:)([A-Za-z0-9]+)/.exec(playlistUrl);
    if (!match) {
      setError("Invalid Spotify playlist URL format");
      return;
    }
    const playlistID = match[1];
    if (playlistID === lastPlaylistId) {
      setError("This playlist has already been scanned.");
      return;
    }
    try {
      const data = await fetchPlaylistRuContent(playlistUrl);
      setRuContent(data);
      setSelected(new Set());
      setLastPlaylistId(playlistID);
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

  const handleToggle = (id: string) => {
    setSelected((prev: Set<string>) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };
  const handleSelectAll = (checked: boolean) => {
    if (ruContent && checked) setSelected(new Set((ruContent.RuArtists ?? []).map((a: Artist) => a.ID)));
    else setSelected(new Set());
  };

  const handleDelete = async () => {
    setLoading("delete");
    setError(null);
    try {
      // Simulate API call
      await new Promise(res => setTimeout(res, 800));
      if (ruContent) {
        const filtered = (ruContent.RuArtists ?? []).filter((a: Artist) => !selected.has(a.ID));
        setRuContent({ ...ruContent, RuArtists: filtered });
      }
      setSelected(new Set());
    } catch {
      setError("Failed to delete selected artists.");
    } finally {
      setLoading(null);
    }
  };

  // Helper: set of RuArtist IDs for quick lookup
  const ruArtistIds = new Set((ruContent?.RuArtists ?? []).map(a => a.ID));

  return (
    <section style={{ maxWidth: 700, margin: "40px auto", padding: 24, background: "#fff", borderRadius: 12, boxShadow: "0 2px 12px #0001" }}>
      <h2 style={{ marginBottom: 24 }}>Music Dashboard</h2>
      <PlaylistInput value={playlistUrl} onChange={setPlaylistUrl} onScan={handleScanPlaylist} loading={loading === "playlist"} />
      <ScanButtons
        onScanLiked={handleScanLiked}
        loading={loading}
      />
      {error && <div style={{ color: "#b00", marginBottom: 16 }}>{error}</div>}

      {/* Show tracks and artists if ruContent is loaded */}
      {ruContent && (
        <div style={{ marginTop: 32 }}>
          <h3>Tracks with Russian Artists</h3>
          <ul style={{ listStyle: "none", padding: 0 }}>
            {(ruContent.RuTracks ?? []).map(track => (
              <li key={track.ID} style={{ marginBottom: 18, padding: 12, border: "1px solid #eee", borderRadius: 8 }}>
                <div style={{ fontWeight: 600 }}>{track.Name}</div>
                <div style={{ marginTop: 4, fontSize: 15 }}>
                  {(track.Artists ?? []).map(artist => (
                    <span
                      key={artist.ID}
                      style={{
                        color: ruArtistIds.has(artist.ID) ? "#e74c3c" : undefined,
                        fontWeight: ruArtistIds.has(artist.ID) ? 700 : 400,
                        marginRight: 12
                      }}
                    >
                      {artist.Name}
                    </span>
                  ))}
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  );
};

export default Dashboard;
