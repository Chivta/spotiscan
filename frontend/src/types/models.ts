export interface User {
  userID: string;
  userRole: "user" | "admin" | "anon";
  Email?: string;
}

export interface Artist {
  ID: number;
  SpotifyID: string;
  Name: string;
  DescriptionUA: string;
  DescriptionEN: string;
  Source: string;
  SourceURL: string;
  Country: string;
  Confirmed: boolean;
}

export interface Track {
  SpotifyID: string;
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
