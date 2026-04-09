package middlewares

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/chivta/ruscan/internal/repository"
)

type Cache interface {
	Allow(ctx context.Context, key string, limit int, windowSeconds int) (bool, error)
}

func NewRateLimitMiddleware(ratelimitRepo *repository.RatelimitRepo) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		ratelimitRepo: ratelimitRepo,
	}
}

type RateLimitMiddleware struct {
	ratelimitRepo *repository.RatelimitRepo
}

const (
	RateLimitContextKey = "ratelimit:"
)

func (rl *RateLimitMiddleware) LimitRequests(limit int, windowSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		allowed, err := rl.ratelimitRepo.Allow(c.Request.Context(), RateLimitContextKey+clientIP, limit, windowSeconds)
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
