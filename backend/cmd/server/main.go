package main

import (
	"github.com/gin-gonic/gin"

	"spotiscan/internal/handlers"
	"spotiscan/internal/services"
)

func main() {
	r := gin.Default()

	playlistService := services.NewPlaylistService()
	playlistHandler := handlers.NewPlaylistHandler(playlistService)
	r.GET("/playlist/ruartists", playlistHandler.GetRussianArtists)

	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	r.Run()
}
