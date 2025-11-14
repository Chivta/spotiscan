package handlers

import (
	"log"
	"net/http"
	"spotiscan/internal/services"

	"github.com/gin-gonic/gin"
)

func NewPlaylistHandler(service *services.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{svc: service}
}

type PlaylistHandler struct {
	svc *services.PlaylistService
}

func (h *PlaylistHandler) GetRussianArtists(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	log.Println(id)
	ruArtists, err := h.svc.GetRuArtists(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println(ruArtists)
	c.JSON(http.StatusOK, ruArtists)
}
