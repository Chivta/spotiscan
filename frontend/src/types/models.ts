export interface User {
  ID: number;
  SpotifyID: string;
}

export interface Artist {
  ID: string;
  URL: string;
  Name: string;
}

export interface Track {
  ID: string;
  Name: string;
  ImageURL: string;
  Artists: Artist[];
}

export interface Playlist {
  ID: string;
  Name: string;
  Description?: string;
  Owned: boolean;
  ImageURL: string;
  TrackCount: number;
  Tracks?: Track[];
}

export interface RuContent {
	AbleToDelete: boolean;
  Tracks: Track[];
  Artists: Artist[];
}
