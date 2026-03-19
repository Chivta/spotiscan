package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/chivta/spotiscan/internal/handlers"
	"github.com/chivta/spotiscan/internal/services"
)

type SpotifyMiddleware struct {
	spotifyService *services.SpotifyService
}

func NewSpotifyMiddleware(spotifyService *services.SpotifyService) *SpotifyMiddleware {
	return &SpotifyMiddleware{
		spotifyService: spotifyService,
	}
}

func (m *SpotifyMiddleware) AttachSpotifyClientCreds() gin.HandlerFunc {
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
