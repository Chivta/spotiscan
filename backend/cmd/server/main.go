package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"spotiscan/internal/config"
	"spotiscan/internal/handlers"
	"spotiscan/internal/services"
	"spotiscan/internal/middlewares"
	"spotiscan/pkg/db"
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

	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health"},
	}))
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "I'm healthy!")
	})

	playlistService := services.NewPlaylistService(db)
	playlistHandler := handlers.NewPlaylistHandler(playlistService)

	userService := services.NewUserService(db)
	signupHandler := handlers.NewSignupHandler(userService)

	authMiddleware := middlewares.NewAuthMiddleware(db)

	api := r.Group("/api")
	{
		api.POST("/signup", signupHandler.PostSignup)
		auth := api.Group("/")
		auth.Use(authMiddleware.RequireSessionToken())
		{
			auth.GET("/playlist/ruartists", playlistHandler.GetRussianArtists)
		}
	}

	r.Run()
}
