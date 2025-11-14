package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"spotiscan/internal/config"
	"spotiscan/internal/handlers"
	"spotiscan/internal/services"
	"spotiscan/pkg/db"
)

type AuthMiddleware struct {
	db *db.DB
}

func (m *AuthMiddleware) RequireSessionToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session")
		if err != nil || token == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userId, err := m.db.GetUserIDBySessionToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		c.Set("user_id", userId)

		c.Next()
	}
}

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	db, err := db.NewDBConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create DB connection: %v", err)
	}

	r := gin.Default()

	playlistService := services.NewPlaylistService(db)
	playlistHandler := handlers.NewPlaylistHandler(playlistService)

	userService := services.NewUserService(db)
	signupHandler := handlers.NewSignupHandler(userService)

	authMiddleware := &AuthMiddleware{db: db}

	api := r.Group("/api")
	{
		api.POST("/signup", signupHandler.PostSignup)
		auth := api.Group("/")
		auth.Use(authMiddleware.RequireSessionToken())
		{
			auth.GET("/playlist/ruartists", playlistHandler.GetRussianArtists)
		}
	}

	r.Static("/static", "./static")

	// TODO: serve static files separately from api
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	r.Run()
}
