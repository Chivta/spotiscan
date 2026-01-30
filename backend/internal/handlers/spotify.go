package handlers

import (
	"net/http"

	"github.com/chivta/spotiscan/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func NewSpotifyHandler(service *services.SpotifyService) *SpotifyHandler {
	return &SpotifyHandler{svc: service}
}

type SpotifyHandler struct {
	svc *services.SpotifyService
}

func (h *SpotifyHandler) GetPlaylistRuContent(c *gin.Context) {
	playlistId := c.Params.ByName("id")
	if playlistId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	tokens := c.MustGet("spotify_tokens").(*oauth2.Token)
	ruContent, err := h.svc.GetPlaylistRuContent(playlistId, tokens)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spotify api error"})
		return
	}
	c.JSON(http.StatusOK, ruContent)
}
