package handlers

import (
	"errors"

	"github.com/chivta/ruscan/internal/shared/domain"
	"github.com/gin-gonic/gin"

	"github.com/rs/zerolog/log"
)

func RespondWithError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPCode, gin.H{"code": appErr.Code})
	} else {
		log.Error().Err(err).Msg("unexpected error")
		c.JSON(500, gin.H{"error": "internal server error", "code": "INTERNAL_ERROR"})
	}
}
