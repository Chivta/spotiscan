package redis_client

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Close() error
	RussianArtistsLoaded(ctx context.Context) (bool, error)
	SetRussianArtistNames(ctx context.Context, names []string) error
	FilterRussianArtistNames(ctx context.Context, names []string) ([]string, error)
}

type redisClient struct {
	client *redis.Client
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

func (r *redisClient) RussianArtistsLoaded(ctx context.Context) (bool, error) {
	key := "ru_artists"
	count, err := r.client.SCard(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
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