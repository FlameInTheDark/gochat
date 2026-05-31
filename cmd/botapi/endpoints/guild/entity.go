package guild

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	guildrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/gofiber/fiber/v2"
)

const entityName = "guild"

type Entity struct {
	name string
	log  *slog.Logger

	bot  botrepo.Bot
	g    guildrepo.Guild
	gc   guildchannels.GuildChannels
	ch   channel.Channel
	perm rolecheck.RoleCheck
}

func New(pg *pgdb.DB, log *slog.Logger) server.Entity {
	return &Entity{
		name: entityName,
		log:  log,
		bot:  botrepo.New(pg.Conn()),
		g:    guildrepo.New(pg.Conn()),
		gc:   guildchannels.New(pg.Conn()),
		ch:   channel.New(pg.Conn()),
		perm: rolecheck.New(pg),
	}
}

func (e *Entity) Name() string {
	return e.name
}

func (e *Entity) Init(router fiber.Router) {
	router.Get("", e.List)
	router.Get("/:guild_id<int>/channels", e.Channels)
}
