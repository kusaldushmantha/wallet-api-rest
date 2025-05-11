package db

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	client *redis.Client
}

func NewRedisCache(cache *redis.Client) Cache {
	return &redisClient{
		client: cache,
	}
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	resp := r.client.Get(ctx, key)
	result, err := resp.Result()
	// This returns redis.NIL as an error if the key is not found.
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return result, err
}

func (r *redisClient) SetWithExpiration(ctx context.Context, key string, value string, duration time.Duration) error {
	resp := r.client.SetEx(ctx, key, value, duration)
	_, err := resp.Result()
	return err
}

func (r *redisClient) SetWithExpirationIfKeyIsNotSet(ctx context.Context, key string, value string, duration time.Duration) (bool, error) {
	resp := r.client.SetNX(ctx, key, value, duration)
	return resp.Result()
}
