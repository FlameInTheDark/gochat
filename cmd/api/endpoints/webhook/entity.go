package webhook

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "webhook"

func (e *entity) Init(router fiber.Router) {
	router.Post("/storage/events", e.StorageEvents)
}

type entity struct {
	name    string
	storage *s3.Client
	at      *attachment.Entity
	log     *slog.Logger
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, storage *s3.Client, log *slog.Logger) server.Entity {
	return &entity{
		name:    entityName,
		storage: storage,
		at:      attachment.New(dbcon),
		log:     log,
	}
}