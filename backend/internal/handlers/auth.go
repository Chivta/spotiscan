package handlers

import (
	"log"
	"net/http"
	"spotiscan/internal/services"

	"github.com/gin-gonic/gin"
)

func NewAuthHandler(service *services.AuthService, frontendRedirectURL string) *AuthHandler {
	return &AuthHandler{
		svc: service,
		frontendRedirectURL: frontendRedirectURL,
	}
}

type AuthHandler struct {
	svc *services.AuthService
	frontendRedirectURL string
}

func (h *AuthHandler) GetAuth(c *gin.Context) {
	redirectUrl, err := h.svc.InitializeSpotifyAuth()
	switch err {
	case nil:
		log.Println("Redirect URL: ", redirectUrl)
		c.Redirect(http.StatusFound, redirectUrl)
	case services.ErrDatabaseFailure:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error"})
	}
}

func (h *AuthHandler) GetCallback(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing state"})
		return
	}

	sessionToken, err := h.svc.AcceptSpotifyAuthCallback(c.Request, state)

	switch err {
	case nil:
		c.SetCookie("session_token", sessionToken, 3600*24*7, "/", "", true, true)
		c.Redirect(http.StatusFound, h.frontendRedirectURL)
		return
	case services.ErrDatabaseFailure:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
			"code":  "DATABASE_ERROR",
		})
	case services.ErrInvalidState:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid OAuth state",
			"code":  "INVALID_OAUTH_STATE",
		})
	case services.ErrSpotifyAPIError:
		c.JSON(http.StatusBadGateway, gin.H{
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

func (s *AuthHandler) PostLogout(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	s.svc.Logout(sessionToken)
	c.SetCookie("session_token", "", -1, "/", "", true, true)
	c.Status(http.StatusNoContent)
}