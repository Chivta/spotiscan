package redis_client

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client

    rateLimitScriptSHA string
}

func NewRedisClient(redisURL string) (*RedisClient, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) LoadRateLimitScript(ctx context.Context, scriptContent string) error {
    script := redis.NewScript(scriptContent)
    sha, err := script.Load(ctx, r.client).Result()
    if err != nil {
        return err
    }

    r.rateLimitScriptSHA = sha
    return nil
}

func (r *RedisClient) Allow(ctx context.Context, key string, limit int, windowSeconds int) (bool, error) {
    result, err := r.client.EvalSha(ctx, r.rateLimitScriptSHA, []string{key}, limit, windowSeconds).Result()
    if err != nil {
        return false, err
    }

    allowed, ok := result.(int64)
    if !ok {
        return false, nil
    }

    return allowed == 1, nil
}

// SetRussianArtistNames: Store the list as a Redis set
func (r *RedisClient) SetRussianArtistNames(ctx context.Context, names []string) error {
    key := "ru_artists"
    // Clear existing set and add new names
    pipe := r.client.Pipeline()
    pipe.Del(ctx, key)
    if len(names) > 0 {
        pipe.SAdd(ctx, key, names)
    }
    _, err := pipe.Exec(ctx)
    return err
}

func (r *RedisClient) FilterRussianArtistNames(ctx context.Context, names []string) ([]string, error) {
    key := "ru_artists"
    if len(names) == 0 {
        return []string{}, nil
    }

    // Use a pipeline to batch SISMEMBER commands
    pipe := r.client.Pipeline()
    cmds := make([]*redis.BoolCmd, len(names))
    for i, name := range names {
        cmds[i] = pipe.SIsMember(ctx, key, name)
    }

    // Execute the pipeline
    _, err := pipe.Exec(ctx)
    if err != nil {
        return nil, err
    }

    // Collect results
    var ruNames []string
    for i, cmd := range cmds {
        isMember, err := cmd.Result()
        if err != nil {
            return nil, err
        }
        if isMember {
            ruNames = append(ruNames, names[i])
        }
    }
    return ruNames, nil
}

func (r *RedisClient) IncrementAnonRequestCounter(ctx context.Context, anonID, path string, ttl time.Duration) (int, error) {
    key := "anon:" + anonID + ":" + path
    newValue, err := r.client.Incr(ctx, key).Result()
    if err != nil {
        return 0, err
    }
    // Set TTL only on key creation so the window is fixed, not sliding.
    if newValue == 1 {
        r.client.Expire(ctx, key, ttl)
    }
    return int(newValue), nil
}