package redis

import (
	"context"
	
	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Close() error
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