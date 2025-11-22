package handlers

import (
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

func (h *SpotifyHandler) GetPlaylistRuContent(c *gin.Context) {
	playlistId := c.Params.ByName("id")
	if playlistId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	tokens := c.MustGet("spotify_tokens").(*oauth2.Token)
	ruArtists, err := h.svc.GetPlaylistRuContent(playlistId, tokens)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spotify api error"})
		return
	}
	c.JSON(http.StatusOK, ruArtists)
}

func (h *SpotifyHandler) GetLikedSongsRuContent(c *gin.Context) {
	tokens := c.MustGet("spotify_tokens").(*oauth2.Token)

	ruContent, err := h.svc.GetUserLikedSongsRuContent(tokens)
	switch err {
	case nil:
		c.JSON(http.StatusOK, ruContent)
	case services.ErrSpotifyAPIError:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Spotify API error",
			"code":  "SPOTIFY_API_ERROR",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	}
}