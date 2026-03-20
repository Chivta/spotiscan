package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/handlers"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/services"
)

type JWTMiddleware struct {
	authService   *services.AuthService
	secureCookies bool
	log           *logger.Logger
}

func NewJWTMiddleware(authService *services.AuthService, secureCookies bool, appLogger *logger.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		authService:   authService,
		secureCookies: secureCookies,
		log:           appLogger,
	}
}

func (m *JWTMiddleware) ProtectRoutes() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtStr, err := c.Cookie(models.CookieJWT)
		if err != nil {
			m.log.Debugf("JWT cookie error: %v:%T", err, err)
			handlers.RespondWithError(c, appErrors.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := m.authService.ParseJWT(jwtStr)
		if err != nil {
			if !errors.Is(err, jwt.ErrTokenExpired) {
				handlers.RespondWithError(c, appErrors.ErrUnauthorized)
				c.Abort()
				return
			}

			// JWT expired — attempt refresh
			refreshStr, err := c.Cookie(models.CookieRefreshToken)
			if err != nil {
				m.log.Debugf("Refresh token cookie error: %v:%T", err, err)
				handlers.RespondWithError(c, appErrors.ErrUnauthorized)
				c.Abort()
				return
			}

			session, err := m.authService.ExchangeRefreshToken(c.Request.Context(), jwtStr, refreshStr)
			if err != nil {
				handlers.RespondWithError(c, appErrors.ErrUnauthorized)
				c.Abort()
				return
			}

			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(models.CookieJWT, session.JWT, models.JWTCookieAge, "/", "", m.secureCookies, true)
			c.SetCookie(models.CookieRefreshToken, session.RefreshToken, models.RefreshTokenCookieAge, "/", "", m.secureCookies, true)

			c.Set("userID", session.UserID)
			c.Set("userRole", session.Role)
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}
