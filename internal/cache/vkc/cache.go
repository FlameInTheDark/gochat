package vkc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	c *redis.Client
}

func New(addr string) (*Cache, error) {
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Cache{c: client}, client.Ping(context.Background()).Err()
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
