package repository

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/redis/go-redis/v9"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func NewTokenRepo(logger *logger.Logger, db *sql.DB, redis *redis.Client, spotifyId, spotifySecret string) *TokenRepo {
	return &TokenRepo{
		logger:        logger,
		db:            db,
		redis:         redis,
		spotifyId:     spotifyId,
		spotifySecret: spotifySecret,
	}
}

type TokenRepo struct {
	logger  *logger.Logger
	db      *sql.DB
	redis   *redis.Client

	token         *oauth2.Token
	spotifyId     string
	spotifySecret string
}

func (r *TokenRepo) GetRefreshTokenByUserID(ctx context.Context, userID int) (string, time.Time, error) {
	var tokenHash string
	var expiresAt time.Time
	err := r.db.QueryRowContext(
		ctx,
		`SELECT token_hash, expires_at FROM refresh_tokens WHERE user_id = $1 ORDER BY id DESC LIMIT 1`,
		userID,
	).Scan(&tokenHash, &expiresAt)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		if err == sql.ErrNoRows {
			return "", time.Time{}, appErrors.ErrUnauthorized
		}
		return "", time.Time{}, appErrors.ErrDatabaseFailure
	}
	return tokenHash, expiresAt, nil
}

func (r *TokenRepo) IncrementAnonRequestCounter(ctx context.Context, anonID, path string, ttl time.Duration) (int, error) {
	key := "anon:" + anonID + ":" + path
	newValue, err := r.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// Set TTL only on key creation so the window is fixed, not sliding.
	if newValue == 1 {
		if err := r.redis.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, appErrors.ErrDatabaseFailure
		}
	}
	return int(newValue), nil
}

func (r *TokenRepo) StoreRefreshTokenHash(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID,
		tokenHash,
		expiresAt.UTC(),
	)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return appErrors.ErrDatabaseFailure
	}

	return nil
}

func (r *TokenRepo) DeleteRefreshTokenHash(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return appErrors.ErrDatabaseFailure
	}
	return nil
}

func (r *TokenRepo) SetSpotifyToken(ctx context.Context, newToken *oauth2.Token) error {
	accessToken := newToken.AccessToken
	expiresAt := newToken.Expiry.UTC()
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO spotify_tokens (singleton, access_token, expires_at) VALUES (true, $1, $2)
		 ON CONFLICT (singleton) DO UPDATE SET access_token = $1, expires_at = $2`,
		accessToken, expiresAt,
	)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return appErrors.ErrDatabaseFailure
	}
	r.token = newToken
	return nil
}

func (r *TokenRepo) GetStoredSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	if r.token != nil {
		return r.token, nil
	}

	var accessToken string
	var expiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT access_token, expires_at FROM spotify_tokens`).Scan(&accessToken, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Infof("no stored spotify token found: %T: %v", err, err)
			return nil, appErrors.ErrNotFound
		}
		r.logger.Errorf("db error: %T: %v", err, err)
		return nil, appErrors.ErrDatabaseFailure
	}
	token := oauth2.Token{
		AccessToken: accessToken,
	}
	if expiresAt.Valid {
		token.Expiry = expiresAt.Time
	}
	r.token = &token
	return &token, nil
}

func (r *TokenRepo) GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	config := &clientcredentials.Config{
		ClientID:     r.spotifyId,
		ClientSecret: r.spotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, translateSpotifyError(err)
	}
	// Ensure the token's expiry is in UTC
	token.Expiry = token.Expiry.UTC()

	return token, nil
}
