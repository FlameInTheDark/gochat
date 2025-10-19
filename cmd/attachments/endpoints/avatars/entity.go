package avatars

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/avatar"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	pguser "github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "avatars"

type entity struct {
	name          string
	log           *slog.Logger
	storage       *s3.Client
	s3ExternalURL string
	av            avatar.Avatar
	usr           pguser.User
	mqt           mq.SendTransporter
}

func (e *entity) Name() string { return e.name }

func New(cql *db.CQLCon, pg *pgdb.DB, storage *s3.Client, externalURL string, mqt mq.SendTransporter, log *slog.Logger) server.Entity {
	return &entity{
		name:          entityName,
		log:           log,
		storage:       storage,
		s3ExternalURL: externalURL,
		av:            avatar.New(cql),
		usr:           pguser.New(pg.Conn()),
		mqt:           mqt,
	}
}

func (e *entity) Init(router fiber.Router) {
	router.Post("/:user_id<int>/:avatar_id<int>", e.Upload)
}
