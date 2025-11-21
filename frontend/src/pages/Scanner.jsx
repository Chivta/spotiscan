import React, { useState } from "react";

export default function Scanner() {
  const [playlistURL, setPlaylistURL] = useState("");
  const [artists, setArtists] = useState([]);
  const IDExtract = /(?:playlist\/|spotify:playlist:)([A-Za-z0-9]+)/;

  const fetchArtists = async (event) => {
    event.preventDefault();
    setArtists([]);
    if (!IDExtract.test(playlistURL)) {
      alert("Invalid Spotify playlist URL format");
      return;
    }
    const match = IDExtract.exec(playlistURL);
    const playlistID = match[1];
    const response = await fetch(`/api/playlist/ruartists?id=${encodeURIComponent(playlistID)}`);
    if (!response.ok) {
      throw new Error(`HTTP error status: ${response.status}`);
    }
    setArtists(await response.json());
  };

  return (
    <main>
      <h1>Paste spotify playlist url</h1>
      <form>
        <input
          type="text"
          value={playlistURL}
          onChange={e => setPlaylistURL(e.target.value)}
        />
        <button onClick={fetchArtists} placeholder="Playlist url">Get ru artists</button>
      </form>
      <ul>
        {artists.map(artist => (
          <li key={artist}>{artist}</li>
        ))}
      </ul>
    </main>
  );
}
