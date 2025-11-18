package handlers

import (
	"net/http"
	"spotiscan/internal/services"

	"github.com/gin-gonic/gin"
)

func NewSpotifyHandler(service *services.SpotifyService, frontendRedirectURL string) *SpotifyHandler {
	return &SpotifyHandler{
		svc: service,
		frontendRedirectURL: frontendRedirectURL,
	}
}

type SpotifyHandler struct {
	svc *services.SpotifyService
	frontendRedirectURL string
}

func (h *SpotifyHandler) GetPlaylistRussianArtists(c *gin.Context) {
	playlistId := c.Query("id")
	if playlistId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	userId := c.GetInt("user_id")
	ruArtists, err := h.svc.GetRuArtistsFromPlaylist(playlistId, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spotify api error"})
		return
	}
	c.JSON(http.StatusOK, ruArtists)
}

func (h *SpotifyHandler) GetAuth(c *gin.Context) {
	userId := c.GetInt("user_id")
	redirectUrl,err := h.svc.InitializeSpotifyAuth(userId)
	switch err { 
	case nil:
		c.Redirect(http.StatusFound, redirectUrl)
	case services.ErrDatabaseFailure:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error"})
	}
}
func (h *SpotifyHandler) PostCallback(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing state"})
		return
	}
	userId := c.GetInt("user_id")

	err := h.svc.AcceptCallback(c.Request,state,userId)

	switch err {
	case nil:
		c.Redirect(http.StatusFound,h.frontendRedirectURL)
	case services.ErrInvalidState:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
	case services.ErrSpotifyAPIError:
		c.JSON(http.StatusBadGateway, gin.H{"error": "spotify api error"})
	case services.ErrDatabaseFailure:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error"})
	}
}
