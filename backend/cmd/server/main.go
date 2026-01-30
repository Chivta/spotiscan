package main

import (
	"log"

	"github.com/chivta/spotiscan/internal/config"
	"github.com/chivta/spotiscan/internal/db"
	"github.com/chivta/spotiscan/internal/handlers"
	"github.com/chivta/spotiscan/internal/middlewares"
	"github.com/chivta/spotiscan/internal/services"
	"github.com/chivta/spotiscan/internal/spotify_client"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	// "github.com/chivta/spotiscan/internal/dbsqlite_migration"
)

func main() {
	// For migrating artists from old sqlite database to postgres database
	// sqlite_migration.Migrate()
	// return

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
		SkipPaths: []string{"/", "/api/me"},
	}))
	r.Use(gin.Recovery())

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.Status(204)
	})

	spotifyClient := spotify_client.NewSpotifyClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	spotifyService := services.NewSpotifyService(db, spotifyClient)
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
