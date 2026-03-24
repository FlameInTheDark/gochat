package emoji

import (
	"context"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/icon"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
	guildrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	memberrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "info"

type entity struct {
	log   *slog.Logger
	cache cache.Cache

	emoji  emojirepo.Emoji
	guild  guildrepo.Guild
	member memberrepo.Member
	icon   icon.Icon
}

type emojiLookupReader interface {
	GetEmojiLookup(ctx context.Context, emojiID int64) (model.EmojiLookup, error)
}

type guildReader interface {
	GetGuildById(ctx context.Context, id int64) (model.Guild, error)
}

type memberChecker interface {
	IsGuildMember(ctx context.Context, guildId, userId int64) (bool, error)
}

func New(cql *db.CQLCon, pg *pgdb.DB, cache cache.Cache, log *slog.Logger) server.Entity {
	return &entity{
		log:    log,
		cache:  cache,
		emoji:  emojirepo.New(pg.Conn()),
		guild:  guildrepo.New(pg.Conn()),
		member: memberrepo.New(pg.Conn()),
		icon:   icon.New(cql),
	}
}

func (e *entity) Name() string { return entityName }

func (e *entity) Init(router fiber.Router) {
	router.Get("/emoji/:emoji_id<int>", e.GetInfo)
}
