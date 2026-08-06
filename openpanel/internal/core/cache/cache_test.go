package cache

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// These tests intentionally point at a redis socket that doesn't exist,
// since CI/dev sandboxes won't have the panel's redis instance running.
// Memoize must degrade gracefully (fall through to fn) rather than fail
// the request when the cache is unreachable.

func unreachableCache(t *testing.T) *Cache {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "no-redis-here.sock"))
}

func TestPingUnreachable(t *testing.T) {
	c := unreachableCache(t)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := c.Ping(ctx); err == nil {
		t.Error("expected Ping to fail against a nonexistent socket")
	}
}

func TestMemoizeFallsThroughWhenCacheUnreachable(t *testing.T) {
	c := unreachableCache(t)
	defer c.Close()

	ctx := context.Background()
	calls := 0

	val, err := Memoize(ctx, c, "some-key", DefaultTTL, func() (string, error) {
		calls++
		return "computed-value", nil
	})
	if err != nil {
		t.Fatalf("Memoize: %v", err)
	}
	if val != "computed-value" {
		t.Errorf("val = %q, want computed-value", val)
	}
	if calls != 1 {
		t.Errorf("fn called %d times, want 1", calls)
	}
}

func TestMemoizePropagatesFnError(t *testing.T) {
	c := unreachableCache(t)
	defer c.Close()

	wantErr := errors.New("boom")
	_, err := Memoize(context.Background(), c, "some-key", DefaultTTL, func() (string, error) {
		return "", wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}
