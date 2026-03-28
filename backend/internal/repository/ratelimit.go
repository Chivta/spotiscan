package repository

import (
	"context"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/redis/go-redis/v9"
)

func NewRatelimitRepo(logger *logger.Logger, redis *redis.Client) *RatelimitRepo {
	return &RatelimitRepo{
		logger:  logger,
		redis:   redis,
	}
}

type RatelimitRepo struct {
	logger  *logger.Logger
	redis   *redis.Client

	rateLimitScriptSHA string
}

func (r *RatelimitRepo) LoadRateLimitScript(ctx context.Context, scriptContent string) error {
	script := redis.NewScript(scriptContent)
	sha, err := script.Load(ctx, r.redis).Result()
	if err != nil {
		return err
	}

	r.rateLimitScriptSHA = sha
	return nil
}

func (r *RatelimitRepo) Allow(ctx context.Context, key string, limit int, windowSeconds int) (bool, error) {
	result, err := r.redis.EvalSha(ctx, r.rateLimitScriptSHA, []string{key}, limit, windowSeconds).Result()
	if err != nil {
		return false, err
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, nil
	}

	return allowed == 1, nil
}
