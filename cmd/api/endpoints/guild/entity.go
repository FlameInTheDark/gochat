package guild

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/database/entities/icon"
	"github.com/FlameInTheDark/gochat/internal/database/entities/message"
	"github.com/FlameInTheDark/gochat/internal/database/entities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/role"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "guild"

func (e *entity) Init(router fiber.Router) {
	router.Post("", e.Create)
	router.Get("/:guild_id<int>", e.Get)
	router.Patch("/:guild_id<int>", e.Update)
	router.Get("/:guild_id<int>/channel/:channel_id<int>", e.GetChannel)
	router.Get("/:guild_id<int>/channel", e.GetChannels)
	router.Post("/:guild_id<int>/channel", e.CreateChannel)
	router.Post("/:guild_id<int>/category", e.CreateCategory)
	router.Delete("/:guild_id<int>/channel/:channel_id<int>", e.DeleteChannel)
	router.Delete("/:guild_id<int>/category/:category_id<int>", e.DeleteCategory)
}

type entity struct {
	name string

	// Services
	log *slog.Logger
	mqt mq.SendTransporter

	// DB entities
	user  user.User
	disc  discriminator.Discriminator
	ch    channel.Channel
	g     guild.Guild
	gc    guildchannels.GuildChannels
	msg   message.Message
	at    attachment.Attachment
	perm  rolecheck.RoleCheck
	uperm channeluserperm.ChannelUserPerm
	rperm channelroleperm.ChannelRolePerm
	role  role.Role
	ur    userrole.UserRole
	icon  icon.Icon
	memb  member.Member
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, pg *pgdb.DB, mqt mq.SendTransporter, log *slog.Logger) server.Entity {
	return &entity{
		name:  entityName,
		log:   log,
		mqt:   mqt,
		user:  user.New(pg.Conn()),
		disc:  discriminator.New(pg.Conn()),
		ch:    channel.New(pg.Conn()),
		g:     guild.New(pg.Conn()),
		gc:    guildchannels.New(pg.Conn()),
		msg:   message.New(dbcon),
		at:    attachment.New(dbcon),
		perm:  rolecheck.New(dbcon, pg),
		uperm: channeluserperm.New(pg.Conn()),
		rperm: channelroleperm.New(pg.Conn()),
		role:  role.New(pg.Conn()),
		ur:    userrole.New(pg.Conn()),
		icon:  icon.New(dbcon),
		memb:  member.New(pg.Conn()),
	}
}
