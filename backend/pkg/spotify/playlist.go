package spotify

import (
	"context"
	"log"
	"spotiscan/models"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	"github.com/zmb3/spotify/v2/auth"
)


func (c *SpotifyClient) GetPlaylistWithTracks(playlistId string, token *oauth2.Token) (*models.Playlist, error) {
	ctx := context.Background()
	

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)


	spotifyPlaylist, err := client.GetPlaylist(ctx, spotify.ID(playlistId))
	if err != nil {
		log.Printf("couldn't fetch playlist: %v", err)
		return nil, err
	}
	var playlist models.Playlist
	playlist.ID = string(spotifyPlaylist.ID)
	playlist.Name = spotifyPlaylist.Name
	// Fill in tracks
	tracks, err := c.getAllPlaylistTracks(client, spotifyPlaylist.ID)
	if err != nil {
		log.Printf("couldn't fetch playlist tracks: %v", err)
		return nil, err
	}
	// Remove duplicate tracks
	trackMap := make(map[string]models.Track)
	for _, track := range tracks {
		trackMap[track.ID] = track
	}
	uniqueTracks := make([]models.Track, 0, len(trackMap))
	for _, track := range trackMap {
		uniqueTracks = append(uniqueTracks, track)
	}
	playlist.Tracks = uniqueTracks
	playlist.Owned = spotifyPlaylist.Owner.ID != ""

	return &playlist, nil
}


func (c *SpotifyClient) getAllPlaylistTracks(client *spotify.Client, playlistId spotify.ID) ([]models.Track, error) {
	var limit = 50
	var offset = 0
	var allTracks []models.Track
	for {
		page, err := client.GetPlaylistItems(context.Background(), playlistId, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			log.Printf("couldn't fetch playlist items: %v", err)
			return nil, err
		}

		for _, item := range page.Items {
			track := item.Track.Track
			var trackArtists []models.Artist
			for _, artist := range track.Artists {
				artist := models.Artist{
					ID:   string(artist.ID),
					URL:  artist.ExternalURLs["spotify"],
					Name: artist.Name,
				}

				trackArtists = append(trackArtists, artist)
			}
			imageURL := ""
			if len(track.Album.Images) > 0 {
				imageURL = track.Album.Images[0].URL
			}
			allTracks = append(allTracks, models.Track{
				ID:       string(track.ID),
				Name:     track.Name,
				ImageURL: imageURL,
				Artists:  trackArtists,
			})
		}

		if len(page.Items) < limit {
			break
		}

		offset += limit
	}
	return allTracks, nil
}