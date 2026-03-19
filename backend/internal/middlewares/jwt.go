package middlewares

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/handlers"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/services"
)

type JWTMiddleware struct {
	authService *services.AuthService
}

func NewJWTMiddleware(authService *services.AuthService) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
	}
}

func (m *JWTMiddleware) ProtectRoutes() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtStr, err := c.Cookie(models.CookieJWT)
		if err != nil {
			log.Printf("JWT cookie error: %v:%T", err, err)
			handlers.RespondWithError(c, appErrors.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := m.authService.ParseJWT(jwtStr)
		if err != nil {
			log.Printf("JWT parsing error: %v:%T", err, err)
			if errors.Is(err, jwt.ErrTokenExpired) {
				// JWT refresh flow
				refreshStr, err := c.Cookie(models.CookieRefreshToken)
				if err != nil {
					log.Printf("Refresh token cookie error: %v:%T", err, err)
					handlers.RespondWithError(c, appErrors.ErrUnauthorized)
					c.Abort()
					return
				}

				session, err := m.authService.ExchangeRefreshToken(c.Request.Context(), jwtStr, refreshStr)
				if err != nil {
					log.Printf("Refresh token exchange error: %v:%T", err, err)
					handlers.RespondWithError(c, appErrors.ErrUnauthorized)
					c.Abort()
					return
				}

				c.SetCookie(models.CookieJWT, session.JWT, models.JWTCookieAge, "/", "", false, true)
				c.SetCookie(models.CookieRefreshToken, session.RefreshToken, models.RefreshTokenCookieAge, "/", "", false, true)
				log.Println("successfully refreshed jwt")
			} else {
				handlers.RespondWithError(c, appErrors.ErrUnauthorized)
				c.Abort()
				return
			}
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}
