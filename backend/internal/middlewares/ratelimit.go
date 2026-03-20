package middlewares

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/chivta/spotiscan/internal/logger"
)

type Cache interface {
	Allow(ctx context.Context, key string, limit int, windowSeconds int) (bool, error)
}

func NewRateLimitMiddleware(redisClient Cache, appLogger *logger.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient: redisClient,
		log:         appLogger,
	}
}

type RateLimitMiddleware struct {
	redisClient Cache
	log         *logger.Logger
}

const (
	RateLimitContextKey = "ratelimit:"
)

func (rl *RateLimitMiddleware) LimitRequests(limit int, windowSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl.redisClient == nil {
			rl.log.Warnf("rate limiting disabled: redis unavailable")
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
