package helper

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

type fakeSessionRepo struct {
	version int64
	err     error
}

func (r fakeSessionRepo) GetSessionVersion(context.Context, int64) (int64, error) {
	return r.version, r.err
}

type fakeSessionCache struct {
	values map[string]int64
}

func (c *fakeSessionCache) GetInt64(_ context.Context, key string) (int64, error) {
	if c.values == nil {
		return 0, redis.Nil
	}
	v, ok := c.values[key]
	if !ok {
		return 0, redis.Nil
	}
	return v, nil
}

func (c *fakeSessionCache) SetInt64(_ context.Context, key string, val int64) error {
	if c.values == nil {
		c.values = make(map[string]int64)
	}
	c.values[key] = val
	return nil
}

func TestSessionVersionCheckerFallsBackToRepoAndCaches(t *testing.T) {
	cache := &fakeSessionCache{}
	checker := NewSessionVersionChecker(fakeSessionRepo{version: 7}, cache)

	got, err := checker.Current(context.Background(), 42)
	if err != nil {
		t.Fatalf("Current returned error: %v", err)
	}
	if got != 7 {
		t.Fatalf("expected version 7, got %d", got)
	}
	if cache.values[SessionVersionCacheKey(42)] != 7 {
		t.Fatalf("expected cache to store version 7, got %d", cache.values[SessionVersionCacheKey(42)])
	}
}

func TestSessionVersionCheckerRejectsStaleClaims(t *testing.T) {
	checker := NewSessionVersionChecker(fakeSessionRepo{version: 3}, &fakeSessionCache{})

	err := checker.ValidateClaims(context.Background(), &Claims{UserID: 99, SessionVersion: 2})
	if !errors.Is(err, ErrInvalidSessionVersion) {
		t.Fatalf("expected ErrInvalidSessionVersion, got %v", err)
	}
}
