package middlewares

import (
	"github.com/gin-gonic/gin"
	"spotiscan/pkg/db"
)


type AuthMiddleware struct {
	db *db.DB
}

func NewAuthMiddleware(db *db.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

func (m *AuthMiddleware) RequireSessionToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil || token == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userId, err := m.db.GetUserIDBySessionToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		c.Set("user_id", userId)

		c.Next()
	}
}