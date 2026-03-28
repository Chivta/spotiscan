package main

import (
	"context"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/chivta/spotiscan/internal/config"
	"github.com/chivta/spotiscan/internal/handlers"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/middlewares"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/repository"
	"github.com/chivta/spotiscan/internal/services"
	"github.com/chivta/spotiscan/scripts"
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

	db, err := repository.InitializeDatabase(cfg.DatabaseURL)
	if err != nil {
		return 1
	}
	defer db.Close()

	redis, err := repository.InitializeRedis(cfg.RedisURL)
	if err != nil {
		appLogger.Errorf("Failed to initialize redis (rate limiting and caching disabled): %v", err)
		return 1
	}
	defer redis.Close()


	repository.MigrateFromSQLite(appLogger, db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	
	ratelimitRepo := repository.NewRatelimitRepo(appLogger, redis)
	artistRepo := repository.NewArtistRepo(appLogger, db, redis)
	tokenRepo := repository.NewTokenRepo(appLogger, db, redis, cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	userRepo := repository.NewUserRepo(appLogger, db, redis)
	playlistRepo := repository.NewPlaylistRepo(appLogger, db, redis)
	
	scriptErr := ratelimitRepo.LoadRateLimitScript(ctx, scripts.RateLimitScript)
	cancel()
	if scriptErr != nil {
		appLogger.Errorf("Failed to load rate limit script to redis (rate limiting disabled): %v", scriptErr)
		return 1
	}
	
	// TODO: remove ctx background here, create init context with timeout for all preparations before starting the server
	artistRepo.LoadRussianArtistsToRedis(context.Background())

	r := gin.New()
	r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/", "/api/me"},
	}))
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) { c.Status(204) })

	validate := validator.New()

	spotifyService := services.NewSpotifyService(appLogger, artistRepo, playlistRepo, tokenRepo)
	spotifyHandler := handlers.NewSpotifyHandler(spotifyService, validate)

	authService := services.NewAuthService(appLogger, []byte(cfg.JWTSecret), tokenRepo, userRepo)
	authHandler := handlers.NewAuthHandler(authService, validate, cfg.SecureCookies)
	spotifyMiddleware := middlewares.NewSpotifyMiddleware(spotifyService)
	jwtMiddleware := middlewares.NewJWTMiddleware(authService, cfg.SecureCookies, appLogger)

	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(ratelimitRepo, appLogger)

	api := r.Group("/api")
	api.Use(jwtMiddleware.ParseAuth())
	api.Use(spotifyMiddleware.AttachSpotifyClientCreds())
	api.Use(rateLimitMiddleware.LimitRequests(cfg.RateLimit.RequestLimit, cfg.RateLimit.WindowSeconds))
	{
		api.GET("/me", authHandler.Me)
		api.POST("/auth/signup", authHandler.Signup)
		api.POST("/auth/login", authHandler.Login)
		api.GET("/playlist/:id/rucontent", jwtMiddleware.RequireAnonQuota("/playlist", models.AnonRequestLimit), spotifyHandler.GetPlaylistRuContent)

		userEndpoints := api.Group("")
		userEndpoints.Use(jwtMiddleware.RequireUserRole())
		{
			userEndpoints.POST("/auth/logout", authHandler.Logout)
		}
	}

	if err = r.Run(); err != nil {
		appLogger.Errorf("Failed to run server: %v", err)
		return 1
	}

	return 0
}
