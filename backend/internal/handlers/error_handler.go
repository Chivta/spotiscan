package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/chivta/spotiscan/internal/appErrors"
)

func RespondWithError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPCode, gin.H{"error": appErr.Message, "code": appErr.Code})
	} else {
		c.JSON(500, gin.H{"error": "internal server error", "code": "INTERNAL_ERROR"})
	}
}

// errorTypeLabel maps an error to a Prometheus label value for spotiscan_errors_total.
func errorTypeLabel(err error) string {
	var appErr *appErrors.AppError
	if !errors.As(err, &appErr) {
		return "internal"
	}
	switch appErr.Code {
	case "SPOTIFY_API_ERROR":
		return "spotify_api"
	case "DATABASE_ERROR":
		return "db"
	case "UNAUTHORIZED", "INVALID_CREDENTIALS", "FORBIDDEN":
		return "auth"
	default:
		return "internal"
	}
}
