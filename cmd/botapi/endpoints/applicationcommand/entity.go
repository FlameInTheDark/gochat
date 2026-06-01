package applicationcommand

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	cqlappcmd "github.com/FlameInTheDark/gochat/internal/database/entities/applicationcommand"
	guildchannelmessagesdb "github.com/FlameInTheDark/gochat/internal/database/entities/guildchannelmessages"
	messagedb "github.com/FlameInTheDark/gochat/internal/database/entities/message"
	readstatesdb "github.com/FlameInTheDark/gochat/internal/database/entities/readstates"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	appcmdrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/applicationcommand"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/gofiber/fiber/v2"
)

const entityName = ""

type Entity struct {
	name  string
	log   *slog.Logger
	cache *kvs.Cache

	appcmd  appcmdrepo.ApplicationCommand
	payload cqlappcmd.ApplicationCommandInteraction
	user    user.User
	disc    discriminator.Discriminator
	ch      channel.Channel
	gc      guildchannels.GuildChannels
	perm    rolecheck.RoleCheck
	msg     messagedb.Message
	rs      readstatesdb.ReadStates
	gclm    guildchannelmessagesdb.GuildChannelMessages
	mqt     mq.SendTransporter
}

func New(cql *db.CQLCon, pg *pgdb.DB, t mq.SendTransporter, cache *kvs.Cache, log *slog.Logger) server.Entity {
	return &Entity{
		name:    entityName,
		log:     log,
		cache:   cache,
		appcmd:  appcmdrepo.New(pg.Conn()),
		payload: cqlappcmd.New(cql),
		user:    user.New(pg.Conn()),
		disc:    discriminator.New(pg.Conn()),
		ch:      channel.New(pg.Conn()),
		gc:      guildchannels.New(pg.Conn()),
		perm:    rolecheck.New(pg),
		msg:     messagedb.New(cql),
		rs:      readstatesdb.New(cql),
		gclm:    guildchannelmessagesdb.New(cql),
		mqt:     t,
	}
}

func (e *Entity) Name() string {
	return e.name
}

func (e *Entity) Init(router fiber.Router) {
	router.Get("/applications/:application_id<int>/commands", e.ListGlobalCommands)
	router.Post("/applications/:application_id<int>/commands", e.CreateGlobalCommand)
	router.Put("/applications/:application_id<int>/commands", e.BulkOverwriteGlobalCommands)
	router.Get("/applications/:application_id<int>/commands/:command_id<int>", e.GetCommand)
	router.Patch("/applications/:application_id<int>/commands/:command_id<int>", e.EditCommand)
	router.Delete("/applications/:application_id<int>/commands/:command_id<int>", e.DeleteCommand)

	router.Get("/applications/:application_id<int>/guilds/:guild_id<int>/commands", e.ListGuildCommands)
	router.Post("/applications/:application_id<int>/guilds/:guild_id<int>/commands", e.CreateGuildCommand)
	router.Put("/applications/:application_id<int>/guilds/:guild_id<int>/commands", e.BulkOverwriteGuildCommands)
	router.Get("/applications/:application_id<int>/guilds/:guild_id<int>/commands/:command_id<int>", e.GetCommand)
	router.Patch("/applications/:application_id<int>/guilds/:guild_id<int>/commands/:command_id<int>", e.EditCommand)
	router.Delete("/applications/:application_id<int>/guilds/:guild_id<int>/commands/:command_id<int>", e.DeleteCommand)

	router.Post("/interactions/:interaction_id<int>/:interaction_token/callback", e.RespondInteraction)
	router.Get("/webhooks/:application_id<int>/:interaction_token/messages/@original", e.GetOriginalInteractionResponse)
	router.Patch("/webhooks/:application_id<int>/:interaction_token/messages/@original", e.EditOriginalInteractionResponse)
	router.Delete("/webhooks/:application_id<int>/:interaction_token/messages/@original", e.DeleteOriginalInteractionResponse)
	router.Post("/webhooks/:application_id<int>/:interaction_token", e.CreateFollowupMessage)
	router.Patch("/webhooks/:application_id<int>/:interaction_token/messages/:message_id<int>", e.EditFollowupMessage)
	router.Delete("/webhooks/:application_id<int>/:interaction_token/messages/:message_id<int>", e.DeleteFollowupMessage)
}
