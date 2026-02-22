package spotify_client

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/chivta/spotiscan/internal/models"
)

func NewSpotifyClient(spotifyId, spotifySecret string) *SpotifyClient {
	return &SpotifyClient{
		spotifyId:     spotifyId,
		spotifySecret: spotifySecret,
	}
}

type SpotifyClient struct {
	spotifyId     string
	spotifySecret string
}

func (c *SpotifyClient) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error) {
	token, ok := ctx.Value("spotify_token").(*oauth2.Token)
	if !ok || token == nil {
		return nil, fmt.Errorf("missing or invalid spotify token in context")
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	spotifyPlaylist, err := client.GetPlaylist(ctx, spotify.ID(playlistId))
	if err != nil {
		return nil, err
	}
	var playlist models.Playlist
	playlist.ID = string(spotifyPlaylist.ID)
	playlist.Name = spotifyPlaylist.Name
	// Fill in tracks
	tracks, err := c.getAllPlaylistTracks(ctx, client, spotifyPlaylist.ID)
	if err != nil {
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

func (c *SpotifyClient) getAllPlaylistTracks(ctx context.Context, client *spotify.Client, playlistId spotify.ID) ([]models.Track, error) {
	var limit = 50
	var offset = 0
	var allTracks []models.Track
	for {
		page, err := client.GetPlaylistItems(ctx, playlistId, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
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

func (c *SpotifyClient) GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	config := &clientcredentials.Config{
		ClientID:     c.spotifyId,
		ClientSecret: c.spotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}
	// Ensure the token's expiry is in UTC
	token.Expiry = token.Expiry.UTC()

	return token, nil
}
