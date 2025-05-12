package db

import (
	"context"
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

func (r *redisClient) SetWithExpirationIfKeyIsNotSet(ctx context.Context, key string, value string, duration time.Duration) (bool, error) {
	resp := r.client.SetNX(ctx, key, value, duration)
	return resp.Result()
}
