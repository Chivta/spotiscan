package middlewares

import (
	"context"
	"github.com/gin-gonic/gin"
)

type Cache interface {
	Allow(ctx context.Context, key string, limit int, windowSeconds int) (bool, error)
}

func NewRateLimitMiddleware(redisClient Cache) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient: redisClient,
	}
}

type RateLimitMiddleware struct {
	redisClient Cache
}

const (
	RateLimitContextKey = "ratelimit:"
)

func (rl *RateLimitMiddleware) LimitRequests(limit int, windowSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		allowed, err := rl.redisClient.Allow(c.Request.Context(), RateLimitContextKey+clientIP, limit, windowSeconds)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(429, gin.H{"error": "Too many requests"})
			return
		}
		c.Next()
	}
}
