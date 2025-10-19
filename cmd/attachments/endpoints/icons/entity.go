package icons

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/icon"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	pgguild "github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "icons"

type entity struct {
	name          string
	log           *slog.Logger
	storage       *s3.Client
	s3ExternalURL string
	ic            icon.Icon
	gld           pgguild.Guild
	mqt           mq.SendTransporter
}

func (e *entity) Name() string { return e.name }

func New(cql *db.CQLCon, pg *pgdb.DB, storage *s3.Client, externalURL string, mqt mq.SendTransporter, log *slog.Logger) server.Entity {
	return &entity{
		name:          entityName,
		log:           log,
		storage:       storage,
		s3ExternalURL: externalURL,
		ic:            icon.New(cql),
		gld:           pgguild.New(pg.Conn()),
		mqt:           mqt,
	}
}

func (e *entity) Init(router fiber.Router) {
	router.Post("/:guild_id<int>/:icon_id<int>", e.Upload)
}
