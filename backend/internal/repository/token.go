package repository

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/oauth2"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/appErrors"
	"github.com/redis/go-redis/v9"
)

func NewTokenRepo(db *sql.DB, redis *redis.Client, spotifyId, spotifySecret string) *TokenRepo {
	return &TokenRepo{
		db:            db,
		redis:         redis,
		spotifyId:     spotifyId,
		spotifySecret: spotifySecret,
	}
}

type TokenRepo struct {
	db    *sql.DB
	redis *redis.Client

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
		log.Error().Msgf("db error: %T: %v", err, err)
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
		log.Error().Msgf("db error: %T: %v", err, err)
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
		log.Error().Msgf("db error: %T: %v", err, err)
		return appErrors.ErrDatabaseFailure
	}
	return nil
}
