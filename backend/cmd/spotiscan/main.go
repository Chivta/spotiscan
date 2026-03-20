package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pressly/goose/v3"

	"github.com/chivta/spotiscan/internal/config"
	"github.com/chivta/spotiscan/internal/handlers"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/middlewares"
	"github.com/chivta/spotiscan/internal/repository"
	"github.com/chivta/spotiscan/internal/repository/db_client"
	"github.com/chivta/spotiscan/internal/repository/redis_client"
	"github.com/chivta/spotiscan/internal/repository/spotify_client"
	"github.com/chivta/spotiscan/internal/services"
	"github.com/chivta/spotiscan/migrations"
	"github.com/chivta/spotiscan/scripts"
)

func main() {
	os.Exit(runApp())
}

func runMigrations(db *db_client.DBClient) error {
	goose.SetBaseFS(migrations.FS)

	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	return goose.Up(db.GetConnection(), ".")
}

// migrateFromSQLite checks for a bot_data.db file and, if present, bulk-inserts
// all artist names into postgres. Safe to call on every startup — the INSERT uses
// ON CONFLICT DO NOTHING so duplicates are silently skipped.
func migrateFromSQLite(appLogger *logger.Logger, db *db_client.DBClient) {
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

	_, err = db.GetConnection().Exec(
		"INSERT INTO ru_artists (name) SELECT unnest($1::text[]) ON CONFLICT (name) DO NOTHING",
		pq.Array(artistsSlice),
	)
	if err != nil {
		appLogger.Warnf("Failed to insert artists into PostgreSQL: %v", err)
		return
	}

	appLogger.Infof("Successfully migrated %d artists from SQLite", len(artistsSlice))
}

func initializeDatabase(dbUrl string, appLogger *logger.Logger) (*db_client.DBClient, error) {
	// db_client.NewDBClient blocks on Ping; wrap it in a goroutine so we can
	// apply a timeout without needing a context-aware driver.
	dbCh := make(chan *db_client.DBClient, 1)
	errCh := make(chan error, 1)
	go func() {
		client, err := db_client.NewDBClient(dbUrl)
		if err != nil {
			errCh <- err
			return
		}
		dbCh <- client
	}()

	var db *db_client.DBClient
	select {
	case result := <-dbCh:
		db = result
		appLogger.Infof("Successfully connected to PostgreSQL")
	case err := <-errCh:
		appLogger.Errorf("Failed to initialize postgres: %v", err)
		return nil, err
	case <-time.After(5 * time.Second):
		appLogger.Errorf("Could not initialize postgres: connection timed out")
		return nil, context.DeadlineExceeded
	}

	if err := runMigrations(db); err != nil {
		appLogger.Errorf("Failed to run database migrations: %v", err)
		return nil, err
	}
	appLogger.Infof("Database migrations completed successfully")
	return db, nil
}

func runApp() int {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.NewLogger(
		cfg.Log.EnableDebug,
		cfg.Log.EnableInfo,
		cfg.Log.ErrorOutput,
		cfg.Log.InfoOutput,
		cfg.Log.DebugOutput,
	)

	db, err := initializeDatabase(cfg.DatabaseURL, appLogger)
	if err != nil {
		return 1
	}
	defer db.Close()

	migrateFromSQLite(appLogger, db)

	var cacheClient repository.CacheClient
	redisClient, err := redis_client.NewRedisClient(cfg.RedisURL)
	if err != nil {
		appLogger.Warnf("Failed to initialize redis (rate limiting and caching disabled): %v", err)
	} else {
		defer redisClient.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		scriptErr := redisClient.LoadRateLimitScript(ctx, scripts.RateLimitScript)
		cancel()
		if scriptErr != nil {
			appLogger.Warnf("Failed to load rate limit script to redis (rate limiting disabled): %v", scriptErr)
		} else {
			cacheClient = redisClient
		}
	}

	spotifyClient := spotify_client.NewSpotifyClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	repo := repository.NewRepo(appLogger, db, cacheClient, spotifyClient)
	repo.LoadRussianArtistsToRedis(context.Background())

	r := gin.New()
	r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/", "/api/me"},
	}))
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) { c.Status(204) })

	validate := validator.New()

	spotifyService := services.NewSpotifyService(appLogger, repo)
	spotifyHandler := handlers.NewSpotifyHandler(spotifyService, validate)

	authService := services.NewAuthService(repo, []byte(cfg.JWTSecret))
	authHandler := handlers.NewAuthHandler(authService, validate, cfg.SecureCookies)
	spotifyMiddleware := middlewares.NewSpotifyMiddleware(spotifyService)
	jwtMiddleware := middlewares.NewJWTMiddleware(authService, cfg.SecureCookies, appLogger)

	var rateLimitCache middlewares.Cache
	if redisClient != nil {
		rateLimitCache = redisClient
	}
	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(rateLimitCache, appLogger)

	api := r.Group("/api")
	api.Use(spotifyMiddleware.AttachSpotifyClientCreds())
	api.Use(rateLimitMiddleware.LimitRequests(cfg.RateLimit.RequestLimit, cfg.RateLimit.WindowSeconds))
	{	
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
		}

		prot := api.Group("")
		prot.Use(jwtMiddleware.ProtectRoutes())
		{
			prot.POST("/auth/logout", authHandler.Logout)
			prot.GET("/me", authHandler.Me)
			prot.GET("/playlist/:id/rucontent", spotifyHandler.GetPlaylistRuContent)

		}

		// TODO: Add admin panel
	}

	if err = r.Run(); err != nil {
		appLogger.Errorf("Failed to run server: %v", err)
		return 1
	}

	return 0
}
