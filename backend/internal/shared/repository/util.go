package repository

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"
	"github.com/chivta/ruscan/internal/spotify"
	"github.com/chivta/ruscan/migrations"
)

func InitializeDatabase(ctx context.Context, dbUrl string) (*sql.DB, error) {
	// db_client.NewDBClient blocks on Ping; wrap it in a goroutine so we can
	// apply a timeout without needing a context-aware driver.
	dbCh := make(chan *sql.DB, 1)
	errCh := make(chan error, 1)
	go func() {
		db, err := sql.Open("postgres", dbUrl)
		if err != nil {
			errCh <- err
		}
		err = db.Ping()
		if err != nil {
			errCh <- err
			db.Close()
			return
		}
		dbCh <- db
	}()

	var db *sql.DB
	select {
	case result := <-dbCh:
		db = result
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return db, nil
}

func RunMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)

	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	return goose.UpContext(ctx, db, ".")
}

// logs unexpected Spotify API and network errors before translating them to application errors.
func translateSpotifyError(err error) error {
	if spotifyErr, ok := err.(*spotify.SpotifyError); ok {
		switch spotifyErr.Status {
		case 404:
			return domain.ErrSpotifyNotFound
		case 400:
			return domain.ErrBadRequest
		case 429:
			return domain.ErrTooManyRequests
		default:
			log.Error().Err(spotifyErr).Int("status", spotifyErr.Status).Msg("spotify API error")
			return domain.ErrSpotifyAPIError
		}
	}
	if _, ok := err.(*url.Error); ok {
		log.Error().Err(err).Msg("network error when calling spotify API")
		return domain.ErrSpotifyAPIError
	}
	log.Error().Err(err).Msg("unexpected error when calling spotify API")
	return domain.ErrSpotifyAPIError
}

func InitializeRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	redis := redis.NewClient(options)

	_, err = redis.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return redis, nil
}
