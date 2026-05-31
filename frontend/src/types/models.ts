export interface User {
  userID: string;
  userRole: "user" | "admin" | "anon";
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

export interface ArtistRef {
  SpotifyID: string;
  Name: string;
}

export interface Track {
  SpotifyID: string;
  Name: string;
  ImageURL: string;
  ArtistRefs: ArtistRef[];
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

// RFC3339 timestamp string as serialized by Go's time.Time JSON marshaler.
export type ISOTimestamp = string;

export interface ArtistDeleteSuggestion {
  ID: number;
  CreatorID: number;
  ArtistName: string;
  Description: string;
  State: "pending" | "approved" | "declined";
  DeclineReason?: string;
  CreatedAt: ISOTimestamp;
  UpdatedAt: ISOTimestamp;
}

export interface ArtistInsertSuggestion {
  ID: number;
  CreatorID: number;
  ArtistName: string;
  Description: string;
  State: "pending" | "approved" | "declined";
  DeclineReason?: string;
  CreatedAt: ISOTimestamp;
  UpdatedAt: ISOTimestamp;
}
