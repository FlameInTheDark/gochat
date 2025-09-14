package server

import (
	"fmt"
	"time"

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
			case "/api/v1/auth/registration":
				return true
			case "/api/v1/auth/confirmation":
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
