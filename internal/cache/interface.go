package cache

import "context"

type Cache interface {
	Set(ctx context.Context, key, val string) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	GetBytes(ctx context.Context, key string) ([]byte, error)
	SetTimed(ctx context.Context, key, val string, ttl int64) error
	SetTimedInt64(ctx context.Context, key string, val int64, ttl int64) error
	SetInt64(ctx context.Context, key string, val int64) error
	SetTTL(ctx context.Context, key string, ttl int64) error
	Incr(ctx context.Context, key string) (int64, error)
	GetInt64(ctx context.Context, key string) (int64, error)
	SetJSON(ctx context.Context, key string, val interface{}) error
	SetTimedJSON(ctx context.Context, key string, val interface{}, ttl int64) error
	GetJSON(ctx context.Context, key string, v interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key, field, value string) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error
}
