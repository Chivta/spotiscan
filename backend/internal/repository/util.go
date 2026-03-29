package repository

import (
	"context"
	"database/sql"
	"net/url"
	"os"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/zmb3/spotify/v2"

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/migrations"
)

// TODO: propage context here
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
		return nil, context.DeadlineExceeded
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

// migrateFromSQLite checks for a bot_data.db file and, if present, bulk-inserts
// all artist names into postgres. Safe to call on every startup — the INSERT uses
// ON CONFLICT DO NOTHING so duplicates are silently skipped.
func MigrateFromSQLite(appLogger *logger.Logger, db *sql.DB) {
	if _, err := os.Stat("bot_data.db"); os.IsNotExist(err) {
		appLogger.Infof("bot_data.db not found, skipping SQLite migration")
		return
	}

	appLogger.Infof("bot_data.db found, migrating artists to PostgreSQL")

	sqliteDB, err := sql.Open("sqlite", "file:bot_data.db")
	if err != nil {
		appLogger.Warnf("Failed to open bot_data.db: %v", err)
		return
	}
	defer sqliteDB.Close()

	rows, err := sqliteDB.Query("SELECT name FROM artists")
	if err != nil {
		appLogger.Warnf("Failed to query artists from SQLite: %v", err)
		return
	}
	defer rows.Close()

	artists := make(map[string]struct{}, 25138) // known size of the old db
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			appLogger.Warnf("Failed to scan artist row: %v", err)
			return
		}
		artists[name] = struct{}{}
	}

	artistsSlice := make([]string, 0, len(artists))
	for name := range artists {
		artistsSlice = append(artistsSlice, name)
	}
	artists = nil // free memory before the bulk insert

	if len(artistsSlice) == 0 {
		appLogger.Infof("No artists found in bot_data.db, nothing to migrate")
		return
	}

	_, err = db.Exec(
		"INSERT INTO ru_artists (name) SELECT unnest($1::text[]) ON CONFLICT (name) DO NOTHING",
		pq.Array(artistsSlice),
	)
	if err != nil {
		appLogger.Warnf("Failed to insert artists into PostgreSQL: %v", err)
		return
	}

	appLogger.Infof("Successfully migrated %d artists from SQLite", len(artistsSlice))
}

func translateSpotifyError(err error) error {
	if spotifyErr, ok := err.(spotify.Error); ok {
		switch spotifyErr.Status {
		case 404:
			return appErrors.ErrPlaylistNotFound
		case 400:
			return appErrors.ErrBadRequest
		default:
			return appErrors.ErrSpotifyAPIError
		}
	}
	if _, ok := err.(*url.Error); ok {
		return appErrors.ErrSpotifyAPIError
	}
	return appErrors.ErrSpotifyAPIError
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
