package developer

import (
	"context"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/avatar"
	"github.com/FlameInTheDark/gochat/internal/database/entities/banner"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/gofiber/fiber/v2"
)

const entityName = "developer"

type entity struct {
	name string
	log  *slog.Logger
	bot  botrepo.Bot
	user botUserRepo
	disc discriminator.Discriminator
	av   avatar.Avatar
	bn   banner.Banner

	attachTTL int64
}

type botUserRepo interface {
	user.User
	CreateUserWithFlags(ctx context.Context, id int64, name string, flags int64) error
}

func New(dbcon *db.CQLCon, pg *pgdb.DB, attachTTLSeconds int64, log *slog.Logger) server.Entity {
	return &entity{
		name:      entityName,
		log:       log,
		bot:       botrepo.New(pg.Conn()),
		user:      user.New(pg.Conn()).(botUserRepo),
		disc:      discriminator.New(pg.Conn()),
		av:        avatar.New(dbcon),
		bn:        banner.New(dbcon),
		attachTTL: attachTTLSeconds,
	}
}

func (e *entity) Name() string {
	return e.name
}

func (e *entity) Init(router fiber.Router) {
	router.Post("/bots", e.CreateBot)
	router.Get("/bots", e.ListBots)
	router.Get("/bots/public", e.SearchPublicBots)
	router.Get("/bots/:bot_id<int>", e.GetBot)
	router.Patch("/bots/:bot_id<int>", e.UpdateBot)
	router.Delete("/bots/:bot_id<int>", e.DeleteBot)
	router.Post("/bots/:bot_id<int>/avatar", e.CreateBotAvatar)
	router.Post("/bots/:bot_id<int>/banner", e.CreateBotBanner)
	router.Get("/bots/authorize/preview", e.PreviewBotAuthorization)

	router.Post("/bots/:bot_id<int>/tokens", e.CreateToken)
	router.Get("/bots/:bot_id<int>/tokens", e.ListTokens)
	router.Delete("/bots/:bot_id<int>/tokens/:token_id<int>", e.RevokeToken)

	router.Post("/bots/:bot_id<int>/grants", e.CreateGrant)
	router.Get("/bots/:bot_id<int>/grants", e.ListGrants)
	router.Delete("/bots/:bot_id<int>/grants/:grant_id<int>", e.RevokeGrant)
}
