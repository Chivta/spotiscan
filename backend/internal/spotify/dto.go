package spotify

type SpotifyPlaylistItemsResponse struct {
	Items  []SpotifyItem `json:"items"`
	Total  int           `json:"total"`
	Next   *string       `json:"next"`
	Offset int           `json:"offset"`
}

type SpotifyPlaylistResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Images      []struct {
		URL string `json:"url"`
	} `json:"images"`
	Tracks SpotifyPlaylistItemsResponse `json:"tracks"`
}

type SpotifyItem struct {
	IsLocal bool         `json:"is_local"`
	Track   SpotifyTrack `json:"track"`
}

type SpotifyTrack *struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Album struct {
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	} `json:"album"`
	Artists []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"artists"`
}
