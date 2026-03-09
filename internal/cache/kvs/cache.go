package kvs

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	c *redis.Client
}

func New(addr string) (*Cache, error) {
	addr = normalizeAddr(addr)

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Cache{c: client}, client.Ping(context.Background()).Err()
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
	res := c.c.Set(ctx, key, val, 0)
	return res.Err()
}

// Get string value
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	res := c.c.Get(ctx, key)
	return res.String(), res.Err()
}

// Delete key
func (c *Cache) Delete(ctx context.Context, key string) error {
	res := c.c.Del(ctx, key)
	return res.Err()
}

func (c *Cache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	res := c.c.Get(ctx, key)
	return res.Bytes()
}

// SetTimed set string value with expiration time in seconds
func (c *Cache) SetTimed(ctx context.Context, key, val string, ttl int64) error {
	res := c.c.Set(ctx, key, val, time.Duration(ttl)*time.Second)
	return res.Err()
}

// SetTimedInt64 set int64 value with expiration time in seconds
func (c *Cache) SetTimedInt64(ctx context.Context, key string, val int64, ttl int64) error {
	res := c.c.Set(ctx, key, val, time.Duration(ttl)*time.Second)
	return res.Err()
}

// SetInt64 set int64 value
func (c *Cache) SetInt64(ctx context.Context, key string, val int64) error {
	res := c.c.Set(ctx, key, val, 0)
	return res.Err()
}

// SetTTL set expiration time for key
func (c *Cache) SetTTL(ctx context.Context, key string, ttl int64) error {
	res := c.c.Expire(ctx, key, time.Duration(ttl)*time.Second)
	return res.Err()
}

// Incr increment numerical value
func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	res := c.c.Incr(ctx, key)
	return res.Val(), res.Err()
}

// GetInt64 return int64 value of key
func (c *Cache) GetInt64(ctx context.Context, key string) (int64, error) {
	res := c.c.Get(ctx, key)
	return res.Int64()
}

// SetJSON marshal set marshaled json of val
func (c *Cache) SetJSON(ctx context.Context, key string, val interface{}) error {
	msg, err := json.Marshal(val)
	if err != nil {
		return err
	}
	res := c.c.Set(ctx, key, string(msg), 0)
	return res.Err()
}

func (c *Cache) SetTimedJSON(ctx context.Context, key string, val interface{}, ttl int64) error {
	msg, err := json.Marshal(val)
	if err != nil {
		return err
	}
	res := c.c.Set(ctx, key, string(msg), time.Duration(ttl)*time.Second)
	return res.Err()
}

func (c *Cache) SetTimedJSONNX(ctx context.Context, key string, val interface{}, ttl int64) (bool, error) {
	msg, err := json.Marshal(val)
	if err != nil {
		return false, err
	}
	res := c.c.SetArgs(ctx, key, string(msg), redis.SetArgs{
		TTL:  time.Duration(ttl) * time.Second,
		Mode: "NX",
	})
	if err := res.Err(); err != nil && err != redis.Nil {
		return false, err
	}
	return res.Val() == "OK", nil
}

// GetJSON unmarshal json into v
func (c *Cache) GetJSON(ctx context.Context, key string, v interface{}) error {
	res := c.c.Get(ctx, key)
	if res.Err() != nil {
		return res.Err()
	}
	b, err := res.Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func (c *Cache) HGet(ctx context.Context, key, field string) (string, error) {
	h := c.c.HGet(ctx, key, field)
	if h.Err() != nil {
		return "", nil
	}
	return h.Val(), nil
}

func (c *Cache) HSet(ctx context.Context, key, field, value string) error {
	h := c.c.HSet(ctx, key, field, value)
	if h.Err() != nil {
		return h.Err()
	}
	return nil
}

func (c *Cache) HDel(ctx context.Context, key, field string) error {
	h := c.c.HDel(ctx, key, field)
	return h.Err()
}

func (c *Cache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	h := c.c.HGetAll(ctx, key)
	if h.Err() != nil {
		return nil, h.Err()
	}
	return h.Val(), nil
}

func (c *Cache) XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error {
	h := c.c.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: maxLen,
		Approx: approx,
		Values: values,
	})
	if h.Err() != nil {
		return h.Err()
	}
	return nil
}
