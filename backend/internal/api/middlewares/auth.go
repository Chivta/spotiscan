package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/appErrors"
	"github.com/chivta/ruscan/internal/api/handlers"
	"github.com/chivta/ruscan/internal/shared/models"
	"github.com/chivta/ruscan/internal/api/services"
)

type AuthMiddleware struct {
	authService   *services.AuthService
	secureCookies bool
}

func NewAuthMiddleware(authService *services.AuthService, secureCookies bool) *AuthMiddleware {
	return &AuthMiddleware{
		authService:   authService,
		secureCookies: secureCookies,
	}
}

func (m *AuthMiddleware) RequireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		assignedRole, exists := c.Get(models.UserRoleKey)
		if !exists {
			log.Warn().Msg("RequireRole: userRole not found in context")
			handlers.RespondWithError(c, appErrors.ErrInternal)
			c.Abort()
			return
		}

		if assignedRole != models.RoleAdmin {
			handlers.RespondWithError(c, appErrors.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RequireUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		assignedRole, exists := c.Get(models.UserRoleKey)
		if !exists {
			log.Warn().Msg("RequireRole: userRole not found in context")
			handlers.RespondWithError(c, appErrors.ErrInternal)
			c.Abort()
			return
		}

		if assignedRole != models.RoleUser && assignedRole != models.RoleAdmin {
			handlers.RespondWithError(c, appErrors.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) issueAnonSession(c *gin.Context) {
	anonSession, err := m.authService.CreateAnonymousSession(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create anonymous session")
		handlers.RespondWithError(c, err)
		c.Abort()
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(models.CookieJWT, anonSession.JWT, models.AnonSessionCookieAge, "/", "", m.secureCookies, true)
	c.Set(models.UserIDKey, anonSession.UserID)
	c.Set(models.UserRoleKey, anonSession.Role)
}

func (m *AuthMiddleware) ParseAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtStr, err := c.Cookie(models.CookieJWT)
		if err != nil {
			// No JWT cookie: treat as anonymous user
			m.issueAnonSession(c)
			c.Next()
			return
		}

		claims, err := m.authService.ParseJWT(jwtStr)
		if err != nil {
			if !errors.Is(err, jwt.ErrTokenExpired) {
				c.Abort()
				c.SetSameSite(http.SameSiteLaxMode)
				c.SetCookie(models.CookieJWT, "", -1, "/", "", m.secureCookies, true)
				c.SetCookie(models.CookieRefreshToken, "", -1, "/", "", m.secureCookies, true)
				handlers.RespondWithError(c, appErrors.ErrUnauthorized)
				return
			}

			if claims.Role == models.RoleAnon {
				// Anon JWT expired — issue new one without hitting the DB (since we don't store anon sessions)
				m.issueAnonSession(c)
				c.Next()
				return
			}

			// JWT expired — attempt refresh
			refreshStr, err := c.Cookie(models.CookieRefreshToken)
			if err != nil {
				c.Abort()
				c.SetSameSite(http.SameSiteLaxMode)
				c.SetCookie(models.CookieJWT, "", -1, "/", "", m.secureCookies, true)
				c.SetCookie(models.CookieRefreshToken, "", -1, "/", "", m.secureCookies, true)
				handlers.RespondWithError(c, appErrors.ErrUnauthorized)
				return
			}

			session, err := m.authService.ExchangeRefreshToken(c.Request.Context(), jwtStr, refreshStr)
			if err != nil {
				c.Abort()
				c.SetSameSite(http.SameSiteLaxMode)
				c.SetCookie(models.CookieJWT, "", -1, "/", "", m.secureCookies, true)
				c.SetCookie(models.CookieRefreshToken, "", -1, "/", "", m.secureCookies, true)
				handlers.RespondWithError(c, appErrors.ErrUnauthorized)
				return
			}

			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(models.CookieJWT, session.JWT, models.JWTCookieAge, "/", "", m.secureCookies, true)
			c.SetCookie(models.CookieRefreshToken, session.RefreshToken, models.RefreshTokenCookieAge, "/", "", m.secureCookies, true)

			c.Set(models.UserIDKey, session.UserID)
			c.Set(models.UserRoleKey, session.Role)
			c.Next()
			return
		}

		c.Set(models.UserIDKey, claims.UserID)
		c.Set(models.UserRoleKey, claims.Role)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireAnonQuota(path string, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		anonIDAny, exists := c.Get(models.UserIDKey)
		if !exists {
			log.Warn().Msg("RequireAnonQuota: userID not found in context")
			handlers.RespondWithError(c, appErrors.ErrInternal)
			c.Abort()
			return
		}
		anonID, ok := anonIDAny.(string)
		if !ok {
			log.Warn().Msg("RequireAnonQuota: userID in context is not a string")
			handlers.RespondWithError(c, appErrors.ErrInternal)
			c.Abort()
			return
		}
		assignedRole, exists := c.Get(models.UserRoleKey)
		if !exists {
			log.Warn().Msg("RequireAnonQuota: userRole not found in context")
			handlers.RespondWithError(c, appErrors.ErrInternal)
			c.Abort()
			return
		}

		if assignedRole != models.RoleAnon {
			c.Next()
			return
		}

		n, err := m.authService.IncrementAnonQuota(c.Request.Context(), anonID, path)
		if err != nil {
			log.Error().Err(err).Msg("Failed to increment anon quota")
			handlers.RespondWithError(c, appErrors.ErrInternal)
			c.Abort()
			return
		}

		if n > limit {
			handlers.RespondWithError(c, appErrors.ErrQuotaExceeded)
			c.Abort()
			return
		}

		c.Next()
	}
}
