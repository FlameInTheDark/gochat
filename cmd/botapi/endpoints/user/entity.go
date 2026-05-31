package user

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/gofiber/fiber/v2"
)

const entityName = "user"

type Entity struct {
	name string
	log  *slog.Logger
}

func New(log *slog.Logger) server.Entity {
	return &Entity{name: entityName, log: log}
}

func (e *Entity) Name() string {
	return e.name
}

func (e *Entity) Init(router fiber.Router) {
	router.Get("/me", e.Me)
}
