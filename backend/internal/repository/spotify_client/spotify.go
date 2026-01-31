package spotify_client

import (
	"context"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/chivta/spotiscan/internal/models"
)

type SpotifyClient interface {
	GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error)
	GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error)
}

func NewSpotifyClient(spotifyId, spotifySecret string) SpotifyClient {
	return &spotifyClient{
		spotifyId:     spotifyId,
		spotifySecret: spotifySecret,
	}
}

type spotifyClient struct {
	spotifyId     string
	spotifySecret string
}


func (c *spotifyClient) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error) {
	httpClient := spotifyauth.New().Client(ctx, ctx.Value("spotify_token").(*oauth2.Token))
	client := spotify.New(httpClient)

	spotifyPlaylist, err := client.GetPlaylist(ctx, spotify.ID(playlistId))
	if err != nil {
		if spotifyErr, ok := err.(spotify.Error); ok && spotifyErr.Status == 404 {
			return nil, ErrNotFound
		}
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

func (c *spotifyClient) getAllPlaylistTracks(ctx context.Context, client *spotify.Client, playlistId spotify.ID) ([]models.Track, error) {
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


func (c *spotifyClient) GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error) {	
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
