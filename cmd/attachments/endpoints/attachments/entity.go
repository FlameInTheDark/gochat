package attachments

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "attachments"

type entity struct {
	name          string
	log           *slog.Logger
	storage       *s3.Client
	s3ExternalURL string
	at            attachment.Attachment
}

func (e *entity) Name() string { return e.name }

func New(cql *db.CQLCon, storage *s3.Client, externalURL string, log *slog.Logger) server.Entity {
	return &entity{
		name:          entityName,
		log:           log,
		storage:       storage,
		s3ExternalURL: externalURL,
		at:            attachment.New(cql),
	}
}

func (e *entity) Init(router fiber.Router) {
	router.Post("/:channel_id<int>/:attachment_id<int>", e.Upload)
}
