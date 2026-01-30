package db

import (
	"database/sql"

	"golang.org/x/oauth2"
)

func (db *DB) StoreSpotifyTokens(token *oauth2.Token) error {
	accessToken := token.AccessToken
	expiresAt := token.Expiry.UTC()
	_, err := db.conn.Exec(
		`INSERT INTO spotify_tokens (access_token, expires_at) VALUES ($1, $2)`,
		accessToken, expiresAt,
	)
	return err
}

func (db *DB) GetSpotifyTokens() (*oauth2.Token, error) {
	var accessToken string
	var expiresAt sql.NullTime
	err := db.conn.QueryRow(`SELECT access_token, expires_at FROM spotify_tokens`).Scan(&accessToken, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	token := oauth2.Token{
		AccessToken: accessToken,
	}
	if expiresAt.Valid {
		token.Expiry = expiresAt.Time
	}
	return &token, nil
}
