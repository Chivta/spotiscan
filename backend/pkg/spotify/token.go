package spotify

import (
	"context"
	"log"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2"
)

func (c *SpotifyClient) GetToken() (*oauth2.Token, error) {
	ctx := context.Background()
	
	config := &clientcredentials.Config{
		ClientID:     c.spotifyId,
		ClientSecret: c.spotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	
	token, err := config.Token(ctx)
	if err != nil {
		log.Printf("couldn't get token: %v", err)
		return nil, err
	}
	// Ensure the token's expiry is in UTC
	token.Expiry = token.Expiry.UTC()
	
	return token, nil
}
