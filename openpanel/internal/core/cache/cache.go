// Package cache wraps the Redis instance the panel uses for caching,
// reached over a unix socket.
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// KeyPrefix is prepended to every cache key.
	KeyPrefix = "openpanel_cache_"
	// DefaultTTL is the default cache entry lifetime.
	DefaultTTL = 300 * time.Second
)

type Cache struct {
	rdb *redis.Client
}

// New connects to Redis over a unix socket at socketPath.
func New(socketPath string) *Cache {
	return &Cache{
		rdb: redis.NewClient(&redis.Options{
			Network: "unix",
			Addr:    socketPath,
		}),
	}
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Raw exposes the underlying redis client for callers that need operations
// Memoize doesn't cover - e.g. a hand-rolled session store using
// hgetall/hset/expire/delete calls directly, which reads/writes a
// different keyspace than the memoize cache.
func (c *Cache) Raw() *redis.Client {
	return c.rdb
}

func (c *Cache) Close() error {
	return c.rdb.Close()
}

// Memoize returns the cached value for key if present, otherwise calls fn,
// caches its result for ttl, and returns it. The key is explicit rather
// than derived from the function name and arguments.
func Memoize[T any](ctx context.Context, c *Cache, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
	var zero T
	fullKey := KeyPrefix + key

	if raw, err := c.rdb.Get(ctx, fullKey).Bytes(); err == nil {
		var cached T
		if json.Unmarshal(raw, &cached) == nil {
			return cached, nil
		}
	}

	val, err := fn()
	if err != nil {
		return zero, err
	}

	if raw, err := json.Marshal(val); err == nil {
		c.rdb.Set(ctx, fullKey, raw, ttl)
	}

	return val, nil
}

// Delete removes a memoized key.
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, KeyPrefix+key).Err()
}
