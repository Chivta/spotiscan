package spotify

type SpotifyArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SpotifyImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type SpotifyPlaylistItemsResponse struct {
	Items  []SpotifyItem `json:"items"`
	Total  int           `json:"total"`
	Next   *string       `json:"next"`
	Offset int           `json:"offset"`
}

type SpotifyPlaylistResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Images      []SpotifyImage `json:"images"`
	Tracks      SpotifyPlaylistItemsResponse `json:"tracks"`
}

type SpotifyItem struct {
	IsLocal bool         `json:"is_local"`
	Track   SpotifyTrack `json:"track"`
}

type SpotifyTrack *struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Album struct {
		Images []SpotifyImage `json:"images"`
	} `json:"album"`
	Artists []SpotifyArtist `json:"artists"`
}

type SpotifyAlbum struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	ReleaseDate string          `json:"release_date"`
	TotalTracks int             `json:"total_tracks"`
	Images      []SpotifyImage  `json:"images"`
	Artists     []SpotifyArtist `json:"artists"`
}

type SpotifyTrackResponse struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Album   SpotifyAlbum `json:"album"`
	Artists []SpotifyArtist `json:"artists"`
}

type SpotifyAlbumTrack struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	DurationMs  int             `json:"duration_ms"`
	TrackNumber int             `json:"track_number"`
	Artists     []SpotifyArtist `json:"artists"`
}

type SpotifyArtistResponse struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Genres     []string       `json:"genres"`
	Popularity int            `json:"popularity"`
	Images     []SpotifyImage `json:"images"`
	Followers  struct {
		Total int `json:"total"`
	} `json:"followers"`
}

type SpotifyAlbumResponse struct {
	SpotifyAlbum
	Tracks struct {
		Items []SpotifyAlbumTrack `json:"items"`
		Total int                 `json:"total"`
		Next  *string             `json:"next"`
	} `json:"tracks"`
}
