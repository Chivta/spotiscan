package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/chivta/ruscan/internal/appErrors"
	"github.com/rs/zerolog/log"
)

func RespondWithError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPCode, gin.H{"error": appErr.Message, "code": appErr.Code})
	} else {
		log.Error().Err(err).Msg("unexpected error")
		c.JSON(500, gin.H{"error": "internal server error", "code": "INTERNAL_ERROR"})
	}
}
