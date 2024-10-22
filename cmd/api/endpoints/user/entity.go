package user

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/entities/member"
	"github.com/FlameInTheDark/gochat/internal/database/entities/user"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "user"

func (e *entity) Init(router fiber.Router) {
	router.Get("/:user_id", e.GetUser)
	router.Get("/@me/guilds", e.GetUserGuilds)
}

type entity struct {
	name   string
	user   *user.Entity
	member *member.Entity
	guild  *guild.Entity
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
		log:    log,
	}
}
