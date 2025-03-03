package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

// Cache описывает базовый интерфейс кеширования.
type Cache interface {
	// Базовые операции ключ-значение
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Операции со списками
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// Операции со множествами
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, members ...interface{}) error

	// Операции со счетчиками
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)

	// Операции с отсортированными множествами (для рейтингов)
	ZAdd(ctx context.Context, key string, score float64, member string) error
	ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error)
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) (map[string]float64, error)
	GetClient() *redis.Client
}
