package middlewares

import (
	"net/http"

	"github.com/chivta/spotiscan/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	spotifyService *services.SpotifyService
}

func NewAuthMiddleware(spotifyService *services.SpotifyService) *AuthMiddleware {
	return &AuthMiddleware{
		spotifyService: spotifyService,
	}
}

func (m *AuthMiddleware) AttachSpotifyClientCreds() gin.HandlerFunc {
	return func(c *gin.Context) {
		spotifyToken, err := m.spotifyService.GetValidSpotifyToken()
		switch err {
		case nil:
			// all good
		case services.ErrSpotifyAPIError:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Spotify API error",
				"code":  "SPOTIFY_API_ERROR",
			})
			c.Abort()
			return
		case services.ErrDatabaseFailure:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
				"code":  "DATABASE_ERROR",
			})
			c.Abort()
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			})
			c.Abort()
			return
		}

		c.Set("spotify_tokens", spotifyToken)

		c.Next()
	}
}
