package main

import (
	"context"
	"database/sql"
	"fmt"
	stdlog "log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/config"
	"github.com/chivta/ruscan/internal/handlers"
	"github.com/chivta/ruscan/internal/metrics"
	"github.com/chivta/ruscan/internal/middlewares"
	"github.com/chivta/ruscan/internal/models"
	"github.com/chivta/ruscan/internal/repository"
	"github.com/chivta/ruscan/internal/services"
	"github.com/chivta/ruscan/internal/spotify"
	"github.com/chivta/ruscan/scripts"
)

func main() {
	os.Exit(runApp())
}

func runApp() int {
	cfg, err := config.Load()
	if err != nil {
		stdlog.Fatalf("Failed to load config: %v", err)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	// redirect stdlib log to zerolog
	stdlog.SetOutput(log.Logger)
	stdlog.SetFlags(0)

	c, err := initApp(cfg)
	if err != nil {
		log.Err(err).Msg("Failed to initialize app")
		return 1
	}
	defer c.db.Close()
	defer c.redis.Close()

	r := gin.New()
	r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
	r.Use(middlewares.Logger("/", "/health", "/api/me", "/metrics"))
	r.Use(gin.Recovery())
	r.Use(metrics.Middleware("/metrics", "/health", "/", "/api/me"))
	r.GET("/metrics", gin.WrapH(metrics.Handler()))
	r.GET("/health", func(c *gin.Context) { c.Status(204) })

	r.GET("/", func(c *gin.Context) { c.Status(204) })

	api := r.Group("/api")
	api.Use(c.authMiddleware.ParseAuth())
	api.Use(c.rateLimitMiddleware.LimitRequests(models.RateLimitRequestLimit, models.RateLimitWindowSeconds))
	{
		api.GET("/me", c.authHandler.Me)
		api.POST("/auth/signup", c.authHandler.Signup)
		api.POST("/auth/login", c.authHandler.Login)

		spotifyEndpoints := api.Group("/spotify")
		spotifyEndpoints.GET("/playlist/:id/rucontent", c.authMiddleware.RequireAnonQuota("/scan", models.AnonRequestLimit), c.spotifyHandler.GetPlaylistRuContent)
		spotifyEndpoints.GET("/track/:id/rucontent", c.authMiddleware.RequireAnonQuota("/scan", models.AnonRequestLimit), c.spotifyHandler.GetTrackRuContent)
		spotifyEndpoints.GET("/album/:id/rucontent", c.authMiddleware.RequireAnonQuota("/scan", models.AnonRequestLimit), c.spotifyHandler.GetAlbumRuContent)
		spotifyEndpoints.GET("/artist/:id/rucontent", c.authMiddleware.RequireAnonQuota("/scan", models.AnonRequestLimit), c.spotifyHandler.GetArtistRuContent)
		spotifyEndpoints.GET("/artist/name/:name/rucontent", c.authMiddleware.RequireAnonQuota("/scan", models.AnonRequestLimit), c.spotifyHandler.GetArtistRuContentByName)

		userEndpoints := api.Group("")
		userEndpoints.Use(c.authMiddleware.RequireUserRole())
		{
			userEndpoints.POST("/auth/logout", c.authHandler.Logout)
			userEndpoints.GET("/suggestions/artist-insert", c.suggestionHandler.GetArtistInsertSuggestions)
			userEndpoints.POST("/suggestions/artist-insert", c.suggestionHandler.CreateArtistInsertSuggestion)
			userEndpoints.DELETE("/suggestions/artist-insert/:id", c.suggestionHandler.DeleteArtistInsertSuggestion)
			userEndpoints.PUT("/suggestions/artist-insert/:id", c.suggestionHandler.UpdateArtistInsertSuggestion)
			userEndpoints.GET("/suggestions/artist-delete", c.suggestionHandler.GetArtistDeleteSuggestions)
			userEndpoints.POST("/suggestions/artist-delete", c.suggestionHandler.CreateArtistDeleteSuggestion)
			userEndpoints.DELETE("/suggestions/artist-delete/:id", c.suggestionHandler.DeleteArtistDeleteSuggestion)
			userEndpoints.PUT("/suggestions/artist-delete/:id", c.suggestionHandler.UpdateArtistDeleteSuggestion)
		}

		adminEndpoints := api.Group("/admin")
		adminEndpoints.Use(c.authMiddleware.RequireAdminRole())
		{
			adminEndpoints.GET("/suggestions/artist-insert", c.suggestionHandler.GetAllArtistInsertSuggestions)
			adminEndpoints.POST("/suggestions/artist-insert/:id/approve", c.suggestionHandler.ApproveArtistInsertSuggestion)
			adminEndpoints.POST("/suggestions/artist-insert/:id/decline", c.suggestionHandler.DeclineArtistInsertSuggestion)
			adminEndpoints.GET("/suggestions/artist-delete", c.suggestionHandler.GetAllArtistDeleteSuggestions)
			adminEndpoints.POST("/suggestions/artist-delete/:id/approve", c.suggestionHandler.ApproveArtistDeleteSuggestion)
			adminEndpoints.POST("/suggestions/artist-delete/:id/decline", c.suggestionHandler.DeclineArtistDeleteSuggestion)
		}
	}

	if err = r.Run(); err != nil {
		log.Err(err).Msg("Failed to run server")
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
	authService         *services.AuthService
	authMiddleware      *middlewares.AuthMiddleware
	spotifyHandler      *handlers.SpotifyHandler
	spotifyService      *services.SpotifyService
	suggestionHandler   *handlers.SuggestionHandler
	suggestionService   *services.SuggestionService
	suggestionRepo      *repository.SuggestionRepo
	rateLimitMiddleware *middlewares.RateLimitMiddleware
	db                  *sql.DB
	redis               *redis.Client
}

func initApp(cfg *config.Config) (*appContainer, error) {
	initCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	db, err := repository.InitializeDatabase(initCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	err = repository.RunMigrations(initCtx, db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	redis, err := repository.InitializeRedis(initCtx, cfg.RedisURL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}
	spotifyClient := spotify.NewSpotifyClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	artistRepo := repository.NewArtistRepo(db, redis)
	tokenRepo := repository.NewTokenRepo(db, redis, cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	userRepo := repository.NewUserRepo(db, redis)
	playlistRepo := repository.NewPlaylistRepo(db, spotifyClient)

	ratelimitRepo := repository.NewRatelimitRepo(redis)
	err = ratelimitRepo.LoadRateLimitScript(initCtx, scripts.RateLimitScript)
	if err != nil {
		redis.Close()
		db.Close()
		return nil, fmt.Errorf("Failed to load rate limit script to redis: %v", err)
	}

	err = artistRepo.LoadRussianArtistsToRedis(initCtx)
	if err != nil {
		redis.Close()
		db.Close()
		return nil, fmt.Errorf("Failed to load Russian artists to redis: %v", err)
	}
	validate := validator.New()

	spotifyService := services.NewSpotifyService(artistRepo, playlistRepo)
	spotifyHandler := handlers.NewSpotifyHandler(spotifyService, validate)

	authService := services.NewAuthService([]byte(cfg.JWTSecret), tokenRepo, userRepo)
	authHandler := handlers.NewAuthHandler(authService, validate, cfg.SecureCookies)
	jwtMiddleware := middlewares.NewAuthMiddleware(authService, cfg.SecureCookies)

	suggestionRepo := repository.NewSuggestionRepo(db, redis)
	suggestionService := services.NewSuggestionService(suggestionRepo, artistRepo)
	suggestionHandler := handlers.NewSuggestionHandler(suggestionService, validate)

	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(ratelimitRepo)

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
		authMiddleware:      jwtMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
		suggestionHandler:   suggestionHandler,
		suggestionService:   suggestionService,
		suggestionRepo:      suggestionRepo,
		db:                  db,
		redis:               redis,
	}, nil
}
