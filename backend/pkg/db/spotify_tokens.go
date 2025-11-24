package db

import (
	"database/sql"
	"golang.org/x/oauth2"
)

func (db *DB) StoreSpotifyTokens(userId int, token *oauth2.Token) error {
	accessToken := token.AccessToken
	refreshToken := token.RefreshToken
	expiresAt := token.Expiry

	_, err := db.conn.Exec(
		`INSERT INTO spotify_tokens (user_id, access_token, refresh_token, expires_at) VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, expires_at = EXCLUDED.expires_at`,
		userId, accessToken, refreshToken, expiresAt,
	)
	return err
}

func (db *DB) GetSpotifyTokensByUserId(userId int) (*oauth2.Token, error) {
	var accessToken, refreshToken string
	var expiresAt sql.NullTime
	err := db.conn.QueryRow(`
		      SELECT access_token, refresh_token, expires_at FROM spotify_tokens 
		      WHERE user_id=$1
	      `, userId).Scan(&accessToken, &refreshToken, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	token := oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if expiresAt.Valid {
		token.Expiry = expiresAt.Time
	}
	return &token, nil
}
