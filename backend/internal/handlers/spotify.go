package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chivta/spotiscan/internal/services"
)

func NewSpotifyHandler(service *services.SpotifyService) *SpotifyHandler {
	return &SpotifyHandler{svc: service}
}

type SpotifyHandler struct {
	svc *services.SpotifyService
}

func (h *SpotifyHandler) GetPlaylistRuContent(c *gin.Context) {
	ctx := context.WithValue(c.Request.Context(),"spotify_token",c.Value("spotify_token"))
	
	playlistId := c.Params.ByName("id")
	if playlistId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	ruContent, err := h.svc.GetPlaylistRuContent(ctx, playlistId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, ruContent)
}
