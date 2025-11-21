package handlers

import (
	"log"
	"net/http"
	"spotiscan/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func NewSpotifyHandler(service *services.SpotifyService) *SpotifyHandler {
	return &SpotifyHandler{svc: service}
}

type SpotifyHandler struct {
	svc *services.SpotifyService
}


func (h *SpotifyHandler) GetPlaylistRussianArtists(c *gin.Context) {
	playlistId := c.Query("id")
	if playlistId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	userId := c.MustGet("user_id").(int)
	tokens := c.MustGet("spotify_tokens").(*oauth2.Token)
	log.Println(userId,tokens)
	ruArtists, err := h.svc.GetRuArtistsFromPlaylist(playlistId, userId,tokens)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spotify api error"})
		return
	}
	c.JSON(http.StatusOK, ruArtists)
}