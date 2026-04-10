package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/chivta/ruscan/internal/models"
	"github.com/chivta/ruscan/internal/services"
)

func NewSpotifyHandler(service *services.SpotifyService, validate *validator.Validate) *SpotifyHandler {
	return &SpotifyHandler{svc: service, validate: validate}
}

type SpotifyHandler struct {
	svc      *services.SpotifyService
	validate *validator.Validate
}

func (h *SpotifyHandler) GetPlaylistRuContent(c *gin.Context) {
	userRole, _ := c.Get(models.UserRoleKey)
	ctx := context.WithValue(c.Request.Context(), models.UserRoleKey, userRole)

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

func (h *SpotifyHandler) GetTrackRuContent(c *gin.Context) {
	userRole, _ := c.Get(models.UserRoleKey)
	ctx := context.WithValue(c.Request.Context(), models.UserRoleKey, userRole)

	trackId := c.Params.ByName("id")
	if trackId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	ruContent, err := h.svc.GetTrackRuContent(ctx, trackId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, ruContent)
}

func (h *SpotifyHandler) GetAlbumRuContent(c *gin.Context) {
	userRole, _ := c.Get(models.UserRoleKey)
	ctx := context.WithValue(c.Request.Context(), models.UserRoleKey, userRole)

	albumId := c.Params.ByName("id")
	if albumId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	ruContent, err := h.svc.GetAlbumRuContent(ctx, albumId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, ruContent)
}

func (h *SpotifyHandler) GetArtistRuContent(c *gin.Context) {
	userRole, _ := c.Get(models.UserRoleKey)
	ctx := context.WithValue(c.Request.Context(), models.UserRoleKey, userRole)

	artistId := c.Params.ByName("id")
	if artistId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	ruContent, err := h.svc.GetArtistRuContent(ctx, artistId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, ruContent)
}

func (h *SpotifyHandler) GetArtistRuContentByName(c *gin.Context) {
	userRole, _ := c.Get(models.UserRoleKey)
	ctx := context.WithValue(c.Request.Context(), models.UserRoleKey, userRole)

	artistName := c.Params.ByName("name")
	if artistName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty name"})
		return
	}
	ruContent, err := h.svc.GetArtistRuContentByName(ctx, artistName)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, ruContent)
}