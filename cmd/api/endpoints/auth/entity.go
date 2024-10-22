package auth

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/authentication"
	"github.com/FlameInTheDark/gochat/internal/database/entities/registration"
	"github.com/FlameInTheDark/gochat/internal/database/entities/user"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mailer"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "auth"

func (e *entity) Init(router fiber.Router) {
	router.Post("/login", e.Login)
	router.Post("/registration", e.Registration)
	router.Post("/confirmation", e.Confirmation)
}

type entity struct {
	name   string
	secret string
	log    *slog.Logger
	id     *idgen.IDGenerator
	auth   *authentication.Entity
	user   *user.Entity
	reg    *registration.Entity
	mailer *mailer.Mailer
}

func (e *entity) Name() string {
	return e.name
}

func New(id *idgen.IDGenerator, dbcon *db.CQLCon, m *mailer.Mailer, secret string, log *slog.Logger) server.Entity {
	return &entity{
		name:   entityName,
		secret: secret,
		log:    log,
		id:     id,
		auth:   authentication.New(dbcon),
		user:   user.New(dbcon),
		reg:    registration.New(dbcon),
		mailer: m,
	}
}
