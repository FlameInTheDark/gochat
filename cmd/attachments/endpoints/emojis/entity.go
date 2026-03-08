package emojis

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

const entityName = "emojis"

type entity struct {
	name     string
	log      *slog.Logger
	cache    cache.Cache
	perm     rolecheck.RoleCheck
	mqt      mq.SendTransporter
	uploader *upload.EmojiService
}

func (e *entity) Name() string { return e.name }

func New(pg *pgdb.DB, storage *s3.Client, cache cache.Cache, mqt mq.SendTransporter, externalURL string, log *slog.Logger) server.Entity {
	return &entity{
		name:     entityName,
		log:      log,
		cache:    cache,
		perm:     rolecheck.New(pg),
		mqt:      mqt,
		uploader: upload.NewEmojiService(emojirepo.New(pg.Conn()), storage, externalURL, upload.NewFFmpegProcessor()),
	}
}

func (e *entity) Init(router fiber.Router) {
	router.Post("/:guild_id<int>/:emoji_id<int>", e.Upload)
}
