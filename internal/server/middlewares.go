package server

import (
	"fmt"
	"time"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

func (s *Server) AuthMiddleware(secret string) {
	s.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
		Filter: func(c *fiber.Ctx) bool {
			switch string(c.Request().RequestURI()) {
			case "/api/v1/webhook/storage/events":
				fallthrough
			case "/api/v1/auth/login":
				fallthrough
			case "/api/v1/auth/registration":
				fallthrough
			case "/docs/swagger":
				fallthrough
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
			switch string(c.Request().RequestURI()) {
			case "/api/v1/webhook/storage/events":
				fallthrough
			case "/api/v1/auth/login":
				fallthrough
			case "/api/v1/auth/registration":
				fallthrough
			case "/docs/swagger":
				fallthrough
			case "/api/v1/auth/confirmation":
				return true
			}
			return false
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			user, err := helper.GetUser(c)
			if err != nil {
				panic("invalid user")
			}
			return fmt.Sprintf("user:%d:rateLimit", user.Id)
		},
		Expiration: time.Second * time.Duration(exp),
		Storage:    cs,
	}))
}
