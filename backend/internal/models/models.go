package models

type User struct {
	ID        int    `json:"ID"`
	SpotifyID string `json:"SpotifyID"`
}

type Track struct {
	ID      	string   `json:"ID"`
	Name    	string   `json:"Name"`
	ImageURL 	string   `json:"ImageURL"`
	Artists 	[]Artist `json:"Artists"`
}

type Artist struct {
	ID   		string `json:"ID"`
	URL			string `json:"URL"`
	Name 		string `json:"Name"`
}

type Playlist struct {
	ID          string  `json:"ID"`
	Name        string  `json:"Name"`
	Description string  `json:"Description"`
	Owned 	  	bool    `json:"Owned"`
	ImageURL    string  `json:"ImageURL"`
	TrackCount  int     `json:"TrackCount"`
	Tracks      []Track `json:"Tracks"`
}

type RuContent struct {
	Tracks  		[]Track  	`json:"Tracks"`
	Artists 		[]Artist 	`json:"Artists"`
}
