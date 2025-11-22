package main

import (
	"log"

	"spotiscan/internal/config"
	"spotiscan/internal/handlers"
	"spotiscan/internal/middlewares"
	"spotiscan/internal/services"
	"spotiscan/pkg/db"
	"spotiscan/pkg/spotify"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	db, err := db.NewDBConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create DB connection: %v", err)
	}
	defer db.Close()

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
		SkipPaths: []string{"/","/api/me"},
	}))
	r.Use(gin.Recovery())

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.Status(204)
	})

	spotifyClient := spotify.NewSpotifyClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRedirectURI)
	spotifyService := services.NewSpotifyService(db, spotifyClient)
	spotifyHandler := handlers.NewSpotifyHandler(spotifyService)

	userService := services.NewUserService(db)
	userHandler := handlers.NewUserHandler(userService)

	authService := services.NewAuthService(db, spotifyClient)
	authHandler := handlers.NewAuthHandler(authService, cfg.FrontendURL)

	authMiddleware := middlewares.NewAuthMiddleware(userService, spotifyService)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.GET("/start", authHandler.GetAuth)
			auth.GET("/callback", authHandler.GetCallback)
		}

		protected := api.Group("/")
		protected.Use(authMiddleware.RequireAuthentication())
		{
			protected.POST("/logout", authHandler.PostLogout)
			protected.GET("/me", userHandler.GetMe)
			protected.GET("/playlist/:id/rucontent", spotifyHandler.GetPlaylistRuContent)
			protected.GET("/user/liked-songs/rucontent", spotifyHandler.GetLikedSongsRuContent)
		}

	}

	r.Run()
}
