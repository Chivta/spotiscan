package db_client

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/chivta/spotiscan/internal/models"
	"github.com/lib/pq"
	"golang.org/x/oauth2"
)

func NewDBClient(DatabaseURL string) (*DBClient, error) {
	db, err := sql.Open("postgres", DatabaseURL)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DBClient{db}, nil
}

type DBClient struct {
	conn *sql.DB
}

func (db *DBClient) Close() error {
	err := db.conn.Close()
	if err != nil {
		log.Println("Error closing database connection:", err)
	}
	return err
}

func (db *DBClient) GetConnection() *sql.DB {
	return db.conn
}

func (db *DBClient) SetSpotifyToken(ctx context.Context, token *oauth2.Token) error {
	accessToken := token.AccessToken
	expiresAt := token.Expiry.UTC()
	_, err := db.conn.ExecContext(
		ctx,
		`INSERT INTO spotify_tokens (singleton, access_token, expires_at) VALUES (true, $1, $2)
		 ON CONFLICT (singleton) DO UPDATE SET access_token = $1, expires_at = $2`,
		accessToken, expiresAt,
	)
	return err
}

func (db *DBClient) GetSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	var accessToken string
	var expiresAt sql.NullTime
	err := db.conn.QueryRowContext(ctx, `SELECT access_token, expires_at FROM spotify_tokens`).Scan(&accessToken, &expiresAt)
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

func (db *DBClient) GetAllRussianArtistNames(ctx context.Context) ([]string, error) {
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

func (db *DBClient) GetRussianArtistNames(ctx context.Context, names []string) ([]string, error) {
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

func (db *DBClient) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := db.conn.QueryRowContext(
		ctx,
		`SELECT id, email, password_hash FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	user.Role = models.RoleUser
	return &user, nil
}

func (db *DBClient) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := db.conn.QueryRowContext(
		ctx,
		`SELECT id, email, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	user.Role = models.RoleUser
	return &user, nil
}

func (db *DBClient) StoreRefreshTokenHash(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := db.conn.ExecContext(
		ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID,
		tokenHash,
		expiresAt.UTC(),
	)
	return err
}

func (db *DBClient) CreateUser(ctx context.Context, user *models.User) (int, error) {
	var userID int
	err := db.conn.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		user.Email,
		user.PasswordHash,
	).Scan(&userID)

	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (db *DBClient) GetRefreshTokenByUserID(ctx context.Context, userID int) (string, time.Time, error) {
	var tokenHash string
	var expiresAt time.Time
	err := db.conn.QueryRowContext(
		ctx,
		`SELECT token_hash, expires_at FROM refresh_tokens WHERE user_id = $1 ORDER BY id DESC LIMIT 1`,
		userID,
	).Scan(&tokenHash, &expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenHash, expiresAt, nil
}

func (db *DBClient) DeleteRefreshTokenHash(ctx context.Context, userID int) error {
	_, err := db.conn.ExecContext(
		ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID,
	)
	return err
}

func (db *DBClient) InsertArtists(ctx context.Context, artists []models.Artist) error {
	if len(artists) == 0 {
		return nil
	}
	var (
		names          = make([]string, 0, len(artists))
		descriptionsUA = make([]string, 0, len(artists))
		descriptionsEN = make([]string, 0, len(artists))
		sources        = make([]string, 0, len(artists))
		sourceURLs     = make([]string, 0, len(artists))
		countries      = make([]string, 0, len(artists))
		confirmed      = make([]bool, 0, len(artists))
	)

	seen := make(map[string]struct{}, len(artists))
	for _, artist := range artists {
		name := strings.ToLower(artist.Name)
		if name == "" {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}

		source := artist.Source
		if source == "" {
			source = "manual"
		}

		country := artist.Country
		if country == "" {
			country = "RU"
		}

		names = append(names, name)
		descriptionsUA = append(descriptionsUA, artist.DescriptionUA)
		descriptionsEN = append(descriptionsEN, artist.DescriptionEN)
		sources = append(sources, source)
		sourceURLs = append(sourceURLs, artist.SourceURL)
		countries = append(countries, country)
		confirmed = append(confirmed, artist.Confirmed)
	}

	if len(names) == 0 {
		return nil
	}

	_, err := db.conn.ExecContext(
		ctx,
		`INSERT INTO ru_artists (name, description_ua, description_en, source, source_url, country, confirmed)
		 SELECT *
		 FROM unnest(
			$1::text[],
			$2::text[],
			$3::text[],
			$4::artist_source[],
			$5::text[],
			$6::text[],
			$7::boolean[]
		 )
		 ON CONFLICT (name) DO UPDATE SET
			description_ua = EXCLUDED.description_ua,
			description_en = EXCLUDED.description_en,
			source = EXCLUDED.source,
			source_url = EXCLUDED.source_url,
			country = EXCLUDED.country,
			confirmed = EXCLUDED.confirmed`,
		pq.Array(names),
		pq.Array(descriptionsUA),
		pq.Array(descriptionsEN),
		pq.Array(sources),
		pq.Array(sourceURLs),
		pq.Array(countries),
		pq.Array(confirmed),
	)
	return err
}

func (db *DBClient) GetRuTags(ctx context.Context) ([]string, error) {
	rows, err := db.conn.QueryContext(ctx, `SELECT name FROM lastfm_ru_tags`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

func (db *DBClient) GetRuRegionIds(ctx context.Context) ([]string, error) {
	rows, err := db.conn.QueryContext(ctx, `SELECT mbid FROM musicbrainz_ru_regions`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}