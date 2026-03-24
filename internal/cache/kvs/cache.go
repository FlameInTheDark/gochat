package kvs

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Cache struct {
	c *redis.Client
}

func New(addr string) (*Cache, error) {
	addr = normalizeAddr(addr)

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, end := observability.StartDependencySpan(context.Background(), "redis", "ping", addr)
	err := client.Ping(ctx).Err()
	end(err)
	return &Cache{c: client}, err
}

func normalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "localhost:6379"
	}
	if strings.Contains(addr, "://") {
		return addr
	}
	if _, _, err := net.SplitHostPort(addr); err == nil {
		return addr
	}
	if strings.HasPrefix(addr, "[") && strings.HasSuffix(addr, "]") {
		return net.JoinHostPort(strings.TrimSuffix(strings.TrimPrefix(addr, "["), "]"), "6379")
	}
	if strings.Count(addr, ":") >= 2 || !strings.Contains(addr, ":") {
		return net.JoinHostPort(addr, "6379")
	}

	var addrErr *net.AddrError
	if _, _, err := net.SplitHostPort(addr); errors.As(err, &addrErr) && strings.Contains(addrErr.Err, "missing port in address") {
		return net.JoinHostPort(addr, "6379")
	}

	return addr
}

func (c *Cache) Client() *redis.Client {
	return c.c
}

func (c *Cache) Close() error {
	return c.c.Close()
}

// Set string value
func (c *Cache) Set(ctx context.Context, key, val string) error {
	ctx, end := c.operation(ctx, "set", key)
	err := c.c.Set(ctx, key, val, 0).Err()
	end(err)
	return err
}

// Get string value
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	ctx, end := c.operation(ctx, "get", key)
	res := c.c.Get(ctx, key)
	err := res.Err()
	finishCacheOperation(ctx, end, err)
	return res.String(), err
}

// Delete key
func (c *Cache) Delete(ctx context.Context, key string) error {
	ctx, end := c.operation(ctx, "delete", key)
	err := c.c.Del(ctx, key).Err()
	end(err)
	return err
}

func (c *Cache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	ctx, end := c.operation(ctx, "get_bytes", key)
	res := c.c.Get(ctx, key)
	val, err := res.Bytes()
	finishCacheOperation(ctx, end, err)
	return val, err
}

// SetTimed set string value with expiration time in seconds
func (c *Cache) SetTimed(ctx context.Context, key, val string, ttl int64) error {
	ctx, end := c.operation(ctx, "set_timed", key)
	err := c.c.Set(ctx, key, val, time.Duration(ttl)*time.Second).Err()
	end(err)
	return err
}

// SetTimedInt64 set int64 value with expiration time in seconds
func (c *Cache) SetTimedInt64(ctx context.Context, key string, val int64, ttl int64) error {
	ctx, end := c.operation(ctx, "set_timed_int64", key)
	err := c.c.Set(ctx, key, val, time.Duration(ttl)*time.Second).Err()
	end(err)
	return err
}

// SetInt64 set int64 value
func (c *Cache) SetInt64(ctx context.Context, key string, val int64) error {
	ctx, end := c.operation(ctx, "set_int64", key)
	err := c.c.Set(ctx, key, val, 0).Err()
	end(err)
	return err
}

// SetTTL set expiration time for key
func (c *Cache) SetTTL(ctx context.Context, key string, ttl int64) error {
	ctx, end := c.operation(ctx, "expire", key)
	err := c.c.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
	end(err)
	return err
}

// Incr increment numerical value
func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	ctx, end := c.operation(ctx, "incr", key)
	res := c.c.Incr(ctx, key)
	err := res.Err()
	end(err)
	return res.Val(), err
}

// GetInt64 return int64 value of key
func (c *Cache) GetInt64(ctx context.Context, key string) (int64, error) {
	ctx, end := c.operation(ctx, "get_int64", key)
	res := c.c.Get(ctx, key)
	val, err := res.Int64()
	finishCacheOperation(ctx, end, err)
	return val, err
}

// SetJSON marshal set marshaled json of val
func (c *Cache) SetJSON(ctx context.Context, key string, val interface{}) error {
	msg, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ctx, end := c.operation(ctx, "set_json", key)
	err = c.c.Set(ctx, key, string(msg), 0).Err()
	end(err)
	return err
}

func (c *Cache) SetTimedJSON(ctx context.Context, key string, val interface{}, ttl int64) error {
	msg, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ctx, end := c.operation(ctx, "set_timed_json", key)
	err = c.c.Set(ctx, key, string(msg), time.Duration(ttl)*time.Second).Err()
	end(err)
	return err
}

func (c *Cache) SetTimedJSONNX(ctx context.Context, key string, val interface{}, ttl int64) (bool, error) {
	msg, err := json.Marshal(val)
	if err != nil {
		return false, err
	}
	ctx, end := c.operation(ctx, "set_timed_json_nx", key)
	res := c.c.SetArgs(ctx, key, string(msg), redis.SetArgs{
		TTL:  time.Duration(ttl) * time.Second,
		Mode: "NX",
	})
	if err := res.Err(); err != nil && err != redis.Nil {
		end(err)
		return false, err
	}
	end(nil)
	return res.Val() == "OK", nil
}

// GetJSON unmarshal json into v
func (c *Cache) GetJSON(ctx context.Context, key string, v interface{}) error {
	ctx, end := c.operation(ctx, "get_json", key)
	res := c.c.Get(ctx, key)
	if res.Err() != nil {
		err := res.Err()
		finishCacheOperation(ctx, end, err)
		return err
	}
	b, err := res.Bytes()
	if err != nil {
		finishCacheOperation(ctx, end, err)
		return err
	}
	err = json.Unmarshal(b, v)
	finishCacheOperation(ctx, end, err)
	return err
}

func (c *Cache) HGet(ctx context.Context, key, field string) (string, error) {
	ctx, end := c.operation(ctx, "hget", key)
	h := c.c.HGet(ctx, key, field)
	if h.Err() != nil {
		finishCacheOperation(ctx, end, h.Err())
		return "", nil
	}
	end(nil)
	return h.Val(), nil
}

func (c *Cache) HSet(ctx context.Context, key, field, value string) error {
	ctx, end := c.operation(ctx, "hset", key)
	h := c.c.HSet(ctx, key, field, value)
	if h.Err() != nil {
		err := h.Err()
		end(err)
		return err
	}
	end(nil)
	return nil
}

func (c *Cache) HDel(ctx context.Context, key, field string) error {
	ctx, end := c.operation(ctx, "hdel", key)
	h := c.c.HDel(ctx, key, field)
	err := h.Err()
	end(err)
	return err
}

func (c *Cache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	ctx, end := c.operation(ctx, "hgetall", key)
	h := c.c.HGetAll(ctx, key)
	if h.Err() != nil {
		err := h.Err()
		end(err)
		return nil, err
	}
	end(nil)
	return h.Val(), nil
}

func (c *Cache) HIncrBy(ctx context.Context, key, field string, delta int64) (int64, error) {
	ctx, end := c.operation(ctx, "hincrby", key)
	h := c.c.HIncrBy(ctx, key, field, delta)
	if h.Err() != nil {
		err := h.Err()
		end(err)
		return 0, err
	}
	end(nil)
	return h.Val(), nil
}

func (c *Cache) ZAdd(ctx context.Context, key string, score float64, member string) error {
	ctx, end := c.operation(ctx, "zadd", key)
	err := c.c.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
	end(err)
	return err
}

func (c *Cache) ZRem(ctx context.Context, key string, members ...string) error {
	ctx, end := c.operation(ctx, "zrem", key)
	args := make([]interface{}, 0, len(members))
	for _, member := range members {
		args = append(args, member)
	}
	err := c.c.ZRem(ctx, key, args...).Err()
	end(err)
	return err
}

func (c *Cache) ZRevRangeByScore(ctx context.Context, key, max, min string, offset, count int64) ([]string, error) {
	ctx, end := c.operation(ctx, "zrevrangebyscore", key)
	res := c.c.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Max:    max,
		Min:    min,
		Offset: offset,
		Count:  count,
	})
	if res.Err() != nil {
		err := res.Err()
		finishCacheOperation(ctx, end, err)
		return nil, err
	}
	end(nil)
	return res.Val(), nil
}

func (c *Cache) XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error {
	ctx, end := c.operation(ctx, "xadd", stream)
	h := c.c.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: maxLen,
		Approx: approx,
		Values: values,
	})
	if h.Err() != nil {
		err := h.Err()
		end(err)
		return err
	}
	end(nil)
	return nil
}

func (c *Cache) operation(ctx context.Context, operation, key string) (context.Context, func(error)) {
	return observability.StartDependencySpan(ctx, "redis", operation, cacheTarget(key))
}

func cacheTarget(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "unknown"
	}
	if idx := strings.Index(key, ":"); idx > 0 {
		return key[:idx]
	}
	return key
}

func finishCacheOperation(ctx context.Context, end func(error), err error) {
	if end == nil {
		return
	}
	if errors.Is(err, redis.Nil) {
		observability.SetDependencyResult(ctx, "miss")
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.Bool("cache.miss", true),
			attribute.String("dependency.result", "miss"),
		)
		span.AddEvent("cache.miss")
		end(nil)
		return
	}
	end(err)
}
