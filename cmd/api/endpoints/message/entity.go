package message

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/database/entities/guildchannelmessages"
	"github.com/FlameInTheDark/gochat/internal/database/entities/readstates"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/database/entities/avatar"
	"github.com/FlameInTheDark/gochat/internal/database/entities/message"
	"github.com/FlameInTheDark/gochat/internal/database/entities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/dmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/groupdmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/role"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
	"github.com/FlameInTheDark/gochat/internal/indexmq"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "message"

func (e *entity) Init(router fiber.Router) {
	router.Post("/channel/:channel_id<int>", e.Send)
	router.Post("/channel/:channel_id<int>/attachment", e.Attachment)
	router.Patch("/channel/:channel_id<int>/:message_id<int>", e.Update)
	router.Delete("/channel/:channel_id<int>/:message_id<int>", e.Delete)
	router.Get("/channel/:channel_id<int>", e.GetMessages)
	router.Post("/channel/:channel_id<int>/:message_id<int>/ack", e.SetReadState)
	router.Post("/channel/:channel_id<int>/typing", e.Typing)
}

type entity struct {
	name        string
	uploadLimit int64
	attachTTL   int64
	// Services
	log *slog.Logger
	mqt mq.SendTransporter
	imq *indexmq.IndexMQ

	// DB entities
	user  user.User
	m     member.Member
	disc  discriminator.Discriminator
	ch    channel.Channel
	g     guild.Guild
	gc    guildchannels.GuildChannels
	dmc   dmchannel.DmChannel
	gdmc  groupdmchannel.GroupDMChannel
	msg   message.Message
	at    attachment.Attachment
	perm  rolecheck.RoleCheck
	uperm channeluserperm.ChannelUserPerm
	rperm channelroleperm.ChannelRolePerm
	role  role.Role
	ur    userrole.UserRole
	rs    readstates.ReadStates
	gclm  guildchannelmessages.GuildChannelMessages
	av    avatar.Avatar
	cache cache.Cache
}

func (e *entity) Name() string {
	return e.name
}

func New(cql *db.CQLCon, pg *pgdb.DB, t mq.SendTransporter, imq *indexmq.IndexMQ, uploadLimit int64, attachTTLSeconds int64, cache cache.Cache, log *slog.Logger) server.Entity {

	return &entity{
		name:        entityName,
		uploadLimit: uploadLimit,
		attachTTL:   attachTTLSeconds,
		log:         log,
		mqt:         t,
		imq:         imq,
		av:          avatar.New(cql),
		cache:       cache,
		user:        user.New(pg.Conn()),
		m:           member.New(pg.Conn()),
		disc:        discriminator.New(pg.Conn()),
		ch:          channel.New(pg.Conn()),
		dmc:         dmchannel.New(pg.Conn()),
		gdmc:        groupdmchannel.New(pg.Conn()),
		g:           guild.New(pg.Conn()),
		gc:          guildchannels.New(pg.Conn()),
		msg:         message.New(cql),
		at:          attachment.New(cql),
		perm:        rolecheck.New(cql, pg),
		uperm:       channeluserperm.New(pg.Conn()),
		rperm:       channelroleperm.New(pg.Conn()),
		role:        role.New(pg.Conn()),
		ur:          userrole.New(pg.Conn()),
		rs:          readstates.New(cql),
		gclm:        guildchannelmessages.New(cql),
	}
}
