package banners

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/banner"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	pguser "github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

const (
	entityName      = "banners"
	coverEntityName = "profile-covers"
)

type entity struct {
	name     string
	log      *slog.Logger
	usr      pguser.User
	bot      botrepo.Bot
	mqt      mq.SendTransporter
	uploader *upload.BannerService
}

func (e *entity) Name() string { return e.name }

func New(cql *db.CQLCon, pg *pgdb.DB, storage *s3.Client, externalURL string, mqt mq.SendTransporter, log *slog.Logger) server.Entity {
	return newEntity(entityName, cql, pg, storage, externalURL, mqt, log)
}

func NewProfileCovers(cql *db.CQLCon, pg *pgdb.DB, storage *s3.Client, externalURL string, mqt mq.SendTransporter, log *slog.Logger) server.Entity {
	return newEntity(coverEntityName, cql, pg, storage, externalURL, mqt, log)
}

func newEntity(name string, cql *db.CQLCon, pg *pgdb.DB, storage *s3.Client, externalURL string, mqt mq.SendTransporter, log *slog.Logger) server.Entity {
	return &entity{
		name:     name,
		log:      log,
		usr:      pguser.New(pg.Conn()),
		bot:      botrepo.New(pg.Conn()),
		mqt:      mqt,
		uploader: upload.NewBannerService(banner.New(cql), storage, externalURL, upload.NewFFmpegProcessor(), bannerMaxDim, bannerMaxSizeBytes, bannerMinWidth, bannerMinHeight),
	}
}

func (e *entity) Init(router fiber.Router) {
	router.Post("/:user_id<int>/:banner_id<int>", e.Upload)
}
