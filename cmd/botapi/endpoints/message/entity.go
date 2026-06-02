package message

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	guildchannelmessagesdb "github.com/FlameInTheDark/gochat/internal/database/entities/guildchannelmessages"
	messagedb "github.com/FlameInTheDark/gochat/internal/database/entities/message"
	reactiondb "github.com/FlameInTheDark/gochat/internal/database/entities/reaction"
	readstatesdb "github.com/FlameInTheDark/gochat/internal/database/entities/readstates"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/gofiber/fiber/v2"
)

const entityName = "message"

type Entity struct {
	name string
	log  *slog.Logger

	user  user.User
	disc  discriminator.Discriminator
	gc    guildchannels.GuildChannels
	ch    channel.Channel
	perm  rolecheck.RoleCheck
	msg   messagedb.Message
	rs    readstatesdb.ReadStates
	gclm  guildchannelmessagesdb.GuildChannelMessages
	react reactiondb.Reaction
	mqt   mq.SendTransporter
}

func New(cql *db.CQLCon, pg *pgdb.DB, t mq.SendTransporter, log *slog.Logger) server.Entity {
	return &Entity{
		name:  entityName,
		log:   log,
		user:  user.New(pg.Conn()),
		disc:  discriminator.New(pg.Conn()),
		gc:    guildchannels.New(pg.Conn()),
		ch:    channel.New(pg.Conn()),
		perm:  rolecheck.New(pg),
		msg:   messagedb.New(cql),
		rs:    readstatesdb.New(cql),
		gclm:  guildchannelmessagesdb.New(cql),
		react: reactiondb.New(cql),
		mqt:   t,
	}
}

func (e *Entity) Name() string {
	return e.name
}

func (e *Entity) Init(router fiber.Router) {
	router.Post("/channel/:channel_id<int>", e.Send)
	router.Get("/channel/:channel_id<int>", e.List)
	router.Patch("/channel/:channel_id<int>/:message_id<int>", e.Update)
	router.Delete("/channel/:channel_id<int>/:message_id<int>", e.Delete)
	router.Post("/channel/:channel_id<int>/typing", e.Typing)
	router.Put("/channel/:channel_id<int>/:message_id<int>/reactions/:reaction_name", e.AddReaction)
	router.Delete("/channel/:channel_id<int>/:message_id<int>/reactions/:reaction_name", e.RemoveReaction)
	router.Get("/channel/:channel_id<int>/:message_id<int>/reactions/:reaction_name", e.GetReactionUsers)
}
