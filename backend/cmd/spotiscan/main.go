package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	// "github.com/chivta/spotiscan/internal/dbsqlite_migration"

	"github.com/chivta/spotiscan/internal/config"
	"github.com/chivta/spotiscan/internal/handlers"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/middlewares"
	"github.com/chivta/spotiscan/internal/services"
	"github.com/chivta/spotiscan/internal/repository"
)

func main() {
	// For migrating artists from old sqlite database to postgres database
	// sqlite_migration.Migrate()
	// return

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

	repo := repository.NewRepo(appLogger)

	err = repo.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	err = repo.InitRedis(cfg.RedisURL)
	if err != nil {
		// Not fatal, just log the error
		appLogger.Warnf("Failed to initialize redis: %v", err)
	}
	repo.InitSpotifyClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret)

	defer repo.Close()

	r := gin.New()

	r.SetTrustedProxies([]string{"172.16.0.0/12"})

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/", "/api/me"},
	}))
	r.Use(gin.Recovery())

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.Status(204)
	})

	spotifyService := services.NewSpotifyService(appLogger,repo)
	spotifyHandler := handlers.NewSpotifyHandler(spotifyService)

	authMiddleware := middlewares.NewAuthMiddleware(spotifyService)

	api := r.Group("/api")
	api.Use(authMiddleware.AttachSpotifyClientCreds())
	{
		api.GET("/playlist/:id/rucontent", spotifyHandler.GetPlaylistRuContent)
		// TODO: Add admin panel
	}

	r.Run()
}
