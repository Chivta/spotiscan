package redis_client

import (
    "os"
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
    Allow(ctx context.Context, key string, limit int, windowSeconds int) (bool, error)
    LoadRateLimitScript(ctx context.Context, scriptFile string) error

	Close() error
	SetRussianArtistNames(ctx context.Context, names []string) error
	FilterRussianArtistNames(ctx context.Context, names []string) ([]string, error)
}

type redisClient struct {
	client *redis.Client

    rateLimitScriptSHA string
}

func NewRedisClient(redisURL string) (RedisClient, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &redisClient{client: client}, nil
}

func (r *redisClient) Close() error {
	return r.client.Close()
}

func (r *redisClient) LoadRateLimitScript(ctx context.Context, scriptFile string) error {
    scriptContent, err := os.ReadFile(scriptFile)
    if err != nil {
        return err
    }
    script := redis.NewScript(string(scriptContent))
    sha, err := script.Load(ctx, r.client).Result()
    if err != nil {
        return err
    }

    r.rateLimitScriptSHA = sha
    return nil
}

func (r *redisClient) Allow(ctx context.Context, key string, limit int, windowSeconds int) (bool, error) {
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
func (r *redisClient) SetRussianArtistNames(ctx context.Context, names []string) error {
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

func (r *redisClient) FilterRussianArtistNames(ctx context.Context, names []string) ([]string, error) {
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