package search

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/database/entities/message"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/role"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
	"github.com/FlameInTheDark/gochat/internal/msgsearch"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "search"

func (e *entity) Init(router fiber.Router) {
	router.Post("/:guild_id<int>/messages", e.Search)
}

type entity struct {
	name string

	// Services
	log    *slog.Logger
	search *msgsearch.Search

	// DB entities
	user  user.User
	disc  discriminator.Discriminator
	ch    channel.Channel
	g     guild.Guild
	gc    guildchannels.GuildChannels
	msg   message.Message
	at    attachment.Attachment
	uperm channeluserperm.ChannelUserPerm
	rperm channelroleperm.ChannelRolePerm
	role  role.Role
	ur    userrole.UserRole
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, pg *pgdb.DB, search *msgsearch.Search, log *slog.Logger) server.Entity {
	return &entity{
		name:   entityName,
		log:    log,
		search: search,
		user:   user.New(pg.Conn()),
		disc:   discriminator.New(pg.Conn()),
		ch:     channel.New(pg.Conn()),
		g:      guild.New(pg.Conn()),
		gc:     guildchannels.New(pg.Conn()),
		msg:    message.New(dbcon),
		at:     attachment.New(dbcon),
		uperm:  channeluserperm.New(pg.Conn()),
		rperm:  channelroleperm.New(pg.Conn()),
		role:   role.New(pg.Conn()),
		ur:     userrole.New(pg.Conn()),
	}
}
