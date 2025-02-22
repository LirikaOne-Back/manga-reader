package cache

import (
	"context"
	"time"
)

// Cache описывает базовый интерфейс кэширования.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expiration time.Duration) error
}
