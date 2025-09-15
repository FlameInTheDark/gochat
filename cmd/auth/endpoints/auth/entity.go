package auth

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/authentication"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/registration"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/mailer"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "auth"

func (e *entity) Init(router fiber.Router) {
	router.Post("/login", e.Login)
	router.Post("/registration", e.Registration)
	router.Post("/confirmation", e.Confirmation)
	router.Post("/recovery", e.PasswordRecovery)
	router.Post("/reset", e.PasswordReset)
	router.Get("/refresh", e.middleware, e.RefreshToken)
}

type entity struct {
	name   string
	secret string

	// Services
	log *slog.Logger

	// DB entities
	auth       authentication.Authentication
	user       user.User
	reg        registration.Registration
	mailer     *mailer.Mailer
	disc       discriminator.Discriminator
	middleware fiber.Handler
}

func (e *entity) Name() string {
	return e.name
}

func New(pg *pgdb.DB, m *mailer.Mailer, secret string, log *slog.Logger, middlewares fiber.Handler) server.Entity {
	return &entity{
		name:       entityName,
		secret:     secret,
		log:        log,
		auth:       authentication.New(pg.Conn()),
		user:       user.New(pg.Conn()),
		reg:        registration.New(pg.Conn()),
		disc:       discriminator.New(pg.Conn()),
		mailer:     m,
		middleware: middlewares,
	}
}
