package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/chivta/spotiscan/internal/config"
	"github.com/chivta/spotiscan/internal/handlers"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/middlewares"
	"github.com/chivta/spotiscan/internal/repository"
	"github.com/chivta/spotiscan/internal/repository/db_client"
	"github.com/chivta/spotiscan/internal/repository/redis_client"
	"github.com/chivta/spotiscan/internal/repository/spotify_client"
	"github.com/chivta/spotiscan/internal/services"

)

func main() {
	os.Exit(runApp())
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

	db, err := db_client.NewDBClient(cfg.DatabaseURL)
	if err != nil {
		appLogger.Errorf("Failed to initialize database: %v", err)
		return 1
	}
	defer db.Close()

	redis, err := redis_client.NewRedisClient(cfg.RedisURL)
	if err != nil {
		appLogger.Errorf("Failed to initialize redis: %v", err)
		return 1
	}
	defer redis.Close()
		
	err = redis.LoadRateLimitScript(context.Background(), cfg.RateLimit.RedisScriptFile)
	if err != nil {
		appLogger.Errorf("Failed to load rate limit script to redis: %v", err)
		return 1
	}

	spotifyClient := spotify_client.NewSpotifyClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	repo := repository.NewRepo(appLogger, db, redis, spotifyClient)
	repo.LoadRussianArtistsToRedis(context.Background())

	r := gin.New()
	r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})

	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/", "/api/me"},
	}))
	r.Use(gin.Recovery())

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.Status(204)
	})

	spotifyService := services.NewSpotifyService(appLogger, repo)
	spotifyHandler := handlers.NewSpotifyHandler(spotifyService)

	authMiddleware := middlewares.NewAuthMiddleware(spotifyService)
	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(redis)

	
	api := r.Group("/api")
	api.Use(authMiddleware.AttachSpotifyClientCreds())
	api.Use(rateLimitMiddleware.LimitRequests(cfg.RateLimit.RequestLimit, cfg.RateLimit.WindowSeconds))
	{
		api.GET("/playlist/:id/rucontent", spotifyHandler.GetPlaylistRuContent)
		// TODO: Add admin panel
	}

	err = r.Run()
	if err != nil {
		appLogger.Errorf("Failed to run server: %v", err)
		return 1
	}

	return 0
}