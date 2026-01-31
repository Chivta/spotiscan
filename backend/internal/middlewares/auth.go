package middlewares

import (
	"github.com/gin-gonic/gin"
	
	"github.com/chivta/spotiscan/internal/services"
	"github.com/chivta/spotiscan/internal/handlers"
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
		spotifyToken, err := m.spotifyService.GetValidSpotifyToken(c.Request.Context())
		if err != nil {
			handlers.RespondWithError(c, err)
			c.Abort()
			return
		}
		c.Set("spotify_token", spotifyToken)

		c.Next()
	}
}
