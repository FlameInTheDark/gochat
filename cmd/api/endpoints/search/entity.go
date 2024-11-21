package search

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/entities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/entities/message"
	"github.com/FlameInTheDark/gochat/internal/database/entities/role"
	"github.com/FlameInTheDark/gochat/internal/database/entities/user"
	"github.com/FlameInTheDark/gochat/internal/database/entities/userrole"
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
	user  *user.Entity
	disc  *discriminator.Entity
	ch    *channel.Entity
	g     *guild.Entity
	gc    *guildchannels.Entity
	msg   *message.Entity
	at    *attachment.Entity
	uperm *channeluserperm.Entity
	rperm *channelroleperm.Entity
	role  *role.Entity
	ur    *userrole.Entity
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, search *msgsearch.Search, log *slog.Logger) server.Entity {
	return &entity{
		name:   entityName,
		log:    log,
		search: search,
		user:   user.New(dbcon),
		disc:   discriminator.New(dbcon),
		ch:     channel.New(dbcon),
		g:      guild.New(dbcon),
		gc:     guildchannels.New(dbcon),
		msg:    message.New(dbcon),
		at:     attachment.New(dbcon),
		uperm:  channeluserperm.New(dbcon),
		rperm:  channelroleperm.New(dbcon),
		role:   role.New(dbcon),
		ur:     userrole.New(dbcon),
	}
}
