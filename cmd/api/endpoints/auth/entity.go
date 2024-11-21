package auth

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/authentication"
	"github.com/FlameInTheDark/gochat/internal/database/entities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/entities/registration"
	"github.com/FlameInTheDark/gochat/internal/database/entities/user"
	"github.com/FlameInTheDark/gochat/internal/mailer"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "auth"

func (e *entity) Init(router fiber.Router) {
	router.Post("/login", e.Login)
	router.Post("/registration", e.Registration)
	router.Post("/confirmation", e.Confirmation)
	// TODO: password recovery method
}

type entity struct {
	name   string
	secret string

	// Services
	log *slog.Logger

	// DB entities
	auth   *authentication.Entity
	user   *user.Entity
	reg    *registration.Entity
	mailer *mailer.Mailer
	disc   *discriminator.Entity
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, m *mailer.Mailer, secret string, log *slog.Logger) server.Entity {
	return &entity{
		name:   entityName,
		secret: secret,
		log:    log,
		auth:   authentication.New(dbcon),
		user:   user.New(dbcon),
		reg:    registration.New(dbcon),
		disc:   discriminator.New(dbcon),
		mailer: m,
	}
}
