package db_client

import (
	"context"
	"database/sql"
	"log"

	"github.com/lib/pq"
	"golang.org/x/oauth2"
)

type DBClient interface {
	Close() error
	GetRussianArtistNames(ctx context.Context, names []string) ([]string, error)
	GetAllRussianArtistNames(ctx context.Context) ([]string, error)
	SetSpotifyToken(ctx context.Context, token *oauth2.Token) error
	GetSpotifyToken(ctx context.Context) (*oauth2.Token, error)
}

func NewDBClient(DatabaseURL string) (DBClient, error) {
	db, err := sql.Open("postgres", DatabaseURL)

	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &dbClient{db}, nil
}

type dbClient struct {
	conn *sql.DB
}

func (db *dbClient) Close() error {
	err := db.conn.Close()
	if err != nil {
		log.Println("Error closing database connection:", err)
	}
	return err
}

func (db *dbClient) SetSpotifyToken(ctx context.Context, token *oauth2.Token) error {
	accessToken := token.AccessToken
	expiresAt := token.Expiry.UTC()
	_, err := db.conn.Exec(
		`INSERT INTO spotify_tokens (access_token, expires_at) VALUES ($1, $2)`,
		accessToken, expiresAt,
	)
	return err
}

func (db *dbClient) GetSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	var accessToken string
	var expiresAt sql.NullTime
	err := db.conn.QueryRow(`SELECT access_token, expires_at FROM spotify_tokens`).Scan(&accessToken, &expiresAt)
	if err != nil {
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

func (db *dbClient) GetAllRussianArtistNames(ctx context.Context) ([]string, error) {
	rows, err := db.conn.Query(`SELECT name FROM ru_artists`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ruNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		ruNames = append(ruNames, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ruNames, nil
}

func (db *dbClient) GetRussianArtistNames(ctx context.Context, names []string) ([]string, error) {
	rows, err := db.conn.Query(`
        SELECT name FROM ru_artists WHERE name = ANY($1)
    `, pq.Array(names))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ruNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		ruNames = append(ruNames, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return ruNames, nil
}
