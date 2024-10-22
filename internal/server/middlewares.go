package server

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) AuthMiddleware(secret string) {
	s.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
		Filter: func(c *fiber.Ctx) bool {
			switch string(c.Request().RequestURI()) {
			case "/api/v1/auth/login":
				fallthrough
			case "/api/v1/auth/registration":
				fallthrough
			case "/api/v1/auth/confirmation":
				return true
			}
			return false
		},
	}))
}
