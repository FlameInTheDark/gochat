package server

import "github.com/gofiber/fiber/v2"

type Entity interface {
	Init(group fiber.Router)
	Name() string
}
