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
		if rl.redisClient == nil {
			c.Next()
			return
		}
		clientIP := c.ClientIP()
		allowed, err := rl.redisClient.Allow(c.Request.Context(), RateLimitContextKey+clientIP, limit, windowSeconds)
		if err != nil {
			// Redis error: let the request through rather than blocking all traffic
			c.Next()
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(429, gin.H{"error": "Too many requests"})
			return
		}
		c.Next()
	}
}
