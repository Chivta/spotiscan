package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/chivta/spotiscan/internal/services"
)

func NewSpotifyHandler(service *services.SpotifyService, validate *validator.Validate) *SpotifyHandler {
	return &SpotifyHandler{svc: service, validate: validate}
}

type SpotifyHandler struct {
	svc      *services.SpotifyService
	validate *validator.Validate
}

func (h *SpotifyHandler) GetPlaylistRuContent(c *gin.Context) {
	ctx := context.WithValue(c.Request.Context(), "spotify_token", c.Value("spotify_token"))

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
