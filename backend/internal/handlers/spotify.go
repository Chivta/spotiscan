package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/chivta/spotiscan/internal/services"
)

type spotifyMetrics interface {
	IncScans()
	ObserveScanDuration(seconds float64)
	IncErrors(errorType string)
}

func NewSpotifyHandler(service *services.SpotifyService, validate *validator.Validate, metrics spotifyMetrics) *SpotifyHandler {
	return &SpotifyHandler{svc: service, validate: validate, metrics: metrics}
}

type SpotifyHandler struct {
	svc      *services.SpotifyService
	validate *validator.Validate
	metrics  spotifyMetrics
}

func (h *SpotifyHandler) GetPlaylistRuContent(c *gin.Context) {
	ctx := context.WithValue(c.Request.Context(), "spotify_token", c.Value("spotify_token"))

	playlistId := c.Params.ByName("id")
	if playlistId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}

	start := time.Now()
	ruContent, err := h.svc.GetPlaylistRuContent(ctx, playlistId)
	if err != nil {
		h.metrics.IncErrors(errorTypeLabel(err))
		RespondWithError(c, err)
		return
	}

	h.metrics.IncScans()
	h.metrics.ObserveScanDuration(time.Since(start).Seconds())
	c.JSON(http.StatusOK, ruContent)
}
