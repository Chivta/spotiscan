package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"spotiscan/internal/config"
	"spotiscan/internal/handlers"
	"spotiscan/internal/middlewares"
	"spotiscan/internal/services"
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
		SkipPaths: []string{"/health"},
	}))
	r.Use(gin.Recovery())

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.Status(204)
	})

	playlistService := services.NewPlaylistService(db)
	playlistHandler := handlers.NewPlaylistHandler(playlistService)

	userService := services.NewUserService(db)
	userHandler := handlers.NewUserHandler(userService)

	authService := services.NewAuthService(db)
	authHandler := handlers.NewAuthHandler(authService)

	authMiddleware := middlewares.NewAuthMiddleware(db)

	api := r.Group("/api")
	{
		api.POST("/signup", authHandler.PostSignup)
		api.POST("/login", authHandler.PostLogin)

		auth := api.Group("/")
		auth.Use(authMiddleware.RequireSessionToken())

		auth.POST("/logout", authHandler.PostLogout)

		{
			auth.GET("/playlist/ruartists", playlistHandler.GetRussianArtists)
		}
		{
			auth.GET("/me", userHandler.GetMe)
		}
	}

	r.Run()
}
