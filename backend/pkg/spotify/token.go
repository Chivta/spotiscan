package spotify

import (
	"context"
	"log"
	"net/http"
	"golang.org/x/oauth2"
)

func (c *SpotifyClient) GetToken(r *http.Request, state string) (*oauth2.Token, error) {
	token, err := c.auth.Token(r.Context(), state, r)
	if err != nil {
		log.Printf("couldn't get token: %v", err)
		return nil, err
	}
	// Ensure the token's expiry is in UTC
	token.Expiry = token.Expiry.UTC()
	return token, nil
}


func (c *SpotifyClient) RefreshToken(token *oauth2.Token) (*oauth2.Token, error) {
	ctx := context.Background()
	newToken, err := c.auth.RefreshToken(ctx, token)
	if err != nil {
		log.Printf("couldn't refresh token: %v", err)
		return nil, err
	}
	return newToken, nil
}
