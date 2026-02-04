package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/chivta/spotiscan/internal/errors"
)

func RespondWithError(c *gin.Context, err error) {
	switch err {
	case nil:
		// no error, do nothing
	case errors.ErrNotFound:
		c.JSON(404, gin.H{"error": "playlist not found", "code": "PLAYLIST_NOT_FOUND"})
	case errors.ErrBadRequest:
		c.JSON(400, gin.H{"error": "bad request", "code": "BAD_REQUEST"})
	case errors.ErrDatabaseFailure:
		c.JSON(500, gin.H{"error": "database error", "code": "DATABASE_ERROR"})
	case errors.ErrSpotifyAPIError:
		c.JSON(500, gin.H{"error": "spotify api error", "code": "SPOTIFY_API_ERROR"})
	case errors.ErrInternal:
		c.JSON(500, gin.H{"error": "internal server error", "code": "INTERNAL_ERROR"})
	default:
		c.JSON(500, gin.H{"error": "internal server error", "code": "INTERNAL_ERROR"})
	}
}
