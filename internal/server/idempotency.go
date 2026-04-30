package server

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"sync"
)

type RedisIdempotency struct {
	c *redis.Client
}

func NewRedisIdempotency(c *redis.Client) *RedisIdempotency {
	return &RedisIdempotency{c: c}
}

func (r RedisIdempotency) Get(key string) ([]byte, error) {
	val, err := r.c.Get(context.Background(), "idempotency:"+key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return val, err
}

func (r RedisIdempotency) Set(key string, val []byte, exp time.Duration) error {
	return r.c.Set(context.Background(), "idempotency:"+key, val, exp).Err()
}

func (r RedisIdempotency) Delete(key string) error {
	return r.c.Del(context.Background(), "idempotency:"+key).Err()
}

func (r RedisIdempotency) Reset() error {
	ctx := context.Background()
	var cursor uint64
	for {
		keys, next, err := r.c.Scan(ctx, cursor, "idempotency:*", 128).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := r.c.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

func (r RedisIdempotency) Close() error {
	return nil
}

type RedisLocker struct {
	c      *redis.Client
	tokens sync.Map
}

func NewRedisLocker(c *redis.Client) *RedisLocker {
	return &RedisLocker{c: c}
}

const (
	redisLockTTL       = 2 * time.Minute
	redisLockRetryWait = 25 * time.Millisecond
)

var releaseLockScript = redis.NewScript(`
local current = redis.call("GET", KEYS[1])
if current == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

func (r *RedisLocker) Lock(key string) error {
	ctx := context.Background()
	lockKey := "lock:" + key
	token := uuid.NewString()

	for {
		acquired, err := r.c.SetNX(ctx, lockKey, token, redisLockTTL).Result()
		if err != nil {
			return err
		}
		if acquired {
			r.tokens.Store(lockKey, token)
			return nil
		}
		time.Sleep(redisLockRetryWait)
	}
}

func (r *RedisLocker) Unlock(key string) error {
	lockKey := "lock:" + key
	tokenValue, ok := r.tokens.LoadAndDelete(lockKey)
	if !ok {
		return nil
	}

	token, _ := tokenValue.(string)
	if token == "" {
		return nil
	}

	_, err := releaseLockScript.Run(context.Background(), r.c, []string{lockKey}, token).Result()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	return err
}
