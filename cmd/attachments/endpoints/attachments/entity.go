package attachments

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

const entityName = "attachments"

type entity struct {
	name     string
	log      *slog.Logger
	uploader *upload.AttachmentService
}

func (e *entity) Name() string { return e.name }

func New(cql *db.CQLCon, pg *pgdb.DB, storage *s3.Client, externalURL string, log *slog.Logger) server.Entity {
	_ = pg
	return &entity{
		name:     entityName,
		log:      log,
		uploader: upload.NewAttachmentService(attachment.New(cql), storage, externalURL, upload.NewFFmpegProcessor()),
	}
}

func (e *entity) Init(router fiber.Router) {
	router.Post("/:channel_id<int>/:attachment_id<int>", e.Upload)
}
