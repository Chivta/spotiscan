package db_client

import (
	"context"
	"database/sql"
	"log"
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
