export interface User {
  userID: string;
  userRole: "user" | "admin" | "anon";
  Email?: string;
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
  ImageURL: string;
  TrackCount: number;
  Tracks?: Track[];
}

export interface RuContent {
  Tracks: Track[];
  Artists: Artist[];
}
