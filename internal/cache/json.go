package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const cacheOperationTimeout = time.Second

func GetJSON[T any](ctx context.Context, client *redis.Client, key string, out *T) (bool, error) {
	if client == nil {
		return false, nil
	}

	cacheCtx, cancel := cacheContext(ctx)
	defer cancel()
	payload, err := client.Get(cacheCtx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read JSON cache: %w", err)
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return false, fmt.Errorf("decode JSON cache: %w", err)
	}
	return true, nil
}

func SetJSON(ctx context.Context, client *redis.Client, key string, value any, ttl time.Duration) error {
	if client == nil {
		return nil
	}
	if ttl <= 0 {
		return errors.New("cache TTL must be positive")
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode JSON cache: %w", err)
	}

	cacheCtx, cancel := cacheContext(ctx)
	defer cancel()
	if err := client.Set(cacheCtx, key, string(payload), ttl).Err(); err != nil {
		return fmt.Errorf("write JSON cache: %w", err)
	}
	return nil
}

func cacheContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, cacheOperationTimeout)
}
