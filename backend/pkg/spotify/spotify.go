package spotify

import (
	"context"
	"log"
	"spotiscan/models"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func NewSpotifyClient(spotifyId, spotifySecret, redirectURI string) *SpotifyClient {
	return &SpotifyClient{
		spotifyId:     spotifyId,
		spotifySecret: spotifySecret,
		redirectURI:   redirectURI,
		auth: spotifyauth.New(
			spotifyauth.WithRedirectURL(redirectURI),
			spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserLibraryRead, spotifyauth.ScopePlaylistModifyPublic, spotifyauth.ScopePlaylistModifyPrivate),
			spotifyauth.WithClientID(spotifyId),
			spotifyauth.WithClientSecret(spotifySecret),
		),
	}
}

type SpotifyClient struct {
	spotifyId     string
	spotifySecret string
	redirectURI   string
	auth          *spotifyauth.Authenticator
}

func (c *SpotifyClient) GetAuthURL(state string) string {
	return c.auth.AuthURL(state)
}


func (c *SpotifyClient) GetUserSavedTracks(token *oauth2.Token) ([]models.Track, error) {
	ctx := context.Background()
	client := spotify.New(c.auth.Client(ctx, token))

	var limit = 50
	var offset = 0
	var allTracks []models.Track
	for {
		page, err := client.CurrentUsersTracks(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			log.Printf("couldn't fetch user library: %v", err)
			return nil, err
		}

		for _, item := range page.Tracks {
			track := item.FullTrack.SimpleTrack
			var trackArtists []models.Artist
			for _, artist := range track.Artists {
				artist := models.Artist{
					ID:   string(artist.ID),
					Name: artist.Name,
				}

				trackArtists = append(trackArtists, artist)
			}
			allTracks = append(allTracks, models.Track{
				ID:      string(track.ID),
				Name:    track.Name,
				Artists: trackArtists,
			})
		}

		if len(page.Tracks) < limit {
			break
		}

		offset += limit
	}
	log.Println("Total liked songs fetched:", len(allTracks))

	return allTracks, nil
}


func (c *SpotifyClient) FetchSpotifyUserId(token *oauth2.Token) (string, error) {
	ctx := context.Background()
	client := spotify.New(c.auth.Client(ctx, token))

	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Printf("couldn't fetch user: %v", err)
		return "", err
	}

	return user.ID, nil
}

