package voice

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/gofiber/fiber/v2"
)

const entityName = "voice"

type entity struct {
	log     *slog.Logger
	regions []Region
}

func New(regions []Region, log *slog.Logger) server.Entity {
	return &entity{log: log, regions: regions}
}

func (e *entity) Name() string { return entityName }

func (e *entity) Init(router fiber.Router) {
	router.Get("/regions", e.GetRegions)
}
