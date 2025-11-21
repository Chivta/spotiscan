package spotify

import (
	"context"
	"log"
	"net/http"

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
			spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate),
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

func (c *SpotifyClient) GetToken(r *http.Request, state string) (*oauth2.Token, error) {
	token, err := c.auth.Token(r.Context(),state, r)
	if err != nil {
		log.Printf("couldn't get token: %v", err)
		return nil, err
	}
	return token, nil
}

func (c *SpotifyClient) GetRuArtistsFromPlaylist(playlistId string, token *oauth2.Token) ([]string, error) {
	ctx := context.Background()
	client := spotify.New(c.auth.Client(ctx, token))

	playlistTracks, err := client.GetPlaylistItems(ctx, spotify.ID(playlistId))
	if err != nil {
		log.Printf("couldn't get playlist items: %v", err)
		return nil, err
	}

	var artists []string
	for _, item := range playlistTracks.Items {
		for _, artist := range item.Track.Track.Artists {
			if isRussianArtist(artist.Name) {
				artists = append(artists, artist.Name)
			}
		}
	}

	return artists, nil
}

func (c *SpotifyClient) RefreshToken(token *oauth2.Token) (*oauth2.Token, error) {
	ctx := context.Background()
	newToken, err := c.auth.RefreshToken(ctx,token)
	if err != nil {
		log.Printf("couldn't refresh token: %v", err)
		return nil, err
	}
	return newToken, nil
}

func (c *SpotifyClient) FetchSpotifyUserId(token *oauth2.Token) (string, error) {
	ctx := context.Background()
	client := spotify.New(c.auth.Client(ctx, token))

	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Printf("couldn't fetch user: %v", err)
		return "", err
	}

	log.Printf("Fetched user: %s", user.ID)

	return user.ID, nil
}

func isRussianArtist(artistName string) bool {
	// TODO: Implement actual logic to determine if an artist is Russian.
	return len(artistName)%2 == 0
}