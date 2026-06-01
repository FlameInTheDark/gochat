package applicationcommand

import (
	"log/slog"
	"time"

	appcmddispatch "github.com/FlameInTheDark/gochat/internal/applicationcommanddispatch"
	"github.com/FlameInTheDark/gochat/internal/botgateway"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	cqlappcmd "github.com/FlameInTheDark/gochat/internal/database/entities/applicationcommand"
	cqlavatar "github.com/FlameInTheDark/gochat/internal/database/entities/avatar"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	appcmdrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/applicationcommand"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/gofiber/fiber/v2"
	natsio "github.com/nats-io/nats.go"
)

const entityName = ""
const sessionRegistryTTL = 75 * time.Second

type entity struct {
	name string
	log  *slog.Logger
	nc   *natsio.Conn

	appcmd   appcmdrepo.ApplicationCommand
	payload  cqlappcmd.ApplicationCommandInteraction
	av       cqlavatar.Avatar
	bot      botrepo.Bot
	user     user.User
	disc     discriminator.Discriminator
	ch       channel.Channel
	gc       guildchannels.GuildChannels
	perm     rolecheck.RoleCheck
	cache    *kvs.Cache
	registry appcmddispatch.BotSessionRegistry
}

func New(cql *db.CQLCon, pg *pgdb.DB, cache *kvs.Cache, nc *natsio.Conn, log *slog.Logger) server.Entity {
	return &entity{
		name:     entityName,
		log:      log,
		nc:       nc,
		appcmd:   appcmdrepo.New(pg.Conn()),
		payload:  cqlappcmd.New(cql),
		av:       cqlavatar.New(cql),
		bot:      botrepo.New(pg.Conn()),
		user:     user.New(pg.Conn()),
		disc:     discriminator.New(pg.Conn()),
		ch:       channel.New(pg.Conn()),
		gc:       guildchannels.New(pg.Conn()),
		perm:     rolecheck.New(pg),
		cache:    cache,
		registry: botgateway.NewRegistry(cache, sessionRegistryTTL),
	}
}

func (e *entity) Name() string {
	return e.name
}

func (e *entity) Init(router fiber.Router) {
	router.Get("/guilds/:guild_id<int>/application-command-index", e.GuildApplicationCommandIndex)
	router.Post("/application-commands/interactions", e.InvokeCommand)
	router.Post("/application-commands/autocomplete", e.AutocompleteCommand)
}
