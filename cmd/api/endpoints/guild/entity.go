package guild

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/entities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/entities/icon"
	"github.com/FlameInTheDark/gochat/internal/database/entities/member"
	"github.com/FlameInTheDark/gochat/internal/database/entities/message"
	"github.com/FlameInTheDark/gochat/internal/database/entities/role"
	"github.com/FlameInTheDark/gochat/internal/database/entities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/entities/user"
	"github.com/FlameInTheDark/gochat/internal/database/entities/userrole"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
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
	user  *user.Entity
	disc  *discriminator.Entity
	ch    *channel.Entity
	g     *guild.Entity
	gc    *guildchannels.Entity
	msg   *message.Entity
	at    *attachment.Entity
	perm  *rolecheck.Entity
	uperm *channeluserperm.Entity
	rperm *channelroleperm.Entity
	role  *role.Entity
	ur    *userrole.Entity
	icon  *icon.Entity
	memb  *member.Entity
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, pg *pgdb.DB, mqt mq.SendTransporter, log *slog.Logger) server.Entity {
	return &entity{
		name:  entityName,
		log:   log,
		mqt:   mqt,
		user:  user.New(dbcon),
		disc:  discriminator.New(dbcon),
		ch:    channel.New(dbcon),
		g:     guild.New(pg.Conn()),
		gc:    guildchannels.New(dbcon),
		msg:   message.New(dbcon),
		at:    attachment.New(dbcon),
		perm:  rolecheck.New(dbcon, pg),
		uperm: channeluserperm.New(dbcon),
		rperm: channelroleperm.New(dbcon),
		role:  role.New(dbcon),
		ur:    userrole.New(dbcon),
		icon:  icon.New(dbcon),
		memb:  member.New(dbcon),
	}
}
