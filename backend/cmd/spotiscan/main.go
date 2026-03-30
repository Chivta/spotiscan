package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"

	"github.com/chivta/spotiscan/internal/config"
	"github.com/chivta/spotiscan/internal/handlers"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/middlewares"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/repository"
	"github.com/chivta/spotiscan/internal/services"
	"github.com/chivta/spotiscan/internal/spotify"
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

	// TODO: remove hardcodes
	appLogger := logger.NewLogger(
		true,
		true,
		"stdout",
		"stdout",
		"stdout",
	)

	c, err := initApp(cfg, appLogger)
	if err != nil {
		log.Printf("Failed to initialize app: %v", err)
		return 1
	}
	defer c.db.Close()
	defer c.redis.Close()

	r := gin.New()
	r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/", "/api/me"},
	}))
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) { c.Status(204) })

	api := r.Group("/api")
	api.Use(c.jwtMiddleware.ParseAuth())
	// api.Use(c.spotifyMiddleware.AttachSpotifyClientCreds())
	api.Use(c.rateLimitMiddleware.LimitRequests(models.RateLimitRequestLimit, models.RateLimitWindowSeconds))
	{
		api.GET("/me", c.authHandler.Me)
		api.POST("/auth/signup", c.authHandler.Signup)
		api.POST("/auth/login", c.authHandler.Login)
		api.GET("/playlist/:id/rucontent", c.jwtMiddleware.RequireAnonQuota("/playlist", models.AnonRequestLimit), c.spotifyHandler.GetPlaylistRuContent)

		userEndpoints := api.Group("")
		userEndpoints.Use(c.jwtMiddleware.RequireUserRole())
		{
			userEndpoints.POST("/auth/logout", c.authHandler.Logout)
		}
	}

	if err = r.Run(); err != nil {
		appLogger.Errorf("Failed to run server: %v", err)
		return 1
	}

	return 0
}

type appContainer struct {
	ratelimitRepo       *repository.RatelimitRepo
	artistRepo          *repository.ArtistRepo
	tokenRepo           *repository.TokenRepo
	userRepo            *repository.UserRepo
	playlistRepo        *repository.PlaylistRepo
	authHandler         *handlers.AuthHandler
	spotifyHandler      *handlers.SpotifyHandler
	authService         *services.AuthService
	spotifyService      *services.SpotifyService
	jwtMiddleware       *middlewares.JWTMiddleware
	rateLimitMiddleware *middlewares.RateLimitMiddleware
	db                  *sql.DB
	redis               *redis.Client
}

func initApp(cfg *config.Config, appLogger *logger.Logger) (*appContainer, error) {
	initCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	db, err := repository.InitializeDatabase(initCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	err = repository.RunMigrations(initCtx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	redis, err := repository.InitializeRedis(initCtx, cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}
	spotifyClient := spotify.NewSpotifyClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	ratelimitRepo := repository.NewRatelimitRepo(appLogger, redis)
	artistRepo := repository.NewArtistRepo(appLogger, db, redis)
	tokenRepo := repository.NewTokenRepo(appLogger, db, redis, cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	userRepo := repository.NewUserRepo(appLogger, db, redis)
	playlistRepo := repository.NewPlaylistRepo(appLogger, db, redis, spotifyClient)

	scriptErr := ratelimitRepo.LoadRateLimitScript(initCtx, scripts.RateLimitScript)
	if scriptErr != nil {
		appLogger.Errorf("Failed to load rate limit script to redis: %v", scriptErr)
		return nil, scriptErr
	}

	err = artistRepo.LoadRussianArtistsToRedis(initCtx)
	if err != nil {
		appLogger.Errorf("Failed to load Russian artists to redis: %v", err)
		return nil, err
	}
	validate := validator.New()

	spotifyService := services.NewSpotifyService(appLogger, artistRepo, playlistRepo)
	spotifyHandler := handlers.NewSpotifyHandler(spotifyService, validate)

	authService := services.NewAuthService(appLogger, []byte(cfg.JWTSecret), tokenRepo, userRepo)
	authHandler := handlers.NewAuthHandler(authService, validate, cfg.SecureCookies)
	jwtMiddleware := middlewares.NewJWTMiddleware(authService, cfg.SecureCookies, appLogger)

	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(ratelimitRepo, appLogger)

	return &appContainer{
		ratelimitRepo:       ratelimitRepo,
		artistRepo:          artistRepo,
		tokenRepo:           tokenRepo,
		userRepo:            userRepo,
		playlistRepo:        playlistRepo,
		authHandler:         authHandler,
		spotifyHandler:      spotifyHandler,
		authService:         authService,
		spotifyService:      spotifyService,
		jwtMiddleware:       jwtMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
		db:                  db,
		redis:               redis,
	}, nil
}
