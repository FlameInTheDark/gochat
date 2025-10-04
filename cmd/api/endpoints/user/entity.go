package user

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/dmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/groupdmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/usersettings"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "user"

func (e *entity) Init(router fiber.Router) {
	router.Get("/:user_id", e.GetUser)
	router.Patch("/me", e.ModifyUser)
	router.Get("/me/guilds", e.GetUserGuilds)
	router.Get("/me/guilds/:guild_id<int>/member", e.GetMyGuildMember)
	router.Delete("/me/guilds/:guild_id<int>", e.LeaveGuild)
	router.Post("/me/channels", e.CreateDM)

	// User settings
	router.Get("/me/settings", e.GetUserSettings)
	router.Post("/me/settings", e.SetUserSettings)
}

type entity struct {
	name string

	// Services
	log *slog.Logger
	mqt mq.SendTransporter

	// DB entities
	user   user.User
	member member.Member
	guild  guild.Guild
	urole  userrole.UserRole
	ch     channel.Channel
	dm     dmchannel.DmChannel
	gdm    groupdmchannel.GroupDMChannel
	disc   discriminator.Discriminator
	uset   usersettings.UserSettings
}

func (e *entity) Name() string {
	return e.name
}

func New(pg *pgdb.DB, mqt mq.SendTransporter, log *slog.Logger) server.Entity {
	return &entity{
		name:   entityName,
		log:    log,
		mqt:    mqt,
		user:   user.New(pg.Conn()),
		member: member.New(pg.Conn()),
		guild:  guild.New(pg.Conn()),
		urole:  userrole.New(pg.Conn()),
		ch:     channel.New(pg.Conn()),
		dm:     dmchannel.New(pg.Conn()),
		gdm:    groupdmchannel.New(pg.Conn()),
		disc:   discriminator.New(pg.Conn()),
		uset:   usersettings.New(pg.Conn()),
	}
}
