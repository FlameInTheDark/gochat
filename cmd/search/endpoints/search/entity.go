package search

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/database/entities/icon"
	"github.com/FlameInTheDark/gochat/internal/database/entities/message"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guilddiscovery"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/role"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
	"github.com/FlameInTheDark/gochat/internal/guildsearch"
	"github.com/FlameInTheDark/gochat/internal/msgsearch"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "search"

type permissionChecker interface {
	ChannelPerm(ctx context.Context, guildID, channelID, userID int64, perm ...permissions.RolePermission) (*model.Channel, *model.GuildChannel, *model.Guild, bool, error)
}

func (e *entity) Init(router fiber.Router) {
	router.Post("/messages", e.SearchChannel)
	router.Post("/:guild_id<int>/messages", e.Search)
	router.Get("/guilds", e.SearchGuilds)
	router.Get("/guild-tags", e.SearchGuildTags)
}

type entity struct {
	name string

	// Services
	log         *slog.Logger
	search      *msgsearch.Search
	guildSearch *guildsearch.Search
	perm        permissionChecker

	// DB entities
	user  user.User
	disc  discriminator.Discriminator
	ch    channel.Channel
	g     guild.Guild
	gc    guildchannels.GuildChannels
	msg   message.Message
	at    attachment.Attachment
	icon  icon.Icon
	gd    guilddiscovery.GuildDiscovery
	uperm channeluserperm.ChannelUserPerm
	rperm channelroleperm.ChannelRolePerm
	role  role.Role
	ur    userrole.UserRole
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, pg *pgdb.DB, search *msgsearch.Search, guildSearch *guildsearch.Search, log *slog.Logger) server.Entity {
	return &entity{
		name:        entityName,
		log:         log,
		search:      search,
		guildSearch: guildSearch,
		perm:        rolecheck.New(pg),
		user:        user.New(pg.Conn()),
		disc:        discriminator.New(pg.Conn()),
		ch:          channel.New(pg.Conn()),
		g:           guild.New(pg.Conn()),
		gc:          guildchannels.New(pg.Conn()),
		msg:         message.New(dbcon),
		at:          attachment.New(dbcon),
		icon:        icon.New(dbcon),
		gd:          guilddiscovery.New(pg.Conn()),
		uperm:       channeluserperm.New(pg.Conn()),
		rperm:       channelroleperm.New(pg.Conn()),
		role:        role.New(pg.Conn()),
		ur:          userrole.New(pg.Conn()),
	}
}
