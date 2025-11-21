package middlewares

import (
	"net/http"
	"spotiscan/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	userService  *services.UserService
	spotifyService *services.SpotifyService
}

func NewAuthMiddleware(userService *services.UserService, spotifyService *services.SpotifyService) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
		spotifyService: spotifyService,
	}
}

func (m *AuthMiddleware) RequireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil || token == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userId, err := m.userService.GetUserIdBySessionToken(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
				"code":  "DATABASE_ERROR",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userId)

		spotifyToken, err := m.spotifyService.GetValidUserSpotifyToken(userId)
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
