package db

import (
	"database/sql"
	"log"
	"golang.org/x/oauth2"
	"github.com/lib/pq"

	"github.com/chivta/spotiscan/internal/models"
)

type DB interface {
	Close() error
	FilterRussian(artists map[string]models.Artist) (map[string]models.Artist, error)
	StoreSpotifyTokens(token *oauth2.Token) error
	GetSpotifyTokens() (*oauth2.Token, error)
}


func NewDBConnection(DatabaseURL string) (DB, error) {
	db, err := sql.Open("postgres", DatabaseURL)

	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &dbConn{db}, nil
}

type dbConn struct {
	conn *sql.DB
}

func (db *dbConn) Close() error {
	err := db.conn.Close()
	if err != nil {
		log.Println("Error closing database connection:", err)
	}
	return err
}


func (db *dbConn) StoreSpotifyTokens(token *oauth2.Token) error {
	accessToken := token.AccessToken
	expiresAt := token.Expiry.UTC()
	_, err := db.conn.Exec(
		`INSERT INTO spotify_tokens (access_token, expires_at) VALUES ($1, $2)`,
		accessToken, expiresAt,
	)
	return err
}

func (db *dbConn) GetSpotifyTokens() (*oauth2.Token, error) {
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


func (db *dbConn) FilterRussian(artists map[string]models.Artist) (map[string]models.Artist, error) {
	var names []string
	for name := range artists {
		names = append(names, name)
	}

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
	ruArtists := make(map[string]models.Artist, len(ruNames))

	for _, name := range ruNames {
		ruArtists[name] = artists[name]
	}

	return ruArtists, nil
}
