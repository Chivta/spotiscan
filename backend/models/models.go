package models

type User struct {
	ID        int    `json:"ID"`
	SpotifyID string `json:"SpotifyID"`
}

type Track struct {
	ID      string   `json:"ID"`
	Name    string   `json:"Name"`
	Artists []Artist `json:"Artists"`
}

type Artist struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

type Playlist struct {
	ID     string  `json:"ID"`
	Name   string  `json:"Name"`
	Tracks []Track `json:"Tracks"`
}


type RuContent struct {
	RuTracks  []Track  `json:"RuTracks"`
	RuArtists []Artist `json:"RuArtists"`
}
