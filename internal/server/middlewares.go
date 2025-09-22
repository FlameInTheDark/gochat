package server

import (
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache/vkcpiped"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/redis/go-redis/v9"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

func (s *Server) AuthMiddleware(secret string) {
	s.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
		Claims:     &helper.Claims{},
		Filter: func(c *fiber.Ctx) bool {
			switch string(c.Request().RequestURI()) {
			case "/docs/swagger":
				return true
			case "/api/v1/webhook/storage/events":
				return true
			case "/api/v1/auth/login":
				return true
			case "/api/v1/auth/registration", "/api/v1/auth/confirmation":
				return true
			case "/api/v1/auth/recovery", "/api/v1/auth/reset":
				return true
			}
			return false
		},
	}))
}

func (s *Server) RateLimitMiddleware(limit, exp int) {
	if s.cache == nil {
		return
	}
	cs := NewVKStorage(s.cache)
	s.app.Use(limiter.New(limiter.Config{
		Max: limit,
		Next: func(c *fiber.Ctx) bool {
			// Skip rate limiting for public endpoints (handle both with and without /api/v1 prefix)
			switch string(c.Request().RequestURI()) {
			case "/api/v1/webhook/storage/events", "/webhook/storage/events":
				return true
			case "/api/v1/auth/login", "/auth/login":
				return true
			case "/api/v1/auth/registration", "/auth/registration":
				return true
			case "/api/v1/auth/confirmation", "/auth/confirmation":
				return true
			case "/api/v1/auth/recovery", "/auth/recovery":
				return true
			case "/docs/swagger":
				return true
			}
			return false
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			// Prefer user-scoped rate limiting when authenticated; otherwise, fall back to IP + path
			if user, err := helper.GetUser(c); err == nil && user != nil {
				return fmt.Sprintf("user:%d:rateLimit", user.Id)
			}
			ip := c.IP()
			path := c.Path()
			return fmt.Sprintf("ip:%s:%s:rateLimit", ip, path)
		},
		Expiration: time.Second * time.Duration(exp),
		Storage:    cs,
	}))
}

func (s *Server) RateLimitPipedMiddleware(limit, exp int) {
	if s.cache == nil {
		return
	}

	write := s.cache.Client()
	read := redis.NewClient(&redis.Options{
		Addr:     write.Options().Addr,
		Password: write.Options().Password,
		DB:       write.Options().DB,
		PoolSize: 2048,
	})

	store, err := vkcpiped.NewVKStorage(vkcpiped.VKOptions{
		WriteClient:        write,
		ReadClient:         read,
		PipeSize:           16,
		FlushInterval:      time.Millisecond,
		Workers:            256,
		EnqueueTimeout:     2 * time.Millisecond,
		WaitAckTimeout:     0,
		ExecTimeout:        250 * time.Millisecond,
		DirectWriteTimeout: 50 * time.Millisecond,
		Prefix:             "",
	})
	if err != nil {
		panic(err)
	}

	s.app.Use(limiter.New(limiter.Config{
		Max:        limit,
		Expiration: time.Second * time.Duration(exp),
		KeyGenerator: func(c *fiber.Ctx) string {
			if user, err := helper.GetUser(c); err == nil && user != nil {
				return fmt.Sprintf("user:%d:rateLimit", user.Id)
			}
			return fmt.Sprintf("ip:%s:%s:rateLimit", c.IP(), c.Path())
		},
		Next: func(c *fiber.Ctx) bool {
			switch string(c.Request().RequestURI()) {
			case "/api/v1/webhook/storage/events", "/webhook/storage/events",
				"/api/v1/auth/login", "/auth/login",
				"/api/v1/auth/registration", "/auth/registration",
				"/api/v1/auth/confirmation", "/auth/confirmation",
				"/api/v1/auth/recovery", "/auth/recovery",
				"/docs/swagger":
				return true
			}
			return false
		},
		Storage: store,
	}))
}

func (s *Server) WithIdempotency(client *redis.Client, lifetimeMinutes int64) {
	s.Use(idempotency.New(idempotency.Config{
		Lifetime:  time.Duration(lifetimeMinutes) * time.Minute,
		KeyHeader: "X-Idempotency-Key",
		KeyHeaderValidate: func(k string) error {
			if l, wl := len(k), 36; l != wl {
				return fmt.Errorf("invalid idempotency key: invalid length: %d != %d", l, wl)
			}

			return nil
		},
		Storage: NewRedisIdempotency(client),
		Lock:    NewRedisLocker(client),
	}))
}
