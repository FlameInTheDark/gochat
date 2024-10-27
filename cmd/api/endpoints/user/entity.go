package user

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/entities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/entities/dmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/entities/groupdmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/entities/member"
	"github.com/FlameInTheDark/gochat/internal/database/entities/user"
	"github.com/FlameInTheDark/gochat/internal/database/entities/userrole"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "user"

func (e *entity) Init(router fiber.Router) {
	router.Get("/:user_id", e.GetUser)
	router.Patch("/me", e.ModifyUser)
	router.Get("/me/guilds", e.GetUserGuilds)
	router.Get("/me/guilds/:guild_id/member", e.GetMyGuildMember)
	router.Delete("/me/guilds/:guild_id", e.LeaveGuild)
	router.Post("/me/channels", e.CreateDM)
}

type entity struct {
	name   string
	user   *user.Entity
	member *member.Entity
	guild  *guild.Entity
	urole  *userrole.Entity
	ch     *channel.Entity
	dm     *dmchannel.Entity
	gdm    *groupdmchannel.Entity
	disc   *discriminator.Entity
	log    *slog.Logger
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, log *slog.Logger) server.Entity {
	return &entity{
		name:   entityName,
		user:   user.New(dbcon),
		member: member.New(dbcon),
		guild:  guild.New(dbcon),
		urole:  userrole.New(dbcon),
		ch:     channel.New(dbcon),
		dm:     dmchannel.New(dbcon),
		gdm:    groupdmchannel.New(dbcon),
		disc:   discriminator.New(dbcon),
		log:    log,
	}
}
