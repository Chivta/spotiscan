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

const (
	userRoleKey = "userRole"
	userIDKey   = "userID"
)

func (m *JWTMiddleware) RequireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		assignedRole, exists := c.Get(userRoleKey)
		if !exists {
			m.log.Warnf("RequireRole: userRole not found in context")
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

func (m *JWTMiddleware) RequireUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		assignedRole, exists := c.Get(userRoleKey)
		if !exists {
			m.log.Warnf("RequireRole: userRole not found in context")
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

func (m *JWTMiddleware) ParseAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtStr, err := c.Cookie(models.CookieJWT)
		if err != nil {
			// No JWT cookie: treat as anonymous user
			m.log.Info("Issuing anon session ")
			anonSession, err := m.authService.CreateAnonymousSession(c.Request.Context())
			if err != nil {
				m.log.Errorf("Failed to create anonymous session: %v", err)
				handlers.RespondWithError(c, err)
				c.Abort()
				return
			}
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(models.CookieJWT, anonSession.JWT, models.AnonSessionCookieAge, "/", "", m.secureCookies, true)
			c.Set(userIDKey, "0")
			c.Set(userRoleKey, anonSession.Role)
			c.Next()
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

		c.Set(userIDKey, claims.UserID)
		c.Set(userRoleKey, claims.Role)
		c.Next()
	}
}

func (m *JWTMiddleware) RequireAnonQuota(limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		anonIDAny, exists := c.Get(userIDKey)
		if !exists {
			m.log.Warnf("RequireAnonQuota: userID not found in context")
			handlers.RespondWithError(c, appErrors.ErrInternal)
			c.Abort()
			return
		}
		anonID, ok := anonIDAny.(string)
		if !ok {
			m.log.Warnf("RequireAnonQuota: userID in context is not a string")
			handlers.RespondWithError(c, appErrors.ErrInternal)
			c.Abort()
			return
		}
		assignedRole, exists := c.Get(userRoleKey)
		if !exists {
			m.log.Warnf("RequireAnonQuota: userRole not found in context")
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
			m.log.Errorf("Failed to increment anon quota: %v", err)
			handlers.RespondWithError(c, appErrors.ErrInternal)
			c.Abort()
			return
		}

		if n > limit {
			handlers.RespondWithError(c, appErrors.ErrQuotaExceeded)
			c.Abort()
			return
		}

	}
}