package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log/slog"
	"time"
)

type RedisCache struct {
	client *redis.Client
	logger *slog.Logger
}

func NewRedisCache(addr, password string, db int, logger *slog.Logger) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisCache{client: client, logger: logger}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}
