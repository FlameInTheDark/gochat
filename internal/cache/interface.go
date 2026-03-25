package cache

import "context"

// ZBatchMember is a (score, member) pair for ZAddBatch.
type ZBatchMember struct {
	Score  float64
	Member string
}

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
	// SetTimedJSONNX marshals val and sets it only if the key does not already exist (SET NX).
	// Returns true if the key was set, false if it already existed.
	SetTimedJSONNX(ctx context.Context, key string, val interface{}, ttl int64) (bool, error)
	GetJSON(ctx context.Context, key string, v interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key, field, value string) error
	HDel(ctx context.Context, key, field string) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	// HGetAllMulti pipelines multiple HGETALL commands in one round-trip.
	// Returns one map per key in the same order; nil maps mean the key was empty/missing.
	HGetAllMulti(ctx context.Context, keys []string) ([]map[string]string, error)
	// MGetBytes fetches multiple string keys in a single MGET round-trip.
	// Returns one []byte per key; nil entries mean key-not-found.
	MGetBytes(ctx context.Context, keys ...string) ([][]byte, error)
	HIncrBy(ctx context.Context, key, field string, delta int64) (int64, error)
	// SetTimedJSONBatch pipelines multiple timed JSON SET commands in one round-trip.
	// keys[i] is the Redis key for vals[i].
	SetTimedJSONBatch(ctx context.Context, keys []string, vals []interface{}, ttl int64) error
	// ZAddBatch adds multiple members to a sorted set in a single ZADD command.
	ZAddBatch(ctx context.Context, key string, members []ZBatchMember) error
	ZAdd(ctx context.Context, key string, score float64, member string) error
	ZRem(ctx context.Context, key string, members ...string) error
	ZRevRangeByScore(ctx context.Context, key, max, min string, offset, count int64) ([]string, error)
	XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error
}
