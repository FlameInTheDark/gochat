package server

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisIdempotency struct {
	c *redis.Client
}

func NewRedisIdempotency(c *redis.Client) *RedisIdempotency {
	return &RedisIdempotency{c: c}
}

func (r RedisIdempotency) Get(key string) ([]byte, error) {
	return r.c.Get(context.Background(), "idempotency:"+key).Bytes()
}

func (r RedisIdempotency) Set(key string, val []byte, exp time.Duration) error {
	return r.c.Set(context.Background(), "idempotency:"+key, val, exp).Err()
}

func (r RedisIdempotency) Delete(key string) error {
	return r.c.Del(context.Background(), "idempotency:"+key).Err()
}

func (r RedisIdempotency) Reset() error {
	return r.c.Del(context.Background(), "idempotency:*").Err()
}

func (r RedisIdempotency) Close() error {
	return nil
}

type RedisLocker struct {
	c *redis.Client
}

func NewRedisLocker(c *redis.Client) *RedisLocker {
	return &RedisLocker{c: c}
}

func (r RedisLocker) Lock(key string) error {
	return r.c.Set(context.Background(), "lock:"+key, "1", time.Second*10).Err()
}

func (r RedisLocker) Unlock(key string) error {
	return r.c.Del(context.Background(), "lock:"+key).Err()
}
