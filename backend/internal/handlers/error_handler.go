package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/chivta/spotiscan/internal/appErrors"
)

func RespondWithError(c *gin.Context, err error) {
	switch err {
	case nil:
		// no error, do nothing
	case appErrors.ErrNotFound:
		c.JSON(404, gin.H{"error": "playlist not found", "code": "PLAYLIST_NOT_FOUND"})
	case appErrors.ErrUnauthorized:
		c.JSON(401, gin.H{"error": "unauthorized", "code": "UNAUTHORIZED"})
	case appErrors.ErrBadRequest:
		c.JSON(400, gin.H{"error": "bad request", "code": "BAD_REQUEST"})
	case appErrors.ErrDatabaseFailure:
		c.JSON(500, gin.H{"error": "database error", "code": "DATABASE_ERROR"})
	case appErrors.ErrSpotifyAPIError:
		c.JSON(500, gin.H{"error": "spotify api error", "code": "SPOTIFY_API_ERROR"})
	case appErrors.ErrEmailExists:
		c.JSON(409, gin.H{"error": "email already exists", "code": "EMAIL_EXISTS"})
	case appErrors.ErrInvalidCredentials:
		c.JSON(401, gin.H{"error": "invalid credentials", "code": "INVALID_CREDENTIALS"})
	case appErrors.ErrInternal:
		c.JSON(500, gin.H{"error": "internal server error", "code": "INTERNAL_ERROR"})
	default:
		c.JSON(500, gin.H{"error": "internal server error", "code": "INTERNAL_ERROR"})
	}
}
